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
	Short: "清理过期或失效的数据",
}

var cleanTokenCmd = &cobra.Command{
	Use:   "token",
	Short: "清理过期或已撤销的令牌",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runTokenCleanupOnce()
	},
}

func runTokenCleanupOnce() error {
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

	tokenStore := store.NewTokenStore(database.DB())
	tokenBiz := biz.NewTokenBiz(tokenStore, nil, nil)

	total, err := tokenBiz.CleanExpiredOrRevoked(ctx)
	if err != nil {
		return fmt.Errorf("清理令牌失败: %w", err)
	}

	logger.Info("token cleanup done, removed %d rows total", total)
	return nil
}

var cleanKeypairCmd = &cobra.Command{
	Use:   "keypair",
	Short: "清理超过保留期限的已退役密钥对",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runKeypairCleanupOnce()
	},
}

func runKeypairCleanupOnce() error {
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

	keypairStore := store.NewKeypairStore(database.DB())
	keypairBiz := biz.NewKeypairBiz(keypairStore, store.NewTransactor(database.DB()))

	total, err := keypairBiz.CleanRetired(ctx)
	if err != nil {
		return fmt.Errorf("清理密钥对失败: %w", err)
	}

	logger.Info("keypair cleanup done, removed %d rows total", total)
	return nil
}

func init() {
	rootCmd.AddCommand(cleanCmd)
	cleanCmd.AddCommand(cleanTokenCmd)
	cleanCmd.AddCommand(cleanKeypairCmd)
}
