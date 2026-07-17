import type { HealthResponse } from '@/types/api';

export function mockHealth(): HealthResponse {
  return {
    status: 'healthy',
    database: 'connected',
    redis: 'connected',
    uptime: '1h23m45s',
    version: '0.1.0',
  };
}
