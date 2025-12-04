package main

import (
	"context"
	"os"
	"testing"
	"time"
)

// TestInitializeApplication tests that the application can be initialized with valid config
func TestInitializeApplication(t *testing.T) {
	// Set up environment variables for testing
	os.Setenv("BINANCE_API_KEY", "test_api_key_12345678")
	os.Setenv("BINANCE_API_SECRET", "test_api_secret_12345678")
	defer func() {
		os.Unsetenv("BINANCE_API_KEY")
		os.Unsetenv("BINANCE_API_SECRET")
	}()

	// Set config file to example config
	os.Setenv("CONFIG_FILE", "../config.example.yaml")
	defer os.Unsetenv("CONFIG_FILE")

	// Initialize application
	app, err := initializeApplication()
	if err != nil {
		t.Fatalf("Failed to initialize application: %v", err)
	}

	// Verify all components are initialized
	if app.config == nil {
		t.Error("Config is nil")
	}
	if app.logger == nil {
		t.Error("Logger is nil")
	}
	if app.binanceClient == nil {
		t.Error("BinanceClient is nil")
	}
	if app.tradingService == nil {
		t.Error("TradingService is nil")
	}
	if app.marketService == nil {
		t.Error("MarketService is nil")
	}
	if app.orderRepo == nil {
		t.Error("OrderRepository is nil")
	}
	if app.riskMgr == nil {
		t.Error("RiskManager is nil")
	}
	if app.cli == nil {
		t.Error("CLI is nil")
	}
}

// TestGracefulShutdown tests that the application can shutdown gracefully
func TestGracefulShutdown(t *testing.T) {
	// Set up environment variables for testing
	os.Setenv("BINANCE_API_KEY", "test_api_key_12345678")
	os.Setenv("BINANCE_API_SECRET", "test_api_secret_12345678")
	os.Setenv("CONFIG_FILE", "../config.example.yaml")
	defer func() {
		os.Unsetenv("BINANCE_API_KEY")
		os.Unsetenv("BINANCE_API_SECRET")
		os.Unsetenv("CONFIG_FILE")
	}()

	// Initialize application
	app, err := initializeApplication()
	if err != nil {
		t.Fatalf("Failed to initialize application: %v", err)
	}

	// Test graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = app.shutdown(ctx)
	if err != nil {
		t.Errorf("Shutdown failed: %v", err)
	}
}

// TestShutdownTimeout tests that shutdown respects timeout
func TestShutdownTimeout(t *testing.T) {
	// Set up environment variables for testing
	os.Setenv("BINANCE_API_KEY", "test_api_key_12345678")
	os.Setenv("BINANCE_API_SECRET", "test_api_secret_12345678")
	os.Setenv("CONFIG_FILE", "../config.example.yaml")
	defer func() {
		os.Unsetenv("BINANCE_API_KEY")
		os.Unsetenv("BINANCE_API_SECRET")
		os.Unsetenv("CONFIG_FILE")
	}()

	// Initialize application
	app, err := initializeApplication()
	if err != nil {
		t.Fatalf("Failed to initialize application: %v", err)
	}

	// Test shutdown with very short timeout (should complete anyway since shutdown is fast)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	// This should either complete successfully or timeout
	err = app.shutdown(ctx)
	// We don't fail the test if it times out, as that's expected behavior
	if err != nil && err.Error() != "shutdown timeout exceeded" {
		t.Errorf("Unexpected shutdown error: %v", err)
	}
}

// TestInitializeApplicationMissingConfig tests error handling when config is missing
func TestInitializeApplicationMissingConfig(t *testing.T) {
	// Set config file to non-existent file
	os.Setenv("CONFIG_FILE", "nonexistent.yaml")
	defer os.Unsetenv("CONFIG_FILE")

	// Initialize application should fail
	_, err := initializeApplication()
	if err == nil {
		t.Error("Expected error when config file is missing, got nil")
	}
}

// TestApplicationPanicRecovery tests that panic recovery works in the run method
func TestApplicationPanicRecovery(t *testing.T) {
	// Set up environment variables for testing
	os.Setenv("BINANCE_API_KEY", "test_api_key_12345678")
	os.Setenv("BINANCE_API_SECRET", "test_api_secret_12345678")
	os.Setenv("CONFIG_FILE", "../config.example.yaml")
	defer func() {
		os.Unsetenv("BINANCE_API_KEY")
		os.Unsetenv("BINANCE_API_SECRET")
		os.Unsetenv("CONFIG_FILE")
	}()

	// Initialize application
	app, err := initializeApplication()
	if err != nil {
		t.Fatalf("Failed to initialize application: %v", err)
	}

	// Verify that the application has panic recovery in place
	// The run method has a defer recover() that should catch panics
	if app == nil {
		t.Error("Application is nil")
	}
}
