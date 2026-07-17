import { motion } from 'framer-motion';
import { Cpu, MemoryStick, HardDrive, Clock } from 'lucide-react';
import { Card } from '@/components/ui/Card';
import { formatPercent, formatBytes, formatRelativeTime } from '@/lib/format';
import type { Snapshot } from '@/types/api';

interface SnapshotComparisonProps {
  snapshot: Snapshot;
}

function MetricRow({ icon, label, value, accent }: { icon: React.ReactNode; label: string; value: string; accent?: string }) {
  return (
    <div className="flex items-center justify-between py-2">
      <div className="flex items-center gap-2 text-slate-400">
        {icon}
        <span className="text-sm">{label}</span>
      </div>
      <span className={`font-mono text-sm ${accent ?? 'text-slate-200'}`}>{value}</span>
    </div>
  );
}

export function SnapshotComparison({ snapshot }: SnapshotComparisonProps) {
  const m = snapshot.metrics;
  return (
    <motion.div
      key={snapshot.event_id}
      initial={{ opacity: 0, scale: 0.98 }}
      animate={{ opacity: 1, scale: 1 }}
      transition={{ duration: 0.2 }}
    >
      <Card>
        <div className="mb-4 flex items-center justify-between">
          <div>
            <h3 className="text-sm font-medium text-slate-200">{snapshot.hostname}</h3>
            <p className="text-xs text-slate-500">Observed {formatRelativeTime(snapshot.observed_at)}</p>
          </div>
          <div className="flex items-center gap-1.5 rounded-lg bg-elevated px-2.5 py-1 text-xs text-slate-400">
            <Clock className="h-3 w-3" />
            {new Date(snapshot.observed_at).toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' })}
          </div>
        </div>

        <div className="divide-y divide-line">
          <MetricRow
            icon={<Cpu className="h-4 w-4 text-accent" />}
            label="CPU"
            value={formatPercent(m.cpu_usage_percent)}
            accent={m.cpu_usage_percent > 80 ? 'text-rose-400' : undefined}
          />
          <MetricRow
            icon={<MemoryStick className="h-4 w-4 text-accent" />}
            label="Memory"
            value={`${formatPercent(m.memory.used_percent)} (${formatBytes(m.memory.used_bytes)})`}
            accent={m.memory.used_percent > 85 ? 'text-amber-400' : undefined}
          />
          {m.disks.map((disk) => (
            <MetricRow
              key={disk.path}
              icon={<HardDrive className="h-4 w-4 text-accent" />}
              label={`Disk ${disk.path}`}
              value={`${formatPercent(disk.used_percent)} (${formatBytes(disk.used_bytes)})`}
            />
          ))}
        </div>
      </Card>
    </motion.div>
  );
}
