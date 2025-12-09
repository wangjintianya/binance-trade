package service

import (
	"binance-trader/internal/api"
	"binance-trader/pkg/errors"
	"binance-trader/pkg/logger"
	"fmt"
)

// FuturesLeverageService defines the interface for leverage and margin management
type FuturesLeverageService interface {
	// Leverage management
	SetLeverage(symbol string, leverage int) (*api.LeverageResponse, error)
	GetLeverage(symbol string) (int, error)
	
	// Margin type management
	SetMarginType(symbol string, marginType api.MarginType) error
	
	// Position mode management
	SetPositionMode(dualSidePosition bool) error
	GetPositionMode() (*api.PositionMode, error)
	
	// Position checking
	HasOpenPositions(symbol string) (bool, error)
	HasAnyOpenPositions() (bool, error)
}

// futuresLeverageService implements FuturesLeverageService interface
type futuresLeverageService struct {
	client api.FuturesClient
	logger logger.Logger
}

// NewFuturesLeverageService creates a new futures leverage service
func NewFuturesLeverageService(client api.FuturesClient, logger logger.Logger) FuturesLeverageService {
	return &futuresLeverageService{
		client: client,
		logger: logger,
	}
}

// SetLeverage sets leverage for a symbol with validation
func (s *futuresLeverageService) SetLeverage(symbol string, leverage int) (*api.LeverageResponse, error) {
	if symbol == "" {
		return nil, errors.NewTradingError(
			errors.ErrInvalidParameter,
			"symbol cannot be empty",
			0,
			nil,
		)
	}
	
	// Validate leverage range (1-125)
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
func (s *futuresLeverageService) GetLeverage(symbol string) (int, error) {
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
	// Return 0 to indicate no leverage is set
	return 0, nil
}

// SetMarginType sets margin type for a symbol with position checking
func (s *futuresLeverageService) SetMarginType(symbol string, marginType api.MarginType) error {
	if symbol == "" {
		return errors.NewTradingError(
			errors.ErrInvalidParameter,
			"symbol cannot be empty",
			0,
			nil,
		)
	}
	
	// Validate margin type
	if marginType != api.MarginTypeIsolated && marginType != api.MarginTypeCrossed {
		return errors.NewTradingError(
			errors.ErrInvalidParameter,
			fmt.Sprintf("invalid margin type: %s", marginType),
			0,
			nil,
		)
	}
	
	// Check if there are open positions for this symbol
	hasPositions, err := s.HasOpenPositions(symbol)
	if err != nil {
		return fmt.Errorf("failed to check positions: %w", err)
	}
	
	if hasPositions {
		return errors.NewTradingError(
			errors.ErrMarginModeConflict,
			fmt.Sprintf("cannot change margin type for %s: open positions exist", symbol),
			0,
			nil,
		)
	}
	
	s.logger.Info("Setting margin type", map[string]interface{}{
		"symbol":      symbol,
		"margin_type": marginType,
	})
	
	err = s.client.SetMarginType(symbol, marginType)
	if err != nil {
		s.logger.Error("Failed to set margin type", map[string]interface{}{
			"symbol":      symbol,
			"margin_type": marginType,
			"error":       err.Error(),
		})
		return fmt.Errorf("failed to set margin type: %w", err)
	}
	
	s.logger.Info("Margin type set successfully", map[string]interface{}{
		"symbol":      symbol,
		"margin_type": marginType,
	})
	
	return nil
}

// SetPositionMode sets position mode (dual side or one way) with position checking
func (s *futuresLeverageService) SetPositionMode(dualSidePosition bool) error {
	// Check if there are any open positions
	hasPositions, err := s.HasAnyOpenPositions()
	if err != nil {
		return fmt.Errorf("failed to check positions: %w", err)
	}
	
	if hasPositions {
		return errors.NewTradingError(
			errors.ErrPositionModeConflict,
			"cannot change position mode: open positions exist",
			0,
			nil,
		)
	}
	
	mode := "one-way"
	if dualSidePosition {
		mode = "dual-side"
	}
	
	s.logger.Info("Setting position mode", map[string]interface{}{
		"mode": mode,
	})
	
	err = s.client.SetPositionMode(dualSidePosition)
	if err != nil {
		s.logger.Error("Failed to set position mode", map[string]interface{}{
			"mode":  mode,
			"error": err.Error(),
		})
		return fmt.Errorf("failed to set position mode: %w", err)
	}
	
	s.logger.Info("Position mode set successfully", map[string]interface{}{
		"mode": mode,
	})
	
	return nil
}

// GetPositionMode retrieves current position mode
func (s *futuresLeverageService) GetPositionMode() (*api.PositionMode, error) {
	mode, err := s.client.GetPositionMode()
	if err != nil {
		s.logger.Error("Failed to get position mode", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to get position mode: %w", err)
	}
	
	modeStr := "one-way"
	if mode.DualSidePosition {
		modeStr = "dual-side"
	}
	
	s.logger.Debug("Retrieved position mode", map[string]interface{}{
		"mode": modeStr,
	})
	
	return mode, nil
}

// HasOpenPositions checks if there are open positions for a specific symbol
func (s *futuresLeverageService) HasOpenPositions(symbol string) (bool, error) {
	if symbol == "" {
		return false, errors.NewTradingError(
			errors.ErrInvalidParameter,
			"symbol cannot be empty",
			0,
			nil,
		)
	}
	
	positions, err := s.client.GetPositions(symbol)
	if err != nil {
		return false, fmt.Errorf("failed to get positions: %w", err)
	}
	
	// Check if any position has non-zero amount
	for _, pos := range positions {
		if pos.PositionAmt != 0 {
			return true, nil
		}
	}
	
	return false, nil
}

// HasAnyOpenPositions checks if there are any open positions across all symbols
func (s *futuresLeverageService) HasAnyOpenPositions() (bool, error) {
	positions, err := s.client.GetAllPositions()
	if err != nil {
		return false, fmt.Errorf("failed to get all positions: %w", err)
	}
	
	// Check if any position has non-zero amount
	for _, pos := range positions {
		if pos.PositionAmt != 0 {
			return true, nil
		}
	}
	
	return false, nil
}
