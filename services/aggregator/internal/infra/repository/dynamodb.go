package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/rubenzito/case/aggregator/internal/domain"
)

type DynamoRepository struct {
	client       *dynamodb.Client
	eventsTable  string
	summaryTable string
}

func NewDynamoRepository(client *dynamodb.Client, eventsTable, summaryTable string) *DynamoRepository {
	return &DynamoRepository{
		client:       client,
		eventsTable:  eventsTable,
		summaryTable: summaryTable,
	}
}

// EventExists verifica se um event_id já foi processado (idempotência)
func (r *DynamoRepository) EventExists(ctx context.Context, eventID string) (bool, error) {
	out, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.eventsTable),
		Key: map[string]types.AttributeValue{
			"event_id": &types.AttributeValueMemberS{Value: eventID},
		},
	})
	if err != nil {
		return false, fmt.Errorf("erro ao verificar evento: %w", err)
	}
	return len(out.Item) > 0, nil
}

// SaveEvent salva um evento individual na tabela events
func (r *DynamoRepository) SaveEvent(ctx context.Context, event domain.ProcessedEvent) error {
	item, err := attributevalue.MarshalMap(event)
	if err != nil {
		return fmt.Errorf("erro ao serializar evento: %w", err)
	}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.eventsTable),
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("erro ao salvar evento: %w", err)
	}
	return nil
}

// UpdateSummary atualiza incrementalmente o resumo do desenvolvedor
func (r *DynamoRepository) UpdateSummary(ctx context.Context, event domain.ProcessedEvent) error {
	// Busca o resumo atual
	out, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.summaryTable),
		Key: map[string]types.AttributeValue{
			"developer_id": &types.AttributeValueMemberS{Value: event.DeveloperID},
		},
	})
	if err != nil {
		return fmt.Errorf("erro ao buscar resumo: %w", err)
	}

	// Reconstrói o resumo atual ou começa do zero
	var summary domain.DeveloperSummary
	if len(out.Item) > 0 {
		if err := attributevalue.UnmarshalMap(out.Item, &summary); err != nil {
			return fmt.Errorf("erro ao desserializar resumo: %w", err)
		}
	} else {
		summary = domain.DeveloperSummary{DeveloperID: event.DeveloperID}
	}

	// Atualiza os totais
	switch event.MetricType {
	case "commits":
		summary.TotalCommits += int64(event.Value)
	case "pull_requests":
		summary.TotalPullRequests += int64(event.Value)
	case "review_time_minutes":
		summary.TotalReviewTime += event.Value
		summary.ReviewTimeCount++
	}
	summary.EventsProcessed++
	if event.Timestamp.After(summary.LastActivity) {
		summary.LastActivity = event.Timestamp
	}

	// Salva de volta
	item, err := attributevalue.MarshalMap(summary)
	if err != nil {
		return fmt.Errorf("erro ao serializar resumo: %w", err)
	}
	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.summaryTable),
		Item:      item,
	})
	return err
}

// GetEvents retorna todos os eventos de um desenvolvedor
// Nota: scan com filtro — aceitável para o volume deste case
func (r *DynamoRepository) GetEvents(ctx context.Context, developerID string) ([]domain.ProcessedEvent, error) {
	out, err := r.client.Scan(ctx, &dynamodb.ScanInput{
		TableName:        aws.String(r.eventsTable),
		FilterExpression: aws.String("developer_id = :dev"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":dev": &types.AttributeValueMemberS{Value: developerID},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar eventos: %w", err)
	}

	var events []domain.ProcessedEvent
	if err := attributevalue.UnmarshalListOfMaps(out.Items, &events); err != nil {
		return nil, fmt.Errorf("erro ao desserializar eventos: %w", err)
	}
	return events, nil
}

// GetSummary retorna o resumo de um desenvolvedor
func (r *DynamoRepository) GetSummary(ctx context.Context, developerID string) (*domain.DeveloperSummary, error) {
	out, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.summaryTable),
		Key: map[string]types.AttributeValue{
			"developer_id": &types.AttributeValueMemberS{Value: developerID},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar resumo: %w", err)
	}
	if len(out.Item) == 0 {
		return nil, errors.New("developer não encontrado")
	}

	var summary domain.DeveloperSummary
	if err := attributevalue.UnmarshalMap(out.Item, &summary); err != nil {
		return nil, fmt.Errorf("erro ao desserializar resumo: %w", err)
	}
	return &summary, nil
}

// HealthCheck verifica se o DynamoDB está acessível
func (r *DynamoRepository) HealthCheck(ctx context.Context) error {
	_, err := r.client.ListTables(ctx, &dynamodb.ListTablesInput{})
	if err != nil {
		slog.Error("dynamodb health check falhou", "erro", err)
	}
	return err
}
