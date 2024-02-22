package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

func WithLogging(next http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		start := time.Now()
		uri := r.RequestURI
		method := r.Method
		next.ServeHTTP(rw, r)
		duration := time.Since(start)

		slog.Info("Response",
			"method", method,
			"uri", uri,
			"duration", duration)
	}
}
