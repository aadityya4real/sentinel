import { useParams, Link } from 'react-router-dom';
import { motion } from 'framer-motion';
import { ArrowLeft, Clock } from 'lucide-react';
import { useHistory } from '@/services/api/dashboard';
import { Badge } from '@/components/ui/Badge';
import { Button } from '@/components/ui/Button';
import { CardSkeleton } from '@/components/ui/Skeleton';
import { ErrorState } from '@/components/ui/ErrorState';
import { EmptyState } from '@/components/ui/EmptyState';
import { AreaChartCard } from '@/components/charts/AreaChartCard';
import { Card } from '@/components/ui/Card';
import { formatPercent, formatBytes, formatUptime } from '@/lib/format';

function toChartData(
  metrics: { timestamp: string; cpu_usage_percent: number; memory: { used_percent: number } }[],
  field: 'cpu' | 'memory',
) {
  return metrics.map((m) => ({
    timestamp: new Date(m.timestamp).toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' }),
    value: field === 'cpu' ? m.cpu_usage_percent : m.memory.used_percent,
  }));
}

export default function HostDetailPage() {
  const { hostname } = useParams<{ hostname: string }>();
  const { data, isLoading, isError, error, refetch } = useHistory(hostname ?? '', 60);

  if (isLoading) {
    return (
      <div className="space-y-6">
        <CardSkeleton />
        <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
          <CardSkeleton />
          <CardSkeleton />
          <CardSkeleton />
        </div>
      </div>
    );
  }

  if (isError) return <ErrorState message={error?.message} onRetry={refetch} />;
  if (!data || data.metrics.length === 0) {
    return <EmptyState title="No history" description={`No metrics recorded for ${hostname}`} />;
  }

  const latest = data.metrics[data.metrics.length - 1];

  return (
    <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <Link to="/hosts" className="text-slate-400 hover:text-slate-200">
            <ArrowLeft className="h-5 w-5" />
          </Link>
          <div>
            <div className="flex items-center gap-3">
              <h1 className="text-xl font-bold text-slate-100">{data.hostname}</h1>
              <Badge variant="active">active</Badge>
            </div>
            <p className="text-sm text-slate-500">
              {latest.os} · Uptime {formatUptime(latest.uptime_seconds)}
            </p>
          </div>
        </div>
        <Link to={`/time-machine?hostname=${data.hostname}`}>
          <Button variant="secondary" size="sm">
            <Clock className="h-4 w-4" /> Time Machine
          </Button>
        </Link>
      </div>

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
        <AreaChartCard title="CPU Usage" data={toChartData(data.metrics, 'cpu')} color="#7c3aed" />
        <AreaChartCard title="Memory Usage" data={toChartData(data.metrics, 'memory')} color="#8b5cf6" />
      </div>

      <Card>
        <h3 className="mb-4 text-sm font-medium text-slate-200">Disks</h3>
        <div className="space-y-2">
          {latest.disks.map((disk) => (
            <div key={disk.path} className="flex items-center justify-between py-2">
              <div>
                <span className="font-mono text-sm text-slate-200">{disk.path}</span>
                <span className="ml-2 text-xs text-slate-500">{disk.filesystem}</span>
              </div>
              <div className="text-right">
                <span className="text-sm text-slate-300">
                  {formatBytes(disk.used_bytes)} / {formatBytes(disk.total_bytes)}
                </span>
                <span className="ml-2 text-xs text-slate-500">{formatPercent(disk.used_percent)}</span>
              </div>
            </div>
          ))}
        </div>
      </Card>
    </motion.div>
  );
}
