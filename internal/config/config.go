package config

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// TradingType represents the type of trading (spot or futures)
type TradingType string

const (
	TradingTypeSpot    TradingType = "spot"
	TradingTypeFutures TradingType = "futures"
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
	Level         string `yaml:"level"`
	File          string `yaml:"file"`
	SpotFile      string `yaml:"spot_file"`
	FuturesFile   string `yaml:"futures_file"`
	MaxSizeMB     int    `yaml:"max_size_mb"`
	MaxBackups    int    `yaml:"max_backups"`
}

// RetryConfig holds retry configuration
type RetryConfig struct {
	MaxAttempts       int     `yaml:"max_attempts"`
	InitialDelayMs    int     `yaml:"initial_delay_ms"`
	BackoffMultiplier float64 `yaml:"backoff_multiplier"`
}

// ConditionalOrdersConfig holds conditional orders configuration
type ConditionalOrdersConfig struct {
	MonitoringIntervalMs      int  `yaml:"monitoring_interval_ms"`
	MaxActiveOrders           int  `yaml:"max_active_orders"`
	TriggerExecutionTimeoutMs int  `yaml:"trigger_execution_timeout_ms"`
	EnableSmartPolling        bool `yaml:"enable_smart_polling"`
}

// StopLossConfig holds stop loss configuration
type StopLossConfig struct {
	DefaultTrailPercent float64 `yaml:"default_trail_percent"`
	MinTrailPercent     float64 `yaml:"min_trail_percent"`
	MaxTrailPercent     float64 `yaml:"max_trail_percent"`
	UpdateIntervalMs    int     `yaml:"update_interval_ms"`
}

// FuturesRiskConfig holds futures-specific risk configuration
type FuturesRiskConfig struct {
	MaxOrderValue         float64 `yaml:"max_order_value"`
	MaxPositionValue      float64 `yaml:"max_position_value"`
	MaxLeverage           int     `yaml:"max_leverage"`
	MinMarginRatio        float64 `yaml:"min_margin_ratio"`
	LiquidationBuffer     float64 `yaml:"liquidation_buffer"`
	MaxDailyOrders        int     `yaml:"max_daily_orders"`
	MaxAPICallsPerMin     int     `yaml:"max_api_calls_per_min"`
}

// FuturesMonitoringConfig holds futures monitoring configuration
type FuturesMonitoringConfig struct {
	PositionUpdateIntervalMs      int `yaml:"position_update_interval_ms"`
	ConditionalOrderIntervalMs    int `yaml:"conditional_order_interval_ms"`
	FundingRateCheckIntervalMs    int `yaml:"funding_rate_check_interval_ms"`
}

// FuturesStopLossConfig holds futures stop loss configuration
type FuturesStopLossConfig struct {
	DefaultCallbackRate float64 `yaml:"default_callback_rate"`
	MinCallbackRate     float64 `yaml:"min_callback_rate"`
	MaxCallbackRate     float64 `yaml:"max_callback_rate"`
}

// FuturesConfig holds futures-specific configuration
type FuturesConfig struct {
	APIKey            string                      `yaml:"api_key"`
	APISecret         string                      `yaml:"api_secret"`
	BaseURL           string                      `yaml:"base_url"`
	Testnet           bool                        `yaml:"testnet"`
	DefaultLeverage   int                         `yaml:"default_leverage"`
	DefaultMarginType string                      `yaml:"default_margin_type"`
	DualSidePosition  bool                        `yaml:"dual_side_position"`
	Risk              FuturesRiskConfig           `yaml:"risk"`
	Monitoring        FuturesMonitoringConfig     `yaml:"monitoring"`
	StopLoss          FuturesStopLossConfig       `yaml:"stop_loss"`
}

// Config represents the application configuration
type Config struct {
	// Legacy fields for backward compatibility
	Binance           BinanceConfig           `yaml:"binance"`
	Risk              RiskConfig              `yaml:"risk"`
	Logging           LoggingConfig           `yaml:"logging"`
	Retry             RetryConfig             `yaml:"retry"`
	ConditionalOrders ConditionalOrdersConfig `yaml:"conditional_orders"`
	StopLoss          StopLossConfig          `yaml:"stop_loss"`
	
	// New fields for multi-trading type support
	Spot    *BinanceConfig `yaml:"spot,omitempty"`
	Futures *FuturesConfig `yaml:"futures,omitempty"`
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

	// Support both old and new config formats
	// If Spot or Futures sections exist, use them; otherwise use legacy Binance section
	hasLegacyConfig := config.Binance.APIKey != ""
	hasSpotConfig := config.Spot != nil && config.Spot.APIKey != ""
	hasFuturesConfig := config.Futures != nil && config.Futures.APIKey != ""
	
	if !hasLegacyConfig && !hasSpotConfig && !hasFuturesConfig {
		return fmt.Errorf("at least one trading configuration (binance, spot, or futures) is required")
	}

	// Validate legacy Binance configuration if present
	if hasLegacyConfig {
		if err := cm.validateBinanceConfig(&config.Binance); err != nil {
			return fmt.Errorf("binance config: %w", err)
		}
	}
	
	// Validate Spot configuration if present
	if hasSpotConfig {
		if err := cm.validateBinanceConfig(config.Spot); err != nil {
			return fmt.Errorf("spot config: %w", err)
		}
	}
	
	// Validate Futures configuration if present
	if hasFuturesConfig {
		if err := cm.validateFuturesConfig(config.Futures); err != nil {
			return fmt.Errorf("futures config: %w", err)
		}
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

	// Validate ConditionalOrders configuration
	if config.ConditionalOrders.MonitoringIntervalMs <= 0 {
		return fmt.Errorf("conditional_orders.monitoring_interval_ms must be greater than 0")
	}
	if config.ConditionalOrders.MaxActiveOrders <= 0 {
		return fmt.Errorf("conditional_orders.max_active_orders must be greater than 0")
	}
	if config.ConditionalOrders.TriggerExecutionTimeoutMs <= 0 {
		return fmt.Errorf("conditional_orders.trigger_execution_timeout_ms must be greater than 0")
	}

	// Validate StopLoss configuration
	if config.StopLoss.DefaultTrailPercent <= 0 {
		return fmt.Errorf("stop_loss.default_trail_percent must be greater than 0")
	}
	if config.StopLoss.MinTrailPercent <= 0 {
		return fmt.Errorf("stop_loss.min_trail_percent must be greater than 0")
	}
	if config.StopLoss.MaxTrailPercent <= 0 {
		return fmt.Errorf("stop_loss.max_trail_percent must be greater than 0")
	}
	if config.StopLoss.MinTrailPercent > config.StopLoss.MaxTrailPercent {
		return fmt.Errorf("stop_loss.min_trail_percent cannot be greater than max_trail_percent")
	}
	if config.StopLoss.DefaultTrailPercent < config.StopLoss.MinTrailPercent || config.StopLoss.DefaultTrailPercent > config.StopLoss.MaxTrailPercent {
		return fmt.Errorf("stop_loss.default_trail_percent must be between min_trail_percent and max_trail_percent")
	}
	if config.StopLoss.UpdateIntervalMs <= 0 {
		return fmt.Errorf("stop_loss.update_interval_ms must be greater than 0")
	}

	return nil
}

// GetConfig returns the current configuration
func (cm *configManager) GetConfig() *Config {
	return cm.config
}

// validateBinanceConfig validates a Binance configuration section
func (cm *configManager) validateBinanceConfig(config *BinanceConfig) error {
	if config.APIKey == "" {
		return fmt.Errorf("api_key is required")
	}
	if config.APISecret == "" {
		return fmt.Errorf("api_secret is required")
	}
	if config.BaseURL == "" {
		return fmt.Errorf("base_url is required")
	}
	if !strings.HasPrefix(config.BaseURL, "https://") {
		return fmt.Errorf("base_url must use HTTPS protocol")
	}
	return nil
}

// validateFuturesConfig validates a Futures configuration section
func (cm *configManager) validateFuturesConfig(config *FuturesConfig) error {
	if config.APIKey == "" {
		return fmt.Errorf("api_key is required")
	}
	if config.APISecret == "" {
		return fmt.Errorf("api_secret is required")
	}
	if config.BaseURL == "" {
		return fmt.Errorf("base_url is required")
	}
	if !strings.HasPrefix(config.BaseURL, "https://") {
		return fmt.Errorf("base_url must use HTTPS protocol")
	}
	
	// Validate futures-specific fields
	if config.DefaultLeverage < 1 || config.DefaultLeverage > 125 {
		return fmt.Errorf("default_leverage must be between 1 and 125")
	}
	
	validMarginTypes := map[string]bool{"CROSSED": true, "ISOLATED": true}
	if config.DefaultMarginType != "" && !validMarginTypes[config.DefaultMarginType] {
		return fmt.Errorf("default_margin_type must be CROSSED or ISOLATED")
	}
	
	// Validate risk config
	if config.Risk.MaxOrderValue <= 0 {
		return fmt.Errorf("risk.max_order_value must be greater than 0")
	}
	if config.Risk.MaxPositionValue <= 0 {
		return fmt.Errorf("risk.max_position_value must be greater than 0")
	}
	if config.Risk.MaxLeverage < 1 || config.Risk.MaxLeverage > 125 {
		return fmt.Errorf("risk.max_leverage must be between 1 and 125")
	}
	if config.Risk.MinMarginRatio < 0 || config.Risk.MinMarginRatio > 1 {
		return fmt.Errorf("risk.min_margin_ratio must be between 0 and 1")
	}
	if config.Risk.LiquidationBuffer < 0 || config.Risk.LiquidationBuffer > 1 {
		return fmt.Errorf("risk.liquidation_buffer must be between 0 and 1")
	}
	
	return nil
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
