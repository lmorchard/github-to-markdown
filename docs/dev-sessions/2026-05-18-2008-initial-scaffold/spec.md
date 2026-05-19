# github-to-markdown — Initial Scaffold

## Purpose

A small Go CLI that fetches the user's recent GitHub activity and renders it as a
markdown document. Intended for weeknotes-style summaries, journaling, and
LLM-friendly activity reports.

## Scope (this session)

- Project scaffolding via `go-cli-builder` skill (no database, templates enabled)
- Single primary command (`fetch` or default-run) that:
  - Authenticates to GitHub via `GITHUB_TOKEN` env var (or `--token` flag)
  - Identifies the user via `/user` endpoint (or `--user` flag override)
  - Fetches recent activity from `/users/{user}/events`
  - Filters to the requested time window (`--since` / `--until`, default last 7 days)
  - Renders to markdown via embedded template (overridable with `--template`)
  - Writes to stdout (or `--output FILE`)
- `init` command — generates example config + customizable template
- `version` command (from scaffold)

## Out of Scope

- PR / issue / commit search APIs (events feed only — broad enough for v1)
- Persistence / caching across runs (no DB)
- Per-repo filtering, event-type filtering (can be added later)
- Pagination beyond the default events page (events endpoint caps at ~300 events / 90 days anyway)

## Configuration

| Key | Flag | Env | Default |
|---|---|---|---|
| `token` | `--token` | `GITHUB_TOKEN` | (required) |
| `user` | `--user` | — | authenticated user from `/user` |
| `since` | `--since` | — | now - 7 days |
| `until` | `--until` | — | now |
| `output` | `--output`, `-o` | — | stdout |
| `template` | `--template` | — | embedded default |

## Output Shape (markdown)

Header with date range, then events grouped by repo. Each event rendered as a
short line with a link. Template is overridable, so exact shape is user-controlled
after `init`.

## Success Criteria

- `go build` succeeds
- `make lint` / `make test` clean
- `github-to-markdown --since 2026-05-11` against a real token produces sensible markdown on stdout
- `github-to-markdown init` produces a config + template file
