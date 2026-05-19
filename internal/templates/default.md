# GitHub Activity — {{ .User }}

_{{ .Since.Format "2006-01-02" }} to {{ .Until.Format "2006-01-02" }} · {{ len .Events }} event(s) · generated {{ .Generated.Format "2006-01-02 15:04 MST" }}_

{{ if not .Events -}}
_No activity in this window._
{{ else -}}
{{ range $repo := .RepoOrder }}
## [{{ $repo }}](https://github.com/{{ $repo }})

{{ range index $.ByRepo $repo -}}
- **{{ .Kind }}** — {{ .Summary }} _{{ .At.Format "2006-01-02 15:04" }}_{{ if .URL }} · [link]({{ .URL }}){{ end }}
{{ end }}
{{ end -}}
{{ end -}}
