package service

import (
	"binance-trader/internal/api"
	"binance-trader/internal/config"
	"binance-trader/pkg/errors"
	"binance-trader/pkg/logger"
	"fmt"
	"math"
	"sync"
)

// FuturesRiskMetrics represents risk metrics for futures trading
type FuturesRiskMetrics struct {
	TotalPositionValue    float64
	TotalUnrealizedPnL    float64
	TotalMarginUsed       float64
	AvailableMargin       float64
	MarginRatio           float64
	PositionsAtRisk       int
	LeverageUtilization   float64
}

// FuturesRiskManager defines the interface for futures risk management
type FuturesRiskManager interface {
	// Risk checks
	ValidateOrder(order *api.FuturesOrderRequest) error
	CheckLiquidationRisk(position *api.Position, markPrice float64) (bool, error)
	CheckMarginSufficiency(symbol string, quantity float64, leverage int) error
	CheckMaxPositionSize(symbol string, quantity float64) error
	
	// Risk monitoring
	MonitorPositions() error
	GetRiskMetrics() (*FuturesRiskMetrics, error)
	
	// Limit management
	UpdateLimits(limits *config.FuturesRiskConfig) error
	GetCurrentLimits() *config.FuturesRiskConfig
}

// futuresRiskManager implements FuturesRiskManager interface
type futuresRiskManager struct {
	limits         *config.FuturesRiskConfig
	client         api.FuturesClient
	positionMgr    FuturesPositionManager
	logger         logger.Logger
	mu             sync.RWMutex
}

// NewFuturesRiskManager creates a new futures risk manager
func NewFuturesRiskManager(
	limits *config.FuturesRiskConfig,
	client api.FuturesClient,
	positionMgr FuturesPositionManager,
	logger logger.Logger,
) FuturesRiskManager {
	if limits == nil {
		// Default limits
		limits = &config.FuturesRiskConfig{
			MaxOrderValue:     50000.0,
			MaxPositionValue:  100000.0,
			MaxLeverage:       20,
			MinMarginRatio:    0.05,
			LiquidationBuffer: 0.02,
			MaxDailyOrders:    200,
			MaxAPICallsPerMin: 2000,
		}
	}
	
	return &futuresRiskManager{
		limits:      limits,
		client:      client,
		positionMgr: positionMgr,
		logger:      logger,
	}
}

// ValidateOrder validates a futures order against risk limits
func (rm *futuresRiskManager) ValidateOrder(order *api.FuturesOrderRequest) error {
	if order == nil {
		return errors.NewTradingError(
			errors.ErrInvalidParameter,
			"order request cannot be nil",
			0,
			nil,
		)
	}
	
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	
	// Calculate order value
	var orderValue float64
	if order.Type == api.OrderTypeMarket {
		// For market orders, get current price
		price, err := rm.client.GetPrice(order.Symbol)
		if err != nil {
			return errors.NewTradingError(
				errors.ErrInvalidParameter,
				fmt.Sprintf("failed to get price for order validation: %v", err),
				0,
				err,
			)
		}
		orderValue = order.Quantity * price.Price
	} else {
		// For limit orders, use specified price
		orderValue = order.Quantity * order.Price
	}
	
	// Check order value limit
	if orderValue > rm.limits.MaxOrderValue {
		rm.logger.Warn("Order value exceeds limit", map[string]interface{}{
			"order_value": orderValue,
			"max_value":   rm.limits.MaxOrderValue,
			"symbol":      order.Symbol,
		})
		return errors.NewTradingError(
			errors.ErrRiskLimitExceeded,
			fmt.Sprintf("order value %.2f exceeds maximum limit %.2f", orderValue, rm.limits.MaxOrderValue),
			0,
			nil,
		)
	}
	
	// Check if this would exceed max position value (only for opening positions)
	if !order.ReduceOnly {
		// Get current positions
		positions, err := rm.client.GetPositions(order.Symbol)
		if err != nil {
			rm.logger.Warn("Failed to get positions for risk check", map[string]interface{}{
				"symbol": order.Symbol,
				"error":  err.Error(),
			})
			// Continue with validation even if we can't get positions
		} else {
			totalPositionValue := orderValue
			for _, pos := range positions {
				posValue := math.Abs(pos.PositionAmt * pos.EntryPrice)
				totalPositionValue += posValue
			}
			
			if totalPositionValue > rm.limits.MaxPositionValue {
				rm.logger.Warn("Total position value would exceed limit", map[string]interface{}{
					"total_value": totalPositionValue,
					"max_value":   rm.limits.MaxPositionValue,
					"symbol":      order.Symbol,
				})
				return errors.NewTradingError(
					errors.ErrMaxPositionExceeded,
					fmt.Sprintf("total position value %.2f would exceed maximum limit %.2f", totalPositionValue, rm.limits.MaxPositionValue),
					0,
					nil,
				)
			}
		}
	}
	
	rm.logger.Debug("Order validated successfully", map[string]interface{}{
		"symbol":      order.Symbol,
		"order_value": orderValue,
	})
	
	return nil
}

// CheckLiquidationRisk checks if a position is at risk of liquidation
func (rm *futuresRiskManager) CheckLiquidationRisk(position *api.Position, markPrice float64) (bool, error) {
	if position == nil {
		return false, errors.NewTradingError(
			errors.ErrInvalidParameter,
			"position cannot be nil",
			0,
			nil,
		)
	}
	
	if markPrice <= 0 {
		return false, errors.NewTradingError(
			errors.ErrInvalidParameter,
			"mark price must be greater than 0",
			0,
			nil,
		)
	}
	
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	
	// Calculate liquidation price
	liquidationPrice, err := rm.positionMgr.CalculateLiquidationPrice(position)
	if err != nil {
		return false, fmt.Errorf("failed to calculate liquidation price: %w", err)
	}
	
	// Check if mark price is within liquidation buffer
	var distanceToLiquidation float64
	if position.PositionAmt > 0 {
		// Long position: liquidation when price drops
		distanceToLiquidation = (markPrice - liquidationPrice) / markPrice
	} else {
		// Short position: liquidation when price rises
		distanceToLiquidation = (liquidationPrice - markPrice) / markPrice
	}
	
	atRisk := distanceToLiquidation < rm.limits.LiquidationBuffer
	
	if atRisk {
		rm.logger.Warn("Position at liquidation risk", map[string]interface{}{
			"symbol":                 position.Symbol,
			"position_side":          position.PositionSide,
			"mark_price":             markPrice,
			"liquidation_price":      liquidationPrice,
			"distance_to_liquidation": distanceToLiquidation,
			"buffer":                 rm.limits.LiquidationBuffer,
		})
	}
	
	return atRisk, nil
}

// CheckMarginSufficiency checks if there's sufficient margin for an order
func (rm *futuresRiskManager) CheckMarginSufficiency(symbol string, quantity float64, leverage int) error {
	if symbol == "" {
		return errors.NewTradingError(
			errors.ErrInvalidParameter,
			"symbol cannot be empty",
			0,
			nil,
		)
	}
	
	if quantity <= 0 {
		return errors.NewTradingError(
			errors.ErrInvalidParameter,
			"quantity must be greater than 0",
			0,
			nil,
		)
	}
	
	if leverage < 1 || leverage > 125 {
		return errors.NewTradingError(
			errors.ErrInvalidLeverage,
			fmt.Sprintf("leverage must be between 1 and 125, got: %d", leverage),
			0,
			nil,
		)
	}
	
	// Get current price
	price, err := rm.client.GetPrice(symbol)
	if err != nil {
		return errors.NewTradingError(
			errors.ErrInvalidParameter,
			fmt.Sprintf("failed to get price for margin check: %v", err),
			0,
			err,
		)
	}
	
	// Calculate required margin
	orderValue := quantity * price.Price
	requiredMargin := orderValue / float64(leverage)
	
	// Get available balance
	balance, err := rm.client.GetBalance()
	if err != nil {
		return errors.NewTradingError(
			errors.ErrInvalidParameter,
			fmt.Sprintf("failed to get balance for margin check: %v", err),
			0,
			err,
		)
	}
	
	// Check if available balance is sufficient
	if balance.AvailableBalance < requiredMargin {
		rm.logger.Warn("Insufficient margin", map[string]interface{}{
			"symbol":            symbol,
			"required_margin":   requiredMargin,
			"available_balance": balance.AvailableBalance,
			"leverage":          leverage,
		})
		return errors.NewTradingError(
			errors.ErrInsufficientMargin,
			fmt.Sprintf("insufficient margin: required %.2f, available %.2f", requiredMargin, balance.AvailableBalance),
			0,
			nil,
		)
	}
	
	rm.logger.Debug("Margin sufficiency check passed", map[string]interface{}{
		"symbol":            symbol,
		"required_margin":   requiredMargin,
		"available_balance": balance.AvailableBalance,
	})
	
	return nil
}

// CheckMaxPositionSize checks if an order would exceed maximum position size
func (rm *futuresRiskManager) CheckMaxPositionSize(symbol string, quantity float64) error {
	if symbol == "" {
		return errors.NewTradingError(
			errors.ErrInvalidParameter,
			"symbol cannot be empty",
			0,
			nil,
		)
	}
	
	if quantity <= 0 {
		return errors.NewTradingError(
			errors.ErrInvalidParameter,
			"quantity must be greater than 0",
			0,
			nil,
		)
	}
	
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	
	// Get current price
	price, err := rm.client.GetPrice(symbol)
	if err != nil {
		return errors.NewTradingError(
			errors.ErrInvalidParameter,
			fmt.Sprintf("failed to get price for position size check: %v", err),
			0,
			err,
		)
	}
	
	// Calculate new order value
	orderValue := quantity * price.Price
	
	// Get current positions
	positions, err := rm.client.GetPositions(symbol)
	if err != nil {
		rm.logger.Warn("Failed to get positions for size check", map[string]interface{}{
			"symbol": symbol,
			"error":  err.Error(),
		})
		// Continue with check using only new order value
		if orderValue > rm.limits.MaxPositionValue {
			return errors.NewTradingError(
				errors.ErrMaxPositionExceeded,
				fmt.Sprintf("order value %.2f exceeds maximum position limit %.2f", orderValue, rm.limits.MaxPositionValue),
				0,
				nil,
			)
		}
		return nil
	}
	
	// Calculate total position value including new order
	totalPositionValue := orderValue
	for _, pos := range positions {
		posValue := math.Abs(pos.PositionAmt * pos.EntryPrice)
		totalPositionValue += posValue
	}
	
	if totalPositionValue > rm.limits.MaxPositionValue {
		rm.logger.Warn("Total position value would exceed limit", map[string]interface{}{
			"symbol":      symbol,
			"total_value": totalPositionValue,
			"max_value":   rm.limits.MaxPositionValue,
		})
		return errors.NewTradingError(
			errors.ErrMaxPositionExceeded,
			fmt.Sprintf("total position value %.2f would exceed maximum limit %.2f", totalPositionValue, rm.limits.MaxPositionValue),
			0,
			nil,
		)
	}
	
	rm.logger.Debug("Position size check passed", map[string]interface{}{
		"symbol":      symbol,
		"total_value": totalPositionValue,
		"max_value":   rm.limits.MaxPositionValue,
	})
	
	return nil
}

// MonitorPositions monitors all positions for risk
func (rm *futuresRiskManager) MonitorPositions() error {
	rm.logger.Debug("Monitoring positions for risk", nil)
	
	// Get all positions
	positions, err := rm.positionMgr.GetAllPositions()
	if err != nil {
		rm.logger.Error("Failed to get positions for monitoring", map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("failed to get positions: %w", err)
	}
	
	positionsAtRisk := 0
	
	// Check each position
	for _, pos := range positions {
		if pos.PositionAmt == 0 {
			continue
		}
		
		// Get current mark price
		markPrice, err := rm.client.GetMarkPrice(pos.Symbol)
		if err != nil {
			rm.logger.Warn("Failed to get mark price for position monitoring", map[string]interface{}{
				"symbol": pos.Symbol,
				"error":  err.Error(),
			})
			continue
		}
		
		// Check liquidation risk
		atRisk, err := rm.CheckLiquidationRisk(pos, markPrice.MarkPrice)
		if err != nil {
			rm.logger.Warn("Failed to check liquidation risk", map[string]interface{}{
				"symbol": pos.Symbol,
				"error":  err.Error(),
			})
			continue
		}
		
		if atRisk {
			positionsAtRisk++
		}
		
		// Check margin ratio
		marginRatio, err := rm.positionMgr.CalculateMarginRatio(pos)
		if err != nil {
			rm.logger.Warn("Failed to calculate margin ratio", map[string]interface{}{
				"symbol": pos.Symbol,
				"error":  err.Error(),
			})
			continue
		}
		
		rm.mu.RLock()
		minMarginRatio := rm.limits.MinMarginRatio
		rm.mu.RUnlock()
		
		if marginRatio > minMarginRatio {
			rm.logger.Warn("Position margin ratio below threshold", map[string]interface{}{
				"symbol":            pos.Symbol,
				"position_side":     pos.PositionSide,
				"margin_ratio":      marginRatio,
				"min_margin_ratio":  minMarginRatio,
			})
		}
	}
	
	rm.logger.Debug("Position monitoring complete", map[string]interface{}{
		"total_positions":    len(positions),
		"positions_at_risk":  positionsAtRisk,
	})
	
	return nil
}

// GetRiskMetrics calculates and returns current risk metrics
func (rm *futuresRiskManager) GetRiskMetrics() (*FuturesRiskMetrics, error) {
	rm.logger.Debug("Calculating risk metrics", nil)
	
	// Get all positions
	positions, err := rm.positionMgr.GetAllPositions()
	if err != nil {
		rm.logger.Error("Failed to get positions for risk metrics", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to get positions: %w", err)
	}
	
	// Get account balance
	balance, err := rm.client.GetBalance()
	if err != nil {
		rm.logger.Error("Failed to get balance for risk metrics", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}
	
	metrics := &FuturesRiskMetrics{
		AvailableMargin: balance.AvailableBalance,
	}
	
	// Calculate metrics from positions
	for _, pos := range positions {
		if pos.PositionAmt == 0 {
			continue
		}
		
		posValue := math.Abs(pos.PositionAmt * pos.EntryPrice)
		metrics.TotalPositionValue += posValue
		metrics.TotalUnrealizedPnL += pos.UnrealizedProfit
		metrics.TotalMarginUsed += pos.PositionInitialMargin
		
		// Check if position is at risk
		markPrice, err := rm.client.GetMarkPrice(pos.Symbol)
		if err != nil {
			rm.logger.Warn("Failed to get mark price for risk metrics", map[string]interface{}{
				"symbol": pos.Symbol,
				"error":  err.Error(),
			})
			continue
		}
		
		atRisk, err := rm.CheckLiquidationRisk(pos, markPrice.MarkPrice)
		if err != nil {
			rm.logger.Warn("Failed to check liquidation risk for metrics", map[string]interface{}{
				"symbol": pos.Symbol,
				"error":  err.Error(),
			})
			continue
		}
		
		if atRisk {
			metrics.PositionsAtRisk++
		}
	}
	
	// Calculate margin ratio
	if metrics.TotalMarginUsed > 0 {
		marginBalance := metrics.TotalMarginUsed + metrics.TotalUnrealizedPnL
		if marginBalance > 0 {
			// Simplified margin ratio calculation
			metrics.MarginRatio = metrics.TotalMarginUsed / marginBalance
		}
	}
	
	// Calculate leverage utilization
	if balance.Balance > 0 {
		metrics.LeverageUtilization = metrics.TotalPositionValue / balance.Balance
	}
	
	rm.logger.Debug("Risk metrics calculated", map[string]interface{}{
		"total_position_value": metrics.TotalPositionValue,
		"total_unrealized_pnl": metrics.TotalUnrealizedPnL,
		"positions_at_risk":    metrics.PositionsAtRisk,
	})
	
	return metrics, nil
}

// UpdateLimits updates the risk limits
func (rm *futuresRiskManager) UpdateLimits(limits *config.FuturesRiskConfig) error {
	if limits == nil {
		return errors.NewTradingError(
			errors.ErrInvalidParameter,
			"limits cannot be nil",
			0,
			nil,
		)
	}
	
	// Validate limits
	if limits.MaxOrderValue <= 0 {
		return errors.NewTradingError(
			errors.ErrInvalidParameter,
			"max order value must be greater than 0",
			0,
			nil,
		)
	}
	if limits.MaxPositionValue <= 0 {
		return errors.NewTradingError(
			errors.ErrInvalidParameter,
			"max position value must be greater than 0",
			0,
			nil,
		)
	}
	if limits.MaxLeverage < 1 || limits.MaxLeverage > 125 {
		return errors.NewTradingError(
			errors.ErrInvalidParameter,
			"max leverage must be between 1 and 125",
			0,
			nil,
		)
	}
	if limits.MinMarginRatio < 0 || limits.MinMarginRatio > 1 {
		return errors.NewTradingError(
			errors.ErrInvalidParameter,
			"min margin ratio must be between 0 and 1",
			0,
			nil,
		)
	}
	if limits.LiquidationBuffer < 0 || limits.LiquidationBuffer > 1 {
		return errors.NewTradingError(
			errors.ErrInvalidParameter,
			"liquidation buffer must be between 0 and 1",
			0,
			nil,
		)
	}
	
	rm.mu.Lock()
	defer rm.mu.Unlock()
	
	rm.limits = limits
	
	rm.logger.Info("Risk limits updated", map[string]interface{}{
		"max_order_value":    limits.MaxOrderValue,
		"max_position_value": limits.MaxPositionValue,
		"max_leverage":       limits.MaxLeverage,
		"min_margin_ratio":   limits.MinMarginRatio,
		"liquidation_buffer": limits.LiquidationBuffer,
	})
	
	return nil
}

// GetCurrentLimits returns the current risk limits
func (rm *futuresRiskManager) GetCurrentLimits() *config.FuturesRiskConfig {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	
	// Return a copy to prevent external modification
	return &config.FuturesRiskConfig{
		MaxOrderValue:     rm.limits.MaxOrderValue,
		MaxPositionValue:  rm.limits.MaxPositionValue,
		MaxLeverage:       rm.limits.MaxLeverage,
		MinMarginRatio:    rm.limits.MinMarginRatio,
		LiquidationBuffer: rm.limits.LiquidationBuffer,
		MaxDailyOrders:    rm.limits.MaxDailyOrders,
		MaxAPICallsPerMin: rm.limits.MaxAPICallsPerMin,
	}
}
