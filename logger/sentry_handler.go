package logger

import (
	"context"
	"log/slog"

	"github.com/getsentry/sentry-go"
)

type SentryHandler struct {
	slog.Handler
}

func NewSentryHandler(h slog.Handler) *SentryHandler {
	return &SentryHandler{Handler: h}
}

func (h *SentryHandler) Handle(ctx context.Context, r slog.Record) error {
	// First, pass to the underlying handler (e.g., stdout)
	if err := h.Handler.Handle(ctx, r); err != nil {
		return err
	}

	// Only capture Error level logs
	if r.Level >= slog.LevelError {
		hub := sentry.CurrentHub().Clone()

		// Extract attributes to add to Sentry context
		scope := hub.Scope()
		r.Attrs(func(a slog.Attr) bool {
			scope.SetExtra(a.Key, a.Value.Any())
			return true
		})

		// Check for an error attribute to use CaptureException
		var err error
		r.Attrs(func(a slog.Attr) bool {
			if a.Key == "err" || a.Key == "error" {
				if e, ok := a.Value.Any().(error); ok {
					err = e
					return false // stop iteration
				}
			}
			return true
		})

		if err != nil {
			hub.CaptureException(err)
		} else {
			hub.CaptureMessage(r.Message)
		}
	}

	return nil
}

func (h *SentryHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &SentryHandler{Handler: h.Handler.WithAttrs(attrs)}
}

func (h *SentryHandler) WithGroup(name string) slog.Handler {
	return &SentryHandler{Handler: h.Handler.WithGroup(name)}
}
