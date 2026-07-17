import { Cpu, MemoryStick, Server, Globe, AlertTriangle } from 'lucide-react';
import { motion } from 'framer-motion';
import { CardSkeleton } from '@/components/ui/Skeleton';
import { ErrorState } from '@/components/ui/ErrorState';
import { formatPercent } from '@/lib/format';
import type { Overview } from '@/types/api';

interface StatCardProps {
  label: string;
  value: string;
  icon: React.ReactNode;
  accent?: string;
}

function StatCard({ label, value, icon, accent = 'text-accent' }: StatCardProps) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 4 }}
      animate={{ opacity: 1, y: 0 }}
      className="card p-5 flex items-start justify-between"
    >
      <div>
        <p className="text-xs font-medium text-slate-500 uppercase tracking-wider">{label}</p>
        <p className="mt-2 text-2xl font-bold text-slate-100">{value}</p>
      </div>
      <div className={`rounded-xl bg-elevated p-2.5 ${accent}`}>
        {icon}
      </div>
    </motion.div>
  );
}

interface FleetOverviewCardsProps {
  data: Overview | undefined;
  isLoading: boolean;
  isError: boolean;
  error: Error | null;
  refetch: () => void;
}

export function FleetOverviewCards({ data, isLoading, isError, error, refetch }: FleetOverviewCardsProps) {
  if (isLoading) {
    return (
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
        {Array.from({ length: 4 }).map((_, i) => (
          <CardSkeleton key={i} />
        ))}
      </div>
    );
  }

  if (isError || !data) {
    return <ErrorState message={error?.message} onRetry={refetch} />;
  }

  const degraded = data.total_hosts - data.active_hosts;

  return (
    <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
      <StatCard label="Active Hosts" value={String(data.active_hosts)} icon={<Server className="h-5 w-5" />} accent="text-emerald-400" />
      <StatCard label="CPU Average" value={formatPercent(data.average_cpu_usage_percent)} icon={<Cpu className="h-5 w-5" />} />
      <StatCard label="Memory Average" value={formatPercent(data.average_memory_usage_percent)} icon={<MemoryStick className="h-5 w-5" />} />
      {degraded > 0 ? (
        <StatCard label="Degraded" value={String(degraded)} icon={<AlertTriangle className="h-5 w-5" />} accent="text-amber-400" />
      ) : (
        <StatCard label="Total Hosts" value={String(data.total_hosts)} icon={<Globe className="h-5 w-5" />} />
      )}
    </div>
  );
}
