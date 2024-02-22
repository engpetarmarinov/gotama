package middleware

import (
	"net/http"
)

func WithAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		//TODO: implement authentication checks
		next.ServeHTTP(rw, r)
	}
}
