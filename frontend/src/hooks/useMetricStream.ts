import { useEffect, useRef, useState } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { MetricStream } from '@/services/stream/websocket';
import type { StreamState } from '@/services/stream/types';
import type { Metrics, HostsPage } from '@/types/api';

const BUFFER_SIZE = 60;
let sharedStream: MetricStream | null = null;
let refCount = 0;

function getStream(): MetricStream {
  if (!sharedStream) {
    sharedStream = new MetricStream();
  }
  return sharedStream;
}

function releaseStream(): void {
  refCount = Math.max(0, refCount - 1);
  if (refCount === 0 && sharedStream) {
    sharedStream.close();
    sharedStream = null;
  }
}

export interface MetricStreamResult {
  state: StreamState;
  attempts: number;
  buffer: Metrics[];
  latest: Metrics | null;
}

/**
 * Subscribes the calling component to the live metric stream. Maintains a
 * singleton stream shared across all hook consumers via refcount, keeps a
 * rolling buffer of the most recent samples, and patches the dashboard
 * React Query cache so the host table updates in real time.
 */
export function useMetricStream(): MetricStreamResult {
  const queryClient = useQueryClient();
  const [state, setState] = useState<StreamState>('disconnected');
  const [attempts, setAttempts] = useState(0);
  const [buffer, setBuffer] = useState<Metrics[]>([]);
  const bufferRef = useRef<Metrics[]>([]);

  useEffect(() => {
    const stream = getStream();
    refCount++;
    stream.connect();

    const offState = stream.onStateChange((next, att) => {
      setState(next);
      setAttempts(att);
    });

    const offMetric = stream.subscribe((metrics) => {
      const next = [...bufferRef.current, metrics].slice(-BUFFER_SIZE);
      bufferRef.current = next;
      setBuffer(next);

      const previous = queryClient.getQueryData<HostsPage>(['dashboard', 'hosts', 100]);
      if (previous) {
        const hosts = previous.hosts.map((h) =>
          h.metrics.hostname === metrics.hostname
            ? { ...h, metrics, status: 'active' as const }
            : h,
        );
        if (!hosts.some((h) => h.metrics.hostname === metrics.hostname)) {
          hosts.unshift({ metrics, status: 'active' });
        }
        queryClient.setQueryData(['dashboard', 'hosts', 100], { ...previous, hosts });
      }
    });

    return () => {
      offState();
      offMetric();
      releaseStream();
    };
  }, [queryClient]);

  return { state, attempts, buffer, latest: buffer.length > 0 ? buffer[buffer.length - 1] : null };
}
