package api

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/aadityya4real/sentinel/backend/internal/timemachine"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// TimeMachineHandler serves point-in-time infrastructure state reconstruction requests.
type TimeMachineHandler struct {
	reader timemachine.Reader
	logger *zap.Logger
}

// NewTimeMachineHandler creates an HTTP handler for Time Machine snapshot requests.
func NewTimeMachineHandler(reader timemachine.Reader, logger *zap.Logger) (*TimeMachineHandler, error) {
	if reader == nil {
		return nil, errors.New("time machine reader is required")
	}
	if logger == nil {
		return nil, errors.New("logger is required")
	}

	return &TimeMachineHandler{reader: reader, logger: logger}, nil
}

// Snapshot returns a host's most recent state at or before the requested timestamp.
func (h *TimeMachineHandler) Snapshot(writer http.ResponseWriter, request *http.Request) {
	hostname := strings.TrimSpace(chi.URLParam(request, "hostname"))
	at, err := parseSnapshotTime(request.URL.Query().Get("at"))
	if err != nil {
		writeError(writer, http.StatusBadRequest, "invalid_snapshot_time", err.Error())
		return
	}

	snapshot, err := h.reader.Snapshot(request.Context(), timemachine.Request{Hostname: hostname, At: at})
	if err != nil {
		var validationError *timemachine.ValidationError
		if errors.As(err, &validationError) {
			writeError(writer, http.StatusBadRequest, "invalid_time_machine_request", validationError.Error())
			return
		}
		var notFoundError *timemachine.NotFoundError
		if errors.As(err, &notFoundError) {
			writeError(writer, http.StatusNotFound, "snapshot_not_found", notFoundError.Error())
			return
		}

		h.logger.Error("reconstruct time machine snapshot", zap.Error(err), zap.String("hostname", hostname))
		writeError(writer, http.StatusServiceUnavailable, "time_machine_unavailable", "time machine data is temporarily unavailable")
		return
	}

	writeJSON(writer, http.StatusOK, snapshot)
}

func parseSnapshotTime(value string) (time.Time, error) {
	if value == "" {
		return time.Time{}, errors.New("at is required and must be an RFC3339 timestamp")
	}

	at, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, errors.New("at must be an RFC3339 timestamp")
	}

	return at.UTC(), nil
}
