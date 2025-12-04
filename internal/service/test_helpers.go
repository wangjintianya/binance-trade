package service

import (
	"binance-trader/internal/api"
	"fmt"
)

// mockBinanceClient is a mock implementation of BinanceClient for testing
type mockBinanceClient struct {
	getPriceFunc            func(symbol string) (*api.Price, error)
	getKlinesFunc           func(symbol string, interval string, limit int) ([]*api.Kline, error)
	getAccountInfoFunc      func() (*api.AccountInfo, error)
	getBalanceFunc          func(asset string) (*api.Balance, error)
	createOrderFunc         func(order *api.OrderRequest) (*api.OrderResponse, error)
	cancelOrderFunc         func(symbol string, orderID int64) (*api.CancelResponse, error)
	getOrderFunc            func(symbol string, orderID int64) (*api.Order, error)
	getOpenOrdersFunc       func(symbol string) ([]*api.Order, error)
	getHistoricalOrdersFunc func(symbol string, startTime, endTime int64) ([]*api.Order, error)
}

func (m *mockBinanceClient) GetPrice(symbol string) (*api.Price, error) {
	if m.getPriceFunc != nil {
		return m.getPriceFunc(symbol)
	}
	return nil, fmt.Errorf("not implemented")
}

func (m *mockBinanceClient) GetKlines(symbol string, interval string, limit int) ([]*api.Kline, error) {
	if m.getKlinesFunc != nil {
		return m.getKlinesFunc(symbol, interval, limit)
	}
	return nil, fmt.Errorf("not implemented")
}

func (m *mockBinanceClient) GetAccountInfo() (*api.AccountInfo, error) {
	if m.getAccountInfoFunc != nil {
		return m.getAccountInfoFunc()
	}
	return nil, fmt.Errorf("not implemented")
}

func (m *mockBinanceClient) GetBalance(asset string) (*api.Balance, error) {
	if m.getBalanceFunc != nil {
		return m.getBalanceFunc(asset)
	}
	return nil, fmt.Errorf("not implemented")
}

func (m *mockBinanceClient) CreateOrder(order *api.OrderRequest) (*api.OrderResponse, error) {
	if m.createOrderFunc != nil {
		return m.createOrderFunc(order)
	}
	return nil, fmt.Errorf("not implemented")
}

func (m *mockBinanceClient) CancelOrder(symbol string, orderID int64) (*api.CancelResponse, error) {
	if m.cancelOrderFunc != nil {
		return m.cancelOrderFunc(symbol, orderID)
	}
	return nil, fmt.Errorf("not implemented")
}

func (m *mockBinanceClient) GetOrder(symbol string, orderID int64) (*api.Order, error) {
	if m.getOrderFunc != nil {
		return m.getOrderFunc(symbol, orderID)
	}
	return nil, fmt.Errorf("not implemented")
}

func (m *mockBinanceClient) GetOpenOrders(symbol string) ([]*api.Order, error) {
	if m.getOpenOrdersFunc != nil {
		return m.getOpenOrdersFunc(symbol)
	}
	return nil, fmt.Errorf("not implemented")
}

func (m *mockBinanceClient) GetHistoricalOrders(symbol string, startTime, endTime int64) ([]*api.Order, error) {
	if m.getHistoricalOrdersFunc != nil {
		return m.getHistoricalOrdersFunc(symbol, startTime, endTime)
	}
	return nil, fmt.Errorf("not implemented")
}

// mockLogger is a mock implementation of Logger for testing
type mockLogger struct{}

func (m *mockLogger) Debug(msg string, fields map[string]interface{})                                                                {}
func (m *mockLogger) Info(msg string, fields map[string]interface{})                                                                 {}
func (m *mockLogger) Warn(msg string, fields map[string]interface{})                                                                 {}
func (m *mockLogger) Error(msg string, fields map[string]interface{})                                                                {}
func (m *mockLogger) Fatal(msg string, fields map[string]interface{})                                                                {}
func (m *mockLogger) LogAPIOperation(operationType string, result string, fields map[string]interface{})                             {}
func (m *mockLogger) LogOrderEvent(eventType string, orderID int64, symbol, side, orderType string, quantity float64, fields map[string]interface{}) {
}
func (m *mockLogger) LogError(err error, context map[string]interface{}) {}
