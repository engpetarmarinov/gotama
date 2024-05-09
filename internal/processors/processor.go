package processors

import (
	"context"
	"fmt"
	"github.com/engpetarmarinov/gotama/internal/config"
	"github.com/engpetarmarinov/gotama/internal/task"
)

type Processor interface {
	ProcessTask(context.Context, *task.Message) error
	ValidatePayload(payload []byte) error
}

func ProcessorFactory(config config.API, name task.Name) (Processor, error) {
	switch name {
	case task.NameEmail:
		return NewEmailProcessor(config), nil
	case task.NameSMS:
		return NewSMSProcessor(config), nil
	case task.NameFoo:
		return NewFooProcessor(), nil
	default:
		return nil, fmt.Errorf("unknown processor type for %s", name.String())
	}
}
