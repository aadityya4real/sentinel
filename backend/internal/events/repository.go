package events

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/aadityya4real/sentinel/backend/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const maxEventQueryLimit = 500

// ErrEventNotFound indicates that no event matched a repository query.
var ErrEventNotFound = errors.New("event not found")

// ErrServerNotFound indicates that no server matched an event creation request.
var ErrServerNotFound = errors.New("server not found")

// Repository persists and queries relational server events in PostgreSQL.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates an event repository backed by the supplied PostgreSQL pool.
func NewRepository(pool *pgxpool.Pool) (*Repository, error) {
	if pool == nil {
		return nil, errors.New("PostgreSQL pool is required")
	}

	return &Repository{pool: pool}, nil
}

// CreateEvent validates and stores an event atomically for an existing server.
func (r *Repository) CreateEvent(ctx context.Context, event models.ServerEvent) (models.ServerEvent, error) {
	if err := event.ValidateCreate(); err != nil {
		return models.ServerEvent{}, fmt.Errorf("validate event: %w", err)
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return models.ServerEvent{}, fmt.Errorf("begin event transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var serverExists bool
	if err := tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM servers WHERE id = $1)", event.ServerID).Scan(&serverExists); err != nil {
		return models.ServerEvent{}, fmt.Errorf("verify server: %w", err)
	}
	if !serverExists {
		return models.ServerEvent{}, fmt.Errorf("server %d: %w", event.ServerID, ErrServerNotFound)
	}

	created, err := scanServerEvent(tx.QueryRow(ctx, `
		INSERT INTO events (server_id, event_type, severity, payload)
		VALUES ($1, $2, $3, $4)
		RETURNING id, server_id, event_type, severity, payload, created_at`,
		event.ServerID, event.EventType, event.Severity, event.Payload,
	))
	if err != nil {
		return models.ServerEvent{}, fmt.Errorf("insert event: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return models.ServerEvent{}, fmt.Errorf("commit event transaction: %w", err)
	}

	return created, nil
}

// GetEvents returns the newest events up to limit across all servers.
func (r *Repository) GetEvents(ctx context.Context, limit int) ([]models.ServerEvent, error) {
	limit, err := normalizeLimit(limit)
	if err != nil {
		return nil, err
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, server_id, event_type, severity, payload, created_at
		FROM events ORDER BY created_at DESC, id DESC LIMIT $1`, limit)
	if err != nil {
		return nil, fmt.Errorf("query events: %w", err)
	}
	defer rows.Close()

	return collectServerEvents(rows)
}

// GetEventsByServer returns the newest events for one server up to limit.
func (r *Repository) GetEventsByServer(ctx context.Context, serverID int64, limit int) ([]models.ServerEvent, error) {
	if serverID < 1 {
		return nil, fmt.Errorf("server_id must be positive")
	}
	limit, err := normalizeLimit(limit)
	if err != nil {
		return nil, err
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, server_id, event_type, severity, payload, created_at
		FROM events WHERE server_id = $1 ORDER BY created_at DESC, id DESC LIMIT $2`, serverID, limit)
	if err != nil {
		return nil, fmt.Errorf("query server events: %w", err)
	}
	defer rows.Close()

	return collectServerEvents(rows)
}

// LatestEvent returns the newest event for one server.
func (r *Repository) LatestEvent(ctx context.Context, serverID int64) (models.ServerEvent, error) {
	if serverID < 1 {
		return models.ServerEvent{}, fmt.Errorf("server_id must be positive")
	}
	event, err := scanServerEvent(r.pool.QueryRow(ctx, `
		SELECT id, server_id, event_type, severity, payload, created_at
		FROM events WHERE server_id = $1 ORDER BY created_at DESC, id DESC LIMIT 1`, serverID))
	if errors.Is(err, pgx.ErrNoRows) {
		return models.ServerEvent{}, ErrEventNotFound
	}
	if err != nil {
		return models.ServerEvent{}, fmt.Errorf("query latest server event: %w", err)
	}

	return event, nil
}

func normalizeLimit(limit int) (int, error) {
	if limit == 0 {
		return 100, nil
	}
	if limit < 1 || limit > maxEventQueryLimit {
		return 0, fmt.Errorf("limit must be between 1 and %d", maxEventQueryLimit)
	}
	return limit, nil
}

func collectServerEvents(rows pgx.Rows) ([]models.ServerEvent, error) {
	events := make([]models.ServerEvent, 0)
	for rows.Next() {
		event, err := scanServerEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate events: %w", err)
	}

	return events, nil
}

func scanServerEvent(row pgx.Row) (models.ServerEvent, error) {
	var event models.ServerEvent
	var payload []byte
	if err := row.Scan(&event.ID, &event.ServerID, &event.EventType, &event.Severity, &payload, &event.CreatedAt); err != nil {
		return models.ServerEvent{}, err
	}
	if !json.Valid(payload) {
		return models.ServerEvent{}, fmt.Errorf("stored event payload is not valid JSON")
	}
	event.Payload = append(event.Payload[:0], payload...)

	return event, nil
}
