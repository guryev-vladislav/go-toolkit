package logger

import (
	"context"
	"log/slog"
	"runtime"
)

type callerHandler struct {
	handler slog.Handler
	skip    int
}

func NewCallerHandler(handler slog.Handler, skip int) slog.Handler {
	return &callerHandler{
		handler: handler,
		skip:    skip,
	}
}

func (h *callerHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *callerHandler) Handle(ctx context.Context, r slog.Record) error {
	if pc, file, line, ok := runtime.Caller(h.skip); ok {
		newRecord := slog.NewRecord(r.Time, r.Level, r.Message, pc)
		newRecord.AddAttrs(slog.String("file", file), slog.Int("line", line))

		r.Attrs(func(attr slog.Attr) bool {
			newRecord.AddAttrs(attr)
			return true
		})

		return h.handler.Handle(ctx, newRecord)
	}
	return h.handler.Handle(ctx, r)
}

func (h *callerHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return h
	}
	return &callerHandler{
		handler: h.handler.WithAttrs(attrs),
		skip:    h.skip,
	}
}

func (h *callerHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	return &callerHandler{
		handler: h.handler.WithGroup(name),
		skip:    h.skip,
	}
}
