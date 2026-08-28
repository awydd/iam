package cmd

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"
	"time"

	"github.com/awydd/iam/conf"
	"github.com/awydd/iam/internal/infra/database"
	"github.com/awydd/iam/internal/logger"
	httpserver "github.com/awydd/iam/internal/transport/http"
	"github.com/awydd/iam/internal/wire"
	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "运行 IAM 服务端",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := writePID(); err != nil {
			return fmt.Errorf("写入 PID 文件失败: %w", err)
		}

		if err := setup(); err != nil {
			removePID()
			return fmt.Errorf("初始化失败: %w", err)
		}
		defer release()
		defer removePID()

		if err := server(); err != nil {
			return fmt.Errorf("服务端错误: %w", err)
		}
		return nil
	},
}

func server() error {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	startTokenCleanupScheduler(ctx, time.Hour)

	app := wire.InitApp(database.DB())
	srv, err := httpserver.New(&conf.Get().HTTP, app)
	if err != nil {
		return fmt.Errorf("创建 HTTP 服务失败: %w", err)
	}

	errCh := make(chan error, 1)
	go func() {
		if err := srv.Start(); err != nil {
			errCh <- err
		}
	}()

	logger.Info("server started, press Ctrl+C to stop")

	select {
	case <-ctx.Done():
	case err := <-errCh:
		stop()
		return fmt.Errorf("HTTP 服务错误: %w", err)
	}

	logger.Info("shutting down server...")

	if err := srv.Shutdown(context.Background()); err != nil {
		logger.Error("优雅关闭失败: %s", err)
	}

	logger.Info("server exited")
	return nil
}

func init() {
	rootCmd.AddCommand(serverCmd)
}
