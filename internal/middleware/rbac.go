package middleware

import (
	"net/http"
)

func WithRBAC(next http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		//TODO: implement RBAC checks
		next.ServeHTTP(rw, r)
	}
}
