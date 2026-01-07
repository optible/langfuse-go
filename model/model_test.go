package model

import (
	"testing"
)

func TestTextPrompt_Compile(t *testing.T) {
	prompt := &TextPrompt{
		Name:    "test-prompt",
		Version: 1,
		Prompt:  "Hello {{name}}, you are a {{role}}. Please {{action}}.",
		Type:    PromptTypeText,
	}

	result := prompt.Compile(map[string]string{
		"name":   "Alice",
		"role":   "helpful assistant",
		"action": "summarize the document",
	})

	expected := "Hello Alice, you are a helpful assistant. Please summarize the document."
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestTextPrompt_Compile_NoVariables(t *testing.T) {
	prompt := &TextPrompt{
		Name:    "test-prompt",
		Version: 1,
		Prompt:  "Hello world!",
		Type:    PromptTypeText,
	}

	result := prompt.Compile(map[string]string{})
	expected := "Hello world!"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestTextPrompt_Compile_UnusedVariables(t *testing.T) {
	prompt := &TextPrompt{
		Name:    "test-prompt",
		Version: 1,
		Prompt:  "Hello {{name}}!",
		Type:    PromptTypeText,
	}

	result := prompt.Compile(map[string]string{
		"name":  "Bob",
		"extra": "ignored",
	})

	expected := "Hello Bob!"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestTextPrompt_Compile_MissingVariables(t *testing.T) {
	prompt := &TextPrompt{
		Name:    "test-prompt",
		Version: 1,
		Prompt:  "Hello {{name}}, your role is {{role}}!",
		Type:    PromptTypeText,
	}

	result := prompt.Compile(map[string]string{
		"name": "Charlie",
	})

	// Missing variables should remain as placeholders
	expected := "Hello Charlie, your role is {{role}}!"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestChatPrompt_Compile(t *testing.T) {
	prompt := &ChatPrompt{
		Name:    "test-chat-prompt",
		Version: 1,
		Prompt: []ChatMessage{
			{Role: "system", Content: "You are a {{role}}."},
			{Role: "user", Content: "Hello, I am {{name}}. Please {{action}}."},
		},
		Type: PromptTypeChat,
	}

	result := prompt.Compile(map[string]string{
		"role":   "helpful assistant",
		"name":   "Alice",
		"action": "help me",
	})

	if len(result) != 2 {
		t.Errorf("expected 2 messages, got %d", len(result))
	}

	if result[0].Role != "system" {
		t.Errorf("expected role 'system', got %q", result[0].Role)
	}
	if result[0].Content != "You are a helpful assistant." {
		t.Errorf("expected content 'You are a helpful assistant.', got %q", result[0].Content)
	}

	if result[1].Role != "user" {
		t.Errorf("expected role 'user', got %q", result[1].Role)
	}
	if result[1].Content != "Hello, I am Alice. Please help me." {
		t.Errorf("expected content 'Hello, I am Alice. Please help me.', got %q", result[1].Content)
	}
}

func TestPrompt_IsText(t *testing.T) {
	textPrompt := &Prompt{
		TextPrompt: &TextPrompt{
			Name:   "test",
			Prompt: "Hello",
		},
	}

	if !textPrompt.IsText() {
		t.Error("expected IsText to return true")
	}
	if textPrompt.IsChat() {
		t.Error("expected IsChat to return false")
	}
}

func TestPrompt_IsChat(t *testing.T) {
	chatPrompt := &Prompt{
		ChatPrompt: &ChatPrompt{
			Name:   "test",
			Prompt: []ChatMessage{{Role: "user", Content: "Hello"}},
		},
	}

	if chatPrompt.IsText() {
		t.Error("expected IsText to return false")
	}
	if !chatPrompt.IsChat() {
		t.Error("expected IsChat to return true")
	}
}

func TestPrompt_GetName(t *testing.T) {
	textPrompt := &Prompt{
		TextPrompt: &TextPrompt{Name: "text-prompt"},
	}
	if textPrompt.GetName() != "text-prompt" {
		t.Errorf("expected 'text-prompt', got %q", textPrompt.GetName())
	}

	chatPrompt := &Prompt{
		ChatPrompt: &ChatPrompt{Name: "chat-prompt"},
	}
	if chatPrompt.GetName() != "chat-prompt" {
		t.Errorf("expected 'chat-prompt', got %q", chatPrompt.GetName())
	}

	emptyPrompt := &Prompt{}
	if emptyPrompt.GetName() != "" {
		t.Errorf("expected empty string, got %q", emptyPrompt.GetName())
	}
}

func TestPrompt_GetVersion(t *testing.T) {
	textPrompt := &Prompt{
		TextPrompt: &TextPrompt{Version: 5},
	}
	if textPrompt.GetVersion() != 5 {
		t.Errorf("expected 5, got %d", textPrompt.GetVersion())
	}

	chatPrompt := &Prompt{
		ChatPrompt: &ChatPrompt{Version: 3},
	}
	if chatPrompt.GetVersion() != 3 {
		t.Errorf("expected 3, got %d", chatPrompt.GetVersion())
	}
}

func TestPrompt_GetLabels(t *testing.T) {
	textPrompt := &Prompt{
		TextPrompt: &TextPrompt{Labels: []string{"production", "v1"}},
	}
	labels := textPrompt.GetLabels()
	if len(labels) != 2 {
		t.Errorf("expected 2 labels, got %d", len(labels))
	}
	if labels[0] != "production" || labels[1] != "v1" {
		t.Errorf("unexpected labels: %v", labels)
	}
}

func TestPrompt_GetTags(t *testing.T) {
	chatPrompt := &Prompt{
		ChatPrompt: &ChatPrompt{Tags: []string{"sales", "customer-support"}},
	}
	tags := chatPrompt.GetTags()
	if len(tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(tags))
	}
	if tags[0] != "sales" || tags[1] != "customer-support" {
		t.Errorf("unexpected tags: %v", tags)
	}
}
