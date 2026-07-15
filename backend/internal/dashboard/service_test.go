package dashboard

import (
	"context"
	"testing"
	"time"

	"github.com/aadityya4real/sentinel/backend/internal/agent"
)

type repositoryStub struct {
	overview      Overview
	metrics       []agent.Metrics
	activeSince   time.Time
	historyResult []agent.Metrics
}

func (s *repositoryStub) Overview(_ context.Context, activeSince time.Time) (Overview, error) {
	s.activeSince = activeSince
	return s.overview, nil
}

func (s *repositoryStub) LatestHosts(context.Context, int) ([]agent.Metrics, error) {
	return s.metrics, nil
}

func (s *repositoryStub) History(context.Context, string, time.Time, time.Time, int) ([]agent.Metrics, error) {
	return s.historyResult, nil
}

func TestServiceGetHostsAssignsFreshnessStatus(t *testing.T) {
	now := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)
	repository := &repositoryStub{metrics: []agent.Metrics{
		{Hostname: "active-node", Timestamp: now.Add(-20 * time.Second)},
		{Hostname: "stale-node", Timestamp: now.Add(-31 * time.Second)},
	}}
	service, err := NewService(repository)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}
	service.now = func() time.Time { return now }

	page, err := service.GetHosts(context.Background(), 100)
	if err != nil {
		t.Fatalf("GetHosts() error = %v", err)
	}
	if page.Hosts[0].Status != "active" || page.Hosts[1].Status != "stale" {
		t.Fatalf("statuses = %q, %q; want active, stale", page.Hosts[0].Status, page.Hosts[1].Status)
	}
}

func TestServiceGetOverviewUsesActiveHostWindow(t *testing.T) {
	now := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)
	repository := &repositoryStub{}
	service, err := NewService(repository)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}
	service.now = func() time.Time { return now }

	overview, err := service.GetOverview(context.Background())
	if err != nil {
		t.Fatalf("GetOverview() error = %v", err)
	}
	if want := now.Add(-activeHostWindow); !repository.activeSince.Equal(want) {
		t.Fatalf("active since = %s, want %s", repository.activeSince, want)
	}
	if overview.ActiveWithinSeconds != 30 {
		t.Fatalf("active within seconds = %d, want 30", overview.ActiveWithinSeconds)
	}
}
