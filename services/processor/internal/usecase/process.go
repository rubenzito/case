package usecase

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/rubenzito/case/processor/internal/domain"
)

// Publisher define o contrato para publicar eventos
type Publisher interface {
	Publish(ctx context.Context, payload any) error
}

// ProcessEvent é o use case principal: valida e enriquece um evento bruto
type ProcessEvent struct {
	publisher   Publisher
	processorID string
}

func NewProcessEvent(publisher Publisher, processorID string) *ProcessEvent {
	return &ProcessEvent{
		publisher:   publisher,
		processorID: processorID,
	}
}

// Execute valida o evento e, se válido, publica o evento enriquecido
// Retorna erro se o evento for inválido ou se falhar ao publicar
func (uc *ProcessEvent) Execute(ctx context.Context, event domain.RawEvent) error {
	log := slog.With("event_id", event.EventID, "developer_id", event.DeveloperID)

	// Validação
	if err := event.Validate(); err != nil {
		log.Warn("evento inválido rejeitado", "motivo", err.Error())
		return fmt.Errorf("validação falhou: %w", err)
	}

	// Enriquecimento
	processed := event.Enrich(uc.processorID)

	// Publicação
	if err := uc.publisher.Publish(ctx, processed); err != nil {
		log.Error("falha ao publicar evento processado", "erro", err)
		return fmt.Errorf("falha ao publicar: %w", err)
	}

	log.Info("evento processado e publicado com sucesso", "metric_type", event.MetricType)
	return nil
}
