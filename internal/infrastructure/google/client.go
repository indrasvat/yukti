package google

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"golang.org/x/oauth2"
)

// BaseURL is the Google Apps Script API base URL.
const BaseURL = "https://script.googleapis.com/v1"

// Common API errors.
var (
	ErrNotFound     = errors.New("resource not found")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("access forbidden")
	ErrRateLimit    = errors.New("rate limit exceeded")
)

// APIError represents an error returned by the Google API.
type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error %d: %s", e.Code, e.Message)
}

// Client is an HTTP client wrapper for the Google Apps Script API.
type Client struct {
	httpClient *http.Client
	baseURL    string
	logger     *slog.Logger
}

// NewClient creates a new API client with the given token source.
func NewClient(ctx context.Context, tokenSource oauth2.TokenSource, logger *slog.Logger) *Client {
	return &Client{
		httpClient: oauth2.NewClient(ctx, tokenSource),
		baseURL:    BaseURL,
		logger:     logger,
	}
}

// do executes an HTTP request and handles the response.
func (c *Client) do(ctx context.Context, method, path string, body io.Reader, result any) error {
	url := c.baseURL + path

	c.logger.Debug("API request",
		slog.String("method", method),
		slog.String("url", url),
	)

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	start := time.Now()
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	duration := time.Since(start)

	c.logger.Debug("API response",
		slog.String("method", method),
		slog.String("url", url),
		slog.Int("status", resp.StatusCode),
		slog.Duration("duration", duration),
	)

	if resp.StatusCode >= 400 {
		return c.handleError(resp)
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("decoding response: %w", err)
		}
	}

	return nil
}

// handleError parses and returns an appropriate error for HTTP error responses.
func (c *Client) handleError(resp *http.Response) error {
	// Try to parse the error response
	var errResp struct {
		Error APIError `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
		// If we can't parse the error, create a generic one
		return &APIError{
			Code:    resp.StatusCode,
			Message: resp.Status,
		}
	}

	// Map common status codes to sentinel errors
	switch resp.StatusCode {
	case http.StatusNotFound:
		return fmt.Errorf("%w: %s", ErrNotFound, errResp.Error.Message)
	case http.StatusUnauthorized:
		return fmt.Errorf("%w: %s", ErrUnauthorized, errResp.Error.Message)
	case http.StatusForbidden:
		return fmt.Errorf("%w: %s", ErrForbidden, errResp.Error.Message)
	case http.StatusTooManyRequests:
		return fmt.Errorf("%w: %s", ErrRateLimit, errResp.Error.Message)
	default:
		return &errResp.Error
	}
}

// Get performs a GET request with a relative path (appended to baseURL).
func (c *Client) Get(ctx context.Context, path string, result any) error {
	return c.do(ctx, http.MethodGet, path, nil, result)
}

// GetAbsolute performs a GET request with an absolute URL.
func (c *Client) GetAbsolute(ctx context.Context, url string, result any) error {
	return c.doAbsolute(ctx, http.MethodGet, url, nil, result)
}

// doAbsolute executes an HTTP request with an absolute URL.
func (c *Client) doAbsolute(ctx context.Context, method, url string, body io.Reader, result any) error {
	c.logger.Debug("API request",
		slog.String("method", method),
		slog.String("url", url),
	)

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	start := time.Now()
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	duration := time.Since(start)

	c.logger.Debug("API response",
		slog.String("method", method),
		slog.String("url", url),
		slog.Int("status", resp.StatusCode),
		slog.Duration("duration", duration),
	)

	if resp.StatusCode >= 400 {
		return c.handleError(resp)
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("decoding response: %w", err)
		}
	}

	return nil
}

// Post performs a POST request with a JSON body.
func (c *Client) Post(ctx context.Context, path string, body, result any) error {
	b, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("encoding request: %w", err)
	}
	return c.do(ctx, http.MethodPost, path, bytes.NewReader(b), result)
}

// Put performs a PUT request with a JSON body.
func (c *Client) Put(ctx context.Context, path string, body, result any) error {
	b, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("encoding request: %w", err)
	}
	return c.do(ctx, http.MethodPut, path, bytes.NewReader(b), result)
}

// Patch performs a PATCH request with a JSON body.
func (c *Client) Patch(ctx context.Context, path string, body, result any) error {
	b, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("encoding request: %w", err)
	}
	return c.do(ctx, http.MethodPatch, path, bytes.NewReader(b), result)
}

// Delete performs a DELETE request.
func (c *Client) Delete(ctx context.Context, path string) error {
	return c.do(ctx, http.MethodDelete, path, nil, nil)
}
