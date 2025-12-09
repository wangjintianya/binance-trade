package api

import (
	"fmt"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Feature: usdt-futures-trading, Property 1: 合约API凭证验证
// Validates: Requirements 1.1, 1.2
// For any API credentials, the validation function must correctly identify valid and invalid credentials
func TestProperty1_FuturesAPICredentialValidation(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("valid credentials should pass validation", prop.ForAll(
		func(apiKey, apiSecret string) bool {
			// Skip empty strings as they are invalid by design
			if apiKey == "" || apiSecret == "" {
				return true
			}

			authMgr, err := NewAuthManager(apiKey, apiSecret)
			if err != nil {
				return false
			}

			// Valid credentials should pass validation
			err = authMgr.ValidateCredentials()
			return err == nil
		},
		gen.AnyString().SuchThat(func(s string) bool { return s != "" }),
		gen.AnyString().SuchThat(func(s string) bool { return s != "" }),
	))

	properties.Property("empty API key should fail validation", prop.ForAll(
		func(apiSecret string) bool {
			_, err := NewAuthManager("", apiSecret)
			return err != nil
		},
		gen.AnyString(),
	))

	properties.Property("empty API secret should fail validation", prop.ForAll(
		func(apiKey string) bool {
			_, err := NewAuthManager(apiKey, "")
			return err != nil
		},
		gen.AnyString(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Feature: usdt-futures-trading, Property 2: HTTPS协议强制使用
// Validates: Requirements 1.3
// For any futures API request URL, the URL must start with "https://fapi.binance.com"
func TestProperty2_HTTPSProtocolEnforcement(t *testing.T) {
	properties := gopter.NewProperties(nil)

	// Generate valid credentials for testing
	authMgr, _ := NewAuthManager("test_key", "test_secret")
	httpClient := &mockHTTPClient{}

	properties.Property("futures client should only accept HTTPS URLs with correct endpoint", prop.ForAll(
		func(protocol, host, path string) bool {
			url := fmt.Sprintf("%s://%s%s", protocol, host, path)
			
			_, err := NewFuturesClient(url, httpClient, authMgr)
			
			// Should succeed only if URL is https://fapi.binance.com or https://testnet.binancefuture.com
			validURL := url == "https://fapi.binance.com" || url == "https://testnet.binancefuture.com"
			
			if validURL {
				return err == nil
			}
			return err != nil
		},
		gen.OneConstOf("http", "https", "ftp"),
		gen.OneConstOf("fapi.binance.com", "api.binance.com", "example.com", "testnet.binancefuture.com"),
		gen.OneConstOf("", "/api", "/fapi"),
	))

	properties.Property("valid futures endpoints should be accepted", prop.ForAll(
		func() bool {
			validURLs := []string{
				"https://fapi.binance.com",
				"https://testnet.binancefuture.com",
			}
			
			for _, url := range validURLs {
				_, err := NewFuturesClient(url, httpClient, authMgr)
				if err != nil {
					return false
				}
			}
			return true
		},
	))

	properties.Property("non-HTTPS URLs should be rejected", prop.ForAll(
		func() bool {
			invalidURLs := []string{
				"http://fapi.binance.com",
				"ftp://fapi.binance.com",
				"https://api.binance.com",
				"https://example.com",
			}
			
			for _, url := range invalidURLs {
				_, err := NewFuturesClient(url, httpClient, authMgr)
				if err == nil {
					return false
				}
			}
			return true
		},
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Feature: usdt-futures-trading, Property 3: 请求签名正确性
// Validates: Requirements 1.4
// For any futures API request parameters, the signature generated using HMAC SHA256 algorithm
// must be verifiable with the same parameters and key
func TestProperty3_RequestSignatureCorrectness(t *testing.T) {
	properties := gopter.NewProperties(nil)

	// Helper to convert generated params to interface{} map
	convertParams := func(params map[string]string) map[string]interface{} {
		result := make(map[string]interface{})
		for k, v := range params {
			result[k] = v
		}
		return result
	}

	properties.Property("signature should be verifiable with same params and key", prop.ForAll(
		func(apiKey, apiSecret string, params map[string]string) bool {
			// Skip empty credentials
			if apiKey == "" || apiSecret == "" {
				return true
			}

			authMgr, err := NewAuthManager(apiKey, apiSecret)
			if err != nil {
				return true // Skip invalid credentials
			}

			// Convert params to interface{} map
			interfaceParams := convertParams(params)
			
			// Serialize params
			queryString := authMgr.SerializeParams(interfaceParams)
			
			// Generate signature
			signature := authMgr.SignRequest(queryString)
			
			// Verify signature
			return authMgr.VerifySignature(queryString, signature)
		},
		gen.AnyString().SuchThat(func(s string) bool { return s != "" }),
		gen.AnyString().SuchThat(func(s string) bool { return s != "" }),
		gen.MapOf(gen.AlphaString(), gen.AlphaString()),
	))

	properties.Property("different keys should produce different signatures", prop.ForAll(
		func(apiKey1, apiKey2, apiSecret1, apiSecret2 string, params map[string]string) bool {
			// Skip empty or identical credentials
			if apiKey1 == "" || apiSecret1 == "" || apiKey2 == "" || apiSecret2 == "" {
				return true
			}
			if apiSecret1 == apiSecret2 {
				return true
			}

			authMgr1, err1 := NewAuthManager(apiKey1, apiSecret1)
			authMgr2, err2 := NewAuthManager(apiKey2, apiSecret2)
			
			if err1 != nil || err2 != nil {
				return true
			}

			interfaceParams := convertParams(params)
			queryString := authMgr1.SerializeParams(interfaceParams)
			
			sig1 := authMgr1.SignRequest(queryString)
			sig2 := authMgr2.SignRequest(queryString)
			
			// Different secrets should produce different signatures
			return sig1 != sig2
		},
		gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
		gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
		gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
		gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
		gen.MapOf(gen.AlphaString(), gen.AlphaString()),
	))

	properties.Property("modified params should fail verification", prop.ForAll(
		func(apiKey, apiSecret string, params map[string]string, extraKey string, extraValue string) bool {
			// Skip empty credentials
			if apiKey == "" || apiSecret == "" || extraKey == "" {
				return true
			}

			authMgr, err := NewAuthManager(apiKey, apiSecret)
			if err != nil {
				return true
			}

			// Convert params to interface{} map
			interfaceParams := convertParams(params)
			
			// Generate signature for original params
			queryString := authMgr.SerializeParams(interfaceParams)
			signature := authMgr.SignRequest(queryString)
			
			// Modify params
			modifiedParams := make(map[string]interface{})
			for k, v := range interfaceParams {
				modifiedParams[k] = v
			}
			modifiedParams[extraKey] = extraValue
			
			// Modified params should not verify with original signature
			modifiedQueryString := authMgr.SerializeParams(modifiedParams)
			
			// If params didn't actually change, skip
			if queryString == modifiedQueryString {
				return true
			}
			
			return !authMgr.VerifySignature(modifiedQueryString, signature)
		},
		gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
		gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
		gen.MapOf(gen.AlphaString(), gen.AlphaString()),
		gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
		gen.AlphaString(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}
