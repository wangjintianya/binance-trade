package repository

import (
	"binance-trader/internal/api"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Feature: binance-auto-trading, Property 15: 订单状态同步一致性
// Validates: Requirements 4.5
// For any order, when fetching the latest status from API, the local stored order status should be consistent with the API returned status
func TestProperty_OrderStatusSyncConsistency(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("syncing order status updates local state to match API state", prop.ForAll(
		func(orderID int64, initialStatus api.OrderStatus, apiStatus api.OrderStatus, apiExecutedQty float64, apiUpdateTime int64) bool {
			// Create repository
			repo := NewMemoryOrderRepository()

			// Create initial order
			initialOrder := &api.Order{
				OrderID:     orderID,
				Symbol:      "BTCUSDT",
				Side:        api.OrderSideBuy,
				Type:        api.OrderTypeMarket,
				Status:      initialStatus,
				Price:       50000.0,
				OrigQty:     1.0,
				ExecutedQty: 0.0,
				Time:        time.Now().Unix(),
				UpdateTime:  time.Now().Unix(),
			}

			// Save initial order
			if err := repo.Save(initialOrder); err != nil {
				return false
			}

			// Sync with API status
			if err := repo.SyncOrderStatus(orderID, apiStatus, apiExecutedQty, apiUpdateTime); err != nil {
				return false
			}

			// Retrieve order and verify it matches API state
			syncedOrder, err := repo.FindByID(orderID)
			if err != nil {
				return false
			}

			// Verify status consistency
			return syncedOrder.Status == apiStatus &&
				syncedOrder.ExecutedQty == apiExecutedQty &&
				syncedOrder.UpdateTime == apiUpdateTime
		},
		gen.Int64Range(1, 1000000),                                                                                                                    // orderID
		gen.OneConstOf(api.OrderStatusNew, api.OrderStatusPartiallyFilled, api.OrderStatusFilled, api.OrderStatusCanceled, api.OrderStatusRejected), // initialStatus
		gen.OneConstOf(api.OrderStatusNew, api.OrderStatusPartiallyFilled, api.OrderStatusFilled, api.OrderStatusCanceled, api.OrderStatusRejected), // apiStatus
		gen.Float64Range(0, 10.0),                                                                                                                     // apiExecutedQty
		gen.Int64Range(1000000000, 9999999999),                                                                                                        // apiUpdateTime
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Unit tests for OrderRepository

func TestSave_ValidOrder(t *testing.T) {
	repo := NewMemoryOrderRepository()
	
	order := &api.Order{
		OrderID:     12345,
		Symbol:      "BTCUSDT",
		Side:        api.OrderSideBuy,
		Type:        api.OrderTypeMarket,
		Status:      api.OrderStatusNew,
		Price:       50000.0,
		OrigQty:     1.0,
		ExecutedQty: 0.0,
		Time:        time.Now().Unix(),
		UpdateTime:  time.Now().Unix(),
	}
	
	err := repo.Save(order)
	if err != nil {
		t.Errorf("Save() failed: %v", err)
	}
	
	// Verify order was saved
	retrieved, err := repo.FindByID(12345)
	if err != nil {
		t.Errorf("FindByID() failed: %v", err)
	}
	
	if retrieved.OrderID != order.OrderID {
		t.Errorf("Expected OrderID %d, got %d", order.OrderID, retrieved.OrderID)
	}
}

func TestSave_NilOrder(t *testing.T) {
	repo := NewMemoryOrderRepository()
	
	err := repo.Save(nil)
	if err == nil {
		t.Error("Expected error when saving nil order")
	}
}

func TestSave_InvalidOrderID(t *testing.T) {
	repo := NewMemoryOrderRepository()
	
	order := &api.Order{
		OrderID: 0,
		Symbol:  "BTCUSDT",
	}
	
	err := repo.Save(order)
	if err == nil {
		t.Error("Expected error when saving order with invalid ID")
	}
}

func TestFindByID_ExistingOrder(t *testing.T) {
	repo := NewMemoryOrderRepository()
	
	order := &api.Order{
		OrderID: 12345,
		Symbol:  "BTCUSDT",
		Status:  api.OrderStatusNew,
	}
	
	repo.Save(order)
	
	retrieved, err := repo.FindByID(12345)
	if err != nil {
		t.Errorf("FindByID() failed: %v", err)
	}
	
	if retrieved.OrderID != 12345 {
		t.Errorf("Expected OrderID 12345, got %d", retrieved.OrderID)
	}
}

func TestFindByID_NonExistingOrder(t *testing.T) {
	repo := NewMemoryOrderRepository()
	
	_, err := repo.FindByID(99999)
	if err == nil {
		t.Error("Expected error when finding non-existing order")
	}
}

func TestFindBySymbol_MultipleOrders(t *testing.T) {
	repo := NewMemoryOrderRepository()
	
	// Save orders for different symbols
	orders := []*api.Order{
		{OrderID: 1, Symbol: "BTCUSDT", Status: api.OrderStatusNew},
		{OrderID: 2, Symbol: "ETHUSDT", Status: api.OrderStatusNew},
		{OrderID: 3, Symbol: "BTCUSDT", Status: api.OrderStatusFilled},
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

func TestUpdate_ExistingOrder(t *testing.T) {
	repo := NewMemoryOrderRepository()
	
	order := &api.Order{
		OrderID: 12345,
		Symbol:  "BTCUSDT",
		Status:  api.OrderStatusNew,
	}
	
	repo.Save(order)
	
	// Update order status
	order.Status = api.OrderStatusFilled
	err := repo.Update(order)
	if err != nil {
		t.Errorf("Update() failed: %v", err)
	}
	
	// Verify update
	retrieved, _ := repo.FindByID(12345)
	if retrieved.Status != api.OrderStatusFilled {
		t.Errorf("Expected status FILLED, got %s", retrieved.Status)
	}
}

func TestUpdate_NonExistingOrder(t *testing.T) {
	repo := NewMemoryOrderRepository()
	
	order := &api.Order{
		OrderID: 99999,
		Symbol:  "BTCUSDT",
		Status:  api.OrderStatusNew,
	}
	
	err := repo.Update(order)
	if err == nil {
		t.Error("Expected error when updating non-existing order")
	}
}

func TestDelete_ExistingOrder(t *testing.T) {
	repo := NewMemoryOrderRepository()
	
	order := &api.Order{
		OrderID: 12345,
		Symbol:  "BTCUSDT",
		Status:  api.OrderStatusNew,
	}
	
	repo.Save(order)
	
	err := repo.Delete(12345)
	if err != nil {
		t.Errorf("Delete() failed: %v", err)
	}
	
	// Verify deletion
	_, err = repo.FindByID(12345)
	if err == nil {
		t.Error("Expected error when finding deleted order")
	}
}

func TestFindOpenOrders_FiltersCorrectly(t *testing.T) {
	repo := NewMemoryOrderRepository()
	
	orders := []*api.Order{
		{OrderID: 1, Symbol: "BTCUSDT", Status: api.OrderStatusNew},
		{OrderID: 2, Symbol: "ETHUSDT", Status: api.OrderStatusPartiallyFilled},
		{OrderID: 3, Symbol: "BTCUSDT", Status: api.OrderStatusFilled},
		{OrderID: 4, Symbol: "BNBUSDT", Status: api.OrderStatusCanceled},
	}
	
	for _, order := range orders {
		repo.Save(order)
	}
	
	openOrders, err := repo.FindOpenOrders()
	if err != nil {
		t.Errorf("FindOpenOrders() failed: %v", err)
	}
	
	if len(openOrders) != 2 {
		t.Errorf("Expected 2 open orders, got %d", len(openOrders))
	}
	
	// Verify only NEW and PARTIALLY_FILLED orders are returned
	for _, order := range openOrders {
		if order.Status != api.OrderStatusNew && order.Status != api.OrderStatusPartiallyFilled {
			t.Errorf("Unexpected order status in open orders: %s", order.Status)
		}
	}
}

func TestFindOrdersByTimeRange_FiltersCorrectly(t *testing.T) {
	repo := NewMemoryOrderRepository()
	
	now := time.Now().Unix()
	
	orders := []*api.Order{
		{OrderID: 1, Symbol: "BTCUSDT", Status: api.OrderStatusNew, Time: now - 3600},
		{OrderID: 2, Symbol: "ETHUSDT", Status: api.OrderStatusNew, Time: now - 1800},
		{OrderID: 3, Symbol: "BTCUSDT", Status: api.OrderStatusNew, Time: now},
		{OrderID: 4, Symbol: "BNBUSDT", Status: api.OrderStatusNew, Time: now + 3600},
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

func TestConcurrentAccess(t *testing.T) {
	repo := NewMemoryOrderRepository()
	
	// Test concurrent writes
	done := make(chan bool, 20) // Buffered channel to prevent blocking
	
	for i := 1; i <= 10; i++ { // Start from 1 to avoid 0 orderID
		go func(id int) {
			order := &api.Order{
				OrderID: int64(id),
				Symbol:  "BTCUSDT",
				Status:  api.OrderStatusNew,
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
			repo.FindByID(int64(id))
			done <- true
		}(i)
	}
	
	// Wait for all read goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
	
	// Verify all orders were saved
	for i := 1; i <= 10; i++ {
		_, err := repo.FindByID(int64(i))
		if err != nil {
			t.Errorf("Order %d not found after concurrent access", i)
		}
	}
}
