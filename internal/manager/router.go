package manager

import (
	"github.com/engpetarmarinov/gotama/internal/broker"
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

func (r *Router) RegisterRoutes(broker broker.Broker) http.Handler {
	//TODO: add swagger
	r.mux.HandleFunc(
		"GET /api/v1/tasks",
		mw.WithLogging(mw.WithCommonHeaders(mw.WithAuth(mw.WithRBAC(getTasksHandler(broker))))))

	r.mux.HandleFunc(
		"GET /api/v1/tasks/{id}",
		mw.WithLogging(mw.WithCommonHeaders(mw.WithAuth(mw.WithRBAC(getTaskHandler(broker))))))

	r.mux.HandleFunc(
		"POST /api/v1/tasks",
		mw.WithLogging(mw.WithCommonHeaders(mw.WithAuth(mw.WithRBAC(postTaskHandler(broker))))))

	r.mux.HandleFunc(
		"PUT /api/v1/tasks/{id}",
		mw.WithLogging(mw.WithCommonHeaders(mw.WithAuth(mw.WithRBAC(putTaskHandler(broker))))))

	r.mux.HandleFunc(
		"DELETE /api/v1/tasks/{id}",
		mw.WithLogging(mw.WithCommonHeaders(mw.WithAuth(mw.WithRBAC(deleteTaskHandler(broker))))))

	return r.mux
}
