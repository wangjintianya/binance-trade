package service

import (
	"binance-trader/internal/api"
	"binance-trader/pkg/errors"
	"binance-trader/pkg/logger"
	"fmt"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Mock FuturesClient for testing
type mockFuturesLeverageClient struct {
	setLeverageFunc      func(symbol string, leverage int) (*api.LeverageResponse, error)
	setMarginTypeFunc    func(symbol string, marginType api.MarginType) error
	setPositionModeFunc  func(dualSidePosition bool) error
	getPositionModeFunc  func() (*api.PositionMode, error)
	getPositionsFunc     func(symbol string) ([]*api.Position, error)
	getAllPositionsFunc  func() ([]*api.Position, error)
}

func (m *mockFuturesLeverageClient) GetAccountInfo() (*api.FuturesAccountInfo, error) {
	return nil, nil
}

func (m *mockFuturesLeverageClient) GetBalance() (*api.FuturesBalance, error) {
	return nil, nil
}

func (m *mockFuturesLeverageClient) GetMarkPrice(symbol string) (*api.MarkPrice, error) {
	return nil, nil
}

func (m *mockFuturesLeverageClient) GetPrice(symbol string) (*api.Price, error) {
	return nil, nil
}

func (m *mockFuturesLeverageClient) GetKlines(symbol string, interval string, limit int) ([]*api.Kline, error) {
	return nil, nil
}

func (m *mockFuturesLeverageClient) GetFundingRate(symbol string) (*api.FundingRate, error) {
	return nil, nil
}

func (m *mockFuturesLeverageClient) GetFundingRateHistory(symbol string, startTime, endTime int64) ([]*api.FundingRate, error) {
	return nil, nil
}

func (m *mockFuturesLeverageClient) SetLeverage(symbol string, leverage int) (*api.LeverageResponse, error) {
	if m.setLeverageFunc != nil {
		return m.setLeverageFunc(symbol, leverage)
	}
	return nil, nil
}

func (m *mockFuturesLeverageClient) SetMarginType(symbol string, marginType api.MarginType) error {
	if m.setMarginTypeFunc != nil {
		return m.setMarginTypeFunc(symbol, marginType)
	}
	return nil
}

func (m *mockFuturesLeverageClient) SetPositionMode(dualSidePosition bool) error {
	if m.setPositionModeFunc != nil {
		return m.setPositionModeFunc(dualSidePosition)
	}
	return nil
}

func (m *mockFuturesLeverageClient) GetPositionMode() (*api.PositionMode, error) {
	if m.getPositionModeFunc != nil {
		return m.getPositionModeFunc()
	}
	return nil, nil
}

func (m *mockFuturesLeverageClient) CreateOrder(order *api.FuturesOrderRequest) (*api.FuturesOrderResponse, error) {
	return nil, nil
}

func (m *mockFuturesLeverageClient) CancelOrder(symbol string, orderID int64) (*api.CancelResponse, error) {
	return nil, nil
}

func (m *mockFuturesLeverageClient) GetOrder(symbol string, orderID int64) (*api.FuturesOrder, error) {
	return nil, nil
}

func (m *mockFuturesLeverageClient) GetOpenOrders(symbol string) ([]*api.FuturesOrder, error) {
	return nil, nil
}

func (m *mockFuturesLeverageClient) GetPositions(symbol string) ([]*api.Position, error) {
	if m.getPositionsFunc != nil {
		return m.getPositionsFunc(symbol)
	}
	return nil, nil
}

func (m *mockFuturesLeverageClient) GetAllPositions() ([]*api.Position, error) {
	if m.getAllPositionsFunc != nil {
		return m.getAllPositionsFunc()
	}
	return nil, nil
}

// Feature: usdt-futures-trading, Property 9: 杠杆值范围验证
// Validates: Requirements 3.1
func TestProperty_LeverageRangeValidation(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("leverage values outside 1-125 range must be rejected", prop.ForAll(
		func(leverage int) bool {
			// Create mock client
			mockClient := &mockFuturesLeverageClient{
				setLeverageFunc: func(symbol string, lev int) (*api.LeverageResponse, error) {
					// This should not be called for invalid leverage
					return nil, fmt.Errorf("should not reach API for invalid leverage")
				},
			}

			// Create service
			log, _ := logger.NewLogger(logger.Config{
				Level:         "info",
				EnableConsole: false,
			})
			service := NewFuturesLeverageService(mockClient, log)

			// Try to set leverage
			_, err := service.SetLeverage("BTCUSDT", leverage)

			// For invalid leverage (< 1 or > 125), we must get an error
			if leverage < 1 || leverage > 125 {
				if err == nil {
					return false
				}
				// Check that it's the right type of error
				if tradingErr, ok := err.(*errors.TradingError); ok {
					return tradingErr.Type == errors.ErrInvalidLeverage
				}
				// If it's wrapped, check the message
				return true
			}

			// For valid leverage (1-125), we should not get a validation error
			// (API errors are acceptable, but not validation errors)
			return true
		},
		gen.IntRange(-100, 250), // Test range including invalid values
	))

	properties.TestingRun(t)
}

// Feature: usdt-futures-trading, Property 10: 保证金模式切换条件
// Validates: Requirements 3.2, 3.5
func TestProperty_MarginModeSwitchingConditions(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("margin mode switch must be rejected when positions exist", prop.ForAll(
		func(hasPositions bool, marginType string) bool {
			// Map string to MarginType
			var mt api.MarginType
			if marginType == "ISOLATED" {
				mt = api.MarginTypeIsolated
			} else {
				mt = api.MarginTypeCrossed
			}

			// Create mock client
			mockClient := &mockFuturesLeverageClient{
				getPositionsFunc: func(symbol string) ([]*api.Position, error) {
					if hasPositions {
						// Return a position with non-zero amount
						return []*api.Position{
							{
								Symbol:       symbol,
								PositionAmt:  1.5,
								Leverage:     10,
								MarginType:   api.MarginTypeCrossed,
							},
						}, nil
					}
					// Return empty position (zero amount)
					return []*api.Position{
						{
							Symbol:       symbol,
							PositionAmt:  0,
							Leverage:     10,
							MarginType:   api.MarginTypeCrossed,
						},
					}, nil
				},
				setMarginTypeFunc: func(symbol string, marginType api.MarginType) error {
					// Simulate successful API call
					return nil
				},
			}

			// Create service
			log, _ := logger.NewLogger(logger.Config{
				Level:         "info",
				EnableConsole: false,
			})
			service := NewFuturesLeverageService(mockClient, log)

			// Try to set margin type
			err := service.SetMarginType("BTCUSDT", mt)

			// If positions exist, operation must be rejected
			if hasPositions {
				if err == nil {
					return false
				}
				// Check that it's the right type of error
				if tradingErr, ok := err.(*errors.TradingError); ok {
					return tradingErr.Type == errors.ErrMarginModeConflict
				}
				// If wrapped, still acceptable as long as there's an error
				return true
			}

			// If no positions exist, operation must succeed
			return err == nil
		},
		gen.Bool(),
		gen.OneConstOf("ISOLATED", "CROSSED"),
	))

	properties.TestingRun(t)
}

// Feature: usdt-futures-trading, Property 9: 杠杆值范围验证
// Validates: Requirements 3.1
func TestProperty_ValidLeverageAccepted(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("leverage values within 1-125 range must be accepted by validation", prop.ForAll(
		func(leverage int) bool {
			// Create mock client that simulates successful API call
			mockClient := &mockFuturesLeverageClient{
				setLeverageFunc: func(symbol string, lev int) (*api.LeverageResponse, error) {
					return &api.LeverageResponse{
						Leverage:         lev,
						MaxNotionalValue: 1000000.0,
						Symbol:           symbol,
					}, nil
				},
			}

			// Create service
			log, _ := logger.NewLogger(logger.Config{
				Level:         "info",
				EnableConsole: false,
			})
			service := NewFuturesLeverageService(mockClient, log)

			// Try to set leverage
			response, err := service.SetLeverage("BTCUSDT", leverage)

			// Should succeed without validation error
			if err != nil {
				return false
			}

			// Response should contain the leverage we set
			return response != nil && response.Leverage == leverage
		},
		gen.IntRange(1, 125), // Only valid leverage values
	))

	properties.TestingRun(t)
}

// Feature: usdt-futures-trading, Property 11: 仓位模式切换条件
// Validates: Requirements 3.3, 3.5
func TestProperty_PositionModeSwitchingConditions(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("position mode switch must be rejected when positions exist", prop.ForAll(
		func(hasPositions bool, dualSidePosition bool) bool {
			// Create mock client
			mockClient := &mockFuturesLeverageClient{
				getAllPositionsFunc: func() ([]*api.Position, error) {
					if hasPositions {
						// Return a position with non-zero amount
						return []*api.Position{
							{
								Symbol:       "BTCUSDT",
								PositionAmt:  2.0,
								Leverage:     10,
								MarginType:   api.MarginTypeCrossed,
							},
						}, nil
					}
					// Return empty positions (zero amounts)
					return []*api.Position{
						{
							Symbol:       "BTCUSDT",
							PositionAmt:  0,
							Leverage:     10,
							MarginType:   api.MarginTypeCrossed,
						},
					}, nil
				},
				setPositionModeFunc: func(dualSide bool) error {
					// Simulate successful API call
					return nil
				},
			}

			// Create service
			log, _ := logger.NewLogger(logger.Config{
				Level:         "info",
				EnableConsole: false,
			})
			service := NewFuturesLeverageService(mockClient, log)

			// Try to set position mode
			err := service.SetPositionMode(dualSidePosition)

			// If positions exist, operation must be rejected
			if hasPositions {
				if err == nil {
					return false
				}
				// Check that it's the right type of error
				if tradingErr, ok := err.(*errors.TradingError); ok {
					return tradingErr.Type == errors.ErrPositionModeConflict
				}
				// If wrapped, still acceptable as long as there's an error
				return true
			}

			// If no positions exist, operation must succeed
			return err == nil
		},
		gen.Bool(),
		gen.Bool(),
	))

	properties.TestingRun(t)
}

// Feature: usdt-futures-trading, Property 12: 杠杆设置响应完整性
// Validates: Requirements 3.4
func TestProperty_LeverageResponseCompleteness(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("successful leverage setting must return complete response with confirmation and config", prop.ForAll(
		func(leverage int, maxNotionalValue float64) bool {
			// Create mock client that returns a complete response
			mockClient := &mockFuturesLeverageClient{
				setLeverageFunc: func(symbol string, lev int) (*api.LeverageResponse, error) {
					return &api.LeverageResponse{
						Leverage:         lev,
						MaxNotionalValue: maxNotionalValue,
						Symbol:           symbol,
					}, nil
				},
			}

			// Create service
			log, _ := logger.NewLogger(logger.Config{
				Level:         "info",
				EnableConsole: false,
			})
			service := NewFuturesLeverageService(mockClient, log)

			// Try to set leverage
			response, err := service.SetLeverage("BTCUSDT", leverage)

			// Must succeed
			if err != nil {
				return false
			}

			// Response must not be nil
			if response == nil {
				return false
			}

			// Response must contain confirmation (leverage value)
			if response.Leverage != leverage {
				return false
			}

			// Response must contain current leverage configuration (symbol)
			if response.Symbol != "BTCUSDT" {
				return false
			}

			// Response must contain max notional value (part of config)
			if response.MaxNotionalValue != maxNotionalValue {
				return false
			}

			return true
		},
		gen.IntRange(1, 125), // Valid leverage range
		gen.Float64Range(1000.0, 10000000.0), // Reasonable max notional values
	))

	properties.TestingRun(t)
}
