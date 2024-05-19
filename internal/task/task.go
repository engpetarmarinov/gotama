package task

import (
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"strings"
	"time"
)

type Name int

const (
	NameUnknown Name = iota
	NameEmail
	NameSMS
	NameSlack
	NameFoo
)

func (n Name) String() string {
	switch n {
	case NameEmail:
		return "EMAIL"
	case NameSMS:
		return "SMS"
	case NameSlack:
		return "SLACK"
	case NameFoo:
		return "FOO"
	}
	panic("task name unknown")
}

func GetName(name string) (Name, error) {
	switch strings.ToUpper(name) {
	case "EMAIL":
		return NameEmail, nil
	case "SMS":
		return NameSMS, nil
	case "SLACK":
		return NameSlack, nil
	case "FOO":
		return NameFoo, nil
	}
	return NameUnknown, errors.New("task name unknown")
}

const (
	QueueDefault string = "default"
)

// Request represents the payload for creating a new task.
// swagger:model taskRequest
type Request struct {
	// The name of the task
	// example: email
	Name string `json:"name"`

	// The type of the task (e.g., once, recurring)
	// example: once
	Type string `json:"type"`

	// The period of the task, applicable if the task is recurring (e.g., 45m, 5s)
	// example: 45m
	Period string `json:"period"`

	// The payload of the task containing task-specific data
	Payload json.RawMessage `json:"payload"`
}

// Response represents the response object for a task.
// swagger:model taskResponse
type Response struct {
	// The unique identifier of the task
	// example: 11ef259c-8523-42e4-8568-9d167dbba9da
	ID string `json:"ID"`

	// The current status of the task
	// example: PENDING
	Status string `json:"status"`

	// The name of the task
	// example: email
	Name string `json:"name"`

	// The type of the task (e.g., once, recurring)
	// example: once
	Type string `json:"type"`

	// The period of the task, applicable if the task is recurring
	// example: 45m
	Period string `json:"period"`

	// The payload of the task containing task-specific data
	Payload any `json:"payload"`

	// Error message, if any
	// example: null
	Error *string `json:"error,omitempty"`

	// The creation timestamp of the task
	// example: 2023-05-19T14:28:23Z
	CreatedAt string `json:"created_at"`

	// The completion timestamp of the task, if completed
	// example: 2023-05-19T15:00:00Z
	CompletedAt *string `json:"completed_at,omitempty"`

	// The failure timestamp of the task, if failed
	// example: 2023-05-19T14:45:00Z
	FailedAt *string `json:"failed_at,omitempty"`
}

type Status int

const (
	StatusPending Status = iota + 1
	StatusRunning
	StatusSucceeded
	StatusFailed
)

func (s Status) String() string {
	switch s {
	case StatusPending:
		return "PENDING"
	case StatusRunning:
		return "RUNNING"
	case StatusSucceeded:
		return "SUCCEEDED"
	case StatusFailed:
		return "FAILED"
	}
	panic("task status unknown")
}

type Type int

const (
	TypeOnce Type = iota + 1
	TypeRecurring
)

func (t Type) String() string {
	switch t {
	case TypeOnce:
		return "ONCE"
	case TypeRecurring:
		return "RECURRING"
	}
	panic("task type unknown")
}

func GetType(t string) (Type, error) {
	switch strings.ToUpper(t) {
	case "ONCE":
		return TypeOnce, nil
	case "RECURRING":
		return TypeRecurring, nil
	}
	return TypeOnce, errors.New("task type unknown")
}

type Message struct {
	ID          string
	Name        string
	Queue       string
	Status      Status
	Type        Type
	Period      time.Duration
	Payload     []byte
	CreatedAt   time.Time
	CompletedAt *time.Time
	FailedAt    *time.Time
	NumRetries  int
	Error       *string
}

func NewMessageFromRequest(req *Request) (*Message, error) {
	id := uuid.New()
	name, err := GetName(req.Name)
	if err != nil {
		return nil, err
	}

	taskType, err := GetType(req.Type)
	if err != nil {
		return nil, err
	}

	var period time.Duration
	if taskType == TypeRecurring {
		period, err = time.ParseDuration(req.Period)
		if err != nil {
			return nil, err
		}

		if period < time.Second {
			return nil, errors.New("period has to be at least 1s")
		}
	}

	return &Message{
		ID:          id.String(),
		Name:        name.String(),
		Queue:       QueueDefault,
		Status:      StatusPending,
		Type:        taskType,
		Period:      period,
		Payload:     req.Payload,
		CreatedAt:   time.Now(),
		CompletedAt: nil,
		FailedAt:    nil,
		NumRetries:  0,
		Error:       nil,
	}, nil
}

func NewResponseFromMessage(msg *Message) (*Response, error) {
	var payload any
	err := json.Unmarshal(msg.Payload, &payload)
	if err != nil {
		return nil, err
	}

	var completedAt *string
	if msg.CompletedAt != nil {
		date := msg.CompletedAt.Format(time.RFC3339)
		completedAt = &date
	}

	var failedAt *string
	if msg.FailedAt != nil {
		date := msg.FailedAt.Format(time.RFC3339)
		failedAt = &date
	}

	return &Response{
		ID:          msg.ID,
		Status:      msg.Status.String(),
		Name:        msg.Name,
		Type:        msg.Type.String(),
		Period:      msg.Period.String(),
		Payload:     payload,
		Error:       msg.Error,
		CreatedAt:   msg.CreatedAt.Format(time.RFC3339),
		CompletedAt: completedAt,
		FailedAt:    failedAt,
	}, nil
}

func EncodeMessage(msg *Message) ([]byte, error) {
	//TODO: use protobuf to save space
	return json.Marshal(msg)
}

func DecodeMessage(encoded string) (*Message, error) {
	var msg Message
	err := json.Unmarshal([]byte(encoded), &msg)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}
