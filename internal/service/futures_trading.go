package service

import (
	"binance-trader/internal/api"
	"binance-trader/internal/repository"
	"binance-trader/pkg/errors"
	"binance-trader/pkg/logger"
	"fmt"
	"time"
)

// FuturesTradingService defines the interface for futures trading operations
type FuturesTradingService interface {
	// Open positions
	OpenLongPosition(symbol string, quantity float64, orderType api.OrderType, price float64) (*api.FuturesOrder, error)
	OpenShortPosition(symbol string, quantity float64, orderType api.OrderType, price float64) (*api.FuturesOrder, error)
	
	// Close positions
	ClosePosition(symbol string, positionSide api.PositionSide, quantity float64) (*api.FuturesOrder, error)
	CloseAllPositions(symbol string) ([]*api.FuturesOrder, error)
	
	// Order management
	CancelOrder(symbol string, orderID int64) error
	GetOrderStatus(orderID int64) (*api.FuturesOrder, error)
	GetActiveOrders(symbol string) ([]*api.FuturesOrder, error)
	
	// Leverage management
	SetLeverage(symbol string, leverage int) (*api.LeverageResponse, error)
	GetLeverage(symbol string) (int, error)
}

// futuresTradingService implements FuturesTradingService interface
type futuresTradingService struct {
	client     api.FuturesClient
	repository repository.FuturesOrderRepository
	logger     logger.Logger
}

// NewFuturesTradingService creates a new futures trading service
func NewFuturesTradingService(
	client api.FuturesClient,
	repository repository.FuturesOrderRepository,
	logger logger.Logger,
) FuturesTradingService {
	return &futuresTradingService{
		client:     client,
		repository: repository,
		logger:     logger,
	}
}

// OpenLongPosition opens a long position (buy)
func (s *futuresTradingService) OpenLongPosition(symbol string, quantity float64, orderType api.OrderType, price float64) (*api.FuturesOrder, error) {
	if symbol == "" {
		return nil, errors.NewTradingError(
			errors.ErrInvalidParameter,
			"symbol cannot be empty",
			0,
			nil,
		)
	}
	
	if quantity <= 0 {
		return nil, errors.NewTradingError(
			errors.ErrInvalidParameter,
			fmt.Sprintf("quantity must be greater than 0, got: %f", quantity),
			0,
			nil,
		)
	}
	
	// Validate price for limit orders
	if orderType == api.OrderTypeLimit && price <= 0 {
		return nil, errors.NewTradingError(
			errors.ErrInvalidParameter,
			"price must be greater than 0 for limit orders",
			0,
			nil,
		)
	}
	
	// Create order request for opening long position
	orderReq := &api.FuturesOrderRequest{
		Symbol:       symbol,
		Side:         api.OrderSideBuy,
		PositionSide: api.PositionSideLong,
		Type:         orderType,
		Quantity:     quantity,
		Price:        price,
	}
	
	s.logger.Info("Opening long position", map[string]interface{}{
		"symbol":    symbol,
		"quantity":  quantity,
		"type":      orderType,
		"price":     price,
	})
	
	// Create order via API
	response, err := s.client.CreateOrder(orderReq)
	if err != nil {
		s.logger.Error("Failed to open long position", map[string]interface{}{
			"symbol":   symbol,
			"quantity": quantity,
			"error":    err.Error(),
		})
		return nil, fmt.Errorf("failed to open long position: %w", err)
	}
	
	// Convert response to FuturesOrder and save to repository
	order := &api.FuturesOrder{
		OrderID:       response.OrderID,
		Symbol:        response.Symbol,
		Side:          response.Side,
		PositionSide:  response.PositionSide,
		Type:          response.Type,
		Status:        response.Status,
		Price:         response.Price,
		StopPrice:     response.StopPrice,
		OrigQty:       response.OrigQty,
		ExecutedQty:   response.ExecutedQty,
		AvgPrice:      response.AvgPrice,
		ReduceOnly:    response.ReduceOnly,
		ClosePosition: response.ClosePosition,
		Time:          time.Now().UnixMilli(),
		UpdateTime:    response.UpdateTime,
	}
	
	if err := s.repository.Save(order); err != nil {
		s.logger.Warn("Failed to save order to repository", map[string]interface{}{
			"order_id": order.OrderID,
			"error":    err.Error(),
		})
	}
	
	s.logger.Info("Long position opened successfully", map[string]interface{}{
		"order_id":     order.OrderID,
		"symbol":       order.Symbol,
		"status":       order.Status,
		"executed_qty": order.ExecutedQty,
		"avg_price":    order.AvgPrice,
	})
	
	return order, nil
}

// OpenShortPosition opens a short position (sell)
func (s *futuresTradingService) OpenShortPosition(symbol string, quantity float64, orderType api.OrderType, price float64) (*api.FuturesOrder, error) {
	if symbol == "" {
		return nil, errors.NewTradingError(
			errors.ErrInvalidParameter,
			"symbol cannot be empty",
			0,
			nil,
		)
	}
	
	if quantity <= 0 {
		return nil, errors.NewTradingError(
			errors.ErrInvalidParameter,
			fmt.Sprintf("quantity must be greater than 0, got: %f", quantity),
			0,
			nil,
		)
	}
	
	// Validate price for limit orders
	if orderType == api.OrderTypeLimit && price <= 0 {
		return nil, errors.NewTradingError(
			errors.ErrInvalidParameter,
			"price must be greater than 0 for limit orders",
			0,
			nil,
		)
	}
	
	// Create order request for opening short position
	orderReq := &api.FuturesOrderRequest{
		Symbol:       symbol,
		Side:         api.OrderSideSell,
		PositionSide: api.PositionSideShort,
		Type:         orderType,
		Quantity:     quantity,
		Price:        price,
	}
	
	s.logger.Info("Opening short position", map[string]interface{}{
		"symbol":    symbol,
		"quantity":  quantity,
		"type":      orderType,
		"price":     price,
	})
	
	// Create order via API
	response, err := s.client.CreateOrder(orderReq)
	if err != nil {
		s.logger.Error("Failed to open short position", map[string]interface{}{
			"symbol":   symbol,
			"quantity": quantity,
			"error":    err.Error(),
		})
		return nil, fmt.Errorf("failed to open short position: %w", err)
	}
	
	// Convert response to FuturesOrder and save to repository
	order := &api.FuturesOrder{
		OrderID:       response.OrderID,
		Symbol:        response.Symbol,
		Side:          response.Side,
		PositionSide:  response.PositionSide,
		Type:          response.Type,
		Status:        response.Status,
		Price:         response.Price,
		StopPrice:     response.StopPrice,
		OrigQty:       response.OrigQty,
		ExecutedQty:   response.ExecutedQty,
		AvgPrice:      response.AvgPrice,
		ReduceOnly:    response.ReduceOnly,
		ClosePosition: response.ClosePosition,
		Time:          time.Now().UnixMilli(),
		UpdateTime:    response.UpdateTime,
	}
	
	if err := s.repository.Save(order); err != nil {
		s.logger.Warn("Failed to save order to repository", map[string]interface{}{
			"order_id": order.OrderID,
			"error":    err.Error(),
		})
	}
	
	s.logger.Info("Short position opened successfully", map[string]interface{}{
		"order_id":     order.OrderID,
		"symbol":       order.Symbol,
		"status":       order.Status,
		"executed_qty": order.ExecutedQty,
		"avg_price":    order.AvgPrice,
	})
	
	return order, nil
}

// ClosePosition closes a position by creating an opposite order
func (s *futuresTradingService) ClosePosition(symbol string, positionSide api.PositionSide, quantity float64) (*api.FuturesOrder, error) {
	if symbol == "" {
		return nil, errors.NewTradingError(
			errors.ErrInvalidParameter,
			"symbol cannot be empty",
			0,
			nil,
		)
	}
	
	if quantity <= 0 {
		return nil, errors.NewTradingError(
			errors.ErrInvalidParameter,
			fmt.Sprintf("quantity must be greater than 0, got: %f", quantity),
			0,
			nil,
		)
	}
	
	// Determine the opposite side for closing
	var closeSide api.OrderSide
	if positionSide == api.PositionSideLong {
		closeSide = api.OrderSideSell
	} else if positionSide == api.PositionSideShort {
		closeSide = api.OrderSideBuy
	} else {
		return nil, errors.NewTradingError(
			errors.ErrInvalidParameter,
			fmt.Sprintf("invalid position side: %s", positionSide),
			0,
			nil,
		)
	}
	
	// Create order request for closing position
	orderReq := &api.FuturesOrderRequest{
		Symbol:       symbol,
		Side:         closeSide,
		PositionSide: positionSide,
		Type:         api.OrderTypeMarket,
		Quantity:     quantity,
		ReduceOnly:   true,
	}
	
	s.logger.Info("Closing position", map[string]interface{}{
		"symbol":        symbol,
		"position_side": positionSide,
		"quantity":      quantity,
		"close_side":    closeSide,
	})
	
	// Create order via API
	response, err := s.client.CreateOrder(orderReq)
	if err != nil {
		s.logger.Error("Failed to close position", map[string]interface{}{
			"symbol":        symbol,
			"position_side": positionSide,
			"quantity":      quantity,
			"error":         err.Error(),
		})
		return nil, fmt.Errorf("failed to close position: %w", err)
	}
	
	// Convert response to FuturesOrder and save to repository
	order := &api.FuturesOrder{
		OrderID:       response.OrderID,
		Symbol:        response.Symbol,
		Side:          response.Side,
		PositionSide:  response.PositionSide,
		Type:          response.Type,
		Status:        response.Status,
		Price:         response.Price,
		StopPrice:     response.StopPrice,
		OrigQty:       response.OrigQty,
		ExecutedQty:   response.ExecutedQty,
		AvgPrice:      response.AvgPrice,
		ReduceOnly:    response.ReduceOnly,
		ClosePosition: response.ClosePosition,
		Time:          time.Now().UnixMilli(),
		UpdateTime:    response.UpdateTime,
	}
	
	if err := s.repository.Save(order); err != nil {
		s.logger.Warn("Failed to save order to repository", map[string]interface{}{
			"order_id": order.OrderID,
			"error":    err.Error(),
		})
	}
	
	s.logger.Info("Position closed successfully", map[string]interface{}{
		"order_id":     order.OrderID,
		"symbol":       order.Symbol,
		"status":       order.Status,
		"executed_qty": order.ExecutedQty,
		"avg_price":    order.AvgPrice,
	})
	
	return order, nil
}

// CloseAllPositions closes all positions for a symbol
func (s *futuresTradingService) CloseAllPositions(symbol string) ([]*api.FuturesOrder, error) {
	if symbol == "" {
		return nil, errors.NewTradingError(
			errors.ErrInvalidParameter,
			"symbol cannot be empty",
			0,
			nil,
		)
	}
	
	s.logger.Info("Closing all positions", map[string]interface{}{
		"symbol": symbol,
	})
	
	// Get current positions
	positions, err := s.client.GetPositions(symbol)
	if err != nil {
		s.logger.Error("Failed to get positions", map[string]interface{}{
			"symbol": symbol,
			"error":  err.Error(),
		})
		return nil, fmt.Errorf("failed to get positions: %w", err)
	}
	
	var closedOrders []*api.FuturesOrder
	
	// Close each position with non-zero amount
	for _, pos := range positions {
		if pos.PositionAmt == 0 {
			continue
		}
		
		// Determine quantity (absolute value)
		quantity := pos.PositionAmt
		if quantity < 0 {
			quantity = -quantity
		}
		
		order, err := s.ClosePosition(symbol, pos.PositionSide, quantity)
		if err != nil {
			s.logger.Error("Failed to close position", map[string]interface{}{
				"symbol":        symbol,
				"position_side": pos.PositionSide,
				"error":         err.Error(),
			})
			continue
		}
		
		closedOrders = append(closedOrders, order)
	}
	
	s.logger.Info("All positions closed", map[string]interface{}{
		"symbol":        symbol,
		"orders_closed": len(closedOrders),
	})
	
	return closedOrders, nil
}

// CancelOrder cancels an existing order
func (s *futuresTradingService) CancelOrder(symbol string, orderID int64) error {
	if symbol == "" {
		return errors.NewTradingError(
			errors.ErrInvalidParameter,
			"symbol cannot be empty",
			0,
			nil,
		)
	}
	
	if orderID <= 0 {
		return errors.NewTradingError(
			errors.ErrInvalidParameter,
			"order ID must be greater than 0",
			0,
			nil,
		)
	}
	
	s.logger.Info("Cancelling order", map[string]interface{}{
		"symbol":   symbol,
		"order_id": orderID,
	})
	
	_, err := s.client.CancelOrder(symbol, orderID)
	if err != nil {
		s.logger.Error("Failed to cancel order", map[string]interface{}{
			"symbol":   symbol,
			"order_id": orderID,
			"error":    err.Error(),
		})
		return fmt.Errorf("failed to cancel order: %w", err)
	}
	
	// Update order status in repository
	if err := s.repository.SyncOrderStatus(orderID, api.OrderStatusCanceled, 0, 0, time.Now().UnixMilli()); err != nil {
		s.logger.Warn("Failed to update order status in repository", map[string]interface{}{
			"order_id": orderID,
			"error":    err.Error(),
		})
	}
	
	s.logger.Info("Order cancelled successfully", map[string]interface{}{
		"symbol":   symbol,
		"order_id": orderID,
	})
	
	return nil
}

// GetOrderStatus retrieves the current status of an order
func (s *futuresTradingService) GetOrderStatus(orderID int64) (*api.FuturesOrder, error) {
	if orderID <= 0 {
		return nil, errors.NewTradingError(
			errors.ErrInvalidParameter,
			"order ID must be greater than 0",
			0,
			nil,
		)
	}
	
	// Try to get from repository first
	order, err := s.repository.FindByID(orderID)
	if err == nil {
		return order, nil
	}
	
	// If not in repository, return error
	return nil, errors.NewTradingError(
		errors.ErrOrderNotFound,
		fmt.Sprintf("order %d not found", orderID),
		0,
		nil,
	)
}

// GetActiveOrders retrieves all active orders for a symbol
func (s *futuresTradingService) GetActiveOrders(symbol string) ([]*api.FuturesOrder, error) {
	if symbol == "" {
		return nil, errors.NewTradingError(
			errors.ErrInvalidParameter,
			"symbol cannot be empty",
			0,
			nil,
		)
	}
	
	s.logger.Debug("Getting active orders", map[string]interface{}{
		"symbol": symbol,
	})
	
	// Get open orders from API
	orders, err := s.client.GetOpenOrders(symbol)
	if err != nil {
		s.logger.Error("Failed to get active orders", map[string]interface{}{
			"symbol": symbol,
			"error":  err.Error(),
		})
		return nil, fmt.Errorf("failed to get active orders: %w", err)
	}
	
	s.logger.Debug("Retrieved active orders", map[string]interface{}{
		"symbol":       symbol,
		"orders_count": len(orders),
	})
	
	return orders, nil
}

// SetLeverage sets leverage for a symbol
func (s *futuresTradingService) SetLeverage(symbol string, leverage int) (*api.LeverageResponse, error) {
	if symbol == "" {
		return nil, errors.NewTradingError(
			errors.ErrInvalidParameter,
			"symbol cannot be empty",
			0,
			nil,
		)
	}
	
	if leverage < 1 || leverage > 125 {
		return nil, errors.NewTradingError(
			errors.ErrInvalidLeverage,
			fmt.Sprintf("leverage must be between 1 and 125, got: %d", leverage),
			0,
			nil,
		)
	}
	
	s.logger.Info("Setting leverage", map[string]interface{}{
		"symbol":   symbol,
		"leverage": leverage,
	})
	
	response, err := s.client.SetLeverage(symbol, leverage)
	if err != nil {
		s.logger.Error("Failed to set leverage", map[string]interface{}{
			"symbol":   symbol,
			"leverage": leverage,
			"error":    err.Error(),
		})
		return nil, fmt.Errorf("failed to set leverage: %w", err)
	}
	
	s.logger.Info("Leverage set successfully", map[string]interface{}{
		"symbol":             symbol,
		"leverage":           response.Leverage,
		"max_notional_value": response.MaxNotionalValue,
	})
	
	return response, nil
}

// GetLeverage retrieves current leverage for a symbol
func (s *futuresTradingService) GetLeverage(symbol string) (int, error) {
	if symbol == "" {
		return 0, errors.NewTradingError(
			errors.ErrInvalidParameter,
			"symbol cannot be empty",
			0,
			nil,
		)
	}
	
	// Get positions to extract leverage information
	positions, err := s.client.GetPositions(symbol)
	if err != nil {
		s.logger.Error("Failed to get positions for leverage info", map[string]interface{}{
			"symbol": symbol,
			"error":  err.Error(),
		})
		return 0, fmt.Errorf("failed to get leverage: %w", err)
	}
	
	// If we have positions, return the leverage from the first position
	if len(positions) > 0 {
		return positions[0].Leverage, nil
	}
	
	// If no positions, we can't determine leverage from positions
	return 0, nil
}
