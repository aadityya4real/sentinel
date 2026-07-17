import type { Event } from '@/types/api';

function recentTimestamp(offsetSeconds = 0): string {
  return new Date(Date.now() - offsetSeconds * 1000).toISOString();
}

const eventTypes = [
  'infrastructure.metrics.collected',
  'infrastructure.host.started',
  'infrastructure.host.stopped',
  'infrastructure.disk.warning',
  'infrastructure.memory.critical',
  'infrastructure.cpu.spike',
  'infrastructure.network.error',
  'infrastructure.service.restarted',
];

export function mockEvents(count = 20): Event[] {
  return Array.from({ length: count }, (_, i) => ({
    id: 1000 - i,
    key: `event-${1000 - i}`,
    type: eventTypes[i % eventTypes.length],
    subject_type: 'host',
    subject_id: ['prod-web-01', 'prod-api-01', 'prod-db-01', 'staging-web-01', 'dev-sandbox-01'][i % 5],
    occurred_at: recentTimestamp(i * 180),
    recorded_at: recentTimestamp(i * 180),
    payload: { source: 'agent', version: '0.1.0' },
  }));
}
