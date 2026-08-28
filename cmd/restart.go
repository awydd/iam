package cmd

import (
	"time"

	"github.com/spf13/cobra"
)

var restartCmd = &cobra.Command{
	Use:   "restart",
	Short: "重启 IAM 后台服务",
	Run: func(cmd *cobra.Command, args []string) {
		stop()

		time.Sleep(500 * time.Millisecond)

		start()
	},
}

func init() {
	rootCmd.AddCommand(restartCmd)
}
