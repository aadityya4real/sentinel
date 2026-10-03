import { useMemo } from 'react';
import { motion } from 'framer-motion';
import { Cpu, MemoryStick, Server, Activity, AlertTriangle } from 'lucide-react';
import { useOverview, useHosts, useHistory } from '@/services/api/dashboard';
import { useEvents } from '@/services/api/events';
import { getRefreshInterval } from '@/services/api/refresh';
import { useMetricStream } from '@/hooks/useMetricStream';
import { MetricCard } from '@/components/dashboard/MetricCard';
import { InfChart } from '@/components/dashboard/InfChart';
import { RecentEventsTimeline } from '@/components/dashboard/RecentEventsTimeline';
import { HostTable } from '@/components/dashboard/HostTable';
import { StreamStatusBadge } from '@/components/dashboard/StreamStatusBadge';

/* ── helpers ─────────────────────────────────────────── */

interface ChartPoint {
  timestamp: string;
  value: number;
}

function calcTrend(before: number, after: number) {
  if (before <= 0) return { direction: 'flat' as const, value: 0 };
  const diff = ((after - before) / before) * 100;
  const dir = Math.abs(diff) < 2 ? ('flat' as const) : diff > 0 ? ('up' as const) : ('down' as const);
  return { direction: dir, value: Math.round(Math.abs(diff)) };
}

/* ── page ────────────────────────────────────────────── */

export default function DashboardPage() {
  const { data: overview, isLoading: oLoading, isError: oError, error: oErr, refetch: refetchOverview } = useOverview();
  const { data: hostsData, isLoading: hLoading, isError: hError, error: hErr, refetch: refetchHosts } = useHosts();
  const { data: eventData, isLoading: eLoading, isError: eError, error: eErr, refetch: refetchEvents } = useEvents(20);
  const { state: wsState, attempts, buffer } = useMetricStream();
  const events = eventData?.events ?? [];

  const connected = wsState === 'connected';
  const hosts = hostsData?.hosts ?? [];
  const loading = oLoading || hLoading;
  const anyError = oError || hError || eError;
  const historyRefreshInterval = connected || wsState === 'mock' ? false : getRefreshInterval();

  /* ── history queries — called at top level so React Query can manage them ── */
  const firstFourHostnames = hosts.slice(0, 4).map((h) => h.metrics.hostname);

  // We always call hooks for a fixed set of hostnames to keep the hook call count stable.
  // If there are fewer than 4 hosts we alias them; extra ones resolve to empty strings
  // and the query simply returns no results.
  const n1 = firstFourHostnames[0] ?? '';
  const n2 = firstFourHostnames[1] ?? '';
  const n3 = firstFourHostnames[2] ?? '';
  const n4 = firstFourHostnames[3] ?? '';

  const h1 = useHistory(n1, 100, historyRefreshInterval);
  const h2 = useHistory(n2, 100, historyRefreshInterval);
  const h3 = useHistory(n3, 100, historyRefreshInterval);
  const h4 = useHistory(n4, 100, historyRefreshInterval);

  /* ── prefers WebSocket buffer when available ── */
  const hasBuffer = buffer.length > 2;

  /* ── build chart data from real history API or WS buffer ── */
  const cpuData = useMemo<ChartPoint[]>(() => {
    if (hasBuffer && buffer.length > 2) {
      return buffer.map((m) => ({
        timestamp: new Date(m.timestamp).toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' }),
        value: m.cpu_usage_percent,
      }));
    }
    let points: ChartPoint[] = [];
    for (const hist of [h1.data, h2.data, h3.data, h4.data]) {
      if (!hist?.metrics?.length) continue;
      for (const m of hist.metrics) {
        points.push({
          timestamp: new Date(m.timestamp).toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' }),
          value: m.cpu_usage_percent,
        });
      }
    }
    return points;
  }, [hasBuffer, buffer, h1.data, h2.data, h3.data, h4.data]);

  const memData = useMemo<ChartPoint[]>(() => {
    if (hasBuffer && buffer.length > 2) {
      return buffer.map((m) => ({
        timestamp: new Date(m.timestamp).toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' }),
        value: m.memory.used_percent,
      }));
    }
    let points: ChartPoint[] = [];
    for (const hist of [h1.data, h2.data, h3.data, h4.data]) {
      if (!hist?.metrics?.length) continue;
      for (const m of hist.metrics) {
        points.push({
          timestamp: new Date(m.timestamp).toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' }),
          value: m.memory.used_percent,
        });
      }
    }
    return points;
  }, [hasBuffer, buffer, h1.data, h2.data, h3.data, h4.data]);

  const diskData = useMemo<ChartPoint[]>(() => {
    if (hasBuffer && buffer.length > 2) {
      return buffer.filter((m) => m.disks.length > 0).map((m) => ({
        timestamp: new Date(m.timestamp).toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' }),
        value: m.disks[0].used_percent,
      }));
    }
    let points: ChartPoint[] = [];
    for (const hist of [h1.data, h2.data, h3.data, h4.data]) {
      if (!hist?.metrics?.length) continue;
      for (const m of hist.metrics) {
        if (!m.disks[0]) continue;
        points.push({
          timestamp: new Date(m.timestamp).toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' }),
          value: m.disks[0].used_percent,
        });
      }
    }
    return points;
  }, [hasBuffer, buffer, h1.data, h2.data, h3.data, h4.data]);

  /* ── sparklines & trends ── */
  const displayCpuSpark = hasBuffer
    ? buffer.slice(-15).map((m) => m.cpu_usage_percent)
    : cpuData.map((p) => p.value);
  const displayMemSpark = hasBuffer
    ? buffer.slice(-15).map((m) => m.memory.used_percent)
    : memData.map((p) => p.value);

  const latestCpu = hasBuffer
    ? buffer[buffer.length - 1].cpu_usage_percent
    : (displayCpuSpark[displayCpuSpark.length - 1] ?? hosts[0]?.metrics.cpu_usage_percent);
  const prevCpu = hasBuffer && buffer.length > 10
    ? buffer[buffer.length - 11].cpu_usage_percent
    : (displayCpuSpark[0] ?? latestCpu);
  const latestMem = hasBuffer
    ? buffer[buffer.length - 1].memory.used_percent
    : (displayMemSpark[displayMemSpark.length - 1] ?? hosts[0]?.metrics.memory.used_percent);
  const prevMem = hasBuffer && buffer.length > 10
    ? buffer[buffer.length - 11].memory.used_percent
    : (displayMemSpark[0] ?? latestMem);
  const trendCPU = prevCpu === undefined || latestCpu === undefined ? undefined : calcTrend(prevCpu, latestCpu);
  const trendMem = prevMem === undefined || latestMem === undefined ? undefined : calcTrend(prevMem, latestMem);

  /* ── card values ── */
  const activeCount = hosts.filter((h) => h.status === 'active').length;
  const staleCount = hosts.filter((h) => h.status === 'stale').length;
  const cpuAverage = overview?.average_cpu_usage_percent ?? (hosts.length ? hosts.reduce((sum, host) => sum + host.metrics.cpu_usage_percent, 0) / hosts.length : undefined);
  const memoryAverage = overview?.average_memory_usage_percent ?? (hosts.length ? hosts.reduce((sum, host) => sum + host.metrics.memory.used_percent, 0) / hosts.length : undefined);
  const cpuValue = oError && hosts.length === 0 ? 'Unavailable' : cpuAverage === undefined ? 'No telemetry' : `${Math.round(cpuAverage)}%`;
  const memoryValue = oError && hosts.length === 0 ? 'Unavailable' : memoryAverage === undefined ? 'No telemetry' : `${Math.round(memoryAverage)}%`;
  const eventValue = eError ? 'Unavailable' : String(events.length);
  const activeValue = hError && !hostsData ? 'Unavailable' : String(activeCount);
  const staleValue = hError && !hostsData ? 'Unavailable' : String(staleCount);
  const historyLoading = [h1, h2, h3, h4].some((history) => history.isLoading);

  return (
    <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} className="space-y-6">
      <div className="flex items-start justify-between">
        <div>
          <h1 className="text-xl font-bold text-slate-100">Infrastructure Overview</h1>
          <p className="mt-1 text-sm text-slate-500">Monitor, replay and analyze infrastructure in real time.</p>
        </div>
        <StreamStatusBadge state={connected ? 'connected' : wsState} attempts={attempts} />
      </div>

      {anyError && !loading && (
        <div className="rounded-lg border border-rose-500/30 bg-rose-500/5 p-3 text-xs text-rose-400 flex items-center gap-2">
          <span>{oErr?.message ?? hErr?.message ?? eErr?.message ?? 'Error loading data'}</span>
          <button onClick={() => { refetchOverview(); refetchHosts(); refetchEvents(); }} className="ml-auto underline cursor-pointer hover:text-rose-300 transition-colors">Retry</button>
        </div>
      )}

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-5">
        <MetricCard data={{ label: 'Active Hosts', value: activeValue, icon: Server, sparklineData: [], accentColor: 'text-emerald-400' }} isLoading={loading} index={0} />
        <MetricCard data={{ label: 'Stale Hosts', value: staleValue, icon: AlertTriangle, sparklineData: [], critical: staleCount > 0, accentColor: 'text-rose-400' }} isLoading={loading} index={1} />
        <MetricCard data={{ label: 'CPU Average', value: cpuValue, icon: Cpu, sparklineData: displayCpuSpark, trend: displayCpuSpark.length > 1 ? trendCPU : undefined, accentColor: 'text-accent-bright' }} isLoading={loading} index={2} />
        <MetricCard data={{ label: 'Memory Average', value: memoryValue, icon: MemoryStick, sparklineData: displayMemSpark, trend: displayMemSpark.length > 1 ? trendMem : undefined, accentColor: 'text-violet-400' }} isLoading={loading} index={3} />
        <MetricCard data={{ label: 'Recent Events', value: eventValue, icon: Activity, sparklineData: [], accentColor: 'text-sky-400' }} isLoading={eLoading} index={4} />
      </div>

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        <InfChart title="CPU Usage" data={cpuData} color="#7c3aed" unit="%" isLoading={loading || (historyLoading && cpuData.length === 0 && !hasBuffer)} emptyMessage={hError && !hostsData ? 'Telemetry unavailable' : 'No data yet'} />
        <InfChart title="Memory Usage" data={memData} color="#8b5cf6" unit="%" isLoading={loading || (historyLoading && memData.length === 0 && !hasBuffer)} emptyMessage={hError && !hostsData ? 'Telemetry unavailable' : 'No data yet'} />
        <InfChart title="Disk Usage" data={diskData} color="#06b6d4" unit="%" isLoading={loading || (historyLoading && diskData.length === 0 && !hasBuffer)} emptyMessage={hError && !hostsData ? 'Telemetry unavailable' : 'No data yet'} />
        <InfChart title="Network Throughput" data={[]} color="#10b981" unit="Mbps" emptyMessage="No network telemetry collected" />
      </div>

      <div className="grid grid-cols-1 gap-6 xl:grid-cols-3">
        <div className="xl:col-span-1">
          <RecentEventsTimeline events={events} isLoading={eLoading} isUnavailable={eError && !eventData} />
        </div>
        <div className="xl:col-span-2">
          {hError && !hostsData ? <div className="card p-5 text-sm text-rose-400">Host data unavailable</div> : <HostTable hosts={hosts} isLoading={loading} />}
        </div>
      </div>
    </motion.div>
  );
}
