CREATE TABLE IF NOT EXISTS infrastructure_metrics (
    id BIGSERIAL PRIMARY KEY,
    hostname TEXT NOT NULL,
    operating_system TEXT NOT NULL,
    uptime_seconds BIGINT NOT NULL CHECK (uptime_seconds >= 0),
    collected_at TIMESTAMPTZ NOT NULL,
    received_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    cpu_usage_percent DOUBLE PRECISION NOT NULL CHECK (cpu_usage_percent >= 0 AND cpu_usage_percent <= 100),
    memory_total_bytes BIGINT NOT NULL CHECK (memory_total_bytes >= 0),
    memory_used_bytes BIGINT NOT NULL CHECK (memory_used_bytes >= 0),
    memory_available_bytes BIGINT NOT NULL CHECK (memory_available_bytes >= 0),
    memory_used_percent DOUBLE PRECISION NOT NULL CHECK (memory_used_percent >= 0 AND memory_used_percent <= 100),
    disks JSONB NOT NULL,
    CONSTRAINT infrastructure_metrics_hostname_collected_at_key UNIQUE (hostname, collected_at)
);

CREATE INDEX IF NOT EXISTS infrastructure_metrics_hostname_collected_at_idx
    ON infrastructure_metrics (hostname, collected_at DESC);
