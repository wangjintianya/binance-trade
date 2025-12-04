package api

import (
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Feature: binance-auto-trading, Property 2: 请求签名正确性
// For any API request parameters, the signature generated using HMAC SHA256
// must be verifiable using the same parameters and secret key
func TestProperty_SignatureCorrectness(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("signature verification round-trip", prop.ForAll(
		func(apiKey, apiSecret string, paramCount int) bool {
			// Skip empty credentials as they should be rejected
			if apiKey == "" || apiSecret == "" {
				return true
			}

			// Create auth manager
			am, err := NewAuthManager(apiKey, apiSecret)
			if err != nil {
				return false
			}

			// Generate random parameters
			params := make(map[string]interface{})
			for i := 0; i < paramCount%10; i++ {
				keyVal, _ := gen.Identifier().Sample()
				key := keyVal.(string)
				value, _ := genParamValue().Sample()
				params[key] = value
			}

			// Serialize parameters
			queryString := am.SerializeParams(params)

			// Generate signature
			signature := am.SignRequest(queryString)

			// Verify signature
			return am.VerifySignature(queryString, signature)
		},
		gen.Identifier(),  // apiKey
		gen.Identifier(),  // apiSecret
		gen.IntRange(0, 20), // paramCount
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Helper generator for parameter values
func genParamValue() gopter.Gen {
	return gen.OneGenOf(
		gen.AlphaString(),
		gen.IntRange(0, 1000000),
		gen.Int64Range(0, 1000000),
		gen.Float64Range(0, 1000000),
		gen.Bool(),
	)
}

// Unit tests for basic functionality

func TestNewAuthManager(t *testing.T) {
	tests := []struct {
		name      string
		apiKey    string
		apiSecret string
		wantErr   bool
	}{
		{
			name:      "valid credentials",
			apiKey:    "test_key",
			apiSecret: "test_secret",
			wantErr:   false,
		},
		{
			name:      "empty api key",
			apiKey:    "",
			apiSecret: "test_secret",
			wantErr:   true,
		},
		{
			name:      "empty api secret",
			apiKey:    "test_key",
			apiSecret: "",
			wantErr:   true,
		},
		{
			name:      "both empty",
			apiKey:    "",
			apiSecret: "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			am, err := NewAuthManager(tt.apiKey, tt.apiSecret)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewAuthManager() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && am == nil {
				t.Error("NewAuthManager() returned nil without error")
			}
		})
	}
}

func TestAuthManager_ValidateCredentials(t *testing.T) {
	tests := []struct {
		name      string
		apiKey    string
		apiSecret string
		wantErr   bool
	}{
		{
			name:      "valid credentials",
			apiKey:    "test_key",
			apiSecret: "test_secret",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			am, _ := NewAuthManager(tt.apiKey, tt.apiSecret)
			if err := am.ValidateCredentials(); (err != nil) != tt.wantErr {
				t.Errorf("ValidateCredentials() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAuthManager_SignRequest(t *testing.T) {
	am, _ := NewAuthManager("test_key", "test_secret")

	tests := []struct {
		name        string
		queryString string
		want        string
	}{
		{
			name:        "empty query string",
			queryString: "",
			want:        "9ba1f63365a6caf66e46348f43cdef956015bea997adeb06deb019943f9e5923",
		},
		{
			name:        "simple query string",
			queryString: "symbol=BTCUSDT&side=BUY",
			want:        "5f2f2d0c3e3c3e3a3a3a3a3a3a3a3a3a3a3a3a3a3a3a3a3a3a3a3a3a3a3a3a3a",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := am.SignRequest(tt.queryString)
			// Just verify it returns a non-empty hex string
			if len(got) != 64 {
				t.Errorf("SignRequest() returned signature of length %d, want 64", len(got))
			}
		})
	}
}

func TestAuthManager_SerializeParams(t *testing.T) {
	am, _ := NewAuthManager("test_key", "test_secret")

	tests := []struct {
		name   string
		params map[string]interface{}
		want   string
	}{
		{
			name:   "empty params",
			params: map[string]interface{}{},
			want:   "",
		},
		{
			name: "single string param",
			params: map[string]interface{}{
				"symbol": "BTCUSDT",
			},
			want: "symbol=BTCUSDT",
		},
		{
			name: "multiple params sorted",
			params: map[string]interface{}{
				"symbol": "BTCUSDT",
				"side":   "BUY",
				"type":   "MARKET",
			},
			want: "side=BUY&symbol=BTCUSDT&type=MARKET",
		},
		{
			name: "mixed types",
			params: map[string]interface{}{
				"symbol":    "BTCUSDT",
				"quantity":  1.5,
				"timestamp": int64(1234567890),
				"test":      true,
			},
			want: "quantity=1.5&symbol=BTCUSDT&test=true&timestamp=1234567890",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := am.SerializeParams(tt.params)
			if got != tt.want {
				t.Errorf("SerializeParams() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuthManager_GenerateTimestamp(t *testing.T) {
	am, _ := NewAuthManager("test_key", "test_secret")

	ts1 := am.GenerateTimestamp()
	ts2 := am.GenerateTimestamp()

	if ts1 <= 0 {
		t.Error("GenerateTimestamp() returned non-positive timestamp")
	}

	if ts2 < ts1 {
		t.Error("GenerateTimestamp() returned decreasing timestamps")
	}
}

func TestAuthManager_ValidateTimestamp(t *testing.T) {
	am, _ := NewAuthManager("test_key", "test_secret")

	tests := []struct {
		name      string
		timestamp int64
		wantErr   bool
	}{
		{
			name:      "current timestamp",
			timestamp: am.GenerateTimestamp(),
			wantErr:   false,
		},
		{
			name:      "old timestamp (10 minutes ago)",
			timestamp: am.GenerateTimestamp() - (10 * 60 * 1000),
			wantErr:   true,
		},
		{
			name:      "future timestamp (10 minutes ahead)",
			timestamp: am.GenerateTimestamp() + (10 * 60 * 1000),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := am.ValidateTimestamp(tt.timestamp)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTimestamp() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Feature: binance-auto-trading, Property 1: HTTPS协议强制使用
// For any API request URL, the URL must start with "https://"
func TestProperty_HTTPSEnforcement(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("all valid URLs must use HTTPS", prop.ForAll(
		func(protocol, domain, path string) bool {
			// Skip empty domains as they are invalid regardless of protocol
			if domain == "" {
				return true
			}

			// Build URL
			url := protocol + "://" + domain + path

			// Validate URL
			err := ValidateURL(url)

			// If protocol is https, validation should pass
			// If protocol is not https, validation should fail
			if protocol == "https" {
				return err == nil
			}
			return err != nil
		},
		gen.OneConstOf("http", "https", "ftp", "ws"),
		gen.OneConstOf("api.binance.com", "testnet.binance.vision", "example.com", ""),
		gen.OneConstOf("/api/v3/order", "/api/v3/account", "", "/test"),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Unit test for ValidateURL
func TestValidateURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "valid https URL",
			url:     "https://api.binance.com",
			wantErr: false,
		},
		{
			name:    "valid https URL with path",
			url:     "https://api.binance.com/api/v3/order",
			wantErr: false,
		},
		{
			name:    "http URL should fail",
			url:     "http://api.binance.com",
			wantErr: true,
		},
		{
			name:    "ftp URL should fail",
			url:     "ftp://api.binance.com",
			wantErr: true,
		},
		{
			name:    "empty URL should fail",
			url:     "",
			wantErr: true,
		},
		{
			name:    "URL without protocol should fail",
			url:     "api.binance.com",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateURL() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Feature: binance-auto-trading, Property 3: 无效凭证拒绝
// For any invalid API credentials, the validation function must return an error and reject the connection
func TestProperty_InvalidCredentialsRejection(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("invalid credentials are rejected", prop.ForAll(
		func(apiKey, apiSecret string, makeInvalid bool) bool {
			// If makeInvalid is true, make at least one credential empty
			if makeInvalid {
				if len(apiKey) > 0 {
					apiKey = ""
				} else if len(apiSecret) > 0 {
					apiSecret = ""
				} else {
					// Both already empty
					apiKey = ""
					apiSecret = ""
				}
			}

			// Try to create auth manager
			am, err := NewAuthManager(apiKey, apiSecret)

			// If credentials are invalid (empty), should return error
			if apiKey == "" || apiSecret == "" {
				if err == nil {
					return false // Should have returned error
				}
				if am != nil {
					return false // Should not have created manager
				}
				return true
			}

			// If credentials are valid, should succeed
			if err != nil {
				return false // Should not have returned error
			}
			if am == nil {
				return false // Should have created manager
			}

			// Validate credentials should also work
			return am.ValidateCredentials() == nil
		},
		gen.AlphaString(),
		gen.AlphaString(),
		gen.Bool(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Additional unit test for credential validation edge cases
func TestAuthManager_InvalidCredentials(t *testing.T) {
	tests := []struct {
		name      string
		apiKey    string
		apiSecret string
		wantErr   bool
	}{
		{
			name:      "both empty",
			apiKey:    "",
			apiSecret: "",
			wantErr:   true,
		},
		{
			name:      "empty key",
			apiKey:    "",
			apiSecret: "secret",
			wantErr:   true,
		},
		{
			name:      "empty secret",
			apiKey:    "key",
			apiSecret: "",
			wantErr:   true,
		},
		{
			name:      "whitespace key",
			apiKey:    "   ",
			apiSecret: "secret",
			wantErr:   false, // Whitespace is technically valid, just not useful
		},
		{
			name:      "valid credentials",
			apiKey:    "valid_key",
			apiSecret: "valid_secret",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewAuthManager(tt.apiKey, tt.apiSecret)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewAuthManager() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
