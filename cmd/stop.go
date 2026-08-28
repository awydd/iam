package cmd

import (
	"time"

	"github.com/awydd/iam/internal/logger"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "停止 IAM 后台服务",
	Run: func(cmd *cobra.Command, args []string) {
		stop()
	},
}

func stop() {
	pid, err := getRunningPID()
	if err != nil {
		logger.Warn("IAM 服务未运行")
		return
	}

	logger.Info("正在停止 IAM 服务 (PID: %d)...", pid)
	if err := killProcess(pid); err != nil {
		logger.Error("停止服务失败: %v", err)
		return
	}

	for range 10 {
		if !isProcessRunning(pid) {
			removePID()
			logger.Info("IAM 服务已成功停止")
			return
		}
		time.Sleep(1 * time.Second)
	}

	logger.Error("验证停止状态失败（超时）")
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
