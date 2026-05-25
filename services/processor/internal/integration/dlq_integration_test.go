package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/google/uuid"
	"github.com/rubenzito/case/processor/internal/domain"
	"github.com/rubenzito/case/processor/internal/infra/queue"
	"github.com/rubenzito/case/processor/internal/infra/worker"
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

func TestDLQ_MessagesMovedAfterMaxReceive(t *testing.T) {
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

	// create DLQ
	dlqName := fmt.Sprintf("test-dlq-%d", time.Now().UnixNano())
	dlqOut, err := client.CreateQueue(context.Background(), &sqs.CreateQueueInput{QueueName: &dlqName})
	if err != nil {
		t.Fatalf("create dlq failed: %v", err)
	}
	// get DLQ ARN
	attrs, err := client.GetQueueAttributes(context.Background(), &sqs.GetQueueAttributesInput{QueueUrl: dlqOut.QueueUrl, AttributeNames: []types.QueueAttributeName{types.QueueAttributeNameQueueArn}})
	if err != nil {
		t.Fatalf("get dlq attrs failed: %v", err)
	}
	dlqArn := attrs.Attributes["QueueArn"]

	// create main queue with redrive policy
	qName := fmt.Sprintf("test-raw-%d", time.Now().UnixNano())
	redrive := fmt.Sprintf(`{"deadLetterTargetArn":"%s","maxReceiveCount":"2"}`, dlqArn)
	qOut, err := client.CreateQueue(context.Background(), &sqs.CreateQueueInput{QueueName: &qName, Attributes: map[string]string{"RedrivePolicy": redrive}})
	if err != nil {
		t.Fatalf("create queue failed: %v", err)
	}

	// send one message that will fail processing
	raw := domain.RawEvent{
		EventID:     uuid.NewString(),
		DeveloperID: "dev-test",
		MetricType:  "commits",
		Value:       1,
		Repository:  "repo/test",
		Timestamp:   time.Now().Add(-time.Minute),
	}
	b, _ := json.Marshal(raw)
	_, err = client.SendMessage(context.Background(), &sqs.SendMessageInput{QueueUrl: qOut.QueueUrl, MessageBody: aws.String(string(b))})
	if err != nil {
		t.Fatalf("send message failed: %v", err)
	}

	// prepare consumer and a processor that always fails
	consumer := queue.NewConsumer(client, *qOut.QueueUrl)
	// mock processor that always returns error
	mockProc := &failingProcessor{}
	pool := worker.NewPool(consumer, mockProc, 1)

	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()
	// start pool
	go pool.Start(ctx)

	// wait until message appears in DLQ
	dlqURL := *dlqOut.QueueUrl
	found := false
	deadline := time.Now().Add(35 * time.Second)
	for time.Now().Before(deadline) {
		out, err := client.ReceiveMessage(context.Background(), &sqs.ReceiveMessageInput{QueueUrl: &dlqURL, MaxNumberOfMessages: 1, WaitTimeSeconds: 2})
		if err != nil {
			t.Logf("receive dlq error: %v", err)
			continue
		}
		if len(out.Messages) > 0 {
			found = true
			break
		}
	}

	if !found {
		t.Fatal("expected message to be moved to DLQ but none found")
	}
}

// failingProcessor implements the worker.Processor interface and always fails
type failingProcessor struct{}

func (f *failingProcessor) Execute(ctx context.Context, event domain.RawEvent) error {
	return fmt.Errorf("forced fail")
}
