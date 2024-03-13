package processors

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/engpetarmarinov/gotama/internal/logger"
	"github.com/engpetarmarinov/gotama/internal/task"
	"math/rand"
	"time"
)

type EmailPayload struct {
	To    string `json:"to"`
	Title string `json:"title"`
	Body  string `json:"body"`
}

type EmailProcessor struct {
}

func (ep *EmailProcessor) ProcessTask(ctx context.Context, t *task.Message) error {
	var payload EmailPayload
	if err := json.Unmarshal(t.Payload, &payload); err != nil {
		logger.Error("error unmarshalling email payload", "err", t.Payload)
		return fmt.Errorf("error unmarshalling email payload %s", err.Error())
	}
	logger.Info("Sending an email", "to", payload.To, "title", payload.Title, "body", payload.Body)
	//simulate dummy load
	tick := time.Tick(time.Second * 2)
	for {
		select {
		case <-ctx.Done():
			return errors.New("email task interrupted by done")
		case <-tick:
			logger.Info("Email sent", "to", payload.To, "title", payload.Title, "body", payload.Body)
			return nil
		default:
			_ = rand.Int() * rand.Int()
		}
	}
}

func (ep *EmailProcessor) ValidatePayload(payload []byte) error {
	var p EmailPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return err
	}

	if len(p.To) <= 0 || len(p.Title) <= 0 || len(p.Body) <= 0 {
		return errors.New("invalid payload: to, title and body are required fields")
	}

	return nil
}

func NewEmailProcessor() *EmailProcessor {
	return &EmailProcessor{}
}
