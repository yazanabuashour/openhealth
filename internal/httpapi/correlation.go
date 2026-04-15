package httpapi

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

type correlationIDKey struct{}

func withCorrelationID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		correlationID := r.Header.Get("X-Correlation-Id")
		if correlationID == "" {
			correlationID = newCorrelationID()
		}

		w.Header().Set("X-Correlation-Id", correlationID)
		ctx := context.WithValue(r.Context(), correlationIDKey{}, correlationID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func correlationIDFromContext(ctx context.Context) string {
	if value, ok := ctx.Value(correlationIDKey{}).(string); ok && value != "" {
		return value
	}
	return "unknown"
}

func newCorrelationID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "generated-correlation-id"
	}
	return hex.EncodeToString(buf)
}
