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

func writeSuccessResponse(w http.ResponseWriter, data interface{}) {
	resp := base.Response{
		Data:  data,
		Error: nil,
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.Error("error when trying to write success base.Response", "error", err.Error())
		writeErrorResponse(w, base.ResponseErrorCodeInternalError, "error when trying to write success base.Response")
	}
}

func writeErrorResponse(w http.ResponseWriter, code int, msg string) {
	resp := base.Response{
		Error: &base.ResponseError{
			Code:    code,
			Message: msg,
		},
	}

	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.Error("error when trying to write error base.Response", "error", err.Error())
	}
}
