import { useQuery } from '@tanstack/react-query';
import { apiGet, USE_MOCK_DATA } from '@/services/http';
import { withMockFallback } from '@/services/mock';
import { mockHealth } from '@/services/mock/health';
import type { HealthResponse } from '@/types/api';
import { getRefreshInterval } from './refresh';

async function fetchHealth(): Promise<HealthResponse> {
  return apiGet<HealthResponse>('/api/v1/health');
}

export function useHealth() {
  return useQuery({
    queryKey: ['health'],
    queryFn: () =>
      withMockFallback(['health'], fetchHealth, mockHealth, {
        enabled: USE_MOCK_DATA,
        isEmpty: () => false,
      }),
    refetchInterval: () => getRefreshInterval(),
  });
}
