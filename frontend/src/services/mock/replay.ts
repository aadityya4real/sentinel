import type { Timeline } from '@/types/api';
import { mockEvents } from './events';

export function mockTimeline(): Timeline {
  return {
    hostname: 'prod-web-01',
    from: new Date(Date.now() - 3600_000).toISOString(),
    to: new Date().toISOString(),
    events: mockEvents(15),
    limit: 50,
  };
}
