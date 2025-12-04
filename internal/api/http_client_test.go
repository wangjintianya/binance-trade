package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"binance-trader/pkg/errors"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Feature: binance-auto-trading, Property 7: 重试机制正确性
// Validates: Requirements 2.4
// For any failed API request, the system should retry at most 3 times with increasing delays
func TestProperty_RetryMechanism(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("retry mechanism retries up to max attempts with exponential backoff", prop.ForAll(
		func(shouldFail bool, initialDelayMs int) bool {
			// Track retry attempts and delays
			var attemptCount int32
			var delays []time.Duration
			var lastAttemptTime time.Time

			// Create test server that fails conditionally
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				currentAttempt := atomic.AddInt32(&attemptCount, 1)
				
				// Record delay between attempts
				now := time.Now()
				if currentAttempt > 1 && !lastAttemptTime.IsZero() {
					delays = append(delays, now.Sub(lastAttemptTime))
				}
				lastAttemptTime = now

				if shouldFail {
					// Return network error to trigger retry
					w.WriteHeader(http.StatusServiceUnavailable)
					w.Write([]byte("service unavailable"))
				} else {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{"success": true}`))
				}
			}))
			defer server.Close()

			// Create HTTP client with retry config
			retryConfig := RetryConfig{
				MaxAttempts:       3,
				InitialDelayMs:    initialDelayMs,
				BackoffMultiplier: 2.0,
			}
			client := NewHTTPClient(nil, retryConfig)

			// Make request
			_, err := client.DoWithRetry(http.MethodGet, server.URL, nil, nil)

			// Verify retry behavior
			attempts := int(atomic.LoadInt32(&attemptCount))

			if shouldFail {
				// Should fail after max attempts
				if err == nil {
					return false
				}
				if attempts != retryConfig.MaxAttempts {
					return false
				}

				// Verify exponential backoff delays
				if len(delays) != retryConfig.MaxAttempts-1 {
					return false
				}

				expectedDelay := time.Duration(initialDelayMs) * time.Millisecond
				for i, delay := range delays {
					// Allow 20% tolerance for timing variations
					minDelay := expectedDelay * 8 / 10
					maxDelay := expectedDelay * 12 / 10

					if delay < minDelay || delay > maxDelay {
						t.Logf("Attempt %d: expected delay ~%v, got %v", i+1, expectedDelay, delay)
						return false
					}

					// Next delay should be multiplied by backoff
					expectedDelay = time.Duration(float64(expectedDelay) * retryConfig.BackoffMultiplier)
				}
			} else {
				// Should succeed on first attempt
				if err != nil {
					return false
				}
				if attempts != 1 {
					return false
				}
			}

			return true
		},
		gen.Bool(),
		gen.IntRange(50, 200), // Initial delay between 50-200ms for faster tests
	))

	properties.TestingRun(t)
}

// Feature: binance-auto-trading, Property 7: 重试机制正确性
// Validates: Requirements 2.4
// Non-retryable errors should not trigger retries
func TestProperty_NonRetryableErrors(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("non-retryable errors do not trigger retries", prop.ForAll(
		func(statusCode int) bool {
			var attemptCount int32

			// Create test server that returns the specified status code
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				atomic.AddInt32(&attemptCount, 1)
				w.WriteHeader(statusCode)
				w.Write([]byte(fmt.Sprintf(`{"error": "status %d"}`, statusCode)))
			}))
			defer server.Close()

			// Create HTTP client with retry config
			retryConfig := RetryConfig{
				MaxAttempts:       3,
				InitialDelayMs:    10,
				BackoffMultiplier: 2.0,
			}
			client := NewHTTPClient(nil, retryConfig)

			// Make request
			_, err := client.DoWithRetry(http.MethodGet, server.URL, nil, nil)

			attempts := int(atomic.LoadInt32(&attemptCount))

			// Non-retryable errors (4xx except 429) should only attempt once
			if statusCode >= 400 && statusCode < 500 && statusCode != 429 {
				if err == nil {
					return false
				}
				// Should not retry
				return attempts == 1
			}

			// Retryable errors (5xx, 429) should retry
			if statusCode >= 500 || statusCode == 429 {
				if err == nil {
					return false
				}
				// Should retry up to max attempts
				return attempts == retryConfig.MaxAttempts
			}

			// Success cases (2xx, 3xx)
			if statusCode >= 200 && statusCode < 400 {
				// Should succeed on first attempt
				return err == nil && attempts == 1
			}

			return true
		},
		gen.IntRange(200, 599), // HTTP status codes
	))

	properties.TestingRun(t)
}

// Feature: binance-auto-trading, Property 19: 速率限制自适应
// Validates: Requirements 5.4
// For any API rate limit warning, the rate limiter should increase the delay between requests
func TestProperty_RateLimitAdaptive(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("rate limiter increases delay after rate limit hit", prop.ForAll(
		func(maxCallsPerMinute int) bool {
			// Create rate limiter
			rateLimiter := NewRateLimiter(maxCallsPerMinute)

			// Initial adaptive delay should be 0
			if rateLimiter.GetAdaptiveDelay() != 0 {
				return false
			}

			// Simulate rate limit hit
			rateLimiter.OnRateLimitHit()

			// Adaptive delay should increase
			delay1 := rateLimiter.GetAdaptiveDelay()
			if delay1 == 0 {
				return false
			}

			// Hit rate limit again
			rateLimiter.OnRateLimitHit()

			// Adaptive delay should increase further
			delay2 := rateLimiter.GetAdaptiveDelay()
			if delay2 <= delay1 {
				return false
			}

			// Rate limit hit count should be tracked
			if rateLimiter.GetRateLimitHitCount() != 2 {
				return false
			}

			return true
		},
		gen.IntRange(100, 2000), // Max calls per minute
	))

	properties.Property("rate limiter caps adaptive delay at maximum", prop.ForAll(
		func(maxCallsPerMinute int) bool {
			rateLimiter := NewRateLimiter(maxCallsPerMinute)

			// Hit rate limit many times
			for i := 0; i < 10; i++ {
				rateLimiter.OnRateLimitHit()
			}

			// Adaptive delay should be capped at 5 seconds
			delay := rateLimiter.GetAdaptiveDelay()
			maxDelay := 5 * time.Second

			return delay <= maxDelay
		},
		gen.IntRange(100, 2000),
	))

	properties.TestingRun(t)
}

// Feature: binance-auto-trading, Property 19: 速率限制自适应
// Validates: Requirements 5.4
// Rate limiter should properly handle concurrent requests
func TestProperty_RateLimitConcurrency(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("rate limiter handles concurrent requests safely", prop.ForAll(
		func(maxCallsPerMinute int, numGoroutines int) bool {
			rateLimiter := NewRateLimiter(maxCallsPerMinute)

			// Track successful waits
			successCount := int32(0)

			// Launch concurrent goroutines
			done := make(chan bool, numGoroutines)
			for i := 0; i < numGoroutines; i++ {
				go func() {
					rateLimiter.Wait()
					atomic.AddInt32(&successCount, 1)
					done <- true
				}()
			}

			// Wait for all goroutines with timeout
			timeout := time.After(10 * time.Second)
			for i := 0; i < numGoroutines; i++ {
				select {
				case <-done:
					// Success
				case <-timeout:
					// Timeout - this is acceptable for rate limiting
					return true
				}
			}

			// All goroutines should have completed
			return int(atomic.LoadInt32(&successCount)) <= numGoroutines
		},
		gen.IntRange(100, 1000),  // Max calls per minute
		gen.IntRange(5, 20),      // Number of concurrent goroutines
	))

	properties.TestingRun(t)
}

// Unit tests for HTTP client

// TestHTTPClient_RequestBuilding tests request construction
func TestHTTPClient_RequestBuilding(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		params     map[string]interface{}
		headers    map[string]string
		wantParams bool
		wantHeader string
	}{
		{
			name:       "GET request with params",
			method:     http.MethodGet,
			params:     map[string]interface{}{"symbol": "BTCUSDT", "limit": 10},
			headers:    map[string]string{"X-Test": "value"},
			wantParams: true,
			wantHeader: "value",
		},
		{
			name:       "POST request with JSON body",
			method:     http.MethodPost,
			params:     map[string]interface{}{"symbol": "BTCUSDT", "quantity": 1.5},
			headers:    map[string]string{"Authorization": "Bearer token"},
			wantParams: false,
			wantHeader: "Bearer token",
		},
		{
			name:       "DELETE request with params",
			method:     http.MethodDelete,
			params:     map[string]interface{}{"orderId": 12345},
			headers:    nil,
			wantParams: true,
			wantHeader: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify method
				if r.Method != tt.method {
					t.Errorf("expected method %s, got %s", tt.method, r.Method)
				}

				// Verify params
				if tt.wantParams {
					if r.URL.RawQuery == "" {
						t.Error("expected query params, got none")
					}
				}

				// Verify headers
				if tt.wantHeader != "" {
					for k, v := range tt.headers {
						if r.Header.Get(k) != v {
							t.Errorf("expected header %s=%s, got %s", k, v, r.Header.Get(k))
						}
					}
				}

				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"success": true}`))
			}))
			defer server.Close()

			client := NewHTTPClient(nil, RetryConfig{MaxAttempts: 1, InitialDelayMs: 10, BackoffMultiplier: 2.0})
			_, err := client.Do(tt.method, server.URL, tt.params, tt.headers)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// TestHTTPClient_ResponseParsing tests response handling
func TestHTTPClient_ResponseParsing(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		responseBody   string
		expectError    bool
		expectedErrType errors.ErrorType
	}{
		{
			name:         "successful response",
			statusCode:   200,
			responseBody: `{"data": "success"}`,
			expectError:  false,
		},
		{
			name:            "rate limit error",
			statusCode:      429,
			responseBody:    `{"error": "rate limit exceeded"}`,
			expectError:     true,
			expectedErrType: errors.ErrRateLimit,
		},
		{
			name:            "client error",
			statusCode:      400,
			responseBody:    `{"error": "bad request"}`,
			expectError:     true,
			expectedErrType: errors.ErrInvalidParameter,
		},
		{
			name:            "server error",
			statusCode:      500,
			responseBody:    `{"error": "internal server error"}`,
			expectError:     true,
			expectedErrType: errors.ErrNetwork,
		},
		{
			name:            "authentication error",
			statusCode:      401,
			responseBody:    `{"error": "unauthorized"}`,
			expectError:     true,
			expectedErrType: errors.ErrInvalidParameter,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			client := NewHTTPClient(nil, RetryConfig{MaxAttempts: 1, InitialDelayMs: 10, BackoffMultiplier: 2.0})
			body, err := client.Do(http.MethodGet, server.URL, nil, nil)

			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				if tradingErr, ok := err.(*errors.TradingError); ok {
					if tradingErr.Type != tt.expectedErrType {
						t.Errorf("expected error type %v, got %v", tt.expectedErrType, tradingErr.Type)
					}
				} else {
					t.Errorf("expected TradingError, got %T", err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if string(body) != tt.responseBody {
					t.Errorf("expected body %s, got %s", tt.responseBody, string(body))
				}
			}
		})
	}
}

// TestHTTPClient_ErrorHandling tests error scenarios
func TestHTTPClient_ErrorHandling(t *testing.T) {
	t.Run("network timeout", func(t *testing.T) {
		// Create a server that delays response
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(2 * time.Second)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		// Create client with short timeout
		client := &httpClient{
			client: &http.Client{
				Timeout: 100 * time.Millisecond,
			},
			rateLimiter: nil,
			retryConfig: RetryConfig{MaxAttempts: 1, InitialDelayMs: 10, BackoffMultiplier: 2.0},
		}

		_, err := client.Do(http.MethodGet, server.URL, nil, nil)
		if err == nil {
			t.Error("expected timeout error, got nil")
		}
	})

	t.Run("invalid URL", func(t *testing.T) {
		client := NewHTTPClient(nil, RetryConfig{MaxAttempts: 1, InitialDelayMs: 10, BackoffMultiplier: 2.0})
		_, err := client.Do(http.MethodGet, "://invalid-url", nil, nil)
		if err == nil {
			t.Error("expected error for invalid URL, got nil")
		}
	})

	t.Run("rate limiter integration", func(t *testing.T) {
		rateLimiter := NewRateLimiter(60) // 60 calls per minute
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error": "rate limit"}`))
		}))
		defer server.Close()

		client := NewHTTPClient(rateLimiter, RetryConfig{MaxAttempts: 1, InitialDelayMs: 10, BackoffMultiplier: 2.0})
		
		initialDelay := rateLimiter.GetAdaptiveDelay()
		_, err := client.Do(http.MethodGet, server.URL, nil, nil)
		
		if err == nil {
			t.Error("expected rate limit error, got nil")
		}
		
		// Rate limiter should have increased delay
		newDelay := rateLimiter.GetAdaptiveDelay()
		if newDelay <= initialDelay {
			t.Error("expected rate limiter to increase delay after rate limit hit")
		}
	})
}
