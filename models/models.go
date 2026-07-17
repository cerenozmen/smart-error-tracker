package models

import (
	"time"
)

type Project struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	APIKey    string    `json:"api_key"`
	CreatedAt time.Time `json:"created_at"`
}


type Issue struct {
	ID           int        `json:"id"`
	ProjectID    int        `json:"project_id"`
	Fingerprint  string     `json:"fingerprint"` // Hatanın benzersiz hash'i
	Title        string     `json:"title"`       // Hata mesajı özeti
	Status       string     `json:"status"`      // unresolved, resolved, ignored
	AIAnalysis   *string    `json:"ai_analysis"` // NULL olabileceği için pointer yapıyoruz
	AIAnalyzedAt *time.Time `json:"ai_analyzed_at"`
	LastSeenAt   time.Time  `json:"last_seen_at"`
	CreatedAt    time.Time  `json:"created_at"`
}

type Error struct {
	ID          int64                  `json:"id"`
	IssueID     int                    `json:"issue_id"`
	Message     string                 `json:"message"`
	StackTrace  string                 `json:"stack_trace"`
	Environment string                 `json:"environment"` // production, development vb.
	Payload     map[string]interface{} `json:"payload"`     // JSONB verisi için Go karşılığı
	CreatedAt   time.Time              `json:"created_at"`
}

type ErrorRequest struct {
	APIKey      string                 `json:"api_key"`
	Message     string                 `json:"message"`
	StackTrace  string                 `json:"stack_trace"`
	Environment string                 `json:"environment"`
	Payload     map[string]interface{} `json:"payload"`
}
