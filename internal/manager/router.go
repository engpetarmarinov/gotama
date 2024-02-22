package manager

import (
	mw "github.com/engpetarmarinov/gotama/internal/middleware"
	"net/http"
)

type Router struct {
	mux *http.ServeMux
}

func NewRouter() *Router {
	return &Router{
		mux: http.NewServeMux(),
	}
}

func (r *Router) RegisterRoutes() http.Handler {
	r.mux.HandleFunc(
		"GET /api/v1/tasks",
		mw.WithLogging(mw.WithCommonHeaders(mw.WithAuth(mw.WithRBAC(getTasksHandler)))))

	r.mux.HandleFunc(
		"GET /api/v1/tasks/{id}",
		mw.WithLogging(mw.WithCommonHeaders(mw.WithAuth(mw.WithRBAC(getTaskHandler)))))

	r.mux.HandleFunc(
		"POST /api/v1/tasks",
		mw.WithLogging(mw.WithCommonHeaders(mw.WithAuth(mw.WithRBAC(postTaskHandler)))))

	r.mux.HandleFunc(
		"PUT /api/v1/tasks/{id}",
		mw.WithLogging(mw.WithCommonHeaders(mw.WithAuth(mw.WithRBAC(putTaskHandler)))))

	r.mux.HandleFunc(
		"DELETE /api/v1/tasks/{id}",
		mw.WithLogging(mw.WithCommonHeaders(mw.WithAuth(mw.WithRBAC(deleteTaskHandler)))))

	return r.mux
}
