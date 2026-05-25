package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rubenzito/case/aggregator/internal/domain"
)

type mockRepo struct {
	exists    bool
	existsErr error

	saveErr   error
	updateErr error

	saved   bool
	updated bool
}

func (m *mockRepo) EventExists(ctx context.Context, eventID string) (bool, error) {
	return m.exists, m.existsErr
}

func (m *mockRepo) SaveEvent(ctx context.Context, event domain.ProcessedEvent) error {
	m.saved = true
	return m.saveErr
}

func (m *mockRepo) UpdateSummary(ctx context.Context, event domain.ProcessedEvent) error {
	m.updated = true
	return m.updateErr
}

func makeEvent() domain.ProcessedEvent {
	return domain.ProcessedEvent{
		EventID:     uuid.NewString(),
		DeveloperID: "dev-1",
		MetricType:  "commits",
		Value:       2,
		Repository:  "repo",
		Timestamp:   time.Now().Add(-time.Minute),
		ProcessedAt: time.Now().UTC(),
		ProcessorID: "proc-1",
	}
}

func TestAggregateEvent_Execute_DuplicateIgnored(t *testing.T) {
	ctx := context.Background()
	repo := &mockRepo{exists: true}
	uc := NewAggregateEvent(repo)

	if err := uc.Execute(ctx, makeEvent()); err != nil {
		t.Fatalf("expected no error for duplicate event: %v", err)
	}
	if repo.saved || repo.updated {
		t.Fatalf("expected no persistence calls for duplicate event")
	}
}

func TestAggregateEvent_Execute_Success(t *testing.T) {
	ctx := context.Background()
	repo := &mockRepo{exists: false}
	uc := NewAggregateEvent(repo)

	if err := uc.Execute(ctx, makeEvent()); err != nil {
		t.Fatalf("expected success but got error: %v", err)
	}
	if !repo.saved || !repo.updated {
		t.Fatalf("expected SaveEvent and UpdateSummary to be called")
	}
}

func TestAggregateEvent_Execute_SaveError(t *testing.T) {
	ctx := context.Background()
	repo := &mockRepo{exists: false, saveErr: errors.New("db write failed")}
	uc := NewAggregateEvent(repo)

	if err := uc.Execute(ctx, makeEvent()); err == nil {
		t.Fatal("expected error when SaveEvent fails")
	}
}
