package agent

import "context"

// Collector produces a point-in-time host metrics snapshot.
type Collector interface {
	Collect(ctx context.Context) (Metrics, error)
}
