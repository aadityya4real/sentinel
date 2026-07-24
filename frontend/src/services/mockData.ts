// ──────────────────────────────────────────────────────
// Deterministic mock data for a fleet of hosts.
// Used by DashboardPage in mock mode so every card, chart,
// table and timeline renders with realistic data.
// ──────────────────────────────────────────────────────
import type { HostSnapshot, Metrics } from '@/types/api';

interface HostConfig {
  hostname: string;
  os: string;
  cpu: number;
  memory: number;
  disk: number;
  network: number;
  status: 'active' | 'stale';
}

const HOSTS: HostConfig[] = [
  { hostname: 'prod-web-01', os: 'linux', cpu: 42, memory: 68, disk: 55, network: 30, status: 'active' },
  { hostname: 'prod-web-02', os: 'linux', cpu: 55, memory: 72, disk: 60, network: 45, status: 'active' },
  { hostname: 'prod-api-01', os: 'linux', cpu: 78, memory: 85, disk: 72, network: 60, status: 'active' },
  { hostname: 'prod-db-01', os: 'linux', cpu: 65, memory: 91, disk: 82, network: 20, status: 'active' },
  { hostname: 'prod-cache-01', os: 'linux', cpu: 31, memory: 45, disk: 30, network: 55, status: 'active' },
  { hostname: 'staging-web-01', os: 'linux', cpu: 12, memory: 34, disk: 25, network: 10, status: 'stale' },
  { hostname: 'staging-api-01', os: 'linux', cpu: 8, memory: 22, disk: 20, network: 5, status: 'stale' },
  { hostname: 'dev-sandbox-01', os: 'linux', cpu: 5, memory: 15, disk: 10, network: 2, status: 'stale' },
];

function timestamp(offsetSeconds: number): string {
  return new Date(Date.now() - offsetSeconds * 1000).toISOString();
}

/** Generate a single Metrics snapshot using config + small noise */
function metricsFromConfig(h: HostConfig, offsetSec = 0): Metrics {
  const jitter = () => (Math.random() - 0.5) * 6;
  return {
    hostname: h.hostname,
    os: h.os,
    timestamp: timestamp(offsetSec),
    uptime_seconds: 86_400 + offsetSec,
    cpu_usage_percent: Math.min(100, Math.max(0, h.cpu + jitter())),
    memory: {
      total_bytes: 16_384_000_000,
      used_bytes: Math.round(16_384_000_000 * ((h.memory + jitter()) / 100)),
      used_percent: Math.min(100, Math.max(0, h.memory + jitter())),
      available_bytes: 0,
    },
    disks: [
      {
        path: '/',
        filesystem: 'ext4',
        total_bytes: 100_000_000_000,
        used_bytes: Math.round(100_000_000_000 * ((h.disk + jitter()) / 100)),
        used_percent: Math.min(100, Math.max(0, h.disk + jitter())),
      },
    ],
  };
}

/** Generate 30 time-series points for one host */
export function generateHistory(hostname: string): Metrics[] {
  const host = HOSTS.find((h) => h.hostname === hostname) || HOSTS[0];
  return Array.from({ length: 30 }, (_, i) => metricsFromConfig(host, 30 - i));
}

/** All host snapshots — latest sample only */
export function generateHostSnapshots(): HostSnapshot[] {
  return HOSTS.map((h) => ({ metrics: metricsFromConfig(h), status: h.status }));
}

/** Fleet summary numbers */
export function calculateOverview() {
  const snaps = generateHostSnapshots();
  const active = snaps.filter((s) => s.status === 'active').length;
  return {
    total_hosts: snaps.length,
    active_hosts: active,
    active_within_seconds: 30,
    average_cpu_usage_percent: Math.round((snaps.reduce((s, h) => s + h.metrics.cpu_usage_percent, 0) / snaps.length) * 10) / 10,
    average_memory_usage_percent: Math.round((snaps.reduce((s, h) => s + h.metrics.memory.used_percent, 0) / snaps.length) * 10) / 10,
    latest_metric_at: timestamp(0),
  };
}
