package worker

import (
	"context"
	"errors"
	"github.com/engpetarmarinov/gotama/internal/base"
	"github.com/engpetarmarinov/gotama/internal/broker"
	"github.com/engpetarmarinov/gotama/internal/config"
	"github.com/engpetarmarinov/gotama/internal/processors"
	"github.com/engpetarmarinov/gotama/internal/task"
	"github.com/engpetarmarinov/gotama/internal/timeutil"
	"log/slog"
	"runtime/debug"
	"strconv"
	"sync"
	"time"
)

var maxRetry = 3

type API interface {
	Run()
	Shutdown()
}

type Worker struct {
	wg     *sync.WaitGroup
	broker broker.Broker
	config config.API
	clock  timeutil.Clock
}

func NewWorker(broker broker.Broker, config config.API, clock timeutil.Clock) *Worker {
	wg := &sync.WaitGroup{}
	return &Worker{
		wg:     wg,
		broker: broker,
		config: config,
		clock:  clock,
	}
}

func (w *Worker) Run() {
	workerGoroutinesStr := w.config.Get("WORKER_GOROUTINES")
	workerGoroutines, err := strconv.Atoi(workerGoroutinesStr)
	if err != nil {
		panic(err.Error())
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for i := 0; i < workerGoroutines; i++ {
		w.wg.Add(1)
		//TODO: add heartbeat and deadline?
		go func(ctx context.Context, wg *sync.WaitGroup) {
			defer wg.Done()
			tick := time.Tick(time.Second)
			for {
				select {
				case <-ctx.Done():
					slog.Info("worker goroutine received done")
					return
				case <-tick:
					err := w.exec(ctx)
					if errors.Is(err, base.ErrorNoTasksInQueue) {
						slog.Info("no tasks in queue")
					} else if err != nil {
						slog.Error("worker exec error", "err", err.Error())
					}
				}
			}

		}(ctx, w.wg)
	}

	w.wg.Wait()
}

func (w *Worker) Shutdown() error {
	//TODO: graceful shutdown
	return nil
}

func (w *Worker) exec(ctx context.Context) error {
	//handle eventual panic in processors, we don't want the worker to stop
	defer func() {
		if r := recover(); r != nil {
			errMsg := string(debug.Stack())
			slog.Error("recovering from panic. stack trace:\n%s", errMsg)
		}
	}()

	msg, err := w.broker.DequeueTask(ctx, task.QueueDefault)
	if err != nil {
		return err
	}

	msgName, err := task.GetName(msg.Name)
	if err != nil {
		return err
	}

	msg.Status = task.StatusRunning
	err = w.broker.UpdateTask(ctx, msg)
	if err != nil {
		return err
	}

	processor, err := processors.ProcessorFactory(msgName)
	if err != nil {
		return err
	}

	err = processor.ProcessTask(ctx, msg)
	if err != nil {
		w.handleProcessTaskError(ctx, msg, err)
		return err
	}

	msg.Status = task.StatusSucceeded
	msg.CompletedAt = w.clock.Now()
	//Reset NumRetries for recurring tasks
	if msg.Type == task.TypeRecurring {
		msg.NumRetries = 0
	}

	err = w.broker.UpdateTask(ctx, msg)
	if err != nil {
		return err
	}

	return w.broker.MarkTaskAsComplete(ctx, msg)
}

func (w *Worker) handleProcessTaskError(ctx context.Context, msg *task.Message, err error) {
	msg.Status = task.StatusFailed
	msg.Error = err.Error()
	msg.FailedAt = w.clock.Now()
	upErr := w.broker.UpdateTask(ctx, msg)
	if upErr != nil {
		slog.Error("error updating task", "err", upErr)
	}

	if msg.NumRetries < maxRetry {
		scheduleErr := w.broker.RequeueTaskRetry(ctx, msg)
		if scheduleErr != nil {
			slog.Error("error schedule retry", "err", upErr)
		}
	} else {
		//dead letter queue
		scheduleErr := w.broker.RequeueTaskFailed(ctx, msg)
		if scheduleErr != nil {
			slog.Error("error schedule retry", "err", upErr)
		}
	}
}
