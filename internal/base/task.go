package base

type TaskStatus string

const (
	TaskStatusPending   = "PENDING"
	TaskStatusRunning   = "RUNNING"
	TaskStatusSucceeded = "SUCCEEDED"
	TaskStatusFailed    = "FAILED"
)

type Task struct {
	Status TaskStatus `json:"status"`
}
