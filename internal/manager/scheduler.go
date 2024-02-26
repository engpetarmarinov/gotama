package manager

import (
	"context"
	"github.com/engpetarmarinov/gotama/internal/broker"
	"github.com/engpetarmarinov/gotama/internal/config"
	"log/slog"
	"time"
)

type schedulerAPI interface {
	Run()
}

type scheduler struct {
	ctx    context.Context
	broker broker.Broker
	config config.API
}

func newScheduler(ctx context.Context, broker broker.Broker, config config.API) *scheduler {
	return &scheduler{
		ctx:    ctx,
		broker: broker,
		config: config,
	}
}

func (s *scheduler) Run() {
	go func() {
		duration := time.Second
		tick := time.Tick(duration)
		slog.Info("scheduler started", "period", duration.String())
		for {
			select {
			case <-s.ctx.Done():
				//TODO: graceful shutdown?
				slog.Info("scheduler received done")
				break
			case <-tick:
				slog.Info("scheduler checking for scheduled tasks...")
				err := s.broker.EnqueueScheduledTasks(s.ctx)
				if err != nil {
					slog.Error("scheduler error during enqueueing scheduled tasks", "err", err)
				}
			}
		}
	}()
}
