import { WS_URL, USE_MOCK_DATA } from '@/config/env';
import { MockStream } from './MockStream';
import type { MetricStreamLike, MetricSubscriber, StateSubscriber, StreamMessage, StreamState } from './types';

const RECONNECT_DELAYS_MS = [1000, 2000, 4000, 8000, 15000];
const HEARTBEAT_INTERVAL_MS = 30_000;
const HEARTBEAT_TIMEOUT_MS = 45_000;

/**
 * MetricStream manages a WebSocket subscription to the Sentinel live metrics
 * endpoint with automatic reconnection, heartbeat, and a mock fallback for
 * development when the backend is unreachable.
 */
export class MetricStream implements MetricStreamLike {
  private socket: WebSocket | null = null;
  private metricHandlers = new Set<MetricSubscriber>();
  private stateHandlers = new Set<StateSubscriber>();
  private state: StreamState = 'disconnected';
  private attempts = 0;
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  private heartbeatTimer: ReturnType<typeof setInterval> | null = null;
  private heartbeatTimeoutTimer: ReturnType<typeof setTimeout> | null = null;
  private closedIntentionally = false;
  private mockFallback: MockStream | null = null;

  connect(): void {
    if (this.state === 'connecting' || this.state === 'connected') return;
    this.closedIntentionally = false;
    this.open();
  }

  close(): void {
    this.closedIntentionally = true;
    this.clearTimers();
    if (this.mockFallback) {
      this.mockFallback.close();
      this.mockFallback = null;
    }
    if (this.socket) {
      this.socket.onclose = null;
      this.socket.onerror = null;
      this.socket.onmessage = null;
      this.socket.onopen = null;
      this.socket.close(1000, 'client shutdown');
      this.socket = null;
    }
    this.setState('disconnected', 0);
  }

  subscribe(handler: MetricSubscriber): () => void {
    this.metricHandlers.add(handler);
    return () => this.metricHandlers.delete(handler);
  }

  onStateChange(handler: StateSubscriber): () => void {
    this.stateHandlers.add(handler);
    handler(this.state, this.attempts);
    return () => this.stateHandlers.delete(handler);
  }

  private open(): void {
    this.setState(this.attempts === 0 ? 'connecting' : 'reconnecting', this.attempts);
    let socket: WebSocket;
    try {
      socket = new WebSocket(WS_URL);
    } catch {
      this.handleFailure();
      return;
    }
    this.socket = socket;

    socket.onopen = () => {
      this.attempts = 0;
      this.setState('connected', 0);
      this.startHeartbeat();
    };

    socket.onmessage = (event) => {
      this.resetHeartbeatTimeout();
      let message: StreamMessage;
      try {
        message = JSON.parse(event.data);
      } catch {
        return;
      }
      if (message?.type === 'metrics' && message.payload) {
        this.metricHandlers.forEach((h) => h(message.payload));
      }
    };

    socket.onerror = () => {
      // The close event will follow; handle reconnect there.
    };

    socket.onclose = () => {
      this.stopHeartbeat();
      this.socket = null;
      if (this.closedIntentionally) return;
      this.scheduleReconnect();
    };
  }

  private scheduleReconnect(): void {
    if (this.attempts >= RECONNECT_DELAYS_MS.length) {
      if (USE_MOCK_DATA) {
        this.switchToMock();
        return;
      }
      this.setState('disconnected', this.attempts);
      return;
    }
    const delay = RECONNECT_DELAYS_MS[this.attempts];
    this.attempts++;
    this.setState('reconnecting', this.attempts);
    this.reconnectTimer = setTimeout(() => this.open(), delay);
  }

  private handleFailure(): void {
    if (USE_MOCK_DATA) {
      this.switchToMock();
    } else {
      this.setState('disconnected', this.attempts);
    }
  }

  private switchToMock(): void {
    this.mockFallback = new MockStream();
    this.mockFallback.subscribe((m) => this.metricHandlers.forEach((h) => h(m)));
    this.mockFallback.onStateChange((s, a) => this.setState(s, a));
    this.mockFallback.connect();
  }

  private startHeartbeat(): void {
    this.stopHeartbeat();
    this.heartbeatTimer = setInterval(() => {
      if (this.socket?.readyState === WebSocket.OPEN) {
        try {
          this.socket.send(JSON.stringify({ type: 'ping' }));
        } catch {
          // socket may have closed between ticks
        }
      }
      this.resetHeartbeatTimeout();
    }, HEARTBEAT_INTERVAL_MS);
    this.resetHeartbeatTimeout();
  }

  private resetHeartbeatTimeout(): void {
    if (this.heartbeatTimeoutTimer) clearTimeout(this.heartbeatTimeoutTimer);
    this.heartbeatTimeoutTimer = setTimeout(() => {
      // No message received in time — force reconnect.
      if (this.socket) {
        this.socket.close();
      }
    }, HEARTBEAT_TIMEOUT_MS);
  }

  private stopHeartbeat(): void {
    if (this.heartbeatTimer) {
      clearInterval(this.heartbeatTimer);
      this.heartbeatTimer = null;
    }
    if (this.heartbeatTimeoutTimer) {
      clearTimeout(this.heartbeatTimeoutTimer);
      this.heartbeatTimeoutTimer = null;
    }
  }

  private clearTimers(): void {
    this.stopHeartbeat();
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
  }

  private setState(state: StreamState, attempts: number): void {
    this.state = state;
    this.attempts = attempts;
    this.stateHandlers.forEach((h) => h(state, attempts));
  }
}
