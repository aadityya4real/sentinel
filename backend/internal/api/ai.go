package api

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/aadityya4real/sentinel/backend/internal/ai"
	"go.uber.org/zap"
)

// AIHandler serves on-demand infrastructure incident explanations.
type AIHandler struct {
	analyzer ai.Analyzer
	logger   *zap.Logger
}

// NewAIHandler creates an HTTP handler for AI incident analysis requests.
func NewAIHandler(analyzer ai.Analyzer, logger *zap.Logger) (*AIHandler, error) {
	if analyzer == nil {
		return nil, errors.New("AI analyzer is required")
	}
	if logger == nil {
		return nil, errors.New("logger is required")
	}
	return &AIHandler{analyzer: analyzer, logger: logger}, nil
}

// AnalyzeIncident explains a bounded host event window using the configured AI analyzer.
func (h *AIHandler) AnalyzeIncident(writer http.ResponseWriter, request *http.Request) {
	if !acceptsJSON(request.Header.Get("Content-Type")) {
		writeError(writer, http.StatusUnsupportedMediaType, "unsupported_media_type", "Content-Type must be application/json")
		return
	}
	var input incidentAnalysisRequest
	if err := decodeJSON(writer, request, &input); err != nil {
		writeError(writer, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	from, to, err := analysisTimeRange(input.From, input.To)
	if err != nil {
		writeError(writer, http.StatusBadRequest, "invalid_time_range", err.Error())
		return
	}
	analysis, err := h.analyzer.Analyze(request.Context(), ai.Request{Hostname: strings.TrimSpace(input.Hostname), From: from, To: to, EventLimit: input.EventLimit})
	if err != nil {
		var validationError *ai.ValidationError
		var notFoundError *ai.NotFoundError
		switch {
		case errors.As(err, &validationError):
			writeError(writer, http.StatusBadRequest, "invalid_analysis_request", validationError.Error())
		case errors.As(err, &notFoundError):
			writeError(writer, http.StatusNotFound, "analysis_events_not_found", notFoundError.Error())
		case errors.Is(err, ai.ErrDisabled):
			writeError(writer, http.StatusServiceUnavailable, "ai_analyzer_disabled", "AI incident analysis is not configured")
		default:
			h.logger.Error("analyze incident", zap.Error(err), zap.String("hostname", input.Hostname))
			writeError(writer, http.StatusServiceUnavailable, "ai_analyzer_unavailable", "AI incident analysis is temporarily unavailable")
		}
		return
	}
	writeJSON(writer, http.StatusOK, analysis)
}

type incidentAnalysisRequest struct {
	Hostname   string `json:"hostname"`
	From       string `json:"from"`
	To         string `json:"to"`
	EventLimit int    `json:"event_limit"`
}

func analysisTimeRange(rawFrom, rawTo string) (time.Time, time.Time, error) {
	now := time.Now().UTC()
	from := now.Add(-15 * time.Minute)
	to := now
	var err error
	if rawFrom != "" {
		from, err = time.Parse(time.RFC3339, rawFrom)
		if err != nil {
			return time.Time{}, time.Time{}, errors.New("from must be an RFC3339 timestamp")
		}
	}
	if rawTo != "" {
		to, err = time.Parse(time.RFC3339, rawTo)
		if err != nil {
			return time.Time{}, time.Time{}, errors.New("to must be an RFC3339 timestamp")
		}
	}
	return from.UTC(), to.UTC(), nil
}
