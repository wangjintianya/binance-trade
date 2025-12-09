package main

import (
	"os"
	"path/filepath"
	"testing"

	"binance-trader/internal/config"
)

// TestSpotEntryPointInitialization verifies that spot entry only initializes spot components
// Validates: Requirements 11.1
func TestSpotEntryPointInitialization(t *testing.T) {
	// Create a temporary config file for testing
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	
	configContent := `
spot:
  api_key: test_spot_key
  api_secret: test_spot_secret
  base_url: https://api.binance.com
  testnet: true

risk:
  max_order_amount: 1000.0
  max_daily_orders: 100
  min_balance_reserve: 100.0
  max_api_calls_per_min: 1200

logging:
  level: info
  file: logs/test.log
  spot_file: logs/spot_test.log
  max_size_mb: 10
  max_backups: 3

retry:
  max_attempts: 3
  initial_delay_ms: 1000
  backoff_multiplier: 2.0

conditional_orders:
  monitoring_interval_ms: 1000
  max_active_orders: 100
  trigger_execution_timeout_ms: 5000
  enable_smart_polling: true

stop_loss:
  default_trail_percent: 1.0
  min_trail_percent: 0.1
  max_trail_percent: 5.0
  update_interval_ms: 1000
`
	
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}
	
	// Set environment variable for config path
	os.Setenv("CONFIG_FILE", configPath)
	defer os.Unsetenv("CONFIG_FILE")
	
	// Initialize spot application
	app, err := initializeApplication(config.TradingTypeSpot)
	if err != nil {
		t.Fatalf("Failed to initialize spot application: %v", err)
	}
	
	// Verify trading type
	if app.tradingType != config.TradingTypeSpot {
		t.Errorf("Expected trading type %s, got %s", config.TradingTypeSpot, app.tradingType)
	}
	
	// Verify spot components are initialized
	if app.spotClient == nil {
		t.Error("Spot client should be initialized")
	}
	if app.spotTradingService == nil {
		t.Error("Spot trading service should be initialized")
	}
	if app.spotMarketService == nil {
		t.Error("Spot market service should be initialized")
	}
	if app.spotOrderRepo == nil {
		t.Error("Spot order repository should be initialized")
	}
	if app.spotRiskMgr == nil {
		t.Error("Spot risk manager should be initialized")
	}
	if app.spotConditionalOrderSvc == nil {
		t.Error("Spot conditional order service should be initialized")
	}
	if app.spotStopLossSvc == nil {
		t.Error("Spot stop loss service should be initialized")
	}
	if app.cli == nil {
		t.Error("CLI should be initialized for spot")
	}
	
	// Verify futures components are NOT initialized
	if app.futuresClient != nil {
		t.Error("Futures client should NOT be initialized for spot entry")
	}
	if app.futuresTradingService != nil {
		t.Error("Futures trading service should NOT be initialized for spot entry")
	}
	if app.futuresMarketService != nil {
		t.Error("Futures market service should NOT be initialized for spot entry")
	}
	if app.futuresPositionManager != nil {
		t.Error("Futures position manager should NOT be initialized for spot entry")
	}
	if app.futuresRiskManager != nil {
		t.Error("Futures risk manager should NOT be initialized for spot entry")
	}
	if app.futuresConditionalOrderSvc != nil {
		t.Error("Futures conditional order service should NOT be initialized for spot entry")
	}
	if app.futuresStopLossSvc != nil {
		t.Error("Futures stop loss service should NOT be initialized for spot entry")
	}
	if app.futuresFundingService != nil {
		t.Error("Futures funding service should NOT be initialized for spot entry")
	}
}

// TestSpotEntryWithLegacyConfig verifies spot entry works with legacy config format
func TestSpotEntryWithLegacyConfig(t *testing.T) {
	// Create a temporary config file with legacy format
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	
	configContent := `
binance:
  api_key: test_legacy_key
  api_secret: test_legacy_secret
  base_url: https://api.binance.com
  testnet: true

risk:
  max_order_amount: 1000.0
  max_daily_orders: 100
  min_balance_reserve: 100.0
  max_api_calls_per_min: 1200

logging:
  level: info
  file: logs/test.log
  max_size_mb: 10
  max_backups: 3

retry:
  max_attempts: 3
  initial_delay_ms: 1000
  backoff_multiplier: 2.0

conditional_orders:
  monitoring_interval_ms: 1000
  max_active_orders: 100
  trigger_execution_timeout_ms: 5000
  enable_smart_polling: true

stop_loss:
  default_trail_percent: 1.0
  min_trail_percent: 0.1
  max_trail_percent: 5.0
  update_interval_ms: 1000
`
	
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}
	
	// Set environment variable for config path
	os.Setenv("CONFIG_FILE", configPath)
	defer os.Unsetenv("CONFIG_FILE")
	
	// Initialize spot application with legacy config
	app, err := initializeApplication(config.TradingTypeSpot)
	if err != nil {
		t.Fatalf("Failed to initialize spot application with legacy config: %v", err)
	}
	
	// Verify spot components are initialized
	if app.spotClient == nil {
		t.Error("Spot client should be initialized with legacy config")
	}
	if app.spotTradingService == nil {
		t.Error("Spot trading service should be initialized with legacy config")
	}
}

// TestFuturesEntryPointInitialization verifies that futures entry only initializes futures components
// Validates: Requirements 11.2
func TestFuturesEntryPointInitialization(t *testing.T) {
	// Create a temporary config file for testing
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	
	configContent := `
futures:
  api_key: test_futures_key
  api_secret: test_futures_secret
  base_url: https://fapi.binance.com
  testnet: true
  default_leverage: 10
  default_margin_type: CROSSED
  dual_side_position: false
  risk:
    max_order_value: 50000.0
    max_position_value: 100000.0
    max_leverage: 20
    min_margin_ratio: 0.05
    liquidation_buffer: 0.02
    max_daily_orders: 200
    max_api_calls_per_min: 2000
  monitoring:
    position_update_interval_ms: 5000
    conditional_order_interval_ms: 1000
    funding_rate_check_interval_ms: 60000

risk:
  max_order_amount: 1000.0
  max_daily_orders: 100
  min_balance_reserve: 100.0
  max_api_calls_per_min: 1200

logging:
  level: info
  file: logs/test.log
  futures_file: logs/futures_test.log
  max_size_mb: 10
  max_backups: 3

retry:
  max_attempts: 3
  initial_delay_ms: 1000
  backoff_multiplier: 2.0

conditional_orders:
  monitoring_interval_ms: 1000
  max_active_orders: 100
  trigger_execution_timeout_ms: 5000
  enable_smart_polling: true

stop_loss:
  default_trail_percent: 1.0
  min_trail_percent: 0.1
  max_trail_percent: 5.0
  update_interval_ms: 1000
`
	
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}
	
	// Set environment variable for config path
	os.Setenv("CONFIG_FILE", configPath)
	defer os.Unsetenv("CONFIG_FILE")
	
	// Initialize futures application
	app, err := initializeApplication(config.TradingTypeFutures)
	if err != nil {
		t.Fatalf("Failed to initialize futures application: %v", err)
	}
	
	// Verify trading type
	if app.tradingType != config.TradingTypeFutures {
		t.Errorf("Expected trading type %s, got %s", config.TradingTypeFutures, app.tradingType)
	}
	
	// Verify futures components are initialized
	if app.futuresClient == nil {
		t.Error("Futures client should be initialized")
	}
	if app.futuresTradingService == nil {
		t.Error("Futures trading service should be initialized")
	}
	if app.futuresMarketService == nil {
		t.Error("Futures market service should be initialized")
	}
	if app.futuresPositionManager == nil {
		t.Error("Futures position manager should be initialized")
	}
	if app.futuresRiskManager == nil {
		t.Error("Futures risk manager should be initialized")
	}
	if app.futuresConditionalOrderSvc == nil {
		t.Error("Futures conditional order service should be initialized")
	}
	if app.futuresStopLossSvc == nil {
		t.Error("Futures stop loss service should be initialized")
	}
	if app.futuresFundingService == nil {
		t.Error("Futures funding service should be initialized")
	}
	
	// Verify spot components are NOT initialized
	if app.spotClient != nil {
		t.Error("Spot client should NOT be initialized for futures entry")
	}
	if app.spotTradingService != nil {
		t.Error("Spot trading service should NOT be initialized for futures entry")
	}
	if app.spotMarketService != nil {
		t.Error("Spot market service should NOT be initialized for futures entry")
	}
	if app.spotOrderRepo != nil {
		t.Error("Spot order repository should NOT be initialized for futures entry")
	}
	if app.spotRiskMgr != nil {
		t.Error("Spot risk manager should NOT be initialized for futures entry")
	}
	if app.spotConditionalOrderSvc != nil {
		t.Error("Spot conditional order service should NOT be initialized for futures entry")
	}
	if app.spotStopLossSvc != nil {
		t.Error("Spot stop loss service should NOT be initialized for futures entry")
	}
}
