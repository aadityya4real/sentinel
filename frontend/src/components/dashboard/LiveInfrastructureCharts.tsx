import { CardSkeleton } from '@/components/ui/Skeleton';
import { ErrorState } from '@/components/ui/ErrorState';
import { AreaChartCard } from '@/components/charts/AreaChartCard';
import type { Metrics } from '@/types/api';

interface Props {
  metrics: Metrics[];
  isLoading: boolean;
  isError: boolean;
  error: Error | null;
  refetch: () => void;
}

function toChartData(metrics: Metrics[], field: 'cpu_usage_percent' | 'memory'): { timestamp: string; value: number }[] {
  return metrics.map((m) => ({
    timestamp: new Date(m.timestamp).toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' }),
    value: field === 'cpu_usage_percent' ? m.cpu_usage_percent : m.memory.used_percent,
  }));
}

export function LiveInfrastructureCharts({ metrics, isLoading, isError, error, refetch }: Props) {
  if (isLoading) {
    return (
      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        <CardSkeleton />
        <CardSkeleton />
      </div>
    );
  }

  if (isError || metrics.length === 0) {
    return <ErrorState message={error?.message ?? 'No metrics data available'} onRetry={refetch} />;
  }

  const cpuData = toChartData(metrics, 'cpu_usage_percent');
  const memData = toChartData(metrics, 'memory');

  return (
    <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
      <AreaChartCard title="CPU Usage" data={cpuData} color="#7c3aed" unit="%" />
      <AreaChartCard title="Memory Usage" data={memData} color="#8b5cf6" unit="%" />
    </div>
  );
}
