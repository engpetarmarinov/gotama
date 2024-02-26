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
	NameFoo
)

func (n Name) String() string {
	switch n {
	case NameEmail:
		return "EMAIL"
	case NameFoo:
		return "FOO"
	}
	panic("task name unknown")
}

func GetName(name string) (Name, error) {
	switch strings.ToUpper(name) {
	case "EMAIL":
		return NameEmail, nil
	case "FOO":
		return NameFoo, nil
	}
	return NameUnknown, errors.New("task name unknown")
}

const (
	QueueDefault string = "default"
)

type Request struct {
	Name    string          `json:"name"`
	Type    string          `json:"type"`
	Period  string          `json:"period"`
	Payload json.RawMessage `json:"payload"`
}

type Response struct {
	ID      string      `json:"ID"`
	Status  string      `json:"status"`
	Name    string      `json:"name"`
	Type    string      `json:"type"`
	Period  string      `json:"period"`
	Payload interface{} `json:"payload"`
	Error   string      `json:"error"`
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
	CompletedAt time.Time
	FailedAt    time.Time
	NumRetries  int
	Error       string
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
		CompletedAt: time.Time{},
		FailedAt:    time.Time{},
		NumRetries:  0,
		Error:       "",
	}, nil
}

func NewResponseFromMessage(msg *Message) (*Response, error) {
	var payload interface{}
	err := json.Unmarshal(msg.Payload, &payload)
	if err != nil {
		return nil, err
	}

	return &Response{
		ID:      msg.ID,
		Status:  msg.Status.String(),
		Name:    msg.Name,
		Type:    msg.Type.String(),
		Period:  msg.Period.String(),
		Payload: payload,
		Error:   msg.Error,
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
