package cmd

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/awydd/iam/conf"
	"github.com/awydd/iam/internal/infra/database"
	"github.com/awydd/iam/internal/infra/database/data"
	"github.com/spf13/cobra"
)

var (
	initUsername string
	initEmail    string
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "初始化系统数据",
	RunE: func(cmd *cobra.Command, args []string) error {
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

		result, err := data.Init(ctx, database.DB(), data.Options{
			Username: initUsername,
			Email:    initEmail,
		})
		if err != nil {
			return fmt.Errorf("系统初始化失败: %w", err)
		}

		fmt.Println("系统初始化完成:")

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

		if result.KeypairCreated {
			fmt.Fprintln(w, "  [new]\t签名密钥: kid =", result.KeypairKid)
		} else {
			fmt.Fprintln(w, "  [skip]\t签名密钥已存在")
		}

		if result.UserCreated {
			fmt.Fprintln(w, "  [new]\t已创建系统用户 — 请立即保存密码，后续将不再显示:")
			fmt.Fprintln(w, "\t用户名:\t", result.Username)
			fmt.Fprintln(w, "\t邮箱:\t", result.Email)
			fmt.Fprintln(w, "\t密码:\t", result.Password)
		} else {
			fmt.Fprintln(w, "  [skip]\t系统用户已存在")
		}

		w.Flush()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().StringVar(&initUsername, "username", "", "初始管理员用户名")
	initCmd.Flags().StringVar(&initEmail, "email", "", "初始管理员邮箱")
	_ = initCmd.MarkFlagRequired("username")
	_ = initCmd.MarkFlagRequired("email")
}
