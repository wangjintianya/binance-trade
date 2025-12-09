package service

import (
	"binance-trader/internal/api"
	"binance-trader/internal/repository"
	"fmt"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Feature: binance-auto-trading, Property 28: Trigger event log completeness
// Validates: Requirements 8.5
func TestProperty_TriggerEventLogCompleteness(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("trigger event logs must contain order_id, trigger_time, and trigger_price", prop.ForAll(
		func(orderID string, symbol string, price float64, quantity float64) bool {
			// Create a mock logger that captures log entries
			mockLogger := &mockLoggerCapture{entries: make([]map[string]interface{}, 0)}

			// Create monitoring engine with mock logger
			repo := repository.NewMemoryConditionalOrderRepository()
			triggerEngine := NewTriggerEngine()
			mockTrading := &mockTradingService{}
			mockMarket := &mockMarketDataService{prices: map[string]float64{symbol: price}}
			stopOrderRepo := repository.NewMemoryStopOrderRepository()
			mockStopLoss := &mockStopLossService{}

			engine := NewMonitoringEngine(repo, stopOrderRepo, triggerEngine, mockTrading, mockMarket, mockStopLoss, mockLogger, nil)

			// Create a conditional order
			order := &repository.ConditionalOrder{
				OrderID:  orderID,
				Symbol:   symbol,
				Side:     api.OrderSideBuy,
				Type:     api.OrderTypeMarket,
				Quantity: quantity,
				Price:    0,
				TriggerCondition: &repository.TriggerCondition{
					Type:     repository.TriggerTypePrice,
					Operator: repository.OperatorGreaterThan,
					Value:    price - 100, // Will trigger
				},
				Status:    repository.ConditionalOrderStatusPending,
				CreatedAt: time.Now().Unix(),
			}

			// Save order to repository
			repo.Save(order)

			// Create market data
			marketData := &MarketData{
				Symbol:    symbol,
				Price:     price,
				Timestamp: time.Now().Unix(),
			}

			// Build trigger log info
			logInfo := engine.buildTriggerLogInfo(order, marketData)

			// Verify log contains required fields
			_, hasOrderID := logInfo["order_id"]
			_, hasTriggerTime := logInfo["trigger_time"]
			_, hasTriggerPrice := logInfo["trigger_price"]

			return hasOrderID && hasTriggerTime && hasTriggerPrice
		},
		gen.Identifier().SuchThat(func(s string) bool { return s != "" }),
		gen.Identifier().SuchThat(func(s string) bool { return s != "" }),
		gen.Float64Range(100, 100000),
		gen.Float64Range(0.001, 1000),
	))

	properties.TestingRun(t)
}

// Feature: binance-auto-trading, Property 42: 复合条件触发日志
// Validates: Requirements 11.5
func TestProperty_CompositeConditionTriggerLog(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("composite condition trigger logs must contain all satisfied sub-conditions and their values", prop.ForAll(
		func(orderID string, symbol string, price float64, basePrice float64) bool {
			// Create a mock logger
			mockLogger := &mockLoggerCapture{entries: make([]map[string]interface{}, 0)}

			// Create monitoring engine
			repo := repository.NewMemoryConditionalOrderRepository()
			triggerEngine := NewTriggerEngine()
			mockTrading := &mockTradingService{}
			mockMarket := &mockMarketDataService{prices: map[string]float64{symbol: price}}
			stopOrderRepo := repository.NewMemoryStopOrderRepository()
			mockStopLoss := &mockStopLossService{}

			engine := NewMonitoringEngine(repo, stopOrderRepo, triggerEngine, mockTrading, mockMarket, mockStopLoss, mockLogger, nil)

			// Create a composite condition (AND)
			compositeCondition := &repository.TriggerCondition{
				CompositeType: repository.LogicAND,
				SubConditions: []*repository.TriggerCondition{
					{
						Type:     repository.TriggerTypePrice,
						Operator: repository.OperatorGreaterThan,
						Value:    basePrice,
					},
					{
						Type:      repository.TriggerTypePriceChangePercent,
						Operator:  repository.OperatorGreaterThan,
						Value:     5.0,
						BasePrice: basePrice,
					},
				},
			}

			// Create order with composite condition
			order := &repository.ConditionalOrder{
				OrderID:          orderID,
				Symbol:           symbol,
				Side:             api.OrderSideBuy,
				Type:             api.OrderTypeMarket,
				Quantity:         1.0,
				TriggerCondition: compositeCondition,
				Status:           repository.ConditionalOrderStatusPending,
				CreatedAt:        time.Now().Unix(),
			}

			// Create market data
			marketData := &MarketData{
				Symbol:    symbol,
				Price:     price,
				Timestamp: time.Now().Unix(),
			}

			// Build trigger log info
			logInfo := engine.buildTriggerLogInfo(order, marketData)

			// Verify log contains composite condition information
			conditionType, hasConditionType := logInfo["condition_type"]
			_, hasCompositeOperator := logInfo["composite_operator"]
			satisfiedConditions, hasSatisfiedConditions := logInfo["satisfied_conditions"]

			// Check that it's marked as composite
			isComposite := hasConditionType && conditionType == "composite"

			// Check that satisfied conditions are logged
			hasSatisfiedInfo := hasSatisfiedConditions && satisfiedConditions != nil

			return isComposite && hasCompositeOperator && hasSatisfiedInfo
		},
		gen.Identifier().SuchThat(func(s string) bool { return s != "" }),
		gen.Identifier().SuchThat(func(s string) bool { return s != "" }),
		gen.Float64Range(1000, 10000),
		gen.Float64Range(100, 900),
	))

	properties.TestingRun(t)
}

// Mock logger that captures log entries
type mockLoggerCapture struct {
	entries []map[string]interface{}
}

func (m *mockLoggerCapture) Debug(msg string, fields map[string]interface{}) {
	entry := make(map[string]interface{})
	entry["level"] = "debug"
	entry["message"] = msg
	for k, v := range fields {
		entry[k] = v
	}
	m.entries = append(m.entries, entry)
}

func (m *mockLoggerCapture) Info(msg string, fields map[string]interface{}) {
	entry := make(map[string]interface{})
	entry["level"] = "info"
	entry["message"] = msg
	for k, v := range fields {
		entry[k] = v
	}
	m.entries = append(m.entries, entry)
}

func (m *mockLoggerCapture) Warn(msg string, fields map[string]interface{}) {
	entry := make(map[string]interface{})
	entry["level"] = "warn"
	entry["message"] = msg
	for k, v := range fields {
		entry[k] = v
	}
	m.entries = append(m.entries, entry)
}

func (m *mockLoggerCapture) Error(msg string, fields map[string]interface{}) {
	entry := make(map[string]interface{})
	entry["level"] = "error"
	entry["message"] = msg
	for k, v := range fields {
		entry[k] = v
	}
	m.entries = append(m.entries, entry)
}

func (m *mockLoggerCapture) Fatal(msg string, fields map[string]interface{}) {
	entry := make(map[string]interface{})
	entry["level"] = "fatal"
	entry["message"] = msg
	for k, v := range fields {
		entry[k] = v
	}
	m.entries = append(m.entries, entry)
}

func (m *mockLoggerCapture) LogAPIOperation(operationType string, result string, fields map[string]interface{}) {
	entry := make(map[string]interface{})
	entry["level"] = "info"
	entry["operation_type"] = operationType
	entry["result"] = result
	for k, v := range fields {
		entry[k] = v
	}
	m.entries = append(m.entries, entry)
}

func (m *mockLoggerCapture) LogOrderEvent(eventType string, orderID int64, symbol, side, orderType string, quantity float64, fields map[string]interface{}) {
	entry := make(map[string]interface{})
	entry["level"] = "info"
	entry["event_type"] = eventType
	entry["order_id"] = orderID
	entry["symbol"] = symbol
	entry["side"] = side
	entry["order_type"] = orderType
	entry["quantity"] = quantity
	for k, v := range fields {
		entry[k] = v
	}
	m.entries = append(m.entries, entry)
}

func (m *mockLoggerCapture) LogError(err error, fields map[string]interface{}) {
	entry := make(map[string]interface{})
	entry["level"] = "error"
	entry["error"] = err.Error()
	for k, v := range fields {
		entry[k] = v
	}
	m.entries = append(m.entries, entry)
}

func (m *mockLoggerCapture) SetTradingType(tradingType string) {
	// No-op for mock
}

func (m *mockLoggerCapture) LogFuturesAPIOperation(operationType string, result string, fields map[string]interface{}) {
	entry := make(map[string]interface{})
	entry["level"] = "info"
	entry["operation_type"] = operationType
	entry["result"] = result
	for k, v := range fields {
		entry[k] = v
	}
	m.entries = append(m.entries, entry)
}

func (m *mockLoggerCapture) LogFuturesOrderEvent(eventType string, orderID int64, symbol, side, orderType string, quantity float64, positionChange map[string]interface{}, fields map[string]interface{}) {
	entry := make(map[string]interface{})
	entry["level"] = "info"
	entry["event_type"] = eventType
	entry["order_id"] = orderID
	entry["symbol"] = symbol
	entry["side"] = side
	entry["order_type"] = orderType
	entry["quantity"] = quantity
	entry["position_change"] = positionChange
	for k, v := range fields {
		entry[k] = v
	}
	m.entries = append(m.entries, entry)
}

func (m *mockLoggerCapture) LogLiquidationEvent(symbol string, positionSide string, liquidationPrice float64, lossAmount float64, reason string, fields map[string]interface{}) {
	entry := make(map[string]interface{})
	entry["level"] = "error"
	entry["symbol"] = symbol
	entry["position_side"] = positionSide
	entry["liquidation_price"] = liquidationPrice
	entry["loss_amount"] = lossAmount
	entry["reason"] = reason
	for k, v := range fields {
		entry[k] = v
	}
	m.entries = append(m.entries, entry)
}

func (m *mockLoggerCapture) LogFundingRateSettlement(symbol string, fundingFee float64, fundingRate float64, positionSize float64, fields map[string]interface{}) {
	entry := make(map[string]interface{})
	entry["level"] = "info"
	entry["symbol"] = symbol
	entry["funding_fee"] = fundingFee
	entry["funding_rate"] = fundingRate
	entry["position_size"] = positionSize
	for k, v := range fields {
		entry[k] = v
	}
	m.entries = append(m.entries, entry)
}

// Mock trading service for testing
type mockTradingService struct{}

func (m *mockTradingService) PlaceMarketBuyOrder(symbol string, quantity float64) (*api.Order, error) {
	return &api.Order{
		OrderID:     12345,
		Symbol:      symbol,
		Side:        api.OrderSideBuy,
		Type:        api.OrderTypeMarket,
		Status:      api.OrderStatusFilled,
		Price:       0,
		OrigQty:     quantity,
		ExecutedQty: quantity,
		Time:        time.Now().Unix(),
		UpdateTime:  time.Now().Unix(),
	}, nil
}

func (m *mockTradingService) PlaceMarketSellOrder(symbol string, quantity float64) (*api.Order, error) {
	return &api.Order{
		OrderID:     12346,
		Symbol:      symbol,
		Side:        api.OrderSideSell,
		Type:        api.OrderTypeMarket,
		Status:      api.OrderStatusFilled,
		Price:       0,
		OrigQty:     quantity,
		ExecutedQty: quantity,
		Time:        time.Now().Unix(),
		UpdateTime:  time.Now().Unix(),
	}, nil
}

func (m *mockTradingService) PlaceLimitSellOrder(symbol string, price, quantity float64) (*api.Order, error) {
	return &api.Order{
		OrderID:     12346,
		Symbol:      symbol,
		Side:        api.OrderSideSell,
		Type:        api.OrderTypeLimit,
		Status:      api.OrderStatusNew,
		Price:       price,
		OrigQty:     quantity,
		ExecutedQty: 0,
		Time:        time.Now().Unix(),
		UpdateTime:  time.Now().Unix(),
	}, nil
}

func (m *mockTradingService) CancelOrder(orderID int64) error {
	return nil
}

func (m *mockTradingService) GetOrderStatus(orderID int64) (*OrderStatus, error) {
	return &OrderStatus{
		OrderID:     orderID,
		Status:      api.OrderStatusFilled,
		ExecutedQty: 1.0,
	}, nil
}

func (m *mockTradingService) GetActiveOrders() ([]*api.Order, error) {
	return []*api.Order{}, nil
}

// Mock market data service for testing
type mockMarketDataService struct {
	prices map[string]float64
}

func (m *mockMarketDataService) GetCurrentPrice(symbol string) (float64, error) {
	if price, ok := m.prices[symbol]; ok {
		return price, nil
	}
	return 1000.0, nil
}

func (m *mockMarketDataService) GetHistoricalData(symbol string, interval string, limit int) ([]*api.Kline, error) {
	return []*api.Kline{}, nil
}

func (m *mockMarketDataService) SubscribeToPrice(symbol string, callback func(float64)) error {
	return nil
}

func (m *mockMarketDataService) GetVolume(symbol string, timeWindow time.Duration) (float64, error) {
	return 1000.0, nil
}

// Unit tests for monitoring engine

func TestMonitoringEngine_StartStop(t *testing.T) {
	// Create monitoring engine
	repo := repository.NewMemoryConditionalOrderRepository()
	triggerEngine := NewTriggerEngine()
	mockTrading := &mockTradingService{}
	mockMarket := &mockMarketDataService{prices: map[string]float64{}}
	mockLogger := &mockLoggerCapture{entries: make([]map[string]interface{}, 0)}
	stopOrderRepo := repository.NewMemoryStopOrderRepository()
	mockStopLoss := &mockStopLossService{}

	engine := NewMonitoringEngine(repo, stopOrderRepo, triggerEngine, mockTrading, mockMarket, mockStopLoss, mockLogger, nil)

	// Test initial state
	if engine.IsRunning() {
		t.Error("Engine should not be running initially")
	}

	// Test start
	err := engine.Start()
	if err != nil {
		t.Errorf("Failed to start engine: %v", err)
	}

	if !engine.IsRunning() {
		t.Error("Engine should be running after start")
	}

	// Test double start
	err = engine.Start()
	if err == nil {
		t.Error("Starting already running engine should return error")
	}

	// Test stop
	err = engine.Stop()
	if err != nil {
		t.Errorf("Failed to stop engine: %v", err)
	}

	if engine.IsRunning() {
		t.Error("Engine should not be running after stop")
	}

	// Test double stop
	err = engine.Stop()
	if err == nil {
		t.Error("Stopping already stopped engine should return error")
	}
}

func TestMonitoringEngine_TriggerDetection(t *testing.T) {
	// Create monitoring engine
	repo := repository.NewMemoryConditionalOrderRepository()
	triggerEngine := NewTriggerEngine()
	mockTrading := &mockTradingService{}
	mockMarket := &mockMarketDataService{prices: map[string]float64{"BTCUSDT": 50000.0}}
	mockLogger := &mockLoggerCapture{entries: make([]map[string]interface{}, 0)}
	stopOrderRepo := repository.NewMemoryStopOrderRepository()
	mockStopLoss := &mockStopLossService{}

	engine := NewMonitoringEngine(repo, stopOrderRepo, triggerEngine, mockTrading, mockMarket, mockStopLoss, mockLogger, nil)

	// Create a conditional order that should trigger
	order := &repository.ConditionalOrder{
		OrderID:  "test-order-1",
		Symbol:   "BTCUSDT",
		Side:     api.OrderSideBuy,
		Type:     api.OrderTypeMarket,
		Quantity: 1.0,
		TriggerCondition: &repository.TriggerCondition{
			Type:     repository.TriggerTypePrice,
			Operator: repository.OperatorGreaterThan,
			Value:    40000.0, // Will trigger since price is 50000
		},
		Status:    repository.ConditionalOrderStatusPending,
		CreatedAt: time.Now().Unix(),
	}

	// Save order to repository
	err := repo.Save(order)
	if err != nil {
		t.Fatalf("Failed to save order: %v", err)
	}

	// Process the order
	engine.processOrder(order)

	// Check that order was executed
	updatedOrder, err := repo.FindByID("test-order-1")
	if err != nil {
		t.Fatalf("Failed to find order: %v", err)
	}

	if updatedOrder.Status != repository.ConditionalOrderStatusExecuted {
		t.Errorf("Expected order status to be EXECUTED, got %s", updatedOrder.Status)
	}

	if updatedOrder.ExecutedOrderID == 0 {
		t.Error("Expected executed order ID to be set")
	}
}

func TestMonitoringEngine_TimeWindowFiltering(t *testing.T) {
	// Create monitoring engine
	repo := repository.NewMemoryConditionalOrderRepository()
	triggerEngine := NewTriggerEngine()
	mockTrading := &mockTradingService{}
	mockMarket := &mockMarketDataService{prices: map[string]float64{"BTCUSDT": 50000.0}}
	mockLogger := &mockLoggerCapture{entries: make([]map[string]interface{}, 0)}
	stopOrderRepo := repository.NewMemoryStopOrderRepository()
	mockStopLoss := &mockStopLossService{}

	engine := NewMonitoringEngine(repo, stopOrderRepo, triggerEngine, mockTrading, mockMarket, mockStopLoss, mockLogger, nil)

	// Create a conditional order with time window in the past (should not trigger)
	pastWindow := &repository.TimeWindow{
		StartTime: time.Now().Add(-2 * time.Hour),
		EndTime:   time.Now().Add(-1 * time.Hour),
	}

	order := &repository.ConditionalOrder{
		OrderID:  "test-order-2",
		Symbol:   "BTCUSDT",
		Side:     api.OrderSideBuy,
		Type:     api.OrderTypeMarket,
		Quantity: 1.0,
		TriggerCondition: &repository.TriggerCondition{
			Type:     repository.TriggerTypePrice,
			Operator: repository.OperatorGreaterThan,
			Value:    40000.0,
		},
		Status:     repository.ConditionalOrderStatusPending,
		CreatedAt:  time.Now().Unix(),
		TimeWindow: pastWindow,
	}

	// Save order to repository
	err := repo.Save(order)
	if err != nil {
		t.Fatalf("Failed to save order: %v", err)
	}

	// Process the order
	engine.processOrder(order)

	// Check that order was NOT executed (time window expired)
	updatedOrder, err := repo.FindByID("test-order-2")
	if err != nil {
		t.Fatalf("Failed to find order: %v", err)
	}

	if updatedOrder.Status != repository.ConditionalOrderStatusPending {
		t.Errorf("Expected order status to remain PENDING, got %s", updatedOrder.Status)
	}
}

func TestMonitoringEngine_ConcurrentSafety(t *testing.T) {
	// Create monitoring engine
	repo := repository.NewMemoryConditionalOrderRepository()
	triggerEngine := NewTriggerEngine()
	mockTrading := &mockTradingService{}
	mockMarket := &mockMarketDataService{prices: map[string]float64{"BTCUSDT": 50000.0}}
	mockLogger := &mockLoggerCapture{entries: make([]map[string]interface{}, 0)}
	stopOrderRepo := repository.NewMemoryStopOrderRepository()
	mockStopLoss := &mockStopLossService{}

	engine := NewMonitoringEngine(repo, stopOrderRepo, triggerEngine, mockTrading, mockMarket, mockStopLoss, mockLogger, nil)

	// Create multiple orders
	for i := 0; i < 10; i++ {
		order := &repository.ConditionalOrder{
			OrderID:  fmt.Sprintf("test-order-%d", i),
			Symbol:   "BTCUSDT",
			Side:     api.OrderSideBuy,
			Type:     api.OrderTypeMarket,
			Quantity: 1.0,
			TriggerCondition: &repository.TriggerCondition{
				Type:     repository.TriggerTypePrice,
				Operator: repository.OperatorGreaterThan,
				Value:    40000.0,
			},
			Status:    repository.ConditionalOrderStatusPending,
			CreatedAt: time.Now().Unix(),
		}
		repo.Save(order)
	}

	// Start engine
	err := engine.Start()
	if err != nil {
		t.Fatalf("Failed to start engine: %v", err)
	}

	// Wait a bit for processing
	time.Sleep(2 * time.Second)

	// Stop engine
	err = engine.Stop()
	if err != nil {
		t.Fatalf("Failed to stop engine: %v", err)
	}

	// Check that all orders were processed
	activeOrders, err := repo.FindActiveOrders()
	if err != nil {
		t.Fatalf("Failed to get active orders: %v", err)
	}

	if len(activeOrders) > 0 {
		t.Errorf("Expected all orders to be processed, but %d are still active", len(activeOrders))
	}
}

func TestMonitoringEngine_MarketDataCaching(t *testing.T) {
	// Create monitoring engine
	repo := repository.NewMemoryConditionalOrderRepository()
	triggerEngine := NewTriggerEngine()
	mockTrading := &mockTradingService{}
	mockMarket := &mockMarketDataService{prices: map[string]float64{"BTCUSDT": 50000.0}}
	mockLogger := &mockLoggerCapture{entries: make([]map[string]interface{}, 0)}
	stopOrderRepo := repository.NewMemoryStopOrderRepository()
	mockStopLoss := &mockStopLossService{}

	engine := NewMonitoringEngine(repo, stopOrderRepo, triggerEngine, mockTrading, mockMarket, mockStopLoss, mockLogger, nil)

	// Get market data twice
	data1, err := engine.getMarketData("BTCUSDT")
	if err != nil {
		t.Fatalf("Failed to get market data: %v", err)
	}

	data2, err := engine.getMarketData("BTCUSDT")
	if err != nil {
		t.Fatalf("Failed to get market data: %v", err)
	}

	// Check that data is cached (same timestamp)
	if data1.Timestamp != data2.Timestamp {
		t.Error("Expected market data to be cached")
	}

	// Wait for cache to expire
	time.Sleep(1100 * time.Millisecond)

	data3, err := engine.getMarketData("BTCUSDT")
	if err != nil {
		t.Fatalf("Failed to get market data: %v", err)
	}

	// Check that cache was refreshed
	if data1.Timestamp == data3.Timestamp {
		t.Error("Expected market data cache to be refreshed")
	}
}

func TestMonitoringEngine_LoadActiveOrders(t *testing.T) {
	// Create monitoring engine
	repo := repository.NewMemoryConditionalOrderRepository()
	triggerEngine := NewTriggerEngine()
	mockTrading := &mockTradingService{}
	mockMarket := &mockMarketDataService{prices: map[string]float64{}}
	mockLogger := &mockLoggerCapture{entries: make([]map[string]interface{}, 0)}

	// Create some orders with different statuses
	pendingOrder := &repository.ConditionalOrder{
		OrderID:  "pending-1",
		Symbol:   "BTCUSDT",
		Side:     api.OrderSideBuy,
		Type:     api.OrderTypeMarket,
		Quantity: 1.0,
		TriggerCondition: &repository.TriggerCondition{
			Type:     repository.TriggerTypePrice,
			Operator: repository.OperatorGreaterThan,
			Value:    40000.0,
		},
		Status:    repository.ConditionalOrderStatusPending,
		CreatedAt: time.Now().Unix(),
	}

	executedOrder := &repository.ConditionalOrder{
		OrderID:  "executed-1",
		Symbol:   "BTCUSDT",
		Side:     api.OrderSideBuy,
		Type:     api.OrderTypeMarket,
		Quantity: 1.0,
		TriggerCondition: &repository.TriggerCondition{
			Type:     repository.TriggerTypePrice,
			Operator: repository.OperatorGreaterThan,
			Value:    40000.0,
		},
		Status:    repository.ConditionalOrderStatusExecuted,
		CreatedAt: time.Now().Unix(),
	}

	repo.Save(pendingOrder)
	repo.Save(executedOrder)
	stopOrderRepo := repository.NewMemoryStopOrderRepository()
	mockStopLoss := &mockStopLossService{}

	engine := NewMonitoringEngine(repo, stopOrderRepo, triggerEngine, mockTrading, mockMarket, mockStopLoss, mockLogger, nil)

	// Start engine (which loads active orders)
	err := engine.Start()
	if err != nil {
		t.Fatalf("Failed to start engine: %v", err)
	}

	// Check that only pending order is loaded
	engine.mu.RLock()
	activeCount := len(engine.activeOrders)
	_, hasPending := engine.activeOrders["pending-1"]
	_, hasExecuted := engine.activeOrders["executed-1"]
	engine.mu.RUnlock()

	if activeCount != 1 {
		t.Errorf("Expected 1 active order, got %d", activeCount)
	}

	if !hasPending {
		t.Error("Expected pending order to be loaded")
	}

	if hasExecuted {
		t.Error("Expected executed order not to be loaded")
	}

	engine.Stop()
}

func TestMonitoringEngine_BuildTriggerLogInfo(t *testing.T) {
	// Create monitoring engine
	repo := repository.NewMemoryConditionalOrderRepository()
	triggerEngine := NewTriggerEngine()
	mockTrading := &mockTradingService{}
	mockMarket := &mockMarketDataService{prices: map[string]float64{}}
	mockLogger := &mockLoggerCapture{entries: make([]map[string]interface{}, 0)}
	stopOrderRepo := repository.NewMemoryStopOrderRepository()
	mockStopLoss := &mockStopLossService{}

	engine := NewMonitoringEngine(repo, stopOrderRepo, triggerEngine, mockTrading, mockMarket, mockStopLoss, mockLogger, nil)

	// Test simple price condition
	order := &repository.ConditionalOrder{
		OrderID:  "test-1",
		Symbol:   "BTCUSDT",
		Side:     api.OrderSideBuy,
		Type:     api.OrderTypeMarket,
		Quantity: 1.0,
		TriggerCondition: &repository.TriggerCondition{
			Type:     repository.TriggerTypePrice,
			Operator: repository.OperatorGreaterThan,
			Value:    50000.0,
		},
		Status:    repository.ConditionalOrderStatusPending,
		CreatedAt: time.Now().Unix(),
	}

	marketData := &MarketData{
		Symbol:    "BTCUSDT",
		Price:     51000.0,
		Timestamp: time.Now().Unix(),
	}

	logInfo := engine.buildTriggerLogInfo(order, marketData)

	// Verify required fields
	if logInfo["order_id"] != "test-1" {
		t.Error("Expected order_id in log info")
	}

	if logInfo["symbol"] != "BTCUSDT" {
		t.Error("Expected symbol in log info")
	}

	if logInfo["trigger_price"] != 51000.0 {
		t.Error("Expected trigger_price in log info")
	}

	if logInfo["trigger_type"] != "price" {
		t.Error("Expected trigger_type in log info")
	}

	if logInfo["operator"] != ">" {
		t.Error("Expected operator in log info")
	}
}

func TestMonitoringEngine_BuildTriggerLogInfo_Composite(t *testing.T) {
	// Create monitoring engine
	repo := repository.NewMemoryConditionalOrderRepository()
	triggerEngine := NewTriggerEngine()
	mockTrading := &mockTradingService{}
	mockMarket := &mockMarketDataService{prices: map[string]float64{}}
	mockLogger := &mockLoggerCapture{entries: make([]map[string]interface{}, 0)}
	stopOrderRepo := repository.NewMemoryStopOrderRepository()
	mockStopLoss := &mockStopLossService{}

	engine := NewMonitoringEngine(repo, stopOrderRepo, triggerEngine, mockTrading, mockMarket, mockStopLoss, mockLogger, nil)

	// Test composite condition
	order := &repository.ConditionalOrder{
		OrderID:  "test-2",
		Symbol:   "BTCUSDT",
		Side:     api.OrderSideBuy,
		Type:     api.OrderTypeMarket,
		Quantity: 1.0,
		TriggerCondition: &repository.TriggerCondition{
			CompositeType: repository.LogicAND,
			SubConditions: []*repository.TriggerCondition{
				{
					Type:     repository.TriggerTypePrice,
					Operator: repository.OperatorGreaterThan,
					Value:    50000.0,
				},
				{
					Type:      repository.TriggerTypePriceChangePercent,
					Operator:  repository.OperatorGreaterThan,
					Value:     5.0,
					BasePrice: 48000.0,
				},
			},
		},
		Status:    repository.ConditionalOrderStatusPending,
		CreatedAt: time.Now().Unix(),
	}

	marketData := &MarketData{
		Symbol:    "BTCUSDT",
		Price:     51000.0,
		Timestamp: time.Now().Unix(),
	}

	logInfo := engine.buildTriggerLogInfo(order, marketData)

	// Verify composite condition fields
	if logInfo["condition_type"] != "composite" {
		t.Error("Expected condition_type to be composite")
	}

	if logInfo["composite_operator"] != "AND" {
		t.Error("Expected composite_operator to be AND")
	}

	if logInfo["satisfied_conditions"] == nil {
		t.Error("Expected satisfied_conditions in log info")
	}
}

// Mock stop loss service for testing
type mockStopLossService struct{}

func (m *mockStopLossService) SetStopLoss(symbol string, position float64, stopPrice float64) (*repository.StopOrder, error) {
	return &repository.StopOrder{}, nil
}

func (m *mockStopLossService) SetTakeProfit(symbol string, position float64, targetPrice float64) (*repository.StopOrder, error) {
	return &repository.StopOrder{}, nil
}

func (m *mockStopLossService) SetStopLossTakeProfit(symbol string, position float64, stopPrice, targetPrice float64) (*repository.StopOrderPair, error) {
	return &repository.StopOrderPair{}, nil
}

func (m *mockStopLossService) SetTrailingStop(symbol string, position float64, trailPercent float64) (*repository.TrailingStopOrder, error) {
	return &repository.TrailingStopOrder{}, nil
}

func (m *mockStopLossService) CancelStopOrder(orderID string) error {
	return nil
}

func (m *mockStopLossService) GetActiveStopOrders(symbol string) ([]*repository.StopOrder, error) {
	return []*repository.StopOrder{}, nil
}

func (m *mockStopLossService) UpdateTrailingStop(orderID string, newTrailPercent float64) error {
	return nil
}
