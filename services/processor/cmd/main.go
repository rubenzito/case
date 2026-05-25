package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sqs"

	"github.com/rubenzito/case/processor/internal/infra/config"
	"github.com/rubenzito/case/processor/internal/infra/queue"
	"github.com/rubenzito/case/processor/internal/infra/worker"
	"github.com/rubenzito/case/processor/internal/usecase"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func main() {
	// Logger estruturado JSON
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	// Inicializa tracing (stdout exporter para demonstração)
	tp, err := initTracer()
	if err != nil {
		slog.Error("erro ao inicializar tracer", "erro", err)
		os.Exit(1)
	}
	defer func() { _ = tp.Shutdown(context.Background()) }()

	cfg := config.Load()

	// Configura cliente AWS apontando para o LocalStack
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

	consumer := queue.NewConsumer(sqsClient, cfg.RawEventsQueue)
	publisher := queue.NewPublisher(sqsClient, cfg.ProcessedEventsQueue)
	processUC := usecase.NewProcessEvent(publisher, cfg.ProcessorID)
	pool := worker.NewPool(consumer, processUC, cfg.WorkerCount)

	// Graceful shutdown: escuta SIGTERM e SIGINT
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, os.Interrupt)
	defer stop()

	slog.Info("processor iniciado", "processor_id", cfg.ProcessorID, "workers", cfg.WorkerCount)
	pool.Start(ctx)
	slog.Info("processor encerrado")
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
