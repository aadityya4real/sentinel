// Package dashboard provides read models for the Sentinel infrastructure dashboard.
package dashboard

import (
	"context"
	"errors"
	"time"

	"github.com/aadityya4real/sentinel/backend/internal/agent"
)

const activeHostWindow = 30 * time.Second

// Overview contains fleet-level metrics for dashboard summary cards.
type Overview struct {
	TotalHosts                int64      `json:"total_hosts"`
	ActiveHosts               int64      `json:"active_hosts"`
	ActiveWithinSeconds       int64      `json:"active_within_seconds"`
	AverageCPUUsagePercent    float64    `json:"average_cpu_usage_percent"`
	AverageMemoryUsagePercent float64    `json:"average_memory_usage_percent"`
	LatestMetricAt            *time.Time `json:"latest_metric_at,omitempty"`
}

// HostSnapshot contains the most recently collected metrics and status for one host.
type HostSnapshot struct {
	Metrics agent.Metrics `json:"metrics"`
	Status  string        `json:"status"`
}

// HostsPage contains the current snapshots requested for the fleet dashboard.
type HostsPage struct {
	Hosts []HostSnapshot `json:"hosts"`
	Limit int            `json:"limit"`
}

// History contains a bounded time series for one host.
type History struct {
	Hostname string          `json:"hostname"`
	From     time.Time       `json:"from"`
	To       time.Time       `json:"to"`
	Metrics  []agent.Metrics `json:"metrics"`
	Limit    int             `json:"limit"`
}

// Repository retrieves dashboard read models from persistent storage.
type Repository interface {
	Overview(ctx context.Context, activeSince time.Time) (Overview, error)
	LatestHosts(ctx context.Context, limit int) ([]agent.Metrics, error)
	History(ctx context.Context, hostname string, from, to time.Time, limit int) ([]agent.Metrics, error)
}

// Reader serves dashboard queries for HTTP and other delivery mechanisms.
type Reader interface {
	GetOverview(ctx context.Context) (Overview, error)
	GetHosts(ctx context.Context, limit int) (HostsPage, error)
	GetHistory(ctx context.Context, hostname string, from, to time.Time, limit int) (History, error)
}

// Service coordinates dashboard read queries and computes host freshness status.
type Service struct {
	repository Repository
	now        func() time.Time
}

// NewService creates a dashboard query service backed by the supplied repository.
func NewService(repository Repository) (*Service, error) {
	if repository == nil {
		return nil, errors.New("dashboard repository is required")
	}

	return &Service{repository: repository, now: time.Now}, nil
}

// GetOverview returns aggregate metrics for the current host fleet.
func (s *Service) GetOverview(ctx context.Context) (Overview, error) {
	now := s.now().UTC()
	overview, err := s.repository.Overview(ctx, now.Add(-activeHostWindow))
	if err != nil {
		return Overview{}, err
	}
	overview.ActiveWithinSeconds = int64(activeHostWindow.Seconds())
	return overview, nil
}

// GetHosts returns latest metrics and freshness status for up to limit hosts.
func (s *Service) GetHosts(ctx context.Context, limit int) (HostsPage, error) {
	metrics, err := s.repository.LatestHosts(ctx, limit)
	if err != nil {
		return HostsPage{}, err
	}

	now := s.now().UTC()
	hosts := make([]HostSnapshot, 0, len(metrics))
	for _, metric := range metrics {
		status := "stale"
		if !metric.Timestamp.Before(now.Add(-activeHostWindow)) {
			status = "active"
		}
		hosts = append(hosts, HostSnapshot{Metrics: metric, Status: status})
	}

	return HostsPage{Hosts: hosts, Limit: limit}, nil
}

// GetHistory returns a bounded metric time series for the requested host and time range.
func (s *Service) GetHistory(ctx context.Context, hostname string, from, to time.Time, limit int) (History, error) {
	metrics, err := s.repository.History(ctx, hostname, from, to, limit)
	if err != nil {
		return History{}, err
	}

	return History{
		Hostname: hostname,
		From:     from,
		To:       to,
		Metrics:  metrics,
		Limit:    limit,
	}, nil
}
