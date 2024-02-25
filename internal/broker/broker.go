package broker

import (
	"context"
	"github.com/engpetarmarinov/gotama/internal/task"
)

type Broker interface {
	Ping(ctx context.Context) error
	Close() error
	GetAllTasks(ctx context.Context, offset int, limit int) (int64, []*task.Message, error)
	GetTask(ctx context.Context, taskID string) (*task.Message, error)
	EnqueueTask(ctx context.Context, msg *task.Message) error
	DequeueTask(ctx context.Context, qname string) (*task.Message, error)
	UpdateTask(ctx context.Context, msg *task.Message) error
	RemoveTask(ctx context.Context, taskID string) error
	RequeueTaskRetry(ctx context.Context, msg *task.Message) error
	RequeueTaskFailed(ctx context.Context, msg *task.Message) error
	MarkTaskAsComplete(ctx context.Context, msg *task.Message) error
}
