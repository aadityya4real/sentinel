import { motion } from 'framer-motion';
import { useState } from 'react';
import { Badge } from '@/components/ui/Badge';
import { Card } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { EmptyState } from '@/components/ui/EmptyState';
import { Spinner } from '@/components/ui/Spinner';
import { useHosts } from '@/services/api/dashboard';
import { useReplay } from '@/services/api/replay';
import { formatRelativeTime } from '@/lib/format';

export default function ReplayPage() {
  const { data: hostsData } = useHosts();
  const [hostname, setHostname] = useState('');
  const [cursor, setCursor] = useState<string | undefined>();
  const { data, isLoading, isFetching } = useReplay(hostname, cursor);

  const hostnames = (hostsData?.hosts ?? []).map((h) => h.metrics.hostname);

  return (
    <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} className="space-y-6">
      <div>
        <h1 className="text-xl font-bold text-slate-100">Replay</h1>
        <p className="text-sm text-slate-500">Replay infrastructure event timelines</p>
      </div>

      <Card>
        <label className="mb-2 block text-xs font-medium text-slate-500">Select Host</label>
        <select
          value={hostname}
          onChange={(e) => {
            setHostname(e.target.value);
            setCursor(undefined);
          }}
          className="h-9 w-full max-w-xs rounded-lg border border-line bg-base px-3 text-sm text-slate-200 focus:border-accent focus:outline-none"
        >
          <option value="">Choose a host...</option>
          {hostnames.map((h) => (
            <option key={h} value={h}>{h}</option>
          ))}
        </select>
      </Card>

      {!hostname ? (
        <EmptyState title="Select a host" description="Choose a host above to replay its event timeline." />
      ) : isLoading ? (
        <div className="flex justify-center py-16"><Spinner /></div>
      ) : data && data.events.length > 0 ? (
        <>
          <Card padding={false}>
            <div className="relative pl-8">
              <div className="absolute left-[19px] top-0 h-full w-px bg-line" />
              <div className="divide-y divide-line">
                {data.events.map((event) => (
                  <motion.div
                    key={event.id}
                    initial={{ opacity: 0, x: -8 }}
                    animate={{ opacity: 1, x: 0 }}
                    className="relative p-4"
                  >
                    <div className="absolute -left-[21px] top-5 h-3 w-3 rounded-full border-2 border-accent bg-base" />
                    <div className="flex items-center gap-2 flex-wrap">
                      <Badge variant="info">{event.type.split('.').pop()}</Badge>
                      <span className="font-mono text-sm text-slate-300">{event.subject_id}</span>
                    </div>
                    <p className="mt-1 text-xs text-slate-500">
                      {new Date(event.occurred_at).toLocaleString('en-US')} · {formatRelativeTime(event.occurred_at)}
                    </p>
                  </motion.div>
                ))}
              </div>
            </div>
          </Card>

          {data.next_cursor && (
            <div className="flex justify-center">
              <Button variant="secondary" size="sm" onClick={() => setCursor(data.next_cursor)} disabled={isFetching}>
                {isFetching ? <Spinner size="sm" /> : null}
                Load More
              </Button>
            </div>
          )}
        </>
      ) : (
        <EmptyState title="No events" description={`No events found for ${hostname} in this time range.`} />
      )}
    </motion.div>
  );
}
