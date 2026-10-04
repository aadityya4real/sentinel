# SENTINEL project context

Sentinel is an infrastructure monitoring MVP for collecting host metrics and immutable infrastructure events, inspecting recent and historical state, replaying event timelines, and requesting AI incident analysis.

## Runtime architecture

- The Go agent collects host metrics and submits them to the Go HTTP API with a Bearer token.
- The API validates and stores metrics and events in PostgreSQL.
- Redis keeps short-lived latest metric and latest event state.
- The API publishes live metric updates to connected WebSocket clients.
- The React/Vite dashboard reads public read-only API routes and the metrics WebSocket. Mock data is enabled only with `VITE_USE_MOCK_DATA=true`.
- Replay, Time Machine, and AI analysis query the PostgreSQL infrastructure event store.

## Configuration

The active server configuration is `backend/internal/config`. The Go server reads process environment variables and does not load `.env` automatically. Docker Compose reads root `.env` values for Compose substitutions. Do not put backend secrets in `VITE_*` variables.

## Development checks

- Backend: `cd backend && go test ./...`
- Frontend: `cd frontend && npm run build`
- Infrastructure config: `docker compose config --quiet`

The unused duplicate configuration package and legacy relational event repository/model have been removed. Migration 003 remains embedded because the state of existing databases and their legacy data has not been verified.
