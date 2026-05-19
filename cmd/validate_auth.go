package cmd

import (
	"context"
	"fmt"

	"github.com/lmorchard/github-to-markdown/internal/github"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// validateAuthCmd answers "do my credentials work?" with a single-line
// stdout and a clean exit code. Suitable for orchestrators or scripts
// gating behavior on auth health.
var validateAuthCmd = &cobra.Command{
	Use:   "validate-auth",
	Short: "Check whether the configured GitHub token is accepted",
	Long: `Run a minimal authenticated request (GET /user) against the GitHub
API and exit 0 if the token is accepted, non-zero otherwise.

Reads the token from the usual sources (--token flag / GITHUB_TOKEN
env var / config file).`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		token := viper.GetString("token")
		if token == "" {
			return fmt.Errorf("no GitHub token configured (set GITHUB_TOKEN or `token:`)")
		}

		client := github.NewClient(token)
		login, err := client.AuthenticatedUser(context.Background())
		if err != nil {
			return fmt.Errorf("github auth check: %w", err)
		}
		fmt.Printf("validate-auth: ok (authenticated as @%s)\n", login)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(validateAuthCmd)
}
