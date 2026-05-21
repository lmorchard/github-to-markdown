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

// As of late 2025, the events API returns PushEvent payloads with only
// before / head / push_id / ref / repository_id — no size, no commits.
// We should still produce a useful summary and a compare-view link.
func TestNormalize_PushEvent_StrippedPayload(t *testing.T) {
	payload := json.RawMessage(`{
		"push_id": 12345,
		"ref": "refs/heads/main",
		"before": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"head":   "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		"repository_id": 999
	}`)
	n := Normalize(Event{Type: "PushEvent", Repo: Repo{Name: "lmorchard/foo"}, Payload: payload})
	if n.Kind != "Push" {
		t.Errorf("Kind = %q, want Push", n.Kind)
	}
	if want := "pushed to `main`"; n.Summary != want {
		t.Errorf("Summary = %q, want %q", n.Summary, want)
	}
	if want := "https://github.com/lmorchard/foo/compare/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa...bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"; n.URL != want {
		t.Errorf("URL = %q, want %q", n.URL, want)
	}
}

// First push to a brand-new branch (or force-push that obliterates
// history) has `before` set to the all-zero sentinel SHA. The compare
// URL would 404, so we fall back to a direct link to the head commit.
func TestNormalize_PushEvent_ZeroBefore(t *testing.T) {
	payload := json.RawMessage(`{
		"ref": "refs/heads/feature-x",
		"before": "0000000000000000000000000000000000000000",
		"head":   "cccccccccccccccccccccccccccccccccccccccc"
	}`)
	n := Normalize(Event{Type: "PushEvent", Repo: Repo{Name: "lmorchard/foo"}, Payload: payload})
	if want := "pushed to `feature-x`"; n.Summary != want {
		t.Errorf("Summary = %q, want %q", n.Summary, want)
	}
	if want := "https://github.com/lmorchard/foo/commit/cccccccccccccccccccccccccccccccccccccccc"; n.URL != want {
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

func TestFilterPublic(t *testing.T) {
	events := []Event{
		{ID: "1", Public: true},
		{ID: "2", Public: false},
		{ID: "3", Public: true},
		{ID: "4", Public: false},
	}
	got := FilterPublic(events)
	if len(got) != 2 {
		t.Fatalf("len(got) = %d, want 2", len(got))
	}
	if got[0].ID != "1" || got[1].ID != "3" {
		t.Errorf("FilterPublic kept wrong events: %+v", got)
	}
}

func TestFilterPublic_Empty(t *testing.T) {
	if got := FilterPublic(nil); len(got) != 0 {
		t.Errorf("nil input should yield empty result, got %v", got)
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
