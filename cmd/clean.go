package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/awydd/iam/conf"
	"github.com/awydd/iam/internal/biz"
	"github.com/awydd/iam/internal/infra/database"
	"github.com/awydd/iam/internal/infra/store"
	"github.com/awydd/iam/internal/logger"
	"github.com/spf13/cobra"
)

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean up stale or expired data (see subcommands)",
}

var cleanTokenCmd = &cobra.Command{
	Use:   "token",
	Short: "Clean up expired or revoked tokens (runs once then exits)",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runTokenCleanupOnce()
	},
}

func runTokenCleanupOnce() error {
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

	tokenStore := store.NewTokenStore(database.DB())
	tokenBiz := biz.NewTokenBiz(tokenStore, nil)

	total, err := tokenBiz.CleanExpiredOrRevoked(ctx)
	if err != nil {
		return fmt.Errorf("token cleanup failed: %w", err)
	}

	logger.Info("token cleanup done, removed %d rows total", total)
	return nil
}

var cleanKeypairCmd = &cobra.Command{
	Use:   "keypair",
	Short: "Clean up retired keypairs older than the retention window (runs once then exits)",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runKeypairCleanupOnce()
	},
}

func runKeypairCleanupOnce() error {
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

	keypairStore := store.NewKeypairStore(database.DB())
	keypairBiz := biz.NewKeypairBiz(keypairStore, store.NewTransactor(database.DB()))

	total, err := keypairBiz.CleanRetired(ctx)
	if err != nil {
		return fmt.Errorf("keypair cleanup failed: %w", err)
	}

	logger.Info("keypair cleanup done, removed %d rows total", total)
	return nil
}

func init() {
	rootCmd.AddCommand(cleanCmd)
	cleanCmd.AddCommand(cleanTokenCmd)
	cleanCmd.AddCommand(cleanKeypairCmd)
}
