package api

import (
	"sync"
	"time"
)

// RateLimiter implements token bucket algorithm for rate limiting
type RateLimiter struct {
	mu                sync.Mutex
	tokens            float64
	maxTokens         float64
	refillRate        float64 // tokens per second
	lastRefill        time.Time
	adaptiveDelay     time.Duration
	rateLimitHitCount int
}

// NewRateLimiter creates a new rate limiter
// maxCallsPerMinute: maximum number of API calls allowed per minute
func NewRateLimiter(maxCallsPerMinute int) *RateLimiter {
	maxTokens := float64(maxCallsPerMinute)
	refillRate := maxTokens / 60.0 // tokens per second

	return &RateLimiter{
		tokens:        maxTokens,
		maxTokens:     maxTokens,
		refillRate:    refillRate,
		lastRefill:    time.Now(),
		adaptiveDelay: 0,
	}
}

// Wait blocks until a token is available
func (rl *RateLimiter) Wait() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Refill tokens based on time elapsed
	rl.refill()

	// Wait until we have at least one token
	for rl.tokens < 1.0 {
		rl.mu.Unlock()
		time.Sleep(100 * time.Millisecond)
		rl.mu.Lock()
		rl.refill()
	}

	// Consume one token
	rl.tokens -= 1.0

	// Apply adaptive delay if rate limit was hit recently
	if rl.adaptiveDelay > 0 {
		rl.mu.Unlock()
		time.Sleep(rl.adaptiveDelay)
		rl.mu.Lock()
	}
}

// refill adds tokens based on elapsed time (must be called with lock held)
func (rl *RateLimiter) refill() {
	now := time.Now()
	elapsed := now.Sub(rl.lastRefill).Seconds()

	// Add tokens based on elapsed time
	tokensToAdd := elapsed * rl.refillRate
	rl.tokens += tokensToAdd

	// Cap at max tokens
	if rl.tokens > rl.maxTokens {
		rl.tokens = rl.maxTokens
	}

	rl.lastRefill = now
}

// OnRateLimitHit is called when a rate limit error is received from the API
// It increases the adaptive delay to slow down requests
func (rl *RateLimiter) OnRateLimitHit() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.rateLimitHitCount++

	// Increase adaptive delay exponentially
	if rl.adaptiveDelay == 0 {
		rl.adaptiveDelay = 100 * time.Millisecond
	} else {
		rl.adaptiveDelay = rl.adaptiveDelay * 2
	}

	// Cap at 5 seconds
	if rl.adaptiveDelay > 5*time.Second {
		rl.adaptiveDelay = 5 * time.Second
	}

	// Reduce tokens to slow down immediately
	rl.tokens = 0
}

// GetAdaptiveDelay returns the current adaptive delay
func (rl *RateLimiter) GetAdaptiveDelay() time.Duration {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	return rl.adaptiveDelay
}

// GetRateLimitHitCount returns the number of times rate limit was hit
func (rl *RateLimiter) GetRateLimitHitCount() int {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	return rl.rateLimitHitCount
}

// ResetAdaptiveDelay resets the adaptive delay (for testing)
func (rl *RateLimiter) ResetAdaptiveDelay() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.adaptiveDelay = 0
	rl.rateLimitHitCount = 0
}
