package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rubenzito/case/processor/internal/domain"
)

type mockPublisher struct {
	called    bool
	published any
	retErr    error
}

func (m *mockPublisher) Publish(ctx context.Context, payload any) error {
	m.called = true
	m.published = payload
	return m.retErr
}

func TestProcessEvent_Execute_Success(t *testing.T) {
	ctx := context.Background()
	raw := domain.RawEvent{
		EventID:     uuid.NewString(),
		DeveloperID: "dev-1",
		MetricType:  "commits",
		Value:       3,
		Repository:  "repo",
		Timestamp:   time.Now().Add(-time.Minute),
	}

	pub := &mockPublisher{}
	uc := NewProcessEvent(pub, "proc-1")

	if err := uc.Execute(ctx, raw); err != nil {
		t.Fatalf("Execute returned unexpected error: %v", err)
	}

	if !pub.called {
		t.Fatalf("expected publisher to be called")
	}

	pe, ok := pub.published.(domain.ProcessedEvent)
	if !ok {
		t.Fatalf("published payload has wrong type: %T", pub.published)
	}
	if pe.ProcessorID != "proc-1" {
		t.Fatalf("unexpected ProcessorID: got %s", pe.ProcessorID)
	}
	if time.Since(pe.ProcessedAt) > time.Minute {
		t.Fatalf("ProcessedAt seems far in the past: %v", pe.ProcessedAt)
	}
}

func TestProcessEvent_Execute_PublishError(t *testing.T) {
	ctx := context.Background()
	raw := domain.RawEvent{
		EventID:     uuid.NewString(),
		DeveloperID: "dev-1",
		MetricType:  "commits",
		Value:       1,
		Repository:  "repo",
		Timestamp:   time.Now().Add(-time.Minute),
	}

	pub := &mockPublisher{retErr: errors.New("send failed")}
	uc := NewProcessEvent(pub, "proc-1")

	err := uc.Execute(ctx, raw)
	if err == nil {
		t.Fatal("expected error when publisher fails")
	}
}

func TestProcessEvent_Execute_InvalidEvent(t *testing.T) {
	ctx := context.Background()
	raw := domain.RawEvent{
		EventID:     "not-a-uuid",
		DeveloperID: "",
		MetricType:  "unknown",
		Value:       -1,
		Timestamp:   time.Time{},
	}

	pub := &mockPublisher{}
	uc := NewProcessEvent(pub, "proc-1")

	if err := uc.Execute(ctx, raw); err == nil {
		t.Fatal("expected validation error for invalid raw event")
	}
}
