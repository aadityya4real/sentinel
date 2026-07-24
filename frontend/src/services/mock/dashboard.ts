import type { Overview, HostsPage, History, Metrics } from '@/types/api';

function recentTimestamp(offsetSeconds = 0): string {
  return new Date(Date.now() - offsetSeconds * 1000).toISOString();
}

function generateMetrics(hostname: string, cpu: number, mem: number, offset = 0): Metrics {
  return {
    cpu_usage_percent: cpu,
    memory: { total_bytes: 16_384_000_000, used_bytes: Math.round(16_384_000_000 * (mem / 100)), used_percent: mem, available_bytes: 0 },
    disks: [
      { path: '/', filesystem: 'ext4', total_bytes: 100_000_000_000, used_bytes: Math.round(100_000_000_000 * (mem / 120)), used_percent: mem / 1.2 },
      { path: '/data', filesystem: 'xfs', total_bytes: 500_000_000_000, used_bytes: Math.round(500_000_000_000 * (mem / 100)), used_percent: mem },
    ],
    hostname,
    os: 'linux',
    uptime_seconds: 86_400 + offset,
    timestamp: recentTimestamp(offset),
  };
}

export { generateMetrics };

const hosts = [
  { name: 'prod-web-01', cpu: 42, mem: 68, status: 'active' as const },
  { name: 'prod-web-02', cpu: 55, mem: 72, status: 'active' as const },
  { name: 'prod-api-01', cpu: 78, mem: 85, status: 'active' as const },
  { name: 'prod-api-02', cpu: 31, mem: 45, status: 'active' as const },
  { name: 'prod-db-01', cpu: 65, mem: 91, status: 'active' as const },
  { name: 'staging-web-01', cpu: 12, mem: 34, status: 'stale' as const },
  { name: 'staging-api-01', cpu: 8, mem: 22, status: 'stale' as const },
  { name: 'dev-sandbox-01', cpu: 5, mem: 15, status: 'stale' as const },
];

export { hosts as mockHostsConfig };

export function mockOverview(): Overview {
  const active = hosts.filter((h) => h.status === 'active').length;
  return {
    total_hosts: hosts.length,
    active_hosts: active,
    active_within_seconds: 30,
    average_cpu_usage_percent: hosts.reduce((s, h) => s + h.cpu, 0) / hosts.length,
    average_memory_usage_percent: hosts.reduce((s, h) => s + h.mem, 0) / hosts.length,
    latest_metric_at: recentTimestamp(),
  };
}

export function mockHostsPage(): HostsPage {
  return {
    hosts: hosts.map((h) => ({
      metrics: generateMetrics(h.name, h.cpu, h.mem),
      status: h.status,
    })),
    limit: 100,
  };
}

export function mockHistory(hostname: string): History {
  const host = hosts.find((h) => h.name === hostname) || hosts[0];
  const metrics: Metrics[] = [];
  for (let i = 59; i >= 0; i--) {
    metrics.push(generateMetrics(host.name, host.cpu + (Math.random() - 0.5) * 20, host.mem + (Math.random() - 0.5) * 10, i * 60));
  }
  return {
    hostname: host.name,
    from: recentTimestamp(3600),
    to: recentTimestamp(),
    metrics,
    limit: 300,
  };
}

export function getMockHosts() {
  return hosts;
}
