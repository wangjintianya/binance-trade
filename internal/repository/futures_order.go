package repository

import (
	"binance-trader/internal/api"
	"binance-trader/pkg/errors"
	"sync"
)

// FuturesOrderRepository defines the interface for futures order data persistence
type FuturesOrderRepository interface {
	// CRUD operations
	Save(order *api.FuturesOrder) error
	FindByID(orderID int64) (*api.FuturesOrder, error)
	FindBySymbol(symbol string) ([]*api.FuturesOrder, error)
	Update(order *api.FuturesOrder) error
	Delete(orderID int64) error

	// Query operations
	FindOpenOrders() ([]*api.FuturesOrder, error)
	FindOrdersByTimeRange(startTime, endTime int64) ([]*api.FuturesOrder, error)
	FindOrdersByPositionSide(symbol string, positionSide api.PositionSide) ([]*api.FuturesOrder, error)
	
	// Sync operation
	SyncOrderStatus(orderID int64, newStatus api.OrderStatus, executedQty float64, avgPrice float64, updateTime int64) error
}

// memoryFuturesOrderRepository implements FuturesOrderRepository using in-memory storage
type memoryFuturesOrderRepository struct {
	mu     sync.RWMutex
	orders map[int64]*api.FuturesOrder
}

// NewMemoryFuturesOrderRepository creates a new in-memory futures order repository
func NewMemoryFuturesOrderRepository() FuturesOrderRepository {
	return &memoryFuturesOrderRepository{
		orders: make(map[int64]*api.FuturesOrder),
	}
}

// Save stores a new futures order
func (r *memoryFuturesOrderRepository) Save(order *api.FuturesOrder) error {
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

// FindByID retrieves a futures order by its ID
func (r *memoryFuturesOrderRepository) FindByID(orderID int64) (*api.FuturesOrder, error) {
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

// FindBySymbol retrieves all futures orders for a specific symbol
func (r *memoryFuturesOrderRepository) FindBySymbol(symbol string) ([]*api.FuturesOrder, error) {
	if symbol == "" {
		return nil, errors.NewTradingError(errors.ErrInvalidParameter, "symbol cannot be empty", 0, nil)
	}
	
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	var result []*api.FuturesOrder
	for _, order := range r.orders {
		if order.Symbol == symbol {
			orderCopy := *order
			result = append(result, &orderCopy)
		}
	}
	
	return result, nil
}

// Update updates an existing futures order
func (r *memoryFuturesOrderRepository) Update(order *api.FuturesOrder) error {
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

// Delete removes a futures order by its ID
func (r *memoryFuturesOrderRepository) Delete(orderID int64) error {
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

// FindOpenOrders retrieves all futures orders with status NEW or PARTIALLY_FILLED
func (r *memoryFuturesOrderRepository) FindOpenOrders() ([]*api.FuturesOrder, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	var result []*api.FuturesOrder
	for _, order := range r.orders {
		if order.Status == api.OrderStatusNew || order.Status == api.OrderStatusPartiallyFilled {
			orderCopy := *order
			result = append(result, &orderCopy)
		}
	}
	
	return result, nil
}

// FindOrdersByTimeRange retrieves futures orders within a time range
func (r *memoryFuturesOrderRepository) FindOrdersByTimeRange(startTime, endTime int64) ([]*api.FuturesOrder, error) {
	if startTime < 0 || endTime < 0 {
		return nil, errors.NewTradingError(errors.ErrInvalidParameter, "time values cannot be negative", 0, nil)
	}
	
	if startTime > endTime {
		return nil, errors.NewTradingError(errors.ErrInvalidParameter, "start time cannot be after end time", 0, nil)
	}
	
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	var result []*api.FuturesOrder
	for _, order := range r.orders {
		if order.Time >= startTime && order.Time <= endTime {
			orderCopy := *order
			result = append(result, &orderCopy)
		}
	}
	
	return result, nil
}

// FindOrdersByPositionSide retrieves futures orders for a specific symbol and position side
func (r *memoryFuturesOrderRepository) FindOrdersByPositionSide(symbol string, positionSide api.PositionSide) ([]*api.FuturesOrder, error) {
	if symbol == "" {
		return nil, errors.NewTradingError(errors.ErrInvalidParameter, "symbol cannot be empty", 0, nil)
	}
	
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	var result []*api.FuturesOrder
	for _, order := range r.orders {
		if order.Symbol == symbol && order.PositionSide == positionSide {
			orderCopy := *order
			result = append(result, &orderCopy)
		}
	}
	
	return result, nil
}

// SyncOrderStatus synchronizes futures order status with API response
func (r *memoryFuturesOrderRepository) SyncOrderStatus(orderID int64, newStatus api.OrderStatus, executedQty float64, avgPrice float64, updateTime int64) error {
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
	order.AvgPrice = avgPrice
	order.UpdateTime = updateTime
	
	return nil
}
