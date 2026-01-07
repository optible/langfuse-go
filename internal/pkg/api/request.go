package api

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/optible/langfuse-go/model"
)

const (
	ContentTypeJSON = "application/json"
)

type Request struct{}

type Ingestion struct {
	Batch []model.IngestionEvent `json:"batch"`
}

func (t *Ingestion) Path() (string, error) {
	return "/api/public/ingestion", nil
}

func (t *Ingestion) Encode() (io.Reader, error) {
	jsonBytes, err := json.Marshal(t)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(jsonBytes), nil
}

func (t *Ingestion) ContentType() string {
	return ContentTypeJSON
}

// GetPromptRequest represents a request to get a prompt by name
type GetPromptRequest struct {
	Name    string
	Version *int
	Label   *string
}

// Path returns the API path for getting a prompt
func (r *GetPromptRequest) Path() (string, error) {
	return "/api/public/v2/prompts/" + r.Name, nil
}

// Encode returns nil for GET requests (no body)
func (r *GetPromptRequest) Encode() (io.Reader, error) {
	return nil, nil
}

// ContentType returns empty string for GET requests
func (r *GetPromptRequest) ContentType() string {
	return ""
}
