package cmd

import (
	"time"

	"github.com/awydd/iam/internal/logger"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop IAM background server",
	Run: func(cmd *cobra.Command, args []string) {
		stop()
	},
}

func stop() {
	pid, err := getRunningPID()
	if err != nil {
		logger.Warn("iam server is not running")
		return
	}

	logger.Info("stopping iam server (PID: %d)...", pid)
	if err := killProcess(pid); err != nil {
		logger.Error("failed to stop server: %v", err)
		return
	}

	for range 10 {
		if !isProcessRunning(pid) {
			removePID()
			logger.Info("iam server stopped successfully")
			return
		}
		time.Sleep(1 * time.Second)
	}

	logger.Error("failed to verify stop status (timeout)")
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
