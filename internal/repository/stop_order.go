package repository

import (
	"binance-trader/pkg/errors"
	"sync"
)

// StopOrderType represents the type of stop order
type StopOrderType int

const (
	StopOrderTypeStopLoss StopOrderType = iota
	StopOrderTypeTakeProfit
)

// StopOrderStatus represents the status of a stop order
type StopOrderStatus string

const (
	StopOrderStatusActive    StopOrderStatus = "ACTIVE"
	StopOrderStatusTriggered StopOrderStatus = "TRIGGERED"
	StopOrderStatusCancelled StopOrderStatus = "CANCELLED"
)

// StopOrder represents a stop loss or take profit order
type StopOrder struct {
	OrderID         string
	Symbol          string
	Position        float64
	StopPrice       float64
	Type            StopOrderType
	Status          StopOrderStatus
	CreatedAt       int64
	TriggeredAt     int64
	ExecutedOrderID int64
}

// StopOrderPair represents a paired stop loss and take profit order
type StopOrderPair struct {
	PairID          string
	Symbol          string
	Position        float64
	StopLossOrder   *StopOrder
	TakeProfitOrder *StopOrder
	Status          string // ACTIVE, PARTIALLY_TRIGGERED, COMPLETED
}

// TrailingStopOrder represents a trailing stop loss order
type TrailingStopOrder struct {
	OrderID          string
	Symbol           string
	Position         float64
	TrailPercent     float64
	HighestPrice     float64
	CurrentStopPrice float64
	Status           StopOrderStatus
	CreatedAt        int64
	LastUpdatedAt    int64
}

// StopOrderRepository defines the interface for stop order data persistence
type StopOrderRepository interface {
	// Stop order CRUD operations
	SaveStopOrder(order *StopOrder) error
	FindStopOrderByID(orderID string) (*StopOrder, error)
	FindStopOrdersBySymbol(symbol string) ([]*StopOrder, error)
	UpdateStopOrder(order *StopOrder) error
	DeleteStopOrder(orderID string) error

	// Stop order query operations
	FindActiveStopOrders(symbol string) ([]*StopOrder, error)
	FindStopOrdersByStatus(status StopOrderStatus) ([]*StopOrder, error)

	// Stop order status management
	UpdateStopOrderStatus(orderID string, newStatus StopOrderStatus, triggeredAt int64, executedOrderID int64) error

	// Stop order pair CRUD operations
	SaveStopOrderPair(pair *StopOrderPair) error
	FindStopOrderPairByID(pairID string) (*StopOrderPair, error)
	FindStopOrderPairsBySymbol(symbol string) ([]*StopOrderPair, error)
	UpdateStopOrderPair(pair *StopOrderPair) error
	DeleteStopOrderPair(pairID string) error

	// Stop order pair query operations
	FindActiveStopOrderPairs(symbol string) ([]*StopOrderPair, error)

	// Trailing stop order CRUD operations
	SaveTrailingStopOrder(order *TrailingStopOrder) error
	FindTrailingStopOrderByID(orderID string) (*TrailingStopOrder, error)
	FindTrailingStopOrdersBySymbol(symbol string) ([]*TrailingStopOrder, error)
	UpdateTrailingStopOrder(order *TrailingStopOrder) error
	DeleteTrailingStopOrder(orderID string) error

	// Trailing stop order query operations
	FindActiveTrailingStopOrders(symbol string) ([]*TrailingStopOrder, error)
}

// memoryStopOrderRepository implements StopOrderRepository using in-memory storage
type memoryStopOrderRepository struct {
	mu                  sync.RWMutex
	stopOrders          map[string]*StopOrder
	stopOrderPairs      map[string]*StopOrderPair
	trailingStopOrders  map[string]*TrailingStopOrder
}

// NewMemoryStopOrderRepository creates a new in-memory stop order repository
func NewMemoryStopOrderRepository() StopOrderRepository {
	return &memoryStopOrderRepository{
		stopOrders:         make(map[string]*StopOrder),
		stopOrderPairs:     make(map[string]*StopOrderPair),
		trailingStopOrders: make(map[string]*TrailingStopOrder),
	}
}

// SaveStopOrder stores a new stop order
func (r *memoryStopOrderRepository) SaveStopOrder(order *StopOrder) error {
	if order == nil {
		return errors.NewTradingError(errors.ErrInvalidParameter, "order cannot be nil", 0, nil)
	}

	if order.OrderID == "" {
		return errors.NewTradingError(errors.ErrInvalidParameter, "order ID cannot be empty", 0, nil)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Create a copy to avoid external modifications
	orderCopy := *order
	r.stopOrders[order.OrderID] = &orderCopy

	return nil
}

// FindStopOrderByID retrieves a stop order by its ID
func (r *memoryStopOrderRepository) FindStopOrderByID(orderID string) (*StopOrder, error) {
	if orderID == "" {
		return nil, errors.NewTradingError(errors.ErrInvalidParameter, "order ID cannot be empty", 0, nil)
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	order, exists := r.stopOrders[orderID]
	if !exists {
		return nil, errors.NewTradingError(errors.ErrStopOrderNotFound, "stop order not found", 0, nil)
	}

	// Return a copy to prevent external modifications
	orderCopy := *order
	return &orderCopy, nil
}

// FindStopOrdersBySymbol retrieves all stop orders for a specific symbol
func (r *memoryStopOrderRepository) FindStopOrdersBySymbol(symbol string) ([]*StopOrder, error) {
	if symbol == "" {
		return nil, errors.NewTradingError(errors.ErrInvalidParameter, "symbol cannot be empty", 0, nil)
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*StopOrder
	for _, order := range r.stopOrders {
		if order.Symbol == symbol {
			orderCopy := *order
			result = append(result, &orderCopy)
		}
	}

	return result, nil
}

// UpdateStopOrder updates an existing stop order
func (r *memoryStopOrderRepository) UpdateStopOrder(order *StopOrder) error {
	if order == nil {
		return errors.NewTradingError(errors.ErrInvalidParameter, "order cannot be nil", 0, nil)
	}

	if order.OrderID == "" {
		return errors.NewTradingError(errors.ErrInvalidParameter, "order ID cannot be empty", 0, nil)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.stopOrders[order.OrderID]; !exists {
		return errors.NewTradingError(errors.ErrStopOrderNotFound, "stop order not found", 0, nil)
	}

	// Create a copy to avoid external modifications
	orderCopy := *order
	r.stopOrders[order.OrderID] = &orderCopy

	return nil
}

// DeleteStopOrder removes a stop order by its ID
func (r *memoryStopOrderRepository) DeleteStopOrder(orderID string) error {
	if orderID == "" {
		return errors.NewTradingError(errors.ErrInvalidParameter, "order ID cannot be empty", 0, nil)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.stopOrders[orderID]; !exists {
		return errors.NewTradingError(errors.ErrStopOrderNotFound, "stop order not found", 0, nil)
	}

	delete(r.stopOrders, orderID)
	return nil
}

// FindActiveStopOrders retrieves all active stop orders for a specific symbol
func (r *memoryStopOrderRepository) FindActiveStopOrders(symbol string) ([]*StopOrder, error) {
	if symbol == "" {
		return nil, errors.NewTradingError(errors.ErrInvalidParameter, "symbol cannot be empty", 0, nil)
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*StopOrder
	for _, order := range r.stopOrders {
		if order.Symbol == symbol && order.Status == StopOrderStatusActive {
			orderCopy := *order
			result = append(result, &orderCopy)
		}
	}

	return result, nil
}

// FindStopOrdersByStatus retrieves all stop orders with a specific status
func (r *memoryStopOrderRepository) FindStopOrdersByStatus(status StopOrderStatus) ([]*StopOrder, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*StopOrder
	for _, order := range r.stopOrders {
		if order.Status == status {
			orderCopy := *order
			result = append(result, &orderCopy)
		}
	}

	return result, nil
}

// UpdateStopOrderStatus updates the status of a stop order
func (r *memoryStopOrderRepository) UpdateStopOrderStatus(orderID string, newStatus StopOrderStatus, triggeredAt int64, executedOrderID int64) error {
	if orderID == "" {
		return errors.NewTradingError(errors.ErrInvalidParameter, "order ID cannot be empty", 0, nil)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	order, exists := r.stopOrders[orderID]
	if !exists {
		return errors.NewTradingError(errors.ErrStopOrderNotFound, "stop order not found", 0, nil)
	}

	// Update order status fields
	order.Status = newStatus
	if triggeredAt > 0 {
		order.TriggeredAt = triggeredAt
	}
	if executedOrderID > 0 {
		order.ExecutedOrderID = executedOrderID
	}

	return nil
}

// SaveStopOrderPair stores a new stop order pair
func (r *memoryStopOrderRepository) SaveStopOrderPair(pair *StopOrderPair) error {
	if pair == nil {
		return errors.NewTradingError(errors.ErrInvalidParameter, "pair cannot be nil", 0, nil)
	}

	if pair.PairID == "" {
		return errors.NewTradingError(errors.ErrInvalidParameter, "pair ID cannot be empty", 0, nil)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Create a copy to avoid external modifications
	pairCopy := *pair
	if pair.StopLossOrder != nil {
		stopLossCopy := *pair.StopLossOrder
		pairCopy.StopLossOrder = &stopLossCopy
	}
	if pair.TakeProfitOrder != nil {
		takeProfitCopy := *pair.TakeProfitOrder
		pairCopy.TakeProfitOrder = &takeProfitCopy
	}

	r.stopOrderPairs[pair.PairID] = &pairCopy

	return nil
}

// FindStopOrderPairByID retrieves a stop order pair by its ID
func (r *memoryStopOrderRepository) FindStopOrderPairByID(pairID string) (*StopOrderPair, error) {
	if pairID == "" {
		return nil, errors.NewTradingError(errors.ErrInvalidParameter, "pair ID cannot be empty", 0, nil)
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	pair, exists := r.stopOrderPairs[pairID]
	if !exists {
		return nil, errors.NewTradingError(errors.ErrStopOrderNotFound, "stop order pair not found", 0, nil)
	}

	// Return a copy to prevent external modifications
	pairCopy := *pair
	if pair.StopLossOrder != nil {
		stopLossCopy := *pair.StopLossOrder
		pairCopy.StopLossOrder = &stopLossCopy
	}
	if pair.TakeProfitOrder != nil {
		takeProfitCopy := *pair.TakeProfitOrder
		pairCopy.TakeProfitOrder = &takeProfitCopy
	}

	return &pairCopy, nil
}

// FindStopOrderPairsBySymbol retrieves all stop order pairs for a specific symbol
func (r *memoryStopOrderRepository) FindStopOrderPairsBySymbol(symbol string) ([]*StopOrderPair, error) {
	if symbol == "" {
		return nil, errors.NewTradingError(errors.ErrInvalidParameter, "symbol cannot be empty", 0, nil)
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*StopOrderPair
	for _, pair := range r.stopOrderPairs {
		if pair.Symbol == symbol {
			pairCopy := *pair
			if pair.StopLossOrder != nil {
				stopLossCopy := *pair.StopLossOrder
				pairCopy.StopLossOrder = &stopLossCopy
			}
			if pair.TakeProfitOrder != nil {
				takeProfitCopy := *pair.TakeProfitOrder
				pairCopy.TakeProfitOrder = &takeProfitCopy
			}
			result = append(result, &pairCopy)
		}
	}

	return result, nil
}

// UpdateStopOrderPair updates an existing stop order pair
func (r *memoryStopOrderRepository) UpdateStopOrderPair(pair *StopOrderPair) error {
	if pair == nil {
		return errors.NewTradingError(errors.ErrInvalidParameter, "pair cannot be nil", 0, nil)
	}

	if pair.PairID == "" {
		return errors.NewTradingError(errors.ErrInvalidParameter, "pair ID cannot be empty", 0, nil)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.stopOrderPairs[pair.PairID]; !exists {
		return errors.NewTradingError(errors.ErrStopOrderNotFound, "stop order pair not found", 0, nil)
	}

	// Create a copy to avoid external modifications
	pairCopy := *pair
	if pair.StopLossOrder != nil {
		stopLossCopy := *pair.StopLossOrder
		pairCopy.StopLossOrder = &stopLossCopy
	}
	if pair.TakeProfitOrder != nil {
		takeProfitCopy := *pair.TakeProfitOrder
		pairCopy.TakeProfitOrder = &takeProfitCopy
	}

	r.stopOrderPairs[pair.PairID] = &pairCopy

	return nil
}

// DeleteStopOrderPair removes a stop order pair by its ID
func (r *memoryStopOrderRepository) DeleteStopOrderPair(pairID string) error {
	if pairID == "" {
		return errors.NewTradingError(errors.ErrInvalidParameter, "pair ID cannot be empty", 0, nil)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.stopOrderPairs[pairID]; !exists {
		return errors.NewTradingError(errors.ErrStopOrderNotFound, "stop order pair not found", 0, nil)
	}

	delete(r.stopOrderPairs, pairID)
	return nil
}

// FindActiveStopOrderPairs retrieves all active stop order pairs for a specific symbol
func (r *memoryStopOrderRepository) FindActiveStopOrderPairs(symbol string) ([]*StopOrderPair, error) {
	if symbol == "" {
		return nil, errors.NewTradingError(errors.ErrInvalidParameter, "symbol cannot be empty", 0, nil)
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*StopOrderPair
	for _, pair := range r.stopOrderPairs {
		if pair.Symbol == symbol && pair.Status == "ACTIVE" {
			pairCopy := *pair
			if pair.StopLossOrder != nil {
				stopLossCopy := *pair.StopLossOrder
				pairCopy.StopLossOrder = &stopLossCopy
			}
			if pair.TakeProfitOrder != nil {
				takeProfitCopy := *pair.TakeProfitOrder
				pairCopy.TakeProfitOrder = &takeProfitCopy
			}
			result = append(result, &pairCopy)
		}
	}

	return result, nil
}

// SaveTrailingStopOrder stores a new trailing stop order
func (r *memoryStopOrderRepository) SaveTrailingStopOrder(order *TrailingStopOrder) error {
	if order == nil {
		return errors.NewTradingError(errors.ErrInvalidParameter, "order cannot be nil", 0, nil)
	}

	if order.OrderID == "" {
		return errors.NewTradingError(errors.ErrInvalidParameter, "order ID cannot be empty", 0, nil)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Create a copy to avoid external modifications
	orderCopy := *order
	r.trailingStopOrders[order.OrderID] = &orderCopy

	return nil
}

// FindTrailingStopOrderByID retrieves a trailing stop order by its ID
func (r *memoryStopOrderRepository) FindTrailingStopOrderByID(orderID string) (*TrailingStopOrder, error) {
	if orderID == "" {
		return nil, errors.NewTradingError(errors.ErrInvalidParameter, "order ID cannot be empty", 0, nil)
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	order, exists := r.trailingStopOrders[orderID]
	if !exists {
		return nil, errors.NewTradingError(errors.ErrStopOrderNotFound, "trailing stop order not found", 0, nil)
	}

	// Return a copy to prevent external modifications
	orderCopy := *order
	return &orderCopy, nil
}

// FindTrailingStopOrdersBySymbol retrieves all trailing stop orders for a specific symbol
func (r *memoryStopOrderRepository) FindTrailingStopOrdersBySymbol(symbol string) ([]*TrailingStopOrder, error) {
	if symbol == "" {
		return nil, errors.NewTradingError(errors.ErrInvalidParameter, "symbol cannot be empty", 0, nil)
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*TrailingStopOrder
	for _, order := range r.trailingStopOrders {
		if order.Symbol == symbol {
			orderCopy := *order
			result = append(result, &orderCopy)
		}
	}

	return result, nil
}

// UpdateTrailingStopOrder updates an existing trailing stop order
func (r *memoryStopOrderRepository) UpdateTrailingStopOrder(order *TrailingStopOrder) error {
	if order == nil {
		return errors.NewTradingError(errors.ErrInvalidParameter, "order cannot be nil", 0, nil)
	}

	if order.OrderID == "" {
		return errors.NewTradingError(errors.ErrInvalidParameter, "order ID cannot be empty", 0, nil)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.trailingStopOrders[order.OrderID]; !exists {
		return errors.NewTradingError(errors.ErrStopOrderNotFound, "trailing stop order not found", 0, nil)
	}

	// Create a copy to avoid external modifications
	orderCopy := *order
	r.trailingStopOrders[order.OrderID] = &orderCopy

	return nil
}

// DeleteTrailingStopOrder removes a trailing stop order by its ID
func (r *memoryStopOrderRepository) DeleteTrailingStopOrder(orderID string) error {
	if orderID == "" {
		return errors.NewTradingError(errors.ErrInvalidParameter, "order ID cannot be empty", 0, nil)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.trailingStopOrders[orderID]; !exists {
		return errors.NewTradingError(errors.ErrStopOrderNotFound, "trailing stop order not found", 0, nil)
	}

	delete(r.trailingStopOrders, orderID)
	return nil
}

// FindActiveTrailingStopOrders retrieves all active trailing stop orders for a specific symbol
func (r *memoryStopOrderRepository) FindActiveTrailingStopOrders(symbol string) ([]*TrailingStopOrder, error) {
	if symbol == "" {
		return nil, errors.NewTradingError(errors.ErrInvalidParameter, "symbol cannot be empty", 0, nil)
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*TrailingStopOrder
	for _, order := range r.trailingStopOrders {
		if order.Symbol == symbol && order.Status == StopOrderStatusActive {
			orderCopy := *order
			result = append(result, &orderCopy)
		}
	}

	return result, nil
}
