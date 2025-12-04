package service

import (
	"binance-trader/internal/api"
	"binance-trader/internal/repository"
	"binance-trader/pkg/errors"
	"fmt"
	"testing"
)

// TestPlaceMarketBuyOrder_Success tests successful market buy order placement
func TestPlaceMarketBuyOrder_Success(t *testing.T) {
	// Setup
	mockClient := &mockBinanceClient{
		createOrderFunc: func(req *api.OrderRequest) (*api.OrderResponse, error) {
			if req.Symbol != "BTCUSDT" {
				t.Errorf("Expected symbol BTCUSDT, got %s", req.Symbol)
			}
			if req.Side != api.OrderSideBuy {
				t.Errorf("Expected side BUY, got %s", req.Side)
			}
			if req.Type != api.OrderTypeMarket {
				t.Errorf("Expected type MARKET, got %s", req.Type)
			}
			if req.Quantity != 0.1 {
				t.Errorf("Expected quantity 0.1, got %f", req.Quantity)
			}
			
			return &api.OrderResponse{
				OrderID:             12345,
				Symbol:              "BTCUSDT",
				Status:              api.OrderStatusFilled,
				Price:               50000.0,
				OrigQty:             0.1,
				ExecutedQty:         0.1,
				CummulativeQuoteQty: 5000.0,
				TransactTime:        1234567890,
			}, nil
		},
		getPriceFunc: func(symbol string) (*api.Price, error) {
			return &api.Price{Symbol: symbol, Price: 50000.0}, nil
		},
		getBalanceFunc: func(asset string) (*api.Balance, error) {
			return &api.Balance{Asset: asset, Free: 10000.0, Locked: 0}, nil
		},
	}
	
	riskMgr := NewRiskManager(&RiskLimits{
		MaxOrderAmount:    10000.0,
		MaxDailyOrders:    100,
		MinBalanceReserve: 100.0,
	}, mockClient)
	
	orderRepo := repository.NewMemoryOrderRepository()
	log := &mockLogger{}
	
	service := NewTradingService(mockClient, riskMgr, orderRepo, log)
	
	// Execute
	order, err := service.PlaceMarketBuyOrder("BTCUSDT", 0.1)
	
	// Verify
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if order == nil {
		t.Fatal("Expected order to be returned")
	}
	
	if order.OrderID != 12345 {
		t.Errorf("Expected order ID 12345, got %d", order.OrderID)
	}
	
	if order.Symbol != "BTCUSDT" {
		t.Errorf("Expected symbol BTCUSDT, got %s", order.Symbol)
	}
	
	if order.Side != api.OrderSideBuy {
		t.Errorf("Expected side BUY, got %s", order.Side)
	}
	
	if order.Type != api.OrderTypeMarket {
		t.Errorf("Expected type MARKET, got %s", order.Type)
	}
	
	// Verify order was saved to repository
	savedOrder, err := orderRepo.FindByID(12345)
	if err != nil {
		t.Errorf("Expected order to be saved in repository, got error: %v", err)
	}
	
	if savedOrder.OrderID != 12345 {
		t.Errorf("Expected saved order ID 12345, got %d", savedOrder.OrderID)
	}
}

// TestPlaceMarketBuyOrder_RiskValidationFailure tests market buy order with risk validation failure
func TestPlaceMarketBuyOrder_RiskValidationFailure(t *testing.T) {
	// Setup
	mockClient := &mockBinanceClient{
		getPriceFunc: func(symbol string) (*api.Price, error) {
			return &api.Price{Symbol: symbol, Price: 50000.0}, nil
		},
	}
	
	// Set very low max order amount to trigger risk validation failure
	riskMgr := NewRiskManager(&RiskLimits{
		MaxOrderAmount:    100.0, // Very low limit
		MaxDailyOrders:    100,
		MinBalanceReserve: 100.0,
	}, mockClient)
	
	orderRepo := repository.NewMemoryOrderRepository()
	log := &mockLogger{}
	
	service := NewTradingService(mockClient, riskMgr, orderRepo, log)
	
	// Execute
	order, err := service.PlaceMarketBuyOrder("BTCUSDT", 0.1)
	
	// Verify
	if err == nil {
		t.Fatal("Expected error due to risk validation failure")
	}
	
	if order != nil {
		t.Error("Expected no order to be returned")
	}
	
	tradingErr, ok := err.(*errors.TradingError)
	if !ok {
		t.Errorf("Expected TradingError, got %T", err)
	}
	
	if tradingErr.Type != errors.ErrRiskLimitExceeded {
		t.Errorf("Expected ErrRiskLimitExceeded, got %v", tradingErr.Type)
	}
}

// TestPlaceLimitSellOrder_Success tests successful limit sell order placement
func TestPlaceLimitSellOrder_Success(t *testing.T) {
	// Setup
	mockClient := &mockBinanceClient{
		createOrderFunc: func(req *api.OrderRequest) (*api.OrderResponse, error) {
			if req.Symbol != "BTCUSDT" {
				t.Errorf("Expected symbol BTCUSDT, got %s", req.Symbol)
			}
			if req.Side != api.OrderSideSell {
				t.Errorf("Expected side SELL, got %s", req.Side)
			}
			if req.Type != api.OrderTypeLimit {
				t.Errorf("Expected type LIMIT, got %s", req.Type)
			}
			if req.Price != 55000.0 {
				t.Errorf("Expected price 55000.0, got %f", req.Price)
			}
			if req.Quantity != 0.1 {
				t.Errorf("Expected quantity 0.1, got %f", req.Quantity)
			}
			
			return &api.OrderResponse{
				OrderID:             12346,
				Symbol:              "BTCUSDT",
				Status:              api.OrderStatusNew,
				Price:               55000.0,
				OrigQty:             0.1,
				ExecutedQty:         0.0,
				CummulativeQuoteQty: 0.0,
				TransactTime:        1234567890,
			}, nil
		},
	}
	
	riskMgr := NewRiskManager(&RiskLimits{
		MaxOrderAmount:    10000.0,
		MaxDailyOrders:    100,
		MinBalanceReserve: 100.0,
	}, mockClient)
	
	orderRepo := repository.NewMemoryOrderRepository()
	log := &mockLogger{}
	
	service := NewTradingService(mockClient, riskMgr, orderRepo, log)
	
	// Execute
	order, err := service.PlaceLimitSellOrder("BTCUSDT", 55000.0, 0.1)
	
	// Verify
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if order == nil {
		t.Fatal("Expected order to be returned")
	}
	
	if order.OrderID != 12346 {
		t.Errorf("Expected order ID 12346, got %d", order.OrderID)
	}
	
	if order.Symbol != "BTCUSDT" {
		t.Errorf("Expected symbol BTCUSDT, got %s", order.Symbol)
	}
	
	if order.Side != api.OrderSideSell {
		t.Errorf("Expected side SELL, got %s", order.Side)
	}
	
	if order.Type != api.OrderTypeLimit {
		t.Errorf("Expected type LIMIT, got %s", order.Type)
	}
	
	if order.Price != 55000.0 {
		t.Errorf("Expected price 55000.0, got %f", order.Price)
	}
}

// TestPlaceLimitSellOrder_InvalidPrice tests limit sell order with invalid price
func TestPlaceLimitSellOrder_InvalidPrice(t *testing.T) {
	// Setup
	mockClient := &mockBinanceClient{}
	riskMgr := NewRiskManager(nil, mockClient)
	orderRepo := repository.NewMemoryOrderRepository()
	log := &mockLogger{}
	
	service := NewTradingService(mockClient, riskMgr, orderRepo, log)
	
	// Execute
	order, err := service.PlaceLimitSellOrder("BTCUSDT", 0, 0.1)
	
	// Verify
	if err == nil {
		t.Fatal("Expected error due to invalid price")
	}
	
	if order != nil {
		t.Error("Expected no order to be returned")
	}
}

// TestCancelOrder_Success tests successful order cancellation
func TestCancelOrder_Success(t *testing.T) {
	// Setup
	mockClient := &mockBinanceClient{
		cancelOrderFunc: func(symbol string, orderID int64) (*api.CancelResponse, error) {
			if symbol != "BTCUSDT" {
				t.Errorf("Expected symbol BTCUSDT, got %s", symbol)
			}
			if orderID != 12345 {
				t.Errorf("Expected order ID 12345, got %d", orderID)
			}
			
			return &api.CancelResponse{
				Symbol:  "BTCUSDT",
				OrderID: 12345,
				Status:  api.OrderStatusCanceled,
			}, nil
		},
	}
	
	riskMgr := NewRiskManager(nil, mockClient)
	orderRepo := repository.NewMemoryOrderRepository()
	log := &mockLogger{}
	
	// Pre-populate repository with an order
	existingOrder := &api.Order{
		OrderID: 12345,
		Symbol:  "BTCUSDT",
		Side:    api.OrderSideBuy,
		Type:    api.OrderTypeLimit,
		Status:  api.OrderStatusNew,
		Price:   50000.0,
		OrigQty: 0.1,
	}
	orderRepo.Save(existingOrder)
	
	service := NewTradingService(mockClient, riskMgr, orderRepo, log)
	
	// Execute
	err := service.CancelOrder(12345)
	
	// Verify
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	// Verify order status was updated in repository
	updatedOrder, err := orderRepo.FindByID(12345)
	if err != nil {
		t.Fatalf("Expected to find order in repository, got error: %v", err)
	}
	
	if updatedOrder.Status != api.OrderStatusCanceled {
		t.Errorf("Expected status CANCELED, got %s", updatedOrder.Status)
	}
}

// TestCancelOrder_OrderNotFound tests canceling a non-existent order
func TestCancelOrder_OrderNotFound(t *testing.T) {
	// Setup
	mockClient := &mockBinanceClient{}
	riskMgr := NewRiskManager(nil, mockClient)
	orderRepo := repository.NewMemoryOrderRepository()
	log := &mockLogger{}
	
	service := NewTradingService(mockClient, riskMgr, orderRepo, log)
	
	// Execute
	err := service.CancelOrder(99999)
	
	// Verify
	if err == nil {
		t.Fatal("Expected error for non-existent order")
	}
	
	tradingErr, ok := err.(*errors.TradingError)
	if !ok {
		t.Errorf("Expected TradingError, got %T", err)
	}
	
	if tradingErr.Type != errors.ErrOrderNotFound {
		t.Errorf("Expected ErrOrderNotFound, got %v", tradingErr.Type)
	}
}

// TestGetOrderStatus_Success tests successful order status retrieval
func TestGetOrderStatus_Success(t *testing.T) {
	// Setup
	mockClient := &mockBinanceClient{
		getOrderFunc: func(symbol string, orderID int64) (*api.Order, error) {
			return &api.Order{
				OrderID:     12345,
				Symbol:      "BTCUSDT",
				Status:      api.OrderStatusPartiallyFilled,
				ExecutedQty: 0.05,
				Price:       50000.0,
				UpdateTime:  1234567890,
			}, nil
		},
	}
	
	riskMgr := NewRiskManager(nil, mockClient)
	orderRepo := repository.NewMemoryOrderRepository()
	log := &mockLogger{}
	
	// Pre-populate repository with an order
	existingOrder := &api.Order{
		OrderID: 12345,
		Symbol:  "BTCUSDT",
		Status:  api.OrderStatusNew,
	}
	orderRepo.Save(existingOrder)
	
	service := NewTradingService(mockClient, riskMgr, orderRepo, log)
	
	// Execute
	status, err := service.GetOrderStatus(12345)
	
	// Verify
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if status == nil {
		t.Fatal("Expected status to be returned")
	}
	
	if status.OrderID != 12345 {
		t.Errorf("Expected order ID 12345, got %d", status.OrderID)
	}
	
	if status.Status != api.OrderStatusPartiallyFilled {
		t.Errorf("Expected status PARTIALLY_FILLED, got %s", status.Status)
	}
	
	if status.ExecutedQty != 0.05 {
		t.Errorf("Expected executed qty 0.05, got %f", status.ExecutedQty)
	}
	
	// Verify repository was synced
	updatedOrder, _ := orderRepo.FindByID(12345)
	if updatedOrder.Status != api.OrderStatusPartiallyFilled {
		t.Errorf("Expected repository status to be synced to PARTIALLY_FILLED, got %s", updatedOrder.Status)
	}
}

// TestGetActiveOrders_Success tests successful retrieval of active orders
func TestGetActiveOrders_Success(t *testing.T) {
	// Setup
	mockClient := &mockBinanceClient{
		getOpenOrdersFunc: func(symbol string) ([]*api.Order, error) {
			return []*api.Order{
				{
					OrderID:     12345,
					Symbol:      "BTCUSDT",
					Status:      api.OrderStatusNew,
					ExecutedQty: 0.0,
					Price:       50000.0,
				},
				{
					OrderID:     12346,
					Symbol:      "ETHUSDT",
					Status:      api.OrderStatusPartiallyFilled,
					ExecutedQty: 0.5,
					Price:       3000.0,
				},
			}, nil
		},
	}
	
	riskMgr := NewRiskManager(nil, mockClient)
	orderRepo := repository.NewMemoryOrderRepository()
	log := &mockLogger{}
	
	service := NewTradingService(mockClient, riskMgr, orderRepo, log)
	
	// Execute
	orders, err := service.GetActiveOrders()
	
	// Verify
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if len(orders) != 2 {
		t.Fatalf("Expected 2 orders, got %d", len(orders))
	}
	
	if orders[0].OrderID != 12345 {
		t.Errorf("Expected first order ID 12345, got %d", orders[0].OrderID)
	}
	
	if orders[1].OrderID != 12346 {
		t.Errorf("Expected second order ID 12346, got %d", orders[1].OrderID)
	}
	
	// Verify orders were saved to repository
	savedOrder1, err := orderRepo.FindByID(12345)
	if err != nil {
		t.Errorf("Expected order 12345 to be saved in repository")
	} else if savedOrder1.OrderID != 12345 {
		t.Errorf("Expected saved order ID 12345, got %d", savedOrder1.OrderID)
	}
	
	savedOrder2, err := orderRepo.FindByID(12346)
	if err != nil {
		t.Errorf("Expected order 12346 to be saved in repository")
	} else if savedOrder2.OrderID != 12346 {
		t.Errorf("Expected saved order ID 12346, got %d", savedOrder2.OrderID)
	}
}

// TestGetActiveOrders_APIError tests error handling when API fails
func TestGetActiveOrders_APIError(t *testing.T) {
	// Setup
	mockClient := &mockBinanceClient{
		getOpenOrdersFunc: func(symbol string) ([]*api.Order, error) {
			return nil, fmt.Errorf("API connection failed")
		},
	}
	
	riskMgr := NewRiskManager(nil, mockClient)
	orderRepo := repository.NewMemoryOrderRepository()
	log := &mockLogger{}
	
	service := NewTradingService(mockClient, riskMgr, orderRepo, log)
	
	// Execute
	orders, err := service.GetActiveOrders()
	
	// Verify
	if err == nil {
		t.Fatal("Expected error due to API failure")
	}
	
	if orders != nil {
		t.Error("Expected no orders to be returned")
	}
}
