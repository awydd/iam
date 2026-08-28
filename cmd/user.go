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
	Short: "用户管理",
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
	Short: "更新系统用户的用户名、邮箱或密码",
	RunE: func(cmd *cobra.Command, args []string) error {
		if updateSystemUsername == "" && updateSystemEmail == "" && updateSystemPassword == "" {
			return fmt.Errorf("必须提供 --username、--email 或 --password 中的至少一项")
		}
		if updateSystemPassword != "" && len(updateSystemPassword) < 6 {
			return fmt.Errorf("密码长度不能少于 6 个字符")
		}
		return runUpdateSystemUser()
	},
}

func init() {
	userCmd.AddCommand(updateSystemCmd)

	updateSystemCmd.Flags().StringVar(&updateSystemUsername, "username", "", "新用户名")
	updateSystemCmd.Flags().StringVar(&updateSystemEmail, "email", "", "新邮箱")
	updateSystemCmd.Flags().StringVar(&updateSystemPassword, "password", "", "新密码")
}

func runUpdateSystemUser() error {
	if err := conf.Init(); err != nil {
		return fmt.Errorf("初始化配置失败: %w", err)
	}
	cfg := conf.Get()

	if err := database.Init(cfg.Database, false); err != nil {
		return fmt.Errorf("初始化数据库失败: %w", err)
	}
	defer database.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := database.Migrate(ctx); err != nil {
		return fmt.Errorf("执行数据库迁移失败: %w", err)
	}

	if err := redis.Init(cfg.Redis); err != nil {
		return fmt.Errorf("初始化 Redis 失败: %w", err)
	}
	defer redis.Close()

	userStore := store.NewUserStore(database.DB())

	sysUser, err := userStore.GetSystem(ctx)
	if err != nil {
		return fmt.Errorf("获取系统用户失败: %w", err)
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
		return fmt.Errorf("检查重复项失败: %w", err)
	}
	if exists {
		return fmt.Errorf("用户名或邮箱已被其他账号使用")
	}

	hashed := ""
	if updateSystemPassword != "" {
		hashed, err = password.Hash(updateSystemPassword)
		if err != nil {
			return fmt.Errorf("密码哈希生成失败: %w", err)
		}
	}

	if _, err := userStore.Update(ctx, sysUser.ID, username, email, sysUser.Status, hashed); err != nil {
		return fmt.Errorf("更新系统用户失败: %w", err)
	}

	tokenStore := store.NewTokenStore(database.DB())
	tokenCache := redis.NewTokenCache()

	if err := invalidateSystemUserCaches(ctx, sysUser.ID, tokenStore, tokenCache); err != nil {
		logger.Error("invalidate system user caches failed: %v", err)
		fmt.Println("警告：完全清除缓存失败，系统用户可能需要手动重新登录")
	}

	logger.Info("system user updated: username=%s email=%s password_changed=%v", username, email, updateSystemPassword != "")
	fmt.Printf("系统用户更新成功: username=%s email=%s password_changed=%v\n", username, email, updateSystemPassword != "")
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
		return fmt.Errorf("获取活跃会话列表失败: %w", err)
	}

	if err := tokenStore.RevokeByUserID(ctx, userID); err != nil {
		return fmt.Errorf("撤销令牌失败: %w", err)
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
