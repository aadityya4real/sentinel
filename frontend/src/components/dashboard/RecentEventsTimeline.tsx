import { motion } from 'framer-motion';
import { Badge } from '@/components/ui/Badge';
import { EmptyState } from '@/components/ui/EmptyState';
import type { Event } from '@/types/api';

interface TimelineEntryProps {
  event: Event;
  index: number;
}

interface RecentEventsTimelineProps {
  events: Event[];
  isLoading?: boolean;
  isUnavailable?: boolean;
}

function formatEventType(type: string) {
  const parts = type.split('.').slice(-2);
  return parts.map((p) => p.charAt(0).toUpperCase() + p.slice(1)).join(' ');
}

export function TimelineEntry({ event, index }: TimelineEntryProps) {
  const color = '#7c3aed';
  const description = formatEventType(event.type);

  return (
    <motion.div
      initial={{ opacity: 0, x: -4 }}
      animate={{ opacity: 1, x: 0 }}
      transition={{ delay: index * 0.03 }}
      className="group flex items-start gap-3 py-2.5 last:pb-0"
    >
      <div className="flex flex-col items-center">
        <div className="mt-1.5 h-2.5 w-2.5 rounded-full shrink-0 ring-4 ring-surface" style={{ backgroundColor: color }} />
      </div>
      <div className="min-w-0 flex-1">
        <div className="flex items-center gap-2 flex-wrap">
          <Badge variant="info">{event.type}</Badge>
          <span className="text-xs font-mono text-text-muted">{event.subject_id}</span>
        </div>
        <p className="mt-0.5 text-sm text-text-primary">{description}</p>
        <p className="mt-0.5 text-xs text-text-muted">
          {new Date(event.occurred_at).toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' })}
        </p>
      </div>
      <span className="hidden sm:block text-xs text-text-muted tabular-nums group-hover:text-text-secondary transition-colors">
        {event.key}
      </span>
    </motion.div>
  );
}

export function RecentEventsTimeline({ events, isLoading, isUnavailable = false }: RecentEventsTimelineProps) {
  if (isLoading) {
    return <RecentEventsSkeleton />;
  }

  if (!events.length && isUnavailable) {
    return (
      <div className="card p-5">
        <h3 className="mb-3 text-sm font-medium text-text-primary">Recent Events</h3>
        <p className="text-sm text-rose-400">Events unavailable</p>
      </div>
    );
  }

  if (!events.length) {
    return (
      <div className="card p-5">
        <EmptyState title="No events" description="Events will appear here as agents start reporting." />
      </div>
    );
  }

  return (
    <div className="card p-5">
      <h3 className="mb-3 text-sm font-medium text-text-primary">Recent Events</h3>
      <div className="max-h-[320px] overflow-y-auto pr-1 scrollbar-thin">
        {events.map((event, i) => (
          <TimelineEntry key={event.id} event={event} index={i} />
        ))}
      </div>
    </div>
  );
}

function RecentEventsSkeleton() {
  return (
    <div className="card p-5 space-y-4">
      <div className="h-4 w-24 rounded bg-elevated animate-shimmer" style={{ backgroundImage: 'linear-gradient(90deg, transparent, rgba(139,92,246,0.06), transparent)', backgroundSize: '200% 100%' }} />
      {Array.from({ length: 5 }).map((_, i) => (
        <div key={i} className="flex items-start gap-3">
          <div className="mt-2 h-2.5 w-2.5 rounded-full bg-elevated" />
          <div className="flex-1 space-y-2">
            <div className="h-3 w-16 rounded bg-elevated animate-shimmer" style={{ backgroundImage: 'linear-gradient(90deg, transparent, rgba(139,92,246,0.06), transparent)', backgroundSize: '200% 100%' }} />
            <div className="h-3 w-40 rounded bg-elevated animate-shimmer" style={{ backgroundImage: 'linear-gradient(90deg, transparent, rgba(139,92,246,0.06), transparent)', backgroundSize: '200% 100%' }} />
          </div>
        </div>
      ))}
    </div>
  );
}
