package api

import (
	"encoding/json"
	"fmt"
)

// AccountInfo represents account information from Binance
type AccountInfo struct {
	MakerCommission  int64
	TakerCommission  int64
	BuyerCommission  int64
	SellerCommission int64
	CanTrade         bool
	CanWithdraw      bool
	CanDeposit       bool
	UpdateTime       int64
}

// Balance represents an asset balance
type Balance struct {
	Asset  string
	Free   float64
	Locked float64
}

// Price represents a symbol price
type Price struct {
	Symbol string
	Price  float64
}

// Kline represents candlestick data
type Kline struct {
	OpenTime  int64
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    float64
	CloseTime int64
}

// OrderSide represents order side (BUY/SELL)
type OrderSide string

const (
	OrderSideBuy  OrderSide = "BUY"
	OrderSideSell OrderSide = "SELL"
)

// OrderType represents order type (MARKET/LIMIT)
type OrderType string

const (
	OrderTypeMarket OrderType = "MARKET"
	OrderTypeLimit  OrderType = "LIMIT"
)

// OrderStatus represents order status
type OrderStatus string

const (
	OrderStatusNew            OrderStatus = "NEW"
	OrderStatusPartiallyFilled OrderStatus = "PARTIALLY_FILLED"
	OrderStatusFilled         OrderStatus = "FILLED"
	OrderStatusCanceled       OrderStatus = "CANCELED"
	OrderStatusRejected       OrderStatus = "REJECTED"
	OrderStatusExpired        OrderStatus = "EXPIRED"
)

// OrderRequest represents a request to create an order
type OrderRequest struct {
	Symbol      string
	Side        OrderSide
	Type        OrderType
	Quantity    float64
	Price       float64
	TimeInForce string
}

// OrderResponse represents the response from creating an order
type OrderResponse struct {
	OrderID                 int64
	Symbol                  string
	Status                  OrderStatus
	Price                   float64
	OrigQty                 float64
	ExecutedQty             float64
	CummulativeQuoteQty     float64
	TransactTime            int64
}

// Order represents an order
type Order struct {
	OrderID                 int64
	Symbol                  string
	Side                    OrderSide
	Type                    OrderType
	Status                  OrderStatus
	Price                   float64
	OrigQty                 float64
	ExecutedQty             float64
	CummulativeQuoteQty     float64
	Time                    int64
	UpdateTime              int64
}

// CancelResponse represents the response from canceling an order
type CancelResponse struct {
	Symbol            string
	OrderID           int64
	OrigClientOrderID string
	Status            OrderStatus
}

// BinanceClient defines the interface for interacting with Binance API
type BinanceClient interface {
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

// HTTPClient defines the interface for making HTTP requests
type HTTPClient interface {
	Do(method, url string, params map[string]interface{}, headers map[string]string) ([]byte, error)
	DoWithRetry(method, url string, params map[string]interface{}, headers map[string]string) ([]byte, error)
}

// binanceClient implements BinanceClient interface
type binanceClient struct {
	baseURL    string
	httpClient HTTPClient
	authMgr    *AuthManager
}

// NewBinanceClient creates a new Binance API client
func NewBinanceClient(baseURL string, httpClient HTTPClient, authMgr *AuthManager) (BinanceClient, error) {
	if err := ValidateURL(baseURL); err != nil {
		return nil, err
	}
	
	if authMgr == nil {
		return nil, fmt.Errorf("auth manager cannot be nil")
	}
	
	if err := authMgr.ValidateCredentials(); err != nil {
		return nil, err
	}
	
	return &binanceClient{
		baseURL:    baseURL,
		httpClient: httpClient,
		authMgr:    authMgr,
	}, nil
}

// GetAccountInfo retrieves account information from Binance
func (c *binanceClient) GetAccountInfo() (*AccountInfo, error) {
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
func (c *binanceClient) GetBalance(asset string) (*Balance, error) {
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
func (c *binanceClient) GetPrice(symbol string) (*Price, error) {
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
func (c *binanceClient) GetKlines(symbol string, interval string, limit int) ([]*Kline, error) {
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
func (c *binanceClient) CreateOrder(order *OrderRequest) (*OrderResponse, error) {
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
func (c *binanceClient) CancelOrder(symbol string, orderID int64) (*CancelResponse, error) {
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
func (c *binanceClient) GetOrder(symbol string, orderID int64) (*Order, error) {
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
func (c *binanceClient) GetOpenOrders(symbol string) ([]*Order, error) {
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
func (c *binanceClient) GetHistoricalOrders(symbol string, startTime, endTime int64) ([]*Order, error) {
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
