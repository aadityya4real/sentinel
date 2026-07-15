package ai

import (
	"context"
	"errors"
)

// ErrDisabled indicates that AI incident analysis is not configured for this server.
var ErrDisabled = errors.New("AI incident analyzer is disabled")

// DisabledAnalyzer keeps the server available when AI configuration is intentionally absent.
type DisabledAnalyzer struct{}

// Analyze reports that the analyzer is disabled.
func (DisabledAnalyzer) Analyze(context.Context, Request) (Analysis, error) {
	return Analysis{}, ErrDisabled
}
