package sdk

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/bytedance/sonic"
)

// defaultBaseURL is the production URL of the Notification Engine API.
const defaultBaseURL = "https://localhost:8080/api/v1"

// Option configures a Client with custom settings.
// Use WithTimeout, WithHTTPClient, or other option functions.
type Option func(*Client)

// Client is the entry point for the Notification Engine SDK.
// It manages authentication, HTTP transport, and sub-clients
// for interacting with different API resources.
//
// Create a client with NewClient:
//
//	client := sdk.NewClient("your-api-key")
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client

	// Notifications provides methods for triggering notification events.
	Notifications *NotificationClient

	// Contacts provides methods for managing subscriber identities and preferences.
	Contacts *ContactClient
}

// NotificationClient provides methods for sending and managing notifications.
// Access it via client.Notifications.
type NotificationClient struct {
	client *Client
}

// ContactClient provides methods for identifying subscribers
// and managing their notification preferences.
// Access it via client.Contacts.
type ContactClient struct {
	client *Client
}

// NewClient creates a new SDK client authenticated with the given API key.
// The client connects to the production API by default.
// Use Option functions like WithTimeout to customize the client.
//
// Example:
//
//	client := sdk.NewClient("your-api-key")
//	err := client.Notifications.Trigger(ctx, &sdk.TriggerRequest{...})
func NewClient(key string, opts ...Option) *Client {
	ct := &Client{
		apiKey:     key,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    defaultBaseURL,
	}

	for _, opt := range opts {
		opt(ct)
	}

	ct.Notifications = &NotificationClient{client: ct}
	ct.Contacts = &ContactClient{client: ct}
	return ct
}

// WithTimeout returns an Option that sets a custom timeout for HTTP requests.
func WithTimeout(d time.Duration) Option {
	return func(c *Client) {
		c.httpClient.Timeout = d
	}
}

// WithBaseURL returns an Option that overrides the default API base URL.
// Use this to connect to a self-hosted Notification Engine instance.
func WithBaseURL(url string) Option {
	return func(c *Client) {
		c.baseURL = url
	}
}

// WithHTTPClient returns an Option that uses a custom HTTP client for all requests.
// This allows advanced configuration like custom transport, TLS settings, or proxies.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// doRequest executes an HTTP request against the Notification Engine API.
// It handles JSON marshaling of the request body, setting auth headers,
// reading the response, and unmarshaling into the res target if provided.
// Returns an APIError for non-2xx responses, or a standard error for network failures.
func (c *Client) doRequest(ctx context.Context, method string, url string, body any, res any) error {
	var buf io.Reader

	if body != nil {
		jsonBody, err := sonic.Marshal(body)
		if err != nil {
			return err
		}
		buf = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, buf)
	if err != nil {
		return err
	}
	req.Header.Set("X-API-Key", c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "notification-engine-sdk/v0.1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return &APIError{
			Code:    resp.StatusCode,
			Message: string(bodyBytes),
		}
	}

	if res != nil {
		return sonic.Unmarshal(bodyBytes, res)
	}

	return nil
}

