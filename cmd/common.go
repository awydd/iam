package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/awydd/iam/conf"
	"github.com/awydd/iam/internal/biz"
	"github.com/awydd/iam/internal/infra/cache/redis"
	"github.com/awydd/iam/internal/infra/database"
	"github.com/awydd/iam/internal/infra/store"
	"github.com/awydd/iam/internal/logger"
	"github.com/awydd/iam/pkg/password"
	"github.com/awydd/iam/pkg/utils"
)

func startTokenCleanupScheduler(ctx context.Context, interval time.Duration) {
	tokenStore := store.NewTokenStore(database.DB())
	tokenBiz := biz.NewTokenBiz(tokenStore, nil, nil)

	go func() {
		cleanup := func() {
			runCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
			defer cancel()

			total, err := tokenBiz.CleanExpiredOrRevoked(runCtx)
			if err != nil {
				logger.Error("清理令牌失败: %s", err)
				return
			}
			if total > 0 {
				logger.Info("token cleanup done, removed %d rows total", total)
			}
		}

		cleanup()

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				logger.Info("令牌清理调度器已停止")
				return
			case <-ticker.C:
				cleanup()
			}
		}
	}()
}

func setup() error {
	if err := conf.Init(); err != nil {
		return fmt.Errorf("初始化配置失败: %w", err)
	}

	cfg := conf.Get()

	if len(cfg.HTTP.TrustedProxies) > 0 {
		utils.SetTrustedProxies(cfg.HTTP.TrustedProxies)
	}

	utils.InitCookieUtil(cfg.Cookie)

	if err := logger.Init(cfg.Logger); err != nil {
		return fmt.Errorf("初始化日志失败: %w", err)
	}

	password.Init(&password.Config{
		Time:           cfg.Password.Time,
		Memory:         cfg.Password.Memory,
		Threads:        cfg.Password.Threads,
		KeyLength:      cfg.Password.KeyLength,
		SaltLen:        cfg.Password.SaltLen,
		MaxConcurrency: cfg.Password.MaxConcurrency,
		WaitTimeout:    cfg.Password.WaitTimeout,
	})

	if err := database.Init(cfg.Database, cfg.IsDev()); err != nil {
		return fmt.Errorf("初始化数据库失败: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := database.Migrate(ctx); err != nil {
		return fmt.Errorf("执行数据库迁移失败: %w", err)
	}

	if err := redis.Init(cfg.Redis); err != nil {
		return fmt.Errorf("初始化 Redis 失败: %w", err)
	}

	logger.Info("setup completed, env=%s", cfg.Env)
	return nil
}

func release() {
	if err := database.Close(); err != nil {
		logger.Error("关闭数据库失败: %s", err)
	}

	if err := redis.Close(); err != nil {
		logger.Error("关闭 Redis 失败: %s", err)
	}

	logger.Info("正在释放资源")
	_ = logger.Sync()
}
