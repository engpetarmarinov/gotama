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

type FooPayload struct {
	Bar string `json:"bar"`
	Baz string `json:"baz"`
}

type FooProcessor struct {
}

func (ep *FooProcessor) ProcessTask(ctx context.Context, t *task.Message) error {
	var payload FooPayload
	if err := json.Unmarshal(t.Payload, &payload); err != nil {
		logger.Error("error unmarshalling foo payload", "err", t.Payload)
		return fmt.Errorf("error unmarshalling foo payload %s", err.Error())
	}
	logger.Info("Doing foo", "bar", payload.Bar, "baz", payload.Baz)
	//simulate dummy load
	tick := time.Tick(time.Second * 1)
	for {
		select {
		case <-ctx.Done():
			return errors.New("foo task interrupted by done")
		case <-tick:
			return errors.New("foo error")
		default:
			_ = rand.Int() * rand.Int()
		}
	}
}

func (ep *FooProcessor) ValidatePayload(payload []byte) error {
	var p FooPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return err
	}

	if len(p.Bar) <= 0 || len(p.Baz) <= 0 {
		return errors.New("invalid payload: bar and baz are required fields")
	}

	return nil
}

func NewFooProcessor() *FooProcessor {
	return &FooProcessor{}
}
