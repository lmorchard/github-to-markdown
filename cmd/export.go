package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/lmorchard/github-to-markdown/internal/timewindow"
	"github.com/spf13/cobra"
)

// exportCmd is the orchestrator-facing entry point with the canonical
// `--since/--until/-o` flag shape shared across all *-to-markdown tools.
// Other settings (token, user, template, include-private) are read from the
// config file or environment.
var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Orchestrator-friendly export of GitHub activity to markdown",
	Long: `Fetch recent GitHub activity and render it as markdown over a canonical
time window.

The --since/--until flag shape matches the contract used by me-to-markdown
and the rest of the *-to-markdown tools. The token, user, template, and
include-private options are read from the config file or environment;
this subcommand exposes only the orchestrator-facing flags.

Example usage:
  github-to-markdown export --since 168h
  github-to-markdown export --since 2026-05-11 --until 2026-05-18 -o activity.md`,
	RunE: func(cmd *cobra.Command, args []string) error {
		since, _ := cmd.Flags().GetString("since")
		until, _ := cmd.Flags().GetString("until")
		output, _ := cmd.Flags().GetString("output")

		now := time.Now()

		sinceTime, err := timewindow.Parse(since, now, false)
		if err != nil {
			return fmt.Errorf("--since: %w", err)
		}

		untilTime := now
		if until != "" {
			untilTime, err = timewindow.Parse(until, now, true)
			if err != nil {
				return fmt.Errorf("--until: %w", err)
			}
		}

		return runExportPipeline(context.Background(), sinceTime, untilTime, output)
	},
}

func init() {
	rootCmd.AddCommand(exportCmd)

	exportCmd.Flags().String("since", "", "Start of time window (YYYY-MM-DD or Go duration like 168h) — required")
	exportCmd.Flags().String("until", "", "End of time window (YYYY-MM-DD, defaults to now)")
	exportCmd.Flags().StringP("output", "o", "", "Output file (default: stdout)")
	_ = exportCmd.MarkFlagRequired("since")
}
