package processors

import (
	"context"
	"fmt"
	"github.com/engpetarmarinov/gotama/internal/task"
)

type Processor interface {
	ProcessTask(context.Context, *task.Message) error
	ValidatePayload(payload []byte) error
}

func ProcessorFactory(name task.Name) (Processor, error) {
	switch name {
	case task.NameEmail:
		return NewEmailProcessor(), nil
	case task.NameFoo:
		return NewFooProcessor(), nil
	default:
		return nil, fmt.Errorf("unknown processor type for %s", name.String())
	}
}
