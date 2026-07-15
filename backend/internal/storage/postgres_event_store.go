package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/aadityya4real/sentinel/backend/internal/eventstore"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgreSQLEventStore stores and retrieves immutable infrastructure events in PostgreSQL.
type PostgreSQLEventStore struct {
	pool *pgxpool.Pool
}

// NewPostgreSQLEventStore creates an event store backed by the supplied PostgreSQL pool.
func NewPostgreSQLEventStore(pool *pgxpool.Pool) (*PostgreSQLEventStore, error) {
	if pool == nil {
		return nil, fmt.Errorf("PostgreSQL pool is required")
	}

	return &PostgreSQLEventStore{pool: pool}, nil
}

// Append durably stores an event once and returns the stored event for new or retried writes.
func (s *PostgreSQLEventStore) Append(ctx context.Context, event eventstore.NewEvent) (eventstore.Event, error) {
	if err := event.Validate(); err != nil {
		return eventstore.Event{}, fmt.Errorf("validate event: %w", err)
	}

	stored, err := scanEvent(s.pool.QueryRow(ctx, `
		INSERT INTO infrastructure_events (
			event_key, event_type, subject_type, subject_id, occurred_at, payload
		) VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (event_key) DO NOTHING
		RETURNING id, event_key, event_type, subject_type, subject_id, occurred_at, recorded_at, payload`,
		event.Key,
		event.Type,
		event.SubjectType,
		event.SubjectID,
		event.OccurredAt,
		event.Payload,
	))
	if err == nil {
		return stored, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return eventstore.Event{}, fmt.Errorf("append infrastructure event: %w", err)
	}

	stored, err = scanEvent(s.pool.QueryRow(ctx, `
		SELECT id, event_key, event_type, subject_type, subject_id, occurred_at, recorded_at, payload
		FROM infrastructure_events
		WHERE event_key = $1`, event.Key))
	if err != nil {
		return eventstore.Event{}, fmt.Errorf("read existing infrastructure event: %w", err)
	}

	return stored, nil
}

// List returns immutable events in chronological order according to the supplied filter.
func (s *PostgreSQLEventStore) List(ctx context.Context, filter eventstore.Filter) ([]eventstore.Event, error) {
	if filter.Limit == 0 {
		filter.Limit = 100
	}
	if filter.Limit < 1 || filter.Limit > 1000 {
		return nil, fmt.Errorf("event list limit must be between 1 and 1000")
	}
	if !filter.From.IsZero() && !filter.To.IsZero() && filter.From.After(filter.To) {
		return nil, fmt.Errorf("event list from must be before or equal to to")
	}
	if !filter.AfterAt.IsZero() && filter.AfterID < 1 {
		return nil, fmt.Errorf("event list cursor ID must be positive")
	}

	query, arguments := buildEventListQuery(filter)
	rows, err := s.pool.Query(ctx, query, arguments...)
	if err != nil {
		return nil, fmt.Errorf("query infrastructure events: %w", err)
	}
	defer rows.Close()

	events := make([]eventstore.Event, 0)
	for rows.Next() {
		event, err := scanEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate infrastructure events: %w", err)
	}

	return events, nil
}

// Latest returns the most recent event matching the supplied filter.
func (s *PostgreSQLEventStore) Latest(ctx context.Context, filter eventstore.Filter) (eventstore.Event, bool, error) {
	if !filter.From.IsZero() || !filter.AfterAt.IsZero() || filter.AfterID != 0 {
		return eventstore.Event{}, false, fmt.Errorf("latest event query does not support lower-bound filters")
	}
	query, arguments := buildLatestEventQuery(filter)
	event, err := scanEvent(s.pool.QueryRow(ctx, query, arguments...))
	if errors.Is(err, pgx.ErrNoRows) {
		return eventstore.Event{}, false, nil
	}
	if err != nil {
		return eventstore.Event{}, false, fmt.Errorf("query latest infrastructure event: %w", err)
	}

	return event, true, nil
}

func buildEventListQuery(filter eventstore.Filter) (string, []any) {
	conditions, arguments := eventConditions(filter)
	query := `SELECT id, event_key, event_type, subject_type, subject_id, occurred_at, recorded_at, payload
		FROM infrastructure_events`
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	arguments = append(arguments, filter.Limit)
	query += fmt.Sprintf(" ORDER BY occurred_at ASC, id ASC LIMIT $%d", len(arguments))
	return query, arguments
}

func buildLatestEventQuery(filter eventstore.Filter) (string, []any) {
	conditions, arguments := eventConditions(filter)
	query := `SELECT id, event_key, event_type, subject_type, subject_id, occurred_at, recorded_at, payload
		FROM infrastructure_events`
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	return query + " ORDER BY occurred_at DESC, id DESC LIMIT 1", arguments
}

func eventConditions(filter eventstore.Filter) ([]string, []any) {
	conditions := make([]string, 0, 5)
	arguments := make([]any, 0, 6)
	addCondition := func(column string, value any) {
		arguments = append(arguments, value)
		conditions = append(conditions, fmt.Sprintf("%s = $%d", column, len(arguments)))
	}
	if filter.Type != "" {
		addCondition("event_type", filter.Type)
	}
	if filter.SubjectType != "" {
		addCondition("subject_type", filter.SubjectType)
	}
	if filter.SubjectID != "" {
		addCondition("subject_id", filter.SubjectID)
	}
	if !filter.From.IsZero() {
		arguments = append(arguments, filter.From)
		conditions = append(conditions, fmt.Sprintf("occurred_at >= $%d", len(arguments)))
	}
	if !filter.To.IsZero() {
		arguments = append(arguments, filter.To)
		conditions = append(conditions, fmt.Sprintf("occurred_at <= $%d", len(arguments)))
	}
	if !filter.AfterAt.IsZero() {
		arguments = append(arguments, filter.AfterAt, filter.AfterID)
		conditions = append(conditions, fmt.Sprintf("(occurred_at, id) > ($%d, $%d)", len(arguments)-1, len(arguments)))
	}

	return conditions, arguments
}

func scanEvent(row pgx.Row) (eventstore.Event, error) {
	var event eventstore.Event
	var payload []byte
	if err := row.Scan(
		&event.ID,
		&event.Key,
		&event.Type,
		&event.SubjectType,
		&event.SubjectID,
		&event.OccurredAt,
		&event.RecordedAt,
		&payload,
	); err != nil {
		return eventstore.Event{}, err
	}
	if !json.Valid(payload) {
		return eventstore.Event{}, fmt.Errorf("stored event payload is not valid JSON")
	}
	event.Payload = append(event.Payload[:0], payload...)

	return event, nil
}
