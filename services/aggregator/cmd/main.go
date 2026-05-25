package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/sqs"

	"github.com/rubenzito/case/aggregator/internal/infra/api"
	"github.com/rubenzito/case/aggregator/internal/infra/config"
	"github.com/rubenzito/case/aggregator/internal/infra/queue"
	"github.com/rubenzito/case/aggregator/internal/infra/repository"
	"github.com/rubenzito/case/aggregator/internal/usecase"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	// Inicializa tracing (stdout exporter para demonstração)
	tp, err := initTracer()
	if err != nil {
		slog.Error("erro ao inicializar tracer", "erro", err)
		os.Exit(1)
	}
	defer func() { _ = tp.Shutdown(context.Background()) }()

	cfg := config.Load()

	// Configura clientes AWS
	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithRegion(cfg.AWSRegion),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("test", "test", "")),
		awsconfig.WithEndpointResolverWithOptions(
			aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{URL: cfg.AWSEndpoint}, nil
			}),
		),
	)
	if err != nil {
		slog.Error("erro ao configurar AWS", "erro", err)
		os.Exit(1)
	}

	sqsClient := sqs.NewFromConfig(awsCfg)
	dynamoClient := dynamodb.NewFromConfig(awsCfg)

	repo := repository.NewDynamoRepository(dynamoClient, cfg.EventsTable, cfg.SummaryTable)
	consumer := queue.NewConsumer(sqsClient, cfg.ProcessedEventsQueue)
	aggregateUC := usecase.NewAggregateEvent(repo)
	handler := api.NewHandler(repo, consumer)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, os.Interrupt)
	defer stop()

	var wg sync.WaitGroup

	// Goroutine do servidor HTTP
	server := &http.Server{
		Addr:    ":" + cfg.APIPort,
		Handler: handler.Router(),
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		slog.Info("API iniciada", "porta", cfg.APIPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("erro no servidor HTTP", "erro", err)
		}
	}()

	// Goroutine do consumer loop
	wg.Add(1)
	go func() {
		defer wg.Done()
		slog.Info("aggregator consumer iniciado", "workers", cfg.WorkerCount)
		jobs := make(chan queue.Message, cfg.WorkerCount*2)

		// Workers de agregação
		var workerWg sync.WaitGroup
		for i := 0; i < cfg.WorkerCount; i++ {
			workerWg.Add(1)
			go func(id int) {
				defer workerWg.Done()
				for msg := range jobs {
					// Extract context from message attributes for distributed tracing
					carrier := propagation.MapCarrier(msg.Attributes)
					ctxMsg := otel.GetTextMapPropagator().Extract(ctx, carrier)

					// Start a span for processing this message
					tracer := otel.Tracer("aggregator")
					ctxSpan, span := tracer.Start(ctxMsg, "AggregateEvent")
					if err := aggregateUC.Execute(ctxSpan, msg.Event); err != nil {
						span.RecordError(err)
						slog.Error("erro ao agregar evento", "worker", id, "event_id", msg.Event.EventID, "erro", err)
						span.End()
						continue
					}
					span.End()

					if err := consumer.Delete(ctx, msg.ReceiptHandle); err != nil {
						slog.Error("erro ao deletar mensagem", "worker", id, "erro", err)
					}
				}
			}(i)
		}

		// Polling loop
		for {
			select {
			case <-ctx.Done():
				close(jobs)
				workerWg.Wait()
				return
			default:
				msgs, err := consumer.Poll(ctx)
				if err != nil {
					if ctx.Err() != nil {
						close(jobs)
						workerWg.Wait()
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
	}()

	// Aguarda sinal de shutdown
	<-ctx.Done()
	slog.Info("sinal de shutdown recebido")
	server.Shutdown(context.Background())
	wg.Wait()
	slog.Info("aggregator encerrado")
}

// initTracer configura um TracerProvider com stdout exporter (apenas para demonstração local)
func initTracer() (*sdktrace.TracerProvider, error) {
	exp, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		return nil, err
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})
	return tp, nil
}
