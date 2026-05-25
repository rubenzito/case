package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/google/uuid"
	"github.com/rubenzito/case/processor/internal/domain"
	"github.com/rubenzito/case/processor/internal/infra/queue"
)

func ensureLocalStack(t *testing.T, endpoint string) {
	t.Helper()
	host := "localhost"
	port := "4566"
	if endpoint != "" {
		var h string
		_, err := fmt.Sscanf(endpoint, "http://%s", &h)
		if err == nil {
			hostPort := h
			for i := 0; i < len(hostPort); i++ {
				if hostPort[i] == '/' {
					hostPort = hostPort[:i]
					break
				}
			}
			host = hostPort
			var hp string
			_, err := fmt.Sscanf(hostPort, "%s:%s", &host, &hp)
			if err == nil {
				port = hp
			}
		}
	}
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, port), 500*time.Millisecond)
	if err != nil {
		t.Skipf("LocalStack not reachable at %s, skipping integration test: %v", endpoint, err)
	}
	conn.Close()
}

func awsConfigFor(endpoint string) (aws.Config, error) {
	return config.LoadDefaultConfig(context.Background(),
		config.WithRegion("us-east-1"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("test", "test", "")),
		config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{URL: endpoint}, nil
		})),
	)
}

func TestWorkerPool_ProcessesConcurrently(t *testing.T) {
	endpoint := os.Getenv("AWS_ENDPOINT")
	if endpoint == "" {
		endpoint = "http://localhost:4566"
	}
	ensureLocalStack(t, endpoint)

	cfg, err := awsConfigFor(endpoint)
	if err != nil {
		t.Fatalf("failed to build aws cfg: %v", err)
	}
	client := sqs.NewFromConfig(cfg)

	// create queue
	qName := fmt.Sprintf("test-pool-%d", time.Now().UnixNano())
	qOut, err := client.CreateQueue(context.Background(), &sqs.CreateQueueInput{QueueName: &qName})
	if err != nil {
		t.Fatalf("create queue failed: %v", err)
	}

	// send messages
	total := 6
	for i := 0; i < total; i++ {
		raw := domain.RawEvent{
			EventID:     uuid.NewString(),
			DeveloperID: "dev-conc",
			MetricType:  "commits",
			Value:       float64(i + 1),
			Repository:  "repo/concurrency",
			Timestamp:   time.Now().Add(-time.Minute),
		}
		b, _ := json.Marshal(raw)
		_, err = client.SendMessage(context.Background(), &sqs.SendMessageInput{QueueUrl: qOut.QueueUrl, MessageBody: aws.String(string(b))})
		if err != nil {
			t.Fatalf("send message failed: %v", err)
		}
	}

	consumer := queue.NewConsumer(client, *qOut.QueueUrl)

	// mock processor that sleeps and tracks concurrency
	var mu sync.Mutex
	running := 0
	maxRunning := 0
	processed := 0
	proc := &mockSleepProcessor{sleep: 500 * time.Millisecond, before: func() {
		mu.Lock()
		running++
		if running > maxRunning {
			maxRunning = running
		}
		mu.Unlock()
	}, after: func() {
		mu.Lock()
		running--
		processed++
		mu.Unlock()
	}}

	pool := NewPool(consumer, proc, 3)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	start := time.Now()
	go pool.Start(ctx)

	// wait until all processed or timeout
	deadline := time.Now().Add(25 * time.Second)
	for time.Now().Before(deadline) {
		mu.Lock()
		if processed >= total {
			mu.Unlock()
			break
		}
		mu.Unlock()
		time.Sleep(200 * time.Millisecond)
	}
	elapsed := time.Since(start)

	if processed < total {
		t.Fatalf("not all messages processed: %d/%d", processed, total)
	}

	// expected parallelism: with 3 workers and 6 messages of 0.5s, ideal ~1s
	if elapsed > 3*time.Second {
		t.Fatalf("expected processing to be concurrent and fast; elapsed %v", elapsed)
	}
	if maxRunning < 2 {
		t.Fatalf("expected at least 2 concurrent workers, got %d", maxRunning)
	}
}

// mockSleepProcessor implements Processor
type mockSleepProcessor struct {
	sleep  time.Duration
	before func()
	after  func()
}

func (m *mockSleepProcessor) Execute(ctx context.Context, event domain.RawEvent) error {
	if m.before != nil {
		m.before()
	}
	time.Sleep(m.sleep)
	if m.after != nil {
		m.after()
	}
	return nil
}
