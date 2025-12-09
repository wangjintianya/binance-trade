package repository

import (
	"binance-trader/pkg/errors"
	"testing"
	"time"
)

// TestStopOrderRepository_SaveAndFind tests saving and finding stop orders
func TestStopOrderRepository_SaveAndFind(t *testing.T) {
	repo := NewMemoryStopOrderRepository()

	order := &StopOrder{
		OrderID:   "stop-001",
		Symbol:    "BTCUSDT",
		Position:  1.5,
		StopPrice: 45000.0,
		Type:      StopOrderTypeStopLoss,
		Status:    StopOrderStatusActive,
		CreatedAt: time.Now().Unix(),
	}

	// Test Save
	err := repo.SaveStopOrder(order)
	if err != nil {
		t.Fatalf("SaveStopOrder failed: %v", err)
	}

	// Test FindByID
	found, err := repo.FindStopOrderByID("stop-001")
	if err != nil {
		t.Fatalf("FindStopOrderByID failed: %v", err)
	}

	if found.OrderID != order.OrderID {
		t.Errorf("Expected OrderID %s, got %s", order.OrderID, found.OrderID)
	}
	if found.Symbol != order.Symbol {
		t.Errorf("Expected Symbol %s, got %s", order.Symbol, found.Symbol)
	}
	if found.Position != order.Position {
		t.Errorf("Expected Position %f, got %f", order.Position, found.Position)
	}
	if found.StopPrice != order.StopPrice {
		t.Errorf("Expected StopPrice %f, got %f", order.StopPrice, found.StopPrice)
	}
}

// TestStopOrderRepository_SaveNilOrder tests saving nil order
func TestStopOrderRepository_SaveNilOrder(t *testing.T) {
	repo := NewMemoryStopOrderRepository()

	err := repo.SaveStopOrder(nil)
	if err == nil {
		t.Fatal("Expected error when saving nil order")
	}

	tradingErr, ok := err.(*errors.TradingError)
	if !ok {
		t.Fatal("Expected TradingError")
	}
	if tradingErr.Type != errors.ErrInvalidParameter {
		t.Errorf("Expected ErrInvalidParameter, got %v", tradingErr.Type)
	}
}

// TestStopOrderRepository_FindBySymbol tests finding orders by symbol
func TestStopOrderRepository_FindBySymbol(t *testing.T) {
	repo := NewMemoryStopOrderRepository()

	orders := []*StopOrder{
		{
			OrderID:   "stop-001",
			Symbol:    "BTCUSDT",
			Position:  1.0,
			StopPrice: 45000.0,
			Type:      StopOrderTypeStopLoss,
			Status:    StopOrderStatusActive,
			CreatedAt: time.Now().Unix(),
		},
		{
			OrderID:   "stop-002",
			Symbol:    "BTCUSDT",
			Position:  0.5,
			StopPrice: 50000.0,
			Type:      StopOrderTypeTakeProfit,
			Status:    StopOrderStatusActive,
			CreatedAt: time.Now().Unix(),
		},
		{
			OrderID:   "stop-003",
			Symbol:    "ETHUSDT",
			Position:  10.0,
			StopPrice: 3000.0,
			Type:      StopOrderTypeStopLoss,
			Status:    StopOrderStatusActive,
			CreatedAt: time.Now().Unix(),
		},
	}

	for _, order := range orders {
		if err := repo.SaveStopOrder(order); err != nil {
			t.Fatalf("SaveStopOrder failed: %v", err)
		}
	}

	// Find BTCUSDT orders
	btcOrders, err := repo.FindStopOrdersBySymbol("BTCUSDT")
	if err != nil {
		t.Fatalf("FindStopOrdersBySymbol failed: %v", err)
	}

	if len(btcOrders) != 2 {
		t.Errorf("Expected 2 BTCUSDT orders, got %d", len(btcOrders))
	}

	// Find ETHUSDT orders
	ethOrders, err := repo.FindStopOrdersBySymbol("ETHUSDT")
	if err != nil {
		t.Fatalf("FindStopOrdersBySymbol failed: %v", err)
	}

	if len(ethOrders) != 1 {
		t.Errorf("Expected 1 ETHUSDT order, got %d", len(ethOrders))
	}
}

// TestStopOrderRepository_Update tests updating stop orders
func TestStopOrderRepository_Update(t *testing.T) {
	repo := NewMemoryStopOrderRepository()

	order := &StopOrder{
		OrderID:   "stop-001",
		Symbol:    "BTCUSDT",
		Position:  1.0,
		StopPrice: 45000.0,
		Type:      StopOrderTypeStopLoss,
		Status:    StopOrderStatusActive,
		CreatedAt: time.Now().Unix(),
	}

	if err := repo.SaveStopOrder(order); err != nil {
		t.Fatalf("SaveStopOrder failed: %v", err)
	}

	// Update order
	order.StopPrice = 46000.0
	order.Status = StopOrderStatusTriggered

	if err := repo.UpdateStopOrder(order); err != nil {
		t.Fatalf("UpdateStopOrder failed: %v", err)
	}

	// Verify update
	found, err := repo.FindStopOrderByID("stop-001")
	if err != nil {
		t.Fatalf("FindStopOrderByID failed: %v", err)
	}

	if found.StopPrice != 46000.0 {
		t.Errorf("Expected StopPrice 46000.0, got %f", found.StopPrice)
	}
	if found.Status != StopOrderStatusTriggered {
		t.Errorf("Expected Status TRIGGERED, got %s", found.Status)
	}
}

// TestStopOrderRepository_Delete tests deleting stop orders
func TestStopOrderRepository_Delete(t *testing.T) {
	repo := NewMemoryStopOrderRepository()

	order := &StopOrder{
		OrderID:   "stop-001",
		Symbol:    "BTCUSDT",
		Position:  1.0,
		StopPrice: 45000.0,
		Type:      StopOrderTypeStopLoss,
		Status:    StopOrderStatusActive,
		CreatedAt: time.Now().Unix(),
	}

	if err := repo.SaveStopOrder(order); err != nil {
		t.Fatalf("SaveStopOrder failed: %v", err)
	}

	// Delete order
	if err := repo.DeleteStopOrder("stop-001"); err != nil {
		t.Fatalf("DeleteStopOrder failed: %v", err)
	}

	// Verify deletion
	_, err := repo.FindStopOrderByID("stop-001")
	if err == nil {
		t.Fatal("Expected error when finding deleted order")
	}

	tradingErr, ok := err.(*errors.TradingError)
	if !ok {
		t.Fatal("Expected TradingError")
	}
	if tradingErr.Type != errors.ErrStopOrderNotFound {
		t.Errorf("Expected ErrStopOrderNotFound, got %v", tradingErr.Type)
	}
}

// TestStopOrderRepository_FindActiveOrders tests finding active orders
func TestStopOrderRepository_FindActiveOrders(t *testing.T) {
	repo := NewMemoryStopOrderRepository()

	orders := []*StopOrder{
		{
			OrderID:   "stop-001",
			Symbol:    "BTCUSDT",
			Position:  1.0,
			StopPrice: 45000.0,
			Type:      StopOrderTypeStopLoss,
			Status:    StopOrderStatusActive,
			CreatedAt: time.Now().Unix(),
		},
		{
			OrderID:   "stop-002",
			Symbol:    "BTCUSDT",
			Position:  0.5,
			StopPrice: 50000.0,
			Type:      StopOrderTypeTakeProfit,
			Status:    StopOrderStatusTriggered,
			CreatedAt: time.Now().Unix(),
		},
		{
			OrderID:   "stop-003",
			Symbol:    "BTCUSDT",
			Position:  2.0,
			StopPrice: 44000.0,
			Type:      StopOrderTypeStopLoss,
			Status:    StopOrderStatusActive,
			CreatedAt: time.Now().Unix(),
		},
	}

	for _, order := range orders {
		if err := repo.SaveStopOrder(order); err != nil {
			t.Fatalf("SaveStopOrder failed: %v", err)
		}
	}

	// Find active orders
	activeOrders, err := repo.FindActiveStopOrders("BTCUSDT")
	if err != nil {
		t.Fatalf("FindActiveStopOrders failed: %v", err)
	}

	if len(activeOrders) != 2 {
		t.Errorf("Expected 2 active orders, got %d", len(activeOrders))
	}

	// Verify all returned orders are active
	for _, order := range activeOrders {
		if order.Status != StopOrderStatusActive {
			t.Errorf("Expected active order, got status %s", order.Status)
		}
	}
}

// TestStopOrderRepository_UpdateStatus tests updating order status
func TestStopOrderRepository_UpdateStatus(t *testing.T) {
	repo := NewMemoryStopOrderRepository()

	order := &StopOrder{
		OrderID:   "stop-001",
		Symbol:    "BTCUSDT",
		Position:  1.0,
		StopPrice: 45000.0,
		Type:      StopOrderTypeStopLoss,
		Status:    StopOrderStatusActive,
		CreatedAt: time.Now().Unix(),
	}

	if err := repo.SaveStopOrder(order); err != nil {
		t.Fatalf("SaveStopOrder failed: %v", err)
	}

	// Update status
	triggeredAt := time.Now().Unix()
	executedOrderID := int64(12345)

	err := repo.UpdateStopOrderStatus("stop-001", StopOrderStatusTriggered, triggeredAt, executedOrderID)
	if err != nil {
		t.Fatalf("UpdateStopOrderStatus failed: %v", err)
	}

	// Verify update
	found, err := repo.FindStopOrderByID("stop-001")
	if err != nil {
		t.Fatalf("FindStopOrderByID failed: %v", err)
	}

	if found.Status != StopOrderStatusTriggered {
		t.Errorf("Expected Status TRIGGERED, got %s", found.Status)
	}
	if found.TriggeredAt != triggeredAt {
		t.Errorf("Expected TriggeredAt %d, got %d", triggeredAt, found.TriggeredAt)
	}
	if found.ExecutedOrderID != executedOrderID {
		t.Errorf("Expected ExecutedOrderID %d, got %d", executedOrderID, found.ExecutedOrderID)
	}
}

// TestStopOrderPairRepository_SaveAndFind tests saving and finding stop order pairs
func TestStopOrderPairRepository_SaveAndFind(t *testing.T) {
	repo := NewMemoryStopOrderRepository()

	stopLoss := &StopOrder{
		OrderID:   "stop-001",
		Symbol:    "BTCUSDT",
		Position:  1.0,
		StopPrice: 45000.0,
		Type:      StopOrderTypeStopLoss,
		Status:    StopOrderStatusActive,
		CreatedAt: time.Now().Unix(),
	}

	takeProfit := &StopOrder{
		OrderID:   "stop-002",
		Symbol:    "BTCUSDT",
		Position:  1.0,
		StopPrice: 50000.0,
		Type:      StopOrderTypeTakeProfit,
		Status:    StopOrderStatusActive,
		CreatedAt: time.Now().Unix(),
	}

	pair := &StopOrderPair{
		PairID:          "pair-001",
		Symbol:          "BTCUSDT",
		Position:        1.0,
		StopLossOrder:   stopLoss,
		TakeProfitOrder: takeProfit,
		Status:          "ACTIVE",
	}

	// Test Save
	err := repo.SaveStopOrderPair(pair)
	if err != nil {
		t.Fatalf("SaveStopOrderPair failed: %v", err)
	}

	// Test FindByID
	found, err := repo.FindStopOrderPairByID("pair-001")
	if err != nil {
		t.Fatalf("FindStopOrderPairByID failed: %v", err)
	}

	if found.PairID != pair.PairID {
		t.Errorf("Expected PairID %s, got %s", pair.PairID, found.PairID)
	}
	if found.Symbol != pair.Symbol {
		t.Errorf("Expected Symbol %s, got %s", pair.Symbol, found.Symbol)
	}
	if found.StopLossOrder == nil {
		t.Fatal("Expected StopLossOrder to be non-nil")
	}
	if found.TakeProfitOrder == nil {
		t.Fatal("Expected TakeProfitOrder to be non-nil")
	}
	if found.StopLossOrder.OrderID != stopLoss.OrderID {
		t.Errorf("Expected StopLoss OrderID %s, got %s", stopLoss.OrderID, found.StopLossOrder.OrderID)
	}
	if found.TakeProfitOrder.OrderID != takeProfit.OrderID {
		t.Errorf("Expected TakeProfit OrderID %s, got %s", takeProfit.OrderID, found.TakeProfitOrder.OrderID)
	}
}

// TestStopOrderPairRepository_FindBySymbol tests finding pairs by symbol
func TestStopOrderPairRepository_FindBySymbol(t *testing.T) {
	repo := NewMemoryStopOrderRepository()

	pairs := []*StopOrderPair{
		{
			PairID:   "pair-001",
			Symbol:   "BTCUSDT",
			Position: 1.0,
			Status:   "ACTIVE",
		},
		{
			PairID:   "pair-002",
			Symbol:   "BTCUSDT",
			Position: 0.5,
			Status:   "ACTIVE",
		},
		{
			PairID:   "pair-003",
			Symbol:   "ETHUSDT",
			Position: 10.0,
			Status:   "ACTIVE",
		},
	}

	for _, pair := range pairs {
		if err := repo.SaveStopOrderPair(pair); err != nil {
			t.Fatalf("SaveStopOrderPair failed: %v", err)
		}
	}

	// Find BTCUSDT pairs
	btcPairs, err := repo.FindStopOrderPairsBySymbol("BTCUSDT")
	if err != nil {
		t.Fatalf("FindStopOrderPairsBySymbol failed: %v", err)
	}

	if len(btcPairs) != 2 {
		t.Errorf("Expected 2 BTCUSDT pairs, got %d", len(btcPairs))
	}
}

// TestStopOrderPairRepository_Update tests updating stop order pairs
func TestStopOrderPairRepository_Update(t *testing.T) {
	repo := NewMemoryStopOrderRepository()

	pair := &StopOrderPair{
		PairID:   "pair-001",
		Symbol:   "BTCUSDT",
		Position: 1.0,
		Status:   "ACTIVE",
	}

	if err := repo.SaveStopOrderPair(pair); err != nil {
		t.Fatalf("SaveStopOrderPair failed: %v", err)
	}

	// Update pair
	pair.Status = "PARTIALLY_TRIGGERED"

	if err := repo.UpdateStopOrderPair(pair); err != nil {
		t.Fatalf("UpdateStopOrderPair failed: %v", err)
	}

	// Verify update
	found, err := repo.FindStopOrderPairByID("pair-001")
	if err != nil {
		t.Fatalf("FindStopOrderPairByID failed: %v", err)
	}

	if found.Status != "PARTIALLY_TRIGGERED" {
		t.Errorf("Expected Status PARTIALLY_TRIGGERED, got %s", found.Status)
	}
}

// TestStopOrderPairRepository_Delete tests deleting stop order pairs
func TestStopOrderPairRepository_Delete(t *testing.T) {
	repo := NewMemoryStopOrderRepository()

	pair := &StopOrderPair{
		PairID:   "pair-001",
		Symbol:   "BTCUSDT",
		Position: 1.0,
		Status:   "ACTIVE",
	}

	if err := repo.SaveStopOrderPair(pair); err != nil {
		t.Fatalf("SaveStopOrderPair failed: %v", err)
	}

	// Delete pair
	if err := repo.DeleteStopOrderPair("pair-001"); err != nil {
		t.Fatalf("DeleteStopOrderPair failed: %v", err)
	}

	// Verify deletion
	_, err := repo.FindStopOrderPairByID("pair-001")
	if err == nil {
		t.Fatal("Expected error when finding deleted pair")
	}
}

// TestStopOrderPairRepository_FindActivePairs tests finding active pairs
func TestStopOrderPairRepository_FindActivePairs(t *testing.T) {
	repo := NewMemoryStopOrderRepository()

	pairs := []*StopOrderPair{
		{
			PairID:   "pair-001",
			Symbol:   "BTCUSDT",
			Position: 1.0,
			Status:   "ACTIVE",
		},
		{
			PairID:   "pair-002",
			Symbol:   "BTCUSDT",
			Position: 0.5,
			Status:   "COMPLETED",
		},
		{
			PairID:   "pair-003",
			Symbol:   "BTCUSDT",
			Position: 2.0,
			Status:   "ACTIVE",
		},
	}

	for _, pair := range pairs {
		if err := repo.SaveStopOrderPair(pair); err != nil {
			t.Fatalf("SaveStopOrderPair failed: %v", err)
		}
	}

	// Find active pairs
	activePairs, err := repo.FindActiveStopOrderPairs("BTCUSDT")
	if err != nil {
		t.Fatalf("FindActiveStopOrderPairs failed: %v", err)
	}

	if len(activePairs) != 2 {
		t.Errorf("Expected 2 active pairs, got %d", len(activePairs))
	}

	// Verify all returned pairs are active
	for _, pair := range activePairs {
		if pair.Status != "ACTIVE" {
			t.Errorf("Expected active pair, got status %s", pair.Status)
		}
	}
}

// TestTrailingStopOrderRepository_SaveAndFind tests saving and finding trailing stop orders
func TestTrailingStopOrderRepository_SaveAndFind(t *testing.T) {
	repo := NewMemoryStopOrderRepository()

	order := &TrailingStopOrder{
		OrderID:          "trail-001",
		Symbol:           "BTCUSDT",
		Position:         1.0,
		TrailPercent:     2.0,
		HighestPrice:     48000.0,
		CurrentStopPrice: 47040.0,
		Status:           StopOrderStatusActive,
		CreatedAt:        time.Now().Unix(),
		LastUpdatedAt:    time.Now().Unix(),
	}

	// Test Save
	err := repo.SaveTrailingStopOrder(order)
	if err != nil {
		t.Fatalf("SaveTrailingStopOrder failed: %v", err)
	}

	// Test FindByID
	found, err := repo.FindTrailingStopOrderByID("trail-001")
	if err != nil {
		t.Fatalf("FindTrailingStopOrderByID failed: %v", err)
	}

	if found.OrderID != order.OrderID {
		t.Errorf("Expected OrderID %s, got %s", order.OrderID, found.OrderID)
	}
	if found.TrailPercent != order.TrailPercent {
		t.Errorf("Expected TrailPercent %f, got %f", order.TrailPercent, found.TrailPercent)
	}
	if found.HighestPrice != order.HighestPrice {
		t.Errorf("Expected HighestPrice %f, got %f", order.HighestPrice, found.HighestPrice)
	}
	if found.CurrentStopPrice != order.CurrentStopPrice {
		t.Errorf("Expected CurrentStopPrice %f, got %f", order.CurrentStopPrice, found.CurrentStopPrice)
	}
}

// TestTrailingStopOrderRepository_Update tests updating trailing stop orders
func TestTrailingStopOrderRepository_Update(t *testing.T) {
	repo := NewMemoryStopOrderRepository()

	order := &TrailingStopOrder{
		OrderID:          "trail-001",
		Symbol:           "BTCUSDT",
		Position:         1.0,
		TrailPercent:     2.0,
		HighestPrice:     48000.0,
		CurrentStopPrice: 47040.0,
		Status:           StopOrderStatusActive,
		CreatedAt:        time.Now().Unix(),
		LastUpdatedAt:    time.Now().Unix(),
	}

	if err := repo.SaveTrailingStopOrder(order); err != nil {
		t.Fatalf("SaveTrailingStopOrder failed: %v", err)
	}

	// Update order - price moved up
	order.HighestPrice = 49000.0
	order.CurrentStopPrice = 48020.0
	order.LastUpdatedAt = time.Now().Unix()

	if err := repo.UpdateTrailingStopOrder(order); err != nil {
		t.Fatalf("UpdateTrailingStopOrder failed: %v", err)
	}

	// Verify update
	found, err := repo.FindTrailingStopOrderByID("trail-001")
	if err != nil {
		t.Fatalf("FindTrailingStopOrderByID failed: %v", err)
	}

	if found.HighestPrice != 49000.0 {
		t.Errorf("Expected HighestPrice 49000.0, got %f", found.HighestPrice)
	}
	if found.CurrentStopPrice != 48020.0 {
		t.Errorf("Expected CurrentStopPrice 48020.0, got %f", found.CurrentStopPrice)
	}
}

// TestTrailingStopOrderRepository_FindActiveOrders tests finding active trailing stop orders
func TestTrailingStopOrderRepository_FindActiveOrders(t *testing.T) {
	repo := NewMemoryStopOrderRepository()

	orders := []*TrailingStopOrder{
		{
			OrderID:          "trail-001",
			Symbol:           "BTCUSDT",
			Position:         1.0,
			TrailPercent:     2.0,
			HighestPrice:     48000.0,
			CurrentStopPrice: 47040.0,
			Status:           StopOrderStatusActive,
			CreatedAt:        time.Now().Unix(),
		},
		{
			OrderID:          "trail-002",
			Symbol:           "BTCUSDT",
			Position:         0.5,
			TrailPercent:     3.0,
			HighestPrice:     49000.0,
			CurrentStopPrice: 47530.0,
			Status:           StopOrderStatusTriggered,
			CreatedAt:        time.Now().Unix(),
		},
		{
			OrderID:          "trail-003",
			Symbol:           "BTCUSDT",
			Position:         2.0,
			TrailPercent:     2.5,
			HighestPrice:     50000.0,
			CurrentStopPrice: 48750.0,
			Status:           StopOrderStatusActive,
			CreatedAt:        time.Now().Unix(),
		},
	}

	for _, order := range orders {
		if err := repo.SaveTrailingStopOrder(order); err != nil {
			t.Fatalf("SaveTrailingStopOrder failed: %v", err)
		}
	}

	// Find active orders
	activeOrders, err := repo.FindActiveTrailingStopOrders("BTCUSDT")
	if err != nil {
		t.Fatalf("FindActiveTrailingStopOrders failed: %v", err)
	}

	if len(activeOrders) != 2 {
		t.Errorf("Expected 2 active trailing stop orders, got %d", len(activeOrders))
	}

	// Verify all returned orders are active
	for _, order := range activeOrders {
		if order.Status != StopOrderStatusActive {
			t.Errorf("Expected active order, got status %s", order.Status)
		}
	}
}
