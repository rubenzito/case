package domain

import "time"

// ProcessedEvent é o evento que chega na fila processed-events
type ProcessedEvent struct {
	EventID     string    `json:"event_id" dynamodbav:"event_id"`
	DeveloperID string    `json:"developer_id" dynamodbav:"developer_id"`
	MetricType  string    `json:"metric_type" dynamodbav:"metric_type"`
	Value       float64   `json:"value" dynamodbav:"value"`
	Repository  string    `json:"repository" dynamodbav:"repository"`
	Timestamp   time.Time `json:"timestamp" dynamodbav:"timestamp"`
	ProcessedAt time.Time `json:"processed_at" dynamodbav:"processed_at"`
	ProcessorID string    `json:"processor_id" dynamodbav:"processor_id"`
}

// DeveloperSummary é o resumo agregado por desenvolvedor no DynamoDB
type DeveloperSummary struct {
	DeveloperID       string    `json:"developer_id" dynamodbav:"developer_id"`
	TotalCommits      int64     `json:"total_commits" dynamodbav:"total_commits"`
	TotalPullRequests int64     `json:"total_pull_requests" dynamodbav:"total_pull_requests"`
	TotalReviewTime   float64   `json:"total_review_time_minutes" dynamodbav:"total_review_time_minutes"`
	ReviewTimeCount   int64     `json:"review_time_count" dynamodbav:"review_time_count"`
	EventsProcessed   int64     `json:"events_processed" dynamodbav:"events_processed"`
	LastActivity      time.Time `json:"last_activity" dynamodbav:"last_activity"`
}

// SummaryResponse é a resposta da API de resumo
type SummaryResponse struct {
	DeveloperID          string    `json:"developer_id"`
	TotalCommits         int64     `json:"total_commits"`
	TotalPullRequests    int64     `json:"total_pull_requests"`
	AvgReviewTimeMinutes float64   `json:"avg_review_time_minutes"`
	EventsProcessed      int64     `json:"events_processed"`
	LastActivity         time.Time `json:"last_activity"`
}

// ToSummaryResponse converte o DeveloperSummary para o formato da API
func (s *DeveloperSummary) ToSummaryResponse() SummaryResponse {
	avg := 0.0
	if s.ReviewTimeCount > 0 {
		avg = s.TotalReviewTime / float64(s.ReviewTimeCount)
	}
	return SummaryResponse{
		DeveloperID:          s.DeveloperID,
		TotalCommits:         s.TotalCommits,
		TotalPullRequests:    s.TotalPullRequests,
		AvgReviewTimeMinutes: avg,
		EventsProcessed:      s.EventsProcessed,
		LastActivity:         s.LastActivity,
	}
}
