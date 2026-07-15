CREATE TABLE IF NOT EXISTS servers (
    id BIGSERIAL PRIMARY KEY,
    hostname TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT servers_hostname_not_blank CHECK (LENGTH(BTRIM(hostname)) > 0)
);

CREATE TABLE IF NOT EXISTS events (
    id BIGSERIAL PRIMARY KEY,
    server_id BIGINT NOT NULL REFERENCES servers(id) ON DELETE RESTRICT,
    event_type TEXT NOT NULL,
    severity TEXT NOT NULL,
    payload JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT events_event_type_not_blank CHECK (LENGTH(BTRIM(event_type)) > 0),
    CONSTRAINT events_severity_not_blank CHECK (LENGTH(BTRIM(severity)) > 0)
);

CREATE INDEX IF NOT EXISTS events_server_id_created_at_idx
    ON events (server_id, created_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS events_event_type_created_at_idx
    ON events (event_type, created_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS events_payload_gin_idx
    ON events USING GIN (payload);
