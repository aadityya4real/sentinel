import { motion } from 'framer-motion';
import { useOverview, useHosts } from '@/services/api/dashboard';
import { useMetricStream } from '@/hooks/useMetricStream';
import { FleetOverviewCards } from '@/components/dashboard/FleetOverviewCards';
import { LiveInfrastructureCharts } from '@/components/dashboard/LiveInfrastructureCharts';
import { HostTable } from '@/components/dashboard/HostTable';
import { RecentEventsTimeline } from '@/components/dashboard/RecentEventsTimeline';
import { AIInsightsCard } from '@/components/dashboard/AIInsightsCard';
import { StreamStatusBadge } from '@/components/dashboard/StreamStatusBadge';

export default function DashboardPage() {
  const overview = useOverview();
  const hosts = useHosts();
  const { state, attempts, buffer } = useMetricStream();
  const fleetMetrics = (hosts.data?.hosts ?? []).map((h) => h.metrics);

  return (
    <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <div className="flex items-center gap-3">
            <h1 className="text-xl font-bold text-slate-100">Dashboard</h1>
            <StreamStatusBadge state={state} attempts={attempts} />
          </div>
          <p className="text-sm text-slate-500">Real-time infrastructure overview</p>
        </div>
      </div>

      <FleetOverviewCards
        data={overview.data}
        isLoading={overview.isLoading}
        isError={overview.isError}
        error={overview.error}
        refetch={overview.refetch}
      />

      <LiveInfrastructureCharts
        metrics={fleetMetrics.length > 0 ? [fleetMetrics[0]] : []}
        stream={buffer}
        isLoading={hosts.isLoading}
        isError={hosts.isError}
        error={hosts.error}
        refetch={hosts.refetch}
      />

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
        <div className="lg:col-span-2">
          <HostTable hosts={hosts.data?.hosts ?? []} isLoading={hosts.isLoading} />
        </div>
        <div className="space-y-6">
          <RecentEventsTimeline />
          <AIInsightsCard />
        </div>
      </div>
    </motion.div>
  );
}
