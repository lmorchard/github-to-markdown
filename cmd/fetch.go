package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/lmorchard/github-to-markdown/internal/github"
	"github.com/lmorchard/github-to-markdown/internal/render"
	"github.com/spf13/cobra"
)

const defaultLookback = 7 * 24 * time.Hour

func runFetch(cmd *cobra.Command, args []string) error {
	c := GetConfig()

	now := time.Now()
	since, err := parseWhen(c.Since, now.Add(-defaultLookback), false)
	if err != nil {
		return fmt.Errorf("--since: %w", err)
	}
	until, err := parseWhen(c.Until, now, true)
	if err != nil {
		return fmt.Errorf("--until: %w", err)
	}

	return runExportPipeline(context.Background(), since, until, c.Output)
}

// runExportPipeline executes the GitHub events fetch + render pipeline for the
// given time window, writing to outputPath (empty = stdout). All other
// settings (token, user, template, include-private) come from the GetConfig()
// values.
//
// Used by both the root command (fetch behavior) and the export subcommand.
func runExportPipeline(ctx context.Context, since, until time.Time, outputPath string) error {
	c := GetConfig()
	log := GetLogger()

	if c.Token == "" {
		return fmt.Errorf("no GitHub token provided (set GITHUB_TOKEN or pass --token)")
	}
	if !until.After(since) {
		return fmt.Errorf("--until (%s) must be after --since (%s)",
			until.Format(time.RFC3339), since.Format(time.RFC3339))
	}

	client := github.NewClient(c.Token)

	user := c.User
	if user == "" {
		log.Debug("resolving authenticated user")
		var err error
		user, err = client.AuthenticatedUser(ctx)
		if err != nil {
			return fmt.Errorf("resolve user: %w", err)
		}
	}
	log.Infof("Fetching events for %s from %s to %s",
		user, since.Format(time.RFC3339), until.Format(time.RFC3339))

	events, err := client.UserEvents(ctx, user, since, until)
	if err != nil {
		return fmt.Errorf("fetch events: %w", err)
	}
	total := len(events)
	if !c.IncludePrivate {
		events = github.FilterPublic(events)
		if dropped := total - len(events); dropped > 0 {
			log.Infof("Dropped %d private event(s) (pass --include-private to keep)", dropped)
		}
	}
	log.Infof("Got %d event(s) in window", len(events))

	data := render.Build(user, since, until, events)

	out, closeFn, err := openOutput(outputPath)
	if err != nil {
		return err
	}
	defer closeFn()

	if err := render.Render(out, c.Template, data); err != nil {
		return err
	}
	if outputPath != "" {
		log.Infof("Wrote %s", outputPath)
	}
	return nil
}

// parseWhen parses a YYYY-MM-DD or RFC3339 timestamp.
// If `s` is empty, fallback is returned. When endOfDay is true, a date-only
// value is interpreted as end-of-day rather than start-of-day.
func parseWhen(s string, fallback time.Time, endOfDay bool) (time.Time, error) {
	if s == "" {
		return fallback, nil
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}
	if t, err := time.ParseInLocation("2006-01-02", s, time.Local); err == nil {
		if endOfDay {
			t = t.Add(24*time.Hour - time.Nanosecond)
		}
		return t, nil
	}
	return time.Time{}, fmt.Errorf("invalid date %q (expected YYYY-MM-DD or RFC3339)", s)
}

// openOutput returns the writer and a close function. If path is empty,
// stdout is used and the close function is a no-op.
func openOutput(path string) (io.Writer, func(), error) {
	if path == "" {
		return os.Stdout, func() {}, nil
	}
	f, err := os.Create(path)
	if err != nil {
		return nil, nil, fmt.Errorf("create output file: %w", err)
	}
	return f, func() { _ = f.Close() }, nil
}
