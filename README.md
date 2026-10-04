# Sentinel

Infrastructure monitoring with AI event analysis, event replay, and a real-time dashboard.

The API accepts authenticated telemetry from the agent, persists metrics and immutable events, and serves a React dashboard with replay, Time Machine, and on-demand AI analysis.

## What it does

- Collects CPU, memory, and disk metrics from agents across your fleet
- Stores metrics and immutable events in PostgreSQL, with Redis caching the latest metric and event state
- Provides on-demand event analysis through configurable OpenAI-compatible endpoints
- Delivers live metric streams to the dashboard over WebSocket
- Lets you replay historical events or time-travel through infrastructure state with snapshots

## Quick start

```bash
git clone https://github.com/aadityya4real/sentinel.git
cd sentinel

cp .env.example .env
docker compose up -d
```

Start the API server:

```bash
cd backend
go run ./cmd/server   # API server on :8080
```

In another terminal, provide the same `SENTINEL_API_TOKEN` to the server and agent, then run `go run ./cmd/agent`. The agent submits host metrics to the API over HTTP. The Go programs read process environment variables; they do not load `.env` automatically. Compose uses the root `.env` for its own substitutions.

Then the frontend:

```bash
cd ../frontend
npm install
npm run dev
```

Dashboard at `http://localhost:5173`.

For a frontend demo without a backend, set `VITE_USE_MOCK_DATA=true` in `frontend/.env`. Mock responses and the mock metric stream are enabled only by this explicit setting.

## Pages

| Route | What it shows |
|---|---|
| `/dashboard` | Fleet overview: host count, live charts, and recent events |
| `/hosts` | Full host inventory with search |
| `/hosts/:hostname` | Single host history: CPU/memory area charts over configurable time range |
| `/events` | Event timeline with severity filtering (CPU, memory, disk, network, etc.) |
| `/replay` | Filter and review past events by time range |
| `/time-machine` | Animated timeline slider for comparing host state at historical times |
| `/ai` | Type a hostname and time window, get a natural-language analysis of what happened |
| `/settings` | App config: API URL, refresh interval, version info |

## Backend structure

The server and agent share the same Go module:

```
backend/
├── cmd/server/      HTTP API and WebSocket hub
└── cmd/agent/       Host-level metric collection and authenticated HTTP submission
    internal/
        ├── agent/       Host metric collection types and collectors
        ├── ai/          OpenAI-compatible endpoint integration
        ├── api/         HTTP handlers per route
        ├── collector/   Metric ingestion from agents
        ├── config/      Canonical server configuration
        ├── dashboard/   Aggregated queries for the fleet overview
        ├── database/    PostgreSQL connection and embedded migrations
        ├── events/      Validated event ingestion service
        ├── eventstore/  Event persistence layer
        ├── logger/      Zap structured logging
        ├── middleware/  CORS, recovery, logging
        ├── models/      Shared structs
        ├── redis/       go-redis client
        ├── replay/      Historical event retrieval
        ├── server/      HTTP server bootstrap
        ├── storage/     PostgreSQL repositories and Redis caches
        ├── timemachine/ Snapshot creation and comparison
        └── websocket/   Broadcast hub for live metric streaming
```

The Go server reads environment variables through Viper and does not automatically load `.env`. Docker Compose reads the root `.env` file for Compose substitutions; provide values to the server and agent process through the shell, service manager, or container environment. See `.env.example` for local development values.

## Security configuration

Set `SENTINEL_API_TOKEN` to the same random value of at least 32 characters for the server and agent. The agent sends it as a Bearer token for metric submissions. Metric/event ingestion and AI analysis require this token; read-only health, dashboard, event-list, replay, and Time Machine endpoints remain public for the browser dashboard. The frontend must not contain the server token, so browser-based AI analysis requires a future session-authenticated server-side path.

Set `SENTINEL_ALLOWED_ORIGINS` to a comma-separated list of exact browser origins. Development defaults are limited to local frontend origins; production requires an explicit list. PostgreSQL and Redis Compose ports bind to loopback for local development. Compose's default PostgreSQL credential and unauthenticated local Redis are development-only; configure credentials and keep both services private in production. When AI is enabled, remote provider URLs must use HTTPS; HTTP is accepted only for loopback endpoints in non-production environments.

## Frontend stack

React 19 + TypeScript, Vite build, TailwindCSS. No Redux or Zustand — TanStack Query handles server state, local state stays in components. Charts are Recharts with custom area/spline rendering. Animations use Framer Motion (page transitions, staggered list entries).

The dashboard uses `MetricCard`, `InfChart`, and `HostTable`; host details use `AreaChartCard`. Time Machine uses `TimelineSlider`, `ReplayControls`, and `SnapshotComparison` alongside shared UI primitives.

Mock data is in `services/mock/events.ts` and generates timestamped infrastructure events across five fake hosts for development without a running backend.

## Architecture

Agent submits metrics over authenticated HTTP → API validates and writes metrics/events to PostgreSQL, updates Redis latest-state caches, and broadcasts metrics to WebSocket clients → Dashboard renders current and historical infrastructure state. AI analysis is an on-demand API operation over the event store; Replay and Time Machine also query that store.

Event replay and Time Machine query the PostgreSQL infrastructure event store to reconstruct and compare historical state.

## Development

```bash
# Backend only (with Docker infra)
cd backend && go run ./cmd/server

# Frontend only (mock mode)
cd frontend && npm run dev

# Both
# Terminal 1: cd backend && go run ./cmd/server
# Terminal 2: cd frontend && npm run dev
```

Docker Compose manages PostgreSQL 17 and Redis 8 with health checks and persistent volumes. Its defaults are for local development only and its database ports are bound to loopback.

## License

MIT
