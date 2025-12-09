package service

import (
	"binance-trader/internal/api"
	"binance-trader/internal/repository"
	"binance-trader/pkg/errors"
	"binance-trader/pkg/logger"
	"fmt"
	"sync"
	"time"
)

// FuturesStopLossService defines the interface for futures stop loss and take profit operations
type FuturesStopLossService interface {
	// Set stop loss and take profit
	SetStopLoss(symbol string, positionSide api.PositionSide, quantity float64, stopPrice float64) (*repository.StopOrder, error)
	SetTakeProfit(symbol string, positionSide api.PositionSide, quantity float64, targetPrice float64) (*repository.StopOrder, error)
	SetStopLossTakeProfit(symbol string, positionSide api.PositionSide, quantity float64, stopPrice, targetPrice float64) (*repository.StopOrderPair, error)
	SetTrailingStop(symbol string, positionSide api.PositionSide, quantity float64, callbackRate float64) (*repository.TrailingStopOrder, error)

	// Manage stop orders
	CancelStopOrder(orderID string) error
	GetActiveStopOrders(symbol string) ([]*repository.StopOrder, error)
	UpdateTrailingStop(orderID string, newCallbackRate float64) error
}

// futuresStopLossService implements the FuturesStopLossService interface
type futuresStopLossService struct {
	stopOrderRepo     repository.StopOrderRepository
	triggerEngine     TriggerEngine
	futuresTradingService FuturesTradingService
	futuresMarketService  FuturesMarketDataService
	logger            logger.Logger
}

// NewFuturesStopLossService creates a new futures stop loss service instance
func NewFuturesStopLossService(
	stopOrderRepo repository.StopOrderRepository,
	triggerEngine TriggerEngine,
	futuresTradingService FuturesTradingService,
	futuresMarketService FuturesMarketDataService,
	log logger.Logger,
) FuturesStopLossService {
	return &futuresStopLossService{
		stopOrderRepo:         stopOrderRepo,
		triggerEngine:         triggerEngine,
		futuresTradingService: futuresTradingService,
		futuresMarketService:  futuresMarketService,
		logger:                log,
	}
}

// SetStopLoss sets a stop loss order for a futures position
func (s *futuresStopLossService) SetStopLoss(symbol string, positionSide api.PositionSide, quantity float64, stopPrice float64) (*repository.StopOrder, error) {
	// Validate input parameters
	if symbol == "" {
		return nil, errors.NewTradingError(errors.ErrInvalidParameter, "symbol cannot be empty", 0, nil)
	}

	if quantity <= 0 {
		return nil, errors.NewTradingError(errors.ErrInvalidParameter, "quantity must be greater than 0", 0, nil)
	}

	if stopPrice <= 0 {
		return nil, errors.NewTradingError(errors.ErrInvalidParameter, "stop price must be greater than 0", 0, nil)
	}

	if positionSide != api.PositionSideLong && positionSide != api.PositionSideShort {
		return nil, errors.NewTradingError(errors.ErrInvalidParameter, "position side must be LONG or SHORT", 0, nil)
	}

	// Create stop order
	stopOrder := &repository.StopOrder{
		OrderID:   generateFuturesStopOrderID("FSL"),
		Symbol:    symbol,
		Position:  quantity,
		StopPrice: stopPrice,
		Type:      repository.StopOrderTypeStopLoss,
		Status:    repository.StopOrderStatusActive,
		CreatedAt: time.Now().Unix(),
	}

	// Save to repository
	if err := s.stopOrderRepo.SaveStopOrder(stopOrder); err != nil {
		s.logger.LogError(err, map[string]interface{}{
			"operation":     "set_futures_stop_loss",
			"symbol":        symbol,
			"position_side": positionSide,
			"quantity":      quantity,
			"stop_price":    stopPrice,
		})
		return nil, err
	}

	// Register trigger condition based on position side
	var operator ComparisonOperator
	if positionSide == api.PositionSideLong {
		// For long positions, stop loss triggers when price falls below stop price
		operator = OperatorLessEqual
	} else {
		// For short positions, stop loss triggers when price rises above stop price
		operator = OperatorGreaterEqual
	}

	condition := &TriggerCondition{
		Type:     TriggerTypePrice,
		Operator: operator,
		Value:    stopPrice,
	}

	if err := s.triggerEngine.RegisterCondition(stopOrder.OrderID, condition); err != nil {
		s.logger.Warn("Failed to register trigger condition", map[string]interface{}{
			"order_id": stopOrder.OrderID,
			"error":    err.Error(),
		})
	}

	s.logger.Info("Futures stop loss order created", map[string]interface{}{
		"order_id":      stopOrder.OrderID,
		"symbol":        symbol,
		"position_side": positionSide,
		"quantity":      quantity,
		"stop_price":    stopPrice,
	})

	return stopOrder, nil
}

// SetTakeProfit sets a take profit order for a futures position
func (s *futuresStopLossService) SetTakeProfit(symbol string, positionSide api.PositionSide, quantity float64, targetPrice float64) (*repository.StopOrder, error) {
	// Validate input parameters
	if symbol == "" {
		return nil, errors.NewTradingError(errors.ErrInvalidParameter, "symbol cannot be empty", 0, nil)
	}

	if quantity <= 0 {
		return nil, errors.NewTradingError(errors.ErrInvalidParameter, "quantity must be greater than 0", 0, nil)
	}

	if targetPrice <= 0 {
		return nil, errors.NewTradingError(errors.ErrInvalidParameter, "target price must be greater than 0", 0, nil)
	}

	if positionSide != api.PositionSideLong && positionSide != api.PositionSideShort {
		return nil, errors.NewTradingError(errors.ErrInvalidParameter, "position side must be LONG or SHORT", 0, nil)
	}

	// Create take profit order
	takeProfitOrder := &repository.StopOrder{
		OrderID:   generateFuturesStopOrderID("FTP"),
		Symbol:    symbol,
		Position:  quantity,
		StopPrice: targetPrice,
		Type:      repository.StopOrderTypeTakeProfit,
		Status:    repository.StopOrderStatusActive,
		CreatedAt: time.Now().Unix(),
	}

	// Save to repository
	if err := s.stopOrderRepo.SaveStopOrder(takeProfitOrder); err != nil {
		s.logger.LogError(err, map[string]interface{}{
			"operation":     "set_futures_take_profit",
			"symbol":        symbol,
			"position_side": positionSide,
			"quantity":      quantity,
			"target_price":  targetPrice,
		})
		return nil, err
	}

	// Register trigger condition based on position side
	var operator ComparisonOperator
	if positionSide == api.PositionSideLong {
		// For long positions, take profit triggers when price rises above target
		operator = OperatorGreaterEqual
	} else {
		// For short positions, take profit triggers when price falls below target
		operator = OperatorLessEqual
	}

	condition := &TriggerCondition{
		Type:     TriggerTypePrice,
		Operator: operator,
		Value:    targetPrice,
	}

	if err := s.triggerEngine.RegisterCondition(takeProfitOrder.OrderID, condition); err != nil {
		s.logger.Warn("Failed to register trigger condition", map[string]interface{}{
			"order_id": takeProfitOrder.OrderID,
			"error":    err.Error(),
		})
	}

	s.logger.Info("Futures take profit order created", map[string]interface{}{
		"order_id":      takeProfitOrder.OrderID,
		"symbol":        symbol,
		"position_side": positionSide,
		"quantity":      quantity,
		"target_price":  targetPrice,
	})

	return takeProfitOrder, nil
}

// SetStopLossTakeProfit sets both stop loss and take profit orders as a pair
func (s *futuresStopLossService) SetStopLossTakeProfit(symbol string, positionSide api.PositionSide, quantity float64, stopPrice, targetPrice float64) (*repository.StopOrderPair, error) {
	// Validate input parameters
	if symbol == "" {
		return nil, errors.NewTradingError(errors.ErrInvalidParameter, "symbol cannot be empty", 0, nil)
	}

	if quantity <= 0 {
		return nil, errors.NewTradingError(errors.ErrInvalidParameter, "quantity must be greater than 0", 0, nil)
	}

	if stopPrice <= 0 {
		return nil, errors.NewTradingError(errors.ErrInvalidParameter, "stop price must be greater than 0", 0, nil)
	}

	if targetPrice <= 0 {
		return nil, errors.NewTradingError(errors.ErrInvalidParameter, "target price must be greater than 0", 0, nil)
	}

	if positionSide != api.PositionSideLong && positionSide != api.PositionSideShort {
		return nil, errors.NewTradingError(errors.ErrInvalidParameter, "position side must be LONG or SHORT", 0, nil)
	}

	// Create stop loss order
	stopLossOrder := &repository.StopOrder{
		OrderID:   generateFuturesStopOrderID("FSL"),
		Symbol:    symbol,
		Position:  quantity,
		StopPrice: stopPrice,
		Type:      repository.StopOrderTypeStopLoss,
		Status:    repository.StopOrderStatusActive,
		CreatedAt: time.Now().Unix(),
	}

	// Create take profit order
	takeProfitOrder := &repository.StopOrder{
		OrderID:   generateFuturesStopOrderID("FTP"),
		Symbol:    symbol,
		Position:  quantity,
		StopPrice: targetPrice,
		Type:      repository.StopOrderTypeTakeProfit,
		Status:    repository.StopOrderStatusActive,
		CreatedAt: time.Now().Unix(),
	}

	// Create order pair
	orderPair := &repository.StopOrderPair{
		PairID:          generateFuturesStopOrderID("FPAIR"),
		Symbol:          symbol,
		Position:        quantity,
		StopLossOrder:   stopLossOrder,
		TakeProfitOrder: takeProfitOrder,
		Status:          "ACTIVE",
	}

	// Save individual orders
	if err := s.stopOrderRepo.SaveStopOrder(stopLossOrder); err != nil {
		s.logger.LogError(err, map[string]interface{}{
			"operation":  "set_futures_stop_loss_take_profit",
			"symbol":     symbol,
			"stop_price": stopPrice,
		})
		return nil, err
	}

	if err := s.stopOrderRepo.SaveStopOrder(takeProfitOrder); err != nil {
		s.logger.LogError(err, map[string]interface{}{
			"operation":    "set_futures_stop_loss_take_profit",
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
			"operation": "set_futures_stop_loss_take_profit",
			"symbol":    symbol,
		})
		// Rollback both orders
		s.stopOrderRepo.DeleteStopOrder(stopLossOrder.OrderID)
		s.stopOrderRepo.DeleteStopOrder(takeProfitOrder.OrderID)
		return nil, err
	}

	// Register trigger conditions based on position side
	var stopLossOperator, takeProfitOperator ComparisonOperator
	if positionSide == api.PositionSideLong {
		stopLossOperator = OperatorLessEqual
		takeProfitOperator = OperatorGreaterEqual
	} else {
		stopLossOperator = OperatorGreaterEqual
		takeProfitOperator = OperatorLessEqual
	}

	stopLossCondition := &TriggerCondition{
		Type:     TriggerTypePrice,
		Operator: stopLossOperator,
		Value:    stopPrice,
	}

	takeProfitCondition := &TriggerCondition{
		Type:     TriggerTypePrice,
		Operator: takeProfitOperator,
		Value:    targetPrice,
	}

	s.triggerEngine.RegisterCondition(stopLossOrder.OrderID, stopLossCondition)
	s.triggerEngine.RegisterCondition(takeProfitOrder.OrderID, takeProfitCondition)

	s.logger.Info("Futures stop loss and take profit pair created", map[string]interface{}{
		"pair_id":       orderPair.PairID,
		"symbol":        symbol,
		"position_side": positionSide,
		"quantity":      quantity,
		"stop_price":    stopPrice,
		"target_price":  targetPrice,
	})

	return orderPair, nil
}

// SetTrailingStop sets a trailing stop order for a futures position
func (s *futuresStopLossService) SetTrailingStop(symbol string, positionSide api.PositionSide, quantity float64, callbackRate float64) (*repository.TrailingStopOrder, error) {
	// Validate input parameters
	if symbol == "" {
		return nil, errors.NewTradingError(errors.ErrInvalidParameter, "symbol cannot be empty", 0, nil)
	}

	if quantity <= 0 {
		return nil, errors.NewTradingError(errors.ErrInvalidParameter, "quantity must be greater than 0", 0, nil)
	}

	if callbackRate <= 0 {
		return nil, errors.NewTradingError(errors.ErrInvalidParameter, "callback rate must be greater than 0", 0, nil)
	}

	if positionSide != api.PositionSideLong && positionSide != api.PositionSideShort {
		return nil, errors.NewTradingError(errors.ErrInvalidParameter, "position side must be LONG or SHORT", 0, nil)
	}

	// Get current mark price to initialize highest/lowest price
	markPrice, err := s.futuresMarketService.GetMarkPrice(symbol)
	if err != nil {
		s.logger.LogError(err, map[string]interface{}{
			"operation": "set_futures_trailing_stop",
			"symbol":    symbol,
		})
		return nil, err
	}

	// Calculate initial stop price based on position side
	var initialStopPrice, extremePrice float64
	if positionSide == api.PositionSideLong {
		// For long positions, track highest price and stop below it
		extremePrice = markPrice
		initialStopPrice = markPrice * (1 - callbackRate/100)
	} else {
		// For short positions, track lowest price and stop above it
		extremePrice = markPrice
		initialStopPrice = markPrice * (1 + callbackRate/100)
	}

	// Create trailing stop order
	trailingStopOrder := &repository.TrailingStopOrder{
		OrderID:          generateFuturesStopOrderID("FTS"),
		Symbol:           symbol,
		Position:         quantity,
		TrailPercent:     callbackRate,
		HighestPrice:     extremePrice,
		CurrentStopPrice: initialStopPrice,
		Status:           repository.StopOrderStatusActive,
		CreatedAt:        time.Now().Unix(),
		LastUpdatedAt:    time.Now().Unix(),
	}

	// Save to repository
	if err := s.stopOrderRepo.SaveTrailingStopOrder(trailingStopOrder); err != nil {
		s.logger.LogError(err, map[string]interface{}{
			"operation":     "set_futures_trailing_stop",
			"symbol":        symbol,
			"position_side": positionSide,
			"quantity":      quantity,
			"callback_rate": callbackRate,
		})
		return nil, err
	}

	s.logger.Info("Futures trailing stop order created", map[string]interface{}{
		"order_id":       trailingStopOrder.OrderID,
		"symbol":         symbol,
		"position_side":  positionSide,
		"quantity":       quantity,
		"callback_rate":  callbackRate,
		"extreme_price":  extremePrice,
		"initial_stop":   initialStopPrice,
	})

	return trailingStopOrder, nil
}

// CancelStopOrder cancels a stop order
func (s *futuresStopLossService) CancelStopOrder(orderID string) error {
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
				"operation": "cancel_futures_stop_order",
				"order_id":  orderID,
			})
			return err
		}

		// Unregister trigger condition
		s.triggerEngine.UnregisterCondition(orderID)

		s.logger.Info("Futures stop order cancelled", map[string]interface{}{
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
				"operation": "cancel_futures_stop_order",
				"order_id":  orderID,
			})
			return err
		}

		s.logger.Info("Futures trailing stop order cancelled", map[string]interface{}{
			"order_id": orderID,
			"symbol":   trailingOrder.Symbol,
		})

		return nil
	}

	// Order not found
	return errors.NewTradingError(errors.ErrStopOrderNotFound, "stop order not found", 0, nil)
}

// GetActiveStopOrders retrieves all active stop orders for a symbol
func (s *futuresStopLossService) GetActiveStopOrders(symbol string) ([]*repository.StopOrder, error) {
	// Validate input
	if symbol == "" {
		return nil, errors.NewTradingError(errors.ErrInvalidParameter, "symbol cannot be empty", 0, nil)
	}

	// Get active stop orders from repository
	orders, err := s.stopOrderRepo.FindActiveStopOrders(symbol)
	if err != nil {
		s.logger.LogError(err, map[string]interface{}{
			"operation": "get_active_futures_stop_orders",
			"symbol":    symbol,
		})
		return nil, err
	}

	s.logger.Debug("Active futures stop orders retrieved", map[string]interface{}{
		"symbol": symbol,
		"count":  len(orders),
	})

	return orders, nil
}

// UpdateTrailingStop updates the callback rate for a trailing stop order
func (s *futuresStopLossService) UpdateTrailingStop(orderID string, newCallbackRate float64) error {
	// Validate input
	if orderID == "" {
		return errors.NewTradingError(errors.ErrInvalidParameter, "order ID cannot be empty", 0, nil)
	}

	if newCallbackRate <= 0 {
		return errors.NewTradingError(errors.ErrInvalidParameter, "callback rate must be greater than 0", 0, nil)
	}

	// Find trailing stop order
	trailingOrder, err := s.stopOrderRepo.FindTrailingStopOrderByID(orderID)
	if err != nil {
		s.logger.LogError(err, map[string]interface{}{
			"operation": "update_futures_trailing_stop",
			"order_id":  orderID,
		})
		return err
	}

	// Update callback rate and recalculate stop price
	trailingOrder.TrailPercent = newCallbackRate
	trailingOrder.CurrentStopPrice = trailingOrder.HighestPrice * (1 - newCallbackRate/100)
	trailingOrder.LastUpdatedAt = time.Now().Unix()

	// Save updated order
	if err := s.stopOrderRepo.UpdateTrailingStopOrder(trailingOrder); err != nil {
		s.logger.LogError(err, map[string]interface{}{
			"operation": "update_futures_trailing_stop",
			"order_id":  orderID,
		})
		return err
	}

	s.logger.Info("Futures trailing stop order updated", map[string]interface{}{
		"order_id":          orderID,
		"new_callback_rate": newCallbackRate,
		"new_stop_price":    trailingOrder.CurrentStopPrice,
	})

	return nil
}

// UpdateTrailingStopPrice updates the trailing stop price based on current market price
// This should be called periodically by the monitoring engine
func (s *futuresStopLossService) UpdateTrailingStopPrice(orderID string, positionSide api.PositionSide, currentPrice float64) (bool, error) {
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

	if positionSide == api.PositionSideLong {
		// For long positions, track highest price
		if currentPrice > trailingOrder.HighestPrice {
			trailingOrder.HighestPrice = currentPrice
			trailingOrder.CurrentStopPrice = currentPrice * (1 - trailingOrder.TrailPercent/100)
			trailingOrder.LastUpdatedAt = time.Now().Unix()
			updated = true

			// Save updated order
			if err := s.stopOrderRepo.UpdateTrailingStopOrder(trailingOrder); err != nil {
				s.logger.LogError(err, map[string]interface{}{
					"operation": "update_futures_trailing_stop_price",
					"order_id":  orderID,
				})
				return false, err
			}

			s.logger.Debug("Futures trailing stop price adjusted (LONG)", map[string]interface{}{
				"order_id":        orderID,
				"current_price":   currentPrice,
				"highest_price":   trailingOrder.HighestPrice,
				"new_stop_price":  trailingOrder.CurrentStopPrice,
				"callback_rate":   trailingOrder.TrailPercent,
			})
		}

		// Check if stop price is triggered (price fell below stop)
		if currentPrice <= trailingOrder.CurrentStopPrice {
			if err := s.triggerTrailingStop(trailingOrder, positionSide, currentPrice); err != nil {
				s.logger.LogError(err, map[string]interface{}{
					"operation": "trigger_futures_trailing_stop",
					"order_id":  orderID,
				})
				return false, err
			}
			return true, nil
		}
	} else {
		// For short positions, track lowest price
		if currentPrice < trailingOrder.HighestPrice || trailingOrder.HighestPrice == 0 {
			trailingOrder.HighestPrice = currentPrice // Using HighestPrice field to store lowest for shorts
			trailingOrder.CurrentStopPrice = currentPrice * (1 + trailingOrder.TrailPercent/100)
			trailingOrder.LastUpdatedAt = time.Now().Unix()
			updated = true

			// Save updated order
			if err := s.stopOrderRepo.UpdateTrailingStopOrder(trailingOrder); err != nil {
				s.logger.LogError(err, map[string]interface{}{
					"operation": "update_futures_trailing_stop_price",
					"order_id":  orderID,
				})
				return false, err
			}

			s.logger.Debug("Futures trailing stop price adjusted (SHORT)", map[string]interface{}{
				"order_id":        orderID,
				"current_price":   currentPrice,
				"lowest_price":    trailingOrder.HighestPrice,
				"new_stop_price":  trailingOrder.CurrentStopPrice,
				"callback_rate":   trailingOrder.TrailPercent,
			})
		}

		// Check if stop price is triggered (price rose above stop)
		if currentPrice >= trailingOrder.CurrentStopPrice {
			if err := s.triggerTrailingStop(trailingOrder, positionSide, currentPrice); err != nil {
				s.logger.LogError(err, map[string]interface{}{
					"operation": "trigger_futures_trailing_stop",
					"order_id":  orderID,
				})
				return false, err
			}
			return true, nil
		}
	}

	return updated, nil
}

// triggerTrailingStop executes a trailing stop order when triggered
func (s *futuresStopLossService) triggerTrailingStop(order *repository.TrailingStopOrder, positionSide api.PositionSide, triggerPrice float64) error {
	// Update status to triggered
	order.Status = repository.StopOrderStatusTriggered
	if err := s.stopOrderRepo.UpdateTrailingStopOrder(order); err != nil {
		return err
	}

	// Execute market order to close position
	executedOrder, err := s.futuresTradingService.ClosePosition(order.Symbol, positionSide, order.Position)
	if err != nil {
		s.logger.LogError(err, map[string]interface{}{
			"operation":     "execute_futures_trailing_stop",
			"order_id":      order.OrderID,
			"symbol":        order.Symbol,
			"position_side": positionSide,
			"quantity":      order.Position,
			"trigger_price": triggerPrice,
		})
		return err
	}

	s.logger.Info("Futures trailing stop order triggered and executed", map[string]interface{}{
		"order_id":          order.OrderID,
		"symbol":            order.Symbol,
		"position_side":     positionSide,
		"quantity":          order.Position,
		"trigger_price":     triggerPrice,
		"stop_price":        order.CurrentStopPrice,
		"extreme_price":     order.HighestPrice,
		"callback_rate":     order.TrailPercent,
		"executed_order_id": executedOrder.OrderID,
	})

	return nil
}

// futuresStopOrderIDCounter is used to generate unique order IDs
var futuresStopOrderIDCounter int64 = 0
var futuresStopOrderIDMutex sync.Mutex

// generateFuturesStopOrderID generates a unique order ID with a prefix
func generateFuturesStopOrderID(prefix string) string {
	futuresStopOrderIDMutex.Lock()
	defer futuresStopOrderIDMutex.Unlock()
	futuresStopOrderIDCounter++
	return fmt.Sprintf("%s_%d_%d", prefix, time.Now().UnixNano(), futuresStopOrderIDCounter)
}
