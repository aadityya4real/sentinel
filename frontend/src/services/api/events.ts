import { useQuery } from '@tanstack/react-query';
import { apiGet, USE_MOCK_DATA } from '@/services/http';
import { withMockFallback } from '@/services/mock';
import { mockEvents } from '@/services/mock/events';
import type { Event } from '@/types/api';
import { getRefreshInterval } from './refresh';

const DEFAULT_LIMIT = 50;

type EventPage = { events: Event[]; next_cursor?: string };

async function fetchEvents(limit = DEFAULT_LIMIT, cursor?: string): Promise<EventPage> {
  const params = new URLSearchParams({ limit: String(limit) });
  if (cursor) params.set('cursor', cursor);
  return apiGet<EventPage>(`/api/v1/events?${params.toString()}`);
}

export function useEvents(limit = DEFAULT_LIMIT, cursor?: string) {
  return useQuery({
    queryKey: ['events', limit, cursor],
    queryFn: () =>
      withMockFallback(['events'], () => fetchEvents(limit, cursor),
        (): EventPage => ({ events: cursor ? [] : mockEvents(limit) }),
        { enabled: USE_MOCK_DATA, isEmpty: (d) => d.events.length === 0 },
      ),
    refetchInterval: () => getRefreshInterval(),
  });
}

export function useHostEvents(hostname: string, limit = 50) {
  return useQuery({
    queryKey: ['events', 'host', hostname, limit],
    queryFn: async (): Promise<Event[]> => {
      const params = new URLSearchParams({
        subject_id: hostname,
        subject_type: 'host',
        limit: String(limit),
      });
      const res = await apiGet<{ events: Event[] }>(`/api/v1/events?${params.toString()}`);
      return res.events;
    },
    enabled: !!hostname,
  });
}
