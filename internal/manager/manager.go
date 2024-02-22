package manager

import (
	"encoding/json"
	"fmt"
	"github.com/engpetarmarinov/gotama/internal/base"
	"github.com/engpetarmarinov/gotama/internal/config"
	"log"
	"log/slog"
	"net/http"
)

type API interface {
	Run()
}

type Manager struct {
}

func NewManager() *Manager {
	return &Manager{}
}

func (m *Manager) Run(config config.API) {
	router := NewRouter().RegisterRoutes()
	go func(mux http.Handler) {
		server := http.Server{
			Addr:    fmt.Sprintf(":%s", config.Get("MANAGER_PORT")),
			Handler: mux,
		}
		slog.Info("Listening on", "address", server.Addr)
		if err := server.ListenAndServe(); err != nil {
			log.Fatal(err)
		}

	}(router)
}

func getTasksHandler(w http.ResponseWriter, r *http.Request) {
	//TODO: get tasks
	resp := []base.Task{
		{Status: base.TaskStatusPending},
	}
	writeSuccessResponse(w, resp)
}

func getTaskHandler(w http.ResponseWriter, r *http.Request) {
	//TODO: get task
	resp := base.Task{
		Status: base.TaskStatusPending,
	}
	writeSuccessResponse(w, resp)
}

func postTaskHandler(w http.ResponseWriter, r *http.Request) {
	//TODO: add task
	resp := base.Task{
		Status: base.TaskStatusPending,
	}
	writeSuccessResponse(w, resp)
}

func putTaskHandler(w http.ResponseWriter, r *http.Request) {
	//TODO: add task
	resp := base.Task{
		Status: base.TaskStatusPending,
	}
	writeSuccessResponse(w, resp)
}

func writeSuccessResponse(w http.ResponseWriter, data interface{}) {
	resp := base.Response{
		Data:  data,
		Error: nil,
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.Error("Error when trying to write success base.Response", "error", err.Error())
		writeErrorResponse(w, base.ResponseErrorCodeInternalError, "Internal Error Occurred")
	}
}

func writeErrorResponse(w http.ResponseWriter, code int, msg string) {
	resp := base.Response{
		Error: &base.ResponseError{
			Code:    code,
			Message: msg,
		},
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.Error("Error when trying to write error base.Response", "error", err.Error())
	}
}
