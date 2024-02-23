package base

import (
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"strings"
	"time"
)

type TaskName int

const (
	TaskNameUnknown TaskName = iota
	TaskNameEmail
)

func (n TaskName) String() string {
	switch n {
	case TaskNameEmail:
		return "EMAIL"
	}
	panic("task name unknown")
}

func GetTaskName(name string) (TaskName, error) {
	switch strings.ToUpper(name) {
	case "EMAIL":
		return TaskNameEmail, nil
	}
	return TaskNameUnknown, errors.New("task name unknown")
}

type TaskRequest struct {
	Name    string          `json:"name"`
	Type    string          `json:"type"`
	Period  string          `json:"period"`
	Payload json.RawMessage `json:"payload"`
}

type TaskResponse struct {
	ID      string      `json:"ID"`
	Status  string      `json:"status"`
	Name    string      `json:"name"`
	Type    string      `json:"type"`
	Period  string      `json:"period"`
	Payload interface{} `json:"payload"`
	Result  interface{} `json:"result"`
	Error   string      `json:"error"`
}

type TaskStatus int

const (
	TaskStatusPending TaskStatus = iota + 1
	TaskStatusRunning
	TaskStatusSucceeded
	TaskStatusFailed
	TaskStatusRetry
)

func (s TaskStatus) String() string {
	switch s {
	case TaskStatusPending:
		return "PENDING"
	case TaskStatusRunning:
		return "RUNNING"
	case TaskStatusSucceeded:
		return "SUCCEEDED"
	case TaskStatusFailed:
		return "FAILED"
	case TaskStatusRetry:
		return "RETRY"
	}
	panic("task status unknown")
}

type TaskType int

const (
	TaskTypeOnce TaskType = iota + 1
	TaskTypeRecurring
)

func (t TaskType) String() string {
	switch t {
	case TaskTypeOnce:
		return "ONCE"
	case TaskTypeRecurring:
		return "RECURRING"
	}
	panic("task type unknown")
}

func GetTaskType(t string) (TaskType, error) {
	switch strings.ToUpper(t) {
	case "ONCE":
		return TaskTypeOnce, nil
	case "RECURRING":
		return TaskTypeRecurring, nil
	}
	return TaskTypeOnce, errors.New("task type unknown")
}

type TaskMessage struct {
	ID          string
	Name        string
	Status      TaskStatus
	Type        TaskType
	Period      time.Duration
	Payload     []byte
	Result      []byte
	CreatedAt   time.Time
	CompletedAt time.Time
	FailedAt    time.Time
	NumRetries  int
	MaxRetries  int
	Error       string
}

func NewTaskMessageFromRequest(req *TaskRequest) (*TaskMessage, error) {
	id := uuid.New()
	name, err := GetTaskName(req.Name)
	if err != nil {
		return nil, err
	}

	taskType, err := GetTaskType(req.Type)
	if err != nil {
		return nil, err
	}

	period, err := time.ParseDuration(req.Period)
	if err != nil {
		return nil, err
	}

	return &TaskMessage{
		ID:          id.String(),
		Name:        name.String(),
		Status:      TaskStatusPending,
		Type:        taskType,
		Period:      period,
		Payload:     req.Payload,
		Result:      nil,
		CreatedAt:   time.Now(),
		CompletedAt: time.Time{},
		FailedAt:    time.Time{},
		NumRetries:  0,
		MaxRetries:  0,
		Error:       "",
	}, nil
}

func NewTaskResponseFromMessage(msg *TaskMessage) (*TaskResponse, error) {
	var payload interface{}
	err := json.Unmarshal(msg.Payload, &payload)
	if err != nil {
		return nil, err
	}

	return &TaskResponse{
		ID:      msg.ID,
		Status:  msg.Status.String(),
		Name:    msg.Name,
		Type:    msg.Type.String(),
		Period:  msg.Period.String(),
		Payload: payload,
		Result:  msg.Result,
		Error:   msg.Error,
	}, nil
}
