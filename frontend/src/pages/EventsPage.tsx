import { useMemo, useState } from 'react';
import { motion } from 'framer-motion';
import { Badge } from '@/components/ui/Badge';
import { Card } from '@/components/ui/Card';
import { EmptyState } from '@/components/ui/EmptyState';
import { ErrorState } from '@/components/ui/ErrorState';
import { Skeleton } from '@/components/ui/Skeleton';
import { useEvents } from '@/services/api/events';
import { formatRelativeTime } from '@/lib/format';
import type { Event } from '@/types/api';

const FILTERS = ['all', 'cpu', 'memory', 'disk', 'host', 'network'] as const;
type Filter = (typeof FILTERS)[number];

function matchesFilter(event: Event, filter: Filter): boolean {
  if (filter === 'all') return true;
  return event.type.toLowerCase().includes(filter);
}

function severityVariant(type: string): 'info' | 'active' | 'critical' {
  if (type.includes('critical')) return 'critical';
  if (type.includes('warning') || type.includes('spike')) return 'active';
  return 'info';
}

export default function EventsPage() {
  const [filter, setFilter] = useState<Filter>('all');
  const { data, isLoading, isError, error, refetch } = useEvents(50);

  const backendEvents: Event[] = data?.events ?? [];
  const events = useMemo(
    () => backendEvents.filter((e) => matchesFilter(e, filter)),
    [backendEvents, filter],
  );

  function renderPayload(payload: unknown): string {
    if (typeof payload === 'string') return payload;
    try {
      return JSON.stringify(payload, null, 2);
    } catch {
      return String(payload);
    }
  }

  return (
    <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} className="space-y-6">
      <div>
        <h1 className="text-xl font-bold text-slate-100">Events</h1>
        <p className="text-sm text-slate-500">Chronological infrastructure event stream</p>
      </div>

      <div className="flex flex-wrap gap-2">
        {FILTERS.map((f) => (
          <button
            key={f}
            onClick={() => setFilter(f)}
            className={`rounded-lg px-3 py-1.5 text-xs font-medium capitalize transition-colors ${
              filter === f ? 'bg-accent text-white' : 'border border-line bg-surface text-slate-400 hover:text-slate-200'
            }`}
          >
            {f}
          </button>
        ))}
      </div>

      {isError && (
        <ErrorState message={error?.message ?? 'Failed to load events'} onRetry={refetch} />
      )}

      {isLoading ? (
        <Card padding={false}>
          {Array.from({ length: 8 }).map((_, i) => (
            <div key={i} className="flex items-start gap-4 p-4">
              <Skeleton className="mt-2 h-2 w-2 rounded-full bg-accent" />
              <div className="min-w-0 flex-1 space-y-2">
                <Skeleton className="h-4 w-24" />
                <Skeleton className="h-3 w-40" />
              </div>
            </div>
          ))}
        </Card>
      ) : events.length === 0 ? (
        <EmptyState title="No events" description={
          backendEvents.length === 0 
            ? 'No events have been collected yet. Ensure agents are running and sending data.' 
            : 'No events match this filter.'
        } />
      ) : (
        <Card padding={false}>
          <div className="divide-y divide-line">
            {events.map((event, i) => (
              <motion.div
                key={`${event.id}-${event.recorded_at}`}
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                transition={{ delay: Math.min(i * 0.02, 0.3) }}
                className="flex items-start gap-4 p-4 transition-colors hover:bg-elevated/40"
              >
                <div className="mt-1.5 h-2 w-2 shrink-0 rounded-full bg-accent" />
                <div className="min-w-0 flex-1">
                  <div className="flex items-center gap-2 flex-wrap">
                    <Badge variant={severityVariant(event.type)}>
                      {event.type.split('.').pop()}
                    </Badge>
                    <span className="font-mono text-sm text-slate-300">{event.subject_id}</span>
                  </div>
                  <p className="mt-1 text-xs text-slate-500">
                    {new Date(event.occurred_at).toLocaleString('en-US')}
                  </p>
                  {event.payload != null && (
                    <pre className="mt-2 overflow-x-auto rounded-lg bg-base p-2 text-xs text-slate-500">
                      {String(renderPayload(event.payload))}
                    </pre>
                  )}
                </div>
                <span className="shrink-0 text-xs text-slate-500">
                  {formatRelativeTime(event.occurred_at)}
                </span>
              </motion.div>
            ))}
          </div>
        </Card>
      )}
    </motion.div>
  );
}