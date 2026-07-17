import { motion } from 'framer-motion';
import { useState, useEffect, useRef } from 'react';
import { useSearchParams } from 'react-router-dom';
import { Card } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { EmptyState } from '@/components/ui/EmptyState';
import { Spinner } from '@/components/ui/Spinner';
import { ErrorState } from '@/components/ui/ErrorState';
import { TimelineSlider } from '@/components/timemachine/TimelineSlider';
import { ReplayControls } from '@/components/timemachine/ReplayControls';
import { SnapshotComparison } from '@/components/timemachine/SnapshotComparison';
import { useHosts } from '@/services/api/dashboard';
import { useSnapshot } from '@/services/api/timemachine';
import { Clock } from 'lucide-react';

const STEP_MS = 60_000;

export default function TimeMachinePage() {
  const [params] = useSearchParams();
  const { data: hostsData } = useHosts();
  const hostnames = (hostsData?.hosts ?? []).map((h) => h.metrics.hostname);

  const [hostname, setHostname] = useState(params.get('hostname') ?? '');
  const now = Date.now();
  const [position, setPosition] = useState(now - 3600_000);
  const [isPlaying, setIsPlaying] = useState(false);
  const playRef = useRef<ReturnType<typeof setInterval> | null>(null);

  const min = now - 7200_000;
  const max = now;

  const atIso = new Date(position).toISOString();
  const { data: snapshot, isLoading, isError, error, refetch } = useSnapshot(hostname, atIso);

  useEffect(() => {
    if (isPlaying) {
      playRef.current = setInterval(() => {
        setPosition((p) => {
          if (p + STEP_MS >= max) {
            setIsPlaying(false);
            return max;
          }
          return p + STEP_MS;
        });
      }, 1000);
    }
    return () => {
      if (playRef.current) clearInterval(playRef.current);
    };
  }, [isPlaying, max]);

  return (
    <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} className="space-y-6">
      <div className="flex items-center gap-3">
        <Clock className="h-6 w-6 text-accent" />
        <div>
          <h1 className="text-xl font-bold text-slate-100">Time Machine</h1>
          <p className="text-sm text-slate-500">Reconstruct point-in-time infrastructure state</p>
        </div>
      </div>

      <Card>
        <label className="mb-2 block text-xs font-medium text-slate-500">Select Host</label>
        <select
          value={hostname}
          onChange={(e) => setHostname(e.target.value)}
          className="h-9 w-full max-w-xs rounded-lg border border-line bg-base px-3 text-sm text-slate-200 focus:border-accent focus:outline-none"
        >
          <option value="">Choose a host...</option>
          {hostnames.map((h) => (
            <option key={h} value={h}>{h}</option>
          ))}
        </select>
      </Card>

      {!hostname ? (
        <EmptyState title="Select a host" description="Choose a host to travel back through its infrastructure history." />
      ) : (
        <>
          <Card>
            <h3 className="mb-4 text-sm font-medium text-slate-200">Timeline</h3>
            <TimelineSlider min={min} max={max} value={position} onChange={(v) => { setPosition(v); setIsPlaying(false); }} />
            <div className="mt-6">
              <ReplayControls
                isPlaying={isPlaying}
                onPlay={() => setIsPlaying(true)}
                onPause={() => setIsPlaying(false)}
                onPrev={() => { setPosition((p) => Math.max(min, p - STEP_MS)); setIsPlaying(false); }}
                onNext={() => { setPosition((p) => Math.min(max, p + STEP_MS)); setIsPlaying(false); }}
                canPrev={position > min}
                canNext={position < max}
              />
            </div>
          </Card>

          <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
            {isLoading ? (
              <div className="flex justify-center py-16"><Spinner /></div>
            ) : isError ? (
              <ErrorState title="No snapshot" message={error?.message ?? 'No data at this point in time'} onRetry={refetch} />
            ) : snapshot ? (
              <SnapshotComparison snapshot={snapshot} />
            ) : (
              <EmptyState title="No data" />
            )}

            <Card>
              <h3 className="mb-4 text-sm font-medium text-slate-200">Observed At</h3>
              <div className="space-y-3">
                <div className="flex items-center justify-between">
                  <span className="text-sm text-slate-400">Requested</span>
                  <span className="font-mono text-sm text-slate-200">{new Date(position).toLocaleString('en-US')}</span>
                </div>
                {snapshot && (
                  <div className="flex items-center justify-between">
                    <span className="text-sm text-slate-400">Observed</span>
                    <span className="font-mono text-sm text-slate-200">{new Date(snapshot.observed_at).toLocaleString('en-US')}</span>
                  </div>
                )}
                {snapshot && (
                  <div className="flex items-center justify-between">
                    <span className="text-sm text-slate-400">Event ID</span>
                    <span className="font-mono text-sm text-accent-bright">#{snapshot.event_id}</span>
                  </div>
                )}
              </div>
              <div className="mt-4">
                <Button variant="ghost" size="sm" onClick={() => setPosition(min)}>Reset to start</Button>
              </div>
            </Card>
          </div>
        </>
      )}
    </motion.div>
  );
}
