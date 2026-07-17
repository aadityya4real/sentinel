import { cn } from '@/lib/cn';

type BadgeVariant = 'active' | 'stale' | 'healthy' | 'degraded' | 'critical' | 'info';

const variants: Record<BadgeVariant, string> = {
  active: 'bg-emerald-500/15 text-emerald-400 ring-emerald-500/30',
  healthy: 'bg-emerald-500/15 text-emerald-400 ring-emerald-500/30',
  degraded: 'bg-amber-500/15 text-amber-400 ring-amber-500/30',
  stale: 'bg-slate-500/15 text-slate-400 ring-slate-500/30',
  critical: 'bg-rose-500/15 text-rose-400 ring-rose-500/30',
  info: 'bg-violet-500/15 text-violet-400 ring-violet-500/30',
};

interface BadgeProps {
  variant: BadgeVariant;
  children: React.ReactNode;
  className?: string;
}

export function Badge({ variant, children, className }: BadgeProps) {
  return (
    <span
      className={cn(
        'inline-flex items-center gap-1.5 rounded-full px-2.5 py-0.5 text-xs font-medium ring-1 ring-inset',
        variants[variant],
        className,
      )}
    >
      <span className="h-1.5 w-1.5 rounded-full bg-current" />
      {children}
    </span>
  );
}
