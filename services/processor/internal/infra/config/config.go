package config

import (
	"log"
	"os"
	"strconv"
)

type Config struct {
	AWSEndpoint          string
	AWSRegion            string
	RawEventsQueue       string
	ProcessedEventsQueue string
	ProcessorID          string
	WorkerCount          int
}

func Load() Config {
	workerCount, err := strconv.Atoi(getEnv("WORKER_COUNT", "5"))
	if err != nil {
		log.Println("WORKER_COUNT inválido, usando 5")
		workerCount = 5
	}

	return Config{
		AWSEndpoint:          getEnv("AWS_ENDPOINT", "http://localhost:4566"),
		AWSRegion:            getEnv("AWS_REGION", "us-east-1"),
		RawEventsQueue:       getEnv("RAW_EVENTS_QUEUE", ""),
		ProcessedEventsQueue: getEnv("PROCESSED_EVENTS_QUEUE", ""),
		ProcessorID:          getEnv("PROCESSOR_ID", "processor-1"),
		WorkerCount:          workerCount,
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
