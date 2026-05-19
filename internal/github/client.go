// Package github is a small HTTP client for the GitHub REST API.
// It only covers what github-to-markdown needs: the authenticated user
// and the user-events feed.
package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"time"
)

const (
	defaultBaseURL = "https://api.github.com"
	userAgent      = "github-to-markdown"
	apiVersion     = "2022-11-28"
	acceptHeader   = "application/vnd.github+json"
	maxEventPages  = 10
)

// Client talks to the GitHub REST API.
type Client struct {
	BaseURL string
	Token   string
	HTTP    *http.Client
}

// NewClient returns a Client that authenticates with the given token.
// An empty token is allowed (public-only access).
func NewClient(token string) *Client {
	return &Client{
		BaseURL: defaultBaseURL,
		Token:   token,
		HTTP:    &http.Client{Timeout: 30 * time.Second},
	}
}

// AuthenticatedUser returns the login of the user the token belongs to.
func (c *Client) AuthenticatedUser(ctx context.Context) (string, error) {
	var u struct {
		Login string `json:"login"`
	}
	if err := c.get(ctx, "/user", &u); err != nil {
		return "", err
	}
	if u.Login == "" {
		return "", fmt.Errorf("github: empty login returned from /user")
	}
	return u.Login, nil
}

// UserEvents fetches the events feed for `user`, paginating until events
// fall before `since` or the API runs out of pages. Events outside
// [since, until] are dropped. Results are sorted newest-first (GitHub's default).
func (c *Client) UserEvents(ctx context.Context, user string, since, until time.Time) ([]Event, error) {
	url := fmt.Sprintf("%s/users/%s/events?per_page=100", c.BaseURL, user)
	var all []Event
	for page := 0; page < maxEventPages && url != ""; page++ {
		var batch []Event
		next, err := c.getPage(ctx, url, &batch)
		if err != nil {
			return nil, err
		}
		stop := false
		for _, ev := range batch {
			if ev.CreatedAt.Before(since) {
				stop = true
				continue
			}
			if !until.IsZero() && ev.CreatedAt.After(until) {
				continue
			}
			all = append(all, ev)
		}
		if stop {
			break
		}
		url = next
	}
	return all, nil
}

func (c *Client) get(ctx context.Context, path string, out interface{}) error {
	_, err := c.getPage(ctx, c.BaseURL+path, out)
	return err
}

// getPage performs a GET on the given absolute URL, decodes into out, and
// returns the "next" page URL parsed from the Link header (empty if none).
func (c *Client) getPage(ctx context.Context, url string, out interface{}) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("github: build request: %w", err)
	}
	req.Header.Set("Accept", acceptHeader)
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("X-GitHub-Api-Version", apiVersion)
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return "", fmt.Errorf("github: request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return "", fmt.Errorf("github: %s %s: %d %s", req.Method, url, resp.StatusCode, string(body))
	}

	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return "", fmt.Errorf("github: decode response: %w", err)
		}
	}

	return parseNextLink(resp.Header.Get("Link")), nil
}

var nextLinkRE = regexp.MustCompile(`<([^>]+)>;\s*rel="next"`)

func parseNextLink(header string) string {
	if header == "" {
		return ""
	}
	m := nextLinkRE.FindStringSubmatch(header)
	if len(m) < 2 {
		return ""
	}
	return m[1]
}
