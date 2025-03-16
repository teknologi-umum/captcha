package main

import (
	"context"
	"log/slog"

	"github.com/getsentry/sentry-go"
)

type SlogSentryBreadcrumb struct {
	Enable     bool
	Level      slog.Level
	attributes []slog.Attr
	group      string
}

var _ slog.Handler = (*SlogSentryBreadcrumb)(nil)

func toSentryLevel(level slog.Level) sentry.Level {
	switch level {
	case slog.LevelDebug:
		return sentry.LevelDebug
	case slog.LevelInfo:
		return sentry.LevelInfo
	case slog.LevelWarn:
		return sentry.LevelWarning
	case slog.LevelError:
		return sentry.LevelError
	default:
		return sentry.LevelInfo
	}
}

func (s *SlogSentryBreadcrumb) Enabled(_ context.Context, _ slog.Level) bool {
	return s.Enable
}

func (s *SlogSentryBreadcrumb) Handle(ctx context.Context, record slog.Record) error {
	hub := sentry.GetHubFromContext(ctx)
	if hub == nil {
		return nil
	}

	var data = make(map[string]any)
	for _, attr := range s.attributes {
		data[attr.Key] = attr.Value
	}
	if s.group != "" {
		data["group"] = s.group
	}

	hub.AddBreadcrumb(&sentry.Breadcrumb{
		Type:      "log",
		Category:  "log",
		Message:   record.Message,
		Data:      data,
		Level:     toSentryLevel(record.Level),
		Timestamp: record.Time,
	}, nil)

	return nil
}

func (s *SlogSentryBreadcrumb) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &SlogSentryBreadcrumb{
		Enable:     s.Enable,
		Level:      s.Level,
		attributes: append(s.attributes, attrs...),
		group:      s.group,
	}
}

func (s *SlogSentryBreadcrumb) WithGroup(name string) slog.Handler {
	return &SlogSentryBreadcrumb{
		Enable:     s.Enable,
		Level:      s.Level,
		attributes: s.attributes,
		group:      name,
	}
}
