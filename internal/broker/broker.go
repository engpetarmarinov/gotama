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
	UpdateTask(ctx context.Context, msg *task.Message) error
	//Dequeue(qnames ...string) (*base.TaskMessage, time.Time, error)
	//Remove(ctx context.Context, msg *base.TaskMessage) error
	//Complete(ctx context.Context, msg *base.TaskMessage) error
	//Schedule(ctx context.Context, msg *base.TaskMessage, processAt time.Time) error
	//WriteResult(qname, id string, data []byte) (n int, err error)
}
