package render

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/lmorchard/github-to-markdown/internal/github"
)

// TestSmokeRender prints a realistic rendered document to stdout when run with -v.
// Run: go test -v -run TestSmokeRender ./internal/render
func TestSmokeRender(t *testing.T) {
	events := []github.Event{
		{
			Type:      "PullRequestEvent",
			Repo:      github.Repo{Name: "lmorchard/foo"},
			Payload:   json.RawMessage(`{"action":"opened","number":7,"pull_request":{"title":"Add markdown rendering","html_url":"https://github.com/lmorchard/foo/pull/7","merged":false}}`),
			CreatedAt: time.Date(2026, 5, 17, 14, 30, 0, 0, time.UTC),
		},
		{
			Type:      "PushEvent",
			Repo:      github.Repo{Name: "lmorchard/foo"},
			Payload:   json.RawMessage(`{"ref":"refs/heads/main","size":1,"commits":[{"sha":"abc1234","message":"Wire up fetch command\n\ndetails"}]}`),
			CreatedAt: time.Date(2026, 5, 18, 9, 15, 0, 0, time.UTC),
		},
		{
			Type:      "IssueCommentEvent",
			Repo:      github.Repo{Name: "someorg/bar"},
			Payload:   json.RawMessage(`{"action":"created","issue":{"number":99,"title":"Bug in widget"},"comment":{"body":"I can reproduce this on linux too.","html_url":"https://github.com/someorg/bar/issues/99#c1"}}`),
			CreatedAt: time.Date(2026, 5, 16, 11, 0, 0, 0, time.UTC),
		},
	}
	data := Build("lmorchard",
		time.Date(2026, 5, 11, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 5, 18, 23, 59, 59, 0, time.UTC),
		events)
	if err := Render(os.Stdout, "", data); err != nil {
		t.Fatal(err)
	}
}
