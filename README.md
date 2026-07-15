<p align="center">
  <img src="https://img.shields.io/badge/Go-1.25-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go 1.25" />
  <img src="https://img.shields.io/badge/PostgreSQL-316192?style=for-the-badge&logo=postgresql&logoColor=white" alt="PostgreSQL" />
  <img src="https://img.shields.io/badge/Redis-DC382D?style=for-the-badge&logo=redis&logoColor=white" alt="Redis" />
  <img src="https://img.shields.io/badge/Docker-2496ED?style=for-the-badge&logo=docker&logoColor=white" alt="Docker" />
  <img src="https://img.shields.io/badge/License-MIT-green?style=for-the-badge" alt="MIT License" />
</p>

# 🛡️ Sentinel

> **Real-time AI-powered security monitoring and alerting platform.**

Sentinel is a full-stack security surveillance system that ingests live video/event streams, applies AI-driven threat detection, and delivers instant alerts through a real-time dashboard. Built with a Go backend and a modern web frontend, it is designed to be deployed as a distributed system with dedicated **server**, **agent**, and **worker** processes.

---

## ✨ Features

| Category | Highlights |
|---|---|
| **AI Detection** | Pluggable AI engine for anomaly & threat detection on video/event streams |
| **Real-time Alerts** | WebSocket-powered live alert delivery to connected dashboards |
| **Event Replay** | Review and replay historical security events for forensic analysis |
| **Metrics & Analytics** | Built-in metrics collection for system health and detection accuracy |
| **Multi-process Architecture** | Separate `server`, `agent`, and `worker` binaries for scalable deployments |
| **Auth & Middleware** | Authentication layer and extensible HTTP middleware pipeline |
| **Object Storage** | Abstracted storage interface for video clips, snapshots, and artifacts |
| **Structured Logging** | Production-grade structured logging via Uber's Zap |

---

## 🏗️ Architecture

```
┌──────────────────────────────────────────────────────────────────┐
│                         Sentinel Platform                        │
├──────────────┬──────────────────┬──────────────┬─────────────────┤
│   Agent      │     Server       │    Worker    │    Frontend     │
│  (collector) │   (API gateway)  │  (processor) │   (dashboard)   │
│              │                  │              │                 │
│  • Captures  │  • REST API      │  • AI infer. │  • Live view    │
│    streams   │  • WebSocket hub │  • Alerting  │  • Event replay │
│  • Forwards  │  • Auth/Authz    │  • Storage   │  • Metrics      │
│    events    │  • Routing       │  • Metrics   │                 │
└──────┬───────┴────────┬─────────┴──────┬───────┴─────────────────┘
       │                │                │
       ▼                ▼                ▼
   ┌────────┐     ┌──────────┐     ┌──────────┐
   │  Redis  │     │ Postgres │     │  Object  │
   │ (pubsub)│     │   (data) │     │  Storage │
   └────────┘     └──────────┘     └──────────┘
```

---

## 📁 Project Structure

```
sentinel/
├── backend/                  # Go backend (module: github.com/aadityya4real/sentinel/backend)
│   ├── cmd/
│   │   ├── server/           # API server entrypoint
│   │   ├── agent/            # Data-collection agent entrypoint
│   │   └── worker/           # Background worker entrypoint
│   ├── config/               # Configuration loader (Viper)
│   └── internal/
│       ├── ai/               # AI/ML inference engine
│       ├── alert/            # Alert generation & dispatch
│       ├── api/              # HTTP router & handlers (Chi)
│       ├── auth/             # Authentication & authorization
│       ├── config/           # Internal config models
│       ├── database/         # PostgreSQL data access (pgx)
│       ├── events/           # Domain event definitions
│       ├── logger/           # Structured logging (Zap)
│       ├── metrics/          # Metrics collection & export
│       ├── middleware/       # HTTP middleware (CORS, rate-limit, etc.)
│       ├── redis/            # Redis client & pub/sub (go-redis)
│       ├── replay/           # Historical event replay
│       ├── storage/          # Object/file storage abstraction
│       └── websocket/        # WebSocket connection manager
├── frontend/                 # Web dashboard (coming soon)
├── deployments/              # Kubernetes / Docker deployment manifests
├── scripts/                  # Dev & CI helper scripts
├── docs/
│   ├── adr/                  # Architectural Decision Records
│   ├── api/                  # API documentation
│   ├── architecture/         # Architecture diagrams & docs
│   └── design/               # Design specifications
├── .github/                  # GitHub Actions CI/CD workflows
├── docker-compose.yml        # Local multi-service orchestration
├── Makefile                  # Build, test, lint shortcuts
└── .env.example              # Environment variable template
```

---

## 🚀 Quick Start

### Prerequisites

- [Go 1.25+](https://go.dev/dl/)
- [Docker & Docker Compose](https://docs.docker.com/get-docker/)
- [Make](https://www.gnu.org/software/make/) (optional, for convenience commands)

### 1. Clone the repository

```bash
git clone https://github.com/aadityya4real/sentinel.git
cd sentinel
```

### 2. Configure environment

```bash
cp .env.example .env
# Edit .env with your database, Redis, and storage credentials
```

### 3. Start infrastructure services

```bash
docker compose up -d
```

### 4. Run the backend

```bash
# Build and run the API server
cd backend
go run ./cmd/server

# In separate terminals, start the agent and worker
go run ./cmd/agent
go run ./cmd/worker
```

---

## ⚙️ Configuration

Sentinel uses [Viper](https://github.com/spf13/viper) for configuration. Settings can be provided via:

| Source | Priority |
|---|---|
| Environment variables | Highest |
| `.env` file | High |
| Config file (`config.yaml` / `config.toml`) | Medium |
| Defaults | Lowest |

See [`.env.example`](.env.example) for all available variables.

---

## 🧰 Tech Stack

| Layer | Technology |
|---|---|
| **Language** | Go 1.25 |
| **HTTP Router** | [Chi v5](https://github.com/go-chi/chi) |
| **Database** | PostgreSQL via [pgx v5](https://github.com/jackc/pgx) |
| **Cache / PubSub** | Redis via [go-redis v9](https://github.com/redis/go-redis) |
| **Configuration** | [Viper](https://github.com/spf13/viper) |
| **Logging** | [Zap](https://github.com/uber-go/zap) |
| **Containerisation** | Docker & Docker Compose |
| **Frontend** | TBD |

---

## 🧪 Development

```bash
# Run all tests
make test

# Lint the codebase
make lint

# Build all binaries
make build

# Format code
make fmt
```

---

## 🗺️ Roadmap

- [ ] Core API server with Chi router & middleware
- [ ] PostgreSQL schema & migrations
- [ ] Redis pub/sub event bus
- [ ] AI inference pipeline integration
- [ ] Real-time WebSocket alert delivery
- [ ] Event replay & forensic search
- [ ] Frontend dashboard (React / Next.js)
- [ ] Kubernetes deployment manifests
- [ ] CI/CD with GitHub Actions
- [ ] Comprehensive API documentation (OpenAPI)

---

## 📄 Documentation

| Document | Path |
|---|---|
| Architecture | [`docs/architecture/`](docs/architecture/) |
| API Reference | [`docs/api/`](docs/api/) |
| Design Specs | [`docs/design/`](docs/design/) |
| ADRs | [`docs/adr/`](docs/adr/) |

---

## 🤝 Contributing

Contributions are welcome! Please follow these steps:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'feat: add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

Please ensure your code passes linting and all tests before submitting.

---

## 📜 License

This project is licensed under the **MIT License** — see the [LICENSE](LICENSE) file for details.

---

<p align="center">
  Built with ❤️ by <a href="https://github.com/aadityya4real">@aadityya4real</a>
</p>
