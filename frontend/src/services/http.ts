import { API_URL, USE_MOCK_DATA } from '@/config/env';
import type { ApiError } from '@/types/api';

export class ApiClientError extends Error {
  code: string;
  status: number;

  constructor(code: string, message: string, status: number) {
    super(message);
    this.code = code;
    this.status = status;
    this.name = 'ApiClientError';
  }
}

function buildUrl(path: string, params?: Record<string, string | number | undefined>): string {
  const base = API_URL ? `${API_URL}${path}` : path;
  if (!params) return base;
  const search = new URLSearchParams();
  for (const [key, value] of Object.entries(params)) {
    if (value !== undefined && value !== '') search.set(key, String(value));
  }
  const qs = search.toString();
  return qs ? `${base}?${qs}` : base;
}

export async function apiGet<T>(path: string, params?: Record<string, string | number | undefined>): Promise<T> {
  const res = await fetch(buildUrl(path, params));
  if (!res.ok) {
    const body = (await res.json()) as ApiError;
    throw new ApiClientError(body.error.code, body.error.message, res.status);
  }
  return res.json();
}

export async function apiPost<T>(path: string, body: unknown): Promise<T> {
  const base = API_URL ? `${API_URL}${path}` : path;
  const res = await fetch(base, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  });
  if (!res.ok) {
    const errBody = (await res.json()) as ApiError;
    throw new ApiClientError(errBody.error.code, errBody.error.message, res.status);
  }
  return res.json();
}

export { USE_MOCK_DATA };
