package errors

import (
	"errors"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Basic unit test
func TestNewTradingError(t *testing.T) {
	cause := errors.New("underlying error")
	err := NewTradingError(ErrNetwork, "network failed", 500, cause)

	if err.Type != ErrNetwork {
		t.Errorf("Expected error type ErrNetwork, got %v", err.Type)
	}
	if err.Message != "network failed" {
		t.Errorf("Expected message 'network failed', got '%s'", err.Message)
	}
	if err.Code != 500 {
		t.Errorf("Expected code 500, got %d", err.Code)
	}
	if err.Cause != cause {
		t.Errorf("Expected cause to be set")
	}
}

func TestTradingErrorError(t *testing.T) {
	// Without cause
	err := NewTradingError(ErrNetwork, "network failed", 500, nil)
	if err.Error() != "network failed" {
		t.Errorf("Expected error string 'network failed', got '%s'", err.Error())
	}

	// With cause
	cause := errors.New("underlying error")
	err2 := NewTradingError(ErrNetwork, "network failed", 500, cause)
	expected := "network failed: underlying error"
	if err2.Error() != expected {
		t.Errorf("Expected error string '%s', got '%s'", expected, err2.Error())
	}
}

// Property-based test to verify gopter is working
// Feature: binance-auto-trading, Property 0: Error message consistency
func TestTradingErrorProperty(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("TradingError always returns non-empty error message", prop.ForAll(
		func(message string) bool {
			err := NewTradingError(ErrNetwork, message, 500, nil)
			return err.Error() == message
		},
		gen.AnyString().SuchThat(func(s string) bool { return len(s) > 0 }),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}
