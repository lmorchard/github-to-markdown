# Session Notes — Initial Scaffold

## What got built

A working Go CLI that fetches the authenticated user's recent GitHub activity
and renders it to Markdown. Single command (`github-to-markdown`) with `init`
and `version` subcommands.

## Architecture

```
main.go                       entry point
cmd/
  root.go                     cobra + viper setup; root.RunE = runFetch
  fetch.go                    runFetch: token check → client → fetch → render
  init.go                     writes example yaml + customizable template
  version.go, constants.go    scaffolding
internal/
  config/config.go            Config struct (Token/User/Since/Until/Output/Template)
  github/
    client.go                 minimal HTTP client (stdlib only), Link-header pagination
    events.go                 raw Event type + Normalize(ev) → Normalized
  render/
    render.go                 Build(data) + Render(w, tmplPath, data) using text/template
  templates/
    templates.go              go:embed default.md
    default.md                Markdown template grouped by repo, recency-ordered
```

No DB. No 3rd-party GitHub SDK — only stdlib for the API.

## Decisions

- **No `go-github`**: 2 endpoints don't justify the dep tree. Stdlib `net/http` +
  one regex for Link-header `rel="next"`.
- **Root command does the work** (vs. a `fetch` subcommand): single-purpose tool,
  cleanest UX is `github-to-markdown --since YYYY-MM-DD`.
- **`text/template` not `html/template`**: output is markdown; no HTML-escaping needed.
- **Recency-ordered repo grouping** via a `RepoOrder` slice on the template data,
  since Go's template `range` over a map sorts by key alphabetically.
- **Date parsing accepts YYYY-MM-DD or RFC3339**; `--until 2026-05-18` is interpreted
  as end-of-day so the inclusive window behaves intuitively.

## Bugs found in the scaffolder

The `go-cli-builder` scaffold's `cmd/init.go.template` and `internal/templates/default.md`
weren't getting their `{{.ProjectName}}` / `{{"{{"}}` placeholders substituted —
generated files still had the raw template syntax. Fixed locally; worth filing
upstream in the skill repo.

## Verified

- `go build ./...` clean
- `make format && make lint` clean (0 issues)
- `go test ./...` clean (3 packages, ~10 tests)
- `--help` output looks right
- `TestSmokeRender` in `internal/render` produces sensible markdown
- No-token error path returns a clear message without dumping Cobra usage

## Not done / follow-ups

- Haven't run against a live token end-to-end (Les can do that — `GITHUB_TOKEN=... ./github-to-markdown -v`)
- No PR/issue/commit search APIs — events feed only (sufficient for v1)
- No event-type filtering / per-repo filtering flags
- README hasn't been written
- The events endpoint caps at ~300 events / 90 days; no warning when the window
  exceeds what the API can return
- Could add `gh auth token` fallback for token discovery
