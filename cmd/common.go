package cmd

import (
	"fmt"

	"github.com/awydd/iam/conf"
	"github.com/awydd/iam/internal/infra/cache/redis"
	"github.com/awydd/iam/internal/logger"
)

func setup() error {
	if err := conf.Init(); err != nil {
		return fmt.Errorf("failed to init config: %w", err)
	}

	cfg := conf.Get()

	if err := logger.Init(cfg.Logger); err != nil {
		return fmt.Errorf("failed to init logger: %w", err)
	}

	if err := redis.Init(cfg.Redis); err != nil {
		return fmt.Errorf("failed to init redis: %w", err)
	}

	logger.Info("setup completed, env=%s", cfg.Env)
	return nil
}

func release() {
	if err := redis.Close(); err != nil {
		logger.Error("failed to close redis: %s", err)
	}

	logger.Info("releasing resources")
	_ = logger.Sync()
}
