// Command agent periodically emits Sentinel host metrics as JSON to standard output.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aadityya4real/sentinel/backend/internal/agent"
)

const (
	collectionInterval = 5 * time.Second
	cpuSampleInterval  = 250 * time.Millisecond
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	collector, err := agent.NewSystemCollector(cpuSampleInterval)
	if err != nil {
		fmt.Fprintf(os.Stderr, "create collector: %v\n", err)
		os.Exit(1)
	}

	encoder := json.NewEncoder(os.Stdout)
	if err := collectAndWrite(ctx, collector, encoder); err != nil {
		fmt.Fprintf(os.Stderr, "collect metrics: %v\n", err)
		os.Exit(1)
	}

	ticker := time.NewTicker(collectionInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := collectAndWrite(ctx, collector, encoder); err != nil {
				fmt.Fprintf(os.Stderr, "collect metrics: %v\n", err)
			}
		}
	}
}

func collectAndWrite(ctx context.Context, collector agent.Collector, encoder *json.Encoder) error {
	metrics, err := collector.Collect(ctx)
	if err != nil {
		return err
	}
	if err := encoder.Encode(metrics); err != nil {
		return fmt.Errorf("encode metrics: %w", err)
	}

	return nil
}
