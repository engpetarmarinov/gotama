package processors

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/engpetarmarinov/gotama/internal/config"
	"github.com/engpetarmarinov/gotama/internal/logger"
	"github.com/engpetarmarinov/gotama/internal/task"
	"regexp"
)

var e164Regex = regexp.MustCompile(`^\+[1-9]\d{1,14}$`)

type SMSPayload struct {
	Phone string `json:"phone"`
	Text  string `json:"text"`
}

type SMSProcessor struct {
	config config.API
}

func NewSMSProcessor(config config.API) *SMSProcessor {
	return &SMSProcessor{
		config: config,
	}
}

func (sp *SMSProcessor) ProcessTask(ctx context.Context, msg *task.Message) error {
	var p SMSPayload
	if err := json.Unmarshal(msg.Payload, &p); err != nil {
		return fmt.Errorf("error unmarshalling sms payload %w", err)
	}

	region := sp.config.Get("AWS_REGION")
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
	if err != nil {
		return err
	}

	client := sns.NewFromConfig(awsCfg)
	input := &sns.PublishInput{
		PhoneNumber: aws.String(p.Phone),
		Message:     aws.String(p.Text),
	}

	logger.Info("Sending SMS", "phone", p.Phone, "text", p.Text)

	// Send direct SMS without SNS topic
	output, err := client.Publish(ctx, input)
	if err != nil {
		return fmt.Errorf("error publishing to sns %w", err)
	}

	logger.Info("Sent SMS", "phone", p.Phone, "text", p.Text, "id", *output.MessageId)

	return nil
}

func (sp *SMSProcessor) ValidatePayload(payload []byte) error {
	var p SMSPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return err
	}

	if len(p.Phone) <= 0 || len(p.Text) <= 0 {
		return errors.New("invalid payload: phone and text are required fields")
	}

	if !e164Regex.MatchString(p.Phone) {
		return errors.New("invalid payload: phone must contain e164 phone number, e.g. +359888123456")
	}

	return nil
}
