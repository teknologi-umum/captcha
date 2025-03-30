package requestid

import (
	"context"
	"log/slog"

	"github.com/getsentry/sentry-go"
	"github.com/oklog/ulid/v2"
)

type RequestIdKeyType string

const RequestIdKey RequestIdKeyType = "request_id"

func SetRequestIdOnContext(ctx context.Context) context.Context {
	requestId := ulid.Make().String()
	if hub := sentry.GetHubFromContext(ctx); hub != nil {
		if scope := hub.Scope(); scope != nil {
			scope.SetTag("request_id", requestId)
		}
	}

	if span := sentry.SpanFromContext(ctx); span != nil {
		span.SetTag("request_id", requestId)
	}

	return context.WithValue(ctx, RequestIdKey, requestId)
}

func GetRequestIdFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	requestID, ok := ctx.Value(RequestIdKey).(string)
	if !ok {
		return ""
	}

	return requestID
}

func GetSlogAttributesFromContext(ctx context.Context) []any {
	if ctx == nil {
		return []any{}
	}

	requestID, ok := ctx.Value(RequestIdKey).(string)
	if !ok {
		return []any{}
	}

	attrs := []any{
		slog.String("request_id", requestID),
	}

	return attrs
}
