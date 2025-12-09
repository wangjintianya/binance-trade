package config

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// TestLoadConfig tests loading a valid configuration file
func TestLoadConfig(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `binance:
  api_key: test_api_key
  api_secret: test_api_secret
  base_url: https://api.binance.com
  testnet: false

risk:
  max_order_amount: 10000.0
  max_daily_orders: 100
  min_balance_reserve: 100.0
  max_api_calls_per_min: 1000

logging:
  level: info
  file: logs/trading.log
  max_size_mb: 100
  max_backups: 5

retry:
  max_attempts: 3
  initial_delay_ms: 1000
  backoff_multiplier: 2.0

conditional_orders:
  monitoring_interval_ms: 1000
  max_active_orders: 500
  trigger_execution_timeout_ms: 3000
  enable_smart_polling: true

stop_loss:
  default_trail_percent: 2.0
  min_trail_percent: 0.1
  max_trail_percent: 10.0
  update_interval_ms: 500
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	cm := NewConfigManager()
	config, err := cm.Load(configPath)

	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify Binance config
	if config.Binance.APIKey != "test_api_key" {
		t.Errorf("Expected APIKey 'test_api_key', got '%s'", config.Binance.APIKey)
	}
	if config.Binance.APISecret != "test_api_secret" {
		t.Errorf("Expected APISecret 'test_api_secret', got '%s'", config.Binance.APISecret)
	}
	if config.Binance.BaseURL != "https://api.binance.com" {
		t.Errorf("Expected BaseURL 'https://api.binance.com', got '%s'", config.Binance.BaseURL)
	}

	// Verify Risk config
	if config.Risk.MaxOrderAmount != 10000.0 {
		t.Errorf("Expected MaxOrderAmount 10000.0, got %f", config.Risk.MaxOrderAmount)
	}
	if config.Risk.MaxDailyOrders != 100 {
		t.Errorf("Expected MaxDailyOrders 100, got %d", config.Risk.MaxDailyOrders)
	}

	// Verify Logging config
	if config.Logging.Level != "info" {
		t.Errorf("Expected Level 'info', got '%s'", config.Logging.Level)
	}

	// Verify Retry config
	if config.Retry.MaxAttempts != 3 {
		t.Errorf("Expected MaxAttempts 3, got %d", config.Retry.MaxAttempts)
	}
}

// TestEnvVarReplacement tests environment variable substitution
func TestEnvVarReplacement(t *testing.T) {
	// Set environment variables
	os.Setenv("TEST_API_KEY", "env_api_key")
	os.Setenv("TEST_API_SECRET", "env_api_secret")
	defer func() {
		os.Unsetenv("TEST_API_KEY")
		os.Unsetenv("TEST_API_SECRET")
	}()

	// Create a temporary config file with env vars
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `binance:
  api_key: ${TEST_API_KEY}
  api_secret: ${TEST_API_SECRET}
  base_url: https://api.binance.com
  testnet: false

risk:
  max_order_amount: 10000.0
  max_daily_orders: 100
  min_balance_reserve: 100.0
  max_api_calls_per_min: 1000

logging:
  level: info
  file: logs/trading.log
  max_size_mb: 100
  max_backups: 5

retry:
  max_attempts: 3
  initial_delay_ms: 1000
  backoff_multiplier: 2.0

conditional_orders:
  monitoring_interval_ms: 1000
  max_active_orders: 500
  trigger_execution_timeout_ms: 3000
  enable_smart_polling: true

stop_loss:
  default_trail_percent: 2.0
  min_trail_percent: 0.1
  max_trail_percent: 10.0
  update_interval_ms: 500
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	cm := NewConfigManager()
	config, err := cm.Load(configPath)

	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify environment variables were replaced
	if config.Binance.APIKey != "env_api_key" {
		t.Errorf("Expected APIKey 'env_api_key', got '%s'", config.Binance.APIKey)
	}
	if config.Binance.APISecret != "env_api_secret" {
		t.Errorf("Expected APISecret 'env_api_secret', got '%s'", config.Binance.APISecret)
	}
}

// TestValidateConfig tests configuration validation
func TestValidateConfig(t *testing.T) {
	cm := NewConfigManager()

	tests := []struct {
		name        string
		config      *Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid config",
			config: &Config{
				Binance: BinanceConfig{
					APIKey:    "test_key",
					APISecret: "test_secret",
					BaseURL:   "https://api.binance.com",
					Testnet:   false,
				},
				Risk: RiskConfig{
					MaxOrderAmount:    10000.0,
					MaxDailyOrders:    100,
					MinBalanceReserve: 100.0,
					MaxAPICallsPerMin: 1000,
				},
				Logging: LoggingConfig{
					Level:      "info",
					File:       "logs/trading.log",
					MaxSizeMB:  100,
					MaxBackups: 5,
				},
				Retry: RetryConfig{
					MaxAttempts:       3,
					InitialDelayMs:    1000,
					BackoffMultiplier: 2.0,
				},
				ConditionalOrders: ConditionalOrdersConfig{
					MonitoringIntervalMs:      1000,
					MaxActiveOrders:           500,
					TriggerExecutionTimeoutMs: 3000,
					EnableSmartPolling:        true,
				},
				StopLoss: StopLossConfig{
					DefaultTrailPercent: 2.0,
					MinTrailPercent:     0.1,
					MaxTrailPercent:     10.0,
					UpdateIntervalMs:    500,
				},
			},
			expectError: false,
		},
		{
			name:        "nil config",
			config:      nil,
			expectError: true,
			errorMsg:    "config cannot be nil",
		},
		{
			name: "missing api key",
			config: &Config{
				Binance: BinanceConfig{
					APIKey:    "",
					APISecret: "test_secret",
					BaseURL:   "https://api.binance.com",
				},
				Risk: RiskConfig{
					MaxOrderAmount:    10000.0,
					MaxDailyOrders:    100,
					MinBalanceReserve: 100.0,
					MaxAPICallsPerMin: 1000,
				},
				Logging: LoggingConfig{
					Level:      "info",
					File:       "logs/trading.log",
					MaxSizeMB:  100,
					MaxBackups: 5,
				},
				Retry: RetryConfig{
					MaxAttempts:       3,
					InitialDelayMs:    1000,
					BackoffMultiplier: 2.0,
				},
				ConditionalOrders: ConditionalOrdersConfig{
					MonitoringIntervalMs:      1000,
					MaxActiveOrders:           500,
					TriggerExecutionTimeoutMs: 3000,
					EnableSmartPolling:        true,
				},
				StopLoss: StopLossConfig{
					DefaultTrailPercent: 2.0,
					MinTrailPercent:     0.1,
					MaxTrailPercent:     10.0,
					UpdateIntervalMs:    500,
				},
			},
			expectError: true,
			errorMsg:    "at least one trading configuration (binance, spot, or futures) is required",
		},
		{
			name: "non-https base url",
			config: &Config{
				Binance: BinanceConfig{
					APIKey:    "test_key",
					APISecret: "test_secret",
					BaseURL:   "http://api.binance.com",
				},
				Risk: RiskConfig{
					MaxOrderAmount:    10000.0,
					MaxDailyOrders:    100,
					MinBalanceReserve: 100.0,
					MaxAPICallsPerMin: 1000,
				},
				Logging: LoggingConfig{
					Level:      "info",
					File:       "logs/trading.log",
					MaxSizeMB:  100,
					MaxBackups: 5,
				},
				Retry: RetryConfig{
					MaxAttempts:       3,
					InitialDelayMs:    1000,
					BackoffMultiplier: 2.0,
				},
				ConditionalOrders: ConditionalOrdersConfig{
					MonitoringIntervalMs:      1000,
					MaxActiveOrders:           500,
					TriggerExecutionTimeoutMs: 3000,
					EnableSmartPolling:        true,
				},
				StopLoss: StopLossConfig{
					DefaultTrailPercent: 2.0,
					MinTrailPercent:     0.1,
					MaxTrailPercent:     10.0,
					UpdateIntervalMs:    500,
				},
			},
			expectError: true,
			errorMsg:    "binance config: base_url must use HTTPS protocol",
		},
		{
			name: "invalid max order amount",
			config: &Config{
				Binance: BinanceConfig{
					APIKey:    "test_key",
					APISecret: "test_secret",
					BaseURL:   "https://api.binance.com",
				},
				Risk: RiskConfig{
					MaxOrderAmount:    -100.0,
					MaxDailyOrders:    100,
					MinBalanceReserve: 100.0,
					MaxAPICallsPerMin: 1000,
				},
				Logging: LoggingConfig{
					Level:      "info",
					File:       "logs/trading.log",
					MaxSizeMB:  100,
					MaxBackups: 5,
				},
				Retry: RetryConfig{
					MaxAttempts:       3,
					InitialDelayMs:    1000,
					BackoffMultiplier: 2.0,
				},
				ConditionalOrders: ConditionalOrdersConfig{
					MonitoringIntervalMs:      1000,
					MaxActiveOrders:           500,
					TriggerExecutionTimeoutMs: 3000,
					EnableSmartPolling:        true,
				},
				StopLoss: StopLossConfig{
					DefaultTrailPercent: 2.0,
					MinTrailPercent:     0.1,
					MaxTrailPercent:     10.0,
					UpdateIntervalMs:    500,
				},
			},
			expectError: true,
			errorMsg:    "risk.max_order_amount must be greater than 0",
		},
		{
			name: "invalid log level",
			config: &Config{
				Binance: BinanceConfig{
					APIKey:    "test_key",
					APISecret: "test_secret",
					BaseURL:   "https://api.binance.com",
				},
				Risk: RiskConfig{
					MaxOrderAmount:    10000.0,
					MaxDailyOrders:    100,
					MinBalanceReserve: 100.0,
					MaxAPICallsPerMin: 1000,
				},
				Logging: LoggingConfig{
					Level:      "invalid",
					File:       "logs/trading.log",
					MaxSizeMB:  100,
					MaxBackups: 5,
				},
				Retry: RetryConfig{
					MaxAttempts:       3,
					InitialDelayMs:    1000,
					BackoffMultiplier: 2.0,
				},
				ConditionalOrders: ConditionalOrdersConfig{
					MonitoringIntervalMs:      1000,
					MaxActiveOrders:           500,
					TriggerExecutionTimeoutMs: 3000,
					EnableSmartPolling:        true,
				},
				StopLoss: StopLossConfig{
					DefaultTrailPercent: 2.0,
					MinTrailPercent:     0.1,
					MaxTrailPercent:     10.0,
					UpdateIntervalMs:    500,
				},
			},
			expectError: true,
			errorMsg:    "logging.level must be one of: debug, info, warn, error",
		},
		{
			name: "invalid backoff multiplier",
			config: &Config{
				Binance: BinanceConfig{
					APIKey:    "test_key",
					APISecret: "test_secret",
					BaseURL:   "https://api.binance.com",
				},
				Risk: RiskConfig{
					MaxOrderAmount:    10000.0,
					MaxDailyOrders:    100,
					MinBalanceReserve: 100.0,
					MaxAPICallsPerMin: 1000,
				},
				Logging: LoggingConfig{
					Level:      "info",
					File:       "logs/trading.log",
					MaxSizeMB:  100,
					MaxBackups: 5,
				},
				Retry: RetryConfig{
					MaxAttempts:       3,
					InitialDelayMs:    1000,
					BackoffMultiplier: 0.5,
				},
				ConditionalOrders: ConditionalOrdersConfig{
					MonitoringIntervalMs:      1000,
					MaxActiveOrders:           500,
					TriggerExecutionTimeoutMs: 3000,
					EnableSmartPolling:        true,
				},
				StopLoss: StopLossConfig{
					DefaultTrailPercent: 2.0,
					MinTrailPercent:     0.1,
					MaxTrailPercent:     10.0,
					UpdateIntervalMs:    500,
				},
			},
			expectError: true,
			errorMsg:    "retry.backoff_multiplier must be greater than 1.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cm.Validate(tt.config)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if err.Error() != tt.errorMsg {
					t.Errorf("Expected error '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

// TestGetConfig tests retrieving the current configuration
func TestGetConfig(t *testing.T) {
	cm := NewConfigManager()

	// Initially should be nil
	if config := cm.GetConfig(); config != nil {
		t.Errorf("Expected nil config before loading, got %v", config)
	}

	// Create and load a config
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `binance:
  api_key: test_key
  api_secret: test_secret
  base_url: https://api.binance.com
  testnet: false

risk:
  max_order_amount: 10000.0
  max_daily_orders: 100
  min_balance_reserve: 100.0
  max_api_calls_per_min: 1000

logging:
  level: info
  file: logs/trading.log
  max_size_mb: 100
  max_backups: 5

retry:
  max_attempts: 3
  initial_delay_ms: 1000
  backoff_multiplier: 2.0

conditional_orders:
  monitoring_interval_ms: 1000
  max_active_orders: 500
  trigger_execution_timeout_ms: 3000
  enable_smart_polling: true

stop_loss:
  default_trail_percent: 2.0
  min_trail_percent: 0.1
  max_trail_percent: 10.0
  update_interval_ms: 500
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	_, err := cm.Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Now should return the loaded config
	config := cm.GetConfig()
	if config == nil {
		t.Errorf("Expected config after loading, got nil")
	}
	if config.Binance.APIKey != "test_key" {
		t.Errorf("Expected APIKey 'test_key', got '%s'", config.Binance.APIKey)
	}
}

// TestLoadConfigFileNotFound tests loading a non-existent config file
func TestLoadConfigFileNotFound(t *testing.T) {
	cm := NewConfigManager()
	_, err := cm.Load("nonexistent.yaml")

	if err == nil {
		t.Errorf("Expected error for non-existent file, got none")
	}
}

// TestLoadConfigInvalidYAML tests loading an invalid YAML file
func TestLoadConfigInvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	invalidYAML := `binance:
  api_key: test_key
  invalid yaml content [[[
`

	if err := os.WriteFile(configPath, []byte(invalidYAML), 0644); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	cm := NewConfigManager()
	_, err := cm.Load(configPath)

	if err == nil {
		t.Errorf("Expected error for invalid YAML, got none")
	}
}

// TestReplaceEnvVars tests the environment variable replacement function
func TestReplaceEnvVars(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		envVars  map[string]string
		expected string
	}{
		{
			name:  "single env var",
			input: "api_key: ${API_KEY}",
			envVars: map[string]string{
				"API_KEY": "test_key",
			},
			expected: "api_key: test_key",
		},
		{
			name:  "multiple env vars",
			input: "api_key: ${API_KEY}\napi_secret: ${API_SECRET}",
			envVars: map[string]string{
				"API_KEY":    "test_key",
				"API_SECRET": "test_secret",
			},
			expected: "api_key: test_key\napi_secret: test_secret",
		},
		{
			name:     "no env vars",
			input:    "api_key: hardcoded_key",
			envVars:  map[string]string{},
			expected: "api_key: hardcoded_key",
		},
		{
			name:  "missing env var",
			input: "api_key: ${MISSING_KEY}",
			envVars: map[string]string{
				"OTHER_KEY": "value",
			},
			expected: "api_key: ${MISSING_KEY}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}
			defer func() {
				for key := range tt.envVars {
					os.Unsetenv(key)
				}
			}()

			result := replaceEnvVars(tt.input)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

// TestLoadConfigWithConditionalOrders tests loading config with conditional orders settings
func TestLoadConfigWithConditionalOrders(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `binance:
  api_key: test_api_key
  api_secret: test_api_secret
  base_url: https://api.binance.com
  testnet: false

risk:
  max_order_amount: 10000.0
  max_daily_orders: 100
  min_balance_reserve: 100.0
  max_api_calls_per_min: 1000

logging:
  level: info
  file: logs/trading.log
  max_size_mb: 100
  max_backups: 5

retry:
  max_attempts: 3
  initial_delay_ms: 1000
  backoff_multiplier: 2.0

conditional_orders:
  monitoring_interval_ms: 1000
  max_active_orders: 500
  trigger_execution_timeout_ms: 3000
  enable_smart_polling: true

stop_loss:
  default_trail_percent: 2.0
  min_trail_percent: 0.1
  max_trail_percent: 10.0
  update_interval_ms: 500
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	cm := NewConfigManager()
	config, err := cm.Load(configPath)

	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify ConditionalOrders config
	if config.ConditionalOrders.MonitoringIntervalMs != 1000 {
		t.Errorf("Expected MonitoringIntervalMs 1000, got %d", config.ConditionalOrders.MonitoringIntervalMs)
	}
	if config.ConditionalOrders.MaxActiveOrders != 500 {
		t.Errorf("Expected MaxActiveOrders 500, got %d", config.ConditionalOrders.MaxActiveOrders)
	}
	if config.ConditionalOrders.TriggerExecutionTimeoutMs != 3000 {
		t.Errorf("Expected TriggerExecutionTimeoutMs 3000, got %d", config.ConditionalOrders.TriggerExecutionTimeoutMs)
	}
	if !config.ConditionalOrders.EnableSmartPolling {
		t.Errorf("Expected EnableSmartPolling true, got false")
	}

	// Verify StopLoss config
	if config.StopLoss.DefaultTrailPercent != 2.0 {
		t.Errorf("Expected DefaultTrailPercent 2.0, got %f", config.StopLoss.DefaultTrailPercent)
	}
	if config.StopLoss.MinTrailPercent != 0.1 {
		t.Errorf("Expected MinTrailPercent 0.1, got %f", config.StopLoss.MinTrailPercent)
	}
	if config.StopLoss.MaxTrailPercent != 10.0 {
		t.Errorf("Expected MaxTrailPercent 10.0, got %f", config.StopLoss.MaxTrailPercent)
	}
	if config.StopLoss.UpdateIntervalMs != 500 {
		t.Errorf("Expected UpdateIntervalMs 500, got %d", config.StopLoss.UpdateIntervalMs)
	}
}

// TestConditionalOrdersConfigDefaults tests default values for conditional orders config
func TestConditionalOrdersConfigDefaults(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Config without conditional_orders section - should use zero values
	configContent := `binance:
  api_key: test_api_key
  api_secret: test_api_secret
  base_url: https://api.binance.com
  testnet: false

risk:
  max_order_amount: 10000.0
  max_daily_orders: 100
  min_balance_reserve: 100.0
  max_api_calls_per_min: 1000

logging:
  level: info
  file: logs/trading.log
  max_size_mb: 100
  max_backups: 5

retry:
  max_attempts: 3
  initial_delay_ms: 1000
  backoff_multiplier: 2.0

conditional_orders:
  monitoring_interval_ms: 0
  max_active_orders: 0
  trigger_execution_timeout_ms: 0
  enable_smart_polling: false

stop_loss:
  default_trail_percent: 0
  min_trail_percent: 0
  max_trail_percent: 0
  update_interval_ms: 0
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	cm := NewConfigManager()
	_, err := cm.Load(configPath)

	// Should fail validation due to zero values
	if err == nil {
		t.Errorf("Expected validation error for zero values, got none")
	}
}

// TestValidateConditionalOrdersConfig tests validation of conditional orders configuration
func TestValidateConditionalOrdersConfig(t *testing.T) {
	cm := NewConfigManager()

	tests := []struct {
		name        string
		config      *Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid conditional orders config",
			config: &Config{
				Binance: BinanceConfig{
					APIKey:    "test_key",
					APISecret: "test_secret",
					BaseURL:   "https://api.binance.com",
				},
				Risk: RiskConfig{
					MaxOrderAmount:    10000.0,
					MaxDailyOrders:    100,
					MinBalanceReserve: 100.0,
					MaxAPICallsPerMin: 1000,
				},
				Logging: LoggingConfig{
					Level:      "info",
					File:       "logs/trading.log",
					MaxSizeMB:  100,
					MaxBackups: 5,
				},
				Retry: RetryConfig{
					MaxAttempts:       3,
					InitialDelayMs:    1000,
					BackoffMultiplier: 2.0,
				},
				ConditionalOrders: ConditionalOrdersConfig{
					MonitoringIntervalMs:      1000,
					MaxActiveOrders:           500,
					TriggerExecutionTimeoutMs: 3000,
					EnableSmartPolling:        true,
				},
				StopLoss: StopLossConfig{
					DefaultTrailPercent: 2.0,
					MinTrailPercent:     0.1,
					MaxTrailPercent:     10.0,
					UpdateIntervalMs:    500,
				},
			},
			expectError: false,
		},
		{
			name: "invalid monitoring interval",
			config: &Config{
				Binance: BinanceConfig{
					APIKey:    "test_key",
					APISecret: "test_secret",
					BaseURL:   "https://api.binance.com",
				},
				Risk: RiskConfig{
					MaxOrderAmount:    10000.0,
					MaxDailyOrders:    100,
					MinBalanceReserve: 100.0,
					MaxAPICallsPerMin: 1000,
				},
				Logging: LoggingConfig{
					Level:      "info",
					File:       "logs/trading.log",
					MaxSizeMB:  100,
					MaxBackups: 5,
				},
				Retry: RetryConfig{
					MaxAttempts:       3,
					InitialDelayMs:    1000,
					BackoffMultiplier: 2.0,
				},
				ConditionalOrders: ConditionalOrdersConfig{
					MonitoringIntervalMs:      0,
					MaxActiveOrders:           500,
					TriggerExecutionTimeoutMs: 3000,
					EnableSmartPolling:        true,
				},
				StopLoss: StopLossConfig{
					DefaultTrailPercent: 2.0,
					MinTrailPercent:     0.1,
					MaxTrailPercent:     10.0,
					UpdateIntervalMs:    500,
				},
			},
			expectError: true,
			errorMsg:    "conditional_orders.monitoring_interval_ms must be greater than 0",
		},
		{
			name: "invalid max active orders",
			config: &Config{
				Binance: BinanceConfig{
					APIKey:    "test_key",
					APISecret: "test_secret",
					BaseURL:   "https://api.binance.com",
				},
				Risk: RiskConfig{
					MaxOrderAmount:    10000.0,
					MaxDailyOrders:    100,
					MinBalanceReserve: 100.0,
					MaxAPICallsPerMin: 1000,
				},
				Logging: LoggingConfig{
					Level:      "info",
					File:       "logs/trading.log",
					MaxSizeMB:  100,
					MaxBackups: 5,
				},
				Retry: RetryConfig{
					MaxAttempts:       3,
					InitialDelayMs:    1000,
					BackoffMultiplier: 2.0,
				},
				ConditionalOrders: ConditionalOrdersConfig{
					MonitoringIntervalMs:      1000,
					MaxActiveOrders:           -1,
					TriggerExecutionTimeoutMs: 3000,
					EnableSmartPolling:        true,
				},
				StopLoss: StopLossConfig{
					DefaultTrailPercent: 2.0,
					MinTrailPercent:     0.1,
					MaxTrailPercent:     10.0,
					UpdateIntervalMs:    500,
				},
			},
			expectError: true,
			errorMsg:    "conditional_orders.max_active_orders must be greater than 0",
		},
		{
			name: "invalid trigger execution timeout",
			config: &Config{
				Binance: BinanceConfig{
					APIKey:    "test_key",
					APISecret: "test_secret",
					BaseURL:   "https://api.binance.com",
				},
				Risk: RiskConfig{
					MaxOrderAmount:    10000.0,
					MaxDailyOrders:    100,
					MinBalanceReserve: 100.0,
					MaxAPICallsPerMin: 1000,
				},
				Logging: LoggingConfig{
					Level:      "info",
					File:       "logs/trading.log",
					MaxSizeMB:  100,
					MaxBackups: 5,
				},
				Retry: RetryConfig{
					MaxAttempts:       3,
					InitialDelayMs:    1000,
					BackoffMultiplier: 2.0,
				},
				ConditionalOrders: ConditionalOrdersConfig{
					MonitoringIntervalMs:      1000,
					MaxActiveOrders:           500,
					TriggerExecutionTimeoutMs: 0,
					EnableSmartPolling:        true,
				},
				StopLoss: StopLossConfig{
					DefaultTrailPercent: 2.0,
					MinTrailPercent:     0.1,
					MaxTrailPercent:     10.0,
					UpdateIntervalMs:    500,
				},
			},
			expectError: true,
			errorMsg:    "conditional_orders.trigger_execution_timeout_ms must be greater than 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cm.Validate(tt.config)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if err.Error() != tt.errorMsg {
					t.Errorf("Expected error '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

// TestValidateStopLossConfig tests validation of stop loss configuration
func TestValidateStopLossConfig(t *testing.T) {
	cm := NewConfigManager()

	tests := []struct {
		name        string
		config      *Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid stop loss config",
			config: &Config{
				Binance: BinanceConfig{
					APIKey:    "test_key",
					APISecret: "test_secret",
					BaseURL:   "https://api.binance.com",
				},
				Risk: RiskConfig{
					MaxOrderAmount:    10000.0,
					MaxDailyOrders:    100,
					MinBalanceReserve: 100.0,
					MaxAPICallsPerMin: 1000,
				},
				Logging: LoggingConfig{
					Level:      "info",
					File:       "logs/trading.log",
					MaxSizeMB:  100,
					MaxBackups: 5,
				},
				Retry: RetryConfig{
					MaxAttempts:       3,
					InitialDelayMs:    1000,
					BackoffMultiplier: 2.0,
				},
				ConditionalOrders: ConditionalOrdersConfig{
					MonitoringIntervalMs:      1000,
					MaxActiveOrders:           500,
					TriggerExecutionTimeoutMs: 3000,
					EnableSmartPolling:        true,
				},
				StopLoss: StopLossConfig{
					DefaultTrailPercent: 2.0,
					MinTrailPercent:     0.1,
					MaxTrailPercent:     10.0,
					UpdateIntervalMs:    500,
				},
			},
			expectError: false,
		},
		{
			name: "invalid default trail percent",
			config: &Config{
				Binance: BinanceConfig{
					APIKey:    "test_key",
					APISecret: "test_secret",
					BaseURL:   "https://api.binance.com",
				},
				Risk: RiskConfig{
					MaxOrderAmount:    10000.0,
					MaxDailyOrders:    100,
					MinBalanceReserve: 100.0,
					MaxAPICallsPerMin: 1000,
				},
				Logging: LoggingConfig{
					Level:      "info",
					File:       "logs/trading.log",
					MaxSizeMB:  100,
					MaxBackups: 5,
				},
				Retry: RetryConfig{
					MaxAttempts:       3,
					InitialDelayMs:    1000,
					BackoffMultiplier: 2.0,
				},
				ConditionalOrders: ConditionalOrdersConfig{
					MonitoringIntervalMs:      1000,
					MaxActiveOrders:           500,
					TriggerExecutionTimeoutMs: 3000,
					EnableSmartPolling:        true,
				},
				StopLoss: StopLossConfig{
					DefaultTrailPercent: 0,
					MinTrailPercent:     0.1,
					MaxTrailPercent:     10.0,
					UpdateIntervalMs:    500,
				},
			},
			expectError: true,
			errorMsg:    "stop_loss.default_trail_percent must be greater than 0",
		},
		{
			name: "min greater than max trail percent",
			config: &Config{
				Binance: BinanceConfig{
					APIKey:    "test_key",
					APISecret: "test_secret",
					BaseURL:   "https://api.binance.com",
				},
				Risk: RiskConfig{
					MaxOrderAmount:    10000.0,
					MaxDailyOrders:    100,
					MinBalanceReserve: 100.0,
					MaxAPICallsPerMin: 1000,
				},
				Logging: LoggingConfig{
					Level:      "info",
					File:       "logs/trading.log",
					MaxSizeMB:  100,
					MaxBackups: 5,
				},
				Retry: RetryConfig{
					MaxAttempts:       3,
					InitialDelayMs:    1000,
					BackoffMultiplier: 2.0,
				},
				ConditionalOrders: ConditionalOrdersConfig{
					MonitoringIntervalMs:      1000,
					MaxActiveOrders:           500,
					TriggerExecutionTimeoutMs: 3000,
					EnableSmartPolling:        true,
				},
				StopLoss: StopLossConfig{
					DefaultTrailPercent: 5.0,
					MinTrailPercent:     10.0,
					MaxTrailPercent:     5.0,
					UpdateIntervalMs:    500,
				},
			},
			expectError: true,
			errorMsg:    "stop_loss.min_trail_percent cannot be greater than max_trail_percent",
		},
		{
			name: "default trail percent out of range",
			config: &Config{
				Binance: BinanceConfig{
					APIKey:    "test_key",
					APISecret: "test_secret",
					BaseURL:   "https://api.binance.com",
				},
				Risk: RiskConfig{
					MaxOrderAmount:    10000.0,
					MaxDailyOrders:    100,
					MinBalanceReserve: 100.0,
					MaxAPICallsPerMin: 1000,
				},
				Logging: LoggingConfig{
					Level:      "info",
					File:       "logs/trading.log",
					MaxSizeMB:  100,
					MaxBackups: 5,
				},
				Retry: RetryConfig{
					MaxAttempts:       3,
					InitialDelayMs:    1000,
					BackoffMultiplier: 2.0,
				},
				ConditionalOrders: ConditionalOrdersConfig{
					MonitoringIntervalMs:      1000,
					MaxActiveOrders:           500,
					TriggerExecutionTimeoutMs: 3000,
					EnableSmartPolling:        true,
				},
				StopLoss: StopLossConfig{
					DefaultTrailPercent: 15.0,
					MinTrailPercent:     0.1,
					MaxTrailPercent:     10.0,
					UpdateIntervalMs:    500,
				},
			},
			expectError: true,
			errorMsg:    "stop_loss.default_trail_percent must be between min_trail_percent and max_trail_percent",
		},
		{
			name: "invalid update interval",
			config: &Config{
				Binance: BinanceConfig{
					APIKey:    "test_key",
					APISecret: "test_secret",
					BaseURL:   "https://api.binance.com",
				},
				Risk: RiskConfig{
					MaxOrderAmount:    10000.0,
					MaxDailyOrders:    100,
					MinBalanceReserve: 100.0,
					MaxAPICallsPerMin: 1000,
				},
				Logging: LoggingConfig{
					Level:      "info",
					File:       "logs/trading.log",
					MaxSizeMB:  100,
					MaxBackups: 5,
				},
				Retry: RetryConfig{
					MaxAttempts:       3,
					InitialDelayMs:    1000,
					BackoffMultiplier: 2.0,
				},
				ConditionalOrders: ConditionalOrdersConfig{
					MonitoringIntervalMs:      1000,
					MaxActiveOrders:           500,
					TriggerExecutionTimeoutMs: 3000,
					EnableSmartPolling:        true,
				},
				StopLoss: StopLossConfig{
					DefaultTrailPercent: 2.0,
					MinTrailPercent:     0.1,
					MaxTrailPercent:     10.0,
					UpdateIntervalMs:    0,
				},
			},
			expectError: true,
			errorMsg:    "stop_loss.update_interval_ms must be greater than 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cm.Validate(tt.config)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if err.Error() != tt.errorMsg {
					t.Errorf("Expected error '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

// ============================================
// Property-Based Tests
// ============================================

// Feature: usdt-futures-trading, Property 42: 合约配置加载正确性
// Validates: Requirements 10.3
// For any config file containing a futures configuration section, 
// futures-specific parameters must be correctly parsed and loaded
func TestProperty42_FuturesConfigLoadingCorrectness(t *testing.T) {
	tests := []struct {
		name           string
		futuresConfig  FuturesConfig
		expectValid    bool
	}{
		{
			name: "valid futures config with all fields",
			futuresConfig: FuturesConfig{
				APIKey:            "test_futures_key",
				APISecret:         "test_futures_secret",
				BaseURL:           "https://fapi.binance.com",
				Testnet:           false,
				DefaultLeverage:   10,
				DefaultMarginType: "CROSSED",
				DualSidePosition:  false,
				Risk: FuturesRiskConfig{
					MaxOrderValue:     50000.0,
					MaxPositionValue:  100000.0,
					MaxLeverage:       20,
					MinMarginRatio:    0.05,
					LiquidationBuffer: 0.02,
					MaxDailyOrders:    200,
					MaxAPICallsPerMin: 2000,
				},
				Monitoring: FuturesMonitoringConfig{
					PositionUpdateIntervalMs:   5000,
					ConditionalOrderIntervalMs: 1000,
					FundingRateCheckIntervalMs: 60000,
				},
				StopLoss: FuturesStopLossConfig{
					DefaultCallbackRate: 1.0,
					MinCallbackRate:     0.1,
					MaxCallbackRate:     5.0,
				},
			},
			expectValid: true,
		},
		{
			name: "valid futures config with minimum leverage",
			futuresConfig: FuturesConfig{
				APIKey:            "test_key",
				APISecret:         "test_secret",
				BaseURL:           "https://fapi.binance.com",
				Testnet:           false,
				DefaultLeverage:   1,
				DefaultMarginType: "ISOLATED",
				DualSidePosition:  true,
				Risk: FuturesRiskConfig{
					MaxOrderValue:     10000.0,
					MaxPositionValue:  20000.0,
					MaxLeverage:       1,
					MinMarginRatio:    0.1,
					LiquidationBuffer: 0.05,
					MaxDailyOrders:    50,
					MaxAPICallsPerMin: 1000,
				},
				Monitoring: FuturesMonitoringConfig{
					PositionUpdateIntervalMs:   1000,
					ConditionalOrderIntervalMs: 500,
					FundingRateCheckIntervalMs: 30000,
				},
				StopLoss: FuturesStopLossConfig{
					DefaultCallbackRate: 0.5,
					MinCallbackRate:     0.1,
					MaxCallbackRate:     2.0,
				},
			},
			expectValid: true,
		},
		{
			name: "valid futures config with maximum leverage",
			futuresConfig: FuturesConfig{
				APIKey:            "test_key",
				APISecret:         "test_secret",
				BaseURL:           "https://fapi.binance.com",
				Testnet:           true,
				DefaultLeverage:   125,
				DefaultMarginType: "CROSSED",
				DualSidePosition:  false,
				Risk: FuturesRiskConfig{
					MaxOrderValue:     100000.0,
					MaxPositionValue:  500000.0,
					MaxLeverage:       125,
					MinMarginRatio:    0.01,
					LiquidationBuffer: 0.01,
					MaxDailyOrders:    1000,
					MaxAPICallsPerMin: 2400,
				},
				Monitoring: FuturesMonitoringConfig{
					PositionUpdateIntervalMs:   10000,
					ConditionalOrderIntervalMs: 2000,
					FundingRateCheckIntervalMs: 120000,
				},
				StopLoss: FuturesStopLossConfig{
					DefaultCallbackRate: 3.0,
					MinCallbackRate:     0.5,
					MaxCallbackRate:     5.0,
				},
			},
			expectValid: true,
		},
		{
			name: "invalid futures config - leverage too low",
			futuresConfig: FuturesConfig{
				APIKey:            "test_key",
				APISecret:         "test_secret",
				BaseURL:           "https://fapi.binance.com",
				DefaultLeverage:   0,
				DefaultMarginType: "CROSSED",
				Risk: FuturesRiskConfig{
					MaxOrderValue:     50000.0,
					MaxPositionValue:  100000.0,
					MaxLeverage:       20,
					MinMarginRatio:    0.05,
					LiquidationBuffer: 0.02,
					MaxDailyOrders:    200,
					MaxAPICallsPerMin: 2000,
				},
			},
			expectValid: false,
		},
		{
			name: "invalid futures config - leverage too high",
			futuresConfig: FuturesConfig{
				APIKey:            "test_key",
				APISecret:         "test_secret",
				BaseURL:           "https://fapi.binance.com",
				DefaultLeverage:   126,
				DefaultMarginType: "CROSSED",
				Risk: FuturesRiskConfig{
					MaxOrderValue:     50000.0,
					MaxPositionValue:  100000.0,
					MaxLeverage:       20,
					MinMarginRatio:    0.05,
					LiquidationBuffer: 0.02,
					MaxDailyOrders:    200,
					MaxAPICallsPerMin: 2000,
				},
			},
			expectValid: false,
		},
		{
			name: "invalid futures config - invalid margin type",
			futuresConfig: FuturesConfig{
				APIKey:            "test_key",
				APISecret:         "test_secret",
				BaseURL:           "https://fapi.binance.com",
				DefaultLeverage:   10,
				DefaultMarginType: "INVALID",
				Risk: FuturesRiskConfig{
					MaxOrderValue:     50000.0,
					MaxPositionValue:  100000.0,
					MaxLeverage:       20,
					MinMarginRatio:    0.05,
					LiquidationBuffer: 0.02,
					MaxDailyOrders:    200,
					MaxAPICallsPerMin: 2000,
				},
			},
			expectValid: false,
		},
		{
			name: "invalid futures config - non-https URL",
			futuresConfig: FuturesConfig{
				APIKey:            "test_key",
				APISecret:         "test_secret",
				BaseURL:           "http://fapi.binance.com",
				DefaultLeverage:   10,
				DefaultMarginType: "CROSSED",
				Risk: FuturesRiskConfig{
					MaxOrderValue:     50000.0,
					MaxPositionValue:  100000.0,
					MaxLeverage:       20,
					MinMarginRatio:    0.05,
					LiquidationBuffer: 0.02,
					MaxDailyOrders:    200,
					MaxAPICallsPerMin: 2000,
				},
			},
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary config file with futures section
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.yaml")

			// Build YAML content
			configContent := fmt.Sprintf(`futures:
  api_key: %s
  api_secret: %s
  base_url: %s
  testnet: %t
  default_leverage: %d
  default_margin_type: %s
  dual_side_position: %t
  risk:
    max_order_value: %f
    max_position_value: %f
    max_leverage: %d
    min_margin_ratio: %f
    liquidation_buffer: %f
    max_daily_orders: %d
    max_api_calls_per_min: %d
  monitoring:
    position_update_interval_ms: %d
    conditional_order_interval_ms: %d
    funding_rate_check_interval_ms: %d
  stop_loss:
    default_callback_rate: %f
    min_callback_rate: %f
    max_callback_rate: %f

risk:
  max_order_amount: 10000.0
  max_daily_orders: 100
  min_balance_reserve: 100.0
  max_api_calls_per_min: 1000

logging:
  level: info
  file: logs/trading.log
  max_size_mb: 100
  max_backups: 5

retry:
  max_attempts: 3
  initial_delay_ms: 1000
  backoff_multiplier: 2.0

conditional_orders:
  monitoring_interval_ms: 1000
  max_active_orders: 500
  trigger_execution_timeout_ms: 3000
  enable_smart_polling: true

stop_loss:
  default_trail_percent: 2.0
  min_trail_percent: 0.1
  max_trail_percent: 10.0
  update_interval_ms: 500
`,
				tt.futuresConfig.APIKey,
				tt.futuresConfig.APISecret,
				tt.futuresConfig.BaseURL,
				tt.futuresConfig.Testnet,
				tt.futuresConfig.DefaultLeverage,
				tt.futuresConfig.DefaultMarginType,
				tt.futuresConfig.DualSidePosition,
				tt.futuresConfig.Risk.MaxOrderValue,
				tt.futuresConfig.Risk.MaxPositionValue,
				tt.futuresConfig.Risk.MaxLeverage,
				tt.futuresConfig.Risk.MinMarginRatio,
				tt.futuresConfig.Risk.LiquidationBuffer,
				tt.futuresConfig.Risk.MaxDailyOrders,
				tt.futuresConfig.Risk.MaxAPICallsPerMin,
				tt.futuresConfig.Monitoring.PositionUpdateIntervalMs,
				tt.futuresConfig.Monitoring.ConditionalOrderIntervalMs,
				tt.futuresConfig.Monitoring.FundingRateCheckIntervalMs,
				tt.futuresConfig.StopLoss.DefaultCallbackRate,
				tt.futuresConfig.StopLoss.MinCallbackRate,
				tt.futuresConfig.StopLoss.MaxCallbackRate,
			)

			if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
				t.Fatalf("Failed to create test config file: %v", err)
			}

			// Load and validate config
			cm := NewConfigManager()
			config, err := cm.Load(configPath)

			if tt.expectValid {
				if err != nil {
					t.Errorf("Expected valid config but got error: %v", err)
					return
				}

				// Verify futures config was loaded correctly
				if config.Futures == nil {
					t.Errorf("Expected futures config to be loaded, got nil")
					return
				}

				// Verify all fields were parsed correctly
				if config.Futures.APIKey != tt.futuresConfig.APIKey {
					t.Errorf("Expected APIKey '%s', got '%s'", tt.futuresConfig.APIKey, config.Futures.APIKey)
				}
				if config.Futures.APISecret != tt.futuresConfig.APISecret {
					t.Errorf("Expected APISecret '%s', got '%s'", tt.futuresConfig.APISecret, config.Futures.APISecret)
				}
				if config.Futures.BaseURL != tt.futuresConfig.BaseURL {
					t.Errorf("Expected BaseURL '%s', got '%s'", tt.futuresConfig.BaseURL, config.Futures.BaseURL)
				}
				if config.Futures.Testnet != tt.futuresConfig.Testnet {
					t.Errorf("Expected Testnet %t, got %t", tt.futuresConfig.Testnet, config.Futures.Testnet)
				}
				if config.Futures.DefaultLeverage != tt.futuresConfig.DefaultLeverage {
					t.Errorf("Expected DefaultLeverage %d, got %d", tt.futuresConfig.DefaultLeverage, config.Futures.DefaultLeverage)
				}
				if config.Futures.DefaultMarginType != tt.futuresConfig.DefaultMarginType {
					t.Errorf("Expected DefaultMarginType '%s', got '%s'", tt.futuresConfig.DefaultMarginType, config.Futures.DefaultMarginType)
				}
				if config.Futures.DualSidePosition != tt.futuresConfig.DualSidePosition {
					t.Errorf("Expected DualSidePosition %t, got %t", tt.futuresConfig.DualSidePosition, config.Futures.DualSidePosition)
				}

				// Verify risk config
				if config.Futures.Risk.MaxOrderValue != tt.futuresConfig.Risk.MaxOrderValue {
					t.Errorf("Expected MaxOrderValue %f, got %f", tt.futuresConfig.Risk.MaxOrderValue, config.Futures.Risk.MaxOrderValue)
				}
				if config.Futures.Risk.MaxLeverage != tt.futuresConfig.Risk.MaxLeverage {
					t.Errorf("Expected MaxLeverage %d, got %d", tt.futuresConfig.Risk.MaxLeverage, config.Futures.Risk.MaxLeverage)
				}

				// Verify monitoring config
				if config.Futures.Monitoring.PositionUpdateIntervalMs != tt.futuresConfig.Monitoring.PositionUpdateIntervalMs {
					t.Errorf("Expected PositionUpdateIntervalMs %d, got %d", 
						tt.futuresConfig.Monitoring.PositionUpdateIntervalMs, 
						config.Futures.Monitoring.PositionUpdateIntervalMs)
				}

				// Verify stop loss config
				if config.Futures.StopLoss.DefaultCallbackRate != tt.futuresConfig.StopLoss.DefaultCallbackRate {
					t.Errorf("Expected DefaultCallbackRate %f, got %f", 
						tt.futuresConfig.StopLoss.DefaultCallbackRate, 
						config.Futures.StopLoss.DefaultCallbackRate)
				}
			} else {
				if err == nil {
					t.Errorf("Expected error for invalid config but got none")
				}
			}
		})
	}
}

// Feature: usdt-futures-trading, Property 44: 配置段独立加载
// Validates: Requirements 11.4
// For any shared config file, spot and futures systems must be able to 
// independently load their respective configuration sections without interfering with each other
func TestProperty44_IndependentConfigSectionLoading(t *testing.T) {
	tests := []struct {
		name              string
		includeSpot       bool
		includeFutures    bool
		includeLegacy     bool
		expectSpotValid   bool
		expectFuturesValid bool
	}{
		{
			name:              "both spot and futures sections",
			includeSpot:       true,
			includeFutures:    true,
			includeLegacy:     false,
			expectSpotValid:   true,
			expectFuturesValid: true,
		},
		{
			name:              "spot section only",
			includeSpot:       true,
			includeFutures:    false,
			includeLegacy:     false,
			expectSpotValid:   true,
			expectFuturesValid: false,
		},
		{
			name:              "futures section only",
			includeSpot:       false,
			includeFutures:    true,
			includeLegacy:     false,
			expectSpotValid:   false,
			expectFuturesValid: true,
		},
		{
			name:              "legacy section only",
			includeSpot:       false,
			includeFutures:    false,
			includeLegacy:     true,
			expectSpotValid:   false,
			expectFuturesValid: false,
		},
		{
			name:              "all three sections",
			includeSpot:       true,
			includeFutures:    true,
			includeLegacy:     true,
			expectSpotValid:   true,
			expectFuturesValid: true,
		},
		{
			name:              "legacy and spot sections",
			includeSpot:       true,
			includeFutures:    false,
			includeLegacy:     true,
			expectSpotValid:   true,
			expectFuturesValid: false,
		},
		{
			name:              "legacy and futures sections",
			includeSpot:       false,
			includeFutures:    true,
			includeLegacy:     true,
			expectSpotValid:   false,
			expectFuturesValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary config file
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.yaml")

			var configContent string

			// Add legacy section if needed
			if tt.includeLegacy {
				configContent += `binance:
  api_key: legacy_api_key
  api_secret: legacy_api_secret
  base_url: https://api.binance.com
  testnet: false

`
			}

			// Add spot section if needed
			if tt.includeSpot {
				configContent += `spot:
  api_key: spot_api_key
  api_secret: spot_api_secret
  base_url: https://api.binance.com
  testnet: false

`
			}

			// Add futures section if needed
			if tt.includeFutures {
				configContent += `futures:
  api_key: futures_api_key
  api_secret: futures_api_secret
  base_url: https://fapi.binance.com
  testnet: false
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
  stop_loss:
    default_callback_rate: 1.0
    min_callback_rate: 0.1
    max_callback_rate: 5.0

`
			}

			// Add required common sections
			configContent += `risk:
  max_order_amount: 10000.0
  max_daily_orders: 100
  min_balance_reserve: 100.0
  max_api_calls_per_min: 1000

logging:
  level: info
  file: logs/trading.log
  max_size_mb: 100
  max_backups: 5

retry:
  max_attempts: 3
  initial_delay_ms: 1000
  backoff_multiplier: 2.0

conditional_orders:
  monitoring_interval_ms: 1000
  max_active_orders: 500
  trigger_execution_timeout_ms: 3000
  enable_smart_polling: true

stop_loss:
  default_trail_percent: 2.0
  min_trail_percent: 0.1
  max_trail_percent: 10.0
  update_interval_ms: 500
`

			if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
				t.Fatalf("Failed to create test config file: %v", err)
			}

			// Load config
			cm := NewConfigManager()
			config, err := cm.Load(configPath)

			if err != nil {
				t.Fatalf("Failed to load config: %v", err)
			}

			// Verify spot section
			if tt.expectSpotValid {
				if config.Spot == nil {
					t.Errorf("Expected spot config to be loaded, got nil")
				} else {
					if config.Spot.APIKey != "spot_api_key" {
						t.Errorf("Expected spot APIKey 'spot_api_key', got '%s'", config.Spot.APIKey)
					}
					if config.Spot.BaseURL != "https://api.binance.com" {
						t.Errorf("Expected spot BaseURL 'https://api.binance.com', got '%s'", config.Spot.BaseURL)
					}
				}
			} else {
				if config.Spot != nil && config.Spot.APIKey != "" {
					t.Errorf("Expected spot config to be nil or empty, got APIKey '%s'", config.Spot.APIKey)
				}
			}

			// Verify futures section
			if tt.expectFuturesValid {
				if config.Futures == nil {
					t.Errorf("Expected futures config to be loaded, got nil")
				} else {
					if config.Futures.APIKey != "futures_api_key" {
						t.Errorf("Expected futures APIKey 'futures_api_key', got '%s'", config.Futures.APIKey)
					}
					if config.Futures.BaseURL != "https://fapi.binance.com" {
						t.Errorf("Expected futures BaseURL 'https://fapi.binance.com', got '%s'", config.Futures.BaseURL)
					}
					if config.Futures.DefaultLeverage != 10 {
						t.Errorf("Expected futures DefaultLeverage 10, got %d", config.Futures.DefaultLeverage)
					}
				}
			} else {
				if config.Futures != nil && config.Futures.APIKey != "" {
					t.Errorf("Expected futures config to be nil or empty, got APIKey '%s'", config.Futures.APIKey)
				}
			}

			// Verify that loading one section doesn't affect the other
			if tt.includeSpot && tt.includeFutures {
				// Both should be independent
				if config.Spot.APIKey == config.Futures.APIKey {
					t.Errorf("Spot and futures should have different API keys, both have '%s'", config.Spot.APIKey)
				}
				if config.Spot.BaseURL == config.Futures.BaseURL {
					t.Errorf("Spot and futures should have different base URLs, both have '%s'", config.Spot.BaseURL)
				}
			}

			// Verify legacy config if present
			if tt.includeLegacy {
				if config.Binance.APIKey != "legacy_api_key" {
					t.Errorf("Expected legacy APIKey 'legacy_api_key', got '%s'", config.Binance.APIKey)
				}
			}
		})
	}
}
