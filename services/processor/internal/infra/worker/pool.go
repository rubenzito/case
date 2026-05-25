package worker

import (
	"context"
	"log/slog"
	"sync"

	"github.com/rubenzito/case/processor/internal/domain"
	"github.com/rubenzito/case/processor/internal/infra/queue"
	"go.opentelemetry.io/otel"
)

// Processor é a interface mínima que um usecase precisa expor para o pool
type Processor interface {
	Execute(ctx context.Context, event domain.RawEvent) error
}

// Pool gerencia um conjunto de workers que processam mensagens da fila
type Pool struct {
	consumer    *queue.Consumer
	processUC   Processor
	workerCount int
}

func NewPool(consumer *queue.Consumer, processUC Processor, workerCount int) *Pool {
	return &Pool{
		consumer:    consumer,
		processUC:   processUC,
		workerCount: workerCount,
	}
}

// Start inicia o pool de workers e bloqueia até o contexto ser cancelado
func (p *Pool) Start(ctx context.Context) {
	jobs := make(chan queue.Message, p.workerCount*2)
	var wg sync.WaitGroup

	// Inicia os workers
	for i := 0; i < p.workerCount; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			p.work(ctx, workerID, jobs)
		}(i)
	}

	slog.Info("worker pool iniciado", "workers", p.workerCount)

	// Loop de polling — alimenta o canal de jobs
	for {
		select {
		case <-ctx.Done():
			slog.Info("contexto cancelado, encerrando polling...")
			close(jobs)
			wg.Wait()
			slog.Info("todos os workers finalizados")
			return
		default:
			msgs, err := p.consumer.Poll(ctx)
			if err != nil {
				if ctx.Err() != nil {
					close(jobs)
					wg.Wait()
					return
				}
				slog.Error("erro no polling", "erro", err)
				continue
			}
			for _, msg := range msgs {
				jobs <- msg
			}
		}
	}
}

// work é a função executada por cada worker
func (p *Pool) work(ctx context.Context, workerID int, jobs <-chan queue.Message) {
	for msg := range jobs {
		log := slog.With("worker_id", workerID, "event_id", msg.Event.EventID)
		// Start a span for processing this message
		tracer := otel.Tracer("processor")
		ctxSpan, span := tracer.Start(ctx, "ProcessEvent")

		err := p.processUC.Execute(ctxSpan, msg.Event)
		span.End()
		if err != nil {
			// Não deleta — SQS vai retentar e depois manda para DLQ
			log.Warn("processamento falhou, mensagem não deletada", "erro", err)
			continue
		}

		// Deletar da fila só após sucesso
		if err := p.consumer.Delete(ctx, msg.ReceiptHandle); err != nil {
			log.Error("erro ao deletar mensagem da fila", "erro", err)
		}
	}
}
