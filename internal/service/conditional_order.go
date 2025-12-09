package service

import (
	"binance-trader/internal/api"
	"binance-trader/internal/repository"
	"binance-trader/pkg/errors"
	"binance-trader/pkg/logger"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ConditionalOrderService defines the interface for conditional order operations
type ConditionalOrderService interface {
	// Create conditional order
	CreateConditionalOrder(request *repository.ConditionalOrderRequest) (*repository.ConditionalOrder, error)

	// Manage conditional orders
	CancelConditionalOrder(orderID string) error
	UpdateConditionalOrder(orderID string, updates *ConditionalOrderUpdate) error
	GetConditionalOrder(orderID string) (*repository.ConditionalOrder, error)
	GetActiveConditionalOrders() ([]*repository.ConditionalOrder, error)
	GetConditionalOrderHistory(startTime, endTime int64) ([]*repository.ConditionalOrder, error)

	// Monitoring and triggering
	StartMonitoring() error
	StopMonitoring() error
}

// ConditionalOrderUpdate represents updates to a conditional order
type ConditionalOrderUpdate struct {
	TriggerCondition *repository.TriggerCondition
	Quantity         *float64
	Price            *float64
	TimeWindow       *repository.TimeWindow
}

// conditionalOrderService implements ConditionalOrderService interface
type conditionalOrderService struct {
	repo              repository.ConditionalOrderRepository
	triggerEngine     TriggerEngine
	tradingService    TradingService
	marketDataService MarketDataService
	logger            logger.Logger
	monitoringEngine  *MonitoringEngine
}

// NewConditionalOrderService creates a new conditional order service
func NewConditionalOrderService(
	repo repository.ConditionalOrderRepository,
	stopOrderRepo repository.StopOrderRepository,
	triggerEngine TriggerEngine,
	tradingService TradingService,
	marketDataService MarketDataService,
	stopLossService StopLossService,
	log logger.Logger,
) ConditionalOrderService {
	// Create monitoring engine
	monitoringEngine := NewMonitoringEngine(
		repo,
		stopOrderRepo,
		triggerEngine,
		tradingService,
		marketDataService,
		stopLossService,
		log,
		nil, // Use default config
	)
	
	return &conditionalOrderService{
		repo:              repo,
		triggerEngine:     triggerEngine,
		tradingService:    tradingService,
		marketDataService: marketDataService,
		logger:            log,
		monitoringEngine:  monitoringEngine,
	}
}

// CreateConditionalOrder creates a new conditional order
func (s *conditionalOrderService) CreateConditionalOrder(request *repository.ConditionalOrderRequest) (*repository.ConditionalOrder, error) {
	// Validate request
	if request == nil {
		return nil, errors.NewTradingError(errors.ErrInvalidParameter, "request cannot be nil", 0, nil)
	}

	if request.Symbol == "" {
		return nil, errors.NewTradingError(errors.ErrInvalidParameter, "symbol cannot be empty", 0, nil)
	}

	if request.Quantity <= 0 {
		return nil, errors.NewTradingError(errors.ErrInvalidParameter, "quantity must be greater than 0", 0, nil)
	}

	if request.Type == api.OrderTypeLimit && request.Price <= 0 {
		return nil, errors.NewTradingError(errors.ErrInvalidParameter, "price must be greater than 0 for limit orders", 0, nil)
	}

	if request.TriggerCondition == nil {
		return nil, errors.NewTradingError(errors.ErrInvalidTriggerCondition, "trigger condition cannot be nil", 0, nil)
	}

	// Validate trigger condition
	if err := s.validateTriggerCondition(request.TriggerCondition); err != nil {
		return nil, err
	}

	// Generate unique order ID
	orderID := uuid.New().String()

	// Create conditional order
	order := &repository.ConditionalOrder{
		OrderID:          orderID,
		Symbol:           request.Symbol,
		Side:             request.Side,
		Type:             request.Type,
		Quantity:         request.Quantity,
		Price:            request.Price,
		TriggerCondition: request.TriggerCondition,
		Status:           repository.ConditionalOrderStatusPending,
		CreatedAt:        time.Now().Unix(),
		TimeWindow:       request.TimeWindow,
	}

	// Save to repository
	if err := s.repo.Save(order); err != nil {
		s.logger.LogError(err, map[string]interface{}{
			"operation": "create_conditional_order",
			"symbol":    request.Symbol,
		})
		return nil, err
	}

	// Register condition with trigger engine
	triggerCond := s.convertToServiceTriggerCondition(request.TriggerCondition)
	if err := s.triggerEngine.RegisterCondition(orderID, triggerCond); err != nil {
		s.logger.Warn("Failed to register condition with trigger engine", map[string]interface{}{
			"order_id": orderID,
			"error":    err.Error(),
		})
	}

	s.logger.Info("Conditional order created", map[string]interface{}{
		"order_id": orderID,
		"symbol":   request.Symbol,
		"side":     string(request.Side),
		"type":     string(request.Type),
		"quantity": request.Quantity,
	})

	return order, nil
}

// CancelConditionalOrder cancels a conditional order
func (s *conditionalOrderService) CancelConditionalOrder(orderID string) error {
	if orderID == "" {
		return errors.NewTradingError(errors.ErrInvalidParameter, "order ID cannot be empty", 0, nil)
	}

	// Get order from repository
	order, err := s.repo.FindByID(orderID)
	if err != nil {
		return err
	}

	// Check if order can be cancelled
	if order.Status != repository.ConditionalOrderStatusPending {
		return errors.NewTradingError(
			errors.ErrInvalidParameter,
			fmt.Sprintf("cannot cancel order with status %s", order.Status),
			0,
			nil,
		)
	}

	// Update status to cancelled
	if err := s.repo.UpdateStatus(orderID, repository.ConditionalOrderStatusCancelled, 0, 0); err != nil {
		s.logger.LogError(err, map[string]interface{}{
			"operation": "cancel_conditional_order",
			"order_id":  orderID,
		})
		return err
	}

	// Unregister from trigger engine
	if err := s.triggerEngine.UnregisterCondition(orderID); err != nil {
		s.logger.Warn("Failed to unregister condition from trigger engine", map[string]interface{}{
			"order_id": orderID,
			"error":    err.Error(),
		})
	}

	s.logger.Info("Conditional order cancelled", map[string]interface{}{
		"order_id": orderID,
		"symbol":   order.Symbol,
	})

	return nil
}

// UpdateConditionalOrder updates a conditional order
func (s *conditionalOrderService) UpdateConditionalOrder(orderID string, updates *ConditionalOrderUpdate) error {
	if orderID == "" {
		return errors.NewTradingError(errors.ErrInvalidParameter, "order ID cannot be empty", 0, nil)
	}

	if updates == nil {
		return errors.NewTradingError(errors.ErrInvalidParameter, "updates cannot be nil", 0, nil)
	}

	// Get existing order
	order, err := s.repo.FindByID(orderID)
	if err != nil {
		return err
	}

	// Check if order can be updated
	if order.Status != repository.ConditionalOrderStatusPending {
		return errors.NewTradingError(
			errors.ErrInvalidParameter,
			fmt.Sprintf("cannot update order with status %s", order.Status),
			0,
			nil,
		)
	}

	// Apply updates
	if updates.TriggerCondition != nil {
		if err := s.validateTriggerCondition(updates.TriggerCondition); err != nil {
			return err
		}
		order.TriggerCondition = updates.TriggerCondition

		// Update trigger engine registration
		triggerCond := s.convertToServiceTriggerCondition(order.TriggerCondition)
		if err := s.triggerEngine.RegisterCondition(orderID, triggerCond); err != nil {
			s.logger.Warn("Failed to update condition in trigger engine", map[string]interface{}{
				"order_id": orderID,
				"error":    err.Error(),
			})
		}
	}

	if updates.Quantity != nil {
		if *updates.Quantity <= 0 {
			return errors.NewTradingError(errors.ErrInvalidParameter, "quantity must be greater than 0", 0, nil)
		}
		order.Quantity = *updates.Quantity
	}

	if updates.Price != nil {
		if order.Type == api.OrderTypeLimit && *updates.Price <= 0 {
			return errors.NewTradingError(errors.ErrInvalidParameter, "price must be greater than 0 for limit orders", 0, nil)
		}
		order.Price = *updates.Price
	}

	if updates.TimeWindow != nil {
		order.TimeWindow = updates.TimeWindow
	}

	// Save updated order
	if err := s.repo.Update(order); err != nil {
		s.logger.LogError(err, map[string]interface{}{
			"operation": "update_conditional_order",
			"order_id":  orderID,
		})
		return err
	}

	s.logger.Info("Conditional order updated", map[string]interface{}{
		"order_id": orderID,
		"symbol":   order.Symbol,
	})

	return nil
}

// GetConditionalOrder retrieves a conditional order by ID
func (s *conditionalOrderService) GetConditionalOrder(orderID string) (*repository.ConditionalOrder, error) {
	if orderID == "" {
		return nil, errors.NewTradingError(errors.ErrInvalidParameter, "order ID cannot be empty", 0, nil)
	}

	return s.repo.FindByID(orderID)
}

// GetActiveConditionalOrders retrieves all active (pending) conditional orders
func (s *conditionalOrderService) GetActiveConditionalOrders() ([]*repository.ConditionalOrder, error) {
	return s.repo.FindActiveOrders()
}

// GetConditionalOrderHistory retrieves conditional order history within a time range
func (s *conditionalOrderService) GetConditionalOrderHistory(startTime, endTime int64) ([]*repository.ConditionalOrder, error) {
	if startTime < 0 || endTime < 0 {
		return nil, errors.NewTradingError(errors.ErrInvalidParameter, "time values cannot be negative", 0, nil)
	}

	if startTime > endTime {
		return nil, errors.NewTradingError(errors.ErrInvalidParameter, "start time cannot be after end time", 0, nil)
	}

	// Get all orders in time range
	allOrders, err := s.repo.FindOrdersByTimeRange(startTime, endTime)
	if err != nil {
		return nil, err
	}

	// Filter for executed or cancelled orders
	var historyOrders []*repository.ConditionalOrder
	for _, order := range allOrders {
		if order.Status == repository.ConditionalOrderStatusExecuted ||
			order.Status == repository.ConditionalOrderStatusCancelled {
			historyOrders = append(historyOrders, order)
		}
	}

	return historyOrders, nil
}

// StartMonitoring starts the monitoring engine
func (s *conditionalOrderService) StartMonitoring() error {
	return s.monitoringEngine.Start()
}

// StopMonitoring stops the monitoring engine
func (s *conditionalOrderService) StopMonitoring() error {
	return s.monitoringEngine.Stop()
}

// validateTriggerCondition validates a trigger condition
func (s *conditionalOrderService) validateTriggerCondition(condition *repository.TriggerCondition) error {
	if condition == nil {
		return errors.NewTradingError(errors.ErrInvalidTriggerCondition, "trigger condition cannot be nil", 0, nil)
	}

	// Validate composite conditions
	if len(condition.SubConditions) > 0 {
		for _, subCond := range condition.SubConditions {
			if err := s.validateTriggerCondition(subCond); err != nil {
				return err
			}
		}
		return nil
	}

	// Validate simple conditions
	if condition.Type == repository.TriggerTypePriceChangePercent && condition.BasePrice <= 0 {
		return errors.NewTradingError(
			errors.ErrInvalidTriggerCondition,
			"base price must be greater than 0 for price change percentage conditions",
			0,
			nil,
		)
	}

	return nil
}

// convertToServiceTriggerCondition converts repository trigger condition to service trigger condition
func (s *conditionalOrderService) convertToServiceTriggerCondition(repoCond *repository.TriggerCondition) *TriggerCondition {
	if repoCond == nil {
		return nil
	}

	serviceCond := &TriggerCondition{
		Type:          TriggerType(repoCond.Type),
		Operator:      ComparisonOperator(repoCond.Operator),
		Value:         repoCond.Value,
		BasePrice:     repoCond.BasePrice,
		TimeWindow:    repoCond.TimeWindow,
		CompositeType: LogicOperator(repoCond.CompositeType),
	}

	// Convert sub-conditions recursively
	if len(repoCond.SubConditions) > 0 {
		serviceCond.SubConditions = make([]*TriggerCondition, len(repoCond.SubConditions))
		for i, subCond := range repoCond.SubConditions {
			serviceCond.SubConditions[i] = s.convertToServiceTriggerCondition(subCond)
		}
	}

	return serviceCond
}
