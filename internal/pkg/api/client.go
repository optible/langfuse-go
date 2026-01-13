package api

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/henomis/restclientgo"
)

const (
	langfuseDefaultEndpoint = "https://cloud.langfuse.com"
)

type Client struct {
	restClient *restclientgo.RestClient
	httpClient *http.Client
	host       string
	publicKey  string
	secretKey  string
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
		httpClient: &http.Client{},
		host:       langfuseHost,
		publicKey:  publicKey,
		secretKey:  secretKey,
	}
}

func (c *Client) Ingestion(ctx context.Context, req *Ingestion, res *IngestionResponse) error {
	return c.restClient.Post(ctx, req, res)
}

// GetHost returns the configured Langfuse host
func (c *Client) GetHost() string {
	return c.host
}

// doRequest performs a raw HTTP request and returns the response body
func (c *Client) doRequest(ctx context.Context, method, urlPath string) ([]byte, int, error) {
	fullURL := c.host + urlPath

	req, err := http.NewRequestWithContext(ctx, method, fullURL, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create request: %w", err)
	}

	// Apply standard headers using the client's stored credentials
	req.Header.Set("Authorization", basicAuth(c.publicKey, c.secretKey))
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("failed to read response body: %w", err)
	}

	return body, resp.StatusCode, nil
}

// DoGetRequest performs a raw GET request and returns the response body
func (c *Client) DoGetRequest(ctx context.Context, urlPath string) ([]byte, int, error) {
	return c.doRequest(ctx, http.MethodGet, urlPath)
}

// DoDeleteRequest performs a raw DELETE request and returns the response body
func (c *Client) DoDeleteRequest(ctx context.Context, urlPath string) ([]byte, int, error) {
	return c.doRequest(ctx, http.MethodDelete, urlPath)
}

func basicAuth(publicKey, secretKey string) string {
	auth := publicKey + ":" + secretKey
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
}
