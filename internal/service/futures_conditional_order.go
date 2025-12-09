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

// FuturesTriggerType represents futures-specific trigger types
type FuturesTriggerType int

const (
	FuturesTriggerTypeMarkPrice FuturesTriggerType = iota
	FuturesTriggerTypeLastPrice
	FuturesTriggerTypeUnrealizedPnL
	FuturesTriggerTypeFundingRate
	FuturesTriggerTypeMarginRatio
)

// FuturesTriggerCondition represents a futures-specific trigger condition
type FuturesTriggerCondition struct {
	Type            FuturesTriggerType
	Operator        ComparisonOperator
	Value           float64
	PriceType       api.PriceType
	BasePrice       float64
	TimeWindow      time.Duration
	CompositeType   LogicOperator
	SubConditions   []*FuturesTriggerCondition
}

// ComparisonOperator and LogicOperator are already defined in trigger.go

// FuturesConditionalOrderRequest represents a request to create a futures conditional order
type FuturesConditionalOrderRequest struct {
	Symbol           string
	Side             api.OrderSide
	PositionSide     api.PositionSide
	Type             api.OrderType
	Quantity         float64
	Price            float64
	TriggerCondition *FuturesTriggerCondition
	ReduceOnly       bool
	TimeWindow       *repository.TimeWindow
}

// FuturesConditionalOrder represents a futures conditional order
type FuturesConditionalOrder struct {
	OrderID          string
	Symbol           string
	Side             api.OrderSide
	PositionSide     api.PositionSide
	Type             api.OrderType
	Quantity         float64
	Price            float64
	TriggerCondition *FuturesTriggerCondition
	Status           repository.ConditionalOrderStatus
	CreatedAt        int64
	TriggeredAt      int64
	ExecutedOrderID  int64
	ReduceOnly       bool
	TimeWindow       *repository.TimeWindow
}

// FuturesConditionalOrderUpdate represents updates to a futures conditional order
type FuturesConditionalOrderUpdate struct {
	TriggerCondition *FuturesTriggerCondition
	Quantity         *float64
	Price            *float64
	TimeWindow       *repository.TimeWindow
}

// FuturesConditionalOrderService defines the interface for futures conditional order operations
type FuturesConditionalOrderService interface {
	// Create conditional order
	CreateConditionalOrder(request *FuturesConditionalOrderRequest) (*FuturesConditionalOrder, error)

	// Manage conditional orders
	CancelConditionalOrder(orderID string) error
	UpdateConditionalOrder(orderID string, updates *FuturesConditionalOrderUpdate) error
	GetConditionalOrder(orderID string) (*FuturesConditionalOrder, error)
	GetActiveConditionalOrders() ([]*FuturesConditionalOrder, error)
	GetConditionalOrderHistory(startTime, endTime int64) ([]*FuturesConditionalOrder, error)

	// Monitoring and triggering
	StartMonitoring() error
	StopMonitoring() error
}

// futuresConditionalOrderService implements FuturesConditionalOrderService interface
type futuresConditionalOrderService struct {
	client            api.FuturesClient
	marketDataService FuturesMarketDataService
	positionManager   FuturesPositionManager
	tradingService    FuturesTradingService
	logger            logger.Logger
	orders            map[string]*FuturesConditionalOrder
	monitoring        bool
	stopChan          chan struct{}
}

// NewFuturesConditionalOrderService creates a new futures conditional order service
func NewFuturesConditionalOrderService(
	client api.FuturesClient,
	marketDataService FuturesMarketDataService,
	positionManager FuturesPositionManager,
	tradingService FuturesTradingService,
	logger logger.Logger,
) FuturesConditionalOrderService {
	return &futuresConditionalOrderService{
		client:            client,
		marketDataService: marketDataService,
		positionManager:   positionManager,
		tradingService:    tradingService,
		logger:            logger,
		orders:            make(map[string]*FuturesConditionalOrder),
		monitoring:        false,
	}
}

// CreateConditionalOrder creates a new futures conditional order
func (s *futuresConditionalOrderService) CreateConditionalOrder(request *FuturesConditionalOrderRequest) (*FuturesConditionalOrder, error) {
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
	order := &FuturesConditionalOrder{
		OrderID:          orderID,
		Symbol:           request.Symbol,
		Side:             request.Side,
		PositionSide:     request.PositionSide,
		Type:             request.Type,
		Quantity:         request.Quantity,
		Price:            request.Price,
		TriggerCondition: request.TriggerCondition,
		Status:           repository.ConditionalOrderStatusPending,
		CreatedAt:        time.Now().Unix(),
		ReduceOnly:       request.ReduceOnly,
		TimeWindow:       request.TimeWindow,
	}

	// Save order
	s.orders[orderID] = order

	s.logger.Info("Futures conditional order created", map[string]interface{}{
		"order_id":      orderID,
		"symbol":        request.Symbol,
		"side":          string(request.Side),
		"position_side": string(request.PositionSide),
		"type":          string(request.Type),
		"quantity":      request.Quantity,
		"trigger_type":  request.TriggerCondition.Type,
	})

	return order, nil
}

// CancelConditionalOrder cancels a futures conditional order
func (s *futuresConditionalOrderService) CancelConditionalOrder(orderID string) error {
	if orderID == "" {
		return errors.NewTradingError(errors.ErrInvalidParameter, "order ID cannot be empty", 0, nil)
	}

	// Get order
	order, exists := s.orders[orderID]
	if !exists {
		return errors.NewTradingError(
			errors.ErrConditionalOrderNotFound,
			"conditional order not found",
			0,
			nil,
		)
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
	order.Status = repository.ConditionalOrderStatusCancelled

	s.logger.Info("Futures conditional order cancelled", map[string]interface{}{
		"order_id": orderID,
		"symbol":   order.Symbol,
	})

	return nil
}

// UpdateConditionalOrder updates a futures conditional order
func (s *futuresConditionalOrderService) UpdateConditionalOrder(orderID string, updates *FuturesConditionalOrderUpdate) error {
	if orderID == "" {
		return errors.NewTradingError(errors.ErrInvalidParameter, "order ID cannot be empty", 0, nil)
	}

	if updates == nil {
		return errors.NewTradingError(errors.ErrInvalidParameter, "updates cannot be nil", 0, nil)
	}

	// Get existing order
	order, exists := s.orders[orderID]
	if !exists {
		return errors.NewTradingError(
			errors.ErrConditionalOrderNotFound,
			"conditional order not found",
			0,
			nil,
		)
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

	s.logger.Info("Futures conditional order updated", map[string]interface{}{
		"order_id": orderID,
		"symbol":   order.Symbol,
	})

	return nil
}

// GetConditionalOrder retrieves a futures conditional order by ID
func (s *futuresConditionalOrderService) GetConditionalOrder(orderID string) (*FuturesConditionalOrder, error) {
	if orderID == "" {
		return nil, errors.NewTradingError(errors.ErrInvalidParameter, "order ID cannot be empty", 0, nil)
	}

	order, exists := s.orders[orderID]
	if !exists {
		return nil, errors.NewTradingError(
			errors.ErrConditionalOrderNotFound,
			"conditional order not found",
			0,
			nil,
		)
	}

	return order, nil
}

// GetActiveConditionalOrders retrieves all active (pending) futures conditional orders
func (s *futuresConditionalOrderService) GetActiveConditionalOrders() ([]*FuturesConditionalOrder, error) {
	var activeOrders []*FuturesConditionalOrder
	for _, order := range s.orders {
		if order.Status == repository.ConditionalOrderStatusPending {
			activeOrders = append(activeOrders, order)
		}
	}
	return activeOrders, nil
}

// GetConditionalOrderHistory retrieves futures conditional order history within a time range
func (s *futuresConditionalOrderService) GetConditionalOrderHistory(startTime, endTime int64) ([]*FuturesConditionalOrder, error) {
	if startTime < 0 || endTime < 0 {
		return nil, errors.NewTradingError(errors.ErrInvalidParameter, "time values cannot be negative", 0, nil)
	}

	if startTime > endTime {
		return nil, errors.NewTradingError(errors.ErrInvalidParameter, "start time cannot be after end time", 0, nil)
	}

	var historyOrders []*FuturesConditionalOrder
	for _, order := range s.orders {
		if order.CreatedAt >= startTime && order.CreatedAt <= endTime {
			if order.Status == repository.ConditionalOrderStatusExecuted ||
				order.Status == repository.ConditionalOrderStatusCancelled {
				historyOrders = append(historyOrders, order)
			}
		}
	}

	return historyOrders, nil
}

// StartMonitoring starts the monitoring engine
func (s *futuresConditionalOrderService) StartMonitoring() error {
	if s.monitoring {
		return errors.NewTradingError(errors.ErrInvalidParameter, "monitoring already started", 0, nil)
	}

	s.monitoring = true
	s.stopChan = make(chan struct{})

	// Start monitoring goroutine
	go s.monitorOrders()

	s.logger.Info("Futures conditional order monitoring started", nil)
	return nil
}

// StopMonitoring stops the monitoring engine
func (s *futuresConditionalOrderService) StopMonitoring() error {
	if !s.monitoring {
		return errors.NewTradingError(errors.ErrInvalidParameter, "monitoring not started", 0, nil)
	}

	s.monitoring = false
	close(s.stopChan)

	s.logger.Info("Futures conditional order monitoring stopped", nil)
	return nil
}

// monitorOrders monitors conditional orders and triggers them when conditions are met
func (s *futuresConditionalOrderService) monitorOrders() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.checkOrders()
		}
	}
}

// checkOrders checks all pending orders for trigger conditions
func (s *futuresConditionalOrderService) checkOrders() {
	for _, order := range s.orders {
		if order.Status != repository.ConditionalOrderStatusPending {
			continue
		}

		// Check if order is within time window
		if order.TimeWindow != nil {
			now := time.Now()
			if !order.TimeWindow.StartTime.IsZero() && now.Before(order.TimeWindow.StartTime) {
				continue
			}
			if !order.TimeWindow.EndTime.IsZero() && now.After(order.TimeWindow.EndTime) {
				continue
			}
		}

		// Evaluate trigger condition
		triggered, triggerValue, err := s.evaluateTriggerCondition(order)
		if err != nil {
			s.logger.Error("Failed to evaluate trigger condition", map[string]interface{}{
				"order_id": order.OrderID,
				"error":    err.Error(),
			})
			continue
		}

		if triggered {
			s.executeTrigger(order, triggerValue)
		}
	}
}

// evaluateTriggerCondition evaluates a futures trigger condition
func (s *futuresConditionalOrderService) evaluateTriggerCondition(order *FuturesConditionalOrder) (bool, float64, error) {
	condition := order.TriggerCondition

	// Handle composite conditions
	if len(condition.SubConditions) > 0 {
		return s.evaluateCompositeCondition(order, condition)
	}

	// Evaluate simple condition based on type
	switch condition.Type {
	case FuturesTriggerTypeMarkPrice:
		return s.evaluateMarkPriceTrigger(order.Symbol, condition)
	case FuturesTriggerTypeLastPrice:
		return s.evaluateLastPriceTrigger(order.Symbol, condition)
	case FuturesTriggerTypeUnrealizedPnL:
		return s.evaluateUnrealizedPnLTrigger(order.Symbol, order.PositionSide, condition)
	case FuturesTriggerTypeFundingRate:
		return s.evaluateFundingRateTrigger(order.Symbol, condition)
	default:
		return false, 0, fmt.Errorf("unknown trigger type: %d", condition.Type)
	}
}

// evaluateMarkPriceTrigger evaluates mark price trigger
func (s *futuresConditionalOrderService) evaluateMarkPriceTrigger(symbol string, condition *FuturesTriggerCondition) (bool, float64, error) {
	markPrice, err := s.marketDataService.GetMarkPrice(symbol)
	if err != nil {
		return false, 0, err
	}

	triggered := s.compareValue(markPrice, condition.Operator, condition.Value)
	return triggered, markPrice, nil
}

// evaluateLastPriceTrigger evaluates last price trigger
func (s *futuresConditionalOrderService) evaluateLastPriceTrigger(symbol string, condition *FuturesTriggerCondition) (bool, float64, error) {
	lastPrice, err := s.marketDataService.GetLastPrice(symbol)
	if err != nil {
		return false, 0, err
	}

	triggered := s.compareValue(lastPrice, condition.Operator, condition.Value)
	return triggered, lastPrice, nil
}

// evaluateUnrealizedPnLTrigger evaluates unrealized PnL trigger
func (s *futuresConditionalOrderService) evaluateUnrealizedPnLTrigger(symbol string, positionSide api.PositionSide, condition *FuturesTriggerCondition) (bool, float64, error) {
	// Get position
	position, err := s.positionManager.GetPosition(symbol, positionSide)
	if err != nil {
		return false, 0, err
	}

	// Get mark price
	markPrice, err := s.marketDataService.GetMarkPrice(symbol)
	if err != nil {
		return false, 0, err
	}

	// Calculate unrealized PnL
	pnl, err := s.positionManager.CalculateUnrealizedPnL(position, markPrice)
	if err != nil {
		return false, 0, err
	}

	triggered := s.compareValue(pnl, condition.Operator, condition.Value)
	return triggered, pnl, nil
}

// evaluateFundingRateTrigger evaluates funding rate trigger
func (s *futuresConditionalOrderService) evaluateFundingRateTrigger(symbol string, condition *FuturesTriggerCondition) (bool, float64, error) {
	fundingRate, err := s.marketDataService.GetFundingRate(symbol)
	if err != nil {
		return false, 0, err
	}

	triggered := s.compareValue(fundingRate.FundingRate, condition.Operator, condition.Value)
	return triggered, fundingRate.FundingRate, nil
}

// evaluateCompositeCondition evaluates a composite condition
func (s *futuresConditionalOrderService) evaluateCompositeCondition(order *FuturesConditionalOrder, condition *FuturesTriggerCondition) (bool, float64, error) {
	results := make([]bool, len(condition.SubConditions))
	var lastValue float64

	for i, subCondition := range condition.SubConditions {
		// Create temporary order with sub-condition
		tempOrder := &FuturesConditionalOrder{
			Symbol:           order.Symbol,
			PositionSide:     order.PositionSide,
			TriggerCondition: subCondition,
		}

		triggered, value, err := s.evaluateTriggerCondition(tempOrder)
		if err != nil {
			return false, 0, err
		}

		results[i] = triggered
		lastValue = value
	}

	// Apply logic operator
	var finalResult bool
	switch condition.CompositeType {
	case LogicOperatorAND:
		finalResult = true
		for _, result := range results {
			if !result {
				finalResult = false
				break
			}
		}
	case LogicOperatorOR:
		finalResult = false
		for _, result := range results {
			if result {
				finalResult = true
				break
			}
		}
	default:
		return false, 0, fmt.Errorf("unknown logic operator: %d", condition.CompositeType)
	}

	return finalResult, lastValue, nil
}

// compareValue compares a value using the specified operator
func (s *futuresConditionalOrderService) compareValue(currentValue float64, operator ComparisonOperator, targetValue float64) bool {
	switch operator {
	case OperatorGreaterThan:
		return currentValue > targetValue
	case OperatorGreaterEqual:
		return currentValue >= targetValue
	case OperatorLessThan:
		return currentValue < targetValue
	case OperatorLessEqual:
		return currentValue <= targetValue
	case OperatorEqual:
		return currentValue == targetValue
	case OperatorNotEqual:
		return currentValue != targetValue
	default:
		return false
	}
}

// executeTrigger executes a triggered order
func (s *futuresConditionalOrderService) executeTrigger(order *FuturesConditionalOrder, triggerValue float64) {
	// Update order status
	order.Status = repository.ConditionalOrderStatusTriggered
	order.TriggeredAt = time.Now().Unix()

	s.logger.Info("Futures conditional order triggered", map[string]interface{}{
		"order_id":      order.OrderID,
		"symbol":        order.Symbol,
		"trigger_value": triggerValue,
		"trigger_time":  order.TriggeredAt,
	})

	// Execute the order
	var executedOrder *api.FuturesOrder
	var err error

	if order.Type == api.OrderTypeMarket {
		if order.Side == api.OrderSideBuy {
			executedOrder, err = s.tradingService.OpenLongPosition(order.Symbol, order.Quantity, api.OrderTypeMarket, 0)
		} else {
			executedOrder, err = s.tradingService.OpenShortPosition(order.Symbol, order.Quantity, api.OrderTypeMarket, 0)
		}
	} else {
		if order.Side == api.OrderSideBuy {
			executedOrder, err = s.tradingService.OpenLongPosition(order.Symbol, order.Quantity, api.OrderTypeLimit, order.Price)
		} else {
			executedOrder, err = s.tradingService.OpenShortPosition(order.Symbol, order.Quantity, api.OrderTypeLimit, order.Price)
		}
	}

	if err != nil {
		s.logger.Error("Failed to execute triggered order", map[string]interface{}{
			"order_id": order.OrderID,
			"error":    err.Error(),
		})
		return
	}

	// Update order status to executed
	order.Status = repository.ConditionalOrderStatusExecuted
	order.ExecutedOrderID = executedOrder.OrderID

	s.logger.Info("Futures conditional order executed", map[string]interface{}{
		"order_id":          order.OrderID,
		"executed_order_id": executedOrder.OrderID,
		"symbol":            order.Symbol,
	})
}

// validateTriggerCondition validates a futures trigger condition
func (s *futuresConditionalOrderService) validateTriggerCondition(condition *FuturesTriggerCondition) error {
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
	if condition.Value == 0 && condition.Type != FuturesTriggerTypeFundingRate {
		return errors.NewTradingError(
			errors.ErrInvalidTriggerCondition,
			"trigger value cannot be zero",
			0,
			nil,
		)
	}

	return nil
}
