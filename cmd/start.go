package cmd

import (
	"os"
	"os/exec"

	"github.com/awydd/iam/internal/logger"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "在后台启动 IAM 服务",
	Run: func(cmd *cobra.Command, args []string) {
		start()
	},
}

func start() {
	if pid, err := getRunningPID(); err == nil {
		logger.Warn("IAM 服务已经在运行中，PID 为 %d", pid)
		return
	}

	binary, err := os.Executable()
	if err != nil {
		logger.Fatal("获取可执行文件路径失败: %v", err)
	}

	subArgs := []string{"server"}

	childCmd := exec.Command(binary, subArgs...)

	childCmd.Stdout = nil
	childCmd.Stderr = nil
	childCmd.Stdin = nil

	if err := childCmd.Start(); err != nil {
		logger.Fatal("启动后台进程失败: %v", err)
	}

	logger.Info("IAM 服务已在后台启动 (PID: %d)", childCmd.Process.Pid)
}

func init() {
	rootCmd.AddCommand(startCmd)
}
