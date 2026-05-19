package github

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNormalize_PushEvent(t *testing.T) {
	payload := json.RawMessage(`{
		"ref": "refs/heads/main",
		"size": 2,
		"commits": [
			{"sha": "abc123", "message": "Fix the bug\n\nDetails here"},
			{"sha": "def456", "message": "Another commit"}
		]
	}`)
	ev := Event{
		Type:      "PushEvent",
		Repo:      Repo{Name: "lmorchard/foo"},
		Payload:   payload,
		CreatedAt: time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
	}

	n := Normalize(ev)
	if n.Kind != "Push" {
		t.Errorf("Kind = %q, want Push", n.Kind)
	}
	if want := "pushed 2 commit(s) to `main`: Fix the bug"; n.Summary != want {
		t.Errorf("Summary = %q, want %q", n.Summary, want)
	}
	if want := "https://github.com/lmorchard/foo/commit/abc123"; n.URL != want {
		t.Errorf("URL = %q, want %q", n.URL, want)
	}
}

func TestNormalize_PullRequestMerged(t *testing.T) {
	payload := json.RawMessage(`{
		"action": "closed",
		"number": 42,
		"pull_request": {
			"title": "Add feature X",
			"html_url": "https://github.com/lmorchard/foo/pull/42",
			"merged": true
		}
	}`)
	n := Normalize(Event{Type: "PullRequestEvent", Repo: Repo{Name: "lmorchard/foo"}, Payload: payload})
	if n.Kind != "PR merged" {
		t.Errorf("Kind = %q, want PR merged", n.Kind)
	}
	if n.Summary != "#42 Add feature X" {
		t.Errorf("Summary = %q", n.Summary)
	}
}

func TestNormalize_UnknownType(t *testing.T) {
	n := Normalize(Event{Type: "MysteryEvent", Repo: Repo{Name: "lmorchard/foo"}})
	if n.Kind != "MysteryEvent" {
		t.Errorf("Kind = %q, want MysteryEvent", n.Kind)
	}
	if n.URL != "https://github.com/lmorchard/foo" {
		t.Errorf("URL = %q", n.URL)
	}
}

func TestParseNextLink(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"", ""},
		{`<https://api.github.com/x?page=2>; rel="next", <https://api.github.com/x?page=5>; rel="last"`, "https://api.github.com/x?page=2"},
		{`<https://api.github.com/x?page=1>; rel="prev"`, ""},
	}
	for _, c := range cases {
		got := parseNextLink(c.in)
		if got != c.want {
			t.Errorf("parseNextLink(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
