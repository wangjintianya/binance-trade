package service

import (
	"binance-trader/internal/repository"
	"binance-trader/pkg/errors"
	"binance-trader/pkg/logger"
	"fmt"
	"sync"
	"time"
)

// StopLossService defines the interface for stop loss and take profit operations
type StopLossService interface {
	// Set stop loss and take profit
	SetStopLoss(symbol string, position float64, stopPrice float64) (*repository.StopOrder, error)
	SetTakeProfit(symbol string, position float64, targetPrice float64) (*repository.StopOrder, error)
	SetStopLossTakeProfit(symbol string, position float64, stopPrice, targetPrice float64) (*repository.StopOrderPair, error)
	SetTrailingStop(symbol string, position float64, trailPercent float64) (*repository.TrailingStopOrder, error)

	// Manage stop orders
	CancelStopOrder(orderID string) error
	GetActiveStopOrders(symbol string) ([]*repository.StopOrder, error)
	UpdateTrailingStop(orderID string, newTrailPercent float64) error
}

// stopLossService implements the StopLossService interface
type stopLossService struct {
	stopOrderRepo  repository.StopOrderRepository
	triggerEngine  TriggerEngine
	tradingService TradingService
	marketService  MarketDataService
	logger         logger.Logger
}

// UpdateTrailingStopPrice updates the trailing stop price based on current market price
// This should be called periodically by the monitoring engine
func (s *stopLossService) UpdateTrailingStopPrice(orderID string, currentPrice float64) (bool, error) {
	// Find trailing stop order
	trailingOrder, err := s.stopOrderRepo.FindTrailingStopOrderByID(orderID)
	if err != nil {
		return false, err
	}

	// Check if order is still active
	if trailingOrder.Status != repository.StopOrderStatusActive {
		return false, nil
	}

	updated := false

	// If current price is higher than highest price, update highest price and stop price
	if currentPrice > trailingOrder.HighestPrice {
		trailingOrder.HighestPrice = currentPrice
		trailingOrder.CurrentStopPrice = currentPrice * (1 - trailingOrder.TrailPercent/100)
		trailingOrder.LastUpdatedAt = time.Now().Unix()
		updated = true

		// Save updated order
		if err := s.stopOrderRepo.UpdateTrailingStopOrder(trailingOrder); err != nil {
			s.logger.LogError(err, map[string]interface{}{
				"operation": "update_trailing_stop_price",
				"order_id":  orderID,
			})
			return false, err
		}

		s.logger.Debug("Trailing stop price adjusted", map[string]interface{}{
			"order_id":        orderID,
			"current_price":   currentPrice,
			"highest_price":   trailingOrder.HighestPrice,
			"new_stop_price":  trailingOrder.CurrentStopPrice,
			"trail_percent":   trailingOrder.TrailPercent,
		})
	}

	// Check if stop price is triggered
	if currentPrice <= trailingOrder.CurrentStopPrice {
		// Trigger the stop order
		if err := s.triggerTrailingStop(trailingOrder, currentPrice); err != nil {
			s.logger.LogError(err, map[string]interface{}{
				"operation": "trigger_trailing_stop",
				"order_id":  orderID,
			})
			return false, err
		}
		return true, nil
	}

	return updated, nil
}

// triggerTrailingStop executes a trailing stop order when triggered
func (s *stopLossService) triggerTrailingStop(order *repository.TrailingStopOrder, triggerPrice float64) error {
	// Update status to triggered
	order.Status = repository.StopOrderStatusTriggered
	if err := s.stopOrderRepo.UpdateTrailingStopOrder(order); err != nil {
		return err
	}

	// Execute market sell order to close position
	executedOrder, err := s.tradingService.PlaceMarketSellOrder(order.Symbol, order.Position)
	if err != nil {
		s.logger.LogError(err, map[string]interface{}{
			"operation":    "execute_trailing_stop",
			"order_id":     order.OrderID,
			"symbol":       order.Symbol,
			"position":     order.Position,
			"trigger_price": triggerPrice,
		})
		return err
	}

	s.logger.Info("Trailing stop order triggered and executed", map[string]interface{}{
		"order_id":          order.OrderID,
		"symbol":            order.Symbol,
		"position":          order.Position,
		"trigger_price":     triggerPrice,
		"stop_price":        order.CurrentStopPrice,
		"highest_price":     order.HighestPrice,
		"trail_percent":     order.TrailPercent,
		"executed_order_id": executedOrder.OrderID,
	})

	return nil
}

// NewStopLossService creates a new stop loss service instance
func NewStopLossService(
	stopOrderRepo repository.StopOrderRepository,
	triggerEngine TriggerEngine,
	tradingService TradingService,
	marketService MarketDataService,
	log logger.Logger,
) StopLossService {
	return &stopLossService{
		stopOrderRepo:  stopOrderRepo,
		triggerEngine:  triggerEngine,
		tradingService: tradingService,
		marketService:  marketService,
		logger:         log,
	}
}

// SetStopLoss sets a stop loss order for a position
func (s *stopLossService) SetStopLoss(symbol string, position float64, stopPrice float64) (*repository.StopOrder, error) {
	// Validate input parameters
	if symbol == "" {
		return nil, errors.NewTradingError(errors.ErrInvalidParameter, "symbol cannot be empty", 0, nil)
	}

	if position <= 0 {
		return nil, errors.NewTradingError(errors.ErrInvalidParameter, "position must be greater than 0", 0, nil)
	}

	if stopPrice <= 0 {
		return nil, errors.NewTradingError(errors.ErrInvalidParameter, "stop price must be greater than 0", 0, nil)
	}

	// Create stop order
	stopOrder := &repository.StopOrder{
		OrderID:   generateOrderID("SL"),
		Symbol:    symbol,
		Position:  position,
		StopPrice: stopPrice,
		Type:      repository.StopOrderTypeStopLoss,
		Status:    repository.StopOrderStatusActive,
		CreatedAt: time.Now().Unix(),
	}

	// Save to repository
	if err := s.stopOrderRepo.SaveStopOrder(stopOrder); err != nil {
		s.logger.LogError(err, map[string]interface{}{
			"operation":  "set_stop_loss",
			"symbol":     symbol,
			"position":   position,
			"stop_price": stopPrice,
		})
		return nil, err
	}

	// Register trigger condition
	condition := &TriggerCondition{
		Type:     TriggerTypePrice,
		Operator: OperatorLessEqual,
		Value:    stopPrice,
	}

	if err := s.triggerEngine.RegisterCondition(stopOrder.OrderID, condition); err != nil {
		s.logger.Warn("Failed to register trigger condition", map[string]interface{}{
			"order_id": stopOrder.OrderID,
			"error":    err.Error(),
		})
	}

	s.logger.Info("Stop loss order created", map[string]interface{}{
		"order_id":   stopOrder.OrderID,
		"symbol":     symbol,
		"position":   position,
		"stop_price": stopPrice,
	})

	return stopOrder, nil
}

// SetTakeProfit sets a take profit order for a position
func (s *stopLossService) SetTakeProfit(symbol string, position float64, targetPrice float64) (*repository.StopOrder, error) {
	// Validate input parameters
	if symbol == "" {
		return nil, errors.NewTradingError(errors.ErrInvalidParameter, "symbol cannot be empty", 0, nil)
	}

	if position <= 0 {
		return nil, errors.NewTradingError(errors.ErrInvalidParameter, "position must be greater than 0", 0, nil)
	}

	if targetPrice <= 0 {
		return nil, errors.NewTradingError(errors.ErrInvalidParameter, "target price must be greater than 0", 0, nil)
	}

	// Create take profit order
	takeProfitOrder := &repository.StopOrder{
		OrderID:   generateOrderID("TP"),
		Symbol:    symbol,
		Position:  position,
		StopPrice: targetPrice,
		Type:      repository.StopOrderTypeTakeProfit,
		Status:    repository.StopOrderStatusActive,
		CreatedAt: time.Now().Unix(),
	}

	// Save to repository
	if err := s.stopOrderRepo.SaveStopOrder(takeProfitOrder); err != nil {
		s.logger.LogError(err, map[string]interface{}{
			"operation":    "set_take_profit",
			"symbol":       symbol,
			"position":     position,
			"target_price": targetPrice,
		})
		return nil, err
	}

	// Register trigger condition
	condition := &TriggerCondition{
		Type:     TriggerTypePrice,
		Operator: OperatorGreaterEqual,
		Value:    targetPrice,
	}

	if err := s.triggerEngine.RegisterCondition(takeProfitOrder.OrderID, condition); err != nil {
		s.logger.Warn("Failed to register trigger condition", map[string]interface{}{
			"order_id": takeProfitOrder.OrderID,
			"error":    err.Error(),
		})
	}

	s.logger.Info("Take profit order created", map[string]interface{}{
		"order_id":     takeProfitOrder.OrderID,
		"symbol":       symbol,
		"position":     position,
		"target_price": targetPrice,
	})

	return takeProfitOrder, nil
}

// SetStopLossTakeProfit sets both stop loss and take profit orders as a pair
func (s *stopLossService) SetStopLossTakeProfit(symbol string, position float64, stopPrice, targetPrice float64) (*repository.StopOrderPair, error) {
	// Validate input parameters
	if symbol == "" {
		return nil, errors.NewTradingError(errors.ErrInvalidParameter, "symbol cannot be empty", 0, nil)
	}

	if position <= 0 {
		return nil, errors.NewTradingError(errors.ErrInvalidParameter, "position must be greater than 0", 0, nil)
	}

	if stopPrice <= 0 {
		return nil, errors.NewTradingError(errors.ErrInvalidParameter, "stop price must be greater than 0", 0, nil)
	}

	if targetPrice <= 0 {
		return nil, errors.NewTradingError(errors.ErrInvalidParameter, "target price must be greater than 0", 0, nil)
	}

	// Create stop loss order
	stopLossOrder := &repository.StopOrder{
		OrderID:   generateOrderID("SL"),
		Symbol:    symbol,
		Position:  position,
		StopPrice: stopPrice,
		Type:      repository.StopOrderTypeStopLoss,
		Status:    repository.StopOrderStatusActive,
		CreatedAt: time.Now().Unix(),
	}

	// Create take profit order
	takeProfitOrder := &repository.StopOrder{
		OrderID:   generateOrderID("TP"),
		Symbol:    symbol,
		Position:  position,
		StopPrice: targetPrice,
		Type:      repository.StopOrderTypeTakeProfit,
		Status:    repository.StopOrderStatusActive,
		CreatedAt: time.Now().Unix(),
	}

	// Create order pair
	orderPair := &repository.StopOrderPair{
		PairID:          generateOrderID("PAIR"),
		Symbol:          symbol,
		Position:        position,
		StopLossOrder:   stopLossOrder,
		TakeProfitOrder: takeProfitOrder,
		Status:          "ACTIVE",
	}

	// Save individual orders
	if err := s.stopOrderRepo.SaveStopOrder(stopLossOrder); err != nil {
		s.logger.LogError(err, map[string]interface{}{
			"operation":  "set_stop_loss_take_profit",
			"symbol":     symbol,
			"stop_price": stopPrice,
		})
		return nil, err
	}

	if err := s.stopOrderRepo.SaveStopOrder(takeProfitOrder); err != nil {
		s.logger.LogError(err, map[string]interface{}{
			"operation":    "set_stop_loss_take_profit",
			"symbol":       symbol,
			"target_price": targetPrice,
		})
		// Rollback stop loss order
		s.stopOrderRepo.DeleteStopOrder(stopLossOrder.OrderID)
		return nil, err
	}

	// Save order pair
	if err := s.stopOrderRepo.SaveStopOrderPair(orderPair); err != nil {
		s.logger.LogError(err, map[string]interface{}{
			"operation": "set_stop_loss_take_profit",
			"symbol":    symbol,
		})
		// Rollback both orders
		s.stopOrderRepo.DeleteStopOrder(stopLossOrder.OrderID)
		s.stopOrderRepo.DeleteStopOrder(takeProfitOrder.OrderID)
		return nil, err
	}

	// Register trigger conditions
	stopLossCondition := &TriggerCondition{
		Type:     TriggerTypePrice,
		Operator: OperatorLessEqual,
		Value:    stopPrice,
	}

	takeProfitCondition := &TriggerCondition{
		Type:     TriggerTypePrice,
		Operator: OperatorGreaterEqual,
		Value:    targetPrice,
	}

	s.triggerEngine.RegisterCondition(stopLossOrder.OrderID, stopLossCondition)
	s.triggerEngine.RegisterCondition(takeProfitOrder.OrderID, takeProfitCondition)

	s.logger.Info("Stop loss and take profit pair created", map[string]interface{}{
		"pair_id":      orderPair.PairID,
		"symbol":       symbol,
		"position":     position,
		"stop_price":   stopPrice,
		"target_price": targetPrice,
	})

	return orderPair, nil
}

// SetTrailingStop sets a trailing stop order for a position
func (s *stopLossService) SetTrailingStop(symbol string, position float64, trailPercent float64) (*repository.TrailingStopOrder, error) {
	// Validate input parameters
	if symbol == "" {
		return nil, errors.NewTradingError(errors.ErrInvalidParameter, "symbol cannot be empty", 0, nil)
	}

	if position <= 0 {
		return nil, errors.NewTradingError(errors.ErrInvalidParameter, "position must be greater than 0", 0, nil)
	}

	if trailPercent <= 0 {
		return nil, errors.NewTradingError(errors.ErrInvalidParameter, "trail percent must be greater than 0", 0, nil)
	}

	// Get current price to initialize highest price
	currentPrice, err := s.marketService.GetCurrentPrice(symbol)
	if err != nil {
		s.logger.LogError(err, map[string]interface{}{
			"operation": "set_trailing_stop",
			"symbol":    symbol,
		})
		return nil, err
	}

	// Calculate initial stop price
	initialStopPrice := currentPrice * (1 - trailPercent/100)

	// Create trailing stop order
	trailingStopOrder := &repository.TrailingStopOrder{
		OrderID:          generateOrderID("TS"),
		Symbol:           symbol,
		Position:         position,
		TrailPercent:     trailPercent,
		HighestPrice:     currentPrice,
		CurrentStopPrice: initialStopPrice,
		Status:           repository.StopOrderStatusActive,
		CreatedAt:        time.Now().Unix(),
		LastUpdatedAt:    time.Now().Unix(),
	}

	// Save to repository
	if err := s.stopOrderRepo.SaveTrailingStopOrder(trailingStopOrder); err != nil {
		s.logger.LogError(err, map[string]interface{}{
			"operation":     "set_trailing_stop",
			"symbol":        symbol,
			"position":      position,
			"trail_percent": trailPercent,
		})
		return nil, err
	}

	s.logger.Info("Trailing stop order created", map[string]interface{}{
		"order_id":       trailingStopOrder.OrderID,
		"symbol":         symbol,
		"position":       position,
		"trail_percent":  trailPercent,
		"highest_price":  currentPrice,
		"initial_stop":   initialStopPrice,
	})

	return trailingStopOrder, nil
}

// CancelStopOrder cancels a stop order
func (s *stopLossService) CancelStopOrder(orderID string) error {
	// Validate input
	if orderID == "" {
		return errors.NewTradingError(errors.ErrInvalidParameter, "order ID cannot be empty", 0, nil)
	}

	// Try to find as regular stop order
	stopOrder, err := s.stopOrderRepo.FindStopOrderByID(orderID)
	if err == nil {
		// Update status to cancelled
		if err := s.stopOrderRepo.UpdateStopOrderStatus(orderID, repository.StopOrderStatusCancelled, 0, 0); err != nil {
			s.logger.LogError(err, map[string]interface{}{
				"operation": "cancel_stop_order",
				"order_id":  orderID,
			})
			return err
		}

		// Unregister trigger condition
		s.triggerEngine.UnregisterCondition(orderID)

		s.logger.Info("Stop order cancelled", map[string]interface{}{
			"order_id": orderID,
			"symbol":   stopOrder.Symbol,
		})

		return nil
	}

	// Try to find as trailing stop order
	trailingOrder, err := s.stopOrderRepo.FindTrailingStopOrderByID(orderID)
	if err == nil {
		// Update status to cancelled
		trailingOrder.Status = repository.StopOrderStatusCancelled
		if err := s.stopOrderRepo.UpdateTrailingStopOrder(trailingOrder); err != nil {
			s.logger.LogError(err, map[string]interface{}{
				"operation": "cancel_stop_order",
				"order_id":  orderID,
			})
			return err
		}

		s.logger.Info("Trailing stop order cancelled", map[string]interface{}{
			"order_id": orderID,
			"symbol":   trailingOrder.Symbol,
		})

		return nil
	}

	// Order not found
	return errors.NewTradingError(errors.ErrStopOrderNotFound, "stop order not found", 0, nil)
}

// GetActiveStopOrders retrieves all active stop orders for a symbol
func (s *stopLossService) GetActiveStopOrders(symbol string) ([]*repository.StopOrder, error) {
	// Validate input
	if symbol == "" {
		return nil, errors.NewTradingError(errors.ErrInvalidParameter, "symbol cannot be empty", 0, nil)
	}

	// Get active stop orders from repository
	orders, err := s.stopOrderRepo.FindActiveStopOrders(symbol)
	if err != nil {
		s.logger.LogError(err, map[string]interface{}{
			"operation": "get_active_stop_orders",
			"symbol":    symbol,
		})
		return nil, err
	}

	s.logger.Debug("Active stop orders retrieved", map[string]interface{}{
		"symbol": symbol,
		"count":  len(orders),
	})

	return orders, nil
}

// UpdateTrailingStop updates the trail percent for a trailing stop order
func (s *stopLossService) UpdateTrailingStop(orderID string, newTrailPercent float64) error {
	// Validate input
	if orderID == "" {
		return errors.NewTradingError(errors.ErrInvalidParameter, "order ID cannot be empty", 0, nil)
	}

	if newTrailPercent <= 0 {
		return errors.NewTradingError(errors.ErrInvalidParameter, "trail percent must be greater than 0", 0, nil)
	}

	// Find trailing stop order
	trailingOrder, err := s.stopOrderRepo.FindTrailingStopOrderByID(orderID)
	if err != nil {
		s.logger.LogError(err, map[string]interface{}{
			"operation": "update_trailing_stop",
			"order_id":  orderID,
		})
		return err
	}

	// Update trail percent and recalculate stop price
	trailingOrder.TrailPercent = newTrailPercent
	trailingOrder.CurrentStopPrice = trailingOrder.HighestPrice * (1 - newTrailPercent/100)
	trailingOrder.LastUpdatedAt = time.Now().Unix()

	// Save updated order
	if err := s.stopOrderRepo.UpdateTrailingStopOrder(trailingOrder); err != nil {
		s.logger.LogError(err, map[string]interface{}{
			"operation": "update_trailing_stop",
			"order_id":  orderID,
		})
		return err
	}

	s.logger.Info("Trailing stop order updated", map[string]interface{}{
		"order_id":          orderID,
		"new_trail_percent": newTrailPercent,
		"new_stop_price":    trailingOrder.CurrentStopPrice,
	})

	return nil
}

// stopOrderIDCounter is used to generate unique order IDs
var stopOrderIDCounter int64 = 0
var stopOrderIDMutex sync.Mutex

// generateOrderID generates a unique order ID with a prefix
func generateOrderID(prefix string) string {
	stopOrderIDMutex.Lock()
	defer stopOrderIDMutex.Unlock()
	stopOrderIDCounter++
	return fmt.Sprintf("%s_%d_%d", prefix, time.Now().UnixNano(), stopOrderIDCounter)
}
