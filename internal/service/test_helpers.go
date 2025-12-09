package service

import (
	"binance-trader/internal/api"
	"binance-trader/internal/repository"
	"time"
)

// Shared mock implementations for testing

type mockFuturesClientShared struct {
	markPrice    float64
	lastPrice    float64
	fundingRate  *api.FundingRate
	positions    []*api.Position
	orders       []*api.FuturesOrder
	nextOrderID  int64
}

func (m *mockFuturesClientShared) GetMarkPrice(symbol string) (*api.MarkPrice, error) {
	return &api.MarkPrice{
		Symbol:    symbol,
		MarkPrice: m.markPrice,
	}, nil
}

func (m *mockFuturesClientShared) GetPrice(symbol string) (*api.Price, error) {
	return &api.Price{
		Symbol: symbol,
		Price:  m.lastPrice,
	}, nil
}

func (m *mockFuturesClientShared) GetFundingRate(symbol string) (*api.FundingRate, error) {
	if m.fundingRate == nil {
		return &api.FundingRate{
			Symbol:      symbol,
			FundingRate: 0.0001,
			FundingTime: time.Now().Unix(),
		}, nil
	}
	return m.fundingRate, nil
}

func (m *mockFuturesClientShared) GetPositions(symbol string) ([]*api.Position, error) {
	var result []*api.Position
	for _, pos := range m.positions {
		if pos.Symbol == symbol {
			result = append(result, pos)
		}
	}
	return result, nil
}

func (m *mockFuturesClientShared) GetAllPositions() ([]*api.Position, error) {
	return m.positions, nil
}

func (m *mockFuturesClientShared) CreateOrder(order *api.FuturesOrderRequest) (*api.FuturesOrderResponse, error) {
	m.nextOrderID++
	return &api.FuturesOrderResponse{
		OrderID: m.nextOrderID,
		Symbol:  order.Symbol,
		Status:  api.OrderStatusFilled,
	}, nil
}

func (m *mockFuturesClientShared) GetAccountInfo() (*api.FuturesAccountInfo, error) { return nil, nil }
func (m *mockFuturesClientShared) GetBalance() (*api.FuturesBalance, error)         { return nil, nil }
func (m *mockFuturesClientShared) GetKlines(symbol string, interval string, limit int) ([]*api.Kline, error) {
	return nil, nil
}
func (m *mockFuturesClientShared) GetFundingRateHistory(symbol string, startTime, endTime int64) ([]*api.FundingRate, error) {
	return nil, nil
}
func (m *mockFuturesClientShared) SetLeverage(symbol string, leverage int) (*api.LeverageResponse, error) {
	return &api.LeverageResponse{
		Leverage: leverage,
		Symbol:   symbol,
	}, nil
}
func (m *mockFuturesClientShared) SetMarginType(symbol string, marginType api.MarginType) error {
	return nil
}
func (m *mockFuturesClientShared) SetPositionMode(dualSidePosition bool) error { return nil }
func (m *mockFuturesClientShared) GetPositionMode() (*api.PositionMode, error) { return nil, nil }
func (m *mockFuturesClientShared) CancelOrder(symbol string, orderID int64) (*api.CancelResponse, error) {
	return nil, nil
}
func (m *mockFuturesClientShared) GetOrder(symbol string, orderID int64) (*api.FuturesOrder, error) {
	return nil, nil
}
func (m *mockFuturesClientShared) GetOpenOrders(symbol string) ([]*api.FuturesOrder, error) {
	return nil, nil
}
func (m *mockFuturesClientShared) GetPositionRisk(symbol string) (interface{}, error) {
	return nil, nil
}
func (m *mockFuturesClientShared) GetIncomeHistory(symbol string, incomeType string, startTime, endTime int64) ([]interface{}, error) {
	return nil, nil
}

type mockFuturesMarketDataServiceShared struct {
	markPrice   float64
	lastPrice   float64
	fundingRate *api.FundingRate
}

func (m *mockFuturesMarketDataServiceShared) GetMarkPrice(symbol string) (float64, error) {
	return m.markPrice, nil
}

func (m *mockFuturesMarketDataServiceShared) GetLastPrice(symbol string) (float64, error) {
	return m.lastPrice, nil
}

func (m *mockFuturesMarketDataServiceShared) GetHistoricalData(symbol string, interval string, limit int) ([]*api.Kline, error) {
	return nil, nil
}

func (m *mockFuturesMarketDataServiceShared) GetFundingRate(symbol string) (*api.FundingRate, error) {
	if m.fundingRate == nil {
		return &api.FundingRate{
			Symbol:      symbol,
			FundingRate: 0.0001,
			FundingTime: time.Now().Unix(),
		}, nil
	}
	return m.fundingRate, nil
}

func (m *mockFuturesMarketDataServiceShared) GetFundingRateHistory(symbol string, startTime, endTime int64) ([]*api.FundingRate, error) {
	return nil, nil
}

func (m *mockFuturesMarketDataServiceShared) SubscribeToMarkPrice(symbol string, callback func(float64)) error {
	return nil
}

type mockFuturesPositionManagerShared struct {
	positions map[string]*api.Position
}

func (m *mockFuturesPositionManagerShared) GetPosition(symbol string, positionSide api.PositionSide) (*api.Position, error) {
	key := symbol + string(positionSide)
	if pos, exists := m.positions[key]; exists {
		return pos, nil
	}
	return &api.Position{
		Symbol:           symbol,
		PositionSide:     positionSide,
		PositionAmt:      0,
		EntryPrice:       0,
		UnrealizedProfit: 0,
	}, nil
}

func (m *mockFuturesPositionManagerShared) GetAllPositions() ([]*api.Position, error) {
	var result []*api.Position
	for _, pos := range m.positions {
		result = append(result, pos)
	}
	return result, nil
}

func (m *mockFuturesPositionManagerShared) CalculateUnrealizedPnL(position *api.Position, markPrice float64) (float64, error) {
	return (markPrice - position.EntryPrice) * position.PositionAmt, nil
}

func (m *mockFuturesPositionManagerShared) CalculateLiquidationPrice(position *api.Position) (float64, error) {
	return 0, nil
}

func (m *mockFuturesPositionManagerShared) CalculateMarginRatio(position *api.Position) (float64, error) {
	return 0, nil
}

func (m *mockFuturesPositionManagerShared) UpdatePosition(symbol string) error {
	return nil
}

func (m *mockFuturesPositionManagerShared) UpdateAllPositions() error {
	return nil
}

func (m *mockFuturesPositionManagerShared) GetPositionHistory(symbol string, startTime, endTime int64) ([]*repository.ClosedPosition, error) {
	return nil, nil
}

type mockFuturesTradingServiceShared struct {
	orders []*api.FuturesOrder
}

func (m *mockFuturesTradingServiceShared) OpenLongPosition(symbol string, quantity float64, orderType api.OrderType, price float64) (*api.FuturesOrder, error) {
	order := &api.FuturesOrder{
		OrderID:      int64(len(m.orders) + 1),
		Symbol:       symbol,
		Side:         api.OrderSideBuy,
		PositionSide: api.PositionSideLong,
		Type:         orderType,
		OrigQty:      quantity,
		Price:        price,
		Status:       api.OrderStatusFilled,
	}
	m.orders = append(m.orders, order)
	return order, nil
}

func (m *mockFuturesTradingServiceShared) OpenShortPosition(symbol string, quantity float64, orderType api.OrderType, price float64) (*api.FuturesOrder, error) {
	order := &api.FuturesOrder{
		OrderID:      int64(len(m.orders) + 1),
		Symbol:       symbol,
		Side:         api.OrderSideSell,
		PositionSide: api.PositionSideShort,
		Type:         orderType,
		OrigQty:      quantity,
		Price:        price,
		Status:       api.OrderStatusFilled,
	}
	m.orders = append(m.orders, order)
	return order, nil
}

func (m *mockFuturesTradingServiceShared) ClosePosition(symbol string, positionSide api.PositionSide, quantity float64) (*api.FuturesOrder, error) {
	return nil, nil
}

func (m *mockFuturesTradingServiceShared) CloseAllPositions(symbol string) ([]*api.FuturesOrder, error) {
	return nil, nil
}

func (m *mockFuturesTradingServiceShared) CancelOrder(symbol string, orderID int64) error {
	return nil
}

func (m *mockFuturesTradingServiceShared) GetOrderStatus(orderID int64) (*api.FuturesOrder, error) {
	return nil, nil
}

func (m *mockFuturesTradingServiceShared) GetActiveOrders(symbol string) ([]*api.FuturesOrder, error) {
	return nil, nil
}

func (m *mockFuturesTradingServiceShared) SetLeverage(symbol string, leverage int) (*api.LeverageResponse, error) {
	return &api.LeverageResponse{
		Leverage: leverage,
		Symbol:   symbol,
	}, nil
}

func (m *mockFuturesTradingServiceShared) GetLeverage(symbol string) (int, error) {
	return 1, nil
}

// mockLogger is a simple mock logger for testing
type mockLogger struct{}

func (m *mockLogger) Debug(msg string, fields map[string]interface{})                                                                                {}
func (m *mockLogger) Info(msg string, fields map[string]interface{})                                                                                 {}
func (m *mockLogger) Warn(msg string, fields map[string]interface{})                                                                                 {}
func (m *mockLogger) Error(msg string, fields map[string]interface{})                                                                                {}
func (m *mockLogger) Fatal(msg string, fields map[string]interface{})                                                                                {}
func (m *mockLogger) LogAPIOperation(operationType string, result string, fields map[string]interface{})                                             {}
func (m *mockLogger) LogOrderEvent(eventType string, orderID int64, symbol, side, orderType string, quantity float64, fields map[string]interface{}) {}
func (m *mockLogger) LogError(err error, context map[string]interface{})                                                                             {}
func (m *mockLogger) LogFuturesAPIOperation(operationType string, result string, fields map[string]interface{})                                      {}
func (m *mockLogger) LogFuturesOrderEvent(eventType string, orderID int64, symbol, side, orderType string, quantity float64, positionChange map[string]interface{}, fields map[string]interface{}) {}
func (m *mockLogger) LogLiquidationEvent(symbol string, positionSide string, liquidationPrice float64, lossAmount float64, reason string, fields map[string]interface{}) {}
func (m *mockLogger) LogFundingRateSettlement(symbol string, fundingFee float64, fundingRate float64, positionSize float64, fields map[string]interface{}) {}
func (m *mockLogger) SetTradingType(tradingType string)                                                                                              {}

// mockBinanceClient is a mock for spot trading client
type mockBinanceClient struct {
	getPriceFunc      func(symbol string) (*api.Price, error)
	getKlinesFunc     func(symbol string, interval string, limit int) ([]*api.Kline, error)
	getBalanceFunc    func(asset string) (*api.Balance, error)
	createOrderFunc   func(order *api.OrderRequest) (*api.OrderResponse, error)
	cancelOrderFunc   func(symbol string, orderID int64) (*api.CancelResponse, error)
	getOrderFunc      func(symbol string, orderID int64) (*api.Order, error)
	getOpenOrdersFunc func(symbol string) ([]*api.Order, error)
}

func (m *mockBinanceClient) GetPrice(symbol string) (*api.Price, error) {
	if m.getPriceFunc != nil {
		return m.getPriceFunc(symbol)
	}
	return &api.Price{Symbol: symbol, Price: 50000.0}, nil
}

func (m *mockBinanceClient) GetKlines(symbol string, interval string, limit int) ([]*api.Kline, error) {
	if m.getKlinesFunc != nil {
		return m.getKlinesFunc(symbol, interval, limit)
	}
	return nil, nil
}

func (m *mockBinanceClient) GetBalance(asset string) (*api.Balance, error) {
	if m.getBalanceFunc != nil {
		return m.getBalanceFunc(asset)
	}
	return &api.Balance{Asset: asset, Free: 1000.0, Locked: 0}, nil
}

func (m *mockBinanceClient) CreateOrder(order *api.OrderRequest) (*api.OrderResponse, error) {
	if m.createOrderFunc != nil {
		return m.createOrderFunc(order)
	}
	return &api.OrderResponse{OrderID: 12345, Symbol: order.Symbol, Status: api.OrderStatusFilled}, nil
}

func (m *mockBinanceClient) CancelOrder(symbol string, orderID int64) (*api.CancelResponse, error) {
	if m.cancelOrderFunc != nil {
		return m.cancelOrderFunc(symbol, orderID)
	}
	return &api.CancelResponse{OrderID: orderID, Symbol: symbol}, nil
}

func (m *mockBinanceClient) GetOrder(symbol string, orderID int64) (*api.Order, error) {
	if m.getOrderFunc != nil {
		return m.getOrderFunc(symbol, orderID)
	}
	return nil, nil
}

func (m *mockBinanceClient) GetOpenOrders(symbol string) ([]*api.Order, error) {
	if m.getOpenOrdersFunc != nil {
		return m.getOpenOrdersFunc(symbol)
	}
	return nil, nil
}

func (m *mockBinanceClient) GetAccountInfo() (*api.AccountInfo, error) {
	return &api.AccountInfo{}, nil
}

func (m *mockBinanceClient) GetHistoricalOrders(symbol string, startTime, endTime int64) ([]*api.Order, error) {
	return nil, nil
}
