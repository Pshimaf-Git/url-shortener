// Package discard provides a no-op log handler that discards all log output.
//
// It implements slog.Handler for use in environments where logging should be
// silenced, such as:
// - Testing environments
// - Production systems with strict performance requirements
// - Cases where logging is explicitly disabled
//
// The handler is thread-safe and incurs minimal overhead since all operations
// are no-ops.
package discard

import (
	"context"
	"log/slog"
)

// Verify DiscardHandler implements slog.Handler at compile time
var _ slog.Handler = &DiscardHandler{}

// DiscardHandler is a no-op log handler that discards all log records.
// It satisfies the slog.Handler interface while performing no actual I/O.
type DiscardHandler struct{}

// NewDiscardLogger creates a new slog.Logger that discards all output.
func NewDiscardLogger() *slog.Logger { return slog.New(NewDiscardHandler()) }

// NewDiscardHandler creates a new DiscardHandler instance.
func NewDiscardHandler() *DiscardHandler { return &DiscardHandler{} }

// Handle processes the log record by doing nothing.
// Implements slog.Handler interface.
func (h *DiscardHandler) Handle(_ context.Context, _ slog.Record) error { return nil }

// Enabled reports whether logging is enabled at the given level.
// Implements slog.Handler interface.
func (h *DiscardHandler) Enabled(_ context.Context, _ slog.Level) bool { return false }

// WithAttrs returns a new handler with additional attributes.
// Implements slog.Handler interface.
func (h *DiscardHandler) WithAttrs(_ []slog.Attr) slog.Handler { return h }

// WithGroup returns a new handler with the given group name.
// Implements slog.Handler interface.
func (h *DiscardHandler) WithGroup(_ string) slog.Handler { return h }
