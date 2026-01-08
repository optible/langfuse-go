package main

import (
	"context"
	"fmt"
	"time"

	"github.com/optible/langfuse-go"
	"github.com/optible/langfuse-go/model"
)

func main() {
	ctx := context.Background()

	// Create a new Langfuse client with custom prompt cache TTL
	l := langfuse.New(ctx).WithPromptCacheTTL(10 * time.Minute)

	// Fetch a prompt by name (defaults to "production" label)
	prompt, err := l.GetPrompt(ctx, "movie-critic", nil)
	if err != nil {
		fmt.Printf("Error fetching prompt: %v\n", err)
		return
	}

	fmt.Printf("Fetched prompt: %s (version %d)\n", prompt.GetName(), prompt.GetVersion())

	// Check if it's a text or chat prompt
	if prompt.IsText() {
		// Compile the text prompt with variables
		compiled := prompt.TextPrompt.Compile(map[string]string{
			"movie": "The Matrix",
			"style": "technical",
		})
		fmt.Printf("Compiled prompt: %s\n", compiled)
	}

	if prompt.IsChat() {
		// Compile the chat prompt with variables
		messages := prompt.ChatPrompt.Compile(map[string]string{
			"movie": "The Matrix",
			"style": "technical",
		})
		for i, msg := range messages {
			fmt.Printf("Message %d [%s]: %s\n", i+1, msg.Role, msg.Content)
		}
	}

	// Fetch a specific version of a prompt
	version := 2
	promptV2, err := l.GetPrompt(ctx, "movie-critic", &langfuse.GetPromptOptions{
		Version: &version,
	})
	if err != nil {
		fmt.Printf("Error fetching prompt v2: %v\n", err)
	} else {
		fmt.Printf("Fetched prompt version %d\n", promptV2.GetVersion())
	}

	// Fetch a prompt with a specific label (e.g., "staging")
	label := "staging"
	promptStaging, err := l.GetPrompt(ctx, "movie-critic", &langfuse.GetPromptOptions{
		Label: &label,
	})
	if err != nil {
		fmt.Printf("Error fetching staging prompt: %v\n", err)
	} else {
		fmt.Printf("Fetched prompt with labels: %v\n", promptStaging.GetLabels())
	}

	// Fetch with a fallback prompt in case of API failure
	fallback := &langfuse.GetPromptOptions{
		FallbackPrompt: &model.Prompt{
			TextPrompt: &model.TextPrompt{
				Name:    "movie-critic",
				Version: 0,
				Prompt:  "Please review {{movie}} in a {{style}} style.",
				Type:    model.PromptTypeText,
			},
		},
	}
	promptWithFallback, err := l.GetPrompt(ctx, "movie-critic", fallback)
	if err != nil {
		fmt.Printf("Error (should not happen with fallback): %v\n", err)
	} else {
		fmt.Printf("Fetched prompt (or fallback): %s\n", promptWithFallback.GetName())
	}

	// Force refresh from API (bypass cache)
	promptFresh, err := l.GetPrompt(ctx, "movie-critic", &langfuse.GetPromptOptions{
		ForceRefresh: true,
	})
	if err != nil {
		fmt.Printf("Error force refreshing prompt: %v\n", err)
	} else {
		fmt.Printf("Fresh prompt fetched: version %d\n", promptFresh.GetVersion())
	}

	// Clear the prompt cache if needed
	l.ClearPromptCache()
}
