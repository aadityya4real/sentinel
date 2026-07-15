package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"

	"github.com/aadityya4real/sentinel/backend/internal/agent"
	"github.com/aadityya4real/sentinel/backend/internal/collector"
	"go.uber.org/zap"
)

const maxMetricsPayloadSize = 1 << 20

var errPayloadTooLarge = errors.New("metrics payload exceeds the maximum allowed size")

// MetricsHandler receives and records metrics submitted by Sentinel agents.
type MetricsHandler struct {
	recorder collector.Recorder
	logger   *zap.Logger
}

// NewMetricsHandler creates an HTTP handler for agent metrics ingestion.
func NewMetricsHandler(recorder collector.Recorder, logger *zap.Logger) (*MetricsHandler, error) {
	if recorder == nil {
		return nil, errors.New("metrics recorder is required")
	}
	if logger == nil {
		return nil, errors.New("logger is required")
	}
	return &MetricsHandler{recorder: recorder, logger: logger}, nil
}

// ServeHTTP validates a metrics request and records it with the collector service.
func (h *MetricsHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	if !acceptsJSON(request.Header.Get("Content-Type")) {
		writeError(writer, http.StatusUnsupportedMediaType, "unsupported_media_type", "Content-Type must be application/json")
		return
	}
	if request.ContentLength > maxMetricsPayloadSize {
		writeError(writer, http.StatusRequestEntityTooLarge, "payload_too_large", errPayloadTooLarge.Error())
		return
	}

	var metrics agent.Metrics
	if err := decodeJSON(writer, request, &metrics); err != nil {
		if errors.Is(err, errPayloadTooLarge) {
			writeError(writer, http.StatusRequestEntityTooLarge, "payload_too_large", err.Error())
			return
		}
		writeError(writer, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}

	if err := h.recorder.Record(request.Context(), metrics); err != nil {
		var validationError *collector.ValidationError
		if errors.As(err, &validationError) {
			writeError(writer, http.StatusUnprocessableEntity, "validation_failed", validationError.Error())
			return
		}
		h.logger.Error("record metrics", zap.Error(err), zap.String("hostname", metrics.Hostname))
		writeError(writer, http.StatusServiceUnavailable, "metrics_unavailable", "metrics storage is temporarily unavailable")
		return
	}

	writeJSON(writer, http.StatusAccepted, map[string]string{"status": "accepted"})
}

func acceptsJSON(contentType string) bool {
	if contentType == "" {
		return false
	}
	mediaType, _, err := mime.ParseMediaType(contentType)
	return err == nil && mediaType == "application/json"
}

func decodeJSON(writer http.ResponseWriter, request *http.Request, destination any) error {
	request.Body = http.MaxBytesReader(writer, request.Body, maxMetricsPayloadSize)
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(destination); err != nil {
		var maxBytesError *http.MaxBytesError
		if errors.As(err, &maxBytesError) {
			return fmt.Errorf("%w (%d bytes)", errPayloadTooLarge, maxMetricsPayloadSize)
		}
		return fmt.Errorf("invalid JSON payload: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		var maxBytesError *http.MaxBytesError
		if errors.As(err, &maxBytesError) {
			return fmt.Errorf("%w (%d bytes)", errPayloadTooLarge, maxMetricsPayloadSize)
		}
		return errors.New("request body must contain exactly one JSON object")
	}
	return nil
}
