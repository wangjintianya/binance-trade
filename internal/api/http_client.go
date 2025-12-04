package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"binance-trader/pkg/errors"
)

// RetryConfig holds retry configuration
type RetryConfig struct {
	MaxAttempts       int
	InitialDelayMs    int
	BackoffMultiplier float64
}

// httpClient implements HTTPClient interface
type httpClient struct {
	client      *http.Client
	rateLimiter *RateLimiter
	retryConfig RetryConfig
}

// NewHTTPClient creates a new HTTP client with rate limiting and retry
func NewHTTPClient(rateLimiter *RateLimiter, retryConfig RetryConfig) HTTPClient {
	return &httpClient{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		rateLimiter: rateLimiter,
		retryConfig: retryConfig,
	}
}

// Do performs a single HTTP request without retry
func (c *httpClient) Do(method, urlStr string, params map[string]interface{}, headers map[string]string) ([]byte, error) {
	// Wait for rate limiter
	if c.rateLimiter != nil {
		c.rateLimiter.Wait()
	}

	// Build request
	req, err := c.buildRequest(method, urlStr, params, headers)
	if err != nil {
		return nil, errors.NewTradingError(errors.ErrInvalidParameter, "failed to build request", 0, err)
	}

	// Execute request
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, errors.NewTradingError(errors.ErrNetwork, "HTTP request failed", 0, err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.NewTradingError(errors.ErrNetwork, "failed to read response body", resp.StatusCode, err)
	}

	// Check for HTTP errors
	if resp.StatusCode >= 400 {
		// Check for rate limit error
		if resp.StatusCode == 429 {
			// Notify rate limiter about rate limit hit
			if c.rateLimiter != nil {
				c.rateLimiter.OnRateLimitHit()
			}
			return nil, errors.NewTradingError(errors.ErrRateLimit, "rate limit exceeded", resp.StatusCode, fmt.Errorf("status: %d, body: %s", resp.StatusCode, string(body)))
		}

		// 4xx errors (except 429) are client errors and should not be retried
		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			return nil, errors.NewTradingError(errors.ErrInvalidParameter, fmt.Sprintf("HTTP client error: %d", resp.StatusCode), resp.StatusCode, fmt.Errorf("body: %s", string(body)))
		}

		// 5xx errors are server errors and should be retried
		return nil, errors.NewTradingError(errors.ErrNetwork, fmt.Sprintf("HTTP server error: %d", resp.StatusCode), resp.StatusCode, fmt.Errorf("body: %s", string(body)))
	}

	return body, nil
}

// DoWithRetry performs an HTTP request with exponential backoff retry
func (c *httpClient) DoWithRetry(method, urlStr string, params map[string]interface{}, headers map[string]string) ([]byte, error) {
	var lastErr error
	delay := time.Duration(c.retryConfig.InitialDelayMs) * time.Millisecond

	for attempt := 1; attempt <= c.retryConfig.MaxAttempts; attempt++ {
		// Try the request
		body, err := c.Do(method, urlStr, params, headers)
		if err == nil {
			return body, nil
		}

		lastErr = err

		// Check if error is retryable
		if !isRetryableError(err) {
			return nil, err
		}

		// Don't sleep after the last attempt
		if attempt < c.retryConfig.MaxAttempts {
			time.Sleep(delay)
			// Exponential backoff
			delay = time.Duration(float64(delay) * c.retryConfig.BackoffMultiplier)
		}
	}

	return nil, lastErr
}

// buildRequest constructs an HTTP request
func (c *httpClient) buildRequest(method, urlStr string, params map[string]interface{}, headers map[string]string) (*http.Request, error) {
	var req *http.Request
	var err error

	if method == http.MethodGet || method == http.MethodDelete {
		// For GET/DELETE, add params to query string
		if len(params) > 0 {
			queryParams := url.Values{}
			for k, v := range params {
				queryParams.Add(k, fmt.Sprintf("%v", v))
			}
			urlStr = urlStr + "?" + queryParams.Encode()
		}
		req, err = http.NewRequest(method, urlStr, nil)
	} else {
		// For POST/PUT, send params as JSON body
		var body []byte
		if len(params) > 0 {
			body, err = json.Marshal(params)
			if err != nil {
				return nil, err
			}
		}
		req, err = http.NewRequest(method, urlStr, bytes.NewBuffer(body))
		if err == nil {
			req.Header.Set("Content-Type", "application/json")
		}
	}

	if err != nil {
		return nil, err
	}

	// Add headers
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	return req, nil
}

// isRetryableError checks if an error is retryable
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Check if it's a TradingError
	if tradingErr, ok := err.(*errors.TradingError); ok {
		// Retry network errors and rate limit errors
		return tradingErr.Type == errors.ErrNetwork || tradingErr.Type == errors.ErrRateLimit
	}

	return false
}
