import React from 'react';
import { setRefreshInterval } from '@/services/api/refresh';

const STORAGE_KEY = 'sentinel-refresh-interval';

/** Available auto-refresh intervals in milliseconds */
export const REFRESH_INTERVALS = [1000, 2000, 5000, 10000] as const;
export type RefreshInterval = (typeof REFRESH_INTERVALS)[number];

export function formatRefreshLabel(ms: number): string {
  if (ms < 1000) return `${ms}ms`;
  return `${ms / 1000}s`;
}

/** Default to the slowest option (10s) when nothing is stored yet. */
const DEFAULT_MS = 10000;

function loadStoredMs(): number {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (!raw) return DEFAULT_MS;
    const n = Number(raw);
    if (Number.isFinite(n) && n >= 1000 && n <= 10000) return n;
    return DEFAULT_MS;
  } catch {
    return DEFAULT_MS;
  }
}

let currentMs = loadStoredMs();

interface UseRefreshIntervalReturn {
  /** Current interval in ms */
  ms: number;
  /** Set a new interval */
  setMs: (n: number) => void;
  /** Human-readable label, e.g. "1s", "5s" */
  label: string;
}

export function useRefreshInterval(): UseRefreshIntervalReturn {
  const [_state, _setState] = React.useState({ tick: 0 });

  function setMs(n: number) {
    if (!REFRESH_INTERVALS.includes(n as RefreshInterval)) return;
    currentMs = n;
    setRefreshInterval(n);
    try { localStorage.setItem(STORAGE_KEY, String(n)); } catch {}
    _setState((s) => ({ ...s, tick: s.tick + 1 }));
  }

  return { ms: currentMs, setMs, label: formatRefreshLabel(currentMs) };
}
