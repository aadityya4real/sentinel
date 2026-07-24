import type { Metrics } from '@/types/api';

export type StreamState = 'connecting' | 'connected' | 'reconnecting' | 'disconnected';

export interface StreamMessage {
  type: 'metrics';
  payload: Metrics;
}

export interface MetricSubscriber {
  (metrics: Metrics): void;
}

export interface StateSubscriber {
  (state: StreamState, attempts: number): void;
}

export interface MetricStreamLike {
  connect(): void;
  close(): void;
  subscribe(handler: MetricSubscriber): () => void;
  onStateChange(handler: StateSubscriber): () => void;
}
