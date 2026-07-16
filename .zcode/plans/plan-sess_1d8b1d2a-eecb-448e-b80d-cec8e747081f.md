## Sentinel Refactor Plan — Comprehensive

All 11 goals addressed. 8 new files, 13 modified files, 1 new test file. Zero architecture changes — same packages, same patterns, same DI style.

---

### GOAL 1 + 7: Refactor `main.go` → `internal/server/` with dependency builder

**NEW `internal/server/dependencies.go`** (~120 lines)
- `Dependencies` struct bundling all 7 handlers
- `buildDependencies(ctx, cfg, db, redis, logger) (*Dependencies, error)` — factory function that wires all repositories → services → handlers in order. Replaces the constructor noise in main.go. Each error is wrapped with context (`fmt.Errorf("create X: %w", err)`). Handles AI-enabled/disabled branching.

**NEW `internal/server/http.go`** (~30 lines)
- `buildHTTPServer()` method on `*App` — constructs `api.NewRouter(api.Handlers{...}, logger)` and wraps it in an `*http.Server` with the same timeouts (ReadHeader 5s, Read 15s, Write 15s, Idle 60s, MaxHeaderBytes 1MiB).

**NEW `internal/server/app.go`** (~100 lines)
- `App` struct holding `cfg`, `logger`, `db`, `redis`, `deps`
- `New(cfg *config.Config) (*App, error)` — validates config, creates logger, connects PostgreSQL + applies migrations, connects Redis, builds dependencies. Cleans up on partial failure (closes Redis/DB if a later step fails).
- `Run() error` — builds HTTP server, starts listening in goroutine, signal handling (`SIGINT`/`SIGTERM`), graceful shutdown (20s budget), closes Redis + DB. Returns error instead of `log.Fatal`.

**MODIFY `cmd/server/main.go`** (~20 lines, down from 174)
```go
func main() {
    cfg, err := config.Load()
    if err != nil { log.Fatal(err) }
    app, err := server.New(cfg)
    if err != nil { log.Fatal(err) }
    if err := app.Run(); err != nil { log.Fatal(err) }
}
```

---

### GOAL 2: Standardize ALL routes to `/api/v1/`

**MODIFY `internal/api/router.go`**

Old → New route mapping:
| Old | New |
|---|---|
| `GET /health` | `GET /api/v1/health` |
| `POST /v1/metrics` | `POST /api/v1/metrics` |
| `POST /api/v1/events` | `POST /api/v1/events` *(unchanged)* |
| `GET /v1/dashboard/overview` | `GET /api/v1/dashboard/overview` |
| `GET /v1/dashboard/hosts` | `GET /api/v1/dashboard/hosts` |
| `GET /v1/dashboard/hosts/{hostname}/metrics` | `GET /api/v1/dashboard/hosts/{hostname}/metrics` |
| `GET /v1/replay/hosts/{hostname}` | `GET /api/v1/replay/hosts/{hostname}` |
| `GET /v1/time-machine/hosts/{hostname}` | `GET /api/v1/time-machine/hosts/{hostname}` |
| `POST /v1/ai/incidents/analyze` | `POST /api/v1/ai/incidents/analyze` |

Also introduces `api.Handlers` struct (bundles 7 handlers) replacing the 7-param `NewRouter` signature → `NewRouter(handlers Handlers, logger *zap.Logger)`. Adds `apiVersion = "0.1.0"` constant.

**Test path updates** (route strings in local chi routers):
- `dashboard_test.go`: `/v1/dashboard/...` → `/api/v1/dashboard/...`
- `replay_test.go`: `/v1/replay/...` → `/api/v1/replay/...`
- `timemachine_test.go`: `/v1/time-machine/...` → `/api/v1/time-machine/...`
- `metrics_test.go`: `/v1/metrics` → `/api/v1/metrics` (cosmetic in request URL)

---

### GOAL 3: Fix Health Endpoint (infrastructure connectivity)

**MODIFY `internal/database/database.go`** — add `Ping(ctx) error` method:
```go
func (d *Database) Ping(ctx context.Context) error { return d.Pool.Ping(ctx) }
```

**MODIFY `internal/redis/redis.go`** — add `Ping(ctx) error` method:
```go
func (r *Redis) Ping(ctx context.Context) error { return r.Client.Ping(ctx).Err() }
```

**REWRITE `internal/api/health.go`** (~90 lines) — new `Pinger` interface:
```go
type Pinger interface { Ping(ctx context.Context) error }
```
Handler takes `database Pinger, redis Pinger, logger`. Pings both on every request, returns:
```json
{"status":"healthy","database":"connected","redis":"connected","uptime":"15m0s","version":"0.1.0"}
```
Returns HTTP 503 when any dependency is disconnected. Tracks `startedAt` for uptime. The old `HealthReader` interface and host-metrics logic are removed (host metrics already live at `/api/v1/dashboard/hosts`). Per your decision, the cache `Get`/`ScanKeys` methods on `RedisLatestMetricsCache` remain for future use.

**REWRITE `internal/api/health_test.go`** — tests using `stubPinger`:
1. Both connected → 200 + `"healthy"`
2. DB down → 503 + `"unhealthy"` + `"database":"disconnected"`
3. Redis down → 503 + `"unhealthy"` + `"redis":"disconnected"`
4. Nil deps rejected

---

### GOAL 4: Fix root route (HTML landing page)

**NEW `internal/api/landing.go`** (~50 lines)
- `landingHandler(version string) http.HandlerFunc` serving a self-contained HTML page:
  - Title: "Sentinel Infrastructure Event Intelligence Platform"
  - Version
  - Available API endpoints list
  - Health endpoint link
- Registered as `GET /` in the router. No more 404.

---

### GOAL 5: Introduce Middleware

**NEW `internal/middleware/` package** (4 files):

| File | Content | Lines |
|---|---|---|
| `recovery.go` | `Recovery(logger)` — panic recovery with zap logging + `runtime/debug` stack trace | ~25 |
| `logging.go` | `Logging(logger)` — request logging (method, path, status, duration) via zap + `statusRecorder` writer | ~35 |
| `cors.go` | `CORS(allowedOrigins []string)` — sets CORS headers, handles OPTIONS preflight | ~30 |
| `chain.go` | `Chain(logger) []func(http.Handler) http.Handler` — assembles full stack: RequestID → RealIP → Recovery → Logging → CORS → Timeout(15s) | ~20 |

Recovery/Logging/CORS are custom (zap-integrated). RequestID/RealIP/Timeout use chi's battle-tested built-ins (no point reinventing). Applied in router via `r.Use(mw.Chain(logger)...)`.

---

### GOAL 6: Validate configuration on startup

**MODIFY `internal/config/config.go`** — add `Validate() error` method:
- Port must be non-empty
- DatabaseURL must be non-empty
- RedisAddress must be non-empty
- If `AIEnabled`: AIAPIKey, AIBaseURL, AIModel must all be non-empty

Called at the end of `Load()` (fail-fast) and again in `server.New()` (defense-in-depth).

**NEW `internal/config/config_test.go`** — tests for valid config, missing port, missing DB URL, AI enabled without API key.

---

### GOAL 8: Package responsibilities (enforced, no violations)
- `database/` — PostgreSQL only ✓ (Ping method is connection management)
- `redis/` — Redis only ✓ (Ping method is connection management)
- `collector/` — metric collection only ✓ (unchanged)
- `api/` — HTTP handlers only ✓ (landing, health, routing)
- `server/` — application bootstrap only ✓ (app, http, dependencies)
- `middleware/` — HTTP middleware only ✓ (new)

---

### GOAL 7 (DI): EventsHandler interface extraction

**MODIFY `internal/api/events.go`** — extract `EventCollector` interface:
```go
type EventCollector interface {
    Collect(ctx context.Context, event models.Event) (eventstore.Event, error)
}
```
Handler now depends on the interface, not the concrete `*events.Collector`. `events_test.go` works unchanged (concrete struct satisfies the interface).

---

### GOAL 9 + 10: Tests
- All existing tests preserved or updated as required
- `go test ./...` must pass 100%

---

### GOAL 11: Deliverables
After implementation: folder structure tree, per-file change explanation, commands to run, verification checklist.

---

### File Summary

| # | File | Action |
|---|---|---|
| 1 | `internal/server/app.go` | NEW |
| 2 | `internal/server/http.go` | NEW |
| 3 | `internal/server/dependencies.go` | NEW |
| 4 | `internal/middleware/recovery.go` | NEW |
| 5 | `internal/middleware/logging.go` | NEW |
| 6 | `internal/middleware/cors.go` | NEW |
| 7 | `internal/middleware/chain.go` | NEW |
| 8 | `internal/api/landing.go` | NEW |
| 9 | `cmd/server/main.go` | MODIFY (shrink) |
| 10 | `internal/api/router.go` | MODIFY (routes + Handlers struct) |
| 11 | `internal/api/health.go` | REWRITE (Pinger-based) |
| 12 | `internal/api/health_test.go` | REWRITE |
| 13 | `internal/api/events.go` | MODIFY (interface extraction) |
| 14 | `internal/api/metrics_test.go` | MODIFY (path) |
| 15 | `internal/api/dashboard_test.go` | MODIFY (paths) |
| 16 | `internal/api/replay_test.go` | MODIFY (paths) |
| 17 | `internal/api/timemachine_test.go` | MODIFY (paths) |
| 18 | `internal/config/config.go` | MODIFY (Validate) |
| 19 | `internal/config/config_test.go` | NEW |
| 20 | `internal/database/database.go` | MODIFY (Ping) |
| 21 | `internal/redis/redis.go` | MODIFY (Ping) |