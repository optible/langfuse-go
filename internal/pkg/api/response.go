package api

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/henomis/restclientgo"
)

type Response struct {
	Code      int       `json:"-"`
	RawBody   *string   `json:"-"`
	Successes []Success `json:"successes"`
	Errors    []Error   `json:"errors"`
}

type Success struct {
	ID     string `json:"id"`
	Status int    `json:"status"`
}

type Error struct {
	ID      string `json:"id"`
	Status  int    `json:"status"`
	Message string `json:"message"`
	Error   string `json:"error"`
}

func (r *Response) IsSuccess() bool {
	return r.Code < http.StatusBadRequest
}

func (r *Response) SetStatusCode(code int) error {
	r.Code = code
	return nil
}

func (r *Response) SetBody(body io.Reader) error {
	b, err := io.ReadAll(body)
	if err != nil {
		return err
	}

	s := string(b)
	r.RawBody = &s

	return nil
}

func (r *Response) AcceptContentType() string {
	return ContentTypeJSON
}

func (r *Response) Decode(body io.Reader) error {
	return json.NewDecoder(body).Decode(r)
}

func (r *Response) SetHeaders(_ restclientgo.Headers) error {
	return nil
}

type IngestionResponse struct {
	Response
}

// PromptResponse represents the response from fetching a prompt
type PromptResponse struct {
	Code    int     `json:"-"`
	RawBody *string `json:"-"`

	// Common fields
	Type          string   `json:"type"`
	Name          string   `json:"name"`
	Version       int      `json:"version"`
	Config        any      `json:"config"`
	Labels        []string `json:"labels"`
	Tags          []string `json:"tags"`
	CommitMessage *string  `json:"commitMessage,omitempty"`

	// For text prompts, Prompt is a string
	// For chat prompts, Prompt is []ChatMessageResponse
	Prompt any `json:"prompt"`

	// Optional fields
	ResolutionGraph map[string]any `json:"resolutionGraph,omitempty"`
}

// ChatMessageResponse represents a chat message in the API response
type ChatMessageResponse struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func (r *PromptResponse) IsSuccess() bool {
	return r.Code < http.StatusBadRequest
}

func (r *PromptResponse) SetStatusCode(code int) error {
	r.Code = code
	return nil
}

func (r *PromptResponse) SetBody(body io.Reader) error {
	b, err := io.ReadAll(body)
	if err != nil {
		return err
	}

	s := string(b)
	r.RawBody = &s

	return nil
}

func (r *PromptResponse) AcceptContentType() string {
	return ContentTypeJSON
}

func (r *PromptResponse) Decode(body io.Reader) error {
	return json.NewDecoder(body).Decode(r)
}

func (r *PromptResponse) SetHeaders(_ restclientgo.Headers) error {
	return nil
}
