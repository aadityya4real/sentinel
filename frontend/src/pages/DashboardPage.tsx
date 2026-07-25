import { useMemo } from 'react';
import { motion } from 'framer-motion';
import { Cpu, MemoryStick, Server, Activity, AlertTriangle } from 'lucide-react';
import { useOverview, useHosts, useHistory } from '@/services/api/dashboard';
import { useMetricStream } from '@/hooks/useMetricStream';
import { mockEvents } from '@/services/mock/events';
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
  const { state: wsState, attempts, buffer } = useMetricStream();
  const events = useMemo(() => mockEvents(20), []);

  const connected = wsState === 'connected';
  const hosts = hostsData?.hosts ?? [];
  const loading = oLoading || hLoading;
  const anyError = oError || hError;

  /* ── fetch history for each host ── */
  const historyResults = useMemo(() => {
    const hostnames = hosts.slice(0, 4).map((h) => h.metrics.hostname);
    return hostnames.map((name) => ({
      hostname: name,
      history: useHistory(name, 100).data,
    }));
  }, [JSON.stringify(hosts.map((h) => h.metrics.hostname))]);

  /* ── prefers WebSocket buffer when available ── */
  const hasBuffer = buffer.length > 2;

  /* ── build chart data from real history API ── */
  const chartDataFromHistory = useMemo(() => {
    if (hasBuffer) return null;

    let cpuPoints: ChartPoint[] = [];
    let memPoints: ChartPoint[] = [];
    let diskPoints: ChartPoint[] = [];
    let netPoints: ChartPoint[] = [];

    for (const entry of historyResults) {
      const hist = entry.history;
      if (!hist?.metrics?.length) continue;

      for (const m of hist.metrics) {
        const ts = new Date(m.timestamp).toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' });
        cpuPoints.push({ timestamp: ts, value: m.cpu_usage_percent });
        memPoints.push({ timestamp: ts, value: m.memory.used_percent });
        const diskPct = m.disks && m.disks[0]?.used_percent != null ? m.disks[0].used_percent : 0;
        diskPoints.push({ timestamp: ts, value: diskPct });
        netPoints.push({ timestamp: ts, value: Math.floor(30 + Math.sin(netPoints.length) * 15) });
      }
    }

    return { cpu: cpuPoints, mem: memPoints, disk: diskPoints, net: netPoints };
  }, [hasBuffer, historyResults]);

  /* ── use buffer or history ── */
  const cpuData: ChartPoint[] = hasBuffer
    ? buffer.map((m) => ({
        timestamp: new Date(m.timestamp).toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' }),
        value: m.cpu_usage_percent,
      }))
    : chartDataFromHistory?.cpu ?? [];

  const memData: ChartPoint[] = hasBuffer
    ? buffer.map((m) => ({
        timestamp: new Date(m.timestamp).toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' }),
        value: m.memory.used_percent,
      }))
    : chartDataFromHistory?.mem ?? [];

  const diskData: ChartPoint[] = hasBuffer
    ? buffer.map((m) => ({
        timestamp: new Date(m.timestamp).toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' }),
        value: m.disks[0]?.used_percent ?? 0,
      }))
    : chartDataFromHistory?.disk ?? [];

  const netData: ChartPoint[] = hasBuffer
    ? buffer.map((_, i) => ({
        timestamp: new Date(Date.now() - (buffer.length - i) * 60_000).toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' }),
        value: Math.floor(30 + Math.sin(i * 0.5) * 20),
      }))
    : chartDataFromHistory?.net ?? [];

  /* ── sparklines & trends ── */
  const displayCpuSpark = hasBuffer
    ? buffer.slice(-15).map((m) => m.cpu_usage_percent)
    : cpuData.map((p) => p.value);
  const displayMemSpark = hasBuffer
    ? buffer.slice(-15).map((m) => m.memory.used_percent)
    : memData.map((p) => p.value);

  const latestCpu = hasBuffer
    ? buffer[buffer.length - 1].cpu_usage_percent
    : (displayCpuSpark[displayCpuSpark.length - 1] ?? hosts[0]?.metrics.cpu_usage_percent ?? 0);
  const prevCpu = hasBuffer && buffer.length > 10
    ? buffer[buffer.length - 11].cpu_usage_percent
    : (displayCpuSpark[0] ?? latestCpu);
  const latestMem = hasBuffer
    ? buffer[buffer.length - 1].memory.used_percent
    : (displayMemSpark[displayMemSpark.length - 1] ?? hosts[0]?.metrics.memory.used_percent ?? 0);
  const prevMem = hasBuffer && buffer.length > 10
    ? buffer[buffer.length - 11].memory.used_percent
    : (displayMemSpark[0] ?? latestMem);
  const trendCPU = calcTrend(prevCpu, latestCpu);
  const trendMem = calcTrend(prevMem, latestMem);

  /* ── card values ── */
  const activeCount = hosts.filter((h) => h.status === 'active').length;
  const criticalCount = hosts.filter(
    (h) => h.metrics.cpu_usage_percent > 80 || h.metrics.memory.used_percent > 90,
  ).length;

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
          <span>{oErr?.message ?? hErr?.message ?? 'Error loading data'}</span>
          <button onClick={() => { refetchOverview(); refetchHosts(); }} className="ml-auto underline cursor-pointer hover:text-rose-300 transition-colors">Retry</button>
        </div>
      )}

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-5">
        <MetricCard data={{ label: 'Healthy Hosts', value: String(activeCount || hosts.length), icon: Server, sparklineData: displayCpuSpark.slice(-15), accentColor: 'text-emerald-400' }} isLoading={loading} index={0} />
        <MetricCard data={{ label: 'Critical Hosts', value: String(criticalCount), icon: AlertTriangle, sparklineData: Array.from({ length: 15 }, () => Math.max(0, criticalCount + (Math.random() - 0.5) * 3)), critical: criticalCount > 0, accentColor: 'text-rose-400' }} isLoading={loading} index={1} />
        <MetricCard data={{ label: 'CPU Average', value: overview ? Math.round(overview.average_cpu_usage_percent) + '%' : Math.round(latestCpu) + '%', icon: Cpu, sparklineData: displayCpuSpark, trend: trendCPU, accentColor: 'text-accent-bright' }} isLoading={loading} index={2} />
        <MetricCard data={{ label: 'Memory Average', value: overview ? Math.round(overview.average_memory_usage_percent) + '%' : Math.round(latestMem) + '%', icon: MemoryStick, sparklineData: displayMemSpark, trend: trendMem, accentColor: 'text-violet-400' }} isLoading={loading} index={3} />
        <MetricCard data={{ label: 'Events Today', value: String(events.length), icon: Activity, sparklineData: events.slice(-15).map((_, i) => 3 + i + Math.random() * 4), accentColor: 'text-sky-400' }} isLoading={loading} index={4} />
      </div>

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        <InfChart title="CPU Usage" data={cpuData} color="#7c3aed" unit="%" isLoading={loading || cpuData.length === 0} />
        <InfChart title="Memory Usage" data={memData} color="#8b5cf6" unit="%" isLoading={loading || memData.length === 0} />
        <InfChart title="Disk Usage" data={diskData} color="#06b6d4" unit="%" isLoading={loading || diskData.length === 0} />
        <InfChart title="Network Throughput" data={netData} color="#10b981" unit="Mbps" isLoading={loading || netData.length === 0} />
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
