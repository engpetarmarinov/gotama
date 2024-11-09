package worker

import (
	"context"
	"errors"
	"github.com/engpetarmarinov/gotama/internal/base"
	"github.com/engpetarmarinov/gotama/internal/config"
	"github.com/engpetarmarinov/gotama/internal/logger"
	"github.com/engpetarmarinov/gotama/internal/processors"
	"github.com/engpetarmarinov/gotama/internal/task"
	"github.com/engpetarmarinov/gotama/internal/timeutil"
	"strconv"
	"sync"
	"time"
)

var maxRetry = 3

type Broker interface {
	UpdateTask(ctx context.Context, msg *task.Message) error
	DequeueTask(ctx context.Context, qname string) (*task.Message, error)
	MarkTaskAsComplete(ctx context.Context, msg *task.Message) error
	RequeueTaskFailed(ctx context.Context, msg *task.Message) error
	RequeueTaskRetry(ctx context.Context, msg *task.Message) error
}

type Worker struct {
	wg     *sync.WaitGroup
	broker Broker
	config config.API
	clock  timeutil.Clock
	cancel context.CancelFunc
}

func NewWorker(config config.API, broker Broker, clock timeutil.Clock) *Worker {
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

	workerCtx, workerCancel := context.WithCancel(context.Background())
	w.cancel = workerCancel

	for i := 0; i < workerGoroutines; i++ {
		w.wg.Add(1)
		go func(ctx context.Context, wg *sync.WaitGroup) {
			defer wg.Done()
			tick := time.Tick(time.Second)
			for {
				select {
				case <-ctx.Done():
					logger.Info("worker goroutine received done")
					return
				case <-tick:
					err := exec(context.Background(), w.config, w.broker, w.clock)
					if errors.Is(err, base.ErrorNoTasksInQueue) {
						logger.Info("no tasks in queue")
					} else if err != nil {
						logger.Error("worker exec error", "error", err)
					}
				}
			}

		}(workerCtx, w.wg)
	}
}

func (w *Worker) Shutdown() error {
	logger.Info("worker shutting down...")
	w.cancel()
	w.wg.Wait()
	logger.Info("worker gracefully shut down all goroutines")
	return nil
}

func exec(ctx context.Context, config config.API, broker Broker, clock timeutil.Clock) error {
	//handle eventual panic in processors, we don't want the worker to stop
	defer func() {
		if r := recover(); r != nil {
			logger.Error("recovering from panic", "error", r)
		}
	}()

	msg, err := broker.DequeueTask(ctx, task.QueueDefault)
	if err != nil {
		return err
	}

	msgName, err := task.GetName(msg.Name)
	if err != nil {
		return err
	}

	msg.Status = task.StatusRunning
	err = broker.UpdateTask(ctx, msg)
	if err != nil {
		return err
	}

	processor, err := processors.ProcessorFactory(config, msgName)
	if err != nil {
		return err
	}

	taskDeadline, err := time.ParseDuration(config.Get("WORKER_TASK_DEADLINE"))
	if err != nil {
		return err
	}

	taskCtx, taskCancel := context.WithDeadline(context.Background(), clock.Now().Add(taskDeadline))
	defer taskCancel()
	err = processor.ProcessTask(taskCtx, msg)
	if err != nil {
		handleProcessTaskError(ctx, broker, clock, msg, err)
		return err
	}

	msg.Status = task.StatusSucceeded
	now := clock.Now()
	msg.CompletedAt = &now
	//Reset NumRetries for recurring tasks
	if msg.Type == task.TypeRecurring {
		msg.NumRetries = 0
	}

	err = broker.UpdateTask(ctx, msg)
	if err != nil {
		return err
	}

	return broker.MarkTaskAsComplete(ctx, msg)
}

func handleProcessTaskError(ctx context.Context, broker Broker, clock timeutil.Clock, msg *task.Message, err error) {
	msg.Status = task.StatusFailed
	errStr := err.Error()
	msg.Error = &errStr
	now := clock.Now()
	msg.FailedAt = &now
	msg.NumRetries = msg.NumRetries + 1
	upErr := broker.UpdateTask(ctx, msg)
	if upErr != nil {
		logger.Error("error updating task when handling task error", "error", upErr)
	}

	if msg.NumRetries < maxRetry {
		scheduleErr := broker.RequeueTaskRetry(ctx, msg)
		if scheduleErr != nil {
			logger.Error("error scheduling retry", "error", scheduleErr)
		}
	} else {
		//dead letter queue
		requeueFailedErr := broker.RequeueTaskFailed(ctx, msg)
		if requeueFailedErr != nil {
			logger.Error("error scheduling retry", "error", requeueFailedErr)
		}
	}
}
