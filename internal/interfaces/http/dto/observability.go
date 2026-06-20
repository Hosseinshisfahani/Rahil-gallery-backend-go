package dto

import (
	"encoding/json"
	"time"

	domain "github.com/rahil-gallery/rahil-gallery-server/internal/domain/observability"
)

type IngestEventRequest struct {
	Source     string          `json:"source"`
	Level      string          `json:"level"`
	Message    string          `json:"message"`
	StackTrace *string         `json:"stackTrace,omitempty"`
	Route      *string         `json:"route,omitempty"`
	Method     *string         `json:"method,omitempty"`
	StatusCode *int            `json:"statusCode,omitempty"`
	RequestID  *string         `json:"requestId,omitempty"`
	Metadata   json.RawMessage `json:"metadata,omitempty"`
}

func (r IngestEventRequest) ToInput(userAgent string) domain.IngestInput {
	return domain.IngestInput{
		Source:     r.Source,
		Level:      r.Level,
		Message:    r.Message,
		StackTrace: r.StackTrace,
		Route:      r.Route,
		Method:     r.Method,
		StatusCode: r.StatusCode,
		RequestID:  r.RequestID,
		UserAgent:  optionalString(userAgent),
		Metadata:   r.Metadata,
	}
}

type ObservabilityEventResponse struct {
	ID         string          `json:"id"`
	Source     string          `json:"source"`
	Level      string          `json:"level"`
	Message    string          `json:"message"`
	StackTrace *string         `json:"stackTrace,omitempty"`
	Route      *string         `json:"route,omitempty"`
	Method     *string         `json:"method,omitempty"`
	StatusCode *int            `json:"statusCode,omitempty"`
	RequestID  *string         `json:"requestId,omitempty"`
	UserAgent  *string         `json:"userAgent,omitempty"`
	Metadata   json.RawMessage `json:"metadata,omitempty"`
	CreatedAt  string          `json:"createdAt"`
}

func ToObservabilityEvent(ev domain.Event) ObservabilityEventResponse {
	return ObservabilityEventResponse{
		ID:         ev.ID.String(),
		Source:     ev.Source,
		Level:      ev.Level,
		Message:    ev.Message,
		StackTrace: ev.StackTrace,
		Route:      ev.Route,
		Method:     ev.Method,
		StatusCode: ev.StatusCode,
		RequestID:  ev.RequestID,
		UserAgent:  ev.UserAgent,
		Metadata:   ev.Metadata,
		CreatedAt:  ev.CreatedAt.UTC().Format(time.RFC3339),
	}
}

func optionalString(v string) *string {
	if v == "" {
		return nil
	}
	return &v
}
