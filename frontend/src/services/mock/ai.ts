import type { Analysis } from '@/types/api';

export function mockAnalysis(): Analysis {
  return {
    hostname: 'prod-api-01',
    from: new Date(Date.now() - 900_000).toISOString(),
    to: new Date().toISOString(),
    analyzed_event_count: 42,
    summary:
      'High memory usage detected on prod-api-01 between 14:15 and 14:30 UTC. Memory utilization peaked at 94.2%, triggering garbage collection thrashing and increased latency. The pattern correlates with a traffic surge from the marketing campaign that started at 14:10 UTC.',
    severity: 'high',
    probable_causes: [
      'Traffic surge exceeding allocated resources',
      'Memory leak in request handler introduced in deploy v2.14.3',
      'Increased cache miss rate due to cold cache after restart at 13:50 UTC',
    ],
    evidence: [
      { event_id: 8921, observation: 'Memory usage crossed 90% threshold at 14:17 UTC' },
      { event_id: 8935, observation: 'Garbage collection frequency increased 340% at 14:19 UTC' },
      { event_id: 8901, observation: 'Request latency P99 exceeded 500ms at 14:18 UTC' },
      { event_id: 8940, observation: 'Connection pool exhausted at 14:20 UTC' },
    ],
    recommended_actions: [
      'Scale horizontally by adding 2 additional API instances',
      'Deploy hotfix for memory leak in request handler (PR #1847)',
      'Increase connection pool size from 50 to 100 as immediate mitigation',
      'Enable memory profiling to pinpoint exact leak source',
    ],
    confidence: 0.87,
  };
}
