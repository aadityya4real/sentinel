## Sentinel Frontend — Complete Architecture Plan

**Stack:** React 19 + TypeScript + Vite + Tailwind v4 + React Router + TanStack Query + Recharts + Framer Motion + Lucide
**Accent:** Indigo/Violet dark theme
**Data:** Real backend via React Query; dev-only mock fallback isolated in `services/mock/` behind `VITE_USE_MOCK_DATA`

---

### Project Structure (feature-based, ≤250 lines/file)

```
frontend/
├── index.html
├── package.json
├── tsconfig.json
├── tsconfig.node.json
├── vite.config.ts
├── .env.example
├── src/
│   ├── main.tsx                    # React root + providers
│   ├── App.tsx                     # Router + route definitions
│   ├── index.css                   # Tailwind directives + design tokens
│   ├── types/
│   │   ├── api.ts                  # HealthResponse, Overview, HostsPage, History, Timeline, Snapshot, Analysis
│   │   └── domain.ts               # UI-only types: NavItem, Severity, HostStatus
│   ├── config/
│   │   └── env.ts                  # Centralized env reading (VITE_API_URL, VITE_USE_MOCK_DATA, refresh interval)
│   ├── lib/
│   │   ├── queryClient.ts          # TanStack Query client factory
│   │   ├── cn.ts                   # classname merge util (clsx + tailwind-merge)
│   │   └── format.ts               # bytes, percent, duration, relative-time formatters
│   ├── services/
│   │   ├── http.ts                 # fetch wrapper: base URL, JSON, typed errors, query-string builder
│   │   ├── api/
│   │   │   ├── health.ts           # useHealth() query
│   │   │   ├── dashboard.ts        # useOverview(), useHosts(), useHistory()
│   │   │   ├── replay.ts           # useReplay()
│   │   │   ├── timemachine.ts      # useSnapshot()
│   │   │   └── ai.ts               # useAnalyzeIncident() mutation
│   │   └── mock/
│   │       ├── index.ts            # shouldUseMock(queryKey) — dev mode + backend down/empty check
│   │       ├── dashboard.ts        # mock Overview + Hosts + History
│   │       ├── hosts.ts            # mock host fleet data
│   │       ├── events.ts           # mock event timeline
│   │       ├── replay.ts           # mock replay timeline
│   │       └── ai.ts               # mock AI analysis response
│   ├── hooks/
│   │   ├── useClock.ts             # live ticking clock for top bar
│   │   ├── useTheme.ts             # dark/light toggle (dark default)
│   │   └── useDebounce.ts          # search input debouncer
│   ├── layouts/
│   │   ├── AppLayout.tsx           # sidebar + topbar + <Outlet/>
│   │   ├── Sidebar.tsx             # logo, nav menu, collapse
│   │   └── TopBar.tsx              # search, clock, backend status badge, theme toggle, avatar
│   ├── components/
│   │   ├── ui/                     # primitives (reusable, no domain logic)
│   │   │   ├── Card.tsx
│   │   │   ├── Button.tsx
│   │   │   ├── Badge.tsx           # status pills (active/stale/healthy/critical)
│   │   │   ├── Skeleton.tsx
│   │   │   ├── Spinner.tsx
│   │   │   ├── EmptyState.tsx
│   │   │   └── ErrorState.tsx
│   │   ├── charts/
│   │   │   ├── AreaChartCard.tsx   # CPU/Memory time series
│   │   │   ├── GaugeCard.tsx       # single-percent radial gauge
│   │   │   └── Sparkline.tsx       # mini inline trend
│   │   ├── dashboard/
│   │   │   ├── FleetOverviewCards.tsx
│   │   │   ├── LiveInfrastructureCharts.tsx
│   │   │   ├── HostTable.tsx
│   │   │   ├── RecentEventsTimeline.tsx
│   │   │   └── AIInsightsCard.tsx
│   │   └── timemachine/
│   │       ├── TimelineSlider.tsx  # the scrubber
│   │       ├── ReplayControls.tsx  # prev/play/pause/next
│   │       └── SnapshotComparison.tsx
│   └── pages/
│       ├── DashboardPage.tsx
│       ├── HostsPage.tsx
│       ├── HostDetailPage.tsx
│       ├── EventsPage.tsx
│       ├── ReplayPage.tsx
│       ├── TimeMachinePage.tsx
│       ├── AIPage.tsx
│       └── SettingsPage.tsx
```

---

### Design System (index.css tokens)

Dark-first palette, violet accent:
- `--bg-base: #0a0a0f` (near-black canvas)
- `--bg-surface: #13131a` (card surface)
- `--bg-elevated: #1c1c26` (hover/elevated)
- `--border: #27272f`
- `--accent: #7c3aed` (violet-600), `--accent-bright: #8b5cf6`
- `--text-primary: #f1f5f9`, `--text-secondary: #94a3b8`, `--text-muted: #64748b`
- Status: emerald (healthy), amber (degraded), rose (critical), slate (stale)

Cards: `rounded-2xl`, `border border-[--border]`, `bg-[--bg-surface]`, `shadow-lg shadow-black/20`, hover lift via Framer Motion.

---

### Data Layer Strategy (the key decision)

**`services/mock/index.ts`** exports `withMockFallback<T>(queryKey, realFetcher, mockGenerator)`:
1. Always tries `realFetcher()` first.
2. In **dev** (`VITE_USE_MOCK_DATA=true`): if fetch throws (backend down) OR response is empty (`hosts.length === 0`, etc.), return `mockGenerator()` instead.
3. In **prod**: never intercepts — errors and empty states propagate normally.
4. UI hooks (`useHosts`, etc.) call `withMockFallback` — the pages never know.

Each query hook uses a **consistent refresh interval** (default 15s, configurable in Settings + `VITE_REFRESH_INTERVAL_MS`).

---

### Routing (React Router)

```
/                    → redirect to /dashboard
/dashboard           → DashboardPage
/hosts               → HostsPage
/hosts/:hostname     → HostDetailPage
/events              → EventsPage
/replay              → ReplayPage  (host selector + timeline)
/replay/:hostname    → ReplayPage  (preselected host)
/time-machine        → TimeMachinePage
/ai                  → AIPage
/settings            → SettingsPage
```

Lazy-loaded page chunks via `React.lazy` + `Suspense` skeleton fallback.

---

### Key Pages

1. **Dashboard** — 6 fleet cards, 2 area charts (CPU/Memory fleet averages), host table, events timeline, AI insights placeholder card. All wired to `useOverview()` + `useHosts()`.

2. **Hosts** — filterable host grid/table, click → `/hosts/:hostname`.

3. **Host Detail** — `useHistory(hostname)` → CPU/Memory/Disk charts, status badge, event history list, "Open in Time Machine" button.

4. **Events** — chronological event list with type filter chips, severity coloring.

5. **Replay** — host selector, `useReplay(hostname)` → scrollable event timeline. Cursor pagination when `next_cursor` present.

6. **Time Machine (flagship)** — host selector + `at` datetime picker → `useSnapshot()`. Timeline slider scrubs through snapshots. Prev/Play/Pause/Next controls. Snapshot comparison (two side-by-side metric cards). Play auto-advances through time steps with Framer Motion transitions.

7. **AI** — incident form (hostname, time range) → `useAnalyzeIncident()` mutation. Chat-style response panel showing summary, severity badge, probable causes, evidence, recommended actions, confidence bar. Analysis history (local state, persisted to localStorage).

8. **Settings** — theme toggle, API URL display, refresh interval slider, version. Read from `config/env.ts`.

---

### State & Loading

- **TanStack Query** for all server state. `staleTime: 10s`, `refetchInterval: 15s` on dashboard/health.
- **Loading**: Skeleton loaders matching each card's shape.
- **Empty**: `EmptyState` component with icon + message.
- **Error**: `ErrorState` with retry button calling `refetch()`.
- **Backend status**: top bar badge polls `/api/v1/health` every 15s — emerald "Operational" / rose "Disconnected".

---

### Files to generate: ~55

Tooling (7): package.json, vite.config.ts, tsconfig.json, tsconfig.node.json, index.html, .env.example, src/index.css
Core (5): main.tsx, App.tsx, config/env.ts, lib/queryClient.ts, lib/cn.ts, lib/format.ts
Types (2): types/api.ts, types/domain.ts
Services (8): services/http.ts + 5 api/ files + services/mock/index.ts
Mock (6): mock/{index,dashboard,hosts,events,replay,ai}.ts
Hooks (3): useClock, useTheme, useDebounce
Layouts (3): AppLayout, Sidebar, TopBar
UI components (7): Card, Button, Badge, Skeleton, Spinner, EmptyState, ErrorState
Charts (3): AreaChartCard, GaugeCard, Sparkline
Dashboard components (5): FleetOverviewCards, LiveInfrastructureCharts, HostTable, RecentEventsTimeline, AIInsightsCard
TimeMachine components (3): TimelineSlider, ReplayControls, SnapshotComparison
Pages (8): Dashboard, Hosts, HostDetail, Events, Replay, TimeMachine, AI, Settings

Everything compiles via `tsc && vite build`. No TODOs, no placeholders.