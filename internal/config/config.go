package config

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// BinanceConfig holds Binance API configuration
type BinanceConfig struct {
	APIKey    string `yaml:"api_key"`
	APISecret string `yaml:"api_secret"`
	BaseURL   string `yaml:"base_url"`
	Testnet   bool   `yaml:"testnet"`
}

// RiskConfig holds risk management configuration
type RiskConfig struct {
	MaxOrderAmount    float64 `yaml:"max_order_amount"`
	MaxDailyOrders    int     `yaml:"max_daily_orders"`
	MinBalanceReserve float64 `yaml:"min_balance_reserve"`
	MaxAPICallsPerMin int     `yaml:"max_api_calls_per_min"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level      string `yaml:"level"`
	File       string `yaml:"file"`
	MaxSizeMB  int    `yaml:"max_size_mb"`
	MaxBackups int    `yaml:"max_backups"`
}

// RetryConfig holds retry configuration
type RetryConfig struct {
	MaxAttempts       int     `yaml:"max_attempts"`
	InitialDelayMs    int     `yaml:"initial_delay_ms"`
	BackoffMultiplier float64 `yaml:"backoff_multiplier"`
}

// Config represents the application configuration
type Config struct {
	Binance BinanceConfig `yaml:"binance"`
	Risk    RiskConfig    `yaml:"risk"`
	Logging LoggingConfig `yaml:"logging"`
	Retry   RetryConfig   `yaml:"retry"`
}

// ConfigManager defines the interface for configuration management
type ConfigManager interface {
	Load(path string) (*Config, error)
	Validate(config *Config) error
	GetConfig() *Config
}

// configManager implements the ConfigManager interface
type configManager struct {
	config *Config
}

// NewConfigManager creates a new ConfigManager instance
func NewConfigManager() ConfigManager {
	return &configManager{}
}

// Load reads and parses the YAML configuration file with environment variable substitution
func (cm *configManager) Load(path string) (*Config, error) {
	// Read the configuration file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Replace environment variables
	content := replaceEnvVars(string(data))

	// Parse YAML
	var config Config
	if err := yaml.Unmarshal([]byte(content), &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate configuration
	if err := cm.Validate(&config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	cm.config = &config
	return &config, nil
}

// Validate checks if the configuration is valid
func (cm *configManager) Validate(config *Config) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	// Validate Binance configuration
	if config.Binance.APIKey == "" {
		return fmt.Errorf("binance.api_key is required")
	}
	if config.Binance.APISecret == "" {
		return fmt.Errorf("binance.api_secret is required")
	}
	if config.Binance.BaseURL == "" {
		return fmt.Errorf("binance.base_url is required")
	}
	if !strings.HasPrefix(config.Binance.BaseURL, "https://") {
		return fmt.Errorf("binance.base_url must use HTTPS protocol")
	}

	// Validate Risk configuration
	if config.Risk.MaxOrderAmount <= 0 {
		return fmt.Errorf("risk.max_order_amount must be greater than 0")
	}
	if config.Risk.MaxDailyOrders <= 0 {
		return fmt.Errorf("risk.max_daily_orders must be greater than 0")
	}
	if config.Risk.MinBalanceReserve < 0 {
		return fmt.Errorf("risk.min_balance_reserve cannot be negative")
	}
	if config.Risk.MaxAPICallsPerMin <= 0 {
		return fmt.Errorf("risk.max_api_calls_per_min must be greater than 0")
	}

	// Validate Logging configuration
	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLogLevels[config.Logging.Level] {
		return fmt.Errorf("logging.level must be one of: debug, info, warn, error")
	}
	if config.Logging.File == "" {
		return fmt.Errorf("logging.file is required")
	}
	if config.Logging.MaxSizeMB <= 0 {
		return fmt.Errorf("logging.max_size_mb must be greater than 0")
	}
	if config.Logging.MaxBackups < 0 {
		return fmt.Errorf("logging.max_backups cannot be negative")
	}

	// Validate Retry configuration
	if config.Retry.MaxAttempts <= 0 {
		return fmt.Errorf("retry.max_attempts must be greater than 0")
	}
	if config.Retry.InitialDelayMs <= 0 {
		return fmt.Errorf("retry.initial_delay_ms must be greater than 0")
	}
	if config.Retry.BackoffMultiplier <= 1.0 {
		return fmt.Errorf("retry.backoff_multiplier must be greater than 1.0")
	}

	return nil
}

// GetConfig returns the current configuration
func (cm *configManager) GetConfig() *Config {
	return cm.config
}

// replaceEnvVars replaces ${VAR_NAME} patterns with environment variable values
func replaceEnvVars(content string) string {
	// Pattern to match ${VAR_NAME}
	re := regexp.MustCompile(`\$\{([^}]+)\}`)

	return re.ReplaceAllStringFunc(content, func(match string) string {
		// Extract variable name (remove ${ and })
		varName := match[2 : len(match)-1]

		// Get environment variable value
		if value := os.Getenv(varName); value != "" {
			return value
		}

		// Return original if not found (will be caught by validation)
		return match
	})
}
