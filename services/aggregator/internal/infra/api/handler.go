package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rubenzito/case/aggregator/internal/domain"
)

// QueryRepository define o contrato de leitura para os handlers
type QueryRepository interface {
	GetEvents(ctx context.Context, developerID string) ([]domain.ProcessedEvent, error)
	GetSummary(ctx context.Context, developerID string) (*domain.DeveloperSummary, error)
	HealthCheck(ctx context.Context) error
}

// SQSChecker define o contrato para o health check da fila
type SQSChecker interface {
	HealthCheck(ctx context.Context) error
}

type Handler struct {
	repo       QueryRepository
	sqsChecker SQSChecker
}

func NewHandler(repo QueryRepository, sqsChecker SQSChecker) *Handler {
	return &Handler{repo: repo, sqsChecker: sqsChecker}
}

func (h *Handler) Router() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", h.health)
	r.Get("/metrics/{developer_id}", h.getEvents)
	r.Get("/metrics/{developer_id}/summary", h.getSummary)

	return r
}

func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	status := map[string]string{"status": "ok"}

	if err := h.repo.HealthCheck(ctx); err != nil {
		status["dynamodb"] = "unhealthy: " + err.Error()
		status["status"] = "degraded"
	} else {
		status["dynamodb"] = "ok"
	}

	if err := h.sqsChecker.HealthCheck(ctx); err != nil {
		status["sqs"] = "unhealthy: " + err.Error()
		status["status"] = "degraded"
	} else {
		status["sqs"] = "ok"
	}

	code := http.StatusOK
	if status["status"] == "degraded" {
		code = http.StatusServiceUnavailable
	}
	writeJSON(w, code, status)
}

func (h *Handler) getEvents(w http.ResponseWriter, r *http.Request) {
	developerID := chi.URLParam(r, "developer_id")
	events, err := h.repo.GetEvents(r.Context(), developerID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if events == nil {
		events = []domain.ProcessedEvent{}
	}
	writeJSON(w, http.StatusOK, events)
}

func (h *Handler) getSummary(w http.ResponseWriter, r *http.Request) {
	developerID := chi.URLParam(r, "developer_id")
	summary, err := h.repo.GetSummary(r.Context(), developerID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "developer não encontrado"})
		return
	}
	writeJSON(w, http.StatusOK, summary.ToSummaryResponse())
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}
