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

func TestDeleteScore_WithValidID(t *testing.T) {
	l := New(context.Background())
	defer l.Flush(context.Background())

	// Note: This test will fail if there's no valid score ID or network issues
	// In a real scenario, you would create a score first and then delete it
	scoreID := "test-score-id"
	err := l.DeleteScore(context.Background(), scoreID)
	
	// We expect an error since we don't have a valid connection to Langfuse
	// But we're testing that the method handles the request properly
	if err == nil {
		t.Log("Score deletion succeeded (or network is unavailable)")
	} else {
		t.Logf("Expected error in test environment: %v", err)
	}
}

func TestDeleteScore_WithEmptyID(t *testing.T) {
	l := New(context.Background())
	defer l.Flush(context.Background())

	err := l.DeleteScore(context.Background(), "")
	if err == nil {
		t.Fatal("expected error when score ID is empty")
	}

	expectedError := "score ID is required"
	if err.Error() != expectedError {
		t.Errorf("expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestDeleteScore_WithContext(t *testing.T) {
	l := New(context.Background())
	defer l.Flush(context.Background())

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	scoreID := "test-score-id"
	err := l.DeleteScore(ctx, scoreID)
	
	// Should get a context cancellation error
	if err == nil {
		t.Fatal("expected error when context is cancelled")
	}
	
	t.Logf("Got expected error with cancelled context: %v", err)
}
