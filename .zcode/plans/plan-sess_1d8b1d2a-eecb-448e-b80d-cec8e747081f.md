## Live Metric Streaming — Full Implementation Plan

**Decision: Option 1** — production WebSocket endpoint in backend + WS client in frontend. Additive only — no existing signatures or architecture changed beyond adding a Hub field through the dependency chain.

---

### PART A: BACKEND (Go) — 5 new files, 5 modified

#### A1. `go.mod` — ADD dependency
- `github.com/gorilla/websocket v1.5.3`

#### A2. NEW `internal/websocket/hub.go` (~110 lines)
The thread-safe broadcast hub. Production features:
- `Hub` struct: `register chan Client`, `unregister chan Client`, `broadcast chan []byte`, `clients map[Client]struct{}`, `mu sync.RWMutex`
- `NewHub(logger) *Hub` constructor
- `Run(ctx context.Context)` — long-lived goroutine; `select`s on register/unregister/broadcast/ctx.Done
- `Publish(metrics agent.Metrics)` — JSON-marshals and pushes to broadcast channel (non-blocking, drops on full buffer to prevent backpressure stalls)
- Thread-safe: all client map access behind RWMutex; buffered channels prevent publisher blocking
- `ClientCount() int` — for health/observability
- Buffer sizes: register/unregister cap 16, broadcast cap 256

#### A3. NEW `internal/websocket/client.go` (~120 lines)
Per-connection client with ping/pong heartbeat:
- `Client` struct: `hub *Hub`, `conn *websocket.Conn`, `send chan []byte`, `logger`
- `readPump()` — reads (discards) messages, enforces read deadline, handles pong → resets deadline. On error → unregister
- `writePump()` — ticker-based ping (54s) + pong deadline (60s); reads from `send` and writes; closes on full backpressure (send cap 64)
- Constants: `writeWait = 10s`, `pongWait = 60s`, `pingPeriod = 54s` (must be < pongWait), `sendBufferSize = 256`, `maxMessageSize = 512KB`
- `ServeClient(hub, conn, logger)` — the per-connection goroutine launcher

#### A4. NEW `internal/websocket/upgrader.go` (~30 lines)
HTTP→WS upgrader:
- `Upgrader` with `CheckOrigin` permitting configured origins (default: allow all for dev; production config-friendly)
- `UpgradeHTTP(hub, logger, w, r)` — upgrades, sets initial deadlines, spawns `ServeClient`

#### A5. NEW `internal/api/websocket.go` (~40 lines)
HTTP handler wrapping the upgrader, mirroring existing handler patterns:
- `WebsocketHandler` struct holding `hub *websocket.Hub`, `logger *zap.Logger`
- `NewWebsocketHandler(hub, logger) (*WebsocketHandler, error)` — nil-checked constructor (mirrors HealthHandler pattern)
- `Metrics(w, r)` method — the `GET /ws/v1/metrics` handler, calls `websocket.UpgradeHTTP`

#### A6. MODIFY `internal/collector/service.go` (+~15 lines)
Add a local interface + field, publish after cache.Store:
```go
// NEW interface (mirrors LatestMetricsCache pattern)
type MetricBroadcaster interface {
    Publish(ctx context.Context, metrics agent.Metrics) error
}

// NEW field on Service
type Service struct {
    repository MetricsRepository
    events     EventAppender
    cache      LatestMetricsCache
    broadcast  MetricBroadcaster  // NEW
}
```
- `NewService` gets a 4th param `broadcast MetricBroadcaster` (nil-validated like the others)
- In `Record`, after successful `cache.Store` (line 81.5): `s.broadcast.Publish(ctx, metrics)` — logged on error, does NOT fail the Record (broadcast is best-effort; the metric is already persisted)

#### A7. MODIFY `internal/server/dependencies.go` (+~10 lines)
- Construct `hub := websocket.NewHub(log)` before `collector.NewService`
- Start it: `go hub.Run(context.Background())`
- Pass `hub` into `collector.NewService(repository, events, cache, hub)`
- Build `wsHandler, _ := api.NewWebsocketHandler(hub, log)`
- Add `Websocket *api.WebsocketHandler` field to `Dependencies` struct, populate it

#### A8. MODIFY `internal/server/app.go` (+~6 lines)
- Store `hub *websocket.Hub` on `App` struct (so it can be closed in cleanup)
- In `cleanup()`: add `if a.hub != nil { a.hub.Close() }` mirroring the redis close pattern

#### A9. MODIFY `internal/api/router.go` (+2 lines)
- Add `Websocket *WebsocketHandler` to `Handlers` struct
- Register `r.Get("/ws/v1/metrics", handlers.Websocket.Metrics)` OUTSIDE the `/api/v1` group (sibling block)

#### A10. MODIFY `internal/server/http.go` (+1 line)
- Add `Websocket: deps.Websocket` to the `api.Handlers{}` literal

#### A11. MODIFY `internal/websocket/hub.go` — include `Close()` method for graceful shutdown

---

### PART B: FRONTEND (TypeScript) — 4 new files, 4 modified

#### B1. MODIFY `src/config/env.ts` (+2 lines)
- `WS_URL`: derive from `API_URL` (http→ws, https→wss) or read `VITE_WS_URL`; default `/ws/v1/metrics` (proxied in dev)

#### B2. MODIFY `vite.config.ts` (+6 lines)
- Add WS proxy entry `/ws` → `ws://localhost:8080` with `ws: true`

#### B3. NEW `src/services/stream/types.ts` (~10 lines)
- `StreamState = 'connecting' | 'connected' | 'reconnecting' | 'disconnected'`
- `StreamMessage = { type: 'metrics'; payload: Metrics }` envelope

#### B4. NEW `src/services/stream/MockStream.ts` (~60 lines)
Dev-only simulated stream — mirrors backend cadence (1 msg/2s per active host), used when `USE_MOCK_DATA` and WS unavailable. Implements the same Observable contract as the real WS client. Reuses `generateMetrics` (export it from mock/dashboard.ts).

#### B5. NEW `src/services/stream/websocket.ts` (~150 lines)
Production WS client class — `MetricStream`:
- `connect()` — opens WebSocket to `WS_URL`, exponential backoff reconnect (1s→2s→4s→8s→15s cap), max 5 attempts then `disconnected`
- Heartbeat: send ping every 30s; if no pong in 45s, force-close → reconnect
- `subscribe(handler: (m: Metrics) => void): () => void` — observer pattern, returns unsubscribe
- `state` observable via callback (`onStateChange`)
- `close()` — clean shutdown (code 1000), no reconnect
- Auto-fallback: if `USE_MOCK_DATA && connection fails` → switch to `MockStream` (same subscribe API)
- Backpressure: drop messages if queue > 1000

#### B6. NEW `src/hooks/useMetricStream.ts` (~80 lines)
React hook bridging the stream to React Query + components:
- Singleton stream instance (module-level, lazy-initialized)
- `useMetricStream()` returns `{ state: StreamState, metrics: Metrics | null }`
- Maintains a rolling buffer (last 60 points) in React state, updated on each message
- Framer Motion hook: triggers re-render on each new metric (animation)
- Calls `queryClient.setQueryData(['dashboard', 'hosts', 100], ...)` to update the host table incrementally (patch the matching host's snapshot in-place) — keeps React Query cache in sync with the stream
- On mount: `connect()`; on unmount of last subscriber (refcount): `close()`

#### B7. MODIFY `src/components/dashboard/LiveInfrastructureCharts.tsx` (+~30 lines)
- Accept optional `stream: Metrics[]` prop (the rolling buffer)
- If stream data is present, use it instead of (or merged with) the REST data
- Merge strategy: prefer stream points for the last 2 minutes, fill older from REST history
- Convert via existing `toChartData()` (unchanged)

#### B8. MODIFY `src/pages/DashboardPage.tsx` (+~15 lines)
- Call `const { state, buffer } = useMetricStream()`
- Pass `stream={buffer}` to `LiveInfrastructureCharts`
- Show a stream-status pill near the page title: green "Live" / amber "Reconnecting" / red "Disconnected"
- Wrap chart container in `<AnimatePresence>` so new data points animate in

#### B9. NEW `src/components/dashboard/StreamStatusBadge.tsx` (~40 lines)
- Small pill showing WS connection state with pulsing dot (green/amber/red)
- "Live" / "Reconnecting..." / "Disconnected"
- Tooltip with attempt count on reconnecting

---

### Design Decisions

1. **Hub pattern (not per-client broadcast from collector)** — the collector calls `hub.Publish()` once; the Hub fans out to N clients. O(1) work for the collector regardless of subscriber count. This is the canonical Go WS pattern (cf. gorilla's chat example, but production-hardened).

2. **Non-blocking publish** — if the broadcast channel is full (slow clients), `Publish` drops the message rather than blocking `Record()`. The metric is already persisted to Postgres + Redis; streaming is best-effort. The collector's primary job is NOT blocked by WS backpressure.

3. **Ping/pong heartbeat** — 54s ping < 60s pong deadline (standard gorilla idiom). Dead connections are reaped by the server; clients detect dead server via missed pongs.

4. **Mock stream parity** — when `VITE_USE_MOCK_DATA=true` and the WS connection fails (no backend), the frontend transparently switches to `MockStream` which emits realistic metrics every 2s. Same subscribe API → zero UI changes. This matches the existing `withMockFallback` philosophy.

5. **No existing signature changes that break callers** — `collector.NewService` gains a parameter, but its only caller is `dependencies.go` (which I'm updating in the same change). `Record(ctx, metrics)` signature is untouched.

6. **React Query stays the source of truth for tables/cards** — the stream patches the cache incrementally via `setQueryData`. Tables update in real time without invalidation storms. Charts read from a dedicated rolling buffer (smoother animation than cache-driven re-renders).

7. **Reconnection strategy** — exponential backoff with jitter (1s, 2s, 4s, 8s, 15s cap), 5 attempts. After exhaustion: `disconnected` state with manual "Reconnect" button surfaced in the UI.

---

### File Summary

| # | File | Action | Lines |
|---|---|---|---|
| 1 | `backend/go.mod` | MODIFY | +1 |
| 2 | `backend/internal/websocket/hub.go` | NEW | ~110 |
| 3 | `backend/internal/websocket/client.go` | NEW | ~120 |
| 4 | `backend/internal/websocket/upgrader.go` | NEW | ~30 |
| 5 | `backend/internal/api/websocket.go` | NEW | ~40 |
| 6 | `backend/internal/collector/service.go` | MODIFY | +15 |
| 7 | `backend/internal/server/dependencies.go` | MODIFY | +10 |
| 8 | `backend/internal/server/app.go` | MODIFY | +6 |
| 9 | `backend/internal/api/router.go` | MODIFY | +2 |
| 10 | `backend/internal/server/http.go` | MODIFY | +1 |
| 11 | `frontend/src/config/env.ts` | MODIFY | +2 |
| 12 | `frontend/vite.config.ts` | MODIFY | +6 |
| 13 | `frontend/src/services/stream/types.ts` | NEW | ~10 |
| 14 | `frontend/src/services/stream/MockStream.ts` | NEW | ~60 |
| 15 | `frontend/src/services/stream/websocket.ts` | NEW | ~150 |
| 16 | `frontend/src/hooks/useMetricStream.ts` | NEW | ~80 |
| 17 | `frontend/src/components/dashboard/LiveInfrastructureCharts.tsx` | MODIFY | +30 |
| 18 | `frontend/src/pages/DashboardPage.tsx` | MODIFY | +15 |
| 19 | `frontend/src/components/dashboard/StreamStatusBadge.tsx` | NEW | ~40 |

### Verification
- `cd backend && go build ./... && go test ./...` — must pass (no existing tests broken)
- `cd frontend && npx tsc -b && npx vite build` — must pass
- Manual: start backend, open dashboard, see "Live" badge + charts updating every 2s
- Without backend: with `VITE_USE_MOCK_DATA=true`, MockStream kicks in, charts still animate