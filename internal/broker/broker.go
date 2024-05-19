package broker

import (
	"context"
	"github.com/engpetarmarinov/gotama/internal/task"
)

type ManagerInterface interface {
	GetTaskInterface
	GetAllTasksInterface
	UpdateTaskInterface
	GetDeleteTaskInterface
	EnqueueTaskInterface
	SchedulerInterface
}

type GetAllTasksInterface interface {
	GetAllTasks(ctx context.Context, offset int, limit int) (int64, []*task.Message, error)
}

type GetTaskInterface interface {
	GetTask(ctx context.Context, taskID string) (*task.Message, error)
}

type UpdateTaskInterface interface {
	UpdateTask(ctx context.Context, msg *task.Message) error
}

type GetUpdateTaskInterface interface {
	GetTaskInterface
	UpdateTaskInterface
}

type GetDeleteTaskInterface interface {
	GetTaskInterface
	RemoveTask(ctx context.Context, taskID string) error
}

type EnqueueTaskInterface interface {
	EnqueueTask(ctx context.Context, msg *task.Message) error
}

type SchedulerInterface interface {
	EnqueueScheduledTasks(ctx context.Context) error
}

type WorkerInterface interface {
	UpdateTaskInterface
	DequeueTask(ctx context.Context, qname string) (*task.Message, error)
	MarkTaskAsComplete(ctx context.Context, msg *task.Message) error
	RequeueTaskFailed(ctx context.Context, msg *task.Message) error
	RequeueTaskRetry(ctx context.Context, msg *task.Message) error
}
