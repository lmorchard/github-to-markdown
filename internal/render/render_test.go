package render

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/lmorchard/github-to-markdown/internal/github"
)

func TestRender_Default_NoEvents(t *testing.T) {
	data := Build("lmorchard",
		time.Date(2026, 5, 11, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 5, 18, 0, 0, 0, 0, time.UTC),
		nil)
	var buf bytes.Buffer
	if err := Render(&buf, "", data); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "lmorchard") {
		t.Errorf("missing user in output: %s", out)
	}
	if !strings.Contains(out, "No activity in this window") {
		t.Errorf("expected empty-state line, got: %s", out)
	}
}

func TestRender_Default_WithEvents_GroupsByRepo(t *testing.T) {
	events := []github.Event{
		{
			Type:      "PullRequestEvent",
			Repo:      github.Repo{Name: "a/old"},
			Payload:   json.RawMessage(`{"action":"opened","number":1,"pull_request":{"title":"Old PR","html_url":"https://example/1"}}`),
			CreatedAt: time.Date(2026, 5, 12, 10, 0, 0, 0, time.UTC),
		},
		{
			Type:      "IssuesEvent",
			Repo:      github.Repo{Name: "b/new"},
			Payload:   json.RawMessage(`{"action":"opened","issue":{"number":2,"title":"New issue","html_url":"https://example/2"}}`),
			CreatedAt: time.Date(2026, 5, 17, 10, 0, 0, 0, time.UTC),
		},
	}
	data := Build("lmorchard",
		time.Date(2026, 5, 11, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 5, 18, 0, 0, 0, 0, time.UTC),
		events)

	if len(data.RepoOrder) != 2 || data.RepoOrder[0] != "b/new" {
		t.Errorf("RepoOrder should be most-recent-first, got %v", data.RepoOrder)
	}

	var buf bytes.Buffer
	if err := Render(&buf, "", data); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "Old PR") || !strings.Contains(out, "New issue") {
		t.Errorf("output missing event summaries: %s", out)
	}
	// b/new (newer) should appear before a/old
	if strings.Index(out, "b/new") > strings.Index(out, "a/old") {
		t.Errorf("repo ordering wrong, more-recent repo should come first: %s", out)
	}
}
