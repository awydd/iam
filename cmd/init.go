package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/awydd/iam/conf"
	"github.com/awydd/iam/internal/infra/database"
	"github.com/awydd/iam/internal/infra/database/data"
	"github.com/spf13/cobra"
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

		result, err := data.Init(ctx, database.DB())
		if err != nil {
			return fmt.Errorf("bootstrap failed: %w", err)
		}

		fmt.Println("bootstrap completed:")
		if result.KeypairCreated {
			fmt.Printf("  [new] signing key: kid=%s\n", result.KeypairKid)
		} else {
			fmt.Println("  [skip] signing key already exists")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
