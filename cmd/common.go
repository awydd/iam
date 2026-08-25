package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/awydd/iam/conf"
	"github.com/awydd/iam/internal/infra/cache/redis"
	"github.com/awydd/iam/internal/infra/database"
	"github.com/awydd/iam/internal/logger"
	"github.com/awydd/iam/pkg/password"
)

func setup() error {
	if err := conf.Init(); err != nil {
		return fmt.Errorf("failed to init config: %w", err)
	}

	cfg := conf.Get()

	if err := logger.Init(cfg.Logger); err != nil {
		return fmt.Errorf("failed to init logger: %w", err)
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
		return fmt.Errorf("failed to init database: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := database.Migrate(ctx); err != nil {
		return fmt.Errorf("failed to run database migration: %w", err)
	}

	if err := redis.Init(cfg.Redis); err != nil {
		return fmt.Errorf("failed to init redis: %w", err)
	}

	logger.Info("setup completed, env=%s", cfg.Env)
	return nil
}

func release() {
	if err := database.Close(); err != nil {
		logger.Error("failed to close database: %s", err)
	}

	if err := redis.Close(); err != nil {
		logger.Error("failed to close redis: %s", err)
	}

	logger.Info("releasing resources")
	_ = logger.Sync()
}
