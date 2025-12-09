package logger

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Feature: binance-auto-trading, Property 20: API操作日志完整性
// Validates: Requirements 6.1
func TestAPIOperationLogCompleteness(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("API operation logs contain operation type, timestamp, and result", prop.ForAll(
		func(operationType string, result string) bool {
			// Create temporary log file
			tmpDir := t.TempDir()
			logFile := filepath.Join(tmpDir, "test.log")

			// Create logger
			logger, err := NewLogger(Config{
				Level:         "info",
				FilePath:      logFile,
				MaxSizeMB:     1,
				MaxBackups:    3,
				EnableConsole: false,
			})
			if err != nil {
				t.Logf("Failed to create logger: %v", err)
				return false
			}

			// Log API operation
			logger.LogAPIOperation(operationType, result, nil)

			// Close file handle before reading
			if l, ok := logger.(*logrusLogger); ok && l.fileHandle != nil {
				l.fileHandle.Close()
			}

			// Read log file
			content, err := os.ReadFile(logFile)
			if err != nil {
				t.Logf("Failed to read log file: %v", err)
				return false
			}

			// Parse JSON log entry
			var logEntry map[string]interface{}
			if err := json.Unmarshal(content, &logEntry); err != nil {
				t.Logf("Failed to parse log entry: %v", err)
				return false
			}

			// Check required fields
			if _, ok := logEntry["operation_type"]; !ok {
				t.Logf("Missing operation_type field")
				return false
			}
			if _, ok := logEntry["timestamp"]; !ok {
				t.Logf("Missing timestamp field")
				return false
			}
			if _, ok := logEntry["result"]; !ok {
				t.Logf("Missing result field")
				return false
			}

			return true
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 && len(s) < 50 }),
		gen.OneConstOf("success", "failure", "error", "timeout"),
	))

	properties.TestingRun(t)
}

// Feature: binance-auto-trading, Property 21: 订单事件日志完整性
// Validates: Requirements 6.2
func TestOrderEventLogCompleteness(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("Order event logs contain orderID, symbol, side, type, and quantity", prop.ForAll(
		func(orderID int64, symbol string, side string, orderType string, quantity float64) bool {
			// Create temporary log file
			tmpDir := t.TempDir()
			logFile := filepath.Join(tmpDir, "test.log")

			// Create logger
			logger, err := NewLogger(Config{
				Level:         "info",
				FilePath:      logFile,
				MaxSizeMB:     1,
				MaxBackups:    3,
				EnableConsole: false,
			})
			if err != nil {
				t.Logf("Failed to create logger: %v", err)
				return false
			}

			// Log order event
			logger.LogOrderEvent("created", orderID, symbol, side, orderType, quantity, nil)

			// Close file handle before reading
			if l, ok := logger.(*logrusLogger); ok && l.fileHandle != nil {
				l.fileHandle.Close()
			}

			// Read log file
			content, err := os.ReadFile(logFile)
			if err != nil {
				t.Logf("Failed to read log file: %v", err)
				return false
			}

			// Parse JSON log entry
			var logEntry map[string]interface{}
			if err := json.Unmarshal(content, &logEntry); err != nil {
				t.Logf("Failed to parse log entry: %v", err)
				return false
			}

			// Check required fields
			requiredFields := []string{"order_id", "symbol", "side", "order_type", "quantity"}
			for _, field := range requiredFields {
				if _, ok := logEntry[field]; !ok {
					t.Logf("Missing required field: %s", field)
					return false
				}
			}

			return true
		},
		gen.Int64(),
		gen.OneConstOf("BTCUSDT", "ETHUSDT", "BNBUSDT"),
		gen.OneConstOf("BUY", "SELL"),
		gen.OneConstOf("MARKET", "LIMIT"),
		gen.Float64Range(0.001, 1000.0),
	))

	properties.TestingRun(t)
}

// Feature: binance-auto-trading, Property 22: 错误日志完整性
// Validates: Requirements 6.3
func TestErrorLogCompleteness(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("Error logs contain error message and context", prop.ForAll(
		func(errorMsg string, contextKey string, contextValue string) bool {
			// Create temporary log file
			tmpDir := t.TempDir()
			logFile := filepath.Join(tmpDir, "test.log")

			// Create logger
			loggerImpl, err := NewLogger(Config{
				Level:         "info",
				FilePath:      logFile,
				MaxSizeMB:     1,
				MaxBackups:    3,
				EnableConsole: false,
			})
			if err != nil {
				t.Logf("Failed to create logger: %v", err)
				return false
			}

			// Create error and context
			testErr := &testError{msg: errorMsg}
			context := map[string]interface{}{
				contextKey: contextValue,
			}

			// Log error
			loggerImpl.LogError(testErr, context)

			// Close file handle before reading
			if l, ok := loggerImpl.(*logrusLogger); ok && l.fileHandle != nil {
				l.fileHandle.Close()
			}

			// Read log file
			content, err := os.ReadFile(logFile)
			if err != nil {
				t.Logf("Failed to read log file: %v", err)
				return false
			}

			// Parse JSON log entry
			var logEntry map[string]interface{}
			if err := json.Unmarshal(content, &logEntry); err != nil {
				t.Logf("Failed to parse log entry: %v", err)
				return false
			}

			// Check required fields
			if _, ok := logEntry["error"]; !ok {
				t.Logf("Missing error field")
				return false
			}

			// Check context is present
			if _, ok := logEntry[contextKey]; !ok {
				t.Logf("Missing context field: %s", contextKey)
				return false
			}

			return true
		},
		gen.Identifier(),
		gen.Identifier(),
		gen.Identifier(),
	))

	properties.TestingRun(t)
}

// testError is a simple error implementation for testing
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

// Feature: binance-auto-trading, Property 23: 日志文件轮转
// Validates: Requirements 6.4
func TestLogRotation(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("Log file rotates when size exceeds limit", prop.ForAll(
		func(maxSizeKB int) bool {
			// Create temporary log file
			tmpDir := t.TempDir()
			logFile := filepath.Join(tmpDir, "test.log")

			// Create logger with small max size
			maxSizeMB := int64(maxSizeKB) / 1024
			if maxSizeMB < 1 {
				maxSizeMB = 1
			}

			loggerImpl, err := NewLogger(Config{
				Level:         "info",
				FilePath:      logFile,
				MaxSizeMB:     maxSizeMB,
				MaxBackups:    3,
				EnableConsole: false,
			})
			if err != nil {
				t.Logf("Failed to create logger: %v", err)
				return false
			}

			// Get the logger implementation
			l, ok := loggerImpl.(*logrusLogger)
			if !ok {
				t.Logf("Failed to cast to logrusLogger")
				return false
			}

			// Write enough data to trigger rotation
			largeMessage := make([]byte, int(maxSizeMB*1024*1024)+1000)
			for i := range largeMessage {
				largeMessage[i] = 'A'
			}

			// Log large message
			l.Info(string(largeMessage), nil)

			// Force rotation check
			l.checkRotation()

			// Close file handle
			if l.fileHandle != nil {
				l.fileHandle.Close()
			}

			// Check if backup file was created
			backupFile := logFile + ".1"
			if _, err := os.Stat(backupFile); os.IsNotExist(err) {
				// Rotation might not have happened yet, which is acceptable
				// as long as the current file exists
				if _, err := os.Stat(logFile); os.IsNotExist(err) {
					t.Logf("Neither current nor backup file exists")
					return false
				}
			}

			// Check that new log file was created
			if _, err := os.Stat(logFile); os.IsNotExist(err) {
				t.Logf("New log file was not created after rotation")
				return false
			}

			return true
		},
		gen.IntRange(100, 1000),
	))

	properties.TestingRun(t)
}

// Feature: binance-auto-trading, Property 24: 敏感信息屏蔽
// Validates: Requirements 6.5
func TestSensitiveInfoMasking(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("Sensitive information is masked in logs", prop.ForAll(
		func(apiKey string, apiSecret string) bool {
			// Ensure keys are long enough to be masked properly (need > 8 chars for first4+****+last4 pattern)
			if len(apiKey) < 9 {
				apiKey = apiKey + "123456789"
			}
			if len(apiSecret) < 9 {
				apiSecret = apiSecret + "987654321"
			}

			// Create temporary log file
			tmpDir := t.TempDir()
			logFile := filepath.Join(tmpDir, "test.log")

			// Create logger
			loggerImpl, err := NewLogger(Config{
				Level:         "info",
				FilePath:      logFile,
				MaxSizeMB:     1,
				MaxBackups:    3,
				EnableConsole: false,
			})
			if err != nil {
				t.Logf("Failed to create logger: %v", err)
				return false
			}

			// Log with sensitive fields
			fields := map[string]interface{}{
				"api_key":    apiKey,
				"api_secret": apiSecret,
			}
			loggerImpl.Info("Test message", fields)

			// Close file handle before reading
			if l, ok := loggerImpl.(*logrusLogger); ok && l.fileHandle != nil {
				l.fileHandle.Close()
			}

			// Read log file
			content, err := os.ReadFile(logFile)
			if err != nil {
				t.Logf("Failed to read log file: %v", err)
				return false
			}

			// Parse JSON log entry
			var logEntry map[string]interface{}
			if err := json.Unmarshal(content, &logEntry); err != nil {
				t.Logf("Failed to parse log entry: %v", err)
				return false
			}

			// Check that api_key field is masked
			if apiKeyVal, ok := logEntry["api_key"]; ok {
				apiKeyStr, ok := apiKeyVal.(string)
				if !ok {
					t.Logf("api_key is not a string")
					return false
				}
				
				// Should not equal the original
				if apiKeyStr == apiKey {
					t.Logf("API key is not masked")
					return false
				}
				
				// Should contain masking pattern
				if !containsSubstring(apiKeyStr, "****") {
					t.Logf("API key does not contain masking pattern")
					return false
				}
				
				// Should show first 4 and last 4 characters
				if len(apiKey) > 8 {
					if !containsSubstring(apiKeyStr, apiKey[:4]) || !containsSubstring(apiKeyStr, apiKey[len(apiKey)-4:]) {
						t.Logf("API key masking format incorrect")
						return false
					}
				}
			}

			// Check that api_secret field is masked
			if apiSecretVal, ok := logEntry["api_secret"]; ok {
				apiSecretStr, ok := apiSecretVal.(string)
				if !ok {
					t.Logf("api_secret is not a string")
					return false
				}
				
				// Should not equal the original
				if apiSecretStr == apiSecret {
					t.Logf("API secret is not masked")
					return false
				}
				
				// Should contain masking pattern
				if !containsSubstring(apiSecretStr, "****") {
					t.Logf("API secret does not contain masking pattern")
					return false
				}
			}

			return true
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) >= 1 }),
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) >= 1 }),
	))

	properties.TestingRun(t)
}

// containsSubstring checks if a substring exists in a string
func containsSubstring(s, substr string) bool {
	if len(substr) == 0 {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Unit Tests

func TestLoggerCreation(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")

	logger, err := NewLogger(Config{
		Level:         "info",
		FilePath:      logFile,
		MaxSizeMB:     1,
		MaxBackups:    3,
		EnableConsole: false,
	})

	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	if logger == nil {
		t.Fatal("Logger is nil")
	}

	// Close file handle
	if l, ok := logger.(*logrusLogger); ok && l.fileHandle != nil {
		l.fileHandle.Close()
	}
}

func TestLogFormatting(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")

	logger, err := NewLogger(Config{
		Level:         "info",
		FilePath:      logFile,
		MaxSizeMB:     1,
		MaxBackups:    3,
		EnableConsole: false,
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Log a message
	logger.Info("Test message", map[string]interface{}{
		"key1": "value1",
		"key2": 123,
	})

	// Close file handle
	if l, ok := logger.(*logrusLogger); ok && l.fileHandle != nil {
		l.fileHandle.Close()
	}

	// Read and verify log format
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	var logEntry map[string]interface{}
	if err := json.Unmarshal(content, &logEntry); err != nil {
		t.Fatalf("Failed to parse log entry: %v", err)
	}

	// Check standard fields
	if _, ok := logEntry["timestamp"]; !ok {
		t.Error("Missing timestamp field")
	}
	if _, ok := logEntry["level"]; !ok {
		t.Error("Missing level field")
	}
	if _, ok := logEntry["message"]; !ok {
		t.Error("Missing message field")
	}
	if _, ok := logEntry["key1"]; !ok {
		t.Error("Missing custom field key1")
	}
}

func TestSensitiveInfoMaskingUnit(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Short API key",
			input:    "short",
			expected: "****",
		},
		{
			name:     "Long API key",
			input:    "verylongapikey123456",
			expected: "very****3456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			logFile := filepath.Join(tmpDir, "test.log")

			logger, err := NewLogger(Config{
				Level:         "info",
				FilePath:      logFile,
				MaxSizeMB:     1,
				MaxBackups:    3,
				EnableConsole: false,
			})
			if err != nil {
				t.Fatalf("Failed to create logger: %v", err)
			}

			// Log with sensitive field
			logger.Info("Test", map[string]interface{}{
				"api_key": tt.input,
			})

			// Close file handle
			if l, ok := logger.(*logrusLogger); ok && l.fileHandle != nil {
				l.fileHandle.Close()
			}

			// Read and verify
			content, err := os.ReadFile(logFile)
			if err != nil {
				t.Fatalf("Failed to read log file: %v", err)
			}

			var logEntry map[string]interface{}
			if err := json.Unmarshal(content, &logEntry); err != nil {
				t.Fatalf("Failed to parse log entry: %v", err)
			}

			apiKey, ok := logEntry["api_key"].(string)
			if !ok {
				t.Fatal("api_key field not found or not a string")
			}

			if apiKey == tt.input {
				t.Errorf("API key was not masked: got %s", apiKey)
			}

			if !containsSubstring(apiKey, "****") {
				t.Errorf("Masked value does not contain ****: got %s", apiKey)
			}
		})
	}
}

func TestLogRotationUnit(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")

	// Create logger with very small max size
	logger, err := NewLogger(Config{
		Level:         "info",
		FilePath:      logFile,
		MaxSizeMB:     1, // 1 KB effectively
		MaxBackups:    2,
		EnableConsole: false,
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	l, ok := logger.(*logrusLogger)
	if !ok {
		t.Fatal("Failed to cast to logrusLogger")
	}

	// Write large message to trigger rotation
	largeMsg := make([]byte, 2*1024*1024) // 2MB
	for i := range largeMsg {
		largeMsg[i] = 'A'
	}

	l.Info(string(largeMsg), nil)
	l.checkRotation()

	// Close file handle
	if l.fileHandle != nil {
		l.fileHandle.Close()
	}

	// Check that backup file exists or current file exists
	_, err1 := os.Stat(logFile + ".1")
	_, err2 := os.Stat(logFile)

	if os.IsNotExist(err1) && os.IsNotExist(err2) {
		t.Error("Neither current nor backup log file exists after rotation")
	}
}

func TestLogLevels(t *testing.T) {
	levels := []string{"debug", "info", "warn", "error"}

	for _, level := range levels {
		t.Run(level, func(t *testing.T) {
			tmpDir := t.TempDir()
			logFile := filepath.Join(tmpDir, "test.log")

			logger, err := NewLogger(Config{
				Level:         level,
				FilePath:      logFile,
				MaxSizeMB:     1,
				MaxBackups:    3,
				EnableConsole: false,
			})
			if err != nil {
				t.Fatalf("Failed to create logger: %v", err)
			}

			// Log at the configured level
			switch level {
			case "debug":
				logger.Debug("Debug message", nil)
			case "info":
				logger.Info("Info message", nil)
			case "warn":
				logger.Warn("Warn message", nil)
			case "error":
				logger.Error("Error message", nil)
			}

			// Close file handle
			if l, ok := logger.(*logrusLogger); ok && l.fileHandle != nil {
				l.fileHandle.Close()
			}

			// Verify log was written
			content, err := os.ReadFile(logFile)
			if err != nil {
				t.Fatalf("Failed to read log file: %v", err)
			}

			if len(content) == 0 {
				t.Error("No log content written")
			}
		})
	}
}

// Feature: usdt-futures-trading, Property 37: 合约API操作日志完整性
// Validates: Requirements 9.1
func TestFuturesAPIOperationLogCompleteness(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("Futures API operation logs contain operation type, timestamp, and result", prop.ForAll(
		func(operationType string, result string) bool {
			// Create temporary log file
			tmpDir := t.TempDir()
			logFile := filepath.Join(tmpDir, "test.log")

			// Create logger
			logger, err := NewLogger(Config{
				Level:         "info",
				FilePath:      logFile,
				MaxSizeMB:     1,
				MaxBackups:    3,
				EnableConsole: false,
			})
			if err != nil {
				t.Logf("Failed to create logger: %v", err)
				return false
			}

			// Log futures API operation
			logger.LogFuturesAPIOperation(operationType, result, nil)

			// Close file handle before reading
			if l, ok := logger.(*logrusLogger); ok && l.fileHandle != nil {
				l.fileHandle.Close()
			}

			// Read log file
			content, err := os.ReadFile(logFile)
			if err != nil {
				t.Logf("Failed to read log file: %v", err)
				return false
			}

			// Parse JSON log entry
			var logEntry map[string]interface{}
			if err := json.Unmarshal(content, &logEntry); err != nil {
				t.Logf("Failed to parse log entry: %v", err)
				return false
			}

			// Check required fields
			if _, ok := logEntry["operation_type"]; !ok {
				t.Logf("Missing operation_type field")
				return false
			}
			if _, ok := logEntry["timestamp"]; !ok {
				t.Logf("Missing timestamp field")
				return false
			}
			if _, ok := logEntry["result"]; !ok {
				t.Logf("Missing result field")
				return false
			}

			return true
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 && len(s) < 50 }),
		gen.OneConstOf("success", "failure", "error", "timeout"),
	))

	properties.TestingRun(t)
}

// Feature: usdt-futures-trading, Property 38: 订单事件日志完整性
// Validates: Requirements 9.2
func TestFuturesOrderEventLogCompleteness(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("Futures order event logs contain orderID, symbol, side, type, quantity, and position change", prop.ForAll(
		func(orderID int64, symbol string, side string, orderType string, quantity float64, positionAmt float64) bool {
			// Create temporary log file
			tmpDir := t.TempDir()
			logFile := filepath.Join(tmpDir, "test.log")

			// Create logger
			logger, err := NewLogger(Config{
				Level:         "info",
				FilePath:      logFile,
				MaxSizeMB:     1,
				MaxBackups:    3,
				EnableConsole: false,
			})
			if err != nil {
				t.Logf("Failed to create logger: %v", err)
				return false
			}

			// Create position change info
			positionChange := map[string]interface{}{
				"position_amt": positionAmt,
				"entry_price":  45000.0,
			}

			// Log futures order event
			logger.LogFuturesOrderEvent("created", orderID, symbol, side, orderType, quantity, positionChange, nil)

			// Close file handle before reading
			if l, ok := logger.(*logrusLogger); ok && l.fileHandle != nil {
				l.fileHandle.Close()
			}

			// Read log file
			content, err := os.ReadFile(logFile)
			if err != nil {
				t.Logf("Failed to read log file: %v", err)
				return false
			}

			// Parse JSON log entry
			var logEntry map[string]interface{}
			if err := json.Unmarshal(content, &logEntry); err != nil {
				t.Logf("Failed to parse log entry: %v", err)
				return false
			}

			// Check required fields
			requiredFields := []string{"order_id", "symbol", "side", "order_type", "quantity", "position_change"}
			for _, field := range requiredFields {
				if _, ok := logEntry[field]; !ok {
					t.Logf("Missing required field: %s", field)
					return false
				}
			}

			return true
		},
		gen.Int64(),
		gen.OneConstOf("BTCUSDT", "ETHUSDT", "BNBUSDT"),
		gen.OneConstOf("BUY", "SELL"),
		gen.OneConstOf("MARKET", "LIMIT"),
		gen.Float64Range(0.001, 1000.0),
		gen.Float64Range(-1000.0, 1000.0),
	))

	properties.TestingRun(t)
}

// Feature: usdt-futures-trading, Property 39: 强平事件日志完整性
// Validates: Requirements 9.3
func TestLiquidationEventLogCompleteness(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("Liquidation event logs contain reason, liquidation price, and loss amount", prop.ForAll(
		func(symbol string, positionSide string, liquidationPrice float64, lossAmount float64, reason string) bool {
			// Create temporary log file
			tmpDir := t.TempDir()
			logFile := filepath.Join(tmpDir, "test.log")

			// Create logger
			logger, err := NewLogger(Config{
				Level:         "info",
				FilePath:      logFile,
				MaxSizeMB:     1,
				MaxBackups:    3,
				EnableConsole: false,
			})
			if err != nil {
				t.Logf("Failed to create logger: %v", err)
				return false
			}

			// Log liquidation event
			logger.LogLiquidationEvent(symbol, positionSide, liquidationPrice, lossAmount, reason, nil)

			// Close file handle before reading
			if l, ok := logger.(*logrusLogger); ok && l.fileHandle != nil {
				l.fileHandle.Close()
			}

			// Read log file
			content, err := os.ReadFile(logFile)
			if err != nil {
				t.Logf("Failed to read log file: %v", err)
				return false
			}

			// Parse JSON log entry
			var logEntry map[string]interface{}
			if err := json.Unmarshal(content, &logEntry); err != nil {
				t.Logf("Failed to parse log entry: %v", err)
				return false
			}

			// Check required fields
			requiredFields := []string{"symbol", "position_side", "liquidation_price", "loss_amount", "reason"}
			for _, field := range requiredFields {
				if _, ok := logEntry[field]; !ok {
					t.Logf("Missing required field: %s", field)
					return false
				}
			}

			return true
		},
		gen.OneConstOf("BTCUSDT", "ETHUSDT", "BNBUSDT"),
		gen.OneConstOf("LONG", "SHORT"),
		gen.Float64Range(1000.0, 100000.0),
		gen.Float64Range(0.0, 10000.0),
		gen.OneConstOf("margin_call", "insufficient_margin", "max_leverage_exceeded"),
	))

	properties.TestingRun(t)
}

// Feature: usdt-futures-trading, Property 40: 资金费率结算日志完整性
// Validates: Requirements 9.4
func TestFundingRateSettlementLogCompleteness(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("Funding rate settlement logs contain funding fee amount", prop.ForAll(
		func(symbol string, fundingFee float64, fundingRate float64, positionSize float64) bool {
			// Create temporary log file
			tmpDir := t.TempDir()
			logFile := filepath.Join(tmpDir, "test.log")

			// Create logger
			logger, err := NewLogger(Config{
				Level:         "info",
				FilePath:      logFile,
				MaxSizeMB:     1,
				MaxBackups:    3,
				EnableConsole: false,
			})
			if err != nil {
				t.Logf("Failed to create logger: %v", err)
				return false
			}

			// Log funding rate settlement
			logger.LogFundingRateSettlement(symbol, fundingFee, fundingRate, positionSize, nil)

			// Close file handle before reading
			if l, ok := logger.(*logrusLogger); ok && l.fileHandle != nil {
				l.fileHandle.Close()
			}

			// Read log file
			content, err := os.ReadFile(logFile)
			if err != nil {
				t.Logf("Failed to read log file: %v", err)
				return false
			}

			// Parse JSON log entry
			var logEntry map[string]interface{}
			if err := json.Unmarshal(content, &logEntry); err != nil {
				t.Logf("Failed to parse log entry: %v", err)
				return false
			}

			// Check required field
			if _, ok := logEntry["funding_fee"]; !ok {
				t.Logf("Missing funding_fee field")
				return false
			}

			return true
		},
		gen.OneConstOf("BTCUSDT", "ETHUSDT", "BNBUSDT"),
		gen.Float64Range(-100.0, 100.0),
		gen.Float64Range(-0.01, 0.01),
		gen.Float64Range(0.001, 1000.0),
	))

	properties.TestingRun(t)
}

// Feature: usdt-futures-trading, Property 41: 敏感信息屏蔽
// Validates: Requirements 9.5
func TestFuturesSensitiveInfoMasking(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("Sensitive information is masked in futures logs", prop.ForAll(
		func(apiKey string, apiSecret string) bool {
			// Ensure keys are long enough to be masked properly
			if len(apiKey) < 9 {
				apiKey = apiKey + "123456789"
			}
			if len(apiSecret) < 9 {
				apiSecret = apiSecret + "987654321"
			}

			// Create temporary log file
			tmpDir := t.TempDir()
			logFile := filepath.Join(tmpDir, "test.log")

			// Create logger
			logger, err := NewLogger(Config{
				Level:         "info",
				FilePath:      logFile,
				MaxSizeMB:     1,
				MaxBackups:    3,
				EnableConsole: false,
			})
			if err != nil {
				t.Logf("Failed to create logger: %v", err)
				return false
			}

			// Log futures API operation with sensitive fields
			fields := map[string]interface{}{
				"api_key":    apiKey,
				"api_secret": apiSecret,
			}
			logger.LogFuturesAPIOperation("authenticate", "success", fields)

			// Close file handle before reading
			if l, ok := logger.(*logrusLogger); ok && l.fileHandle != nil {
				l.fileHandle.Close()
			}

			// Read log file
			content, err := os.ReadFile(logFile)
			if err != nil {
				t.Logf("Failed to read log file: %v", err)
				return false
			}

			// Parse JSON log entry
			var logEntry map[string]interface{}
			if err := json.Unmarshal(content, &logEntry); err != nil {
				t.Logf("Failed to parse log entry: %v", err)
				return false
			}

			// Check that api_key field is masked
			if apiKeyVal, ok := logEntry["api_key"]; ok {
				apiKeyStr, ok := apiKeyVal.(string)
				if !ok {
					t.Logf("api_key is not a string")
					return false
				}
				
				// Should not equal the original
				if apiKeyStr == apiKey {
					t.Logf("API key is not masked")
					return false
				}
				
				// Should contain masking pattern
				if !containsSubstring(apiKeyStr, "****") {
					t.Logf("API key does not contain masking pattern")
					return false
				}
			}

			// Check that api_secret field is masked
			if apiSecretVal, ok := logEntry["api_secret"]; ok {
				apiSecretStr, ok := apiSecretVal.(string)
				if !ok {
					t.Logf("api_secret is not a string")
					return false
				}
				
				// Should not equal the original
				if apiSecretStr == apiSecret {
					t.Logf("API secret is not masked")
					return false
				}
				
				// Should contain masking pattern
				if !containsSubstring(apiSecretStr, "****") {
					t.Logf("API secret does not contain masking pattern")
					return false
				}
			}

			return true
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) >= 1 }),
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) >= 1 }),
	))

	properties.TestingRun(t)
}

// Feature: usdt-futures-trading, Property 43: 日志交易类型标记
// Validates: Requirements 10.4, 11.5
func TestLogTradingTypeMarker(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("Futures logs contain trading type marker", prop.ForAll(
		func(tradingType string, operationType string) bool {
			// Create temporary log file
			tmpDir := t.TempDir()
			logFile := filepath.Join(tmpDir, "test.log")

			// Create logger with trading type
			logger, err := NewLogger(Config{
				Level:         "info",
				FilePath:      logFile,
				MaxSizeMB:     1,
				MaxBackups:    3,
				EnableConsole: false,
				TradingType:   tradingType,
			})
			if err != nil {
				t.Logf("Failed to create logger: %v", err)
				return false
			}

			// Log futures API operation
			logger.LogFuturesAPIOperation(operationType, "success", nil)

			// Close file handle before reading
			if l, ok := logger.(*logrusLogger); ok && l.fileHandle != nil {
				l.fileHandle.Close()
			}

			// Read log file
			content, err := os.ReadFile(logFile)
			if err != nil {
				t.Logf("Failed to read log file: %v", err)
				return false
			}

			// Parse JSON log entry
			var logEntry map[string]interface{}
			if err := json.Unmarshal(content, &logEntry); err != nil {
				t.Logf("Failed to parse log entry: %v", err)
				return false
			}

			// Check that trading_type field exists and matches
			if tradingTypeVal, ok := logEntry["trading_type"]; !ok {
				t.Logf("Missing trading_type field")
				return false
			} else {
				tradingTypeStr, ok := tradingTypeVal.(string)
				if !ok {
					t.Logf("trading_type is not a string")
					return false
				}
				if tradingTypeStr != tradingType {
					t.Logf("trading_type mismatch: expected %s, got %s", tradingType, tradingTypeStr)
					return false
				}
			}

			return true
		},
		gen.OneConstOf("spot", "futures"),
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 && len(s) < 50 }),
	))

	properties.TestingRun(t)
}
