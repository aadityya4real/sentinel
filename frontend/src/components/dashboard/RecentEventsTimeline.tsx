import { Badge } from '@/components/ui/Badge';
import { EmptyState } from '@/components/ui/EmptyState';
import { mockEvents } from '@/services/mock/events';
import { formatRelativeTime } from '@/lib/format';
import type { Event } from '@/types/api';

function eventTypeVariant(type: string): 'info' | 'active' | 'critical' {
  if (type.includes('critical') || type.includes('warning')) return 'critical';
  if (type.includes('started')) return 'active';
  return 'info';
}

interface EventRowProps {
  event: Event;
}

function EventRow({ event }: EventRowProps) {
  return (
    <div className="flex items-start gap-3 py-2">
      <div className="mt-1.5 h-2 w-2 rounded-full bg-accent shrink-0" />
      <div className="min-w-0 flex-1">
        <div className="flex items-center gap-2 flex-wrap">
          <Badge variant={eventTypeVariant(event.type)}>{event.type.split('.').pop()}</Badge>
          <span className="text-xs text-slate-400">{event.subject_id}</span>
        </div>
        <p className="mt-0.5 text-xs text-slate-500">{formatRelativeTime(event.occurred_at)}</p>
      </div>
    </div>
  );
}

export function RecentEventsTimeline() {
  const events = mockEvents(8);

  if (events.length === 0) {
    return <EmptyState title="No events yet" description="Infrastructure events will appear here as agents report." />;
  }

  return (
    <div className="card p-5">
      <h3 className="mb-4 text-sm font-medium text-slate-200">Recent Events</h3>
      <div className="divide-y divide-line/50">
        {events.map((event) => (
          <EventRow key={event.id} event={event} />
        ))}
      </div>
    </div>
  );
}
