## Issue #4: Infrastructure Health — Implementation Plan

### Summary of Changes

This issue adds a **read path to Redis** (currently both caches are write-only) and wires it into the health endpoint so it reports live infrastructure status from cache. We also fix TTL from `0` (no expiry) to `5 minutes`.

### File-by-File Changes (6 files modified, 2 new files)

---

#### 1. `backend/internal/storage/redis_metrics.go` — **MODIFY**
- Change TTL from `0` to `5 * time.Minute` in the `Set` call
- Add a `Get(ctx context.Context, hostname string) (agent.Metrics, error)` method that reads back the cached metrics for a host
- Return `ErrCacheMiss` sentinel error when the key does not exist (matches Go `sync.Map` pattern)

#### 2. `backend/internal/collector/service.go` — **MODIFY**
- Expand the `LatestMetricsCache` interface to add a `Get` method:
  ```go
  type LatestMetricsCache interface {
      Store(ctx context.Context, metrics agent.Metrics) error
      Get(ctx context.Context, hostname string) (agent.Metrics, error)
  }
  ```
- Define `ErrCacheMiss` as a package-level sentinel error

#### 3. `backend/internal/collector/service_test.go` — **MODIFY**
- Update the `memoryCache` stub to implement the new `Get` method (returns `collector.ErrCacheMiss` by default)
- Add test for cache miss scenario

#### 4. `backend/internal/api/router.go` — **MODIFY**
- Replace the static inline `/health` handler with a new `HealthHandler` parameter
- The new handler reads latest metrics from Redis cache and computes host health status

#### 5. `backend/internal/api/health.go` — **NEW FILE**
- `HealthHandler` struct with `LatestMetricsCache` dependency (reuses the interface from `collector`)
- `GetHealth(ctx) -> HealthResponse` method that:
  - Calls `SCAN` on Redis with key pattern `sentinel:metrics:latest:*` to get all cached hosts
  - For each host, unmarshal metrics and classify as `healthy` (last seen within 5 min) or `degraded` (older)
  - Returns `{ status: "healthy"|"degraded"|"unhealthy", hosts: [...], version: "0.1.0" }`
- `ServeHTTP` method for the Chi router (follows existing handler pattern: nil-check constructor, structured error responses)
- Kept under 200 lines

#### 6. `backend/internal/api/health_test.go` — **NEW FILE**
- Unit tests using a stub cache:
  - `TestHealthHandlerReturnsDegradedWhenHostsStale` — metrics exist but are older than 5 min
  - `TestHealthHandlerReturnsHealthyWhenHostsActive` — metrics within 5 min window
  - `TestHealthHandlerReturnsUnhealthyWhenNoHosts` — empty cache
  - `TestNewHealthHandlerRejectsNilDependencies`

#### 7. `backend/cmd/server/main.go` — **MODIFY**
- Wire `cache` (which now has both `Store` and `Get`) into a new `api.NewHealthHandler(cache, logg)`
- Pass `healthHandler` into `api.NewRouter(...)` as the new last parameter

#### 8. `backend/internal/storage/redis_latest_events.go` — **MODIFY**
- Change TTL from `0` to `5 * time.Minute` in the `Set` call (matches the acceptance criteria for cache TTL)

---

### Design Decisions

- **Reuse `LatestMetricsCache` interface** (defined in `collector`) for the health handler read path — no new interface needed. The health handler reads from the same cache the collector writes to.
- **`SCAN` pattern** to discover all hosts in Redis — avoids maintaining a separate host list. Uses `MATCH sentinel:metrics:latest:*` with `COUNT 100`.
- **Three health states**: `healthy` (at least one active host, none stale), `degraded` (hosts exist but some are stale), `unhealthy` (no hosts at all).
- **Sentinel error `ErrCacheMiss`** — clean pattern for distinguishing cache miss from internal errors. `redis.Nil` from go-redis maps to this.
- **No new packages** — everything fits into existing `storage`, `collector`, and `api` packages.
- **No architecture changes** — same DI pattern, same handler struct pattern, same Redis client pattern.

### Acceptance Criteria Mapping

| Criteria | How it's met |
|---|---|
| Redis caches latest metrics | Already exists, adding TTL fix |
| Redis caches latest health | Already exists for events, TTL fix |
| Cache TTL = 5 minutes | `5 * time.Minute` in both `Set` calls |
| Collector updates cache | Already happens, no change needed |
| Health endpoint reads cache | New `HealthHandler` reads from Redis via `LatestMetricsCache.Get` |
| Unit tested | `health_test.go` + updated `service_test.go` |
| Production-ready | Structured errors, ctx propagation, graceful SCAN |
| Compiles successfully | Verified via `go build ./...` |