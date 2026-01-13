package langfuse

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/optible/langfuse-go/internal/pkg/api"
	"github.com/optible/langfuse-go/internal/pkg/cache"
	"github.com/optible/langfuse-go/internal/pkg/observer"
	"github.com/optible/langfuse-go/model"
)

const (
	defaultFlushInterval   = 500 * time.Millisecond
	defaultPromptCacheTTL  = 5 * time.Minute
)

type Langfuse struct {
	flushInterval  time.Duration
	client         *api.Client
	observer       *observer.Observer[model.IngestionEvent]
	promptCache    *cache.Cache[*model.Prompt]
	promptCacheTTL time.Duration
}

// GetPromptOptions contains options for fetching a prompt
type GetPromptOptions struct {
	// Version specifies which version of the prompt to fetch.
	// If not set, fetches by label (defaults to "production").
	Version *int

	// Label specifies which label to fetch (e.g., "production", "staging").
	// If not set and Version is not set, defaults to "production".
	Label *string

	// FallbackPrompt is returned if the prompt cannot be fetched from the API.
	// This provides guaranteed availability even during network issues.
	FallbackPrompt *model.Prompt

	// CacheTTL overrides the default cache TTL for this specific request.
	// If not set, uses the client's default cache TTL.
	CacheTTL *time.Duration

	// FetchTimeout sets the timeout for the API request.
	// If not set, uses the context's deadline.
	FetchTimeout *time.Duration

	// ForceRefresh bypasses the cache and fetches directly from the API.
	ForceRefresh bool
}

func New(ctx context.Context) *Langfuse {
	client := api.New()

	l := &Langfuse{
		flushInterval:  defaultFlushInterval,
		client:         client,
		promptCache:    cache.New[*model.Prompt](defaultPromptCacheTTL),
		promptCacheTTL: defaultPromptCacheTTL,
		observer: observer.NewObserver(
			ctx,
			func(ctx context.Context, events []model.IngestionEvent) {
				err := ingest(ctx, client, events)
				if err != nil {
					fmt.Println(err)
				}
			},
		),
	}

	return l
}

func (l *Langfuse) WithFlushInterval(d time.Duration) *Langfuse {
	l.flushInterval = d
	return l
}

// WithPromptCacheTTL sets the default cache TTL for prompts.
// Default is 5 minutes.
func (l *Langfuse) WithPromptCacheTTL(ttl time.Duration) *Langfuse {
	l.promptCacheTTL = ttl
	l.promptCache.SetTTL(ttl)
	return l
}

// GetPrompt fetches a prompt by name.
// By default, fetches the "production" labeled version.
// Uses caching to minimize API calls.
func (l *Langfuse) GetPrompt(ctx context.Context, name string, opts *GetPromptOptions) (*model.Prompt, error) {
	if opts == nil {
		opts = &GetPromptOptions{}
	}

	// Build cache key
	cacheKey := l.buildPromptCacheKey(name, opts)

	// Check cache unless force refresh is requested
	if !opts.ForceRefresh {
		if prompt, found, expired := l.promptCache.Get(cacheKey); found {
			if !expired {
				return prompt, nil
			}
			// If expired but we have a cached value, try to refresh
			// but return cached value if refresh fails
			newPrompt, err := l.fetchPrompt(ctx, name, opts)
			if err != nil {
				// Return stale cached value on error
				return prompt, nil
			}
			return newPrompt, nil
		}
	}

	// Fetch from API
	prompt, err := l.fetchPrompt(ctx, name, opts)
	if err != nil {
		// If we have a fallback, return it
		if opts.FallbackPrompt != nil {
			return opts.FallbackPrompt, nil
		}
		return nil, err
	}

	return prompt, nil
}

// buildPromptCacheKey creates a unique cache key for a prompt request
func (l *Langfuse) buildPromptCacheKey(name string, opts *GetPromptOptions) string {
	key := "prompt:" + name
	if opts.Version != nil {
		key += ":v" + strconv.Itoa(*opts.Version)
	} else if opts.Label != nil {
		key += ":l:" + *opts.Label
	} else {
		key += ":l:production"
	}
	return key
}

// fetchPrompt fetches a prompt from the Langfuse API
func (l *Langfuse) fetchPrompt(ctx context.Context, name string, opts *GetPromptOptions) (*model.Prompt, error) {
	// Apply fetch timeout if specified
	if opts.FetchTimeout != nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, *opts.FetchTimeout)
		defer cancel()
	}

	// Build URL path with query parameters
	path := "/api/public/v2/prompts/" + url.PathEscape(name)

	params := url.Values{}
	if opts.Version != nil {
		params.Set("version", strconv.Itoa(*opts.Version))
	} else if opts.Label != nil {
		params.Set("label", *opts.Label)
	}
	// Note: if neither version nor label is set, API defaults to "production"

	if len(params) > 0 {
		path = path + "?" + params.Encode()
	}

	// Make the API request
	body, statusCode, err := l.client.DoGetRequest(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch prompt: %w", err)
	}

	if statusCode >= 400 {
		return nil, fmt.Errorf("failed to fetch prompt: HTTP %d: %s", statusCode, string(body))
	}

	// Parse the response
	prompt, err := parsePromptResponse(body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse prompt response: %w", err)
	}

	// Cache the result
	cacheTTL := l.promptCacheTTL
	if opts.CacheTTL != nil {
		cacheTTL = *opts.CacheTTL
	}

	cacheKey := l.buildPromptCacheKey(name, opts)
	l.promptCache.SetWithTTL(cacheKey, prompt, cacheTTL)

	return prompt, nil
}

// parsePromptResponse parses the API response into a Prompt
func parsePromptResponse(body []byte) (*model.Prompt, error) {
	// First, parse the common fields to determine the type
	var rawPrompt struct {
		Type            string         `json:"type"`
		Name            string         `json:"name"`
		Version         int            `json:"version"`
		Config          any            `json:"config"`
		Labels          []string       `json:"labels"`
		Tags            []string       `json:"tags"`
		CommitMessage   *string        `json:"commitMessage,omitempty"`
		ResolutionGraph map[string]any `json:"resolutionGraph,omitempty"`
		Prompt          json.RawMessage `json:"prompt"`
	}

	if err := json.Unmarshal(body, &rawPrompt); err != nil {
		return nil, fmt.Errorf("failed to unmarshal prompt: %w", err)
	}

	result := &model.Prompt{}

	switch rawPrompt.Type {
	case "text":
		var promptStr string
		if err := json.Unmarshal(rawPrompt.Prompt, &promptStr); err != nil {
			return nil, fmt.Errorf("failed to unmarshal text prompt: %w", err)
		}

		result.TextPrompt = &model.TextPrompt{
			Name:            rawPrompt.Name,
			Version:         rawPrompt.Version,
			Config:          rawPrompt.Config,
			Labels:          rawPrompt.Labels,
			Tags:            rawPrompt.Tags,
			Prompt:          promptStr,
			Type:            model.PromptTypeText,
			CommitMessage:   rawPrompt.CommitMessage,
			ResolutionGraph: rawPrompt.ResolutionGraph,
		}

	case "chat":
		var messages []model.ChatMessage
		if err := json.Unmarshal(rawPrompt.Prompt, &messages); err != nil {
			return nil, fmt.Errorf("failed to unmarshal chat prompt: %w", err)
		}

		result.ChatPrompt = &model.ChatPrompt{
			Name:            rawPrompt.Name,
			Version:         rawPrompt.Version,
			Config:          rawPrompt.Config,
			Labels:          rawPrompt.Labels,
			Tags:            rawPrompt.Tags,
			Prompt:          messages,
			Type:            model.PromptTypeChat,
			CommitMessage:   rawPrompt.CommitMessage,
			ResolutionGraph: rawPrompt.ResolutionGraph,
		}

	default:
		return nil, fmt.Errorf("unknown prompt type: %s", rawPrompt.Type)
	}

	return result, nil
}

// ClearPromptCache clears all cached prompts
func (l *Langfuse) ClearPromptCache() {
	l.promptCache.Clear()
}

func ingest(ctx context.Context, client *api.Client, events []model.IngestionEvent) error {
	req := api.Ingestion{
		Batch: events,
	}

	res := api.IngestionResponse{}
	return client.Ingestion(ctx, &req, &res)
}

func (l *Langfuse) Trace(t *model.Trace) (*model.Trace, error) {
	t.ID = buildID(&t.ID)
	l.observer.Dispatch(
		model.IngestionEvent{
			ID:        buildID(nil),
			Type:      model.IngestionEventTypeTraceCreate,
			Timestamp: time.Now().UTC(),
			Body:      t,
		},
	)
	return t, nil
}

func (l *Langfuse) Generation(g *model.Generation, parentID *string) (*model.Generation, error) {
	if g.TraceID == "" {
		traceID, err := l.createTrace(g.Name)
		if err != nil {
			return nil, err
		}

		g.TraceID = traceID
	}

	g.ID = buildID(&g.ID)

	if parentID != nil {
		g.ParentObservationID = *parentID
	}

	l.observer.Dispatch(
		model.IngestionEvent{
			ID:        buildID(nil),
			Type:      model.IngestionEventTypeGenerationCreate,
			Timestamp: time.Now().UTC(),
			Body:      g,
		},
	)
	return g, nil
}

func (l *Langfuse) GenerationEnd(g *model.Generation) (*model.Generation, error) {
	if g.ID == "" {
		return nil, fmt.Errorf("generation ID is required")
	}

	if g.TraceID == "" {
		return nil, fmt.Errorf("trace ID is required")
	}

	l.observer.Dispatch(
		model.IngestionEvent{
			ID:        buildID(nil),
			Type:      model.IngestionEventTypeGenerationUpdate,
			Timestamp: time.Now().UTC(),
			Body:      g,
		},
	)

	return g, nil
}

func (l *Langfuse) Score(s *model.Score) (*model.Score, error) {
	if s.TraceID == "" && s.SessionID == "" {
		return nil, fmt.Errorf("either trace ID or session ID is required")
	}
	s.ID = buildID(&s.ID)

	l.observer.Dispatch(
		model.IngestionEvent{
			ID:        buildID(nil),
			Type:      model.IngestionEventTypeScoreCreate,
			Timestamp: time.Now().UTC(),
			Body:      s,
		},
	)
	return s, nil
}

func (l *Langfuse) Span(s *model.Span, parentID *string) (*model.Span, error) {
	if s.TraceID == "" {
		traceID, err := l.createTrace(s.Name)
		if err != nil {
			return nil, err
		}

		s.TraceID = traceID
	}

	s.ID = buildID(&s.ID)

	if parentID != nil {
		s.ParentObservationID = *parentID
	}

	l.observer.Dispatch(
		model.IngestionEvent{
			ID:        buildID(nil),
			Type:      model.IngestionEventTypeSpanCreate,
			Timestamp: time.Now().UTC(),
			Body:      s,
		},
	)

	return s, nil
}

func (l *Langfuse) SpanEnd(s *model.Span) (*model.Span, error) {
	if s.ID == "" {
		return nil, fmt.Errorf("generation ID is required")
	}

	if s.TraceID == "" {
		return nil, fmt.Errorf("trace ID is required")
	}

	l.observer.Dispatch(
		model.IngestionEvent{
			ID:        buildID(nil),
			Type:      model.IngestionEventTypeSpanUpdate,
			Timestamp: time.Now().UTC(),
			Body:      s,
		},
	)

	return s, nil
}

func (l *Langfuse) Event(e *model.Event, parentID *string) (*model.Event, error) {
	if e.TraceID == "" {
		traceID, err := l.createTrace(e.Name)
		if err != nil {
			return nil, err
		}

		e.TraceID = traceID
	}

	e.ID = buildID(&e.ID)

	if parentID != nil {
		e.ParentObservationID = *parentID
	}

	l.observer.Dispatch(
		model.IngestionEvent{
			ID:        uuid.New().String(),
			Type:      model.IngestionEventTypeEventCreate,
			Timestamp: time.Now().UTC(),
			Body:      e,
		},
	)

	return e, nil
}

func (l *Langfuse) createTrace(traceName string) (string, error) {
	trace, errTrace := l.Trace(
		&model.Trace{
			Name: traceName,
		},
	)
	if errTrace != nil {
		return "", errTrace
	}

	return trace.ID, nil
}

func (l *Langfuse) Flush(ctx context.Context) {
	l.observer.Wait(ctx)
}

func buildID(id *string) string {
	if id == nil {
		return uuid.New().String()
	} else if *id == "" {
		return uuid.New().String()
	}

	return *id
}
