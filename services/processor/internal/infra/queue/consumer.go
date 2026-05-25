package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/rubenzito/case/processor/internal/domain"
)

// Message representa uma mensagem recebida da fila com metadados para delete
type Message struct {
	Event         domain.RawEvent
	ReceiptHandle string
}

// Consumer consome mensagens da fila SQS
type Consumer struct {
	client   *sqs.Client
	queueURL string
}

func NewConsumer(client *sqs.Client, queueURL string) *Consumer {
	return &Consumer{client: client, queueURL: queueURL}
}

// Poll faz long polling e retorna mensagens recebidas
func (c *Consumer) Poll(ctx context.Context) ([]Message, error) {
	out, err := c.client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(c.queueURL),
		MaxNumberOfMessages: 10,
		WaitTimeSeconds:     5, // long polling
	})
	if err != nil {
		return nil, fmt.Errorf("erro ao receber mensagens: %w", err)
	}

	var messages []Message
	for _, msg := range out.Messages {
		var event domain.RawEvent
		if err := json.Unmarshal([]byte(*msg.Body), &event); err != nil {
			slog.Error("erro ao desserializar mensagem", "erro", err, "body", *msg.Body)
			// Não deleta — vai para DLQ após 3 tentativas
			continue
		}
		messages = append(messages, Message{
			Event:         event,
			ReceiptHandle: *msg.ReceiptHandle,
		})
	}
	return messages, nil
}

// Delete remove a mensagem da fila após processamento bem-sucedido
func (c *Consumer) Delete(ctx context.Context, receiptHandle string) error {
	_, err := c.client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(c.queueURL),
		ReceiptHandle: aws.String(receiptHandle),
	})
	return err
}

// ChangeVisibility aumenta o timeout de visibilidade para retry com backoff
func (c *Consumer) ChangeVisibility(ctx context.Context, msg types.Message, attempt int) {
	// backoff exponencial: 2^attempt segundos (máx 30s)
	delay := int32(math.Min(math.Pow(2, float64(attempt)), 30))
	_, err := c.client.ChangeMessageVisibility(ctx, &sqs.ChangeMessageVisibilityInput{
		QueueUrl:          aws.String(c.queueURL),
		ReceiptHandle:     msg.ReceiptHandle,
		VisibilityTimeout: delay,
	})
	if err != nil {
		slog.Error("erro ao alterar visibilidade", "erro", err)
	}
	time.Sleep(time.Duration(delay) * time.Second)
}
