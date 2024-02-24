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
	//Dequeue(qnames ...string) (*task.TaskMessage, time.Time, error)
	UpdateTask(ctx context.Context, msg *task.Message) error
	RemoveTask(ctx context.Context, taskID string) error
	//Complete(ctx context.Context, msg *task.TaskMessage) error
	//Schedule(ctx context.Context, msg *task.TaskMessage, processAt time.Time) error
	//WriteResult(qname, id string, data []byte) (n int, err error)
}
