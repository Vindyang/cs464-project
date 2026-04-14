package app

import (
	"bytes"
	"html/template"
	"net/http"
	"strings"
)

type docsRoute struct {
	Group   string
	Method  string
	Path    string
	Summary string
}

type docsEndpoint struct {
	Method      string
	MethodClass string
	Path        string
	Summary     string
}

type docsSection struct {
	Title     string
	Endpoints []docsEndpoint
}

type docsPageData struct {
	Sections []docsSection
	SpecPath string
	Total    int
}

var documentedRoutes = []docsRoute{
	{Group: "System", Method: http.MethodGet, Path: "/health", Summary: "Service health check."},
	{Group: "System", Method: http.MethodGet, Path: "/api/v1/health", Summary: "Versioned service health check."},
	{Group: "Docs", Method: http.MethodGet, Path: "/api/v1/docs", Summary: "Interactive index of available backend endpoints."},
	{Group: "Docs", Method: http.MethodGet, Path: "/api/v1/docs/openapi.yml", Summary: "Raw OpenAPI YAML specification."},

	{Group: "Providers", Method: http.MethodGet, Path: "/api/providers", Summary: "List currently connected providers."},
	{Group: "Providers", Method: http.MethodGet, Path: "/api/v1/providers", Summary: "Versioned alias for listing providers."},
	{Group: "Providers", Method: http.MethodGet, Path: "/api/oauth/gdrive/authorize", Summary: "Start Google Drive OAuth authorization."},
	{Group: "Providers", Method: http.MethodGet, Path: "/api/oauth/gdrive/callback", Summary: "Complete Google Drive OAuth authorization."},
	{Group: "Providers", Method: http.MethodPost, Path: "/api/oauth/gdrive/disconnect", Summary: "Disconnect the Google Drive provider."},
	{Group: "Providers", Method: http.MethodGet, Path: "/api/oauth/onedrive/authorize", Summary: "Start OneDrive OAuth authorization."},
	{Group: "Providers", Method: http.MethodGet, Path: "/api/oauth/onedrive/callback", Summary: "Complete OneDrive OAuth authorization."},
	{Group: "Providers", Method: http.MethodPost, Path: "/api/oauth/onedrive/disconnect", Summary: "Disconnect the OneDrive provider."},
	{Group: "Providers", Method: http.MethodPost, Path: "/api/providers/awsS3/connect", Summary: "Connect the AWS S3 provider."},
	{Group: "Providers", Method: http.MethodPost, Path: "/api/providers/awsS3/disconnect", Summary: "Disconnect the AWS S3 provider."},

	{Group: "Credentials", Method: http.MethodGet, Path: "/api/credentials", Summary: "List stored provider credentials."},
	{Group: "Credentials", Method: http.MethodGet, Path: "/api/credentials/status", Summary: "Show whether any provider credentials are configured."},
	{Group: "Credentials", Method: http.MethodGet, Path: "/api/credentials/{provider}", Summary: "Read a provider credential record without the secret."},
	{Group: "Credentials", Method: http.MethodPut, Path: "/api/credentials/{provider}", Summary: "Create or update provider credentials."},
	{Group: "Credentials", Method: http.MethodDelete, Path: "/api/credentials/{provider}", Summary: "Delete a provider credential record."},
	{Group: "Credentials", Method: http.MethodGet, Path: "/api/credentials/{provider}/secret", Summary: "Reveal the stored provider secret."},

	{Group: "Settings", Method: http.MethodGet, Path: "/api/settings", Summary: "Read current application settings."},
	{Group: "Settings", Method: http.MethodPut, Path: "/api/settings", Summary: "Update application settings."},
	{Group: "Settings", Method: http.MethodPost, Path: "/api/settings/reset", Summary: "Reset application settings and stored credentials."},

	{Group: "Files", Method: http.MethodGet, Path: "/api/v1/files", Summary: "List uploaded files."},
	{Group: "Files", Method: http.MethodGet, Path: "/api/v1/files/{fileId}", Summary: "Read file metadata."},
	{Group: "Files", Method: http.MethodDelete, Path: "/api/v1/files/{fileId}", Summary: "Delete a file via the versioned workflow route."},
	{Group: "Files", Method: http.MethodDelete, Path: "/api/orchestrator/files/{fileId}", Summary: "Delete a file via the compatibility orchestrator route."},
	{Group: "Files", Method: http.MethodGet, Path: "/api/v1/shards/file/{fileId}", Summary: "Read shard map details for a file."},
	{Group: "Files", Method: http.MethodPost, Path: "/api/v1/files/health/refresh", Summary: "Refresh health for all files."},
	{Group: "Files", Method: http.MethodPost, Path: "/api/orchestrator/files/health/refresh", Summary: "Compatibility route for refreshing all file health."},
	{Group: "Files", Method: http.MethodPost, Path: "/api/v1/files/{fileId}/health/refresh", Summary: "Refresh health for one file."},
	{Group: "Files", Method: http.MethodPost, Path: "/api/orchestrator/files/{fileId}/health/refresh", Summary: "Compatibility route for refreshing one file's health."},

	{Group: "Workflows", Method: http.MethodPost, Path: "/api/v1/upload", Summary: "Upload and shard a file."},
	{Group: "Workflows", Method: http.MethodPost, Path: "/api/orchestrator/upload", Summary: "Compatibility route for uploads."},
	{Group: "Workflows", Method: http.MethodGet, Path: "/api/v1/download/{fileId}", Summary: "Download and reconstruct a file."},
	{Group: "Workflows", Method: http.MethodGet, Path: "/api/orchestrator/files/{fileId}/download", Summary: "Compatibility route for downloads."},
	{Group: "Workflows", Method: http.MethodGet, Path: "/api/v1/history", Summary: "List global lifecycle history."},
	{Group: "Workflows", Method: http.MethodGet, Path: "/api/orchestrator/history", Summary: "Compatibility route for global lifecycle history."},
	{Group: "Workflows", Method: http.MethodGet, Path: "/api/v1/history/{fileId}", Summary: "List lifecycle history for one file."},
	{Group: "Workflows", Method: http.MethodGet, Path: "/api/orchestrator/files/{fileId}/history", Summary: "Compatibility route for file lifecycle history."},
}

var docsPageTemplate = template.Must(template.New("docs-page").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Omnishard Monolith API</title>
  <style>
    :root {
      color-scheme: light;
      --bg: #f7f4ed;
      --panel: #fffdf8;
      --text: #1f2937;
      --muted: #5b6472;
      --line: #ddd4c4;
      --accent: #0f766e;
      --get: #0f766e;
      --post: #b45309;
      --put: #1d4ed8;
      --delete: #b91c1c;
    }

    * { box-sizing: border-box; }
    body {
      margin: 0;
      font-family: "Segoe UI", "Helvetica Neue", sans-serif;
      background: radial-gradient(circle at top, #fff8e8 0%, var(--bg) 45%, #f1eee7 100%);
      color: var(--text);
    }
    main {
      max-width: 1100px;
      margin: 0 auto;
      padding: 32px 20px 48px;
    }
    header {
      background: rgba(255, 253, 248, 0.88);
      border: 1px solid var(--line);
      border-radius: 20px;
      padding: 24px;
      box-shadow: 0 18px 48px rgba(80, 64, 32, 0.08);
      backdrop-filter: blur(8px);
    }
    h1 {
      margin: 0 0 8px;
      font-size: clamp(2rem, 3.5vw, 3rem);
      line-height: 1.05;
    }
    p {
      margin: 0;
      color: var(--muted);
      line-height: 1.6;
      max-width: 70ch;
    }
    .meta {
      display: flex;
      gap: 16px;
      flex-wrap: wrap;
      margin-top: 16px;
      align-items: center;
    }
    .pill {
      display: inline-flex;
      align-items: center;
      padding: 8px 12px;
      border-radius: 999px;
      border: 1px solid var(--line);
      background: #fff8ec;
      font-size: 0.95rem;
      color: var(--text);
    }
    .spec-link {
      color: var(--accent);
      text-decoration: none;
      font-weight: 600;
    }
    .spec-link:hover { text-decoration: underline; }
    section {
      margin-top: 22px;
      background: rgba(255, 253, 248, 0.92);
      border: 1px solid var(--line);
      border-radius: 18px;
      overflow: hidden;
      box-shadow: 0 14px 36px rgba(80, 64, 32, 0.06);
    }
    h2 {
      margin: 0;
      padding: 16px 20px;
      border-bottom: 1px solid var(--line);
      background: linear-gradient(90deg, rgba(15, 118, 110, 0.08), rgba(255, 248, 236, 0));
      font-size: 1.1rem;
    }
    table {
      width: 100%;
      border-collapse: collapse;
    }
    th, td {
      padding: 14px 20px;
      text-align: left;
      vertical-align: top;
      border-bottom: 1px solid rgba(221, 212, 196, 0.75);
    }
    tr:last-child td { border-bottom: none; }
    th {
      color: var(--muted);
      font-size: 0.78rem;
      letter-spacing: 0.08em;
      text-transform: uppercase;
    }
    code {
      font-family: "Cascadia Code", "Consolas", monospace;
      font-size: 0.92rem;
      word-break: break-word;
    }
    .method {
      display: inline-block;
      min-width: 72px;
      padding: 6px 10px;
      border-radius: 999px;
      font-size: 0.8rem;
      font-weight: 700;
      letter-spacing: 0.06em;
      text-align: center;
      color: #fff;
    }
    .method-get { background: var(--get); }
    .method-post { background: var(--post); }
    .method-put { background: var(--put); }
    .method-delete { background: var(--delete); }
    @media (max-width: 760px) {
      th:nth-child(3), td:nth-child(3) {
        display: none;
      }
      main {
        padding: 20px 12px 32px;
      }
      header, section {
        border-radius: 16px;
      }
      th, td, h2 {
        padding-left: 14px;
        padding-right: 14px;
      }
    }
  </style>
</head>
<body>
  <main>
    <header>
      <h1>Omnishard Monolith API</h1>
      <p>This page lists every HTTP endpoint currently exposed by the monolith backend. The raw OpenAPI document is still available if you need the machine-readable spec.</p>
      <div class="meta">
        <span class="pill">{{.Total}} routes</span>
        <a class="spec-link" href="{{.SpecPath}}">Open raw OpenAPI YAML</a>
      </div>
    </header>
    {{range .Sections}}
    <section>
      <h2>{{.Title}}</h2>
      <table>
        <thead>
          <tr>
            <th>Method</th>
            <th>Path</th>
            <th>Summary</th>
          </tr>
        </thead>
        <tbody>
          {{range .Endpoints}}
          <tr>
            <td><span class="method method-{{.MethodClass}}">{{.Method}}</span></td>
            <td><code>{{.Path}}</code></td>
            <td>{{.Summary}}</td>
          </tr>
          {{end}}
        </tbody>
      </table>
    </section>
    {{end}}
  </main>
</body>
</html>`))

func buildDocsPageData(specPath string) docsPageData {
	sections := make([]docsSection, 0)
	sectionIndex := make(map[string]int)
	total := 0

	for _, route := range documentedRoutes {
		index, ok := sectionIndex[route.Group]
		if !ok {
			index = len(sections)
			sectionIndex[route.Group] = index
			sections = append(sections, docsSection{Title: route.Group})
		}

		sections[index].Endpoints = append(sections[index].Endpoints, docsEndpoint{
			Method:      route.Method,
			MethodClass: strings.ToLower(route.Method),
			Path:        route.Path,
			Summary:     route.Summary,
		})
		total++
	}

	return docsPageData{
		Sections: sections,
		SpecPath: specPath,
		Total:    total,
	}
}

func (a *App) serveDocsIndex(w http.ResponseWriter) {
	var body bytes.Buffer
	if err := docsPageTemplate.Execute(&body, buildDocsPageData("/api/v1/docs/openapi.yml")); err != nil {
		http.Error(w, "failed to render docs page", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(body.Bytes())
}

func (a *App) serveOpenAPISpec(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/yaml; charset=utf-8")
	w.Header().Set("Content-Disposition", "inline; filename=\"openapi.yml\"")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(openAPISpec)
}
