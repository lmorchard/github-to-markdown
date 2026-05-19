# github-to-markdown

A small Go CLI that fetches your recent GitHub activity from the events feed
and renders it as a Markdown document. Useful for weeknotes, journals, and
LLM-friendly activity summaries.

## What it does

Calls GitHub's `/users/{user}/events` endpoint, filters to a time window,
groups events by repository (most-recent first), and renders the result
through a template. By default it writes to stdout; you can also write to a
file and customize the template.

## Installation

From source (requires Go 1.21+):

```sh
go install github.com/lmorchard/github-to-markdown@latest
```

Or clone and build:

```sh
git clone https://github.com/lmorchard/github-to-markdown.git
cd github-to-markdown
make build
./github-to-markdown --help
```

Pre-built binaries are produced by the release workflow for Linux, macOS, and
Windows on tagged versions and as a rolling "latest" build from `main`.

## Quick start

```sh
export GITHUB_TOKEN=ghp_xxxxxxxxxxxxxxxxxxxx
github-to-markdown --since 2026-05-11 --output activity.md
```

That fetches activity from May 11 through now, drops events from private
repositories, and writes Markdown to `activity.md`.

## Authentication

The tool reads a GitHub token from the `--token` flag or the `GITHUB_TOKEN`
environment variable.

### Token scopes

The tool hits two endpoints: `GET /user` and `GET /users/{user}/events`.

**Classic personal access token:**

- No scopes are strictly required for public activity. `/user` works without
  any scope.
- `repo` is required to include private-repo activity in the events feed.
  GitHub returns private events only when authenticated as that user *and*
  the token can see the repo.

**Fine-grained personal access token:**

- Resource owner: yourself
- Repository access: "All repositories" for full private activity, or
  "Selected repositories" for a subset
- Permissions: **Metadata: Read-only** is sufficient

Pragmatic recommendation: classic PAT with the `repo` scope.

## Usage

```
github-to-markdown [flags]
github-to-markdown export --since <date|duration> [--until <date>] [-o <file>]
github-to-markdown init
github-to-markdown version
```

The `export` subcommand exposes the canonical `--since/--until/-o` flag
shape used by [`me-to-markdown`](https://github.com/lmorchard/me-to-markdown)
and the rest of the `*-to-markdown` family. It's a thin wrapper over the
same pipeline as the root command — `--since` additionally accepts a Go
duration (e.g. `168h`) on `export`, and other settings (token, user,
template, include-private) come from the config file or environment.

### Flags

| Flag | Env | Description | Default |
|---|---|---|---|
| `--token` | `GITHUB_TOKEN` | GitHub personal access token | (required) |
| `--user` | | GitHub username to fetch | authenticated user |
| `--since` | | Start of window (`YYYY-MM-DD` or RFC3339) | 7 days ago |
| `--until` | | End of window (`YYYY-MM-DD` or RFC3339) | now |
| `-o`, `--output` | | Write output to a file | stdout |
| `--template` | | Path to a custom template file | embedded default |
| `--include-private` | | Include events from private repositories | false |
| `--config` | | Path to a config file | `./github-to-markdown.yaml` |
| `-v`, `--verbose` | | Verbose logging | false |
| `--debug` | | Debug logging | false |
| `--log-json` | | JSON log format | false |

`YYYY-MM-DD` dates are parsed in the local timezone. `--until` interpreted as
a date is treated as end-of-day, so `--since 2026-05-11 --until 2026-05-18`
gives an inclusive 8-day window.

### Examples

Last 7 days to stdout:

```sh
github-to-markdown
```

Specific week, to file, including private activity:

```sh
github-to-markdown --since 2026-05-11 --until 2026-05-18 \
  --include-private -o week-of-may-11.md
```

Activity for a different user (public events only — token still required for
the `/user` lookup):

```sh
github-to-markdown --user some-other-user --since 2026-05-01
```

## Configuration file

Settings can also live in a YAML config file. By default the tool looks for
`github-to-markdown.yaml` in the current directory; override with `--config`.

CLI flags override config values; env vars override file values.

```yaml
# Prefer GITHUB_TOKEN over storing this in the file.
# token: "ghp_xxxxxxxxxxxxxxxxxxxx"

# GitHub username to fetch. Defaults to the authenticated user.
# user: "lmorchard"

# Time window. YYYY-MM-DD or RFC3339. Defaults: since = now-7d, until = now.
# since: "2026-05-11"
# until: "2026-05-18"

# Output file. Empty / unset means stdout.
# output: "activity.md"

# Path to a custom template file. Empty / unset means the embedded default.
# template: "github-to-markdown.md"

# Include events from private repositories.
# include_private: false

verbose: false
debug: false
log_json: false
```

Run `github-to-markdown init` to generate both a config file and a starter
template you can edit.

## Custom templates

Templates use Go's [`text/template`](https://pkg.go.dev/text/template) syntax.
The data passed in:

```go
type Data struct {
    User      string              // GitHub login
    Since     time.Time           // start of window
    Until     time.Time           // end of window
    Generated time.Time           // when the report was generated
    Events    []Normalized        // flat list, newest-first
    ByRepo    map[string][]Normalized
    RepoOrder []string            // repo names sorted by most-recent event
}

type Normalized struct {
    Kind    string    // human label ("PR opened", "Push", "Issue closed", ...)
    Summary string    // one-line description
    URL     string    // best-effort link to the relevant page
    Repo    string    // owner/name
    Actor   string    // login of the user who triggered the event
    At      time.Time // event timestamp
    Type    string    // raw GitHub event type (PullRequestEvent, PushEvent, etc.)
}
```

The embedded default template is in `internal/templates/default.md` and is a
good starting point.

Useful patterns:

```
{{ range $repo := .RepoOrder }}
## {{ $repo }}
{{ range index $.ByRepo $repo -}}
- {{ .Kind }}: {{ .Summary }}
{{ end -}}
{{ end -}}
```

## Limitations

GitHub's events feed is rate- and size-limited by the server:

- The feed returns **at most ~300 events / 90 days**, regardless of token
  scopes. Asking for a longer `--since` window won't backfill earlier
  activity.
- **Payloads are heavily stripped** on this endpoint. Pull-request events
  don't include the title or HTML URL; push events don't include the commit
  list or count. The tool synthesizes what it can from the data that is
  available; for richer detail, an API call per PR would be needed.
- The feed only includes events the token's user can see. Private-repo
  events appear when the token has access; the `--include-private` flag
  controls whether they're rendered.

## Development

```sh
make setup    # install gofumpt and golangci-lint
make build    # build binary with version metadata
make test     # go test ./...
make lint     # golangci-lint run
make format   # go fmt + gofumpt
```

### Project layout

```
cmd/                  Cobra commands
internal/config/      Config struct
internal/github/      Stdlib HTTP client + event normalization
internal/render/      Template data builder + renderer
internal/templates/   Embedded default markdown template
```

## License

MIT.
