package processors

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/engpetarmarinov/gotama/internal/config"
	"github.com/engpetarmarinov/gotama/internal/logger"
	"github.com/engpetarmarinov/gotama/internal/task"
	"github.com/slack-go/slack"
)

type SlackPayload struct {
	Channel string `json:"channel"`
	Text    string `json:"text"`
}

type SlackProcessor struct {
	config config.API
}

func NewSlackProcessor(config config.API) *SlackProcessor {
	return &SlackProcessor{
		config: config,
	}
}

func (sp *SlackProcessor) ProcessTask(ctx context.Context, msg *task.Message) error {
	var p SlackPayload
	if err := json.Unmarshal(msg.Payload, &p); err != nil {
		return fmt.Errorf("error unmarshalling slack payload %w", err)
	}

	token := sp.config.Get("SLACK_TOKEN")
	client := slack.New(token)
	logger.Info("Sending Slack", "channel", p.Channel, "text", p.Text)
	channel, timestamp, err := client.PostMessage(p.Channel, slack.MsgOptionText(p.Text, true))
	if err != nil {
		return fmt.Errorf("error sending slack message: %w", err)
	}

	logger.Info("Sent Slack", "channel", channel, "text", p.Text, "timestamp", timestamp)

	return nil
}

func (sp *SlackProcessor) ValidatePayload(payload []byte) error {
	var p SlackPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return err
	}

	if len(p.Channel) <= 0 || len(p.Text) <= 0 {
		return errors.New("invalid payload: phone and text are required fields")
	}

	return nil
}
