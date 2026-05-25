package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Tipos válidos de métrica
var validMetricTypes = map[string]bool{
	"commits":             true,
	"pull_requests":       true,
	"review_time_minutes": true,
}

// RawEvent representa um evento bruto que chega na fila raw-events
type RawEvent struct {
	EventID     string    `json:"event_id"`
	DeveloperID string    `json:"developer_id"`
	MetricType  string    `json:"metric_type"`
	Value       float64   `json:"value"`
	Repository  string    `json:"repository"`
	Timestamp   time.Time `json:"timestamp"`
}

// ProcessedEvent é o evento enriquecido que vai para processed-events
type ProcessedEvent struct {
	EventID     string    `json:"event_id"`
	DeveloperID string    `json:"developer_id"`
	MetricType  string    `json:"metric_type"`
	Value       float64   `json:"value"`
	Repository  string    `json:"repository"`
	Timestamp   time.Time `json:"timestamp"`
	ProcessedAt time.Time `json:"processed_at"`
	ProcessorID string    `json:"processor_id"`
}

// Validate valida as regras de negócio do evento bruto
func (e *RawEvent) Validate() error {
	// event_id: obrigatório e deve ser UUID válido
	if e.EventID == "" {
		return errors.New("event_id é obrigatório")
	}
	if _, err := uuid.Parse(e.EventID); err != nil {
		return errors.New("event_id deve ser um UUID válido")
	}

	// developer_id: obrigatório
	if e.DeveloperID == "" {
		return errors.New("developer_id é obrigatório")
	}

	// metric_type: deve ser um dos valores permitidos
	if !validMetricTypes[e.MetricType] {
		return errors.New("metric_type inválido: deve ser commits, pull_requests ou review_time_minutes")
	}

	// value: deve ser >= 0
	if e.Value < 0 {
		return errors.New("value deve ser maior ou igual a 0")
	}

	// review_time_minutes: máximo de 1440 (24h)
	if e.MetricType == "review_time_minutes" && e.Value > 1440 {
		return errors.New("review_time_minutes não pode ser maior que 1440 (24h)")
	}

	// timestamp: obrigatório e não pode ser futuro
	if e.Timestamp.IsZero() {
		return errors.New("timestamp é obrigatório")
	}
	if e.Timestamp.After(time.Now()) {
		return errors.New("timestamp não pode ser uma data futura")
	}

	return nil
}

// Enrich cria um ProcessedEvent a partir do RawEvent
func (e *RawEvent) Enrich(processorID string) ProcessedEvent {
	return ProcessedEvent{
		EventID:     e.EventID,
		DeveloperID: e.DeveloperID,
		MetricType:  e.MetricType,
		Value:       e.Value,
		Repository:  e.Repository,
		Timestamp:   e.Timestamp,
		ProcessedAt: time.Now().UTC(),
		ProcessorID: processorID,
	}
}
