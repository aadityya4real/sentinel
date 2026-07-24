import { motion } from 'framer-motion';
import { CardSkeleton } from '@/components/ui/Skeleton';
import { ErrorState } from '@/components/ui/ErrorState';
import { AreaChartCard } from '@/components/charts/AreaChartCard';
import type { Metrics } from '@/types/api';

interface Props {
  metrics: Metrics[];
  stream: Metrics[];
  isLoading: boolean;
  isError: boolean;
  error: Error | null;
  refetch: () => void;
}

function toChartData(metrics: Metrics[], field: 'cpu_usage_percent' | 'memory'): { timestamp: string; value: number }[] {
  return metrics.map((m) => ({
    timestamp: new Date(m.timestamp).toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit', second: '2-digit' }),
    value: field === 'cpu_usage_percent' ? m.cpu_usage_percent : m.memory.used_percent,
  }));
}

/**
 * Merges REST and stream samples, preferring fresh stream points while falling
 * back to REST history when the stream is empty (e.g. on first load).
 */
function mergeSeries(rest: Metrics[], stream: Metrics[]): Metrics[] {
  if (stream.length === 0) return rest;
  if (rest.length === 0) return stream;
  const streamStart = new Date(stream[0].timestamp).getTime();
  const older = rest.filter((m) => new Date(m.timestamp).getTime() < streamStart);
  return [...older, ...stream];
}

export function LiveInfrastructureCharts({ metrics, stream, isLoading, isError, error, refetch }: Props) {
  if (isLoading && metrics.length === 0 && stream.length === 0) {
    return (
      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        <CardSkeleton />
        <CardSkeleton />
      </div>
    );
  }

  if (isError && metrics.length === 0 && stream.length === 0) {
    return <ErrorState message={error?.message ?? 'No metrics data available'} onRetry={refetch} />;
  }

  const merged = mergeSeries(metrics, stream).slice(-60);
  const cpuData = toChartData(merged, 'cpu_usage_percent');
  const memData = toChartData(merged, 'memory');

  return (
    <motion.div
      key={stream.length}
      initial={{ opacity: 0.85 }}
      animate={{ opacity: 1 }}
      transition={{ duration: 0.4 }}
      className="grid grid-cols-1 gap-4 lg:grid-cols-2"
    >
      <AreaChartCard title="CPU Usage" data={cpuData} color="#7c3aed" unit="%" />
      <AreaChartCard title="Memory Usage" data={memData} color="#8b5cf6" unit="%" />
    </motion.div>
  );
}
