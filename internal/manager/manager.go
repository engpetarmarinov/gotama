package manager

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/engpetarmarinov/gotama/internal/base"
	"github.com/engpetarmarinov/gotama/internal/broker"
	"github.com/engpetarmarinov/gotama/internal/config"
	"github.com/engpetarmarinov/gotama/internal/logger"
	"log"
	"net/http"
)

type Manager struct {
	server    *http.Server
	scheduler base.Service
	broker    broker.Broker
	config    config.API
}

func NewManager(broker broker.Broker, config config.API) *Manager {
	return &Manager{
		broker:    broker,
		config:    config,
		scheduler: newScheduler(broker, config),
	}
}

func (m *Manager) Shutdown() error {
	logger.Info("manager shutting down...")
	if err := m.scheduler.Shutdown(); err != nil {
		return err
	}
	if err := m.server.Shutdown(context.Background()); err != nil {
		return err
	}
	return nil
}

func (m *Manager) Run() {
	router := NewRouter().RegisterRoutes(m.broker)
	go func(mux http.Handler) {
		server := http.Server{
			Addr:    fmt.Sprintf(":%s", m.config.Get("MANAGER_PORT")),
			Handler: mux,
		}

		m.server = &server
		logger.Info("Listening on", "address", server.Addr)
		if err := server.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}(router)

	m.scheduler.Run()
}

func writeSuccessResponse(w http.ResponseWriter, data interface{}) {
	resp := base.Response{
		Data:  data,
		Error: nil,
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		logger.Error("error when trying to write success base.Response", "error", err.Error())
		writeErrorResponse(w, http.StatusInternalServerError, "error when trying to write success base.Response")
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
		logger.Error("error when trying to write error base.Response", "error", err.Error())
	}
}
