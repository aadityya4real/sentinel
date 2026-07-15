package api

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/aadityya4real/sentinel/backend/internal/dashboard"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

const (
	defaultHostsLimit   = 100
	maxHostsLimit       = 250
	defaultHistoryLimit = 300
	maxHistoryLimit     = 1000
	maxHistoryRange     = 7 * 24 * time.Hour
)

// DashboardHandler serves read-only dashboard views of collected infrastructure metrics.
type DashboardHandler struct {
	reader dashboard.Reader
	logger *zap.Logger
}

// NewDashboardHandler creates an HTTP handler for dashboard read APIs.
func NewDashboardHandler(reader dashboard.Reader, logger *zap.Logger) (*DashboardHandler, error) {
	if reader == nil {
		return nil, errors.New("dashboard reader is required")
	}
	if logger == nil {
		return nil, errors.New("logger is required")
	}

	return &DashboardHandler{reader: reader, logger: logger}, nil
}

// Overview returns aggregate fleet metrics for dashboard summary cards.
func (h *DashboardHandler) Overview(writer http.ResponseWriter, request *http.Request) {
	overview, err := h.reader.GetOverview(request.Context())
	if err != nil {
		h.writeUnavailable(writer, "load dashboard overview", err)
		return
	}

	writeJSON(writer, http.StatusOK, overview)
}

// Hosts returns current metrics and freshness status for the requested number of hosts.
func (h *DashboardHandler) Hosts(writer http.ResponseWriter, request *http.Request) {
	limit, err := parseLimit(request.URL.Query().Get("limit"), defaultHostsLimit, maxHostsLimit)
	if err != nil {
		writeError(writer, http.StatusBadRequest, "invalid_limit", err.Error())
		return
	}

	hosts, err := h.reader.GetHosts(request.Context(), limit)
	if err != nil {
		h.writeUnavailable(writer, "load dashboard hosts", err)
		return
	}

	writeJSON(writer, http.StatusOK, hosts)
}

// History returns a bounded metric time series for a selected host.
func (h *DashboardHandler) History(writer http.ResponseWriter, request *http.Request) {
	hostname := strings.TrimSpace(chi.URLParam(request, "hostname"))
	if hostname == "" || len(hostname) > 255 {
		writeError(writer, http.StatusBadRequest, "invalid_hostname", "hostname must be between 1 and 255 characters")
		return
	}

	from, to, err := parseTimeRange(request.URL.Query().Get("from"), request.URL.Query().Get("to"))
	if err != nil {
		writeError(writer, http.StatusBadRequest, "invalid_time_range", err.Error())
		return
	}
	limit, err := parseLimit(request.URL.Query().Get("limit"), defaultHistoryLimit, maxHistoryLimit)
	if err != nil {
		writeError(writer, http.StatusBadRequest, "invalid_limit", err.Error())
		return
	}

	history, err := h.reader.GetHistory(request.Context(), hostname, from, to, limit)
	if err != nil {
		h.writeUnavailable(writer, "load host metric history", err)
		return
	}

	writeJSON(writer, http.StatusOK, history)
}

func (h *DashboardHandler) writeUnavailable(writer http.ResponseWriter, operation string, err error) {
	h.logger.Error(operation, zap.Error(err))
	writeError(writer, http.StatusServiceUnavailable, "dashboard_unavailable", "dashboard data is temporarily unavailable")
}

func parseLimit(raw string, defaultValue, maximum int) (int, error) {
	if raw == "" {
		return defaultValue, nil
	}

	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 || value > maximum {
		return 0, fmt.Errorf("limit must be an integer between 1 and %d", maximum)
	}

	return value, nil
}

func parseTimeRange(rawFrom, rawTo string) (time.Time, time.Time, error) {
	now := time.Now().UTC()
	from := now.Add(-time.Hour)
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
	if from.After(to) {
		return time.Time{}, time.Time{}, errors.New("from must be before or equal to to")
	}
	if to.Sub(from) > maxHistoryRange {
		return time.Time{}, time.Time{}, errors.New("time range must not exceed seven days")
	}

	return from.UTC(), to.UTC(), nil
}
