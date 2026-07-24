import type { LucideIcon } from 'lucide-react';
import { ArrowUp, ArrowDown, Minus } from 'lucide-react';
import { motion } from 'framer-motion';
import { cn } from '@/lib/cn';

interface SparklineProps {
  data: number[];
  color?: string;
}

/** Minimal SVG sparkline — no deps beyond lucide-react */
function Sparkline({ data, color = '#7c3aed' }: SparklineProps) {
  if (data.length < 2) return null;

  const min = Math.min(...data);
  const max = Math.max(...data);
  const range = max - min || 1;
  const width = 80;
  const height = 24;
  const pad = 2;

  const points = data.map((v, i) => ({
    x: pad + (i / (data.length - 1)) * (width - 2 * pad),
    y: height - pad - ((v - min) / range) * (height - 2 * pad),
  }));

  const pathD = points.map((p, i) => `${i === 0 ? 'M' : 'L'} ${p.x} ${p.y}`).join(' ');

  // area fill path
  const areaD = pathD + ` L ${points[points.length - 1].x} ${height} L ${points[0].x} ${height} Z`;

  return (
    <svg width={width} height={height} className="overflow-visible">
      <defs>
        <linearGradient id={`sg-${color.replace('#', '')}`} x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stopColor={color} stopOpacity={0.35} />
          <stop offset="100%" stopColor={color} stopOpacity={0} />
        </linearGradient>
      </defs>
      <path d={areaD} fill={`url(#sg-${color.replace('#', '')})`} />
      <path d={pathD} fill="none" stroke={color} strokeWidth={1.5} strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  );
}

export interface MetricCardValue {
  label: string;
  value: string | number;
  icon: LucideIcon;
  sparklineData: number[];
  trend?: { direction: 'up' | 'down' | 'flat'; value: number };
  accentColor?: string;
  critical?: boolean;
}

const TREND_COLORS = {
  up: 'text-emerald-400',
  down: 'text-amber-400',
  flat: 'text-slate-500',
};

function TrendArrow({ direction, value }: { direction: 'up' | 'down' | 'flat'; value: number }) {
  const Icon = direction === 'up' ? ArrowUp : direction === 'down' ? ArrowDown : Minus;
  const colorClass = TREND_COLORS[direction];
  return (
    <span className={cn('inline-flex items-center gap-0.5 text-xs font-medium', colorClass)}>
      <Icon className="h-3 w-3" />
      {direction === 'flat' ? '0%' : `${direction === 'down' ? '-' : '+'}${value}%`}
    </span>
  );
}

interface MetricCardProps {
  data: MetricCardValue;
  isLoading: boolean;
  index: number;
}

export function MetricCard({ data, isLoading, index }: MetricCardProps) {
  const Icon = data.icon;
  const accent = data.accentColor || 'text-accent-bright';
  const borderColor = data.critical ? 'border-rose-500/40 shadow-glow' : '';

  if (isLoading) {
    return (
      <motion.div
        initial={{ opacity: 0, y: 4 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: index * 0.06 }}
        className="card p-5 flex flex-col justify-between min-h-[96px]"
      >
        <div className="flex items-start justify-between">
          <div className="space-y-2">
            <div className="h-3 w-16 rounded bg-elevated animate-shimmer" style={{ backgroundImage: 'linear-gradient(90deg, transparent, rgba(139,92,246,0.06), transparent)', backgroundSize: '200% 100%' }} />
            <div className="h-7 w-20 rounded bg-elevated animate-shimmer" style={{ backgroundImage: 'linear-gradient(90deg, transparent, rgba(139,92,246,0.06), transparent)', backgroundSize: '200% 100%' }} />
          </div>
          <div className="rounded-xl bg-elevated h-10 w-10" />
        </div>
      </motion.div>
    );
  }

  return (
    <motion.div
      initial={{ opacity: 0, y: 4 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ delay: index * 0.06 }}
      className={cn('card p-5 flex flex-col justify-between min-h-[96px]', borderColor)}
    >
      <div className="flex items-start justify-between">
        <div>
          <p className="text-xs font-medium text-slate-500 uppercase tracking-wider">{data.label}</p>
          <div className="mt-1 flex items-baseline gap-2">
            <p className="text-2xl font-bold text-slate-100">{data.value}</p>
            {data.trend && <TrendArrow direction={data.trend.direction} value={data.trend.value} />}
          </div>
        </div>
        <div className={cn('rounded-xl bg-elevated p-2.5', accent)}>
          <Icon className="h-5 w-5" />
        </div>
      </div>
      <div className="mt-3">
        <Sparkline data={data.sparklineData} color={data.critical ? '#f43f5e' : '#7c3aed'} />
      </div>
    </motion.div>
  );
}
