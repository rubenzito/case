package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/rubenzito/case/aggregator/internal/domain"
)

type Message struct {
	Event         domain.ProcessedEvent
	ReceiptHandle string
	Attributes    map[string]string
}

type Consumer struct {
	client   *sqs.Client
	queueURL string
}

func NewConsumer(client *sqs.Client, queueURL string) *Consumer {
	return &Consumer{client: client, queueURL: queueURL}
}

func (c *Consumer) Poll(ctx context.Context) ([]Message, error) {
	out, err := c.client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:              aws.String(c.queueURL),
		MaxNumberOfMessages:   10,
		WaitTimeSeconds:       5,
		MessageAttributeNames: []string{"All"},
	})
	if err != nil {
		return nil, fmt.Errorf("erro ao receber mensagens: %w", err)
	}

	var messages []Message
	for _, msg := range out.Messages {
		var event domain.ProcessedEvent
		if err := json.Unmarshal([]byte(*msg.Body), &event); err != nil {
			slog.Error("erro ao desserializar mensagem", "erro", err)
			continue
		}

		attrs := map[string]string{}
		for k, v := range msg.MessageAttributes {
			if v.StringValue != nil {
				attrs[k] = *v.StringValue
			}
		}

		messages = append(messages, Message{
			Event:         event,
			ReceiptHandle: *msg.ReceiptHandle,
			Attributes:    attrs,
		})
	}
	return messages, nil
}

func (c *Consumer) Delete(ctx context.Context, receiptHandle string) error {
	_, err := c.client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(c.queueURL),
		ReceiptHandle: aws.String(receiptHandle),
	})
	return err
}

// HealthCheck verifica se consegue acessar a fila
func (c *Consumer) HealthCheck(ctx context.Context) error {
	_, err := c.client.GetQueueAttributes(ctx, &sqs.GetQueueAttributesInput{
		QueueUrl: aws.String(c.queueURL),
	})
	return err
}
