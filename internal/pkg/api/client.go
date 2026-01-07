package api

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/henomis/restclientgo"
)

const (
	langfuseDefaultEndpoint = "https://cloud.langfuse.com"
)

type Client struct {
	restClient *restclientgo.RestClient
}

func New() *Client {
	langfuseHost := os.Getenv("LANGFUSE_HOST")
	if langfuseHost == "" {
		langfuseHost = langfuseDefaultEndpoint
	}

	publicKey := os.Getenv("LANGFUSE_PUBLIC_KEY")
	secretKey := os.Getenv("LANGFUSE_SECRET_KEY")

	restClient := restclientgo.New(langfuseHost)
	restClient.SetRequestModifier(func(req *http.Request) *http.Request {
		req.Header.Set("Authorization", basicAuth(publicKey, secretKey))
		return req
	})

	return &Client{
		restClient: restClient,
	}
}

func (c *Client) Ingestion(ctx context.Context, req *Ingestion, res *IngestionResponse) error {
	return c.restClient.Post(ctx, req, res)
}

// GetPrompt fetches a prompt by name with optional version or label
func (c *Client) GetPrompt(ctx context.Context, req *GetPromptRequest, res *PromptResponse) error {
	// Build URL with query parameters
	path := "/api/public/v2/prompts/" + url.PathEscape(req.Name)

	params := url.Values{}
	if req.Version != nil {
		params.Set("version", strconv.Itoa(*req.Version))
	}
	if req.Label != nil {
		params.Set("label", *req.Label)
	}

	if len(params) > 0 {
		path = path + "?" + params.Encode()
	}

	// Create a simple GET request wrapper
	getReq := &simpleGetRequest{path: path}
	return c.restClient.Get(ctx, getReq, res)
}

// simpleGetRequest is a helper for GET requests with custom paths
type simpleGetRequest struct {
	path string
}

func (r *simpleGetRequest) Path() (string, error) {
	return r.path, nil
}

func (r *simpleGetRequest) Encode() (io.Reader, error) {
	return nil, nil
}

func (r *simpleGetRequest) ContentType() string {
	return ""
}

// GetHost returns the configured Langfuse host
func (c *Client) GetHost() string {
	host := os.Getenv("LANGFUSE_HOST")
	if host == "" {
		return langfuseDefaultEndpoint
	}
	return host
}

// DoGetRequest performs a raw GET request and returns the response body
func (c *Client) DoGetRequest(ctx context.Context, urlPath string) ([]byte, int, error) {
	fullURL := c.GetHost() + urlPath

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create request: %w", err)
	}

	publicKey := os.Getenv("LANGFUSE_PUBLIC_KEY")
	secretKey := os.Getenv("LANGFUSE_SECRET_KEY")
	req.Header.Set("Authorization", basicAuth(publicKey, secretKey))
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body := make([]byte, 0)
	buf := make([]byte, 1024)
	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			body = append(body, buf[:n]...)
		}
		if readErr != nil {
			break
		}
	}

	return body, resp.StatusCode, nil
}

func basicAuth(publicKey, secretKey string) string {
	auth := publicKey + ":" + secretKey
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
}
