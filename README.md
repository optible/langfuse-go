# Langfuse Go SDK

[![GoDoc](https://godoc.org/github.com/optible/langfuse-go?status.svg)](https://godoc.org/github.com/optible/langfuse-go) [![Go Report Card](https://goreportcard.com/badge/github.com/optible/langfuse-go)](https://goreportcard.com/report/github.com/optible/langfuse-go) [![GitHub release](https://img.shields.io/github/release/optible/langfuse-go.svg)](https://github.com/optible/langfuse-go/releases)

**Maintained by [Optible](https://github.com/optible)**

This is [Langfuse](https://langfuse.com)'s **unofficial** Go client, designed to enable you to use Langfuse's services easily from your own applications.

## Langfuse

[Langfuse](https://langfuse.com) provides traces, evals, prompt management and metrics to debug and improve your LLM application.

## Features

- **Full Ingestion API Support**: Traces, Generations, Spans, Events, and Scores
- **Prompt Management**: Fetch and cache prompts with version control and labels
- **Smart Caching**: Built-in caching with configurable TTL for prompts
- **Fallback Support**: Graceful degradation with fallback prompts
- **Batch Processing**: Efficient batching of ingestion events
- **Type Safety**: Strongly typed models for all API interactions
- **Context Support**: Full Go context support for cancellation and timeouts

## API Support

| **Feature**  | **Status** | **Description** |
| --- | --- | --- |
| Trace | 🟢 | Create and manage execution traces |
| Generation | 🟢 | Track LLM generations with metadata |
| Span | 🟢 | Measure execution spans within traces |
| Event | 🟢 | Log custom events in traces |
| Score | 🟢 | Add evaluations and scores to traces/sessions |
| GetPrompt | 🟢 | Fetch prompts with caching, versioning, and labels |




## Getting started

### Installation

You can load langfuse-go into your project by using:
```
go get github.com/optible/langfuse-go
```


### Configuration
Just like the official Python SDK, these three environment variables will be used to configure the Langfuse client:

- `LANGFUSE_HOST`: The host of the Langfuse service.
- `LANGFUSE_PUBLIC_KEY`: Your public key for the Langfuse service.
- `LANGFUSE_SECRET_KEY`: Your secret key for the Langfuse service.


### Usage

Please refer to the [examples folder](examples/cmd/) to see how to use the SDK.

#### Basic Ingestion Example

Here's a simple example showing how to create traces, spans, generations, events, and scores:

```go
package main

import (
	"context"

	"github.com/optible/langfuse-go"
	"github.com/optible/langfuse-go/model"
)

func main() {
	ctx := context.Background()
	l := langfuse.New(ctx)

	// Create a trace
	trace, err := l.Trace(&model.Trace{
		Name:      "my-llm-app",
		SessionID: "user-session-123",
	})
	if err != nil {
		panic(err)
	}

	// Create a span within the trace
	span, err := l.Span(&model.Span{
		Name:    "data-processing",
		TraceID: trace.ID,
	}, nil)
	if err != nil {
		panic(err)
	}

	// Track an LLM generation
	generation, err := l.Generation(
		&model.Generation{
			TraceID: trace.ID,
			Name:    "chat-completion",
			Model:   "gpt-3.5-turbo",
			ModelParameters: model.M{
				"maxTokens":   "1000",
				"temperature": "0.9",
			},
			Input: []model.M{
				{
					"role":    "system",
					"content": "You are a helpful assistant.",
				},
				{
					"role":    "user",
					"content": "Please generate a summary of the following documents...",
				},
			},
			Metadata: model.M{
				"environment": "production",
			},
		},
		&span.ID,
	)
	if err != nil {
		panic(err)
	}

	// Log an event
	_, err = l.Event(
		&model.Event{
			Name:    "user-feedback",
			TraceID: trace.ID,
			Input: model.M{
				"feedback": "positive",
			},
		},
		&generation.ID,
	)
	if err != nil {
		panic(err)
	}

	// Update generation with output
	generation.Output = model.M{
		"completion": "Here is the summary...",
	}
	_, err = l.GenerationEnd(generation)
	if err != nil {
		panic(err)
	}

	// Add a score
	_, err = l.Score(
		&model.Score{
			TraceID: trace.ID,
			Name:    "quality-score",
			Value:   0.95,
			Comment: "High quality response",
		},
	)
	if err != nil {
		panic(err)
	}

	// End the span
	_, err = l.SpanEnd(span)
	if err != nil {
		panic(err)
	}

	// Flush all pending events
	l.Flush(ctx)
}
```

#### Prompt Management Example

The SDK includes powerful prompt management capabilities with caching, versioning, and fallback support:

```go
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

	// Create client with custom cache TTL
	l := langfuse.New(ctx).WithPromptCacheTTL(10 * time.Minute)

	// Fetch a prompt (defaults to "production" label)
	prompt, err := l.GetPrompt(ctx, "movie-critic", nil)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Fetched prompt: %s (version %d)\n", prompt.GetName(), prompt.GetVersion())

	// Use text prompts
	if prompt.IsText() {
		compiled := prompt.TextPrompt.Compile(map[string]string{
			"movie": "The Matrix",
			"style": "technical",
		})
		fmt.Printf("Compiled prompt: %s\n", compiled)
	}

	// Use chat prompts
	if prompt.IsChat() {
		messages := prompt.ChatPrompt.Compile(map[string]string{
			"movie": "The Matrix",
			"style": "technical",
		})
		for i, msg := range messages {
			fmt.Printf("Message %d [%s]: %s\n", i+1, msg.Role, msg.Content)
		}
	}

	// Fetch a specific version
	version := 2
	promptV2, err := l.GetPrompt(ctx, "movie-critic", &langfuse.GetPromptOptions{
		Version: &version,
	})
	if err != nil {
		panic(err)
	}

	// Fetch by label (e.g., "staging")
	label := "staging"
	promptStaging, err := l.GetPrompt(ctx, "movie-critic", &langfuse.GetPromptOptions{
		Label: &label,
	})
	if err != nil {
		panic(err)
	}

	// Use fallback prompt for high availability
	promptWithFallback, err := l.GetPrompt(ctx, "movie-critic", &langfuse.GetPromptOptions{
		FallbackPrompt: &model.Prompt{
			TextPrompt: &model.TextPrompt{
				Name:    "movie-critic",
				Version: 0,
				Prompt:  "Please review {{movie}} in a {{style}} style.",
				Type:    model.PromptTypeText,
			},
		},
	})
	if err != nil {
		panic(err)
	}

	// Force refresh (bypass cache)
	promptFresh, err := l.GetPrompt(ctx, "movie-critic", &langfuse.GetPromptOptions{
		ForceRefresh: true,
	})
	if err != nil {
		panic(err)
	}

	// Clear cache when needed
	l.ClearPromptCache()
}
```
