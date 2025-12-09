package service

import (
	"binance-trader/internal/api"
	"binance-trader/internal/repository"
	"binance-trader/pkg/logger"
	"fmt"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Feature: binance-auto-trading, Property 34: 活跃条件订单过滤
// Validates: Requirements 10.1
func TestProperty_ActiveConditionalOrderFiltering(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("GetActiveConditionalOrders should only return PENDING orders", prop.ForAll(
		func(orders []*repository.ConditionalOrder) bool {
			// Setup
			repo := repository.NewMemoryConditionalOrderRepository()
			triggerEngine := NewTriggerEngine()
			log, _ := logger.NewLogger(logger.Config{Level: "info", EnableConsole: false})
			stopOrderRepo := repository.NewMemoryStopOrderRepository()
			service := NewConditionalOrderService(repo, stopOrderRepo, triggerEngine, nil, nil, nil, log)

			// Save all orders to repository
			for _, order := range orders {
				if err := repo.Save(order); err != nil {
					t.Logf("Failed to save order: %v", err)
					return false
				}
			}

			// Get active orders
			activeOrders, err := service.GetActiveConditionalOrders()
			if err != nil {
				t.Logf("Failed to get active orders: %v", err)
				return false
			}

			// Verify all returned orders have PENDING status
			for _, order := range activeOrders {
				if order.Status != repository.ConditionalOrderStatusPending {
					t.Logf("Non-pending order returned: %s", order.Status)
					return false
				}
			}

			// Count expected pending orders
			expectedCount := 0
			for _, order := range orders {
				if order.Status == repository.ConditionalOrderStatusPending {
					expectedCount++
				}
			}

			// Verify count matches
			if len(activeOrders) != expectedCount {
				t.Logf("Expected %d active orders, got %d", expectedCount, len(activeOrders))
				return false
			}

			return true
		},
		genConditionalOrderSlice(),
	))

	properties.TestingRun(t)
}

// Feature: binance-auto-trading, Property 35: 条件订单取消效果
// Validates: Requirements 10.2
func TestProperty_ConditionalOrderCancellationEffect(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("Cancelled orders should not appear in active orders list", prop.ForAll(
		func(order *repository.ConditionalOrder) bool {
			// Setup
			repo := repository.NewMemoryConditionalOrderRepository()
			triggerEngine := NewTriggerEngine()
			log, _ := logger.NewLogger(logger.Config{Level: "info", EnableConsole: false})
			stopOrderRepo := repository.NewMemoryStopOrderRepository()
			service := NewConditionalOrderService(repo, stopOrderRepo, triggerEngine, nil, nil, nil, log)

			// Ensure order is pending
			order.Status = repository.ConditionalOrderStatusPending

			// Save order
			if err := repo.Save(order); err != nil {
				t.Logf("Failed to save order: %v", err)
				return false
			}

			// Verify order is in active list before cancellation
			activeOrdersBefore, err := service.GetActiveConditionalOrders()
			if err != nil {
				t.Logf("Failed to get active orders before: %v", err)
				return false
			}

			foundBefore := false
			for _, o := range activeOrdersBefore {
				if o.OrderID == order.OrderID {
					foundBefore = true
					break
				}
			}

			if !foundBefore {
				t.Logf("Order not found in active list before cancellation")
				return false
			}

			// Cancel order
			if err := service.CancelConditionalOrder(order.OrderID); err != nil {
				t.Logf("Failed to cancel order: %v", err)
				return false
			}

			// Verify order is not in active list after cancellation
			activeOrdersAfter, err := service.GetActiveConditionalOrders()
			if err != nil {
				t.Logf("Failed to get active orders after: %v", err)
				return false
			}

			for _, o := range activeOrdersAfter {
				if o.OrderID == order.OrderID {
					t.Logf("Cancelled order still in active list")
					return false
				}
			}

			return true
		},
		genConditionalOrder(),
	))

	properties.TestingRun(t)
}

// Feature: binance-auto-trading, Property 36: 条件订单更新一致�?
// Validates: Requirements 10.3
func TestProperty_ConditionalOrderUpdateConsistency(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("Updated order parameters should be reflected when queried", prop.ForAll(
		func(order *repository.ConditionalOrder, newQuantity float64, newPrice float64) bool {
			// Ensure valid values
			if newQuantity <= 0 {
				newQuantity = 1.0
			}
			if newPrice <= 0 {
				newPrice = 100.0
			}

			// Setup
			repo := repository.NewMemoryConditionalOrderRepository()
			triggerEngine := NewTriggerEngine()
			log, _ := logger.NewLogger(logger.Config{Level: "info", EnableConsole: false})
			stopOrderRepo := repository.NewMemoryStopOrderRepository()
			service := NewConditionalOrderService(repo, stopOrderRepo, triggerEngine, nil, nil, nil, log)

			// Ensure order is pending
			order.Status = repository.ConditionalOrderStatusPending
			order.Type = api.OrderTypeLimit // Ensure it's a limit order for price updates

			// Save order
			if err := repo.Save(order); err != nil {
				t.Logf("Failed to save order: %v", err)
				return false
			}

			// Update order
			updates := &ConditionalOrderUpdate{
				Quantity: &newQuantity,
				Price:    &newPrice,
			}

			if err := service.UpdateConditionalOrder(order.OrderID, updates); err != nil {
				t.Logf("Failed to update order: %v", err)
				return false
			}

			// Query updated order
			updatedOrder, err := service.GetConditionalOrder(order.OrderID)
			if err != nil {
				t.Logf("Failed to get updated order: %v", err)
				return false
			}

			// Verify updates
			if updatedOrder.Quantity != newQuantity {
				t.Logf("Quantity not updated: expected %f, got %f", newQuantity, updatedOrder.Quantity)
				return false
			}

			if updatedOrder.Price != newPrice {
				t.Logf("Price not updated: expected %f, got %f", newPrice, updatedOrder.Price)
				return false
			}

			return true
		},
		genConditionalOrder(),
		gen.Float64Range(0.1, 1000.0),
		gen.Float64Range(1.0, 100000.0),
	))

	properties.TestingRun(t)
}

// Feature: binance-auto-trading, Property 38: 历史订单查询过滤
// Validates: Requirements 10.5
func TestProperty_HistoricalOrderQueryFiltering(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("History query should only return EXECUTED or CANCELLED orders", prop.ForAll(
		func(orders []*repository.ConditionalOrder) bool {
			// Setup
			repo := repository.NewMemoryConditionalOrderRepository()
			triggerEngine := NewTriggerEngine()
			log, _ := logger.NewLogger(logger.Config{Level: "info", EnableConsole: false})
			stopOrderRepo := repository.NewMemoryStopOrderRepository()
			service := NewConditionalOrderService(repo, stopOrderRepo, triggerEngine, nil, nil, nil, log)

			// Save all orders
			for _, order := range orders {
				if err := repo.Save(order); err != nil {
					t.Logf("Failed to save order: %v", err)
					return false
				}
			}

			// Query history (use wide time range to get all orders)
			startTime := int64(0)
			endTime := time.Now().Unix() + 1000000

			historyOrders, err := service.GetConditionalOrderHistory(startTime, endTime)
			if err != nil {
				t.Logf("Failed to get history: %v", err)
				return false
			}

			// Verify all returned orders are EXECUTED or CANCELLED
			for _, order := range historyOrders {
				if order.Status != repository.ConditionalOrderStatusExecuted &&
					order.Status != repository.ConditionalOrderStatusCancelled {
					t.Logf("Invalid status in history: %s", order.Status)
					return false
				}
			}

			// Count expected history orders (must be in time range AND have correct status)
			expectedCount := 0
			for _, order := range orders {
				if (order.Status == repository.ConditionalOrderStatusExecuted ||
					order.Status == repository.ConditionalOrderStatusCancelled) &&
					order.CreatedAt >= startTime && order.CreatedAt <= endTime {
					expectedCount++
				}
			}

			// Verify count matches
			if len(historyOrders) != expectedCount {
				t.Logf("Expected %d history orders, got %d", expectedCount, len(historyOrders))
				return false
			}

			return true
		},
		genConditionalOrderSlice(),
	))

	properties.TestingRun(t)
}

// Generators

var orderIDCounter int64 = 0

func genConditionalOrder() gopter.Gen {
	return gopter.CombineGens(
		gen.OneConstOf("BTCUSDT", "ETHUSDT", "BNBUSDT"),    // Symbol
		gen.OneConstOf(api.OrderSideBuy, api.OrderSideSell), // Side
		gen.OneConstOf(api.OrderTypeMarket, api.OrderTypeLimit), // Type
		gen.Float64Range(0.001, 100.0),                     // Quantity
		gen.Float64Range(1.0, 100000.0),                    // Price
		genTriggerCondition(),                               // TriggerCondition
		genConditionalOrderStatus(),                         // Status
		gen.Int64Range(1000000000, 2000000000),             // CreatedAt
	).Map(func(values []interface{}) *repository.ConditionalOrder {
		// Generate unique order ID using counter
		orderIDCounter++
		orderID := fmt.Sprintf("order-%d", orderIDCounter)
		
		return &repository.ConditionalOrder{
			OrderID:          orderID,
			Symbol:           values[0].(string),
			Side:             values[1].(api.OrderSide),
			Type:             values[2].(api.OrderType),
			Quantity:         values[3].(float64),
			Price:            values[4].(float64),
			TriggerCondition: values[5].(*repository.TriggerCondition),
			Status:           values[6].(repository.ConditionalOrderStatus),
			CreatedAt:        values[7].(int64),
		}
	})
}

func genConditionalOrderSlice() gopter.Gen {
	return gen.SliceOf(genConditionalOrder())
}

func genTriggerCondition() gopter.Gen {
	return gopter.CombineGens(
		gen.OneConstOf(
			repository.TriggerTypePrice,
			repository.TriggerTypePriceChangePercent,
			repository.TriggerTypeVolume,
		),
		gen.OneConstOf(
			repository.OperatorGreaterThan,
			repository.OperatorLessThan,
			repository.OperatorGreaterEqual,
			repository.OperatorLessEqual,
		),
		gen.Float64Range(1.0, 100000.0), // Value
		gen.Float64Range(1.0, 100000.0), // BasePrice
	).Map(func(values []interface{}) *repository.TriggerCondition {
		return &repository.TriggerCondition{
			Type:      values[0].(repository.TriggerType),
			Operator:  values[1].(repository.ComparisonOperator),
			Value:     values[2].(float64),
			BasePrice: values[3].(float64),
		}
	})
}

func genConditionalOrderStatus() gopter.Gen {
	return gen.OneConstOf(
		repository.ConditionalOrderStatusPending,
		repository.ConditionalOrderStatusTriggered,
		repository.ConditionalOrderStatusExecuted,
		repository.ConditionalOrderStatusCancelled,
	)
}


// Unit Tests

func TestConditionalOrderService_CreateConditionalOrder(t *testing.T) {
	// Setup
	repo := repository.NewMemoryConditionalOrderRepository()
	triggerEngine := NewTriggerEngine()
	log, _ := logger.NewLogger(logger.Config{Level: "info", EnableConsole: false})
	stopOrderRepo := repository.NewMemoryStopOrderRepository()
			service := NewConditionalOrderService(repo, stopOrderRepo, triggerEngine, nil, nil, nil, log)

	// Test valid order creation
	request := &repository.ConditionalOrderRequest{
		Symbol:   "BTCUSDT",
		Side:     api.OrderSideBuy,
		Type:     api.OrderTypeMarket,
		Quantity: 1.0,
		TriggerCondition: &repository.TriggerCondition{
			Type:     repository.TriggerTypePrice,
			Operator: repository.OperatorGreaterThan,
			Value:    50000.0,
		},
	}

	order, err := service.CreateConditionalOrder(request)
	if err != nil {
		t.Fatalf("Failed to create conditional order: %v", err)
	}

	if order.OrderID == "" {
		t.Error("Order ID should not be empty")
	}

	if order.Status != repository.ConditionalOrderStatusPending {
		t.Errorf("Expected status PENDING, got %s", order.Status)
	}

	// Test invalid request (nil)
	_, err = service.CreateConditionalOrder(nil)
	if err == nil {
		t.Error("Expected error for nil request")
	}

	// Test invalid request (empty symbol)
	invalidRequest := &repository.ConditionalOrderRequest{
		Symbol:   "",
		Quantity: 1.0,
		TriggerCondition: &repository.TriggerCondition{
			Type:     repository.TriggerTypePrice,
			Operator: repository.OperatorGreaterThan,
			Value:    50000.0,
		},
	}
	_, err = service.CreateConditionalOrder(invalidRequest)
	if err == nil {
		t.Error("Expected error for empty symbol")
	}
}

func TestConditionalOrderService_CancelConditionalOrder(t *testing.T) {
	// Setup
	repo := repository.NewMemoryConditionalOrderRepository()
	triggerEngine := NewTriggerEngine()
	log, _ := logger.NewLogger(logger.Config{Level: "info", EnableConsole: false})
	stopOrderRepo := repository.NewMemoryStopOrderRepository()
			service := NewConditionalOrderService(repo, stopOrderRepo, triggerEngine, nil, nil, nil, log)

	// Create an order first
	request := &repository.ConditionalOrderRequest{
		Symbol:   "BTCUSDT",
		Side:     api.OrderSideBuy,
		Type:     api.OrderTypeMarket,
		Quantity: 1.0,
		TriggerCondition: &repository.TriggerCondition{
			Type:     repository.TriggerTypePrice,
			Operator: repository.OperatorGreaterThan,
			Value:    50000.0,
		},
	}

	order, err := service.CreateConditionalOrder(request)
	if err != nil {
		t.Fatalf("Failed to create order: %v", err)
	}

	// Cancel the order
	err = service.CancelConditionalOrder(order.OrderID)
	if err != nil {
		t.Fatalf("Failed to cancel order: %v", err)
	}

	// Verify order is cancelled
	cancelledOrder, err := service.GetConditionalOrder(order.OrderID)
	if err != nil {
		t.Fatalf("Failed to get cancelled order: %v", err)
	}

	if cancelledOrder.Status != repository.ConditionalOrderStatusCancelled {
		t.Errorf("Expected status CANCELLED, got %s", cancelledOrder.Status)
	}

	// Test cancelling non-existent order
	err = service.CancelConditionalOrder("non-existent-id")
	if err == nil {
		t.Error("Expected error for non-existent order")
	}
}

func TestConditionalOrderService_UpdateConditionalOrder(t *testing.T) {
	// Setup
	repo := repository.NewMemoryConditionalOrderRepository()
	triggerEngine := NewTriggerEngine()
	log, _ := logger.NewLogger(logger.Config{Level: "info", EnableConsole: false})
	stopOrderRepo := repository.NewMemoryStopOrderRepository()
			service := NewConditionalOrderService(repo, stopOrderRepo, triggerEngine, nil, nil, nil, log)

	// Create an order first
	request := &repository.ConditionalOrderRequest{
		Symbol:   "BTCUSDT",
		Side:     api.OrderSideBuy,
		Type:     api.OrderTypeLimit,
		Quantity: 1.0,
		Price:    50000.0,
		TriggerCondition: &repository.TriggerCondition{
			Type:     repository.TriggerTypePrice,
			Operator: repository.OperatorGreaterThan,
			Value:    50000.0,
		},
	}

	order, err := service.CreateConditionalOrder(request)
	if err != nil {
		t.Fatalf("Failed to create order: %v", err)
	}

	// Update the order
	newQuantity := 2.0
	newPrice := 55000.0
	updates := &ConditionalOrderUpdate{
		Quantity: &newQuantity,
		Price:    &newPrice,
	}

	err = service.UpdateConditionalOrder(order.OrderID, updates)
	if err != nil {
		t.Fatalf("Failed to update order: %v", err)
	}

	// Verify updates
	updatedOrder, err := service.GetConditionalOrder(order.OrderID)
	if err != nil {
		t.Fatalf("Failed to get updated order: %v", err)
	}

	if updatedOrder.Quantity != newQuantity {
		t.Errorf("Expected quantity %f, got %f", newQuantity, updatedOrder.Quantity)
	}

	if updatedOrder.Price != newPrice {
		t.Errorf("Expected price %f, got %f", newPrice, updatedOrder.Price)
	}
}

func TestConditionalOrderService_GetActiveConditionalOrders(t *testing.T) {
	// Setup
	repo := repository.NewMemoryConditionalOrderRepository()
	triggerEngine := NewTriggerEngine()
	log, _ := logger.NewLogger(logger.Config{Level: "info", EnableConsole: false})
	stopOrderRepo := repository.NewMemoryStopOrderRepository()
			service := NewConditionalOrderService(repo, stopOrderRepo, triggerEngine, nil, nil, nil, log)

	// Create multiple orders with different statuses
	for i := 0; i < 3; i++ {
		request := &repository.ConditionalOrderRequest{
			Symbol:   "BTCUSDT",
			Side:     api.OrderSideBuy,
			Type:     api.OrderTypeMarket,
			Quantity: 1.0,
			TriggerCondition: &repository.TriggerCondition{
				Type:     repository.TriggerTypePrice,
				Operator: repository.OperatorGreaterThan,
				Value:    50000.0,
			},
		}
		_, err := service.CreateConditionalOrder(request)
		if err != nil {
			t.Fatalf("Failed to create order: %v", err)
		}
	}

	// Get active orders
	activeOrders, err := service.GetActiveConditionalOrders()
	if err != nil {
		t.Fatalf("Failed to get active orders: %v", err)
	}

	if len(activeOrders) != 3 {
		t.Errorf("Expected 3 active orders, got %d", len(activeOrders))
	}

	// Verify all are pending
	for _, order := range activeOrders {
		if order.Status != repository.ConditionalOrderStatusPending {
			t.Errorf("Expected PENDING status, got %s", order.Status)
		}
	}
}

func TestConditionalOrderService_GetConditionalOrderHistory(t *testing.T) {
	// Setup
	repo := repository.NewMemoryConditionalOrderRepository()
	triggerEngine := NewTriggerEngine()
	log, _ := logger.NewLogger(logger.Config{Level: "info", EnableConsole: false})
	stopOrderRepo := repository.NewMemoryStopOrderRepository()
			service := NewConditionalOrderService(repo, stopOrderRepo, triggerEngine, nil, nil, nil, log)

	// Create and cancel an order
	request := &repository.ConditionalOrderRequest{
		Symbol:   "BTCUSDT",
		Side:     api.OrderSideBuy,
		Type:     api.OrderTypeMarket,
		Quantity: 1.0,
		TriggerCondition: &repository.TriggerCondition{
			Type:     repository.TriggerTypePrice,
			Operator: repository.OperatorGreaterThan,
			Value:    50000.0,
		},
	}

	order, err := service.CreateConditionalOrder(request)
	if err != nil {
		t.Fatalf("Failed to create order: %v", err)
	}

	err = service.CancelConditionalOrder(order.OrderID)
	if err != nil {
		t.Fatalf("Failed to cancel order: %v", err)
	}

	// Get history
	startTime := time.Now().Unix() - 1000
	endTime := time.Now().Unix() + 1000

	history, err := service.GetConditionalOrderHistory(startTime, endTime)
	if err != nil {
		t.Fatalf("Failed to get history: %v", err)
	}

	if len(history) != 1 {
		t.Errorf("Expected 1 history order, got %d", len(history))
	}

	if history[0].Status != repository.ConditionalOrderStatusCancelled {
		t.Errorf("Expected CANCELLED status, got %s", history[0].Status)
	}
}



