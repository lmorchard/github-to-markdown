package github

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Event is the raw GitHub events-feed entry.
type Event struct {
	ID        string          `json:"id"`
	Type      string          `json:"type"`
	Actor     Actor           `json:"actor"`
	Repo      Repo            `json:"repo"`
	Payload   json.RawMessage `json:"payload"`
	Public    bool            `json:"public"`
	CreatedAt time.Time       `json:"created_at"`
}

// Actor identifies the user who triggered the event.
type Actor struct {
	Login string `json:"login"`
}

// Repo identifies the repository the event happened in.
type Repo struct {
	Name string `json:"name"` // "owner/name"
}

// Normalized is a flattened, template-friendly view of an Event.
type Normalized struct {
	Kind    string // human-readable label (e.g. "PR opened")
	Summary string // one-line description
	URL     string // best-effort link to the relevant page on github.com
	Repo    string // owner/name
	Actor   string
	At      time.Time
	Type    string // raw event type, for template authors who want it
}

// Normalize converts the raw event into a template-friendly form.
// Unknown event types still produce a usable summary.
func Normalize(ev Event) Normalized {
	n := Normalized{
		Repo:  ev.Repo.Name,
		Actor: ev.Actor.Login,
		At:    ev.CreatedAt,
		Type:  ev.Type,
	}

	switch ev.Type {
	case "PushEvent":
		// As of late 2025 the events API (/users/{user}/events and friends)
		// stopped returning size / distinct_size / commits on PushEvent
		// payloads — only ref / before / head / push_id / repository_id
		// come back. We still parse the legacy shape in case it ever
		// reappears (and for tests), but compose output that's honest
		// when those fields are missing.
		var p struct {
			Ref     string `json:"ref"`
			Before  string `json:"before"`
			Head    string `json:"head"`
			Size    int    `json:"size"`
			Commits []struct {
				SHA     string `json:"sha"`
				Message string `json:"message"`
			} `json:"commits"`
		}
		_ = json.Unmarshal(ev.Payload, &p)
		branch := strings.TrimPrefix(p.Ref, "refs/heads/")
		n.Kind = "Push"
		switch {
		case p.Size > 0 && len(p.Commits) > 0:
			// Legacy / test path: full payload available.
			n.Summary = fmt.Sprintf("pushed %d commit(s) to `%s`: %s", p.Size, branch, firstLine(p.Commits[0].Message))
			n.URL = fmt.Sprintf("https://github.com/%s/commit/%s", ev.Repo.Name, p.Commits[0].SHA)
		case isRealSHA(p.Before) && isRealSHA(p.Head):
			// Current API path: only before/head SHAs. Link to the
			// compare view so the reader can see what actually shipped.
			n.Summary = fmt.Sprintf("pushed to `%s`", branch)
			n.URL = fmt.Sprintf("https://github.com/%s/compare/%s...%s", ev.Repo.Name, p.Before, p.Head)
		case isRealSHA(p.Head):
			// First push to a branch, or before is the zero SHA. Link
			// to the head commit directly.
			n.Summary = fmt.Sprintf("pushed to `%s`", branch)
			n.URL = fmt.Sprintf("https://github.com/%s/commit/%s", ev.Repo.Name, p.Head)
		default:
			// No useful SHAs; fall back to the branch's commit listing.
			n.Summary = fmt.Sprintf("pushed to `%s`", branch)
			n.URL = fmt.Sprintf("https://github.com/%s/commits/%s", ev.Repo.Name, branch)
		}

	case "PullRequestEvent":
		var p struct {
			Action      string `json:"action"`
			Number      int    `json:"number"`
			PullRequest struct {
				Title   string `json:"title"`
				HTMLURL string `json:"html_url"`
				Merged  bool   `json:"merged"`
			} `json:"pull_request"`
		}
		_ = json.Unmarshal(ev.Payload, &p)
		action := p.Action
		if action == "closed" && p.PullRequest.Merged {
			action = "merged"
		}
		n.Kind = "PR " + action
		n.Summary = fmt.Sprintf("#%d %s", p.Number, p.PullRequest.Title)
		n.URL = p.PullRequest.HTMLURL

	case "PullRequestReviewEvent":
		var p struct {
			Action string `json:"action"`
			Review struct {
				State   string `json:"state"`
				HTMLURL string `json:"html_url"`
			} `json:"review"`
			PullRequest struct {
				Number  int    `json:"number"`
				Title   string `json:"title"`
				HTMLURL string `json:"html_url"`
			} `json:"pull_request"`
		}
		_ = json.Unmarshal(ev.Payload, &p)
		n.Kind = "PR review (" + p.Review.State + ")"
		n.Summary = fmt.Sprintf("#%d %s", p.PullRequest.Number, p.PullRequest.Title)
		if p.Review.HTMLURL != "" {
			n.URL = p.Review.HTMLURL
		} else {
			n.URL = p.PullRequest.HTMLURL
		}

	case "PullRequestReviewCommentEvent":
		var p struct {
			Comment struct {
				Body    string `json:"body"`
				HTMLURL string `json:"html_url"`
			} `json:"comment"`
			PullRequest struct {
				Number int    `json:"number"`
				Title  string `json:"title"`
			} `json:"pull_request"`
		}
		_ = json.Unmarshal(ev.Payload, &p)
		n.Kind = "PR review comment"
		n.Summary = fmt.Sprintf("#%d %s — %s", p.PullRequest.Number, p.PullRequest.Title, snippet(p.Comment.Body, 80))
		n.URL = p.Comment.HTMLURL

	case "IssuesEvent":
		var p struct {
			Action string `json:"action"`
			Issue  struct {
				Number  int    `json:"number"`
				Title   string `json:"title"`
				HTMLURL string `json:"html_url"`
			} `json:"issue"`
		}
		_ = json.Unmarshal(ev.Payload, &p)
		n.Kind = "Issue " + p.Action
		n.Summary = fmt.Sprintf("#%d %s", p.Issue.Number, p.Issue.Title)
		n.URL = p.Issue.HTMLURL

	case "IssueCommentEvent":
		var p struct {
			Action string `json:"action"`
			Issue  struct {
				Number      int    `json:"number"`
				Title       string `json:"title"`
				PullRequest *struct {
					HTMLURL string `json:"html_url"`
				} `json:"pull_request"`
			} `json:"issue"`
			Comment struct {
				Body    string `json:"body"`
				HTMLURL string `json:"html_url"`
			} `json:"comment"`
		}
		_ = json.Unmarshal(ev.Payload, &p)
		kind := "Issue comment"
		if p.Issue.PullRequest != nil {
			kind = "PR comment"
		}
		n.Kind = kind
		n.Summary = fmt.Sprintf("#%d %s — %s", p.Issue.Number, p.Issue.Title, snippet(p.Comment.Body, 80))
		n.URL = p.Comment.HTMLURL

	case "CreateEvent":
		var p struct {
			RefType string `json:"ref_type"`
			Ref     string `json:"ref"`
		}
		_ = json.Unmarshal(ev.Payload, &p)
		n.Kind = "Create " + p.RefType
		if p.Ref != "" {
			n.Summary = p.Ref
		} else {
			n.Summary = ev.Repo.Name
		}
		n.URL = "https://github.com/" + ev.Repo.Name

	case "DeleteEvent":
		var p struct {
			RefType string `json:"ref_type"`
			Ref     string `json:"ref"`
		}
		_ = json.Unmarshal(ev.Payload, &p)
		n.Kind = "Delete " + p.RefType
		n.Summary = p.Ref
		n.URL = "https://github.com/" + ev.Repo.Name

	case "ForkEvent":
		var p struct {
			Forkee struct {
				FullName string `json:"full_name"`
				HTMLURL  string `json:"html_url"`
			} `json:"forkee"`
		}
		_ = json.Unmarshal(ev.Payload, &p)
		n.Kind = "Fork"
		n.Summary = "forked to " + p.Forkee.FullName
		n.URL = p.Forkee.HTMLURL

	case "WatchEvent":
		n.Kind = "Star"
		n.Summary = "starred " + ev.Repo.Name
		n.URL = "https://github.com/" + ev.Repo.Name

	case "ReleaseEvent":
		var p struct {
			Action  string `json:"action"`
			Release struct {
				Name    string `json:"name"`
				TagName string `json:"tag_name"`
				HTMLURL string `json:"html_url"`
			} `json:"release"`
		}
		_ = json.Unmarshal(ev.Payload, &p)
		name := p.Release.Name
		if name == "" {
			name = p.Release.TagName
		}
		n.Kind = "Release " + p.Action
		n.Summary = name
		n.URL = p.Release.HTMLURL

	case "PublicEvent":
		n.Kind = "Made public"
		n.Summary = ev.Repo.Name
		n.URL = "https://github.com/" + ev.Repo.Name

	case "GollumEvent":
		var p struct {
			Pages []struct {
				PageName string `json:"page_name"`
				Action   string `json:"action"`
				HTMLURL  string `json:"html_url"`
			} `json:"pages"`
		}
		_ = json.Unmarshal(ev.Payload, &p)
		n.Kind = "Wiki"
		if len(p.Pages) > 0 {
			n.Summary = fmt.Sprintf("%s %s", p.Pages[0].Action, p.Pages[0].PageName)
			n.URL = p.Pages[0].HTMLURL
		} else {
			n.Summary = ev.Repo.Name
			n.URL = "https://github.com/" + ev.Repo.Name + "/wiki"
		}

	default:
		n.Kind = ev.Type
		n.Summary = ev.Repo.Name
		n.URL = "https://github.com/" + ev.Repo.Name
	}

	return n
}

// FilterPublic returns only the public events from the input slice.
// The events-feed endpoint mixes public and private-repo events whenever
// the token can see them; this drops the private ones.
func FilterPublic(events []Event) []Event {
	out := events[:0:0]
	for _, ev := range events {
		if ev.Public {
			out = append(out, ev)
		}
	}
	return out
}

func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}

// isRealSHA returns true for a 40-char hex SHA that isn't the all-zero
// sentinel git uses for "no parent" (initial branch push, force-push to
// delete commits, etc.).
func isRealSHA(s string) bool {
	if len(s) != 40 {
		return false
	}
	for _, c := range s {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			return false
		}
	}
	return s != "0000000000000000000000000000000000000000"
}

func snippet(s string, max int) string {
	s = strings.TrimSpace(firstLine(s))
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}
