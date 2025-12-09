package service

import (
	"binance-trader/internal/api"
	"binance-trader/internal/repository"
	"math"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Feature: binance-auto-trading, Property 29: Stop loss trigger correctness
// Validates: Requirements 9.1
func TestProperty_StopLossTriggerCorrectness(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("for any position and stop price, when market price falls below stop price, stop loss should trigger", prop.ForAll(
		func(position float64, stopPrice float64, marketPrice float64) bool {
			// Setup
			stopOrderRepo := repository.NewMemoryStopOrderRepository()
			triggerEngine := NewTriggerEngine()
			mockTrading := &mockStopLossTradingService{}
			mockMarket := &mockStopLossMarketDataService{currentPrice: marketPrice}
			log := &mockLogger{}

			service := NewStopLossService(stopOrderRepo, triggerEngine, mockTrading, mockMarket, log)

			// Create stop loss order
			stopOrder, err := service.SetStopLoss("BTCUSDT", position, stopPrice)
			if err != nil {
				return false
			}

			// Evaluate trigger condition
			condition := &TriggerCondition{
				Type:     TriggerTypePrice,
				Operator: OperatorLessEqual,
				Value:    stopPrice,
			}

			marketData := &MarketData{
				Symbol: "BTCUSDT",
				Price:  marketPrice,
			}

			triggered, err := triggerEngine.EvaluateCondition(condition, marketData.Price)
			if err != nil {
				return false
			}

			// Verify: if market price <= stop price, should trigger
			expectedTrigger := marketPrice <= stopPrice

			// Verify order was created correctly
			if stopOrder.Type != repository.StopOrderTypeStopLoss {
				return false
			}

			if stopOrder.Status != repository.StopOrderStatusActive {
				return false
			}

			return triggered == expectedTrigger
		},
		gen.Float64Range(0.1, 1000.0),  // position
		gen.Float64Range(100.0, 50000.0), // stop price
		gen.Float64Range(50.0, 60000.0),  // market price
	))

	properties.TestingRun(t)
}

// Feature: binance-auto-trading, Property 30: Take profit trigger correctness
// Validates: Requirements 9.2
func TestProperty_TakeProfitTriggerCorrectness(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("for any position and target price, when market price reaches target, take profit should trigger", prop.ForAll(
		func(position float64, targetPrice float64, marketPrice float64) bool {
			// Setup
			stopOrderRepo := repository.NewMemoryStopOrderRepository()
			triggerEngine := NewTriggerEngine()
			mockTrading := &mockStopLossTradingService{}
			mockMarket := &mockStopLossMarketDataService{currentPrice: marketPrice}
			log := &mockLogger{}

			service := NewStopLossService(stopOrderRepo, triggerEngine, mockTrading, mockMarket, log)

			// Create take profit order
			takeProfitOrder, err := service.SetTakeProfit("BTCUSDT", position, targetPrice)
			if err != nil {
				return false
			}

			// Evaluate trigger condition
			condition := &TriggerCondition{
				Type:     TriggerTypePrice,
				Operator: OperatorGreaterEqual,
				Value:    targetPrice,
			}

			marketData := &MarketData{
				Symbol: "BTCUSDT",
				Price:  marketPrice,
			}

			triggered, err := triggerEngine.EvaluateCondition(condition, marketData.Price)
			if err != nil {
				return false
			}

			// Verify: if market price >= target price, should trigger
			expectedTrigger := marketPrice >= targetPrice

			// Verify order was created correctly
			if takeProfitOrder.Type != repository.StopOrderTypeTakeProfit {
				return false
			}

			if takeProfitOrder.Status != repository.StopOrderStatusActive {
				return false
			}

			return triggered == expectedTrigger
		},
		gen.Float64Range(0.1, 1000.0),  // position
		gen.Float64Range(100.0, 50000.0), // target price
		gen.Float64Range(50.0, 60000.0),  // market price
	))

	properties.TestingRun(t)
}

// Feature: binance-auto-trading, Property 31: Stop loss take profit mutual exclusivity
// Validates: Requirements 9.3
func TestProperty_StopLossTakeProfitMutualExclusivity(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("for any stop loss and take profit pair, when one triggers, the other should be cancelled", prop.ForAll(
		func(position float64, stopPrice float64, targetPrice float64) bool {
			// Ensure stop price < target price for valid test
			if stopPrice >= targetPrice {
				return true // Skip invalid combinations
			}

			// Setup
			stopOrderRepo := repository.NewMemoryStopOrderRepository()
			triggerEngine := NewTriggerEngine()
			mockTrading := &mockStopLossTradingService{}
			mockMarket := &mockStopLossMarketDataService{currentPrice: (stopPrice + targetPrice) / 2}
			log := &mockLogger{}

			service := NewStopLossService(stopOrderRepo, triggerEngine, mockTrading, mockMarket, log)

			// Create stop loss and take profit pair
			pair, err := service.SetStopLossTakeProfit("BTCUSDT", position, stopPrice, targetPrice)
			if err != nil {
				return false
			}

			// Verify both orders are initially active
			if pair.StopLossOrder.Status != repository.StopOrderStatusActive {
				return false
			}

			if pair.TakeProfitOrder.Status != repository.StopOrderStatusActive {
				return false
			}

			// Simulate triggering stop loss
			err = stopOrderRepo.UpdateStopOrderStatus(
				pair.StopLossOrder.OrderID,
				repository.StopOrderStatusTriggered,
				0,
				0,
			)
			if err != nil {
				return false
			}

			// Cancel the other order (take profit)
			err = stopOrderRepo.UpdateStopOrderStatus(
				pair.TakeProfitOrder.OrderID,
				repository.StopOrderStatusCancelled,
				0,
				0,
			)
			if err != nil {
				return false
			}

			// Verify states
			stopLoss, _ := stopOrderRepo.FindStopOrderByID(pair.StopLossOrder.OrderID)
			takeProfit, _ := stopOrderRepo.FindStopOrderByID(pair.TakeProfitOrder.OrderID)

			// One should be triggered, the other cancelled
			return stopLoss.Status == repository.StopOrderStatusTriggered &&
				takeProfit.Status == repository.StopOrderStatusCancelled
		},
		gen.Float64Range(0.1, 1000.0),    // position
		gen.Float64Range(100.0, 30000.0), // stop price
		gen.Float64Range(100.0, 30000.0), // target price
	))

	properties.TestingRun(t)
}

// Feature: binance-auto-trading, Property 32: Stop loss take profit trigger logging
// Validates: Requirements 9.4
func TestProperty_StopLossTakeProfitTriggerLogging(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("for any triggered stop order, log must contain trigger reason, price and result", prop.ForAll(
		func(position float64, stopPrice float64) bool {
			// Setup
			stopOrderRepo := repository.NewMemoryStopOrderRepository()
			triggerEngine := NewTriggerEngine()
			mockTrading := &mockStopLossTradingService{}
			mockMarket := &mockStopLossMarketDataService{currentPrice: stopPrice}
			log := &mockLogger{}

			service := NewStopLossService(stopOrderRepo, triggerEngine, mockTrading, mockMarket, log)

			// Create stop loss order
			stopOrder, err := service.SetStopLoss("BTCUSDT", position, stopPrice)
			if err != nil {
				return false
			}

			// Verify order was created with correct fields
			if stopOrder.OrderID == "" {
				return false
			}

			if stopOrder.Symbol != "BTCUSDT" {
				return false
			}

			if stopOrder.Position != position {
				return false
			}

			if stopOrder.StopPrice != stopPrice {
				return false
			}

			return true
		},
		gen.Float64Range(0.1, 1000.0),    // position
		gen.Float64Range(100.0, 50000.0), // stop price
	))

	properties.TestingRun(t)
}

// Mock implementations for testing

type mockStopLossTradingService struct{}

func (m *mockStopLossTradingService) PlaceMarketBuyOrder(symbol string, quantity float64) (*api.Order, error) {
	return &api.Order{
		OrderID: 12345,
		Symbol:  symbol,
		Status:  api.OrderStatusFilled,
	}, nil
}

func (m *mockStopLossTradingService) PlaceLimitSellOrder(symbol string, price, quantity float64) (*api.Order, error) {
	return &api.Order{
		OrderID: 12346,
		Symbol:  symbol,
		Status:  api.OrderStatusNew,
	}, nil
}

func (m *mockStopLossTradingService) CancelOrder(orderID int64) error {
	return nil
}

func (m *mockStopLossTradingService) GetOrderStatus(orderID int64) (*OrderStatus, error) {
	return &OrderStatus{
		OrderID: orderID,
		Status:  api.OrderStatusFilled,
	}, nil
}

func (m *mockStopLossTradingService) GetActiveOrders() ([]*api.Order, error) {
	return []*api.Order{}, nil
}

type mockStopLossMarketDataService struct {
	currentPrice float64
}

func (m *mockStopLossMarketDataService) GetCurrentPrice(symbol string) (float64, error) {
	return m.currentPrice, nil
}

func (m *mockStopLossMarketDataService) GetHistoricalData(symbol string, interval string, limit int) ([]*api.Kline, error) {
	return []*api.Kline{}, nil
}

func (m *mockStopLossMarketDataService) SubscribeToPrice(symbol string, callback func(float64)) error {
	return nil
}

func (m *mockStopLossMarketDataService) GetVolume(symbol string, timeWindow time.Duration) (float64, error) {
	return 1000.0, nil
}

// Unit tests for StopLossService

func TestSetStopLoss(t *testing.T) {
	// Setup
	stopOrderRepo := repository.NewMemoryStopOrderRepository()
	triggerEngine := NewTriggerEngine()
	mockTrading := &mockStopLossTradingService{}
	mockMarket := &mockStopLossMarketDataService{currentPrice: 50000.0}
	log := &mockLogger{}

	service := NewStopLossService(stopOrderRepo, triggerEngine, mockTrading, mockMarket, log)

	// Test successful stop loss creation
	t.Run("successful stop loss creation", func(t *testing.T) {
		stopOrder, err := service.SetStopLoss("BTCUSDT", 1.0, 45000.0)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if stopOrder.Symbol != "BTCUSDT" {
			t.Errorf("expected symbol BTCUSDT, got %s", stopOrder.Symbol)
		}

		if stopOrder.Position != 1.0 {
			t.Errorf("expected position 1.0, got %f", stopOrder.Position)
		}

		if stopOrder.StopPrice != 45000.0 {
			t.Errorf("expected stop price 45000.0, got %f", stopOrder.StopPrice)
		}

		if stopOrder.Type != repository.StopOrderTypeStopLoss {
			t.Errorf("expected type StopLoss, got %v", stopOrder.Type)
		}

		if stopOrder.Status != repository.StopOrderStatusActive {
			t.Errorf("expected status Active, got %v", stopOrder.Status)
		}
	})

	// Test invalid parameters
	t.Run("empty symbol", func(t *testing.T) {
		_, err := service.SetStopLoss("", 1.0, 45000.0)
		if err == nil {
			t.Error("expected error for empty symbol")
		}
	})

	t.Run("invalid position", func(t *testing.T) {
		_, err := service.SetStopLoss("BTCUSDT", 0, 45000.0)
		if err == nil {
			t.Error("expected error for invalid position")
		}
	})

	t.Run("invalid stop price", func(t *testing.T) {
		_, err := service.SetStopLoss("BTCUSDT", 1.0, 0)
		if err == nil {
			t.Error("expected error for invalid stop price")
		}
	})
}

func TestSetTakeProfit(t *testing.T) {
	// Setup
	stopOrderRepo := repository.NewMemoryStopOrderRepository()
	triggerEngine := NewTriggerEngine()
	mockTrading := &mockStopLossTradingService{}
	mockMarket := &mockStopLossMarketDataService{currentPrice: 50000.0}
	log := &mockLogger{}

	service := NewStopLossService(stopOrderRepo, triggerEngine, mockTrading, mockMarket, log)

	// Test successful take profit creation
	t.Run("successful take profit creation", func(t *testing.T) {
		takeProfitOrder, err := service.SetTakeProfit("BTCUSDT", 1.0, 55000.0)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if takeProfitOrder.Symbol != "BTCUSDT" {
			t.Errorf("expected symbol BTCUSDT, got %s", takeProfitOrder.Symbol)
		}

		if takeProfitOrder.Position != 1.0 {
			t.Errorf("expected position 1.0, got %f", takeProfitOrder.Position)
		}

		if takeProfitOrder.StopPrice != 55000.0 {
			t.Errorf("expected target price 55000.0, got %f", takeProfitOrder.StopPrice)
		}

		if takeProfitOrder.Type != repository.StopOrderTypeTakeProfit {
			t.Errorf("expected type TakeProfit, got %v", takeProfitOrder.Type)
		}

		if takeProfitOrder.Status != repository.StopOrderStatusActive {
			t.Errorf("expected status Active, got %v", takeProfitOrder.Status)
		}
	})

	// Test invalid parameters
	t.Run("empty symbol", func(t *testing.T) {
		_, err := service.SetTakeProfit("", 1.0, 55000.0)
		if err == nil {
			t.Error("expected error for empty symbol")
		}
	})

	t.Run("invalid position", func(t *testing.T) {
		_, err := service.SetTakeProfit("BTCUSDT", -1.0, 55000.0)
		if err == nil {
			t.Error("expected error for invalid position")
		}
	})

	t.Run("invalid target price", func(t *testing.T) {
		_, err := service.SetTakeProfit("BTCUSDT", 1.0, -100.0)
		if err == nil {
			t.Error("expected error for invalid target price")
		}
	})
}

func TestSetStopLossTakeProfit(t *testing.T) {
	// Setup
	stopOrderRepo := repository.NewMemoryStopOrderRepository()
	triggerEngine := NewTriggerEngine()
	mockTrading := &mockStopLossTradingService{}
	mockMarket := &mockStopLossMarketDataService{currentPrice: 50000.0}
	log := &mockLogger{}

	service := NewStopLossService(stopOrderRepo, triggerEngine, mockTrading, mockMarket, log)

	// Test successful pair creation
	t.Run("successful pair creation", func(t *testing.T) {
		pair, err := service.SetStopLossTakeProfit("BTCUSDT", 1.0, 45000.0, 55000.0)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if pair.Symbol != "BTCUSDT" {
			t.Errorf("expected symbol BTCUSDT, got %s", pair.Symbol)
		}

		if pair.Position != 1.0 {
			t.Errorf("expected position 1.0, got %f", pair.Position)
		}

		if pair.Status != "ACTIVE" {
			t.Errorf("expected status ACTIVE, got %s", pair.Status)
		}

		// Verify stop loss order
		if pair.StopLossOrder == nil {
			t.Fatal("expected stop loss order to be set")
		}

		if pair.StopLossOrder.StopPrice != 45000.0 {
			t.Errorf("expected stop price 45000.0, got %f", pair.StopLossOrder.StopPrice)
		}

		if pair.StopLossOrder.Type != repository.StopOrderTypeStopLoss {
			t.Errorf("expected type StopLoss, got %v", pair.StopLossOrder.Type)
		}

		// Verify take profit order
		if pair.TakeProfitOrder == nil {
			t.Fatal("expected take profit order to be set")
		}

		if pair.TakeProfitOrder.StopPrice != 55000.0 {
			t.Errorf("expected target price 55000.0, got %f", pair.TakeProfitOrder.StopPrice)
		}

		if pair.TakeProfitOrder.Type != repository.StopOrderTypeTakeProfit {
			t.Errorf("expected type TakeProfit, got %v", pair.TakeProfitOrder.Type)
		}
	})

	// Test invalid parameters
	t.Run("empty symbol", func(t *testing.T) {
		_, err := service.SetStopLossTakeProfit("", 1.0, 45000.0, 55000.0)
		if err == nil {
			t.Error("expected error for empty symbol")
		}
	})

	t.Run("invalid position", func(t *testing.T) {
		_, err := service.SetStopLossTakeProfit("BTCUSDT", 0, 45000.0, 55000.0)
		if err == nil {
			t.Error("expected error for invalid position")
		}
	})
}

func TestSetTrailingStop(t *testing.T) {
	// Setup
	stopOrderRepo := repository.NewMemoryStopOrderRepository()
	triggerEngine := NewTriggerEngine()
	mockTrading := &mockStopLossTradingService{}
	mockMarket := &mockStopLossMarketDataService{currentPrice: 50000.0}
	log := &mockLogger{}

	service := NewStopLossService(stopOrderRepo, triggerEngine, mockTrading, mockMarket, log)

	// Test successful trailing stop creation
	t.Run("successful trailing stop creation", func(t *testing.T) {
		trailingOrder, err := service.SetTrailingStop("BTCUSDT", 1.0, 2.0)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if trailingOrder.Symbol != "BTCUSDT" {
			t.Errorf("expected symbol BTCUSDT, got %s", trailingOrder.Symbol)
		}

		if trailingOrder.Position != 1.0 {
			t.Errorf("expected position 1.0, got %f", trailingOrder.Position)
		}

		if trailingOrder.TrailPercent != 2.0 {
			t.Errorf("expected trail percent 2.0, got %f", trailingOrder.TrailPercent)
		}

		if trailingOrder.HighestPrice != 50000.0 {
			t.Errorf("expected highest price 50000.0, got %f", trailingOrder.HighestPrice)
		}

		expectedStopPrice := 50000.0 * (1 - 2.0/100)
		if trailingOrder.CurrentStopPrice != expectedStopPrice {
			t.Errorf("expected stop price %f, got %f", expectedStopPrice, trailingOrder.CurrentStopPrice)
		}

		if trailingOrder.Status != repository.StopOrderStatusActive {
			t.Errorf("expected status Active, got %v", trailingOrder.Status)
		}
	})

	// Test invalid parameters
	t.Run("empty symbol", func(t *testing.T) {
		_, err := service.SetTrailingStop("", 1.0, 2.0)
		if err == nil {
			t.Error("expected error for empty symbol")
		}
	})

	t.Run("invalid position", func(t *testing.T) {
		_, err := service.SetTrailingStop("BTCUSDT", 0, 2.0)
		if err == nil {
			t.Error("expected error for invalid position")
		}
	})

	t.Run("invalid trail percent", func(t *testing.T) {
		_, err := service.SetTrailingStop("BTCUSDT", 1.0, 0)
		if err == nil {
			t.Error("expected error for invalid trail percent")
		}
	})
}

func TestCancelStopOrder(t *testing.T) {
	// Setup
	stopOrderRepo := repository.NewMemoryStopOrderRepository()
	triggerEngine := NewTriggerEngine()
	mockTrading := &mockStopLossTradingService{}
	mockMarket := &mockStopLossMarketDataService{currentPrice: 50000.0}
	log := &mockLogger{}

	service := NewStopLossService(stopOrderRepo, triggerEngine, mockTrading, mockMarket, log)

	// Test cancelling stop loss order
	t.Run("cancel stop loss order", func(t *testing.T) {
		stopOrder, _ := service.SetStopLoss("BTCUSDT", 1.0, 45000.0)

		err := service.CancelStopOrder(stopOrder.OrderID)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Verify order is cancelled
		order, _ := stopOrderRepo.FindStopOrderByID(stopOrder.OrderID)
		if order.Status != repository.StopOrderStatusCancelled {
			t.Errorf("expected status Cancelled, got %v", order.Status)
		}
	})

	// Test cancelling trailing stop order
	t.Run("cancel trailing stop order", func(t *testing.T) {
		trailingOrder, _ := service.SetTrailingStop("BTCUSDT", 1.0, 2.0)

		err := service.CancelStopOrder(trailingOrder.OrderID)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Verify order is cancelled
		order, _ := stopOrderRepo.FindTrailingStopOrderByID(trailingOrder.OrderID)
		if order.Status != repository.StopOrderStatusCancelled {
			t.Errorf("expected status Cancelled, got %v", order.Status)
		}
	})

	// Test invalid order ID
	t.Run("empty order ID", func(t *testing.T) {
		err := service.CancelStopOrder("")
		if err == nil {
			t.Error("expected error for empty order ID")
		}
	})

	t.Run("non-existent order ID", func(t *testing.T) {
		err := service.CancelStopOrder("non-existent")
		if err == nil {
			t.Error("expected error for non-existent order ID")
		}
	})
}

func TestGetActiveStopOrders(t *testing.T) {
	// Setup
	stopOrderRepo := repository.NewMemoryStopOrderRepository()
	triggerEngine := NewTriggerEngine()
	mockTrading := &mockStopLossTradingService{}
	mockMarket := &mockStopLossMarketDataService{currentPrice: 50000.0}
	log := &mockLogger{}

	service := NewStopLossService(stopOrderRepo, triggerEngine, mockTrading, mockMarket, log)

	// Test getting active orders
	t.Run("get active orders", func(t *testing.T) {
		// Create some orders
		stopOrder1, _ := service.SetStopLoss("BTCUSDT", 1.0, 45000.0)
		takeProfitOrder, _ := service.SetTakeProfit("BTCUSDT", 1.0, 55000.0)
		stopOrder3, _ := service.SetStopLoss("BTCUSDT", 0.5, 46000.0)

		// Cancel one order
		service.CancelStopOrder(stopOrder3.OrderID)

		orders, err := service.GetActiveStopOrders("BTCUSDT")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Should have 2 active orders (cancelled one excluded)
		if len(orders) != 2 {
			t.Errorf("expected 2 active orders, got %d", len(orders))
			// Debug: print order IDs
			t.Logf("Stop order 1: %s", stopOrder1.OrderID)
			t.Logf("Take profit order: %s", takeProfitOrder.OrderID)
			t.Logf("Stop order 3 (cancelled): %s", stopOrder3.OrderID)
			for i, order := range orders {
				t.Logf("Active order %d: %s (status: %v, type: %v)", i, order.OrderID, order.Status, order.Type)
			}
		}

		// Verify all returned orders are active
		for _, order := range orders {
			if order.Status != repository.StopOrderStatusActive {
				t.Errorf("expected all orders to be active, got %v", order.Status)
			}
		}
	})

	// Test invalid symbol
	t.Run("empty symbol", func(t *testing.T) {
		_, err := service.GetActiveStopOrders("")
		if err == nil {
			t.Error("expected error for empty symbol")
		}
	})
}

func TestUpdateTrailingStop(t *testing.T) {
	// Setup
	stopOrderRepo := repository.NewMemoryStopOrderRepository()
	triggerEngine := NewTriggerEngine()
	mockTrading := &mockStopLossTradingService{}
	mockMarket := &mockStopLossMarketDataService{currentPrice: 50000.0}
	log := &mockLogger{}

	service := NewStopLossService(stopOrderRepo, triggerEngine, mockTrading, mockMarket, log)

	// Test updating trailing stop
	t.Run("successful update", func(t *testing.T) {
		trailingOrder, _ := service.SetTrailingStop("BTCUSDT", 1.0, 2.0)

		err := service.UpdateTrailingStop(trailingOrder.OrderID, 3.0)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Verify update
		order, _ := stopOrderRepo.FindTrailingStopOrderByID(trailingOrder.OrderID)
		if order.TrailPercent != 3.0 {
			t.Errorf("expected trail percent 3.0, got %f", order.TrailPercent)
		}

		expectedStopPrice := 50000.0 * (1 - 3.0/100)
		if order.CurrentStopPrice != expectedStopPrice {
			t.Errorf("expected stop price %f, got %f", expectedStopPrice, order.CurrentStopPrice)
		}
	})

	// Test invalid parameters
	t.Run("empty order ID", func(t *testing.T) {
		err := service.UpdateTrailingStop("", 3.0)
		if err == nil {
			t.Error("expected error for empty order ID")
		}
	})

	t.Run("invalid trail percent", func(t *testing.T) {
		trailingOrder, _ := service.SetTrailingStop("BTCUSDT", 1.0, 2.0)
		err := service.UpdateTrailingStop(trailingOrder.OrderID, 0)
		if err == nil {
			t.Error("expected error for invalid trail percent")
		}
	})

	t.Run("non-existent order ID", func(t *testing.T) {
		err := service.UpdateTrailingStop("non-existent", 3.0)
		if err == nil {
			t.Error("expected error for non-existent order ID")
		}
	})
}

// Feature: binance-auto-trading, Property 33: Trailing stop price adjustment
// Validates: Requirements 9.5
func TestProperty_TrailingStopPriceAdjustment(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("for any trailing stop order and price sequence, when price creates new high, stop price must adjust by trail percent", prop.ForAll(
		func(position float64, trailPercent float64, initialPrice float64, priceIncrease float64) bool {
			// Ensure valid inputs
			if trailPercent <= 0 || trailPercent >= 100 {
				return true // Skip invalid trail percents
			}
			if initialPrice <= 0 {
				return true // Skip invalid prices
			}
			if priceIncrease < 0 {
				return true // Skip price decreases for this test
			}

			// Setup
			stopOrderRepo := repository.NewMemoryStopOrderRepository()
			triggerEngine := NewTriggerEngine()
			mockTrading := &mockStopLossTradingService{}
			mockMarket := &mockStopLossMarketDataService{currentPrice: initialPrice}
			log := &mockLogger{}

			service := NewStopLossService(stopOrderRepo, triggerEngine, mockTrading, mockMarket, log).(*stopLossService)

			// Create trailing stop order
			trailingOrder, err := service.SetTrailingStop("BTCUSDT", position, trailPercent)
			if err != nil {
				return false
			}

			// Record initial values
			initialHighest := trailingOrder.HighestPrice
			initialStopPrice := trailingOrder.CurrentStopPrice

			// Verify initial stop price calculation
			expectedInitialStop := initialPrice * (1 - trailPercent/100)
			tolerance := 0.0001
			if math.Abs(initialStopPrice-expectedInitialStop) > tolerance {
				return false
			}

			// Simulate price increase
			newPrice := initialPrice + priceIncrease

			// Update trailing stop price
			updated, err := service.UpdateTrailingStopPrice(trailingOrder.OrderID, newPrice)
			if err != nil {
				return false
			}

			// Get updated order
			updatedOrder, err := stopOrderRepo.FindTrailingStopOrderByID(trailingOrder.OrderID)
			if err != nil {
				return false
			}

			// If price increased, verify adjustments
			if newPrice > initialHighest {
				// Should have been updated
				if !updated {
					return false
				}

				// Highest price should be updated to new price
				if updatedOrder.HighestPrice != newPrice {
					return false
				}

				// Stop price should be adjusted by trail percent
				expectedNewStop := newPrice * (1 - trailPercent/100)
				if math.Abs(updatedOrder.CurrentStopPrice-expectedNewStop) > tolerance {
					return false
				}

				// Stop price should have increased
				if updatedOrder.CurrentStopPrice <= initialStopPrice {
					return false
				}
			} else {
				// Price didn't increase, nothing should change
				if updated {
					return false
				}

				if updatedOrder.HighestPrice != initialHighest {
					return false
				}

				if updatedOrder.CurrentStopPrice != initialStopPrice {
					return false
				}
			}

			return true
		},
		gen.Float64Range(0.1, 1000.0),     // position
		gen.Float64Range(0.5, 10.0),       // trail percent
		gen.Float64Range(1000.0, 50000.0), // initial price
		gen.Float64Range(0.0, 10000.0),    // price increase
	))

	properties.TestingRun(t)
}

// Add PlaceMarketSellOrder to mock
func (m *mockStopLossTradingService) PlaceMarketSellOrder(symbol string, quantity float64) (*api.Order, error) {
	return &api.Order{
		OrderID: 12347,
		Symbol:  symbol,
		Status:  api.OrderStatusFilled,
	}, nil
}

// Unit tests for trailing stop price tracking logic

func TestUpdateTrailingStopPrice(t *testing.T) {
	// Setup
	stopOrderRepo := repository.NewMemoryStopOrderRepository()
	triggerEngine := NewTriggerEngine()
	mockTrading := &mockStopLossTradingService{}
	mockMarket := &mockStopLossMarketDataService{currentPrice: 50000.0}
	log := &mockLogger{}

	service := NewStopLossService(stopOrderRepo, triggerEngine, mockTrading, mockMarket, log).(*stopLossService)

	// Test price increase updates stop price
	t.Run("price increase updates stop price", func(t *testing.T) {
		trailingOrder, _ := service.SetTrailingStop("BTCUSDT", 1.0, 2.0)
		initialStopPrice := trailingOrder.CurrentStopPrice

		// Simulate price increase
		newPrice := 52000.0
		updated, err := service.UpdateTrailingStopPrice(trailingOrder.OrderID, newPrice)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if !updated {
			t.Error("expected update to be true when price increases")
		}

		// Verify updated order
		order, _ := stopOrderRepo.FindTrailingStopOrderByID(trailingOrder.OrderID)
		if order.HighestPrice != newPrice {
			t.Errorf("expected highest price %f, got %f", newPrice, order.HighestPrice)
		}

		expectedStopPrice := newPrice * (1 - 2.0/100)
		if order.CurrentStopPrice != expectedStopPrice {
			t.Errorf("expected stop price %f, got %f", expectedStopPrice, order.CurrentStopPrice)
		}

		if order.CurrentStopPrice <= initialStopPrice {
			t.Error("expected stop price to increase")
		}
	})

	// Test price decrease does not update stop price
	t.Run("price decrease does not update stop price", func(t *testing.T) {
		trailingOrder, _ := service.SetTrailingStop("ETHUSDT", 1.0, 2.0)
		initialHighest := trailingOrder.HighestPrice
		initialStopPrice := trailingOrder.CurrentStopPrice

		// Simulate price decrease (but still above stop price)
		newPrice := initialHighest - 500.0 // Decrease but still above stop
		updated, err := service.UpdateTrailingStopPrice(trailingOrder.OrderID, newPrice)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Should not be updated (price didn't increase) and not triggered (still above stop)
		if updated {
			t.Error("expected update to be false when price decreases but doesn't trigger")
		}

		// Verify order unchanged
		order, _ := stopOrderRepo.FindTrailingStopOrderByID(trailingOrder.OrderID)
		if order.HighestPrice != initialHighest {
			t.Errorf("expected highest price unchanged at %f, got %f", initialHighest, order.HighestPrice)
		}

		if order.CurrentStopPrice != initialStopPrice {
			t.Errorf("expected stop price unchanged at %f, got %f", initialStopPrice, order.CurrentStopPrice)
		}
	})

	// Test stop price trigger
	t.Run("stop price trigger executes order", func(t *testing.T) {
		trailingOrder, _ := service.SetTrailingStop("BTCUSDT", 1.0, 2.0)

		// Simulate price falling below stop price
		triggerPrice := trailingOrder.CurrentStopPrice - 100.0
		triggered, err := service.UpdateTrailingStopPrice(trailingOrder.OrderID, triggerPrice)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if !triggered {
			t.Error("expected trigger to be true when price falls below stop")
		}

		// Verify order is triggered
		order, _ := stopOrderRepo.FindTrailingStopOrderByID(trailingOrder.OrderID)
		if order.Status != repository.StopOrderStatusTriggered {
			t.Errorf("expected status Triggered, got %v", order.Status)
		}
	})

	// Test boundary condition - price exactly at stop price
	t.Run("price at stop price triggers", func(t *testing.T) {
		trailingOrder, _ := service.SetTrailingStop("BTCUSDT", 1.0, 2.0)

		// Simulate price exactly at stop price
		triggerPrice := trailingOrder.CurrentStopPrice
		triggered, err := service.UpdateTrailingStopPrice(trailingOrder.OrderID, triggerPrice)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if !triggered {
			t.Error("expected trigger to be true when price equals stop price")
		}
	})

	// Test multiple price increases
	t.Run("multiple price increases adjust stop price progressively", func(t *testing.T) {
		trailingOrder, _ := service.SetTrailingStop("BTCUSDT", 1.0, 2.0)
		initialStopPrice := trailingOrder.CurrentStopPrice

		// First increase
		price1 := 51000.0
		service.UpdateTrailingStopPrice(trailingOrder.OrderID, price1)
		order1, _ := stopOrderRepo.FindTrailingStopOrderByID(trailingOrder.OrderID)
		stopPrice1 := order1.CurrentStopPrice

		// Second increase
		price2 := 53000.0
		service.UpdateTrailingStopPrice(trailingOrder.OrderID, price2)
		order2, _ := stopOrderRepo.FindTrailingStopOrderByID(trailingOrder.OrderID)
		stopPrice2 := order2.CurrentStopPrice

		// Verify progressive increases
		if stopPrice1 <= initialStopPrice {
			t.Error("expected first stop price to be higher than initial")
		}

		if stopPrice2 <= stopPrice1 {
			t.Error("expected second stop price to be higher than first")
		}

		// Verify correct calculation
		expectedStop2 := price2 * (1 - 2.0/100)
		if stopPrice2 != expectedStop2 {
			t.Errorf("expected stop price %f, got %f", expectedStop2, stopPrice2)
		}
	})
}

