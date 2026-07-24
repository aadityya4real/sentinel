export const API_URL = import.meta.env.VITE_API_URL || '';
export const USE_MOCK_DATA = import.meta.env.VITE_USE_MOCK_DATA === 'true';
export const REFRESH_INTERVAL_MS = Number(import.meta.env.VITE_REFRESH_INTERVAL_MS) || 15000;
export const APP_VERSION = '0.1.0';

function resolveWsUrl(): string {
  const explicit = import.meta.env.VITE_WS_URL;
  if (explicit) return explicit;
  if (API_URL) {
    const url = new URL(API_URL);
    return `${url.protocol === 'https:' ? 'wss' : 'ws'}://${url.host}/ws/v1/metrics`;
  }
  return '/ws/v1/metrics';
}

export const WS_URL = resolveWsUrl();

