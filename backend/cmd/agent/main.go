// Command agent periodically collects host metrics and posts them to the Sentinel API.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/aadityya4real/sentinel/backend/internal/agent"
)

const (
	collectionInterval = 5 * time.Second
	cpuSampleInterval  = 250 * time.Millisecond
	requestTimeout     = 5 * time.Second
	maxSendAttempts    = 3
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	collector, err := agent.NewSystemCollector(cpuSampleInterval)
	if err != nil {
		fmt.Fprintf(os.Stderr, "create collector: %v\n", err)
		os.Exit(1)
	}
	endpoint, err := metricsEndpoint(os.Getenv("SENTINEL_API_URL"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "configure API endpoint: %v\n", err)
		os.Exit(1)
	}
	token := os.Getenv("SENTINEL_API_TOKEN")
	if strings.TrimSpace(token) == "" {
		fmt.Fprintln(os.Stderr, "configure API authentication: SENTINEL_API_TOKEN is required")
		os.Exit(1)
	}
	client := &http.Client{Timeout: requestTimeout}

	if err := collectAndSend(ctx, collector, client, endpoint, token); err != nil {
		log.Printf("send metrics: %v", err)
	}

	ticker := time.NewTicker(collectionInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := collectAndSend(ctx, collector, client, endpoint, token); err != nil {
				log.Printf("send metrics: %v", err)
			}
		}
	}
}

func metricsEndpoint(apiURL string) (string, error) {
	if strings.TrimSpace(apiURL) == "" {
		apiURL = "http://localhost:8080"
	}
	base, err := url.Parse(strings.TrimSpace(apiURL))
	if err != nil || (base.Scheme != "http" && base.Scheme != "https") || base.Host == "" || base.RawQuery != "" || base.Fragment != "" || base.User != nil {
		return "", fmt.Errorf("SENTINEL_API_URL must be an absolute HTTP(S) base URL without query or fragment")
	}
	if base.Scheme != "https" && !isLoopbackHost(base.Hostname()) {
		return "", fmt.Errorf("SENTINEL_API_URL must use HTTPS; HTTP is allowed only for local development endpoints")
	}
	base.Path = strings.TrimRight(base.Path, "/") + "/api/v1/metrics"
	return base.String(), nil
}

func isLoopbackHost(host string) bool {
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func collectAndSend(ctx context.Context, collector agent.Collector, client *http.Client, endpoint, token string) error {
	metrics, err := collector.Collect(ctx)
	if err != nil {
		return fmt.Errorf("collect metrics: %w", err)
	}

	var lastErr error
	for attempt := 1; attempt <= maxSendAttempts; attempt++ {
		if attempt > 1 {
			delay := time.Duration(attempt-1) * 500 * time.Millisecond
			timer := time.NewTimer(delay)
			select {
			case <-ctx.Done():
				timer.Stop()
				return ctx.Err()
			case <-timer.C:
			}
		}
		lastErr = postMetrics(ctx, client, endpoint, token, metrics)
		if lastErr == nil {
			return nil
		}
		var responseErr *apiResponseError
		if errors.As(lastErr, &responseErr) && !responseErr.retryable() {
			return lastErr
		}
	}
	return fmt.Errorf("failed after %d attempts: %w", maxSendAttempts, lastErr)
}

type apiResponseError struct {
	statusCode int
	message    string
}

func (e *apiResponseError) Error() string {
	return fmt.Sprintf("API returned %d: %s", e.statusCode, e.message)
}

func (e *apiResponseError) retryable() bool {
	return e.statusCode == http.StatusTooManyRequests || e.statusCode >= http.StatusInternalServerError
}

func postMetrics(ctx context.Context, client *http.Client, endpoint, token string, metrics agent.Metrics) error {
	payload, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("encode metrics: %w", err)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("create metrics request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+token)
	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("post metrics: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		message, _ := io.ReadAll(io.LimitReader(response.Body, 4*1024))
		return &apiResponseError{statusCode: response.StatusCode, message: strings.TrimSpace(string(message))}
	}
	_, _ = io.Copy(io.Discard, response.Body)

	return nil
}
