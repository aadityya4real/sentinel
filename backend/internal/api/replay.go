package api

import (
	"errors"
	"net/http"
	"strings"

	"github.com/aadityya4real/sentinel/backend/internal/replay"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

const maxReplayAPILimit = 999

// ReplayHandler serves chronological infrastructure timelines for the Time Machine.
type ReplayHandler struct {
	reader replay.Reader
	logger *zap.Logger
}

// NewReplayHandler creates an HTTP handler for infrastructure replay requests.
func NewReplayHandler(reader replay.Reader, logger *zap.Logger) (*ReplayHandler, error) {
	if reader == nil {
		return nil, errors.New("replay reader is required")
	}
	if logger == nil {
		return nil, errors.New("logger is required")
	}

	return &ReplayHandler{reader: reader, logger: logger}, nil
}

// Replay returns an ordered, cursor-paginated event timeline for one host.
func (h *ReplayHandler) Replay(writer http.ResponseWriter, request *http.Request) {
	hostname := strings.TrimSpace(chi.URLParam(request, "hostname"))
	from, to, err := parseTimeRange(request.URL.Query().Get("from"), request.URL.Query().Get("to"))
	if err != nil {
		writeError(writer, http.StatusBadRequest, "invalid_time_range", err.Error())
		return
	}
	limit, err := parseLimit(request.URL.Query().Get("limit"), defaultHistoryLimit, maxReplayAPILimit)
	if err != nil {
		writeError(writer, http.StatusBadRequest, "invalid_limit", err.Error())
		return
	}

	timeline, err := h.reader.Replay(request.Context(), replay.Request{
		Hostname: hostname,
		From:     from,
		To:       to,
		Limit:    limit,
		Cursor:   request.URL.Query().Get("cursor"),
	})
	if err != nil {
		var validationError *replay.ValidationError
		if errors.As(err, &validationError) {
			writeError(writer, http.StatusBadRequest, "invalid_replay_request", validationError.Error())
			return
		}

		h.logger.Error("replay host timeline", zap.Error(err), zap.String("hostname", hostname))
		writeError(writer, http.StatusServiceUnavailable, "replay_unavailable", "replay data is temporarily unavailable")
		return
	}

	writeJSON(writer, http.StatusOK, timeline)
}
