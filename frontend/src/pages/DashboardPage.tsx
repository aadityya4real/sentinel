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
import type { HostSnapshot, Metrics } from '@/types/api';

/** Convert host snapshots into time-series data for charts */
function snapshotToSeries(hosts: HostSnapshot[], field: (m: Metrics) => number) {
  const now = Date.now();
  return hosts.slice(-30).reverse().map((h, i) => ({
    timestamp: new Date(now - (30 - i) * 60_000).toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' }),
    value: field(h.metrics),
  }));
}

/** Prefer buffer data; fall back to REST snapshots */
function buildChartSeries(buffer: Metrics[], hosts: HostSnapshot[], field: (m: Metrics) => number) {
  if (buffer.length > 2) {
    return buffer.map((m) => ({
      timestamp: new Date(m.timestamp).toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' }),
      value: field(m),
    }));
  }
  return snapshotToSeries(hosts, field);
}

/** Compute trend direction */
function calcTrend(before: number, after: number) {
  if (before <= 0) return { direction: 'flat' as const, value: 0 };
  const diff = ((after - before) / before) * 100;
  const dir = Math.abs(diff) < 2 ? ('flat' as const) : diff > 0 ? ('up' as const) : ('down' as const);
  return { direction: dir, value: Math.round(Math.abs(diff)) };
}

/** Generate stable network throughput data based on host metrics */
function generateNetworkData(hosts: HostSnapshot[]): { timestamp: string; value: number }[] {
  const now = Date.now();
  return hosts.slice(-30).reverse().map((_, i) => ({
    timestamp: new Date(now - (30 - i) * 60_000).toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' }),
    value: Math.floor(30 + Math.sin(i * 0.5) * 20),
  }));
}

export default function DashboardPage() {
  const { data: overview, isLoading: oLoading, isError: oError, error: oErr, refetch: refetchOverview } = useOverview();
  const { data: hostsData, isLoading: hLoading, isError: hError, error: hErr, refetch: refetchHosts } = useHosts();
  const { state, attempts, buffer } = useMetricStream();

  const connected = state === 'connected';
  const hosts = hostsData?.hosts ?? [];
  const events = useMemo(() => mockEvents(20), []);

  // Use stream data when available, otherwise REST data
  const displayHosts = buffer.length > 0
    ? buffer.map((m) => ({ metrics: m, status: 'active' as const }))
    : hosts;

  const activeCount = displayHosts.filter((h) => h.status === 'active').length;
  const criticalCount = displayHosts.filter(
    (h) => h.metrics.cpu_usage_percent > 80 || h.metrics.memory.used_percent > 90,
  ).length;

  const latestCpu = buffer.length > 0 ? buffer[buffer.length - 1].cpu_usage_percent : (displayHosts[0]?.metrics.cpu_usage_percent ?? 0);
  const prevCpu = buffer.length > 10 ? buffer[buffer.length - 11].cpu_usage_percent : latestCpu;
  const latestMem = buffer.length > 0 ? buffer[buffer.length - 1].memory.used_percent : (displayHosts[0]?.metrics.memory.used_percent ?? 0);
  const prevMem = buffer.length > 10 ? buffer[buffer.length - 11].memory.used_percent : latestMem;
  const trendCPU = calcTrend(prevCpu, latestCpu);
  const trendMem = calcTrend(prevMem, latestMem);

  const cpuSpark = buffer.slice(-20).map((m) => m.cpu_usage_percent);
  const memSpark = buffer.slice(-20).map((m) => m.memory.used_percent);

  const cpuData = buildChartSeries(buffer, displayHosts, (m) => m.cpu_usage_percent);
  const memData = buildChartSeries(buffer, displayHosts, (m) => m.memory.used_percent);
  const diskData = buildChartSeries(buffer, displayHosts, (m) => m.disks[0]?.used_percent ?? 0);
  const netData = buffer.length > 0
    ? buffer.map((_, i) => ({
        timestamp: new Date(Date.now() - (buffer.length - i) * 60_000).toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' }),
        value: Math.floor(30 + Math.sin(i * 0.5) * 20),
      }))
    : generateNetworkData(displayHosts);

  const loading = oLoading || hLoading;
  const anyError = oError || hError;

  return (
    <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} className="space-y-6">
      <div className="flex items-start justify-between">
        <div>
          <h1 className="text-xl font-bold text-slate-100">Infrastructure Overview</h1>
          <p className="mt-1 text-sm text-slate-500">Monitor, replay and analyze infrastructure in real time.</p>
        </div>
        <StreamStatusBadge state={connected ? 'connected' : state} attempts={attempts} />
      </div>

      {anyError && !loading && (
        <div className="rounded-lg border border-rose-500/30 bg-rose-500/5 p-3 text-xs text-rose-400 flex items-center gap-2">
          <span>{oErr?.message ?? hErr?.message ?? 'Error loading data'}</span>
          <button onClick={() => { refetchOverview(); refetchHosts(); }} className="ml-auto underline cursor-pointer hover:text-rose-300 transition-colors">Retry</button>
        </div>
      )}

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-5">
        <MetricCard data={{ label: 'Healthy Hosts', value: String(activeCount || hosts.length), icon: Server, sparklineData: cpuSpark.slice(-15), accentColor: 'text-emerald-400' }} isLoading={loading} index={0} />
        <MetricCard data={{ label: 'Critical Hosts', value: String(criticalCount), icon: AlertTriangle, sparklineData: Array.from({ length: 15 }, () => Math.max(0, criticalCount + (Math.random() - 0.5) * 3)), critical: criticalCount > 0, accentColor: 'text-rose-400' }} isLoading={loading} index={1} />
        <MetricCard data={{ label: 'CPU Average', value: overview ? Math.round(overview.average_cpu_usage_percent) + '%' : Math.round(latestCpu) + '%', icon: Cpu, sparklineData: cpuSpark.slice(-15), trend: trendCPU, accentColor: 'text-accent-bright' }} isLoading={loading} index={2} />
        <MetricCard data={{ label: 'Memory Average', value: overview ? Math.round(overview.average_memory_usage_percent) + '%' : Math.round(latestMem) + '%', icon: MemoryStick, sparklineData: memSpark.slice(-15), trend: trendMem, accentColor: 'text-violet-400' }} isLoading={loading} index={3} />
        <MetricCard data={{ label: 'Events Today', value: String(events.length), icon: Activity, sparklineData: events.slice(-15).map((_, i) => 3 + i + Math.random() * 4), accentColor: 'text-sky-400' }} isLoading={loading} index={4} />
      </div>

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        <InfChart title="CPU Usage" data={cpuData} color="#7c3aed" unit="%" isLoading={loading} />
        <InfChart title="Memory Usage" data={memData} color="#8b5cf6" unit="%" isLoading={loading} />
        <InfChart title="Disk Usage" data={diskData} color="#06b6d4" unit="%" isLoading={loading} />
        <InfChart title="Network Throughput" data={netData} color="#10b981" unit="Mbps" isLoading={loading} />
      </div>

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
