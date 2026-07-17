# 🛡️ Sentinel

<div align="center">

**Record. Replay. Explain.**

_An AI-powered infrastructure event intelligence platform that ingests live streams, detects anomalies, and delivers real-time insights._

[![Go Version](https://img.shields.io/badge/Go-1.25.2-00ADD8?style=for-the-badge&logo=go)](https://go.dev)
[![React](https://img.shields.io/badge/React-19-61DAFB?style=for-the-badge&logo=react)](https://react.dev)
[![TypeScript](https://img.shields.io/badge/TypeScript-5.7-3178C6?style=for-the-badge&logo=typescript)](https://typescriptlang.org)
[![Tailwind CSS](https://img.shields.io/badge/Tailwind-3-06B6D4?style=for-the-badge&logo=tailwind-css)](https://tailwindcss.com)

[![License: MIT](https://img.shields.io/badge/License-MIT-green?style=for-the-badge)](LICENSE)
[![Docker](https://img.shields.io/badge/Docker-Compose-2496ED?style=for-the-badge&logo=docker)](https://compose.docker.io)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-17-316192?style=for-the-badge&logo=postgresql)](https://www.postgresql.org)
[![Redis](https://img.shields.io/badge/Redis-8-DC382D?style=for-the-badge&logo=redis)](https://redis.io)

[Report Bug](https://github.com/aadityya4real/sentinel/issues) · [Request Feature](https://github.com/aadityya4real/sentinel/issues)

</div>

---

## 🔭 Overview

Sentinel is a **full-stack security observability platform** that gives you complete visibility into your infrastructure. It captures events from agents across your fleet, stores them in a high-performance event store, runs AI-powered analysis for anomaly detection, and presents everything through a sleek real-time dashboard.

Think of it as your **central nervous system for infrastructure intelligence** — watching, learning, and alerting.

<div align="center">

```
┌──────────┐     ┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│  Agent   │────▶│   Collector  │────▶│  Event Store │────▶│   Replay     │
│(capture) │     │  (ingest)    │     │(PostgreSQL)  │     │ (forensic)   │
└──────────┘     └──────┬───────┘     └──────┬───────┘     └──────┬───────┘
                        │                    │                     │
                        ▼                    ▼                     ▼
                  ┌──────────────┐     ┌──────────────┐     ┌──────────────┐
                  │    Redis     │     │     AI       │     │  Dashboard   │
                  │  (pub/sub)   │     │  Engine      │     │  (React)     │
                  └──────────────┘     └──────────────┘     └──────────────┘
```

</div>

---

## ✨ Features

<table>
<tr>
<td width="50%" valign="top">

### 🤖 AI-Powered Analysis

Pluggable AI inference pipeline that analyzes events in real-time, detecting anomalies and generating natural-language explanations of what is happening in your infrastructure.

</td>
<td width="50%" valign="top">

### ⚡ Real-Time Streaming

WebSocket-powered live event delivery. Watch your infrastructure unfold in real-time with instant alert notifications pushed directly to your dashboard.

</td>
</tr>
<tr>
<td valign="top">

### 🎬 Event Replay

Full forensic replay capability. Time-travel through your infrastructure history with snapshot comparison and timeline scrubbing.

</td>
<td valign="top">

### 📊 Rich Dashboards

Beautiful, dark-themed dashboards with interactive charts, fleet overviews, host tables, and event timelines — all built with React and Recharts.

</td>
</tr>
<tr>
<td valign="top">

### 🏗️ Multi-Process Architecture

Separate `server`, `agent`, and `worker` binaries for horizontal scalability. Each component can scale independently.

</td>
<td valign="top">

### 🔐 Secure by Design

Built-in authentication middleware, structured logging with Zap, and environment-based configuration management via Viper.

</td>
</tr>
</table>

---

## 📁 Project Structure

```
sentinel/
├── backend/                 # Go backend (server, agent, worker)
│   ├── cmd/
│   │   ├── server/          # API gateway and HTTP server
│   │   ├── agent/           # Event capture agent
│   │   └── worker/          # Background processor
│   └── internal/
│       ├── ai/              # AI inference pipeline
│       ├── alert/           # Alerting engine
│       ├── api/             # HTTP handlers and routes
│       ├── auth/            # Authentication middleware
│       ├── collector/       # Event ingestion
│       ├── dashboard/       # Dashboard data aggregation
│       ├── database/        # PostgreSQL access (pgx)
│       ├── events/          # Domain event definitions
│       ├── eventstore/      # Event persistence
│       ├── logger/          # Structured logging (Zap)
│       ├── metrics/         # System metrics
│       ├── middleware/      # CORS, rate-limit, etc.
│       ├── models/          # Shared data models
│       ├── redis/           # Redis pub/sub (go-redis)
│       ├── replay/          # Historical event replay
│       ├── server/          # HTTP server setup
│       ├── storage/         # Object storage abstraction
│       ├── timemachine/     # Time-travel debugging
│       └── websocket/       # WebSocket connection hub
├── frontend/                # React + TypeScript dashboard
│   ├── src/
│   │   ├── components/
│   │   │   ├── charts/      # Area charts, gauges, sparklines
│   │   │   ├── dashboard/   # Fleet cards, host table, timeline
│   │   │   ├── timemachine/ # Replay controls, sliders
│   │   │   └── ui/          # Buttons, badges, skeletons
│   │   ├── layouts/         # App layout with sidebar and topbar
│   │   ├── pages/           # Dashboard, hosts, events, AI, etc.
│   │   ├── services/        # API client and hooks
│   │   └── types/           # TypeScript type definitions
│   └── public/
├── docs/                    # Architecture and design docs
│   ├── adr/                 # Architectural Decision Records
│   ├── api/                 # API documentation
│   ├── architecture/        # System architecture docs
│   └── design/              # Design specifications
├── deployments/             # Kubernetes and deployment manifests
├── scripts/                 # Dev and CI helper scripts
├── docker-compose.yml       # Local development stack
└── .env.example             # Configuration template
```

---

## 🚀 Quick Start

### Prerequisites

| Requirement | Version |
|---|---|
| [Go](https://go.dev/dl/) | 1.25.2+ |
| [Node.js](https://nodejs.org/) | 18+ |
| [Docker and Compose](https://docs.docker.com/get-docker/) | Latest |

### One-Command Setup

```bash
# Clone and enter
git clone https://github.com/aadityya4real/sentinel.git
cd sentinel

# Copy configuration
cp .env.example .env

# Start infrastructure (PostgreSQL + Redis)
docker compose up -d

# Backend
cd backend && go run ./cmd/server
# In separate terminals:
go run ./cmd/agent
go run ./cmd/worker

# Frontend (new terminal)
cd frontend && npm install && npm run dev
```

Open **[http://localhost:5173](http://localhost:5173)** and watch Sentinel come alive.

---

## ⚙️ Configuration

Sentinel uses **[Viper](https://github.com/spf13/viper)** for configuration management with a clear priority hierarchy:

```
Environment Variables > .env file > config.yaml > defaults
```

Key configuration options:

| Variable | Description | Default |
|---|---|---|
| `APP_PORT` | HTTP server port | `8080` |
| `LOG_LEVEL` | Logging verbosity | `debug` |
| `POSTGRES_HOST` | Database connection | `localhost` |
| `REDIS_HOST` | Redis connection | `localhost` |
| `AI_ENABLED` | Toggle AI pipeline | `false` |
| `AI_MODEL` | Inference model | `gpt-5-mini` |

See [`.env.example`](.env.example) for the full list.

---

## 🧪 Development

```bash
# Backend
cd backend
go run ./cmd/server    # Start API server
go run ./cmd/agent     # Start event agent
go run ./cmd/worker    # Start background worker

# Frontend
cd frontend
npm run dev            # Vite dev server
npm run build          # Production build
npm run typecheck      # TypeScript validation
```

---

## 🎨 Frontend Preview

Sentinel features a **dark-themed, production-quality dashboard** with:

- **Fleet Overview** — Real-time host metrics at a glance
- **Live Charts** — Infrastructure telemetry with animated area charts and gauges
- **Event Timeline** — Chronological security event feed
- **Host Management** — Detailed per-host breakdowns
- **AI Insights** — Natural-language anomaly explanations
- **Time Machine** — Forensic event replay with snapshot comparison

<div align="center">

_Built with Inter + JetBrains Mono, TailwindCSS, Framer Motion animations, and a custom purple-accented dark theme._

</div>

---

## 📖 Documentation

| Resource | Location |
|---|---|
| Architecture Decisions | [`docs/adr/`](docs/adr/) |
| API Reference | [`docs/api/`](docs/api/) |
| System Architecture | [`docs/architecture/`](docs/architecture/) |
| Design Specifications | [`docs/design/`](docs/design/) |

---

## 🗺️ Roadmap

| Status | Feature |
|---|---|
| ✅ | Go backend with Chi router and middleware |
| ✅ | Structured logging (Zap) |
| ✅ | Configuration management (Viper) |
| ✅ | React dashboard with TailwindCSS |
| ✅ | Real-time WebSocket communication |
| ✅ | AI inference pipeline integration |
| ✅ | Event replay and forensic search |
| ✅ | Time Machine with snapshot comparison |
| 🚧 | PostgreSQL schema and migrations |
| 🚧 | Kubernetes deployment manifests |
| 🚧 | CI/CD with GitHub Actions |
| 🚧 | OpenAPI specification |

---

## 👥 Contributing

Contributions are welcome! Here is how to get started:

1. **Fork** the repository
2. **Create** a feature branch → `git checkout -b feat/amazing-feature`
3. **Commit** your changes → `git commit -m 'feat: add amazing feature'`
4. **Push** → `git push origin feat/amazing-feature`
5. **Open** a Pull Request

Please ensure your code follows Go formatting standards and TypeScript type-checks cleanly.

---

## 📜 License

This project is licensed under the **MIT License** — see the [LICENSE](LICENSE) file for details.

---

<div align="center">

**Built with 💜 by [aadityya4real](https://github.com/aadityya4real)**

_Sentinel — Your infrastructure, under surveillance._

</div>
