package usecase

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/rubenzito/case/aggregator/internal/domain"
)

// Repository define o contrato de persistência
type Repository interface {
	EventExists(ctx context.Context, eventID string) (bool, error)
	SaveEvent(ctx context.Context, event domain.ProcessedEvent) error
	UpdateSummary(ctx context.Context, event domain.ProcessedEvent) error
}

// AggregateEvent é o use case principal do aggregator
type AggregateEvent struct {
	repo Repository
}

func NewAggregateEvent(repo Repository) *AggregateEvent {
	return &AggregateEvent{repo: repo}
}

// Execute processa um evento: verifica idempotência, salva e agrega
func (uc *AggregateEvent) Execute(ctx context.Context, event domain.ProcessedEvent) error {
	log := slog.With("event_id", event.EventID, "developer_id", event.DeveloperID)

	// Idempotência: ignora eventos duplicados
	exists, err := uc.repo.EventExists(ctx, event.EventID)
	if err != nil {
		return fmt.Errorf("erro ao verificar idempotência: %w", err)
	}
	if exists {
		log.Info("evento duplicado ignorado")
		return nil
	}

	// Persiste o evento individual
	if err := uc.repo.SaveEvent(ctx, event); err != nil {
		return fmt.Errorf("erro ao salvar evento: %w", err)
	}

	// Atualiza o resumo do desenvolvedor
	if err := uc.repo.UpdateSummary(ctx, event); err != nil {
		return fmt.Errorf("erro ao atualizar resumo: %w", err)
	}

	log.Info("evento agregado com sucesso", "metric_type", event.MetricType, "value", event.Value)
	return nil
}
