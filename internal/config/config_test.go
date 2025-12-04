package config

import (
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
			},
			expectError: true,
			errorMsg:    "binance.api_key is required",
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
			},
			expectError: true,
			errorMsg:    "binance.base_url must use HTTPS protocol",
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
