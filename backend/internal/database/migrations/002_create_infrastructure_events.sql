CREATE TABLE IF NOT EXISTS infrastructure_events (
    id BIGSERIAL PRIMARY KEY,
    event_key TEXT NOT NULL UNIQUE,
    event_type TEXT NOT NULL,
    subject_type TEXT NOT NULL,
    subject_id TEXT NOT NULL,
    occurred_at TIMESTAMPTZ NOT NULL,
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    payload JSONB NOT NULL
);

CREATE INDEX IF NOT EXISTS infrastructure_events_subject_occurred_at_idx
    ON infrastructure_events (subject_type, subject_id, occurred_at ASC, id ASC);

CREATE INDEX IF NOT EXISTS infrastructure_events_type_occurred_at_idx
    ON infrastructure_events (event_type, occurred_at ASC, id ASC);
