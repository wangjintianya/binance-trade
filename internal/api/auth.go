package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

// AuthManager handles API authentication and request signing
type AuthManager struct {
	apiKey    string
	apiSecret string
}

// NewAuthManager creates a new AuthManager instance
func NewAuthManager(apiKey, apiSecret string) (*AuthManager, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("API key cannot be empty")
	}
	if apiSecret == "" {
		return nil, fmt.Errorf("API secret cannot be empty")
	}
	return &AuthManager{
		apiKey:    apiKey,
		apiSecret: apiSecret,
	}, nil
}

// GetAPIKey returns the API key
func (am *AuthManager) GetAPIKey() string {
	return am.apiKey
}

// ValidateCredentials checks if the credentials are valid (non-empty)
func (am *AuthManager) ValidateCredentials() error {
	if am.apiKey == "" {
		return fmt.Errorf("invalid credentials: API key is empty")
	}
	if am.apiSecret == "" {
		return fmt.Errorf("invalid credentials: API secret is empty")
	}
	return nil
}

// GenerateTimestamp generates a current timestamp in milliseconds
func (am *AuthManager) GenerateTimestamp() int64 {
	return time.Now().UnixMilli()
}

// ValidateTimestamp checks if a timestamp is within acceptable range (5 minutes)
func (am *AuthManager) ValidateTimestamp(timestamp int64) error {
	now := time.Now().UnixMilli()
	diff := now - timestamp
	
	// Allow 5 minutes window (5 * 60 * 1000 milliseconds)
	maxDiff := int64(5 * 60 * 1000)
	
	if diff < -maxDiff || diff > maxDiff {
		return fmt.Errorf("timestamp out of acceptable range: %d ms difference", diff)
	}
	
	return nil
}

// SerializeParams converts a map of parameters to a sorted query string
func (am *AuthManager) SerializeParams(params map[string]interface{}) string {
	if len(params) == 0 {
		return ""
	}
	
	// Sort keys for consistent ordering
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	
	// Build query string
	values := url.Values{}
	for _, k := range keys {
		v := params[k]
		var strValue string
		
		switch val := v.(type) {
		case string:
			strValue = val
		case int:
			strValue = strconv.Itoa(val)
		case int64:
			strValue = strconv.FormatInt(val, 10)
		case float64:
			strValue = strconv.FormatFloat(val, 'f', -1, 64)
		case bool:
			strValue = strconv.FormatBool(val)
		default:
			strValue = fmt.Sprintf("%v", val)
		}
		
		values.Add(k, strValue)
	}
	
	return values.Encode()
}

// SignRequest generates HMAC SHA256 signature for the request
func (am *AuthManager) SignRequest(queryString string) string {
	mac := hmac.New(sha256.New, []byte(am.apiSecret))
	mac.Write([]byte(queryString))
	return hex.EncodeToString(mac.Sum(nil))
}

// VerifySignature verifies if a signature is valid for the given query string
func (am *AuthManager) VerifySignature(queryString, signature string) bool {
	expectedSignature := am.SignRequest(queryString)
	return hmac.Equal([]byte(expectedSignature), []byte(signature))
}

// SignRequestWithParams signs a request with the given parameters and adds timestamp
func (am *AuthManager) SignRequestWithParams(params map[string]interface{}) (string, error) {
	if err := am.ValidateCredentials(); err != nil {
		return "", err
	}
	
	// Add timestamp if not present
	if _, exists := params["timestamp"]; !exists {
		params["timestamp"] = am.GenerateTimestamp()
	}
	
	// Serialize parameters
	queryString := am.SerializeParams(params)
	
	// Generate signature
	signature := am.SignRequest(queryString)
	
	// Append signature to query string
	if queryString != "" {
		queryString += "&signature=" + signature
	} else {
		queryString = "signature=" + signature
	}
	
	return queryString, nil
}

// ValidateURL checks if a URL uses HTTPS protocol
func ValidateURL(urlStr string) error {
	if urlStr == "" {
		return fmt.Errorf("URL cannot be empty")
	}
	
	if !strings.HasPrefix(urlStr, "https://") {
		return fmt.Errorf("URL must use HTTPS protocol, got: %s", urlStr)
	}
	
	return nil
}
