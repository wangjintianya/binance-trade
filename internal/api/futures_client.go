package api

import (
	"encoding/json"
	"fmt"
)

// PositionSide represents position side for futures
type PositionSide string

const (
	PositionSideLong  PositionSide = "LONG"
	PositionSideShort PositionSide = "SHORT"
	PositionSideBoth  PositionSide = "BOTH"
)

// MarginType represents margin type for futures
type MarginType string

const (
	MarginTypeIsolated MarginType = "ISOLATED"
	MarginTypeCrossed  MarginType = "CROSSED"
)

// PriceType represents price type for triggers
type PriceType string

const (
	PriceTypeMark PriceType = "MARK"
	PriceTypeLast PriceType = "LAST"
)

// FuturesAccountInfo represents futures account information
type FuturesAccountInfo struct {
	Assets                      []FuturesAsset
	Positions                   []Position
	CanDeposit                  bool
	CanTrade                    bool
	CanWithdraw                 bool
	FeeTier                     int
	MaxWithdrawAmount           float64
	TotalInitialMargin          float64
	TotalMaintMargin            float64
	TotalMarginBalance          float64
	TotalOpenOrderInitialMargin float64
	TotalPositionInitialMargin  float64
	TotalUnrealizedProfit       float64
	TotalWalletBalance          float64
	UpdateTime                  int64
}

// FuturesAsset represents a futures asset balance
type FuturesAsset struct {
	Asset                  string
	WalletBalance          float64
	UnrealizedProfit       float64
	MarginBalance          float64
	MaintMargin            float64
	InitialMargin          float64
	PositionInitialMargin  float64
	OpenOrderInitialMargin float64
	MaxWithdrawAmount      float64
	CrossWalletBalance     float64
	CrossUnPnl             float64
	AvailableBalance       float64
	MarginAvailable        bool
	UpdateTime             int64
}

// FuturesBalance represents futures balance
type FuturesBalance struct {
	Asset                string
	Balance              float64
	AvailableBalance     float64
	CrossWalletBalance   float64
	CrossUnPnl           float64
	MaxWithdrawAmount    float64
	MarginAvailable      bool
	UpdateTime           int64
}

// MarkPrice represents mark price information
type MarkPrice struct {
	Symbol          string
	MarkPrice       float64
	IndexPrice      float64
	LastFundingRate float64
	NextFundingTime int64
	Time            int64
}

// FundingRate represents funding rate information
type FundingRate struct {
	Symbol       string
	FundingRate  float64
	FundingTime  int64
}

// Position represents a futures position
type Position struct {
	Symbol              string
	PositionSide        PositionSide
	PositionAmt         float64
	EntryPrice          float64
	MarkPrice           float64
	UnrealizedProfit    float64
	LiquidationPrice    float64
	Leverage            int
	MarginType          MarginType
	IsolatedMargin      float64
	IsAutoAddMargin     bool
	PositionInitialMargin float64
	MaintenanceMargin   float64
	UpdateTime          int64
}

// FuturesOrderRequest represents a futures order request
type FuturesOrderRequest struct {
	Symbol           string
	Side             OrderSide
	PositionSide     PositionSide
	Type             OrderType
	Quantity         float64
	Price            float64
	StopPrice        float64
	TimeInForce      string
	ReduceOnly       bool
	ClosePosition    bool
}

// FuturesOrderResponse represents a futures order response
type FuturesOrderResponse struct {
	OrderID                 int64
	Symbol                  string
	Status                  OrderStatus
	Price                   float64
	AvgPrice                float64
	OrigQty                 float64
	ExecutedQty             float64
	CumQty                  float64
	CumQuote                float64
	TimeInForce             string
	Type                    OrderType
	ReduceOnly              bool
	ClosePosition           bool
	Side                    OrderSide
	PositionSide            PositionSide
	StopPrice               float64
	WorkingType             string
	PriceProtect            bool
	OrigType                OrderType
	UpdateTime              int64
}

// FuturesOrder represents a futures order
type FuturesOrder struct {
	OrderID                 int64
	Symbol                  string
	Side             OrderSide
	PositionSide     PositionSide
	Type             OrderType
	Status           OrderStatus
	Price            float64
	StopPrice        float64
	OrigQty          float64
	ExecutedQty      float64
	AvgPrice         float64
	ReduceOnly       bool
	ClosePosition    bool
	Time             int64
	UpdateTime       int64
}

// PositionMode represents position mode configuration
type PositionMode struct {
	DualSidePosition bool
}

// LeverageResponse represents leverage setting response
type LeverageResponse struct {
	Leverage         int
	MaxNotionalValue float64
	Symbol           string
}

// FuturesClient defines the interface for interacting with Binance Futures API
type FuturesClient interface {
	// Account information
	GetAccountInfo() (*FuturesAccountInfo, error)
	GetBalance() (*FuturesBalance, error)

	// Market data
	GetMarkPrice(symbol string) (*MarkPrice, error)
	GetPrice(symbol string) (*Price, error)
	GetKlines(symbol string, interval string, limit int) ([]*Kline, error)
	GetFundingRate(symbol string) (*FundingRate, error)
	GetFundingRateHistory(symbol string, startTime, endTime int64) ([]*FundingRate, error)

	// Leverage and margin
	SetLeverage(symbol string, leverage int) (*LeverageResponse, error)
	SetMarginType(symbol string, marginType MarginType) error
	SetPositionMode(dualSidePosition bool) error
	GetPositionMode() (*PositionMode, error)

	// Order operations
	CreateOrder(order *FuturesOrderRequest) (*FuturesOrderResponse, error)
	CancelOrder(symbol string, orderID int64) (*CancelResponse, error)
	GetOrder(symbol string, orderID int64) (*FuturesOrder, error)
	GetOpenOrders(symbol string) ([]*FuturesOrder, error)

	// Position queries
	GetPositions(symbol string) ([]*Position, error)
	GetAllPositions() ([]*Position, error)
}

// futuresClient implements FuturesClient interface
type futuresClient struct {
	baseURL    string
	httpClient HTTPClient
	authMgr    *AuthManager
}

// NewFuturesClient creates a new Binance Futures API client
func NewFuturesClient(baseURL string, httpClient HTTPClient, authMgr *AuthManager) (FuturesClient, error) {
	if err := ValidateURL(baseURL); err != nil {
		return nil, err
	}

	if authMgr == nil {
		return nil, fmt.Errorf("auth manager cannot be nil")
	}

	if err := authMgr.ValidateCredentials(); err != nil {
		return nil, err
	}

	// Validate that the URL is the futures endpoint
	if baseURL != "https://fapi.binance.com" && baseURL != "https://testnet.binancefuture.com" {
		return nil, fmt.Errorf("invalid futures API endpoint: %s", baseURL)
	}

	return &futuresClient{
		baseURL:    baseURL,
		httpClient: httpClient,
		authMgr:    authMgr,
	}, nil
}

// GetAccountInfo retrieves futures account information
func (c *futuresClient) GetAccountInfo() (*FuturesAccountInfo, error) {
	params := make(map[string]interface{})
	params["timestamp"] = c.authMgr.GenerateTimestamp()

	queryString, err := c.authMgr.SignRequestWithParams(params)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/fapi/v2/account?%s", c.baseURL, queryString)
	headers := map[string]string{
		"X-MBX-APIKEY": c.authMgr.GetAPIKey(),
	}

	body, err := c.httpClient.DoWithRetry("GET", url, nil, headers)
	if err != nil {
		return nil, err
	}

	var accountInfo FuturesAccountInfo
	if err := json.Unmarshal(body, &accountInfo); err != nil {
		return nil, fmt.Errorf("failed to parse futures account info: %w", err)
	}

	return &accountInfo, nil
}

// GetBalance retrieves the USDT balance for futures account
func (c *futuresClient) GetBalance() (*FuturesBalance, error) {
	accountInfo, err := c.GetAccountInfo()
	if err != nil {
		return nil, err
	}

	// Find USDT asset
	for _, asset := range accountInfo.Assets {
		if asset.Asset == "USDT" {
			return &FuturesBalance{
				Asset:                asset.Asset,
				Balance:              asset.WalletBalance,
				AvailableBalance:     asset.AvailableBalance,
				CrossWalletBalance:   asset.CrossWalletBalance,
				CrossUnPnl:           asset.CrossUnPnl,
				MaxWithdrawAmount:    asset.MaxWithdrawAmount,
				MarginAvailable:      asset.MarginAvailable,
				UpdateTime:           asset.UpdateTime,
			}, nil
		}
	}

	return nil, fmt.Errorf("USDT asset not found in futures account")
}

// GetMarkPrice retrieves mark price for a symbol
func (c *futuresClient) GetMarkPrice(symbol string) (*MarkPrice, error) {
	params := map[string]interface{}{
		"symbol": symbol,
	}

	url := fmt.Sprintf("%s/fapi/v1/premiumIndex", c.baseURL)

	body, err := c.httpClient.DoWithRetry("GET", url, params, nil)
	if err != nil {
		return nil, err
	}

	var markPrice MarkPrice
	if err := json.Unmarshal(body, &markPrice); err != nil {
		return nil, fmt.Errorf("failed to parse mark price: %w", err)
	}

	return &markPrice, nil
}

// GetPrice retrieves the current price for a symbol
func (c *futuresClient) GetPrice(symbol string) (*Price, error) {
	params := map[string]interface{}{
		"symbol": symbol,
	}

	url := fmt.Sprintf("%s/fapi/v1/ticker/price", c.baseURL)

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
func (c *futuresClient) GetKlines(symbol string, interval string, limit int) ([]*Kline, error) {
	params := map[string]interface{}{
		"symbol":   symbol,
		"interval": interval,
		"limit":    limit,
	}

	url := fmt.Sprintf("%s/fapi/v1/klines", c.baseURL)

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

// GetFundingRate retrieves current funding rate for a symbol
func (c *futuresClient) GetFundingRate(symbol string) (*FundingRate, error) {
	params := map[string]interface{}{
		"symbol": symbol,
	}

	url := fmt.Sprintf("%s/fapi/v1/premiumIndex", c.baseURL)

	body, err := c.httpClient.DoWithRetry("GET", url, params, nil)
	if err != nil {
		return nil, err
	}

	var data struct {
		Symbol          string  `json:"symbol"`
		MarkPrice       string  `json:"markPrice"`
		LastFundingRate string  `json:"lastFundingRate"`
		NextFundingTime int64   `json:"nextFundingTime"`
		Time            int64   `json:"time"`
	}

	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("failed to parse funding rate: %w", err)
	}

	var fundingRate float64
	if _, err := fmt.Sscanf(data.LastFundingRate, "%f", &fundingRate); err != nil {
		return nil, fmt.Errorf("failed to parse funding rate value: %w", err)
	}

	return &FundingRate{
		Symbol:      data.Symbol,
		FundingRate: fundingRate,
		FundingTime: data.NextFundingTime,
	}, nil
}

// GetFundingRateHistory retrieves funding rate history
func (c *futuresClient) GetFundingRateHistory(symbol string, startTime, endTime int64) ([]*FundingRate, error) {
	params := map[string]interface{}{
		"symbol": symbol,
	}
	if startTime > 0 {
		params["startTime"] = startTime
	}
	if endTime > 0 {
		params["endTime"] = endTime
	}

	url := fmt.Sprintf("%s/fapi/v1/fundingRate", c.baseURL)

	body, err := c.httpClient.DoWithRetry("GET", url, params, nil)
	if err != nil {
		return nil, err
	}

	var rawRates []struct {
		Symbol       string `json:"symbol"`
		FundingRate  string `json:"fundingRate"`
		FundingTime  int64  `json:"fundingTime"`
	}

	if err := json.Unmarshal(body, &rawRates); err != nil {
		return nil, fmt.Errorf("failed to parse funding rate history: %w", err)
	}

	rates := make([]*FundingRate, 0, len(rawRates))
	for _, raw := range rawRates {
		var rate float64
		fmt.Sscanf(raw.FundingRate, "%f", &rate)
		rates = append(rates, &FundingRate{
			Symbol:      raw.Symbol,
			FundingRate: rate,
			FundingTime: raw.FundingTime,
		})
	}

	return rates, nil
}

// SetLeverage sets leverage for a symbol
func (c *futuresClient) SetLeverage(symbol string, leverage int) (*LeverageResponse, error) {
	if leverage < 1 || leverage > 125 {
		return nil, fmt.Errorf("leverage must be between 1 and 125, got: %d", leverage)
	}

	params := make(map[string]interface{})
	params["symbol"] = symbol
	params["leverage"] = leverage
	params["timestamp"] = c.authMgr.GenerateTimestamp()

	queryString, err := c.authMgr.SignRequestWithParams(params)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/fapi/v1/leverage?%s", c.baseURL, queryString)
	headers := map[string]string{
		"X-MBX-APIKEY": c.authMgr.GetAPIKey(),
	}

	body, err := c.httpClient.DoWithRetry("POST", url, nil, headers)
	if err != nil {
		return nil, err
	}

	var response LeverageResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse leverage response: %w", err)
	}

	return &response, nil
}

// SetMarginType sets margin type for a symbol
func (c *futuresClient) SetMarginType(symbol string, marginType MarginType) error {
	params := make(map[string]interface{})
	params["symbol"] = symbol
	params["marginType"] = string(marginType)
	params["timestamp"] = c.authMgr.GenerateTimestamp()

	queryString, err := c.authMgr.SignRequestWithParams(params)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/fapi/v1/marginType?%s", c.baseURL, queryString)
	headers := map[string]string{
		"X-MBX-APIKEY": c.authMgr.GetAPIKey(),
	}

	_, err = c.httpClient.DoWithRetry("POST", url, nil, headers)
	return err
}

// SetPositionMode sets position mode (dual side or one way)
func (c *futuresClient) SetPositionMode(dualSidePosition bool) error {
	params := make(map[string]interface{})
	params["dualSidePosition"] = dualSidePosition
	params["timestamp"] = c.authMgr.GenerateTimestamp()

	queryString, err := c.authMgr.SignRequestWithParams(params)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/fapi/v1/positionSide/dual?%s", c.baseURL, queryString)
	headers := map[string]string{
		"X-MBX-APIKEY": c.authMgr.GetAPIKey(),
	}

	_, err = c.httpClient.DoWithRetry("POST", url, nil, headers)
	return err
}

// GetPositionMode retrieves current position mode
func (c *futuresClient) GetPositionMode() (*PositionMode, error) {
	params := make(map[string]interface{})
	params["timestamp"] = c.authMgr.GenerateTimestamp()

	queryString, err := c.authMgr.SignRequestWithParams(params)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/fapi/v1/positionSide/dual?%s", c.baseURL, queryString)
	headers := map[string]string{
		"X-MBX-APIKEY": c.authMgr.GetAPIKey(),
	}

	body, err := c.httpClient.DoWithRetry("GET", url, nil, headers)
	if err != nil {
		return nil, err
	}

	var mode PositionMode
	if err := json.Unmarshal(body, &mode); err != nil {
		return nil, fmt.Errorf("failed to parse position mode: %w", err)
	}

	return &mode, nil
}

// CreateOrder creates a new futures order
func (c *futuresClient) CreateOrder(order *FuturesOrderRequest) (*FuturesOrderResponse, error) {
	if order == nil {
		return nil, fmt.Errorf("order request cannot be nil")
	}

	params := make(map[string]interface{})
	params["symbol"] = order.Symbol
	params["side"] = string(order.Side)
	params["type"] = string(order.Type)
	params["quantity"] = order.Quantity
	params["timestamp"] = c.authMgr.GenerateTimestamp()

	if order.PositionSide != "" {
		params["positionSide"] = string(order.PositionSide)
	}

	if order.Type == OrderTypeLimit {
		if order.Price <= 0 {
			return nil, fmt.Errorf("price must be greater than 0 for limit orders")
		}
		params["price"] = order.Price
		if order.TimeInForce == "" {
			params["timeInForce"] = "GTC"
		} else {
			params["timeInForce"] = order.TimeInForce
		}
	}

	if order.StopPrice > 0 {
		params["stopPrice"] = order.StopPrice
	}

	if order.ReduceOnly {
		params["reduceOnly"] = "true"
	}

	if order.ClosePosition {
		params["closePosition"] = "true"
	}

	queryString, err := c.authMgr.SignRequestWithParams(params)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/fapi/v1/order?%s", c.baseURL, queryString)
	headers := map[string]string{
		"X-MBX-APIKEY": c.authMgr.GetAPIKey(),
	}

	body, err := c.httpClient.DoWithRetry("POST", url, nil, headers)
	if err != nil {
		return nil, err
	}

	var response FuturesOrderResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse futures order response: %w", err)
	}

	return &response, nil
}

// CancelOrder cancels a futures order
func (c *futuresClient) CancelOrder(symbol string, orderID int64) (*CancelResponse, error) {
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

	url := fmt.Sprintf("%s/fapi/v1/order?%s", c.baseURL, queryString)
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

// GetOrder retrieves futures order details
func (c *futuresClient) GetOrder(symbol string, orderID int64) (*FuturesOrder, error) {
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

	url := fmt.Sprintf("%s/fapi/v1/order?%s", c.baseURL, queryString)
	headers := map[string]string{
		"X-MBX-APIKEY": c.authMgr.GetAPIKey(),
	}

	body, err := c.httpClient.DoWithRetry("GET", url, nil, headers)
	if err != nil {
		return nil, err
	}

	var order FuturesOrder
	if err := json.Unmarshal(body, &order); err != nil {
		return nil, fmt.Errorf("failed to parse futures order: %w", err)
	}

	return &order, nil
}

// GetOpenOrders retrieves all open futures orders
func (c *futuresClient) GetOpenOrders(symbol string) ([]*FuturesOrder, error) {
	params := make(map[string]interface{})
	if symbol != "" {
		params["symbol"] = symbol
	}
	params["timestamp"] = c.authMgr.GenerateTimestamp()

	queryString, err := c.authMgr.SignRequestWithParams(params)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/fapi/v1/openOrders?%s", c.baseURL, queryString)
	headers := map[string]string{
		"X-MBX-APIKEY": c.authMgr.GetAPIKey(),
	}

	body, err := c.httpClient.DoWithRetry("GET", url, nil, headers)
	if err != nil {
		return nil, err
	}

	var orders []*FuturesOrder
	if err := json.Unmarshal(body, &orders); err != nil {
		return nil, fmt.Errorf("failed to parse open orders: %w", err)
	}

	return orders, nil
}

// GetPositions retrieves positions for a specific symbol
func (c *futuresClient) GetPositions(symbol string) ([]*Position, error) {
	params := make(map[string]interface{})
	params["timestamp"] = c.authMgr.GenerateTimestamp()

	queryString, err := c.authMgr.SignRequestWithParams(params)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/fapi/v2/positionRisk?%s", c.baseURL, queryString)
	headers := map[string]string{
		"X-MBX-APIKEY": c.authMgr.GetAPIKey(),
	}

	body, err := c.httpClient.DoWithRetry("GET", url, nil, headers)
	if err != nil {
		return nil, err
	}

	var allPositions []*Position
	if err := json.Unmarshal(body, &allPositions); err != nil {
		return nil, fmt.Errorf("failed to parse positions: %w", err)
	}

	// Filter by symbol if specified
	if symbol != "" {
		positions := make([]*Position, 0)
		for _, pos := range allPositions {
			if pos.Symbol == symbol {
				positions = append(positions, pos)
			}
		}
		return positions, nil
	}

	return allPositions, nil
}

// GetAllPositions retrieves all positions
func (c *futuresClient) GetAllPositions() ([]*Position, error) {
	return c.GetPositions("")
}
