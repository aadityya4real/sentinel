package agent_test

import (
	"context"
	"testing"
	"time"

	"github.com/aadityya4real/sentinel/backend/internal/agent"
	"github.com/aadityya4real/sentinel/backend/internal/collector"
)

func TestSystemCollectorProducesValidMetrics(t *testing.T) {
	systemCollector, err := agent.NewSystemCollector(10 * time.Millisecond)
	if err != nil {
		t.Fatalf("NewSystemCollector() error = %v", err)
	}

	metrics, err := systemCollector.Collect(context.Background())
	if err != nil {
		t.Fatalf("Collect() error = %v", err)
	}
	if err := collector.Validate(metrics); err != nil {
		t.Fatalf("collected metrics do not satisfy API validation: %v", err)
	}
	if metrics.Hostname == "" || metrics.OS == "" || metrics.UptimeSeconds == 0 || metrics.Timestamp.IsZero() {
		t.Fatalf("host metadata is incomplete: %+v", metrics)
	}
	if len(metrics.Disks) == 0 {
		t.Fatal("Collect() returned no disks")
	}
}
