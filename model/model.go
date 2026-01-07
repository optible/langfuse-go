package model

import "time"

type IngestionEventType string

const (
	IngestionEventTypeTraceCreate      = "trace-create"
	IngestionEventTypeGenerationCreate = "generation-create"
	IngestionEventTypeGenerationUpdate = "generation-update"
	IngestionEventTypeScoreCreate      = "score-create"
	IngestionEventTypeSpanCreate       = "span-create"
	IngestionEventTypeSpanUpdate       = "span-update"
	IngestionEventTypeEventCreate      = "event-create"
)

type IngestionEvent struct {
	Type      IngestionEventType `json:"type"`
	ID        string             `json:"id"`
	Timestamp time.Time          `json:"timestamp"`
	Metadata  any
	Body      any `json:"body"`
}

type Trace struct {
	ID        string     `json:"id,omitempty"`
	Timestamp *time.Time `json:"timestamp,omitempty"`
	Name      string     `json:"name,omitempty"`
	UserID    string     `json:"userId,omitempty"`
	Input     any        `json:"input,omitempty"`
	Output    any        `json:"output,omitempty"`
	SessionID string     `json:"sessionId,omitempty"`
	Release   string     `json:"release,omitempty"`
	Version   string     `json:"version,omitempty"`
	Metadata  any        `json:"metadata,omitempty"`
	Tags      []string   `json:"tags,omitempty"`
	Public    bool       `json:"public,omitempty"`
}

type ObservationLevel string

const (
	ObservationLevelDebug   ObservationLevel = "DEBUG"
	ObservationLevelDefault ObservationLevel = "DEFAULT"
	ObservationLevelWarning ObservationLevel = "WARNING"
	ObservationLevelError   ObservationLevel = "ERROR"
)

type Generation struct {
	TraceID             string           `json:"traceId,omitempty"`
	Name                string           `json:"name,omitempty"`
	StartTime           *time.Time       `json:"startTime,omitempty"`
	Metadata            any              `json:"metadata,omitempty"`
	Input               any              `json:"input,omitempty"`
	Output              any              `json:"output,omitempty"`
	Level               ObservationLevel `json:"level,omitempty"`
	StatusMessage       string           `json:"statusMessage,omitempty"`
	ParentObservationID string           `json:"parentObservationId,omitempty"`
	Version             string           `json:"version,omitempty"`
	ID                  string           `json:"id,omitempty"`
	EndTime             *time.Time       `json:"endTime,omitempty"`
	CompletionStartTime *time.Time       `json:"completionStartTime,omitempty"`
	Model               string           `json:"model,omitempty"`
	ModelParameters     any              `json:"modelParameters,omitempty"`
	Usage               Usage            `json:"usage,omitempty"`
	PromptName          string           `json:"promptName,omitempty"`
	PromptVersion       int              `json:"promptVersion,omitempty"`
}

type Usage struct {
	Input      int       `json:"input,omitempty"`
	Output     int       `json:"output,omitempty"`
	Total      int       `json:"total,omitempty"`
	Unit       UsageUnit `json:"unit,omitempty"`
	InputCost  float64   `json:"inputCost,omitempty"`
	OutputCost float64   `json:"outputCost,omitempty"`
	TotalCost  float64   `json:"totalCost,omitempty"`

	PromptTokens     int `json:"promptTokens,omitempty"`
	CompletionTokens int `json:"completionTokens,omitempty"`
	TotalTokens      int `json:"totalTokens,omitempty"`
}

type UsageUnit string

const (
	ModelUsageUnitCharacters   UsageUnit = "CHARACTERS"
	ModelUsageUnitTokens       UsageUnit = "TOKENS"
	ModelUsageUnitMilliseconds UsageUnit = "MILLISECONDS"
	ModelUsageUnitSeconds      UsageUnit = "SECONDS"
	ModelUsageUnitImages       UsageUnit = "IMAGES"
)

type Score struct {
	ID            string  `json:"id,omitempty"`
	TraceID       string  `json:"traceId,omitempty"`
	Name          string  `json:"name,omitempty"`
	Value         float64 `json:"value,omitempty"`
	ObservationID string  `json:"observationId,omitempty"`
	Comment       string  `json:"comment,omitempty"`
}

type Span struct {
	TraceID             string           `json:"traceId,omitempty"`
	Name                string           `json:"name,omitempty"`
	StartTime           *time.Time       `json:"startTime,omitempty"`
	Metadata            any              `json:"metadata,omitempty"`
	Input               any              `json:"input,omitempty"`
	Output              any              `json:"output,omitempty"`
	Level               ObservationLevel `json:"level,omitempty"`
	StatusMessage       string           `json:"statusMessage,omitempty"`
	ParentObservationID string           `json:"parentObservationId,omitempty"`
	Version             string           `json:"version,omitempty"`
	ID                  string           `json:"id,omitempty"`
	EndTime             *time.Time       `json:"endTime,omitempty"`
}

type Event struct {
	TraceID             string           `json:"traceId,omitempty"`
	Name                string           `json:"name,omitempty"`
	StartTime           *time.Time       `json:"startTime,omitempty"`
	Metadata            any              `json:"metadata,omitempty"`
	Input               any              `json:"input,omitempty"`
	Output              any              `json:"output,omitempty"`
	Level               ObservationLevel `json:"level,omitempty"`
	StatusMessage       string           `json:"statusMessage,omitempty"`
	ParentObservationID string           `json:"parentObservationId,omitempty"`
	Version             string           `json:"version,omitempty"`
	ID                  string           `json:"id,omitempty"`
}

type M map[string]interface{}

// PromptType represents the type of prompt (text or chat)
type PromptType string

const (
	PromptTypeText PromptType = "text"
	PromptTypeChat PromptType = "chat"
)

// ChatMessage represents a single message in a chat prompt
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// TextPrompt represents a text-based prompt
type TextPrompt struct {
	Name            string         `json:"name"`
	Version         int            `json:"version"`
	Config          any            `json:"config"`
	Labels          []string       `json:"labels"`
	Tags            []string       `json:"tags"`
	Prompt          string         `json:"prompt"`
	Type            PromptType     `json:"type"`
	CommitMessage   *string        `json:"commitMessage,omitempty"`
	ResolutionGraph map[string]any `json:"resolutionGraph,omitempty"`
}

// Compile replaces variables in the prompt with the provided values.
// Variables are in the format {{variableName}}.
func (p *TextPrompt) Compile(variables map[string]string) string {
	result := p.Prompt
	for key, value := range variables {
		placeholder := "{{" + key + "}}"
		result = replaceAll(result, placeholder, value)
	}
	return result
}

// ChatPrompt represents a chat-based prompt with multiple messages
type ChatPrompt struct {
	Name            string         `json:"name"`
	Version         int            `json:"version"`
	Config          any            `json:"config"`
	Labels          []string       `json:"labels"`
	Tags            []string       `json:"tags"`
	Prompt          []ChatMessage  `json:"prompt"`
	Type            PromptType     `json:"type"`
	CommitMessage   *string        `json:"commitMessage,omitempty"`
	ResolutionGraph map[string]any `json:"resolutionGraph,omitempty"`
}

// Compile replaces variables in all chat messages with the provided values.
// Variables are in the format {{variableName}}.
func (p *ChatPrompt) Compile(variables map[string]string) []ChatMessage {
	result := make([]ChatMessage, len(p.Prompt))
	for i, msg := range p.Prompt {
		content := msg.Content
		for key, value := range variables {
			placeholder := "{{" + key + "}}"
			content = replaceAll(content, placeholder, value)
		}
		result[i] = ChatMessage{
			Role:    msg.Role,
			Content: content,
		}
	}
	return result
}

// Prompt is a union type that can be either TextPrompt or ChatPrompt
type Prompt struct {
	*TextPrompt
	*ChatPrompt
}

// IsText returns true if the prompt is a text prompt
func (p *Prompt) IsText() bool {
	return p.TextPrompt != nil
}

// IsChat returns true if the prompt is a chat prompt
func (p *Prompt) IsChat() bool {
	return p.ChatPrompt != nil
}

// GetName returns the prompt name
func (p *Prompt) GetName() string {
	if p.TextPrompt != nil {
		return p.TextPrompt.Name
	}
	if p.ChatPrompt != nil {
		return p.ChatPrompt.Name
	}
	return ""
}

// GetVersion returns the prompt version
func (p *Prompt) GetVersion() int {
	if p.TextPrompt != nil {
		return p.TextPrompt.Version
	}
	if p.ChatPrompt != nil {
		return p.ChatPrompt.Version
	}
	return 0
}

// GetConfig returns the prompt config
func (p *Prompt) GetConfig() any {
	if p.TextPrompt != nil {
		return p.TextPrompt.Config
	}
	if p.ChatPrompt != nil {
		return p.ChatPrompt.Config
	}
	return nil
}

// GetLabels returns the prompt labels
func (p *Prompt) GetLabels() []string {
	if p.TextPrompt != nil {
		return p.TextPrompt.Labels
	}
	if p.ChatPrompt != nil {
		return p.ChatPrompt.Labels
	}
	return nil
}

// GetTags returns the prompt tags
func (p *Prompt) GetTags() []string {
	if p.TextPrompt != nil {
		return p.TextPrompt.Tags
	}
	if p.ChatPrompt != nil {
		return p.ChatPrompt.Tags
	}
	return nil
}

// replaceAll is a simple string replacement function
func replaceAll(s, old, new string) string {
	result := ""
	for {
		i := indexOf(s, old)
		if i == -1 {
			return result + s
		}
		result += s[:i] + new
		s = s[i+len(old):]
	}
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
