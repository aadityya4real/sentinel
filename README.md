# Sentinel

Infrastructure observability, reimagined.

Sentinel captures, stores, and analyzes infrastructure events in real time — giving you end-to-end visibility into your entire fleet with AI-powered anomaly detection, forensic event replay, and a production-grade dashboard.

## Architecture

```
┌──────────────────────────────────────────────────────────────┐
│                    Sentinel Platform                         │
├───────────┬──────────────┬──────────────┬───────────────────┤
│   Agent   │    Server    │    Worker    │     Frontend      │
│  Capture  │  API Gateway │  Processing  │   React Dashboard │
└─────┬─────┴──────┬───────┴──────┬───────┴────────┬──────────┘
      │           │              │                 │
      ▼           ▼              ▼                 ▼
  ┌────────┐ ┌──────────┐ ┌──────────┐    ┌──────────────┐
  │ Events │ │ PostgreSQL│ │  Redis   │    │ Live Charts  │
  │  Store │ │  (pgx)    │ │(pub/sub) │    │ WebSocket Hub│
  └────────┘ └──────────┘ └──────────┘    └──────────────┘
                                         ┌──────────────┐
                                         │   AI Engine   │
                                         │ Anomaly Detect│
                                         └──────────────┘
```

## Core Capabilities

### Fleet Monitoring
- Real-time host discovery and metrics collection
- Per-host CPU, memory, disk, and network telemetry
- Interactive fleet overview with live status indicators
- Deep-dive host detail pages with granular breakdowns

### Event Intelligence
- Structured event ingestion via agents across the fleet
- Filterable event timeline with severity classification
- Type-based filtering: CPU, memory, disk, host, network
- Custom badge system for info, warning, and critical states

### AI-Powered Analysis
- Pluggable AI inference pipeline for anomaly detection
- Natural-language explanations of infrastructure incidents
- Configurable hostname and time-window analysis
- Persistent analysis history with local storage

### Forensic Time Machine
- Replay infrastructure state at any point in the past
- Interactive timeline slider with configurable step size
- Snapshot comparison between two time points
- 2-hour lookback window with 1-minute resolution

### Real-Time Streaming
- WebSocket-powered metric stream to connected dashboards
- Connection status indicator with reconnection attempts
- Live buffering with smooth chart updates
- Framer Motion animations for fluid transitions

## Tech Stack

| Layer | Technology |
|---|---|
| **Backend** | Go 1.25.2, Chi v5 router |
| **Frontend** | React 19, TypeScript 5.7, Vite 6 |
| **UI Framework** | TailwindCSS 3, Framer Motion |
| **Data Viz** | Recharts, Sparklines, Gauges |
| **State Management** | TanStack Query v5 |
| **Database** | PostgreSQL 17 (pgx v5) |
| **Pub/Sub** | Redis 8 (go-redis v9) |
| **Configuration** | Viper |
| **Logging** | Zap (Uber) |
| **Containerization** | Docker Compose |

## Project Layout

```
sentinel/
├── backend/                      # Go backend
│   ├── cmd/
│   │   ├── server/               # API server & HTTP routes
│   │   ├── agent/                # Infrastructure capture agent
│   │   └── worker/               # Background processing daemon
│   └── internal/
│       ├── ai/                   # AI inference engine
│       ├── alert/                # Alerting logic
│       ├── api/                  # HTTP handlers
│       ├── auth/                 # Authentication middleware
│       ├── collector/            # Event ingestion layer
│       ├── dashboard/            # Aggregated dashboard queries
│       ├── database/             # PostgreSQL data access
│       ├── events/               # Domain event types
│       ├── eventstore/           # Event persistence
│       ├── logger/               # Structured logging setup
│       ├── metrics/              # System metrics collection
│       ├── middleware/           # CORS, rate limiting, etc.
│       ├── models/               # Shared type definitions
│       ├── redis/                # Redis client & pub/sub
│       ├── replay/               # Historical event retrieval
│       ├── server/               # HTTP server configuration
│       ├── storage/              # Object storage abstraction
│       ├── timemachine/          # Time-travel debugging
│       └── websocket/            # WebSocket hub implementation
├── frontend/                     # React dashboard
│   ├── public/
│   │   └── sentinel.svg          # Brand asset
│   └── src/
│       ├── components/
│       │   ├── charts/           # AreaChartCard, GaugeCard, Sparkline
│       │   ├── dashboard/        # FleetOverviewCards, HostTable, etc.
│       │   ├── timemachine/      # TimelineSlider, ReplayControls
│       │   └── ui/               # Button, Badge, Card, Skeleton...
│       ├── layouts/              # AppLayout, Sidebar, TopBar
│       ├── pages/
│       │   ├── DashboardPage     # Fleet overview + live metrics
│       │   ├── EventsPage        # Filterable event timeline
│       │   ├── HostsPage         # Host inventory list
│       │   ├── HostDetailPage    # Single host deep-dive
│       │   ├── ReplayPage        # Event replay viewer
│       │   ├── TimeMachinePage   # Forensic timeline tool
│       │   ├── AIPage            # AI incident analyzer
│       │   └── SettingsPage      # Configuration panel
│       ├── services/             # API client + mock data
│       ├── hooks/                # useMetricStream, etc.
│       └── types/                # TypeScript type definitions
├── docs/
│   ├── adr/                      # Architectural Decision Records
│   ├── api/                      # API documentation
│   ├── architecture/             # System architecture docs
│   └── design/                   # Design specifications
├── deployments/                  # Kubernetes manifests
├── scripts/                      # Dev helper scripts
├── docker-compose.yml            # Local infrastructure stack
└── .env.example                  # Environment template
```

## Getting Started

### Prerequisites

- **Go** 1.25.2 or later
- **Node.js** 18+ with npm
- **Docker** and Docker Compose

### Infrastructure

```bash
docker compose up -d
```

This starts PostgreSQL 17 and Redis 8 with health checks, persistent volumes, and sensible defaults.

### Backend

```bash
cd backend

# Copy configuration
cp ../.env.example .env

# Start the API server
go run ./cmd/server

# Start the event capture agent (separate terminal)
go run ./cmd/agent

# Start the background processor (separate terminal)
go run ./cmd/worker
```

### Frontend

```bash
cd frontend

npm install
npm run dev
```

The dashboard opens at [http://localhost:5173](http://localhost:5173).

## Configuration

All settings are managed through **[Viper](https://github.com/spf13/viper)**. The configuration priority is:

```
Environment Variables > .env file > config file > defaults
```

### Key Variables

| Variable | Purpose | Example |
|---|---|---|
| `APP_PORT` | HTTP listener port | `8080` |
| `APP_ENV` | Runtime environment | `development` |
| `LOG_LEVEL` | Verbosity level | `debug` |
| `POSTGRES_HOST` | Database address | `localhost` |
| `POSTGRES_PORT` | Database port | `5432` |
| `POSTGRES_USER` | Database user | `sentinel` |
| `POSTGRES_PASSWORD` | Database password | `sentinel` |
| `POSTGRES_DB` | Database name | `sentinel` |
| `REDIS_HOST` | Redis address | `localhost` |
| `REDIS_PORT` | Redis port | `6379` |
| `AI_ENABLED` | Toggle AI pipeline | `false` |
| `AI_BASE_URL` | Inference endpoint | `https://api.openai.com/v1` |
| `AI_API_KEY` | API credential | — |
| `AI_MODEL` | Model identifier | `gpt-5-mini` |

See `.env.example` for the complete reference.

## Dashboard

| Page | Description |
|---|---|
| **Dashboard** | Fleet-wide overview with live metric streams, host table, recent events, and AI insights |
| **Hosts** | Inventory of all monitored hosts with aggregate health status |
| **Host Detail** | Granular per-host metrics, events, and configuration |
| **Events** | Full event log with severity badges and multi-category filters |
| **Replay** | Review historical events with time-filtered search |
| **Time Machine** | Forensic playback of infrastructure state with snapshot comparison |
| **AI** | Incident analyzer — select a host and time range, get natural-language analysis |
| **Settings** | System configuration and preferences |

### Visual Design

- Dark color scheme (`#0a0a0f` base) with purple accent (`#7c3aed`)
- Inter for body text, JetBrains Mono for code and monospaced data
- Rounded cards with subtle borders and glow effects
- Smooth page transitions via Framer Motion

## Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/name`
3. Make your changes
4. Commit with conventional commits: `git commit -m "feat: description"`
5. Push: `git push origin feature/name`
6. Open a Pull Request

Code should follow Go formatting standards (`gofmt`, `golint`) and TypeScript strict mode rules.

## License

MIT
