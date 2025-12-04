package repository

import (
	"binance-trader/internal/api"
	"binance-trader/pkg/errors"
	"sync"
)

// OrderRepository defines the interface for order data persistence
type OrderRepository interface {
	// CRUD operations
	Save(order *api.Order) error
	FindByID(orderID int64) (*api.Order, error)
	FindBySymbol(symbol string) ([]*api.Order, error)
	Update(order *api.Order) error
	Delete(orderID int64) error

	// Query operations
	FindOpenOrders() ([]*api.Order, error)
	FindOrdersByTimeRange(startTime, endTime int64) ([]*api.Order, error)
	
	// Sync operation
	SyncOrderStatus(orderID int64, newStatus api.OrderStatus, executedQty float64, updateTime int64) error
}

// memoryOrderRepository implements OrderRepository using in-memory storage
type memoryOrderRepository struct {
	mu     sync.RWMutex
	orders map[int64]*api.Order
}

// NewMemoryOrderRepository creates a new in-memory order repository
func NewMemoryOrderRepository() OrderRepository {
	return &memoryOrderRepository{
		orders: make(map[int64]*api.Order),
	}
}

// Save stores a new order
func (r *memoryOrderRepository) Save(order *api.Order) error {
	if order == nil {
		return errors.NewTradingError(errors.ErrInvalidParameter, "order cannot be nil", 0, nil)
	}
	
	if order.OrderID <= 0 {
		return errors.NewTradingError(errors.ErrInvalidParameter, "order ID must be greater than 0", 0, nil)
	}
	
	r.mu.Lock()
	defer r.mu.Unlock()
	
	// Create a copy to avoid external modifications
	orderCopy := *order
	r.orders[order.OrderID] = &orderCopy
	
	return nil
}

// FindByID retrieves an order by its ID
func (r *memoryOrderRepository) FindByID(orderID int64) (*api.Order, error) {
	if orderID <= 0 {
		return nil, errors.NewTradingError(errors.ErrInvalidParameter, "order ID must be greater than 0", 0, nil)
	}
	
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	order, exists := r.orders[orderID]
	if !exists {
		return nil, errors.NewTradingError(errors.ErrOrderNotFound, "order not found", 0, nil)
	}
	
	// Return a copy to prevent external modifications
	orderCopy := *order
	return &orderCopy, nil
}

// FindBySymbol retrieves all orders for a specific symbol
func (r *memoryOrderRepository) FindBySymbol(symbol string) ([]*api.Order, error) {
	if symbol == "" {
		return nil, errors.NewTradingError(errors.ErrInvalidParameter, "symbol cannot be empty", 0, nil)
	}
	
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	var result []*api.Order
	for _, order := range r.orders {
		if order.Symbol == symbol {
			orderCopy := *order
			result = append(result, &orderCopy)
		}
	}
	
	return result, nil
}

// Update updates an existing order
func (r *memoryOrderRepository) Update(order *api.Order) error {
	if order == nil {
		return errors.NewTradingError(errors.ErrInvalidParameter, "order cannot be nil", 0, nil)
	}
	
	if order.OrderID <= 0 {
		return errors.NewTradingError(errors.ErrInvalidParameter, "order ID must be greater than 0", 0, nil)
	}
	
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if _, exists := r.orders[order.OrderID]; !exists {
		return errors.NewTradingError(errors.ErrOrderNotFound, "order not found", 0, nil)
	}
	
	// Create a copy to avoid external modifications
	orderCopy := *order
	r.orders[order.OrderID] = &orderCopy
	
	return nil
}

// Delete removes an order by its ID
func (r *memoryOrderRepository) Delete(orderID int64) error {
	if orderID <= 0 {
		return errors.NewTradingError(errors.ErrInvalidParameter, "order ID must be greater than 0", 0, nil)
	}
	
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if _, exists := r.orders[orderID]; !exists {
		return errors.NewTradingError(errors.ErrOrderNotFound, "order not found", 0, nil)
	}
	
	delete(r.orders, orderID)
	return nil
}

// FindOpenOrders retrieves all orders with status NEW or PARTIALLY_FILLED
func (r *memoryOrderRepository) FindOpenOrders() ([]*api.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	var result []*api.Order
	for _, order := range r.orders {
		if order.Status == api.OrderStatusNew || order.Status == api.OrderStatusPartiallyFilled {
			orderCopy := *order
			result = append(result, &orderCopy)
		}
	}
	
	return result, nil
}

// FindOrdersByTimeRange retrieves orders within a time range
func (r *memoryOrderRepository) FindOrdersByTimeRange(startTime, endTime int64) ([]*api.Order, error) {
	if startTime < 0 || endTime < 0 {
		return nil, errors.NewTradingError(errors.ErrInvalidParameter, "time values cannot be negative", 0, nil)
	}
	
	if startTime > endTime {
		return nil, errors.NewTradingError(errors.ErrInvalidParameter, "start time cannot be after end time", 0, nil)
	}
	
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	var result []*api.Order
	for _, order := range r.orders {
		if order.Time >= startTime && order.Time <= endTime {
			orderCopy := *order
			result = append(result, &orderCopy)
		}
	}
	
	return result, nil
}

// SyncOrderStatus synchronizes order status with API response
func (r *memoryOrderRepository) SyncOrderStatus(orderID int64, newStatus api.OrderStatus, executedQty float64, updateTime int64) error {
	if orderID <= 0 {
		return errors.NewTradingError(errors.ErrInvalidParameter, "order ID must be greater than 0", 0, nil)
	}
	
	r.mu.Lock()
	defer r.mu.Unlock()
	
	order, exists := r.orders[orderID]
	if !exists {
		return errors.NewTradingError(errors.ErrOrderNotFound, "order not found", 0, nil)
	}
	
	// Update order status fields
	order.Status = newStatus
	order.ExecutedQty = executedQty
	order.UpdateTime = updateTime
	
	return nil
}
