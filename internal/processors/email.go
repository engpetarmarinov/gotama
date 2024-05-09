package processors

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
	"github.com/engpetarmarinov/gotama/internal/config"
	"github.com/engpetarmarinov/gotama/internal/logger"
	"github.com/engpetarmarinov/gotama/internal/task"
	"net/mail"
)

const charSet = "UTF-8"

type EmailPayload struct {
	To    string `json:"to"`
	Title string `json:"title"`
	Body  string `json:"body"`
}

func NewEmailProcessor(config config.API) *EmailProcessor {
	return &EmailProcessor{
		config: config,
	}
}

type EmailProcessor struct {
	config config.API
}

func (ep *EmailProcessor) ProcessTask(ctx context.Context, msg *task.Message) error {
	var payload EmailPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return fmt.Errorf("error unmarshalling email payload %w", err)
	}

	region := ep.config.Get("AWS_REGION")
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
	if err != nil {
		return err
	}

	client := ses.NewFromConfig(awsCfg)
	from := ep.config.Get("EMAIL_FROM")
	input := &ses.SendEmailInput{
		Destination: &types.Destination{
			ToAddresses: []string{payload.To},
		},
		Message: &types.Message{
			Subject: &types.Content{
				Data:    aws.String(payload.Title),
				Charset: aws.String(charSet),
			},
			Body: &types.Body{
				Text: &types.Content{
					Data:    aws.String(payload.Body),
					Charset: aws.String(charSet),
				},
			},
		},
		Source: aws.String(from),
	}

	logger.Info("Sending an email", "to", payload.To, "title", payload.Title, "body", payload.Body)

	output, err := client.SendEmail(ctx, input)
	if err != nil {
		return fmt.Errorf("error sending an email %w", err)
	}

	logger.Info("Sending an email", "to", payload.To, "title", payload.Title, "body", payload.Body, "id", *output.MessageId)

	return nil
}

func (ep *EmailProcessor) ValidatePayload(payload []byte) error {
	var p EmailPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return err
	}

	if len(p.To) <= 0 || len(p.Title) <= 0 || len(p.Body) <= 0 {
		return errors.New("invalid payload: to, title and body are required fields")
	}

	_, err := mail.ParseAddress(p.To)
	if err != nil {
		return errors.New("invalid payload: to must be a valid email address")
	}

	return nil
}
