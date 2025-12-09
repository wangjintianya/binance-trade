package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// Logger defines the interface for logging operations
type Logger interface {
	Debug(msg string, fields map[string]interface{})
	Info(msg string, fields map[string]interface{})
	Warn(msg string, fields map[string]interface{})
	Error(msg string, fields map[string]interface{})
	Fatal(msg string, fields map[string]interface{})
	
	// Additional methods for specific logging scenarios
	LogAPIOperation(operationType string, result string, fields map[string]interface{})
	LogOrderEvent(eventType string, orderID int64, symbol, side, orderType string, quantity float64, fields map[string]interface{})
	LogError(err error, context map[string]interface{})
	
	// Futures-specific logging methods
	LogFuturesAPIOperation(operationType string, result string, fields map[string]interface{})
	LogFuturesOrderEvent(eventType string, orderID int64, symbol, side, orderType string, quantity float64, positionChange map[string]interface{}, fields map[string]interface{})
	LogLiquidationEvent(symbol string, positionSide string, liquidationPrice float64, lossAmount float64, reason string, fields map[string]interface{})
	LogFundingRateSettlement(symbol string, fundingFee float64, fundingRate float64, positionSize float64, fields map[string]interface{})
	
	// Set trading type for log entries
	SetTradingType(tradingType string)
}

// Config holds logger configuration
type Config struct {
	Level         string // debug, info, warn, error
	FilePath      string // path to log file
	MaxSizeMB     int64  // max size in MB before rotation
	MaxBackups    int    // max number of backup files
	EnableConsole bool   // also log to console
	TradingType   string // trading type marker (spot, futures)
}

// logrusLogger implements Logger interface using logrus
type logrusLogger struct {
	logger      *logrus.Logger
	config      Config
	mu          sync.Mutex
	currentSize int64
	fileHandle  *os.File
	tradingType string
}

// sensitivePatterns are regex patterns for sensitive information
var sensitivePatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(api[_-]?key|apikey)[\s:=]+([a-zA-Z0-9]{8,})`),
	regexp.MustCompile(`(?i)(api[_-]?secret|apisecret|secret[_-]?key)[\s:=]+([a-zA-Z0-9]{8,})`),
	regexp.MustCompile(`(?i)(password|passwd|pwd)[\s:=]+([^\s]+)`),
	regexp.MustCompile(`(?i)(token|auth[_-]?token)[\s:=]+([a-zA-Z0-9]{8,})`),
}

// NewLogger creates a new logger instance
func NewLogger(config Config) (Logger, error) {
	log := logrus.New()
	
	// Set log level
	level, err := logrus.ParseLevel(config.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	log.SetLevel(level)
	
	// Set formatter
	log.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339,
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "timestamp",
			logrus.FieldKeyLevel: "level",
			logrus.FieldKeyMsg:   "message",
		},
	})
	
	logger := &logrusLogger{
		logger:      log,
		config:      config,
		tradingType: config.TradingType,
	}
	
	// Setup output
	if config.FilePath != "" {
		if err := logger.setupFileOutput(); err != nil {
			return nil, err
		}
	}
	
	if config.EnableConsole {
		if config.FilePath != "" {
			// Write to both file and console
			logger.logger.SetOutput(io.MultiWriter(logger.fileHandle, os.Stdout))
		} else {
			logger.logger.SetOutput(os.Stdout)
		}
	}
	
	return logger, nil
}

// setupFileOutput initializes file output with rotation support
func (l *logrusLogger) setupFileOutput() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	// Create directory if it doesn't exist
	dir := filepath.Dir(l.config.FilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}
	
	// Open or create log file
	file, err := os.OpenFile(l.config.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	
	// Get current file size
	info, err := file.Stat()
	if err != nil {
		file.Close()
		return fmt.Errorf("failed to stat log file: %w", err)
	}
	
	l.fileHandle = file
	l.currentSize = info.Size()
	l.logger.SetOutput(file)
	
	return nil
}

// checkRotation checks if log rotation is needed and performs it
func (l *logrusLogger) checkRotation() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	if l.fileHandle == nil {
		return nil
	}
	
	maxSize := l.config.MaxSizeMB * 1024 * 1024
	if l.currentSize < maxSize {
		return nil
	}
	
	// Close current file
	if err := l.fileHandle.Close(); err != nil {
		return err
	}
	
	// Rotate backup files
	for i := l.config.MaxBackups - 1; i >= 1; i-- {
		oldPath := fmt.Sprintf("%s.%d", l.config.FilePath, i)
		newPath := fmt.Sprintf("%s.%d", l.config.FilePath, i+1)
		os.Rename(oldPath, newPath) // Ignore error if file doesn't exist
	}
	
	// Move current file to .1
	backupPath := fmt.Sprintf("%s.1", l.config.FilePath)
	if err := os.Rename(l.config.FilePath, backupPath); err != nil {
		return err
	}
	
	// Create new file
	file, err := os.OpenFile(l.config.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	
	l.fileHandle = file
	l.currentSize = 0
	l.logger.SetOutput(file)
	
	return nil
}

// maskSensitiveInfo masks sensitive information in strings
func maskSensitiveInfo(s string) string {
	result := s
	for _, pattern := range sensitivePatterns {
		result = pattern.ReplaceAllStringFunc(result, func(match string) string {
			parts := pattern.FindStringSubmatch(match)
			if len(parts) >= 3 {
				key := parts[1]
				value := parts[2]
				if len(value) > 8 {
					masked := value[:4] + "****" + value[len(value)-4:]
					return key + "=" + masked
				}
			}
			return match
		})
	}
	return result
}

// maskSensitiveFields masks sensitive information in field values
func maskSensitiveFields(fields map[string]interface{}) map[string]interface{} {
	if fields == nil {
		return nil
	}
	
	masked := make(map[string]interface{})
	sensitiveKeys := []string{"api_key", "apikey", "api_secret", "apisecret", "secret_key", "password", "passwd", "pwd", "token", "auth_token"}
	
	for k, v := range fields {
		lowerKey := strings.ToLower(k)
		isSensitive := false
		for _, sk := range sensitiveKeys {
			if strings.Contains(lowerKey, sk) {
				isSensitive = true
				break
			}
		}
		
		if isSensitive {
			if str, ok := v.(string); ok && len(str) > 8 {
				masked[k] = str[:4] + "****" + str[len(str)-4:]
			} else {
				masked[k] = "****"
			}
		} else {
			// Also mask string values that might contain sensitive info
			if str, ok := v.(string); ok {
				masked[k] = maskSensitiveInfo(str)
			} else {
				masked[k] = v
			}
		}
	}
	
	return masked
}

// log is the internal logging method
func (l *logrusLogger) log(level logrus.Level, msg string, fields map[string]interface{}) {
	// Check rotation before logging
	l.checkRotation()
	
	// Mask sensitive information
	maskedMsg := maskSensitiveInfo(msg)
	maskedFields := maskSensitiveFields(fields)
	
	// Add trading type marker if set
	if l.tradingType != "" {
		if maskedFields == nil {
			maskedFields = make(map[string]interface{})
		}
		maskedFields["trading_type"] = l.tradingType
	}
	
	// Create entry with fields
	entry := l.logger.WithFields(logrus.Fields(maskedFields))
	
	// Update size estimate
	l.mu.Lock()
	l.currentSize += int64(len(maskedMsg) + 100) // Rough estimate
	l.mu.Unlock()
	
	// Log at appropriate level
	switch level {
	case logrus.DebugLevel:
		entry.Debug(maskedMsg)
	case logrus.InfoLevel:
		entry.Info(maskedMsg)
	case logrus.WarnLevel:
		entry.Warn(maskedMsg)
	case logrus.ErrorLevel:
		entry.Error(maskedMsg)
	case logrus.FatalLevel:
		entry.Fatal(maskedMsg)
	}
}

func (l *logrusLogger) Debug(msg string, fields map[string]interface{}) {
	l.log(logrus.DebugLevel, msg, fields)
}

func (l *logrusLogger) Info(msg string, fields map[string]interface{}) {
	l.log(logrus.InfoLevel, msg, fields)
}

func (l *logrusLogger) Warn(msg string, fields map[string]interface{}) {
	l.log(logrus.WarnLevel, msg, fields)
}

func (l *logrusLogger) Error(msg string, fields map[string]interface{}) {
	l.log(logrus.ErrorLevel, msg, fields)
}

func (l *logrusLogger) Fatal(msg string, fields map[string]interface{}) {
	l.log(logrus.FatalLevel, msg, fields)
}

// LogAPIOperation logs API operations with required fields
func (l *logrusLogger) LogAPIOperation(operationType string, result string, fields map[string]interface{}) {
	if fields == nil {
		fields = make(map[string]interface{})
	}
	fields["operation_type"] = operationType
	fields["result"] = result
	fields["timestamp"] = time.Now().Unix()
	
	l.Info("API operation", fields)
}

// LogOrderEvent logs order events with complete details
func (l *logrusLogger) LogOrderEvent(eventType string, orderID int64, symbol, side, orderType string, quantity float64, fields map[string]interface{}) {
	if fields == nil {
		fields = make(map[string]interface{})
	}
	fields["event_type"] = eventType
	fields["order_id"] = orderID
	fields["symbol"] = symbol
	fields["side"] = side
	fields["order_type"] = orderType
	fields["quantity"] = quantity
	
	l.Info("Order event", fields)
}

// LogError logs errors with context information
func (l *logrusLogger) LogError(err error, context map[string]interface{}) {
	if context == nil {
		context = make(map[string]interface{})
	}
	context["error"] = err.Error()
	
	l.Error("Error occurred", context)
}

// SetTradingType sets the trading type marker for all log entries
func (l *logrusLogger) SetTradingType(tradingType string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.tradingType = tradingType
}

// LogFuturesAPIOperation logs futures API operations with required fields
func (l *logrusLogger) LogFuturesAPIOperation(operationType string, result string, fields map[string]interface{}) {
	if fields == nil {
		fields = make(map[string]interface{})
	}
	fields["operation_type"] = operationType
	fields["result"] = result
	fields["timestamp"] = time.Now().Unix()
	fields["api_type"] = "futures"
	
	l.Info("Futures API operation", fields)
}

// LogFuturesOrderEvent logs futures order events with complete details including position changes
func (l *logrusLogger) LogFuturesOrderEvent(eventType string, orderID int64, symbol, side, orderType string, quantity float64, positionChange map[string]interface{}, fields map[string]interface{}) {
	if fields == nil {
		fields = make(map[string]interface{})
	}
	fields["event_type"] = eventType
	fields["order_id"] = orderID
	fields["symbol"] = symbol
	fields["side"] = side
	fields["order_type"] = orderType
	fields["quantity"] = quantity
	
	// Add position change information
	if positionChange != nil {
		fields["position_change"] = positionChange
	}
	
	l.Info("Futures order event", fields)
}

// LogLiquidationEvent logs liquidation events with complete details
func (l *logrusLogger) LogLiquidationEvent(symbol string, positionSide string, liquidationPrice float64, lossAmount float64, reason string, fields map[string]interface{}) {
	if fields == nil {
		fields = make(map[string]interface{})
	}
	fields["symbol"] = symbol
	fields["position_side"] = positionSide
	fields["liquidation_price"] = liquidationPrice
	fields["loss_amount"] = lossAmount
	fields["reason"] = reason
	fields["timestamp"] = time.Now().Unix()
	
	l.Error("Liquidation event", fields)
}

// LogFundingRateSettlement logs funding rate settlement events
func (l *logrusLogger) LogFundingRateSettlement(symbol string, fundingFee float64, fundingRate float64, positionSize float64, fields map[string]interface{}) {
	if fields == nil {
		fields = make(map[string]interface{})
	}
	fields["symbol"] = symbol
	fields["funding_fee"] = fundingFee
	fields["funding_rate"] = fundingRate
	fields["position_size"] = positionSize
	fields["timestamp"] = time.Now().Unix()
	
	l.Info("Funding rate settlement", fields)
}
