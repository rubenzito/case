package config

import (
	"os"
	"strconv"
)

type Config struct {
	AWSEndpoint          string
	AWSRegion            string
	ProcessedEventsQueue string
	DynamoDBEndpoint     string
	EventsTable          string
	SummaryTable         string
	APIPort              string
	WorkerCount          int
}

func Load() Config {
	workerCount, err := strconv.Atoi(getEnv("WORKER_COUNT", "3"))
	if err != nil {
		workerCount = 3
	}

	return Config{
		AWSEndpoint:          getEnv("AWS_ENDPOINT", "http://localhost:4566"),
		AWSRegion:            getEnv("AWS_REGION", "us-east-1"),
		ProcessedEventsQueue: getEnv("PROCESSED_EVENTS_QUEUE", ""),
		DynamoDBEndpoint:     getEnv("DYNAMODB_ENDPOINT", "http://localhost:4566"),
		EventsTable:          getEnv("EVENTS_TABLE", "events"),
		SummaryTable:         getEnv("SUMMARY_TABLE", "developer_summary"),
		APIPort:              getEnv("API_PORT", "8080"),
		WorkerCount:          workerCount,
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
