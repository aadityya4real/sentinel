# Sentinel

Infrastructure monitoring with AI event analysis, event replay, and a real-time dashboard.

Three processes ingest host metrics, store events, and run AI analysis — all visible through a React frontend.

## What it does

- Collects CPU, memory, disk, network metrics from agents across your fleet
- Stores everything in PostgreSQL with Redis for pub/sub between components
- Runs anomaly detection through a pluggable AI pipeline (OpenAI-compatible endpoints)
- Delivers live metric streams to the dashboard over WebSocket
- Lets you replay historical events or time-travel through infrastructure state with snapshots

## Quick start

```bash
git clone https://github.com/aadityya4real/sentinel.git
cd sentinel

docker compose up -d

cp .env.example .env
```

Start the three backend processes:

```bash
cd backend
go run ./cmd/server   # API server on :8080
go run ./cmd/agent    # event collector
go run ./cmd/worker   # AI inference + background processing
```

Then the frontend:

```bash
cd ../frontend
npm install
npm run dev
```

Dashboard at `http://localhost:5173`.

If you haven't wired up the backend yet, set `VITE_USE_MOCK_DATA=true` in `.env` — the frontend falls back to mock data for every page.

## Pages

| Route | What it shows |
|---|---|
| `/dashboard` | Fleet overview: host count, live charts, recent events, AI insights card |
| `/hosts` | Full host inventory with search |
| `/hosts/:hostname` | Single host history: CPU/memory area charts over configurable time range |
| `/events` | Event timeline with severity filtering (CPU, memory, disk, network, etc.) |
| `/replay` | Filter and review past events by time range |
| `/time-machine` | Animated timeline slider — scrub through the last 2 hours of host snapshots with comparison views |
| `/ai` | Type a hostname and time window, get a natural-language analysis of what happened |
| `/settings` | App config: API URL, refresh interval, version info |

## Backend structure

Three binaries share the same Go module:

```
backend/
├── cmd/server/      REST API (Chi router), WebSocket hub, auth middleware
├── cmd/agent/       Host-level metric collection via gopsutil
└── cmd/worker/      AI inference dispatch + alert evaluation
    internal/
        ├── ai/          OpenAI-compatible endpoint integration
        ├── alert/       Rule-based alerting
        ├── api/         HTTP handlers per route
        ├── auth/        Token validation middleware
        ├── collector/   Metric ingestion from agents
        ├── dashboard/   Aggregated queries for the fleet overview
        ├── database/    PostgreSQL via pgx
        ├── events/      Domain event types
        ├── eventstore/  Event persistence layer
        ├── logger/      Zap structured logging
        ├── metrics/     Host metric types and aggregation
        ├── middleware/  CORS, recovery, logging
        ├── models/      Shared structs
        ├── redis/       go-redis client + pub/sub channels
        ├── replay/      Historical event retrieval
        ├── server/      HTTP server bootstrap
        ├── storage/     Object storage interface (local/S3)
        ├── timemachine/ Snapshot creation and comparison
        └── websocket/   Broadcast hub for live metric streaming
```

Configuration uses Viper — env vars, `.env` file, then defaults. See `.env.example`.

## Frontend stack

React 19 + TypeScript, Vite build, TailwindCSS. No Redux or Zustand — TanStack Query handles server state, local state stays in components. Charts are Recharts with custom area/spline rendering. Animations use Framer Motion (page transitions, staggered list entries).

Components fall into four buckets: charts (AreaChartCard, GaugeCard, Sparkline), dashboard (FleetOverviewCards, HostTable, LiveInfrastructureCharts, RecentEventsTimeline), timemachine (TimelineSlider, ReplayControls, SnapshotComparison), and shared UI primitives (Badge, Button, Card, EmptyState, ErrorState, Skeleton, Spinner).

Mock data is in `services/mock/events.ts` and generates timestamped infrastructure events across five fake hosts for development without a running backend.

## Architecture

Agent collects metrics → forwards to Server via gRPC/HTTP → Server writes to PostgreSQL and publishes to Redis pub/sub → Dashboard subscribes via WebSocket → AI Worker analyzes events in parallel.

Event replay pulls from the PostgreSQL event store. Time Machine takes periodic snapshots and lets you compare two points in time.

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

Docker Compose manages PostgreSQL 17 and Redis 8 with health checks and persistent volumes. The default credentials in `.env.example` match the compose setup out of the box.

## License

MIT
