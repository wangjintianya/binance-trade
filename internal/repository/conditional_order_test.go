package repository

import (
	"binance-trader/internal/api"
	"fmt"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Feature: binance-auto-trading, Property 37: 条件订单状态转换
// Validates: Requirements 10.4
// For any conditional order, when trigger condition is met and executed, order status must change from PENDING to EXECUTED
func TestProperty_ConditionalOrderStatusTransition(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("triggering and executing order transitions status from PENDING to EXECUTED", prop.ForAll(
		func(orderID string, symbol string, quantity float64, price float64, executedOrderID int64) bool {
			// Create repository
			repo := NewMemoryConditionalOrderRepository()

			// Create initial conditional order with PENDING status
			initialOrder := &ConditionalOrder{
				OrderID:  orderID,
				Symbol:   symbol,
				Side:     api.OrderSideBuy,
				Type:     api.OrderTypeLimit,
				Quantity: quantity,
				Price:    price,
				TriggerCondition: &TriggerCondition{
					Type:     TriggerTypePrice,
					Operator: OperatorGreaterThan,
					Value:    price * 0.9,
				},
				Status:    ConditionalOrderStatusPending,
				CreatedAt: time.Now().Unix(),
			}

			// Save initial order
			if err := repo.Save(initialOrder); err != nil {
				return false
			}

			// Verify initial status is PENDING
			retrieved, err := repo.FindByID(orderID)
			if err != nil || retrieved.Status != ConditionalOrderStatusPending {
				return false
			}

			// Simulate trigger and execution by updating status to EXECUTED
			triggeredAt := time.Now().Unix()
			if err := repo.UpdateStatus(orderID, ConditionalOrderStatusExecuted, triggeredAt, executedOrderID); err != nil {
				return false
			}

			// Retrieve order and verify status changed to EXECUTED
			executedOrder, err := repo.FindByID(orderID)
			if err != nil {
				return false
			}

			// Verify status transition
			return executedOrder.Status == ConditionalOrderStatusExecuted &&
				executedOrder.TriggeredAt == triggeredAt &&
				executedOrder.ExecutedOrderID == executedOrderID
		},
		gen.Identifier().SuchThat(func(v string) bool { return v != "" }),  // orderID
		gen.OneConstOf("BTCUSDT", "ETHUSDT", "BNBUSDT"),                    // symbol
		gen.Float64Range(0.001, 100.0),                                     // quantity
		gen.Float64Range(100.0, 100000.0),                                  // price
		gen.Int64Range(1, 1000000),                                         // executedOrderID
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Unit tests for ConditionalOrderRepository

func TestSave_ValidConditionalOrder(t *testing.T) {
	repo := NewMemoryConditionalOrderRepository()

	order := &ConditionalOrder{
		OrderID:  "cond-12345",
		Symbol:   "BTCUSDT",
		Side:     api.OrderSideBuy,
		Type:     api.OrderTypeLimit,
		Quantity: 1.0,
		Price:    50000.0,
		TriggerCondition: &TriggerCondition{
			Type:     TriggerTypePrice,
			Operator: OperatorGreaterThan,
			Value:    49000.0,
		},
		Status:    ConditionalOrderStatusPending,
		CreatedAt: time.Now().Unix(),
	}

	err := repo.Save(order)
	if err != nil {
		t.Errorf("Save() failed: %v", err)
	}

	// Verify order was saved
	retrieved, err := repo.FindByID("cond-12345")
	if err != nil {
		t.Errorf("FindByID() failed: %v", err)
	}

	if retrieved.OrderID != order.OrderID {
		t.Errorf("Expected OrderID %s, got %s", order.OrderID, retrieved.OrderID)
	}
}

func TestSave_NilConditionalOrder(t *testing.T) {
	repo := NewMemoryConditionalOrderRepository()

	err := repo.Save(nil)
	if err == nil {
		t.Error("Expected error when saving nil order")
	}
}

func TestSave_EmptyOrderID(t *testing.T) {
	repo := NewMemoryConditionalOrderRepository()

	order := &ConditionalOrder{
		OrderID: "",
		Symbol:  "BTCUSDT",
	}

	err := repo.Save(order)
	if err == nil {
		t.Error("Expected error when saving order with empty ID")
	}
}

func TestFindByID_ExistingConditionalOrder(t *testing.T) {
	repo := NewMemoryConditionalOrderRepository()

	order := &ConditionalOrder{
		OrderID: "cond-12345",
		Symbol:  "BTCUSDT",
		Status:  ConditionalOrderStatusPending,
		TriggerCondition: &TriggerCondition{
			Type:     TriggerTypePrice,
			Operator: OperatorGreaterThan,
			Value:    50000.0,
		},
	}

	repo.Save(order)

	retrieved, err := repo.FindByID("cond-12345")
	if err != nil {
		t.Errorf("FindByID() failed: %v", err)
	}

	if retrieved.OrderID != "cond-12345" {
		t.Errorf("Expected OrderID cond-12345, got %s", retrieved.OrderID)
	}
}

func TestFindByID_NonExistingConditionalOrder(t *testing.T) {
	repo := NewMemoryConditionalOrderRepository()

	_, err := repo.FindByID("non-existent")
	if err == nil {
		t.Error("Expected error when finding non-existing order")
	}
}

func TestFindBySymbol_MultipleConditionalOrders(t *testing.T) {
	repo := NewMemoryConditionalOrderRepository()

	// Save orders for different symbols
	orders := []*ConditionalOrder{
		{OrderID: "cond-1", Symbol: "BTCUSDT", Status: ConditionalOrderStatusPending},
		{OrderID: "cond-2", Symbol: "ETHUSDT", Status: ConditionalOrderStatusPending},
		{OrderID: "cond-3", Symbol: "BTCUSDT", Status: ConditionalOrderStatusExecuted},
	}

	for _, order := range orders {
		repo.Save(order)
	}

	// Find BTCUSDT orders
	btcOrders, err := repo.FindBySymbol("BTCUSDT")
	if err != nil {
		t.Errorf("FindBySymbol() failed: %v", err)
	}

	if len(btcOrders) != 2 {
		t.Errorf("Expected 2 BTCUSDT orders, got %d", len(btcOrders))
	}
}

func TestUpdate_ExistingConditionalOrder(t *testing.T) {
	repo := NewMemoryConditionalOrderRepository()

	order := &ConditionalOrder{
		OrderID: "cond-12345",
		Symbol:  "BTCUSDT",
		Status:  ConditionalOrderStatusPending,
		TriggerCondition: &TriggerCondition{
			Type:     TriggerTypePrice,
			Operator: OperatorGreaterThan,
			Value:    50000.0,
		},
	}

	repo.Save(order)

	// Update order status
	order.Status = ConditionalOrderStatusExecuted
	err := repo.Update(order)
	if err != nil {
		t.Errorf("Update() failed: %v", err)
	}

	// Verify update
	retrieved, _ := repo.FindByID("cond-12345")
	if retrieved.Status != ConditionalOrderStatusExecuted {
		t.Errorf("Expected status EXECUTED, got %s", retrieved.Status)
	}
}

func TestUpdate_NonExistingConditionalOrder(t *testing.T) {
	repo := NewMemoryConditionalOrderRepository()

	order := &ConditionalOrder{
		OrderID: "non-existent",
		Symbol:  "BTCUSDT",
		Status:  ConditionalOrderStatusPending,
	}

	err := repo.Update(order)
	if err == nil {
		t.Error("Expected error when updating non-existing order")
	}
}

func TestDelete_ExistingConditionalOrder(t *testing.T) {
	repo := NewMemoryConditionalOrderRepository()

	order := &ConditionalOrder{
		OrderID: "cond-12345",
		Symbol:  "BTCUSDT",
		Status:  ConditionalOrderStatusPending,
	}

	repo.Save(order)

	err := repo.Delete("cond-12345")
	if err != nil {
		t.Errorf("Delete() failed: %v", err)
	}

	// Verify deletion
	_, err = repo.FindByID("cond-12345")
	if err == nil {
		t.Error("Expected error when finding deleted order")
	}
}

func TestFindActiveOrders_FiltersCorrectly(t *testing.T) {
	repo := NewMemoryConditionalOrderRepository()

	orders := []*ConditionalOrder{
		{OrderID: "cond-1", Symbol: "BTCUSDT", Status: ConditionalOrderStatusPending},
		{OrderID: "cond-2", Symbol: "ETHUSDT", Status: ConditionalOrderStatusPending},
		{OrderID: "cond-3", Symbol: "BTCUSDT", Status: ConditionalOrderStatusExecuted},
		{OrderID: "cond-4", Symbol: "BNBUSDT", Status: ConditionalOrderStatusCancelled},
	}

	for _, order := range orders {
		repo.Save(order)
	}

	activeOrders, err := repo.FindActiveOrders()
	if err != nil {
		t.Errorf("FindActiveOrders() failed: %v", err)
	}

	if len(activeOrders) != 2 {
		t.Errorf("Expected 2 active orders, got %d", len(activeOrders))
	}

	// Verify only PENDING orders are returned
	for _, order := range activeOrders {
		if order.Status != ConditionalOrderStatusPending {
			t.Errorf("Unexpected order status in active orders: %s", order.Status)
		}
	}
}

func TestFindOrdersByStatus_FiltersCorrectly(t *testing.T) {
	repo := NewMemoryConditionalOrderRepository()

	orders := []*ConditionalOrder{
		{OrderID: "cond-1", Symbol: "BTCUSDT", Status: ConditionalOrderStatusPending},
		{OrderID: "cond-2", Symbol: "ETHUSDT", Status: ConditionalOrderStatusExecuted},
		{OrderID: "cond-3", Symbol: "BTCUSDT", Status: ConditionalOrderStatusExecuted},
		{OrderID: "cond-4", Symbol: "BNBUSDT", Status: ConditionalOrderStatusCancelled},
	}

	for _, order := range orders {
		repo.Save(order)
	}

	executedOrders, err := repo.FindOrdersByStatus(ConditionalOrderStatusExecuted)
	if err != nil {
		t.Errorf("FindOrdersByStatus() failed: %v", err)
	}

	if len(executedOrders) != 2 {
		t.Errorf("Expected 2 executed orders, got %d", len(executedOrders))
	}

	// Verify only EXECUTED orders are returned
	for _, order := range executedOrders {
		if order.Status != ConditionalOrderStatusExecuted {
			t.Errorf("Unexpected order status: %s", order.Status)
		}
	}
}

func TestFindConditionalOrdersByTimeRange_FiltersCorrectly(t *testing.T) {
	repo := NewMemoryConditionalOrderRepository()

	now := time.Now().Unix()

	orders := []*ConditionalOrder{
		{OrderID: "cond-1", Symbol: "BTCUSDT", Status: ConditionalOrderStatusPending, CreatedAt: now - 3600},
		{OrderID: "cond-2", Symbol: "ETHUSDT", Status: ConditionalOrderStatusPending, CreatedAt: now - 1800},
		{OrderID: "cond-3", Symbol: "BTCUSDT", Status: ConditionalOrderStatusPending, CreatedAt: now},
		{OrderID: "cond-4", Symbol: "BNBUSDT", Status: ConditionalOrderStatusPending, CreatedAt: now + 3600},
	}

	for _, order := range orders {
		repo.Save(order)
	}

	// Find orders in range
	rangeOrders, err := repo.FindOrdersByTimeRange(now-2000, now+1000)
	if err != nil {
		t.Errorf("FindOrdersByTimeRange() failed: %v", err)
	}

	if len(rangeOrders) != 2 {
		t.Errorf("Expected 2 orders in range, got %d", len(rangeOrders))
	}
}

func TestUpdateStatus_UpdatesCorrectly(t *testing.T) {
	repo := NewMemoryConditionalOrderRepository()

	order := &ConditionalOrder{
		OrderID:  "cond-12345",
		Symbol:   "BTCUSDT",
		Status:   ConditionalOrderStatusPending,
		CreatedAt: time.Now().Unix(),
	}

	repo.Save(order)

	// Update status
	triggeredAt := time.Now().Unix()
	executedOrderID := int64(98765)
	err := repo.UpdateStatus("cond-12345", ConditionalOrderStatusExecuted, triggeredAt, executedOrderID)
	if err != nil {
		t.Errorf("UpdateStatus() failed: %v", err)
	}

	// Verify update
	retrieved, _ := repo.FindByID("cond-12345")
	if retrieved.Status != ConditionalOrderStatusExecuted {
		t.Errorf("Expected status EXECUTED, got %s", retrieved.Status)
	}
	if retrieved.TriggeredAt != triggeredAt {
		t.Errorf("Expected TriggeredAt %d, got %d", triggeredAt, retrieved.TriggeredAt)
	}
	if retrieved.ExecutedOrderID != executedOrderID {
		t.Errorf("Expected ExecutedOrderID %d, got %d", executedOrderID, retrieved.ExecutedOrderID)
	}
}

func TestConcurrentAccess_ConditionalOrders(t *testing.T) {
	repo := NewMemoryConditionalOrderRepository()

	// Test concurrent writes
	done := make(chan bool, 20)

	for i := 1; i <= 10; i++ {
		go func(id int) {
			order := &ConditionalOrder{
				OrderID: fmt.Sprintf("cond-%d", id),
				Symbol:  "BTCUSDT",
				Status:  ConditionalOrderStatusPending,
			}
			repo.Save(order)
			done <- true
		}(i)
	}

	// Wait for all write goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Test concurrent reads
	for i := 1; i <= 10; i++ {
		go func(id int) {
			repo.FindByID(fmt.Sprintf("cond-%d", id))
			done <- true
		}(i)
	}

	// Wait for all read goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all orders were saved
	for i := 1; i <= 10; i++ {
		_, err := repo.FindByID(fmt.Sprintf("cond-%d", i))
		if err != nil {
			t.Errorf("Order cond-%d not found after concurrent access", i)
		}
	}
}
