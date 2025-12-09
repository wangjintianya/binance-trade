package api

import (
	"encoding/json"
	"fmt"
)

// SpotClient defines the interface for interacting with Binance Spot API
type SpotClient interface {
	// Account information
	GetAccountInfo() (*AccountInfo, error)
	GetBalance(asset string) (*Balance, error)

	// Market data
	GetPrice(symbol string) (*Price, error)
	GetKlines(symbol string, interval string, limit int) ([]*Kline, error)

	// Order operations
	CreateOrder(order *OrderRequest) (*OrderResponse, error)
	CancelOrder(symbol string, orderID int64) (*CancelResponse, error)
	GetOrder(symbol string, orderID int64) (*Order, error)
	GetOpenOrders(symbol string) ([]*Order, error)
	GetHistoricalOrders(symbol string, startTime, endTime int64) ([]*Order, error)
}

// spotClient implements SpotClient interface
type spotClient struct {
	baseURL    string
	httpClient HTTPClient
	authMgr    *AuthManager
}

// NewSpotClient creates a new Binance Spot API client
func NewSpotClient(baseURL string, httpClient HTTPClient, authMgr *AuthManager) (SpotClient, error) {
	if err := ValidateURL(baseURL); err != nil {
		return nil, err
	}
	
	if authMgr == nil {
		return nil, fmt.Errorf("auth manager cannot be nil")
	}
	
	if err := authMgr.ValidateCredentials(); err != nil {
		return nil, err
	}
	
	return &spotClient{
		baseURL:    baseURL,
		httpClient: httpClient,
		authMgr:    authMgr,
	}, nil
}

// GetAccountInfo retrieves account information from Binance
func (c *spotClient) GetAccountInfo() (*AccountInfo, error) {
	params := make(map[string]interface{})
	params["timestamp"] = c.authMgr.GenerateTimestamp()
	
	queryString, err := c.authMgr.SignRequestWithParams(params)
	if err != nil {
		return nil, err
	}
	
	url := fmt.Sprintf("%s/api/v3/account?%s", c.baseURL, queryString)
	headers := map[string]string{
		"X-MBX-APIKEY": c.authMgr.GetAPIKey(),
	}
	
	body, err := c.httpClient.DoWithRetry("GET", url, nil, headers)
	if err != nil {
		return nil, err
	}
	
	var accountInfo AccountInfo
	if err := json.Unmarshal(body, &accountInfo); err != nil {
		return nil, fmt.Errorf("failed to parse account info: %w", err)
	}
	
	return &accountInfo, nil
}

// GetBalance retrieves the balance for a specific asset
func (c *spotClient) GetBalance(asset string) (*Balance, error) {
	// Parse balances from account info
	var rawData map[string]interface{}
	params := make(map[string]interface{})
	params["timestamp"] = c.authMgr.GenerateTimestamp()
	
	queryString, err := c.authMgr.SignRequestWithParams(params)
	if err != nil {
		return nil, err
	}
	
	url := fmt.Sprintf("%s/api/v3/account?%s", c.baseURL, queryString)
	headers := map[string]string{
		"X-MBX-APIKEY": c.authMgr.GetAPIKey(),
	}
	
	body, err := c.httpClient.DoWithRetry("GET", url, nil, headers)
	if err != nil {
		return nil, err
	}
	
	if err := json.Unmarshal(body, &rawData); err != nil {
		return nil, fmt.Errorf("failed to parse account data: %w", err)
	}
	
	// Extract balances array
	balances, ok := rawData["balances"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("balances field not found or invalid")
	}
	
	// Find the requested asset
	for _, b := range balances {
		balanceMap, ok := b.(map[string]interface{})
		if !ok {
			continue
		}
		
		assetName, _ := balanceMap["asset"].(string)
		if assetName == asset {
			freeStr, _ := balanceMap["free"].(string)
			lockedStr, _ := balanceMap["locked"].(string)
			
			var free, locked float64
			fmt.Sscanf(freeStr, "%f", &free)
			fmt.Sscanf(lockedStr, "%f", &locked)
			
			return &Balance{
				Asset:  asset,
				Free:   free,
				Locked: locked,
			}, nil
		}
	}
	
	return nil, fmt.Errorf("asset %s not found in account", asset)
}

// GetPrice retrieves the current price for a symbol
func (c *spotClient) GetPrice(symbol string) (*Price, error) {
	params := map[string]interface{}{
		"symbol": symbol,
	}
	
	url := fmt.Sprintf("%s/api/v3/ticker/price", c.baseURL)
	
	body, err := c.httpClient.DoWithRetry("GET", url, params, nil)
	if err != nil {
		return nil, err
	}
	
	var priceData struct {
		Symbol string `json:"symbol"`
		Price  string `json:"price"`
	}
	
	if err := json.Unmarshal(body, &priceData); err != nil {
		return nil, fmt.Errorf("failed to parse price data: %w", err)
	}
	
	var price float64
	if _, err := fmt.Sscanf(priceData.Price, "%f", &price); err != nil {
		return nil, fmt.Errorf("failed to parse price value: %w", err)
	}
	
	return &Price{
		Symbol: priceData.Symbol,
		Price:  price,
	}, nil
}

// GetKlines retrieves candlestick data for a symbol
func (c *spotClient) GetKlines(symbol string, interval string, limit int) ([]*Kline, error) {
	params := map[string]interface{}{
		"symbol":   symbol,
		"interval": interval,
		"limit":    limit,
	}
	
	url := fmt.Sprintf("%s/api/v3/klines", c.baseURL)
	
	body, err := c.httpClient.DoWithRetry("GET", url, params, nil)
	if err != nil {
		return nil, err
	}
	
	var rawKlines [][]interface{}
	if err := json.Unmarshal(body, &rawKlines); err != nil {
		return nil, fmt.Errorf("failed to parse klines data: %w", err)
	}
	
	klines := make([]*Kline, 0, len(rawKlines))
	for _, raw := range rawKlines {
		if len(raw) < 7 {
			continue
		}
		
		kline := &Kline{}
		
		// Parse each field
		if openTime, ok := raw[0].(float64); ok {
			kline.OpenTime = int64(openTime)
		}
		if openStr, ok := raw[1].(string); ok {
			fmt.Sscanf(openStr, "%f", &kline.Open)
		}
		if highStr, ok := raw[2].(string); ok {
			fmt.Sscanf(highStr, "%f", &kline.High)
		}
		if lowStr, ok := raw[3].(string); ok {
			fmt.Sscanf(lowStr, "%f", &kline.Low)
		}
		if closeStr, ok := raw[4].(string); ok {
			fmt.Sscanf(closeStr, "%f", &kline.Close)
		}
		if volumeStr, ok := raw[5].(string); ok {
			fmt.Sscanf(volumeStr, "%f", &kline.Volume)
		}
		if closeTime, ok := raw[6].(float64); ok {
			kline.CloseTime = int64(closeTime)
		}
		
		klines = append(klines, kline)
	}
	
	return klines, nil
}

// CreateOrder creates a new order
func (c *spotClient) CreateOrder(order *OrderRequest) (*OrderResponse, error) {
	if order == nil {
		return nil, fmt.Errorf("order request cannot be nil")
	}
	
	params := make(map[string]interface{})
	params["symbol"] = order.Symbol
	params["side"] = string(order.Side)
	params["type"] = string(order.Type)
	params["quantity"] = order.Quantity
	params["timestamp"] = c.authMgr.GenerateTimestamp()
	
	// Only include price for limit orders
	if order.Type == OrderTypeLimit {
		if order.Price <= 0 {
			return nil, fmt.Errorf("price must be greater than 0 for limit orders")
		}
		params["price"] = order.Price
		
		// TimeInForce is required for limit orders
		if order.TimeInForce == "" {
			params["timeInForce"] = "GTC" // Good Till Cancel by default
		} else {
			params["timeInForce"] = order.TimeInForce
		}
	}
	
	queryString, err := c.authMgr.SignRequestWithParams(params)
	if err != nil {
		return nil, err
	}
	
	url := fmt.Sprintf("%s/api/v3/order?%s", c.baseURL, queryString)
	headers := map[string]string{
		"X-MBX-APIKEY": c.authMgr.GetAPIKey(),
	}
	
	body, err := c.httpClient.DoWithRetry("POST", url, nil, headers)
	if err != nil {
		return nil, err
	}
	
	var response OrderResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse order response: %w", err)
	}
	
	return &response, nil
}

// CancelOrder cancels an order
func (c *spotClient) CancelOrder(symbol string, orderID int64) (*CancelResponse, error) {
	if symbol == "" {
		return nil, fmt.Errorf("symbol cannot be empty")
	}
	if orderID <= 0 {
		return nil, fmt.Errorf("orderID must be greater than 0")
	}
	
	params := make(map[string]interface{})
	params["symbol"] = symbol
	params["orderId"] = orderID
	params["timestamp"] = c.authMgr.GenerateTimestamp()
	
	queryString, err := c.authMgr.SignRequestWithParams(params)
	if err != nil {
		return nil, err
	}
	
	url := fmt.Sprintf("%s/api/v3/order?%s", c.baseURL, queryString)
	headers := map[string]string{
		"X-MBX-APIKEY": c.authMgr.GetAPIKey(),
	}
	
	body, err := c.httpClient.DoWithRetry("DELETE", url, nil, headers)
	if err != nil {
		return nil, err
	}
	
	var response CancelResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse cancel response: %w", err)
	}
	
	return &response, nil
}

// GetOrder retrieves order details
func (c *spotClient) GetOrder(symbol string, orderID int64) (*Order, error) {
	if symbol == "" {
		return nil, fmt.Errorf("symbol cannot be empty")
	}
	if orderID <= 0 {
		return nil, fmt.Errorf("orderID must be greater than 0")
	}
	
	params := make(map[string]interface{})
	params["symbol"] = symbol
	params["orderId"] = orderID
	params["timestamp"] = c.authMgr.GenerateTimestamp()
	
	queryString, err := c.authMgr.SignRequestWithParams(params)
	if err != nil {
		return nil, err
	}
	
	url := fmt.Sprintf("%s/api/v3/order?%s", c.baseURL, queryString)
	headers := map[string]string{
		"X-MBX-APIKEY": c.authMgr.GetAPIKey(),
	}
	
	body, err := c.httpClient.DoWithRetry("GET", url, nil, headers)
	if err != nil {
		return nil, err
	}
	
	var order Order
	if err := json.Unmarshal(body, &order); err != nil {
		return nil, fmt.Errorf("failed to parse order: %w", err)
	}
	
	return &order, nil
}

// GetOpenOrders retrieves all open orders
func (c *spotClient) GetOpenOrders(symbol string) ([]*Order, error) {
	params := make(map[string]interface{})
	if symbol != "" {
		params["symbol"] = symbol
	}
	params["timestamp"] = c.authMgr.GenerateTimestamp()
	
	queryString, err := c.authMgr.SignRequestWithParams(params)
	if err != nil {
		return nil, err
	}
	
	url := fmt.Sprintf("%s/api/v3/openOrders?%s", c.baseURL, queryString)
	headers := map[string]string{
		"X-MBX-APIKEY": c.authMgr.GetAPIKey(),
	}
	
	body, err := c.httpClient.DoWithRetry("GET", url, nil, headers)
	if err != nil {
		return nil, err
	}
	
	var orders []*Order
	if err := json.Unmarshal(body, &orders); err != nil {
		return nil, fmt.Errorf("failed to parse open orders: %w", err)
	}
	
	return orders, nil
}

// GetHistoricalOrders retrieves historical orders
func (c *spotClient) GetHistoricalOrders(symbol string, startTime, endTime int64) ([]*Order, error) {
	if symbol == "" {
		return nil, fmt.Errorf("symbol cannot be empty")
	}
	
	params := make(map[string]interface{})
	params["symbol"] = symbol
	if startTime > 0 {
		params["startTime"] = startTime
	}
	if endTime > 0 {
		params["endTime"] = endTime
	}
	params["timestamp"] = c.authMgr.GenerateTimestamp()
	
	queryString, err := c.authMgr.SignRequestWithParams(params)
	if err != nil {
		return nil, err
	}
	
	url := fmt.Sprintf("%s/api/v3/allOrders?%s", c.baseURL, queryString)
	headers := map[string]string{
		"X-MBX-APIKEY": c.authMgr.GetAPIKey(),
	}
	
	body, err := c.httpClient.DoWithRetry("GET", url, nil, headers)
	if err != nil {
		return nil, err
	}
	
	var orders []*Order
	if err := json.Unmarshal(body, &orders); err != nil {
		return nil, fmt.Errorf("failed to parse historical orders: %w", err)
	}
	
	return orders, nil
}
