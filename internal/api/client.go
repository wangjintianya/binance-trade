package api

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

// BinanceClient is an alias for SpotClient for backward compatibility
// Deprecated: Use SpotClient instead
type BinanceClient = SpotClient

// HTTPClient defines the interface for making HTTP requests
type HTTPClient interface {
	Do(method, url string, params map[string]interface{}, headers map[string]string) ([]byte, error)
	DoWithRetry(method, url string, params map[string]interface{}, headers map[string]string) ([]byte, error)
}

// NewBinanceClient creates a new Binance API client
// Deprecated: Use NewSpotClient instead
func NewBinanceClient(baseURL string, httpClient HTTPClient, authMgr *AuthManager) (BinanceClient, error) {
	return NewSpotClient(baseURL, httpClient, authMgr)
}
