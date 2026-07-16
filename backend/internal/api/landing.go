package api

import (
	"fmt"
	"net/http"
)

// landingHandler serves a self-contained HTML overview of the Sentinel API.
func landingHandler(version string) http.HandlerFunc {
	body := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Sentinel</title>
  <style>
    body { font-family: system-ui, sans-serif; background: #0f172a; color: #e2e8f0; margin: 0; padding: 3rem 1rem; }
    .container { max-width: 720px; margin: 0 auto; }
    h1 { font-size: 1.75rem; margin-bottom: 0.25rem; }
    .version { color: #38bdf8; font-size: 0.95rem; margin-bottom: 1.5rem; }
    p.lead { color: #94a3b8; line-height: 1.6; margin-bottom: 1.5rem; }
    h2 { font-size: 1.1rem; margin-bottom: 0.5rem; color: #f1f5f9; }
    ul { list-style: none; padding: 0; }
    li { padding: 0.4rem 0; border-bottom: 1px solid #1e293b; }
    code { background: #1e293b; padding: 0.15rem 0.4rem; border-radius: 4px; font-size: 0.85rem; color: #38bdf8; }
    a { color: #38bdf8; text-decoration: none; }
    a:hover { text-decoration: underline; }
  </style>
</head>
<body>
  <div class="container">
    <h1>Sentinel Infrastructure Event Intelligence Platform</h1>
    <div class="version">Version %s</div>
    <p class="lead">Real-time infrastructure monitoring, event replay, and AI-driven incident analysis.</p>
    <h2>Available API Endpoints</h2>
    <ul>
      <li><code>GET</code>&nbsp;&nbsp;<a href="/api/v1/health">/api/v1/health</a> &mdash; Infrastructure health</li>
      <li><code>POST</code>&nbsp;<a href="/api/v1/metrics">/api/v1/metrics</a> &mdash; Ingest agent metrics</li>
      <li><code>POST</code>&nbsp;<a href="/api/v1/events">/api/v1/events</a> &mdash; Ingest agent events</li>
      <li><code>GET</code>&nbsp;&nbsp;<a href="/api/v1/dashboard/overview">/api/v1/dashboard/overview</a> &mdash; Fleet overview</li>
      <li><code>GET</code>&nbsp;&nbsp;<a href="/api/v1/dashboard/hosts">/api/v1/dashboard/hosts</a> &mdash; Host metrics</li>
      <li><code>GET</code>&nbsp;&nbsp;/api/v1/dashboard/hosts/&#123;hostname&#125;/metrics &mdash; Host history</li>
      <li><code>GET</code>&nbsp;&nbsp;/api/v1/replay/hosts/&#123;hostname&#125; &mdash; Event replay</li>
      <li><code>GET</code>&nbsp;&nbsp;/api/v1/time-machine/hosts/&#123;hostname&#125; &mdash; Point-in-time snapshot</li>
      <li><code>POST</code>&nbsp;<a href="/api/v1/ai/incidents/analyze">/api/v1/ai/incidents/analyze</a> &mdash; AI incident analysis</li>
    </ul>
  </div>
</body>
</html>
`, version)

	return func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = writer.Write([]byte(body))
	}
}
