import { REFRESH_INTERVAL_MS as DEFAULT_MS } from '@/config/env';

let _interval = DEFAULT_MS;

/** Get the current auto-refresh interval in milliseconds */
export function getRefreshInterval(): number { return _interval; }

/** Set a new refresh interval (called by Settings UI) */
export function setRefreshInterval(ms: number): void { _interval = ms; }
