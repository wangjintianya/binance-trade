package service

import (
	"binance-trader/internal/api"
	"binance-trader/internal/repository"
	"binance-trader/pkg/errors"
	"binance-trader/pkg/logger"
	"strings"
)

// SpotTradingService defines the interface for spot trading operations
type SpotTradingService interface {
	// Order creation
	PlaceMarketBuyOrder(symbol string, quantity float64) (*api.Order, error)
	PlaceMarketSellOrder(symbol string, quantity float64) (*api.Order, error)
	PlaceLimitSellOrder(symbol string, price, quantity float64) (*api.Order, error)

	// Order management
	CancelOrder(orderID int64) error
	GetOrderStatus(orderID int64) (*OrderStatus, error)
	GetActiveOrders() ([]*api.Order, error)
}

// spotTradingService implements the SpotTradingService interface
type spotTradingService struct {
	client    api.SpotClient
	riskMgr   RiskManager
	orderRepo repository.OrderRepository
	logger    logger.Logger
}

// NewSpotTradingService creates a new spot trading service instance
func NewSpotTradingService(
	client api.SpotClient,
	riskMgr RiskManager,
	orderRepo repository.OrderRepository,
	log logger.Logger,
) SpotTradingService {
	return &spotTradingService{
		client:    client,
		riskMgr:   riskMgr,
		orderRepo: orderRepo,
		logger:    log,
	}
}

// PlaceMarketBuyOrder places a market buy order
func (s *spotTradingService) PlaceMarketBuyOrder(symbol string, quantity float64) (*api.Order, error) {
	// Validate input parameters
	if symbol == "" {
		s.logger.Error("Market buy order failed: empty symbol", map[string]interface{}{
			"quantity": quantity,
		})
		return nil, errors.NewTradingError(
			errors.ErrInvalidParameter,
			"symbol cannot be empty",
			0,
			nil,
		)
	}
	
	if quantity <= 0 {
		s.logger.Error("Market buy order failed: invalid quantity", map[string]interface{}{
			"symbol":   symbol,
			"quantity": quantity,
		})
		return nil, errors.NewTradingError(
			errors.ErrInvalidParameter,
			"quantity must be greater than 0",
			0,
			nil,
		)
	}
	
	// Create order request
	orderReq := &api.OrderRequest{
		Symbol:   symbol,
		Side:     api.OrderSideBuy,
		Type:     api.OrderTypeMarket,
		Quantity: quantity,
	}
	
	// Validate order with risk manager
	if err := s.riskMgr.ValidateOrder(orderReq); err != nil {
		s.logger.Error("Market buy order failed risk validation", map[string]interface{}{
			"symbol":   symbol,
			"quantity": quantity,
			"error":    err.Error(),
		})
		return nil, err
	}
	
	// Check daily limit
	if err := s.riskMgr.CheckDailyLimit(); err != nil {
		s.logger.Error("Market buy order failed daily limit check", map[string]interface{}{
			"symbol":   symbol,
			"quantity": quantity,
			"error":    err.Error(),
		})
		return nil, err
	}
	
	// Extract quote asset from symbol (e.g., USDT from BTCUSDT)
	quoteAsset := extractQuoteAsset(symbol)
	if quoteAsset != "" {
		if err := s.riskMgr.CheckMinimumBalance(quoteAsset); err != nil {
			s.logger.Error("Market buy order failed minimum balance check", map[string]interface{}{
				"symbol":      symbol,
				"quantity":    quantity,
				"quote_asset": quoteAsset,
				"error":       err.Error(),
			})
			return nil, err
		}
	}
	
	// Place order via API
	s.logger.Info("Placing market buy order", map[string]interface{}{
		"symbol":   symbol,
		"quantity": quantity,
	})
	
	orderResp, err := s.client.CreateOrder(orderReq)
	if err != nil {
		s.logger.LogError(err, map[string]interface{}{
			"operation": "place_market_buy_order",
			"symbol":    symbol,
			"quantity":  quantity,
		})
		return nil, err
	}
	
	// Convert response to Order
	order := &api.Order{
		OrderID:             orderResp.OrderID,
		Symbol:              orderResp.Symbol,
		Side:                api.OrderSideBuy,
		Type:                api.OrderTypeMarket,
		Status:              orderResp.Status,
		Price:               orderResp.Price,
		OrigQty:             orderResp.OrigQty,
		ExecutedQty:         orderResp.ExecutedQty,
		CummulativeQuoteQty: orderResp.CummulativeQuoteQty,
		Time:                orderResp.TransactTime,
		UpdateTime:          orderResp.TransactTime,
	}
	
	// Save order to repository
	if err := s.orderRepo.Save(order); err != nil {
		s.logger.Warn("Failed to save order to repository", map[string]interface{}{
			"order_id": order.OrderID,
			"error":    err.Error(),
		})
	}
	
	// Record order in risk manager
	if rm, ok := s.riskMgr.(*riskManager); ok {
		rm.RecordOrder(orderResp.CummulativeQuoteQty)
	}
	
	// Log order event
	s.logger.LogOrderEvent(
		"order_created",
		order.OrderID,
		order.Symbol,
		string(order.Side),
		string(order.Type),
		order.OrigQty,
		map[string]interface{}{
			"status":       string(order.Status),
			"executed_qty": order.ExecutedQty,
			"price":        order.Price,
		},
	)
	
	return order, nil
}

// PlaceMarketSellOrder places a market sell order
func (s *spotTradingService) PlaceMarketSellOrder(symbol string, quantity float64) (*api.Order, error) {
	// Validate input parameters
	if symbol == "" {
		s.logger.Error("Market sell order failed: empty symbol", map[string]interface{}{
			"quantity": quantity,
		})
		return nil, errors.NewTradingError(
			errors.ErrInvalidParameter,
			"symbol cannot be empty",
			0,
			nil,
		)
	}
	
	if quantity <= 0 {
		s.logger.Error("Market sell order failed: invalid quantity", map[string]interface{}{
			"symbol":   symbol,
			"quantity": quantity,
		})
		return nil, errors.NewTradingError(
			errors.ErrInvalidParameter,
			"quantity must be greater than 0",
			0,
			nil,
		)
	}
	
	// Create order request
	orderReq := &api.OrderRequest{
		Symbol:   symbol,
		Side:     api.OrderSideSell,
		Type:     api.OrderTypeMarket,
		Quantity: quantity,
	}
	
	// Validate order with risk manager
	if err := s.riskMgr.ValidateOrder(orderReq); err != nil {
		s.logger.Error("Market sell order failed risk validation", map[string]interface{}{
			"symbol":   symbol,
			"quantity": quantity,
			"error":    err.Error(),
		})
		return nil, err
	}
	
	// Check daily limit
	if err := s.riskMgr.CheckDailyLimit(); err != nil {
		s.logger.Error("Market sell order failed daily limit check", map[string]interface{}{
			"symbol":   symbol,
			"quantity": quantity,
			"error":    err.Error(),
		})
		return nil, err
	}
	
	// Place order via API
	s.logger.Info("Placing market sell order", map[string]interface{}{
		"symbol":   symbol,
		"quantity": quantity,
	})
	
	orderResp, err := s.client.CreateOrder(orderReq)
	if err != nil {
		s.logger.LogError(err, map[string]interface{}{
			"operation": "place_market_sell_order",
			"symbol":    symbol,
			"quantity":  quantity,
		})
		return nil, err
	}
	
	// Convert response to Order
	order := &api.Order{
		OrderID:             orderResp.OrderID,
		Symbol:              orderResp.Symbol,
		Side:                api.OrderSideSell,
		Type:                api.OrderTypeMarket,
		Status:              orderResp.Status,
		Price:               orderResp.Price,
		OrigQty:             orderResp.OrigQty,
		ExecutedQty:         orderResp.ExecutedQty,
		CummulativeQuoteQty: orderResp.CummulativeQuoteQty,
		Time:                orderResp.TransactTime,
		UpdateTime:          orderResp.TransactTime,
	}
	
	// Save order to repository
	if err := s.orderRepo.Save(order); err != nil {
		s.logger.Warn("Failed to save order to repository", map[string]interface{}{
			"order_id": order.OrderID,
			"error":    err.Error(),
		})
	}
	
	// Record order in risk manager
	if rm, ok := s.riskMgr.(*riskManager); ok {
		rm.RecordOrder(orderResp.CummulativeQuoteQty)
	}
	
	// Log order event
	s.logger.LogOrderEvent(
		"order_created",
		order.OrderID,
		order.Symbol,
		string(order.Side),
		string(order.Type),
		order.OrigQty,
		map[string]interface{}{
			"status":       string(order.Status),
			"executed_qty": order.ExecutedQty,
			"price":        order.Price,
		},
	)
	
	return order, nil
}

// PlaceLimitSellOrder places a limit sell order
func (s *spotTradingService) PlaceLimitSellOrder(symbol string, price, quantity float64) (*api.Order, error) {
	// Validate input parameters
	if symbol == "" {
		s.logger.Error("Limit sell order failed: empty symbol", map[string]interface{}{
			"price":    price,
			"quantity": quantity,
		})
		return nil, errors.NewTradingError(
			errors.ErrInvalidParameter,
			"symbol cannot be empty",
			0,
			nil,
		)
	}
	
	if price <= 0 {
		s.logger.Error("Limit sell order failed: invalid price", map[string]interface{}{
			"symbol":   symbol,
			"price":    price,
			"quantity": quantity,
		})
		return nil, errors.NewTradingError(
			errors.ErrInvalidParameter,
			"price must be greater than 0",
			0,
			nil,
		)
	}
	
	if quantity <= 0 {
		s.logger.Error("Limit sell order failed: invalid quantity", map[string]interface{}{
			"symbol":   symbol,
			"price":    price,
			"quantity": quantity,
		})
		return nil, errors.NewTradingError(
			errors.ErrInvalidParameter,
			"quantity must be greater than 0",
			0,
			nil,
		)
	}
	
	// Create order request
	orderReq := &api.OrderRequest{
		Symbol:      symbol,
		Side:        api.OrderSideSell,
		Type:        api.OrderTypeLimit,
		Quantity:    quantity,
		Price:       price,
		TimeInForce: "GTC", // Good Till Cancel
	}
	
	// Validate order with risk manager
	if err := s.riskMgr.ValidateOrder(orderReq); err != nil {
		s.logger.Error("Limit sell order failed risk validation", map[string]interface{}{
			"symbol":   symbol,
			"price":    price,
			"quantity": quantity,
			"error":    err.Error(),
		})
		return nil, err
	}
	
	// Check daily limit
	if err := s.riskMgr.CheckDailyLimit(); err != nil {
		s.logger.Error("Limit sell order failed daily limit check", map[string]interface{}{
			"symbol":   symbol,
			"price":    price,
			"quantity": quantity,
			"error":    err.Error(),
		})
		return nil, err
	}
	
	// Place order via API
	s.logger.Info("Placing limit sell order", map[string]interface{}{
		"symbol":   symbol,
		"price":    price,
		"quantity": quantity,
	})
	
	orderResp, err := s.client.CreateOrder(orderReq)
	if err != nil {
		s.logger.LogError(err, map[string]interface{}{
			"operation": "place_limit_sell_order",
			"symbol":    symbol,
			"price":     price,
			"quantity":  quantity,
		})
		return nil, err
	}
	
	// Convert response to Order
	order := &api.Order{
		OrderID:             orderResp.OrderID,
		Symbol:              orderResp.Symbol,
		Side:                api.OrderSideSell,
		Type:                api.OrderTypeLimit,
		Status:              orderResp.Status,
		Price:               orderResp.Price,
		OrigQty:             orderResp.OrigQty,
		ExecutedQty:         orderResp.ExecutedQty,
		CummulativeQuoteQty: orderResp.CummulativeQuoteQty,
		Time:                orderResp.TransactTime,
		UpdateTime:          orderResp.TransactTime,
	}
	
	// Save order to repository
	if err := s.orderRepo.Save(order); err != nil {
		s.logger.Warn("Failed to save order to repository", map[string]interface{}{
			"order_id": order.OrderID,
			"error":    err.Error(),
		})
	}
	
	// Record order in risk manager
	if rm, ok := s.riskMgr.(*riskManager); ok {
		rm.RecordOrder(price * quantity)
	}
	
	// Log order event
	s.logger.LogOrderEvent(
		"order_created",
		order.OrderID,
		order.Symbol,
		string(order.Side),
		string(order.Type),
		order.OrigQty,
		map[string]interface{}{
			"status":       string(order.Status),
			"executed_qty": order.ExecutedQty,
			"price":        order.Price,
		},
	)
	
	return order, nil
}

// CancelOrder cancels an existing order
func (s *spotTradingService) CancelOrder(orderID int64) error {
	// Validate input
	if orderID <= 0 {
		s.logger.Error("Cancel order failed: invalid order ID", map[string]interface{}{
			"order_id": orderID,
		})
		return errors.NewTradingError(
			errors.ErrInvalidParameter,
			"order ID must be greater than 0",
			0,
			nil,
		)
	}
	
	// Get order from repository to find symbol
	order, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		s.logger.Error("Cancel order failed: order not found in repository", map[string]interface{}{
			"order_id": orderID,
			"error":    err.Error(),
		})
		return err
	}
	
	// Cancel order via API
	s.logger.Info("Canceling order", map[string]interface{}{
		"order_id": orderID,
		"symbol":   order.Symbol,
	})
	
	cancelResp, err := s.client.CancelOrder(order.Symbol, orderID)
	if err != nil {
		s.logger.LogError(err, map[string]interface{}{
			"operation": "cancel_order",
			"order_id":  orderID,
			"symbol":    order.Symbol,
		})
		return err
	}
	
	// Update order status in repository
	if err := s.orderRepo.SyncOrderStatus(orderID, cancelResp.Status, order.ExecutedQty, order.UpdateTime); err != nil {
		s.logger.Warn("Failed to update order status in repository", map[string]interface{}{
			"order_id": orderID,
			"error":    err.Error(),
		})
	}
	
	// Log order event
	s.logger.LogOrderEvent(
		"order_canceled",
		orderID,
		order.Symbol,
		string(order.Side),
		string(order.Type),
		order.OrigQty,
		map[string]interface{}{
			"status": string(cancelResp.Status),
		},
	)
	
	return nil
}

// GetOrderStatus retrieves the current status of an order
func (s *spotTradingService) GetOrderStatus(orderID int64) (*OrderStatus, error) {
	// Validate input
	if orderID <= 0 {
		s.logger.Error("Get order status failed: invalid order ID", map[string]interface{}{
			"order_id": orderID,
		})
		return nil, errors.NewTradingError(
			errors.ErrInvalidParameter,
			"order ID must be greater than 0",
			0,
			nil,
		)
	}
	
	// Get order from repository
	order, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		s.logger.Error("Get order status failed: order not found in repository", map[string]interface{}{
			"order_id": orderID,
			"error":    err.Error(),
		})
		return nil, err
	}
	
	// Query API for latest status
	s.logger.Debug("Querying order status from API", map[string]interface{}{
		"order_id": orderID,
		"symbol":   order.Symbol,
	})
	
	apiOrder, err := s.client.GetOrder(order.Symbol, orderID)
	if err != nil {
		s.logger.LogError(err, map[string]interface{}{
			"operation": "get_order_status",
			"order_id":  orderID,
			"symbol":    order.Symbol,
		})
		return nil, err
	}
	
	// Sync status with repository
	if err := s.orderRepo.SyncOrderStatus(orderID, apiOrder.Status, apiOrder.ExecutedQty, apiOrder.UpdateTime); err != nil {
		s.logger.Warn("Failed to sync order status in repository", map[string]interface{}{
			"order_id": orderID,
			"error":    err.Error(),
		})
	}
	
	// Create and return order status
	status := &OrderStatus{
		OrderID:     apiOrder.OrderID,
		Symbol:      apiOrder.Symbol,
		Status:      apiOrder.Status,
		ExecutedQty: apiOrder.ExecutedQty,
		Price:       apiOrder.Price,
	}
	
	s.logger.Debug("Order status retrieved", map[string]interface{}{
		"order_id":     status.OrderID,
		"symbol":       status.Symbol,
		"status":       string(status.Status),
		"executed_qty": status.ExecutedQty,
	})
	
	return status, nil
}

// GetActiveOrders retrieves all active (open) orders
func (s *spotTradingService) GetActiveOrders() ([]*api.Order, error) {
	s.logger.Debug("Retrieving active orders", nil)
	
	// Get open orders from API
	apiOrders, err := s.client.GetOpenOrders("")
	if err != nil {
		s.logger.LogError(err, map[string]interface{}{
			"operation": "get_active_orders",
		})
		return nil, err
	}
	
	// Sync with repository
	for _, apiOrder := range apiOrders {
		// Try to find in repository
		repoOrder, err := s.orderRepo.FindByID(apiOrder.OrderID)
		if err != nil {
			// Order not in repository, save it
			if err := s.orderRepo.Save(apiOrder); err != nil {
				s.logger.Warn("Failed to save order to repository", map[string]interface{}{
					"order_id": apiOrder.OrderID,
					"error":    err.Error(),
				})
			}
		} else {
			// Order exists, sync status
			if repoOrder.Status != apiOrder.Status || repoOrder.ExecutedQty != apiOrder.ExecutedQty {
				if err := s.orderRepo.SyncOrderStatus(apiOrder.OrderID, apiOrder.Status, apiOrder.ExecutedQty, apiOrder.UpdateTime); err != nil {
					s.logger.Warn("Failed to sync order status", map[string]interface{}{
						"order_id": apiOrder.OrderID,
						"error":    err.Error(),
					})
				}
			}
		}
	}
	
	s.logger.Info("Active orders retrieved", map[string]interface{}{
		"count": len(apiOrders),
	})
	
	return apiOrders, nil
}

// extractQuoteAsset extracts the quote asset from a trading pair symbol
// For example: BTCUSDT -> USDT, ETHBTC -> BTC
func extractQuoteAsset(symbol string) string {
	// Common quote assets in order of priority
	quoteAssets := []string{"USDT", "BUSD", "USDC", "BTC", "ETH", "BNB"}
	
	for _, quote := range quoteAssets {
		if strings.HasSuffix(symbol, quote) {
			return quote
		}
	}
	
	// If no common quote asset found, return empty string
	return ""
}
