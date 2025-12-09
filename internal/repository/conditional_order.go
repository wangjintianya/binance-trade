package repository

import (
	"binance-trader/internal/api"
	"binance-trader/pkg/errors"
	"sync"
	"time"
)

// ConditionalOrderStatus represents the status of a conditional order
type ConditionalOrderStatus string

const (
	ConditionalOrderStatusPending   ConditionalOrderStatus = "PENDING"
	ConditionalOrderStatusTriggered ConditionalOrderStatus = "TRIGGERED"
	ConditionalOrderStatusExecuted  ConditionalOrderStatus = "EXECUTED"
	ConditionalOrderStatusCancelled ConditionalOrderStatus = "CANCELLED"
)

// TriggerType represents the type of trigger condition
type TriggerType int

const (
	TriggerTypePrice TriggerType = iota
	TriggerTypePriceChangePercent
	TriggerTypeVolume
)

// ComparisonOperator represents comparison operators for trigger conditions
type ComparisonOperator int

const (
	OperatorGreaterThan ComparisonOperator = iota
	OperatorLessThan
	OperatorGreaterEqual
	OperatorLessEqual
)

// LogicOperator represents logical operators for composite conditions
type LogicOperator int

const (
	LogicAND LogicOperator = iota
	LogicOR
)

// TriggerCondition represents a trigger condition for conditional orders
type TriggerCondition struct {
	Type          TriggerType
	Operator      ComparisonOperator
	Value         float64
	BasePrice     float64       // For price change percentage calculations
	TimeWindow    time.Duration // For volume calculations
	CompositeType LogicOperator // For composite conditions
	SubConditions []*TriggerCondition
}

// TimeWindow represents a time range for filtering
type TimeWindow struct {
	StartTime time.Time
	EndTime   time.Time
}

// ConditionalOrderRequest represents a request to create a conditional order
type ConditionalOrderRequest struct {
	Symbol           string
	Side             api.OrderSide
	Type             api.OrderType
	Quantity         float64
	Price            float64
	TriggerCondition *TriggerCondition
	TimeWindow       *TimeWindow
}

// ConditionalOrder represents a conditional order
type ConditionalOrder struct {
	OrderID          string
	Symbol           string
	Side             api.OrderSide
	Type             api.OrderType
	Quantity         float64
	Price            float64
	TriggerCondition *TriggerCondition
	Status           ConditionalOrderStatus
	CreatedAt        int64
	TriggeredAt      int64
	ExecutedOrderID  int64
	TimeWindow       *TimeWindow
}

// ConditionalOrderRepository defines the interface for conditional order data persistence
type ConditionalOrderRepository interface {
	// CRUD operations
	Save(order *ConditionalOrder) error
	FindByID(orderID string) (*ConditionalOrder, error)
	FindBySymbol(symbol string) ([]*ConditionalOrder, error)
	Update(order *ConditionalOrder) error
	Delete(orderID string) error

	// Query operations
	FindActiveOrders() ([]*ConditionalOrder, error)
	FindOrdersByStatus(status ConditionalOrderStatus) ([]*ConditionalOrder, error)
	FindOrdersByTimeRange(startTime, endTime int64) ([]*ConditionalOrder, error)

	// Status management
	UpdateStatus(orderID string, newStatus ConditionalOrderStatus, triggeredAt int64, executedOrderID int64) error
}

// memoryConditionalOrderRepository implements ConditionalOrderRepository using in-memory storage
type memoryConditionalOrderRepository struct {
	mu     sync.RWMutex
	orders map[string]*ConditionalOrder
}

// NewMemoryConditionalOrderRepository creates a new in-memory conditional order repository
func NewMemoryConditionalOrderRepository() ConditionalOrderRepository {
	return &memoryConditionalOrderRepository{
		orders: make(map[string]*ConditionalOrder),
	}
}

// Save stores a new conditional order
func (r *memoryConditionalOrderRepository) Save(order *ConditionalOrder) error {
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
	if order.TriggerCondition != nil {
		conditionCopy := *order.TriggerCondition
		orderCopy.TriggerCondition = &conditionCopy
	}
	if order.TimeWindow != nil {
		timeWindowCopy := *order.TimeWindow
		orderCopy.TimeWindow = &timeWindowCopy
	}

	r.orders[order.OrderID] = &orderCopy

	return nil
}

// FindByID retrieves a conditional order by its ID
func (r *memoryConditionalOrderRepository) FindByID(orderID string) (*ConditionalOrder, error) {
	if orderID == "" {
		return nil, errors.NewTradingError(errors.ErrInvalidParameter, "order ID cannot be empty", 0, nil)
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	order, exists := r.orders[orderID]
	if !exists {
		return nil, errors.NewTradingError(errors.ErrConditionalOrderNotFound, "conditional order not found", 0, nil)
	}

	// Return a copy to prevent external modifications
	orderCopy := *order
	if order.TriggerCondition != nil {
		conditionCopy := *order.TriggerCondition
		orderCopy.TriggerCondition = &conditionCopy
	}
	if order.TimeWindow != nil {
		timeWindowCopy := *order.TimeWindow
		orderCopy.TimeWindow = &timeWindowCopy
	}

	return &orderCopy, nil
}

// FindBySymbol retrieves all conditional orders for a specific symbol
func (r *memoryConditionalOrderRepository) FindBySymbol(symbol string) ([]*ConditionalOrder, error) {
	if symbol == "" {
		return nil, errors.NewTradingError(errors.ErrInvalidParameter, "symbol cannot be empty", 0, nil)
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*ConditionalOrder
	for _, order := range r.orders {
		if order.Symbol == symbol {
			orderCopy := *order
			if order.TriggerCondition != nil {
				conditionCopy := *order.TriggerCondition
				orderCopy.TriggerCondition = &conditionCopy
			}
			if order.TimeWindow != nil {
				timeWindowCopy := *order.TimeWindow
				orderCopy.TimeWindow = &timeWindowCopy
			}
			result = append(result, &orderCopy)
		}
	}

	return result, nil
}

// Update updates an existing conditional order
func (r *memoryConditionalOrderRepository) Update(order *ConditionalOrder) error {
	if order == nil {
		return errors.NewTradingError(errors.ErrInvalidParameter, "order cannot be nil", 0, nil)
	}

	if order.OrderID == "" {
		return errors.NewTradingError(errors.ErrInvalidParameter, "order ID cannot be empty", 0, nil)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.orders[order.OrderID]; !exists {
		return errors.NewTradingError(errors.ErrConditionalOrderNotFound, "conditional order not found", 0, nil)
	}

	// Create a copy to avoid external modifications
	orderCopy := *order
	if order.TriggerCondition != nil {
		conditionCopy := *order.TriggerCondition
		orderCopy.TriggerCondition = &conditionCopy
	}
	if order.TimeWindow != nil {
		timeWindowCopy := *order.TimeWindow
		orderCopy.TimeWindow = &timeWindowCopy
	}

	r.orders[order.OrderID] = &orderCopy

	return nil
}

// Delete removes a conditional order by its ID
func (r *memoryConditionalOrderRepository) Delete(orderID string) error {
	if orderID == "" {
		return errors.NewTradingError(errors.ErrInvalidParameter, "order ID cannot be empty", 0, nil)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.orders[orderID]; !exists {
		return errors.NewTradingError(errors.ErrConditionalOrderNotFound, "conditional order not found", 0, nil)
	}

	delete(r.orders, orderID)
	return nil
}

// FindActiveOrders retrieves all orders with status PENDING
func (r *memoryConditionalOrderRepository) FindActiveOrders() ([]*ConditionalOrder, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*ConditionalOrder
	for _, order := range r.orders {
		if order.Status == ConditionalOrderStatusPending {
			orderCopy := *order
			if order.TriggerCondition != nil {
				conditionCopy := *order.TriggerCondition
				orderCopy.TriggerCondition = &conditionCopy
			}
			if order.TimeWindow != nil {
				timeWindowCopy := *order.TimeWindow
				orderCopy.TimeWindow = &timeWindowCopy
			}
			result = append(result, &orderCopy)
		}
	}

	return result, nil
}

// FindOrdersByStatus retrieves all orders with a specific status
func (r *memoryConditionalOrderRepository) FindOrdersByStatus(status ConditionalOrderStatus) ([]*ConditionalOrder, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*ConditionalOrder
	for _, order := range r.orders {
		if order.Status == status {
			orderCopy := *order
			if order.TriggerCondition != nil {
				conditionCopy := *order.TriggerCondition
				orderCopy.TriggerCondition = &conditionCopy
			}
			if order.TimeWindow != nil {
				timeWindowCopy := *order.TimeWindow
				orderCopy.TimeWindow = &timeWindowCopy
			}
			result = append(result, &orderCopy)
		}
	}

	return result, nil
}

// FindOrdersByTimeRange retrieves conditional orders within a time range
func (r *memoryConditionalOrderRepository) FindOrdersByTimeRange(startTime, endTime int64) ([]*ConditionalOrder, error) {
	if startTime < 0 || endTime < 0 {
		return nil, errors.NewTradingError(errors.ErrInvalidParameter, "time values cannot be negative", 0, nil)
	}

	if startTime > endTime {
		return nil, errors.NewTradingError(errors.ErrInvalidParameter, "start time cannot be after end time", 0, nil)
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*ConditionalOrder
	for _, order := range r.orders {
		if order.CreatedAt >= startTime && order.CreatedAt <= endTime {
			orderCopy := *order
			if order.TriggerCondition != nil {
				conditionCopy := *order.TriggerCondition
				orderCopy.TriggerCondition = &conditionCopy
			}
			if order.TimeWindow != nil {
				timeWindowCopy := *order.TimeWindow
				orderCopy.TimeWindow = &timeWindowCopy
			}
			result = append(result, &orderCopy)
		}
	}

	return result, nil
}

// UpdateStatus updates the status of a conditional order
func (r *memoryConditionalOrderRepository) UpdateStatus(orderID string, newStatus ConditionalOrderStatus, triggeredAt int64, executedOrderID int64) error {
	if orderID == "" {
		return errors.NewTradingError(errors.ErrInvalidParameter, "order ID cannot be empty", 0, nil)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	order, exists := r.orders[orderID]
	if !exists {
		return errors.NewTradingError(errors.ErrConditionalOrderNotFound, "conditional order not found", 0, nil)
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
