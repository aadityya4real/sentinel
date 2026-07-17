import { useQuery } from '@tanstack/react-query';
import { apiGet, USE_MOCK_DATA } from '@/services/http';
import { withMockFallback } from '@/services/mock';
import { mockTimeline } from '@/services/mock/replay';
import type { Timeline } from '@/types/api';

async function fetchReplay(hostname: string, cursor?: string, limit = 50): Promise<Timeline> {
  return apiGet<Timeline>(`/api/v1/replay/hosts/${hostname}`, { limit, cursor });
}

function isEmptyTimeline(data: Timeline): boolean {
  return data.events.length === 0;
}

export function useReplay(hostname: string, cursor?: string, limit = 50) {
  return useQuery({
    queryKey: ['replay', hostname, cursor, limit],
    queryFn: () =>
      withMockFallback(['replay', hostname], () => fetchReplay(hostname, cursor, limit), mockTimeline, {
        enabled: USE_MOCK_DATA,
        isEmpty: isEmptyTimeline,
      }),
    enabled: !!hostname,
  });
}
