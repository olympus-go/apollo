package spotify

import (
	"context"
	"log/slog"
)

type nopLogHandler struct{}

func (n nopLogHandler) Enabled(_ context.Context, _ slog.Level) bool  { return false }
func (n nopLogHandler) Handle(_ context.Context, _ slog.Record) error { return nil }
func (n nopLogHandler) WithAttrs(_ []slog.Attr) slog.Handler          { return n }
func (n nopLogHandler) WithGroup(_ string) slog.Handler               { return n }
