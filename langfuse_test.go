package langfuse

import (
	"context"
	"testing"

	"github.com/optible/langfuse-go/model"
)

func TestScore_WithTraceID(t *testing.T) {
	l := New(context.Background())
	defer l.Flush(context.Background())

	score := &model.Score{
		TraceID: "test-trace-id",
		Name:    "test-score",
		Value:   0.9,
	}

	result, err := l.Score(score)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.TraceID != "test-trace-id" {
		t.Errorf("expected traceID 'test-trace-id', got '%s'", result.TraceID)
	}

	if result.ID == "" {
		t.Error("expected score ID to be generated")
	}
}

func TestScore_WithSessionID(t *testing.T) {
	l := New(context.Background())
	defer l.Flush(context.Background())

	score := &model.Score{
		SessionID: "test-session-id",
		Name:      "test-score",
		Value:     0.8,
	}

	result, err := l.Score(score)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.SessionID != "test-session-id" {
		t.Errorf("expected sessionID 'test-session-id', got '%s'", result.SessionID)
	}

	if result.ID == "" {
		t.Error("expected score ID to be generated")
	}
}

func TestScore_WithBothTraceIDAndSessionID(t *testing.T) {
	l := New(context.Background())
	defer l.Flush(context.Background())

	score := &model.Score{
		TraceID:   "test-trace-id",
		SessionID: "test-session-id",
		Name:      "test-score",
		Value:     0.7,
	}

	result, err := l.Score(score)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.TraceID != "test-trace-id" {
		t.Errorf("expected traceID 'test-trace-id', got '%s'", result.TraceID)
	}

	if result.SessionID != "test-session-id" {
		t.Errorf("expected sessionID 'test-session-id', got '%s'", result.SessionID)
	}

	if result.ID == "" {
		t.Error("expected score ID to be generated")
	}
}

func TestScore_WithoutTraceIDOrSessionID(t *testing.T) {
	l := New(context.Background())
	defer l.Flush(context.Background())

	score := &model.Score{
		Name:  "test-score",
		Value: 0.9,
	}

	_, err := l.Score(score)
	if err == nil {
		t.Fatal("expected error when neither traceID nor sessionID is provided")
	}

	expectedError := "either trace ID or session ID is required"
	if err.Error() != expectedError {
		t.Errorf("expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestScore_PreservesExistingID(t *testing.T) {
	l := New(context.Background())
	defer l.Flush(context.Background())

	existingID := "existing-score-id"
	score := &model.Score{
		ID:        existingID,
		TraceID:   "test-trace-id",
		Name:      "test-score",
		Value:     0.9,
	}

	result, err := l.Score(score)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.ID != existingID {
		t.Errorf("expected ID to be preserved as '%s', got '%s'", existingID, result.ID)
	}
}
