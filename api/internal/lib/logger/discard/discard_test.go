package discard

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDiscardLogger(t *testing.T) {
	t.Run("happy_path", func(t *testing.T) {
		logger := NewDiscardLogger()
		assert.NotNil(t, logger)
	})
}

func TestNewDiscardHandler(t *testing.T) {
	t.Run("happy_path", func(t *testing.T) {
		handler := NewDiscardHandler()
		assert.NotNil(t, handler)
	})
}

func TestHandle(t *testing.T) {
	t.Run("happy_path", func(t *testing.T) {
		h := NewDiscardHandler()

		err := h.Handle(context.Background(), slog.Record{})
		assert.NoError(t, err)
	})
}

func TestEnabled(t *testing.T) {
	t.Run("happy_path", func(t *testing.T) {
		h := NewDiscardHandler()

		ok := h.Enabled(context.Background(), slog.Level(-1))
		assert.False(t, ok)
	})
}

func TestWithAttrs(t *testing.T) {
	t.Run("happy_path", func(t *testing.T) {
		h := NewDiscardHandler()

		withAttr := h.WithAttrs([]slog.Attr{})
		assert.Same(t, h, withAttr)
	})
}

func TestWithGroup(t *testing.T) {
	t.Run("happy_path", func(t *testing.T) {
		h := NewDiscardHandler()

		withGroup := h.WithGroup("hello world")
		assert.Same(t, h, withGroup)
	})
}
