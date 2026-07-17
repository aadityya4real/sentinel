import { Link } from 'react-router-dom';
import { motion } from 'framer-motion';
import { ExternalLink } from 'lucide-react';
import { Badge } from '@/components/ui/Badge';
import { TableSkeleton } from '@/components/ui/Skeleton';
import { EmptyState } from '@/components/ui/EmptyState';
import { formatPercent, formatRelativeTime } from '@/lib/format';
import type { HostSnapshot } from '@/types/api';

interface Props {
  hosts: HostSnapshot[];
  isLoading: boolean;
}

export function HostTable({ hosts, isLoading }: Props) {
  if (isLoading) {
    return (
      <div className="card p-5">
        <TableSkeleton rows={6} />
      </div>
    );
  }

  if (hosts.length === 0) {
    return <EmptyState title="No hosts found" description="Hosts will appear here once agents start reporting." />;
  }

  return (
    <div className="card overflow-hidden">
      <div className="overflow-x-auto">
        <table className="w-full text-left text-sm">
          <thead>
            <tr className="border-b border-line">
              <th className="px-5 py-3 text-xs font-medium text-slate-500 uppercase tracking-wider">Hostname</th>
              <th className="px-5 py-3 text-xs font-medium text-slate-500 uppercase tracking-wider">OS</th>
              <th className="px-5 py-3 text-xs font-medium text-slate-500 uppercase tracking-wider">CPU</th>
              <th className="px-5 py-3 text-xs font-medium text-slate-500 uppercase tracking-wider">Memory</th>
              <th className="px-5 py-3 text-xs font-medium text-slate-500 uppercase tracking-wider">Status</th>
              <th className="px-5 py-3 text-xs font-medium text-slate-500 uppercase tracking-wider">Last Seen</th>
              <th className="px-5 py-3" />
            </tr>
          </thead>
          <tbody>
            {hosts.map((host, i) => (
              <motion.tr
                key={host.metrics.hostname}
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                transition={{ delay: i * 0.03 }}
                className="group border-b border-line/50 last:border-0 transition-colors hover:bg-elevated/50"
              >
                <td className="px-5 py-3 font-medium text-slate-200">{host.metrics.hostname}</td>
                <td className="px-5 py-3 text-slate-400">{host.metrics.os}</td>
                <td className="px-5 py-3">
                  <span className={host.metrics.cpu_usage_percent > 80 ? 'text-rose-400' : 'text-slate-300'}>
                    {formatPercent(host.metrics.cpu_usage_percent)}
                  </span>
                </td>
                <td className="px-5 py-3">
                  <span className={host.metrics.memory.used_percent > 85 ? 'text-amber-400' : 'text-slate-300'}>
                    {formatPercent(host.metrics.memory.used_percent)}
                  </span>
                </td>
                <td className="px-5 py-3">
                  <Badge variant={host.status === 'active' ? 'active' : 'stale'}>{host.status}</Badge>
                </td>
                <td className="px-5 py-3 text-slate-500">{formatRelativeTime(host.metrics.timestamp)}</td>
                <td className="px-5 py-3">
                  <Link
                    to={`/hosts/${host.metrics.hostname}`}
                    className="inline-flex items-center gap-1 text-xs text-accent opacity-0 transition-opacity group-hover:opacity-100"
                  >
                    View <ExternalLink className="h-3 w-3" />
                  </Link>
                </td>
              </motion.tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
