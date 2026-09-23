package domain

import (
	"encoding/json"
	"time"
)

const (
	ImportStatusDraft     = "draft"
	ImportStatusValidated = "validated"
	ImportStatusPublished = "published"
)

type Import struct {
	ID           string          `json:"id"`
	ResourceType string          `json:"resource_type"`
	Status       string          `json:"status"`
	CreatedBy    *string         `json:"created_by"`
	LineCount    int32           `json:"line_count"`
	ValidCount   int32           `json:"valid_count"`
	InvalidCount int32           `json:"invalid_count"`
	Report       json.RawMessage `json:"report" swaggertype:"array,object"`
	CreatedAt    time.Time       `json:"created_at"`
	ValidatedAt  *time.Time      `json:"validated_at"`
	PublishedAt  *time.Time      `json:"published_at"`
}

type ImportRow struct {
	ID         string          `json:"id"`
	LineNumber int32           `json:"line_number"`
	Payload    json.RawMessage `json:"payload" swaggertype:"object"`
	Valid      bool            `json:"valid"`
	Error      *string         `json:"error,omitempty"`
}

type ImportCreate struct {
	ResourceType string            `json:"resource_type"`
	Rows         []json.RawMessage `json:"rows" swaggertype:"array,object"`
}

type ImportReportEntry struct {
	Line  int32  `json:"line"`
	Error string `json:"error"`
}
