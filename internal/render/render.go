// Package render produces a markdown document from normalized GitHub events.
package render

import (
	"fmt"
	"io"
	"os"
	"sort"
	"text/template"
	"time"

	"github.com/lmorchard/github-to-markdown/internal/github"
	"github.com/lmorchard/github-to-markdown/internal/templates"
)

// Data is the value passed to the template.
type Data struct {
	User      string
	Since     time.Time
	Until     time.Time
	Generated time.Time
	Events    []github.Normalized
	ByRepo    map[string][]github.Normalized
	RepoOrder []string // repos sorted by most-recent event first
}

// Render writes the markdown document to w. If templatePath is empty, the
// embedded default template is used.
func Render(w io.Writer, templatePath string, data Data) error {
	tmplText, err := loadTemplate(templatePath)
	if err != nil {
		return err
	}
	tmpl, err := template.New("activity").Parse(tmplText)
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}
	if err := tmpl.Execute(w, data); err != nil {
		return fmt.Errorf("execute template: %w", err)
	}
	return nil
}

func loadTemplate(path string) (string, error) {
	if path == "" {
		t, err := templates.GetDefaultTemplate()
		if err != nil {
			return "", fmt.Errorf("load embedded template: %w", err)
		}
		return t, nil
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read template %s: %w", path, err)
	}
	return string(b), nil
}

// Build constructs a Data value from raw events plus the requested window.
// Events are normalized and grouped by repository; repos are ordered by the
// timestamp of their most recent event.
func Build(user string, since, until time.Time, events []github.Event) Data {
	normalized := make([]github.Normalized, 0, len(events))
	byRepo := make(map[string][]github.Normalized)
	latest := make(map[string]time.Time)

	for _, ev := range events {
		n := github.Normalize(ev)
		normalized = append(normalized, n)
		byRepo[n.Repo] = append(byRepo[n.Repo], n)
		if n.At.After(latest[n.Repo]) {
			latest[n.Repo] = n.At
		}
	}

	repos := make([]string, 0, len(byRepo))
	for r := range byRepo {
		repos = append(repos, r)
	}
	sort.Slice(repos, func(i, j int) bool {
		return latest[repos[i]].After(latest[repos[j]])
	})

	return Data{
		User:      user,
		Since:     since,
		Until:     until,
		Generated: time.Now(),
		Events:    normalized,
		ByRepo:    byRepo,
		RepoOrder: repos,
	}
}
