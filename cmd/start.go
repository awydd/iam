package cmd

import (
	"os"
	"os/exec"

	"github.com/awydd/iam/internal/logger"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start IAM server in background",
	Run: func(cmd *cobra.Command, args []string) {
		start()
	},
}

func start() {
	if pid, err := getRunningPID(); err == nil {
		logger.Warn("iam server is already running with PID %d", pid)
		return
	}

	binary, err := os.Executable()
	if err != nil {
		logger.Fatal("failed to get executable path: %v", err)
	}

	subArgs := []string{"server"}

	childCmd := exec.Command(binary, subArgs...)

	childCmd.Stdout = nil
	childCmd.Stderr = nil
	childCmd.Stdin = nil

	if err := childCmd.Start(); err != nil {
		logger.Fatal("failed to start background process: %v", err)
	}

	logger.Info("iam server started in background (PID: %d)", childCmd.Process.Pid)
}

func init() {
	rootCmd.AddCommand(startCmd)
}
