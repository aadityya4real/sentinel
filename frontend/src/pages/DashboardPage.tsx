import { useMemo } from 'react';
import { motion } from 'framer-motion';
import { Cpu, MemoryStick, Server, Activity, AlertTriangle } from 'lucide-react';
import { useOverview, useHosts } from '@/services/api/dashboard';
import { useMetricStream } from '@/hooks/useMetricStream';
import { mockEvents } from '@/services/mock/events';
import { MetricCard } from '@/components/dashboard/MetricCard';
import { InfChart } from '@/components/dashboard/InfChart';
import { RecentEventsTimeline } from '@/components/dashboard/RecentEventsTimeline';
import { HostTable } from '@/components/dashboard/HostTable';
import { StreamStatusBadge } from '@/components/dashboard/StreamStatusBadge';
import type { Metrics } from '@/types/api';

/* ── sparkline trend helper ─────────────────────────────────────── */
function extractTrend(buffer: Metrics[], field: (m: Metrics) => number) {
  if (buffer.length < 20) return { direction: 'flat' as const, value: 0 };
  const recent = buffer.slice(-10).map(field);
  const older = buffer.slice(-20, -10).map(field);
  const avgOld = older.reduce((s, v) => s + v, 0) / older.length;
  const avgNew = recent.reduce((s, v) => s + v, 0) / recent.length;
  const diff = ((avgNew - avgOld) / Math.max(avgOld, 1)) * 100;
  const dir: 'up' | 'down' | 'flat' = Math.abs(diff) < 2 ? 'flat' : diff > 0 ? 'up' : 'down';
  return { direction: dir, value: Math.round(Math.abs(diff)) };
}

/* ── dashboard page ─────────────────────────────────────────────── */
export default function DashboardPage() {
  const { data: overview, isLoading: oLoading, isError: oError, error: oErr, refetch: refetchOverview } = useOverview();
  const { data: hostsData, isLoading: hLoading, isError: hError, error: hErr, refetch: refetchHosts } = useHosts();
  const { state, attempts, buffer } = useMetricStream();

  const connected = state === 'connected';
  const hosts = hostsData?.hosts ?? [];
  const events = useMemo(() => mockEvents(20), []);

  const activeCount = hosts.filter((h) => h.status === 'active').length;
  const criticalCount = hosts.filter(
    (h) => h.metrics.cpu_usage_percent > 80 || h.metrics.memory.used_percent > 90,
  ).length;

  const trendCPU = extractTrend(buffer, (m) => m.cpu_usage_percent);
  const trendMem = extractTrend(buffer, (m) => m.memory.used_percent);

  const metricCards = [
    {
      label: 'Healthy Hosts',
      value: String(activeCount),
      icon: Server,
      sparklineData: Array.from({ length: 15 }, () => activeCount + (Math.random() - 0.5) * 2),
      trend: { direction: 'flat' as const, value: 0 },
      accentColor: 'text-emerald-400',
    },
    {
      label: 'Critical Hosts',
      value: String(criticalCount),
      icon: AlertTriangle,
      sparklineData: Array.from({ length: 15 }, () => Math.max(0, criticalCount + (Math.random() - 0.5) * 3)),
      critical: criticalCount > 0,
      accentColor: 'text-rose-400',
    },
    {
      label: 'CPU Average',
      value: overview ? `${Math.round(overview.average_cpu_usage_percent)}%` : '—',
      icon: Cpu,
      sparklineData: buffer.slice(-15).map((m) => m.cpu_usage_percent),
      trend: trendCPU,
      accentColor: 'text-accent-bright',
    },
    {
      label: 'Memory Average',
      value: overview ? `${Math.round(overview.average_memory_usage_percent)}%` : '—',
      icon: MemoryStick,
      sparklineData: buffer.slice(-15).map((m) => m.memory.used_percent),
      trend: trendMem,
      accentColor: 'text-violet-400',
    },
    {
      label: 'Events Today',
      value: String(events.length),
      icon: Activity,
      sparklineData: events.slice(-15).map((_, i) => 3 + i + Math.random() * 4),
      accentColor: 'text-sky-400',
    },
  ];

  /* chart data from live stream */
  const cpuData = buffer.map((m) => ({
    timestamp: new Date(m.timestamp).toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' }),
    value: m.cpu_usage_percent,
  }));
  const memData = buffer.map((m) => ({
    timestamp: new Date(m.timestamp).toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' }),
    value: m.memory.used_percent,
  }));
  const diskData = buffer.map((m) => ({
    timestamp: new Date(m.timestamp).toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' }),
    value: m.disks[0]?.used_percent ?? 0,
  }));
  const netData = buffer.map((_, i) => ({
    timestamp: new Date(Date.now() - (buffer.length - i) * 60_000)
      .toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' }),
    value: Math.floor(20 + Math.random() * 40),
  }));

  const loading = oLoading || hLoading;
  const anyError = oError || hError;

  return (
    <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} className="space-y-6">
      {/* Hero */}
      <div className="flex items-start justify-between">
        <div>
          <h1 className="text-xl font-bold text-slate-100">Infrastructure Overview</h1>
          <p className="mt-1 text-sm text-slate-500">Monitor, replay and analyze infrastructure in real time.</p>
        </div>
        <StreamStatusBadge state={connected ? 'connected' : state} attempts={attempts} />
      </div>

      {/* Error banner */}
      {anyError && !loading && (
        <div className="rounded-lg border border-rose-500/30 bg-rose-500/5 p-3 text-xs text-rose-400 flex items-center gap-2">
          <span>{oErr?.message ?? hErr?.message ?? 'Error loading data'}</span>
          <button onClick={() => { refetchOverview(); refetchHosts(); }} className="ml-auto underline cursor-pointer hover:text-rose-300 transition-colors">
            Retry
          </button>
        </div>
      )}

      {/* Metric Cards */}
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-5">
        {metricCards.map((card, i) => (
          <MetricCard key={card.label} data={card} isLoading={loading} index={i} />
        ))}
      </div>

      {/* Charts */}
      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        <InfChart title="CPU Usage" data={cpuData} color="#7c3aed" unit="%" isLoading={loading} />
        <InfChart title="Memory Usage" data={memData} color="#8b5cf6" unit="%" isLoading={loading} />
        <InfChart title="Disk Usage" data={diskData} color="#06b6d4" unit="%" isLoading={loading} />
        <InfChart title="Network Throughput" data={netData} color="#10b981" unit="Mbps" isLoading={loading} />
      </div>

      {/* Events + Host Table */}
      <div className="grid grid-cols-1 gap-6 xl:grid-cols-3">
        <div className="xl:col-span-1">
          <RecentEventsTimeline events={events} isLoading={loading} />
        </div>
        <div className="xl:col-span-2">
          <HostTable hosts={hosts} isLoading={loading} />
        </div>
      </div>
    </motion.div>
  );
}



