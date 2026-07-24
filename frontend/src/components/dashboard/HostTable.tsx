import { useState, useMemo } from 'react';
import { Link } from 'react-router-dom';
import { motion } from 'framer-motion';
import { ChevronUp, ChevronDown, ExternalLink, Search } from 'lucide-react';
import { cn } from '@/lib/cn';
import { Badge } from '@/components/ui/Badge';
import { TableSkeleton } from '@/components/ui/Skeleton';
import { EmptyState } from '@/components/ui/EmptyState';
import { Card } from '@/components/ui/Card';
import { formatPercent, formatRelativeTime } from '@/lib/format';
import type { HostSnapshot } from '@/types/api';

type SortKey = 'hostname' | 'cpu_usage_percent' | 'memory.used_percent' | 'status';
type SortDir = 'asc' | 'desc';

interface Props {
  hosts: HostSnapshot[];
  isLoading: boolean;
}

export function HostTable({ hosts, isLoading }: Props) {
  const [query, setQuery] = useState('');
  const [sortKey, setSortKey] = useState<SortKey>('hostname');
  const [sortDir, setSortDir] = useState<SortDir>('asc');

  const filteredAndSorted = useMemo(() => {
    let result = [...hosts];

    if (query.trim()) {
      const q = query.toLowerCase().trim();
      result = result.filter((h) => h.metrics.hostname.toLowerCase().includes(q));
    }

    result.sort((a, b) => {
      let av: string | number;
      let bv: string | number;

      switch (sortKey) {
        case 'hostname':
          av = a.metrics.hostname.toLowerCase();
          bv = b.metrics.hostname.toLowerCase();
          break;
        case 'cpu_usage_percent':
          av = a.metrics.cpu_usage_percent;
          bv = b.metrics.cpu_usage_percent;
          break;
        case 'memory.used_percent':
          av = a.metrics.memory.used_percent;
          bv = b.metrics.memory.used_percent;
          break;
        case 'status':
          av = a.status === 'active' ? 1 : 0;
          bv = b.status === 'active' ? 1 : 0;
          break;
        default:
          return 0;
      }

      if (typeof av === 'string' && typeof bv === 'string') {
        return sortDir === 'asc' ? av.localeCompare(bv) : bv.localeCompare(av);
      }
      return sortDir === 'asc' ? Number(av) - Number(bv) : Number(bv) - Number(av);
    });

    return result;
  }, [hosts, query, sortKey, sortDir]);

  const handleSort = (key: SortKey) => {
    if (sortKey === key) {
      setSortDir((d) => (d === 'asc' ? 'desc' : 'asc'));
    } else {
      setSortKey(key);
      setSortDir('asc');
    }
  };

  const statusVariant = (s: string): 'healthy' | 'critical' | 'stale' => {
    if (s === 'stale') return 'stale';
    return 'healthy';
  };

  if (isLoading) {
    return (
      <Card>
        <div className="px-5 py-3 border-b border-line">
          <TableSkeleton rows={6} />
        </div>
      </Card>
    );
  }

  if (hosts.length === 0) {
    return (
      <Card>
        <EmptyState title="No hosts found" description="Hosts will appear here once agents start reporting." />
      </Card>
    );
  }

  return (
    <Card className="overflow-hidden">
      {/* header row with search */}
      <div className="flex items-center justify-between px-5 py-3 border-b border-line">
        <h3 className="text-sm font-medium text-slate-200">Hosts ({filteredAndSorted.length})</h3>
        <div className="relative">
          <Search className="absolute left-2.5 top-1/2 h-3.5 w-3.5 -translate-y-1/2 text-slate-500" />
          <input
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Filter hosts..."
            className="h-7 w-48 rounded-lg border border-line bg-base pl-7 pr-2.5 text-xs text-slate-200 placeholder:text-slate-500 focus:border-accent focus:outline-none"
          />
        </div>
      </div>

      <div className="overflow-x-auto">
        <table className="w-full text-left text-sm">
          <thead>
            <tr className="border-b border-line">
              <SortableHeader label="Hostname" onClick={() => handleSort('hostname')} active={sortKey === 'hostname'}>
                <SortIcon col="hostname" current={sortKey} dir={sortDir} />
              </SortableHeader>
              <th className="px-5 py-2.5 text-xs font-medium text-slate-500 uppercase tracking-wider">OS</th>
              <SortableHeader label="CPU" onClick={() => handleSort('cpu_usage_percent')} active={sortKey === 'cpu_usage_percent'}>
                <SortIcon col="cpu_usage_percent" current={sortKey} dir={sortDir} />
              </SortableHeader>
              <SortableHeader label="Memory" onClick={() => handleSort('memory.used_percent')} active={sortKey === 'memory.used_percent'}>
                <SortIcon col="memory.used_percent" current={sortKey} dir={sortDir} />
              </SortableHeader>
              <SortableHeader label="Status" onClick={() => handleSort('status')} active={sortKey === 'status'}>
                <SortIcon col="status" current={sortKey} dir={sortDir} />
              </SortableHeader>
              <th className="px-5 py-2.5 text-xs font-medium text-slate-500 uppercase tracking-wider">Last Seen</th>
              <th className="px-5 py-2.5" />
            </tr>
          </thead>
          <tbody>
            {filteredAndSorted.map((host, i) => (
              <motion.tr
                key={host.metrics.hostname}
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                transition={{ delay: Math.min(i * 0.02, 0.2) }}
                className="group border-b border-line/50 last:border-0 hover:bg-elevated/50 transition-colors"
              >
                <td className="px-5 py-2.5 font-medium text-slate-200">{host.metrics.hostname}</td>
                <td className="px-5 py-2.5 text-slate-400">{host.metrics.os}</td>
                <td className="px-5 py-2.5">
                  <span className={cn('tabular-nums', host.metrics.cpu_usage_percent > 80 ? 'text-rose-400' : 'text-slate-300')}>
                    {formatPercent(host.metrics.cpu_usage_percent)}
                  </span>
                </td>
                <td className="px-5 py-2.5">
                  <span className={cn('tabular-nums', host.metrics.memory.used_percent > 85 ? 'text-amber-400' : 'text-slate-300')}>
                    {formatPercent(host.metrics.memory.used_percent)}
                  </span>
                </td>
                <td className="px-5 py-2.5">
                  <Badge variant={statusVariant(host.status)}>{host.status === 'active' ? 'Active' : 'Stale'}</Badge>
                </td>
                <td className="px-5 py-2.5 text-slate-500 tabular-nums">{formatRelativeTime(host.metrics.timestamp)}</td>
                <td className="px-5 py-2.5">
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
    </Card>
  );
}

/* ── helpers ─────────────────────────────────────── */

interface SortIconProps {
  col: SortKey;
  current: SortKey;
  dir: SortDir;
}

function SortIcon({ col, current, dir }: SortIconProps) {
  if (current !== col) return <div className="h-4 w-4" />;
  return dir === 'asc' ? <ChevronUp className="h-3.5 w-3.5" /> : <ChevronDown className="h-3.5 w-3.5" />;
}

interface SortableHeaderProps {
  label: string;
  children?: React.ReactNode;
  onClick: () => void;
  active: boolean;
}

function SortableHeader({ label, children, onClick, active }: SortableHeaderProps) {
  return (
    <th
      className="px-5 py-2.5 text-xs font-medium text-slate-500 uppercase tracking-wider cursor-pointer select-none hover:text-slate-300 transition-colors"
      onClick={onClick}
    >
      <span className="inline-flex items-center gap-1">
        {label}
        <span className={active ? 'text-accent-bright' : 'text-slate-600'}>{children}</span>
      </span>
    </th>
  );
}
