import { useId } from 'react';
import type { StreamState } from '@/services/stream/types';
import { cn } from '@/lib/cn';

interface StreamStatusBadgeProps {
  state: StreamState;
  attempts?: number;
}

const COPY: Record<StreamState, { label: string; dot: string; text: string; ring: string }> = {
  connected: { label: 'Live', dot: 'bg-emerald-400', text: 'text-emerald-400', ring: 'ring-emerald-500/30' },
  connecting: { label: 'Connecting', dot: 'bg-slate-400', text: 'text-slate-400', ring: 'ring-slate-500/30' },
  reconnecting: { label: 'Reconnecting', dot: 'bg-amber-400', text: 'text-amber-400', ring: 'ring-amber-500/30' },
  disconnected: { label: 'Disconnected', dot: 'bg-rose-400', text: 'text-rose-400', ring: 'ring-rose-500/30' },
};

export function StreamStatusBadge({ state, attempts = 0 }: StreamStatusBadgeProps) {
  const tipId = useId();
  const variant = COPY[state];
  const tooltip =
    state === 'reconnecting'
      ? `Reconnect attempt ${attempts}`
      : state === 'disconnected'
        ? 'Stream offline — showing last known data'
        : 'Streaming live metrics';

  return (
    <span
      className={cn(
        'inline-flex items-center gap-1.5 rounded-full px-2.5 py-0.5 text-xs font-medium ring-1 ring-inset',
        variant.text,
        variant.ring,
        'bg-surface',
      )}
      aria-describedby={tipId}
    >
      <span className="relative flex h-1.5 w-1.5">
        {state === 'connected' && (
          <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-emerald-400 opacity-75" />
        )}
        <span className={cn('relative inline-flex h-1.5 w-1.5 rounded-full', variant.dot)} />
      </span>
      {variant.label}
      <span id={tipId} role="tooltip" className="sr-only">
        {tooltip}
      </span>
    </span>
  );
}
