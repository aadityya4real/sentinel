export interface HealthResponse {
  status: 'healthy' | 'unhealthy';
  database: 'connected' | 'disconnected';
  redis: 'connected' | 'disconnected';
  uptime: string;
  version: string;
}

export interface DiskUsage {
  path: string;
  filesystem: string;
  total_bytes: number;
  used_bytes: number;
  used_percent: number;
}

export interface MemoryUsage {
  total_bytes: number;
  used_bytes: number;
  used_percent: number;
  available_bytes: number;
}

export interface Metrics {
  cpu_usage_percent: number;
  memory: MemoryUsage;
  disks: DiskUsage[];
  hostname: string;
  os: string;
  uptime_seconds: number;
  timestamp: string;
}

export interface Overview {
  total_hosts: number;
  active_hosts: number;
  active_within_seconds: number;
  average_cpu_usage_percent: number;
  average_memory_usage_percent: number;
  latest_metric_at?: string;
}

export interface HostSnapshot {
  metrics: Metrics;
  status: 'active' | 'stale';
}

export interface HostsPage {
  hosts: HostSnapshot[];
  limit: number;
}

export interface History {
  hostname: string;
  from: string;
  to: string;
  metrics: Metrics[];
  limit: number;
}

export interface Event {
  id: number;
  key: string;
  type: string;
  subject_type: string;
  subject_id: string;
  occurred_at: string;
  recorded_at: string;
  payload: unknown;
}

export interface Timeline {
  hostname: string;
  from: string;
  to: string;
  events: Event[];
  limit: number;
  next_cursor?: string;
}

export interface Snapshot {
  hostname: string;
  requested_at: string;
  observed_at: string;
  event_id: number;
  metrics: Metrics;
}

export interface Evidence {
  event_id: number;
  observation: string;
}

export interface Analysis {
  hostname: string;
  from: string;
  to: string;
  analyzed_event_count: number;
  summary: string;
  severity: string;
  probable_causes: string[];
  evidence: Evidence[];
  recommended_actions: string[];
  confidence: number;
}

export interface IncidentRequest {
  hostname: string;
  from: string;
  to: string;
  event_limit: number;
}

export interface ApiError {
  error: {
    code: string;
    message: string;
  };
}
