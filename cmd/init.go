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
	Short: "Initialize system data",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := conf.Init(); err != nil {
			return fmt.Errorf("failed to init config: %w", err)
		}
		cfg := conf.Get()
		if err := database.Init(cfg.Database, false); err != nil {
			return fmt.Errorf("failed to init database: %w", err)
		}
		defer database.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := database.Migrate(ctx); err != nil {
			return fmt.Errorf("failed to run database migration: %w", err)
		}

		result, err := data.Init(ctx, database.DB(), data.Options{
			Username: initUsername,
			Email:    initEmail,
		})
		if err != nil {
			return fmt.Errorf("bootstrap failed: %w", err)
		}

		fmt.Println("bootstrap completed:")

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

		if result.KeypairCreated {
			fmt.Fprintln(w, "  [new]\tsigning key: kid =", result.KeypairKid)
		} else {
			fmt.Fprintln(w, "  [skip]\tsigning key already exists")
		}

		if result.UserCreated {
			fmt.Fprintln(w, "  [new]\tsystem user created — save the password now, it will not be shown again:")
			fmt.Fprintln(w, "\tusername:\t", result.Username)
			fmt.Fprintln(w, "\temail:\t", result.Email)
			fmt.Fprintln(w, "\tpassword:\t", result.Password)
		} else {
			fmt.Fprintln(w, "  [skip]\tsystem user already exists")
		}

		w.Flush()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().StringVar(&initUsername, "username", "", "username for the bootstrap admin user")
	initCmd.Flags().StringVar(&initEmail, "email", "", "email for the bootstrap admin user")
	_ = initCmd.MarkFlagRequired("username")
	_ = initCmd.MarkFlagRequired("email")
}
