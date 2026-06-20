package observability

import (
	"encoding/json"
	"time"

	"github.com/rahil-gallery/rahil-gallery-server/internal/domain/shared"
)

const (
	SourceAPI    = "api"
	SourceClient = "client"

	LevelError = "error"
	LevelWarn  = "warn"
	LevelInfo  = "info"

	DefaultRetentionDays = 7
)

type Event struct {
	ID         shared.ID
	Source     string
	Level      string
	Message    string
	StackTrace *string
	Route      *string
	Method     *string
	StatusCode *int
	RequestID  *string
	UserAgent  *string
	Metadata   json.RawMessage
	CreatedAt  time.Time
}

type IngestInput struct {
	Source     string
	Level      string
	Message    string
	StackTrace *string
	Route      *string
	Method     *string
	StatusCode *int
	RequestID  *string
	UserAgent  *string
	Metadata   json.RawMessage
}

type ListFilter struct {
	Source   string
	Level    string
	From     *time.Time
	To       *time.Time
	Route    string
	Page     int
	PerPage  int
}

type DailyCount struct {
	Day    string `json:"day"`
	Source string `json:"source"`
	Level  string `json:"level"`
	Count  int    `json:"count"`
}

type Summary struct {
	RetentionDays int          `json:"retentionDays"`
	Total         int          `json:"total"`
	Last24Hours   int          `json:"last24Hours"`
	BySource      map[string]int `json:"bySource"`
	Daily         []DailyCount `json:"daily"`
}
