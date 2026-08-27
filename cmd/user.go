package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/awydd/iam/conf"
	"github.com/awydd/iam/internal/infra/cache/redis"
	"github.com/awydd/iam/internal/infra/database"
	"github.com/awydd/iam/internal/infra/store"
	"github.com/awydd/iam/internal/logger"
	"github.com/awydd/iam/pkg/password"
	"github.com/spf13/cobra"
)

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "Manage users",
}

func init() {
	rootCmd.AddCommand(userCmd)
}

var (
	updateSystemUsername string
	updateSystemEmail    string
	updateSystemPassword string
)

var updateSystemCmd = &cobra.Command{
	Use:   "update-system",
	Short: "Update the system user's username/email/password",
	RunE: func(cmd *cobra.Command, args []string) error {
		if updateSystemUsername == "" && updateSystemEmail == "" && updateSystemPassword == "" {
			return fmt.Errorf("at least one of --username, --email, or --password must be provided")
		}
		if updateSystemPassword != "" && len(updateSystemPassword) < 6 {
			return fmt.Errorf("password must be at least 6 characters")
		}
		return runUpdateSystemUser()
	},
}

func init() {
	userCmd.AddCommand(updateSystemCmd)

	updateSystemCmd.Flags().StringVar(&updateSystemUsername, "username", "", "new username")
	updateSystemCmd.Flags().StringVar(&updateSystemEmail, "email", "", "new email")
	updateSystemCmd.Flags().StringVar(&updateSystemPassword, "password", "", "new password")
}

func runUpdateSystemUser() error {
	if err := conf.Init(); err != nil {
		return fmt.Errorf("failed to init config: %w", err)
	}
	cfg := conf.Get()

	if err := database.Init(cfg.Database, false); err != nil {
		return fmt.Errorf("failed to init database: %w", err)
	}
	defer database.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := database.Migrate(ctx); err != nil {
		return fmt.Errorf("failed to run database migration: %w", err)
	}

	if err := redis.Init(cfg.Redis); err != nil {
		return fmt.Errorf("failed to init redis: %w", err)
	}
	defer redis.Close()

	userStore := store.NewUserStore(database.DB())

	sysUser, err := userStore.GetSystem(ctx)
	if err != nil {
		return fmt.Errorf("get system user: %w", err)
	}

	username := sysUser.Username
	if updateSystemUsername != "" {
		username = updateSystemUsername
	}
	email := sysUser.Email
	if updateSystemEmail != "" {
		email = updateSystemEmail
	}

	exists, err := userStore.Duplicate(ctx, username, email, sysUser.ID)
	if err != nil {
		return fmt.Errorf("check duplicate: %w", err)
	}
	if exists {
		return fmt.Errorf("username or email already in use by another account")
	}

	hashed := ""
	if updateSystemPassword != "" {
		hashed, err = password.Hash(updateSystemPassword)
		if err != nil {
			return fmt.Errorf("hash password: %w", err)
		}
	}

	if _, err := userStore.Update(ctx, sysUser.ID, username, email, sysUser.Status, hashed); err != nil {
		return fmt.Errorf("update system user: %w", err)
	}

	tokenStore := store.NewTokenStore(database.DB())
	tokenCache := redis.NewTokenCache()

	if err := invalidateSystemUserCaches(ctx, sysUser.ID, tokenStore, tokenCache); err != nil {
		logger.Error("invalidate system user caches failed: %v", err)
		fmt.Println("warning: failed to fully invalidate caches, system user may need to be manually logged out")
	}

	logger.Info("system user updated: username=%s email=%s password_changed=%v", username, email, updateSystemPassword != "")
	fmt.Printf("system user updated successfully: username=%s email=%s password_changed=%v\n", username, email, updateSystemPassword != "")
	return nil
}

func invalidateSystemUserCaches(
	ctx context.Context,
	userID int,
	tokenStore *store.TokenStore,
	tokenCache *redis.TokenCache,
) error {
	sessionIDs, err := tokenStore.ListActiveSessionsByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("list active sessions: %w", err)
	}

	if err := tokenStore.RevokeByUserID(ctx, userID); err != nil {
		return fmt.Errorf("revoke tokens: %w", err)
	}

	for _, sid := range sessionIDs {
		if err := tokenCache.DelAccess(ctx, sid); err != nil {
			logger.Error("del access cache failed: session_id=%s err=%v", sid, err)
		}
		if err := tokenCache.DelRefresh(ctx, sid); err != nil {
			logger.Error("del refresh cache failed: session_id=%s err=%v", sid, err)
		}
	}

	return nil
}
