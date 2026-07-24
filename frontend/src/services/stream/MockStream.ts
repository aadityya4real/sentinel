import type { MetricStreamLike, MetricSubscriber, StateSubscriber, StreamState } from './types';
import type { Metrics } from '@/types/api';
import { generateMetrics, mockHostsConfig } from '@/services/mock/dashboard';

const INTERVAL_MS = 2000;

/**
 * Dev-only simulated metric stream. Used as a transparent fallback when
 * VITE_USE_MOCK_DATA is enabled and the WebSocket backend is unreachable.
 * Mirrors the MetricStream contract so the UI cannot tell them apart.
 */
export class MockStream implements MetricStreamLike {
  private metricHandlers = new Set<MetricSubscriber>();
  private stateHandlers = new Set<StateSubscriber>();
  private timer: ReturnType<typeof setInterval> | null = null;
  private state: StreamState = 'disconnected';
  private index = 0;

  connect(): void {
    if (this.timer) return;
    this.setState('connected', 0);
    this.timer = setInterval(() => this.emit(), INTERVAL_MS);
  }

  close(): void {
    if (this.timer) {
      clearInterval(this.timer);
      this.timer = null;
    }
    this.setState('disconnected', 0);
  }

  subscribe(handler: MetricSubscriber): () => void {
    this.metricHandlers.add(handler);
    return () => this.metricHandlers.delete(handler);
  }

  onStateChange(handler: StateSubscriber): () => void {
    this.stateHandlers.add(handler);
    handler(this.state, 0);
    return () => this.stateHandlers.delete(handler);
  }

  private emit(): void {
    const host = mockHostsConfig[this.index % mockHostsConfig.length];
    this.index++;
    const jitter = () => (Math.random() - 0.5) * 8;
    const metrics: Metrics = generateMetrics(
      host.name,
      Math.max(0, Math.min(100, host.cpu + jitter())),
      Math.max(0, Math.min(100, host.mem + jitter())),
    );
    this.metricHandlers.forEach((h) => h(metrics));
  }

  private setState(state: StreamState, attempts: number): void {
    this.state = state;
    this.stateHandlers.forEach((h) => h(state, attempts));
  }
}
