package manager

import (
	"context"
	"github.com/engpetarmarinov/gotama/internal/broker"
	"github.com/engpetarmarinov/gotama/internal/config"
	"github.com/engpetarmarinov/gotama/internal/logger"
	"time"
)

type scheduler struct {
	ctx    context.Context
	broker broker.Broker
	config config.API
	cancel context.CancelFunc
}

func newScheduler(broker broker.Broker, config config.API) *scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	return &scheduler{
		ctx:    ctx,
		broker: broker,
		config: config,
		cancel: cancel,
	}
}

func (s *scheduler) Run() {
	go func() {
		duration := time.Second
		tick := time.Tick(duration)
		logger.Info("scheduler started", "period", duration.String())
		for {
			select {
			case <-s.ctx.Done():
				logger.Info("scheduler goroutine received done")
				return
			case <-tick:
				logger.Info("scheduler checking for scheduled tasks...")
				err := s.broker.EnqueueScheduledTasks(s.ctx)
				if err != nil {
					logger.Error("scheduler error during enqueueing scheduled tasks", "err", err)
				}
			}
		}
	}()
}

func (s *scheduler) Shutdown() error {
	logger.Info("scheduler shutting down...")
	s.cancel()
	return nil
}
