package service

import (
	"binance-trader/internal/api"
	"binance-trader/internal/repository"
	"binance-trader/pkg/errors"
	"binance-trader/pkg/logger"
	"fmt"
	"math"
)

// FuturesPositionManager defines the interface for futures position management
type FuturesPositionManager interface {
	// Position queries
	GetPosition(symbol string, positionSide api.PositionSide) (*api.Position, error)
	GetAllPositions() ([]*api.Position, error)
	
	// Position calculations
	CalculateUnrealizedPnL(position *api.Position, markPrice float64) (float64, error)
	CalculateLiquidationPrice(position *api.Position) (float64, error)
	CalculateMarginRatio(position *api.Position) (float64, error)
	
	// Position updates
	UpdatePosition(symbol string) error
	UpdateAllPositions() error
	
	// Position history
	GetPositionHistory(symbol string, startTime, endTime int64) ([]*repository.ClosedPosition, error)
}

// futuresPositionManager implements FuturesPositionManager interface
type futuresPositionManager struct {
	client     api.FuturesClient
	repository repository.FuturesPositionRepository
	logger     logger.Logger
}

// NewFuturesPositionManager creates a new futures position manager
func NewFuturesPositionManager(
	client api.FuturesClient,
	repository repository.FuturesPositionRepository,
	logger logger.Logger,
) FuturesPositionManager {
	return &futuresPositionManager{
		client:     client,
		repository: repository,
		logger:     logger,
	}
}

// GetPosition retrieves a specific position
func (m *futuresPositionManager) GetPosition(symbol string, positionSide api.PositionSide) (*api.Position, error) {
	if symbol == "" {
		return nil, errors.NewTradingError(
			errors.ErrInvalidParameter,
			"symbol cannot be empty",
			0,
			nil,
		)
	}
	
	m.logger.Debug("Getting position", map[string]interface{}{
		"symbol":        symbol,
		"position_side": positionSide,
	})
	
	// Get positions from API
	positions, err := m.client.GetPositions(symbol)
	if err != nil {
		m.logger.Error("Failed to get position", map[string]interface{}{
			"symbol": symbol,
			"error":  err.Error(),
		})
		return nil, fmt.Errorf("failed to get position: %w", err)
	}
	
	// Find the specific position side
	for _, pos := range positions {
		if pos.PositionSide == positionSide {
			// Save to repository
			if err := m.repository.SavePosition(pos); err != nil {
				m.logger.Warn("Failed to save position to repository", map[string]interface{}{
					"symbol": symbol,
					"error":  err.Error(),
				})
			}
			return pos, nil
		}
	}
	
	return nil, errors.NewTradingError(
		errors.ErrPositionNotFound,
		fmt.Sprintf("position not found for %s %s", symbol, positionSide),
		0,
		nil,
	)
}

// GetAllPositions retrieves all positions
func (m *futuresPositionManager) GetAllPositions() ([]*api.Position, error) {
	m.logger.Debug("Getting all positions", nil)
	
	// Get all positions from API
	positions, err := m.client.GetAllPositions()
	if err != nil {
		m.logger.Error("Failed to get all positions", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to get all positions: %w", err)
	}
	
	// Save all positions to repository
	for _, pos := range positions {
		if err := m.repository.SavePosition(pos); err != nil {
			m.logger.Warn("Failed to save position to repository", map[string]interface{}{
				"symbol": pos.Symbol,
				"error":  err.Error(),
			})
		}
	}
	
	m.logger.Debug("Retrieved all positions", map[string]interface{}{
		"count": len(positions),
	})
	
	return positions, nil
}

// CalculateUnrealizedPnL calculates unrealized profit and loss
func (m *futuresPositionManager) CalculateUnrealizedPnL(position *api.Position, markPrice float64) (float64, error) {
	if position == nil {
		return 0, errors.NewTradingError(
			errors.ErrInvalidParameter,
			"position cannot be nil",
			0,
			nil,
		)
	}
	
	if markPrice <= 0 {
		return 0, errors.NewTradingError(
			errors.ErrInvalidParameter,
			"mark price must be greater than 0",
			0,
			nil,
		)
	}
	
	// Calculate PnL based on position direction
	// For long: (markPrice - entryPrice) * positionAmt
	// For short: (entryPrice - markPrice) * abs(positionAmt)
	pnl := (markPrice - position.EntryPrice) * position.PositionAmt
	
	return pnl, nil
}

// CalculateLiquidationPrice calculates the liquidation price
func (m *futuresPositionManager) CalculateLiquidationPrice(position *api.Position) (float64, error) {
	if position == nil {
		return 0, errors.NewTradingError(
			errors.ErrInvalidParameter,
			"position cannot be nil",
			0,
			nil,
		)
	}
	
	if position.PositionAmt == 0 {
		return 0, nil
	}
	
	// Simplified liquidation price calculation
	// For long: entryPrice - (margin / abs(positionAmt))
	// For short: entryPrice + (margin / abs(positionAmt))
	
	leverage := float64(position.Leverage)
	if leverage == 0 {
		leverage = 1
	}
	
	// Maintenance margin rate (simplified, typically 0.4% for low leverage)
	maintenanceMarginRate := 0.004
	
	var liquidationPrice float64
	
	if position.PositionAmt > 0 {
		// Long position
		liquidationPrice = position.EntryPrice * (1 - (1/leverage) + maintenanceMarginRate)
	} else {
		// Short position
		liquidationPrice = position.EntryPrice * (1 + (1/leverage) - maintenanceMarginRate)
	}
	
	return liquidationPrice, nil
}

// CalculateMarginRatio calculates the margin ratio
func (m *futuresPositionManager) CalculateMarginRatio(position *api.Position) (float64, error) {
	if position == nil {
		return 0, errors.NewTradingError(
			errors.ErrInvalidParameter,
			"position cannot be nil",
			0,
			nil,
		)
	}
	
	if position.PositionInitialMargin == 0 {
		return 0, nil
	}
	
	// Margin ratio = MaintenanceMargin / (PositionInitialMargin + UnrealizedProfit)
	marginBalance := position.PositionInitialMargin + position.UnrealizedProfit
	
	if marginBalance <= 0 {
		return math.Inf(1), nil // Infinite margin ratio indicates liquidation risk
	}
	
	marginRatio := position.MaintenanceMargin / marginBalance
	
	return marginRatio, nil
}

// UpdatePosition updates a specific position from the API
func (m *futuresPositionManager) UpdatePosition(symbol string) error {
	if symbol == "" {
		return errors.NewTradingError(
			errors.ErrInvalidParameter,
			"symbol cannot be empty",
			0,
			nil,
		)
	}
	
	m.logger.Debug("Updating position", map[string]interface{}{
		"symbol": symbol,
	})
	
	// Get positions from API
	positions, err := m.client.GetPositions(symbol)
	if err != nil {
		m.logger.Error("Failed to update position", map[string]interface{}{
			"symbol": symbol,
			"error":  err.Error(),
		})
		return fmt.Errorf("failed to update position: %w", err)
	}
	
	// Save all positions for this symbol
	for _, pos := range positions {
		if err := m.repository.SavePosition(pos); err != nil {
			m.logger.Warn("Failed to save position to repository", map[string]interface{}{
				"symbol": symbol,
				"error":  err.Error(),
			})
		}
	}
	
	m.logger.Debug("Position updated", map[string]interface{}{
		"symbol": symbol,
		"count":  len(positions),
	})
	
	return nil
}

// UpdateAllPositions updates all positions from the API
func (m *futuresPositionManager) UpdateAllPositions() error {
	m.logger.Debug("Updating all positions", nil)
	
	// Get all positions from API
	positions, err := m.client.GetAllPositions()
	if err != nil {
		m.logger.Error("Failed to update all positions", map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("failed to update all positions: %w", err)
	}
	
	// Save all positions
	for _, pos := range positions {
		if err := m.repository.SavePosition(pos); err != nil {
			m.logger.Warn("Failed to save position to repository", map[string]interface{}{
				"symbol": pos.Symbol,
				"error":  err.Error(),
			})
		}
	}
	
	m.logger.Debug("All positions updated", map[string]interface{}{
		"count": len(positions),
	})
	
	return nil
}

// GetPositionHistory retrieves closed position history
func (m *futuresPositionManager) GetPositionHistory(symbol string, startTime, endTime int64) ([]*repository.ClosedPosition, error) {
	if symbol == "" {
		return nil, errors.NewTradingError(
			errors.ErrInvalidParameter,
			"symbol cannot be empty",
			0,
			nil,
		)
	}
	
	if startTime < 0 || endTime < 0 {
		return nil, errors.NewTradingError(
			errors.ErrInvalidParameter,
			"time values cannot be negative",
			0,
			nil,
		)
	}
	
	if startTime > endTime {
		return nil, errors.NewTradingError(
			errors.ErrInvalidParameter,
			"start time cannot be after end time",
			0,
			nil,
		)
	}
	
	m.logger.Debug("Getting position history", map[string]interface{}{
		"symbol":     symbol,
		"start_time": startTime,
		"end_time":   endTime,
	})
	
	// Get position history from repository
	history, err := m.repository.GetPositionHistory(symbol, startTime, endTime)
	if err != nil {
		m.logger.Error("Failed to get position history", map[string]interface{}{
			"symbol": symbol,
			"error":  err.Error(),
		})
		return nil, fmt.Errorf("failed to get position history: %w", err)
	}
	
	m.logger.Debug("Retrieved position history", map[string]interface{}{
		"symbol": symbol,
		"count":  len(history),
	})
	
	return history, nil
}
