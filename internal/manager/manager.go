package manager

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/engpetarmarinov/gotama/internal/base"
	"github.com/engpetarmarinov/gotama/internal/broker"
	"github.com/engpetarmarinov/gotama/internal/config"
	"log"
	"log/slog"
	"net/http"
)

type API interface {
	Run()
	Shutdown()
}

type Manager struct {
	ctx    context.Context
	server *http.Server
	broker broker.Broker
	config config.API
	cancel context.CancelFunc
}

func NewManager(broker broker.Broker, config config.API) *Manager {
	ctx, cancel := context.WithCancel(context.Background())
	return &Manager{
		ctx:    ctx,
		broker: broker,
		config: config,
		cancel: cancel,
	}
}

func (m *Manager) Shutdown() error {
	if err := m.server.Shutdown(m.ctx); err != nil {
		return err
	}
	m.cancel()
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
		slog.Info("Listening on", "address", server.Addr)
		if err := server.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}(router)

	newScheduler(m.ctx, m.broker, m.config).Run()
}

func writeSuccessResponse(w http.ResponseWriter, data interface{}) {
	resp := base.Response{
		Data:  data,
		Error: nil,
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.Error("error when trying to write success base.Response", "error", err.Error())
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
		slog.Error("error when trying to write error base.Response", "error", err.Error())
	}
}
