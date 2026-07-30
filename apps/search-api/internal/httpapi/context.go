package httpapi

import (
	"context"
	"net/http"
)

func withRequestID(
	ctx context.Context,
	requestID string,
) context.Context {
	return context.WithValue(ctx, requestIDContextKey, requestID)
}

func requestIDFromRequest(request *http.Request) string {
	value, _ := request.Context().Value(
		requestIDContextKey,
	).(string)

	return value
}
