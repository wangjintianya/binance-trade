package service

import (
	"binance-trader/internal/api"
	"binance-trader/internal/repository"
	"binance-trader/pkg/errors"
	"binance-trader/pkg/logger"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// symbolGen generates non-empty symbol strings
func symbolGen() gopter.Gen {
	// Generate uppercase alphanumeric strings that look like trading symbols
	return gen.Identifier().Map(func(s string) string {
		// Ensure it's uppercase and between 3-10 characters
		if len(s) < 3 {
			s = s + "BTC"
		}
		if len(s) > 10 {
			s = s[:10]
		}
		return s + "USDT"
	})
}

// Feature: usdt-futures-trading, Property 13: 市价开多单参数正确性
// Validates: Requirements 4.1
func TestProperty13_MarketLongOrderParametersCorrectness(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("market long orders have correct parameters", prop.ForAll(
		func(symbol string, quantity float64) bool {
			mockClient := &mockFuturesClient{
				createOrderFunc: func(req *api.FuturesOrderRequest) (*api.FuturesOrderResponse, error) {
					// Return response with the request parameters
					return &api.FuturesOrderResponse{
						OrderID:      12345,
						Symbol:       req.Symbol,
						Status:       api.OrderStatusNew,
						Side:         req.Side,
						PositionSide: req.PositionSide,
						Type:         req.Type,
						OrigQty:      req.Quantity,
					}, nil
				},
			}

			repo := repository.NewMemoryFuturesOrderRepository()
			log, _ := logger.NewLogger(logger.Config{Level: "info"})
			service := NewFuturesTradingService(mockClient, repo, log)

			order, err := service.OpenLongPosition(symbol, quantity, api.OrderTypeMarket, 0)
			if err != nil {
				return false
			}

			// Verify order parameters - for any market long order, these must be correct
			return order.Type == api.OrderTypeMarket &&
				order.Side == api.OrderSideBuy &&
				order.PositionSide == api.PositionSideLong
		},
		symbolGen(),
		gen.Float64Range(0.001, 1000.0),
	))

	properties.TestingRun(t)
}

// Feature: usdt-futures-trading, Property 14: 限价开空单参数完整性
// Validates: Requirements 4.2
func TestProperty14_LimitShortOrderParametersCompleteness(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("limit short orders have complete parameters", prop.ForAll(
		func(symbol string, quantity float64, price float64) bool {
			mockClient := &mockFuturesClient{
				createOrderFunc: func(req *api.FuturesOrderRequest) (*api.FuturesOrderResponse, error) {
					// Verify parameters
					if req.Type != api.OrderTypeLimit {
						return nil, nil
					}
					if req.Side != api.OrderSideSell {
						return nil, nil
					}
					if req.PositionSide != api.PositionSideShort {
						return nil, nil
					}
					if req.Price <= 0 {
						return nil, nil
					}
					
					return &api.FuturesOrderResponse{
						OrderID:      12345,
						Symbol:       req.Symbol,
						Status:       api.OrderStatusNew,
						Side:         req.Side,
						PositionSide: req.PositionSide,
						Type:         req.Type,
						Price:        req.Price,
						OrigQty:      req.Quantity,
					}, nil
				},
			}

			repo := repository.NewMemoryFuturesOrderRepository()
			log, _ := logger.NewLogger(logger.Config{Level: "info"})
			service := NewFuturesTradingService(mockClient, repo, log)

			order, err := service.OpenShortPosition(symbol, quantity, api.OrderTypeLimit, price)
			if err != nil {
				return false
			}

			// Verify order has all required fields
			return order.Type == api.OrderTypeLimit &&
				order.Side == api.OrderSideSell &&
				order.PositionSide == api.PositionSideShort &&
				order.Price > 0
		},
		symbolGen(),
		gen.Float64Range(0.001, 1000.0),
		gen.Float64Range(1.0, 100000.0),
	))

	properties.TestingRun(t)
}

// Feature: usdt-futures-trading, Property 15: 平仓方向正确性
// Validates: Requirements 4.3
func TestProperty15_ClosePositionDirectionCorrectness(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("close position has opposite direction", prop.ForAll(
		func(symbol string, quantity float64, isLong bool) bool {
			positionSide := api.PositionSideLong
			expectedCloseSide := api.OrderSideSell
			if !isLong {
				positionSide = api.PositionSideShort
				expectedCloseSide = api.OrderSideBuy
			}

			mockClient := &mockFuturesClient{
				createOrderFunc: func(req *api.FuturesOrderRequest) (*api.FuturesOrderResponse, error) {
					// Verify close order has opposite side
					if req.PositionSide == api.PositionSideLong && req.Side != api.OrderSideSell {
						return nil, nil
					}
					if req.PositionSide == api.PositionSideShort && req.Side != api.OrderSideBuy {
						return nil, nil
					}
					
					return &api.FuturesOrderResponse{
						OrderID:      12345,
						Symbol:       req.Symbol,
						Status:       api.OrderStatusNew,
						Side:         req.Side,
						PositionSide: req.PositionSide,
						Type:         req.Type,
						OrigQty:      req.Quantity,
						ReduceOnly:   req.ReduceOnly,
					}, nil
				},
			}

			repo := repository.NewMemoryFuturesOrderRepository()
			log, _ := logger.NewLogger(logger.Config{Level: "info"})
			service := NewFuturesTradingService(mockClient, repo, log)

			order, err := service.ClosePosition(symbol, positionSide, quantity)
			if err != nil {
				return false
			}

			// Verify close order has opposite side
			return order.Side == expectedCloseSide && order.ReduceOnly
		},
		symbolGen(),
		gen.Float64Range(0.001, 1000.0),
		gen.Bool(),
	))

	properties.TestingRun(t)
}

// Feature: usdt-futures-trading, Property 16: 合约订单响应完整性
// Validates: Requirements 4.4
func TestProperty16_FuturesOrderResponseCompleteness(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("order response contains all required fields", prop.ForAll(
		func(symbol string, quantity float64, isLong bool) bool {
			orderType := api.OrderTypeMarket
			side := api.OrderSideBuy
			positionSide := api.PositionSideLong
			if !isLong {
				side = api.OrderSideSell
				positionSide = api.PositionSideShort
			}

			mockClient := &mockFuturesClient{
				createOrderFunc: func(req *api.FuturesOrderRequest) (*api.FuturesOrderResponse, error) {
					return &api.FuturesOrderResponse{
						OrderID:      12345,
						Symbol:       req.Symbol,
						Status:       api.OrderStatusFilled,
						Side:         req.Side,
						PositionSide: req.PositionSide,
						Type:         req.Type,
						Price:        50000.0,
						AvgPrice:     50000.0,
						OrigQty:      req.Quantity,
						ExecutedQty:  req.Quantity,
						UpdateTime:   1234567890,
					}, nil
				},
			}

			repo := repository.NewMemoryFuturesOrderRepository()
			log, _ := logger.NewLogger(logger.Config{Level: "info"})
			service := NewFuturesTradingService(mockClient, repo, log)

			var order *api.FuturesOrder
			var err error
			if isLong {
				order, err = service.OpenLongPosition(symbol, quantity, orderType, 0)
			} else {
				order, err = service.OpenShortPosition(symbol, quantity, orderType, 0)
			}
			
			if err != nil {
				return false
			}

			// Verify response contains all required fields
			return order.OrderID > 0 &&
				order.Status != "" &&
				order.AvgPrice >= 0 &&
				order.Symbol == symbol &&
				order.Side == side &&
				order.PositionSide == positionSide
		},
		symbolGen(),
		gen.Float64Range(0.001, 1000.0),
		gen.Bool(),
	))

	properties.TestingRun(t)
}

// Feature: usdt-futures-trading, Property 17: 保证金充足性检查
// Validates: Requirements 4.5
func TestProperty17_MarginSufficiencyCheck(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("orders rejected when margin insufficient", prop.ForAll(
		func(symbol string, quantity float64, hasMargin bool) bool {
			mockClient := &mockFuturesClient{
				createOrderFunc: func(req *api.FuturesOrderRequest) (*api.FuturesOrderResponse, error) {
					// Simulate margin check
					if !hasMargin {
						// Return error for insufficient margin
						return nil, errors.NewTradingError(
							errors.ErrInsufficientMargin,
							"Margin is insufficient",
							-2019,
							nil,
						)
					}
					
					return &api.FuturesOrderResponse{
						OrderID:      12345,
						Symbol:       req.Symbol,
						Status:       api.OrderStatusNew,
						Side:         req.Side,
						PositionSide: req.PositionSide,
						Type:         req.Type,
						OrigQty:      req.Quantity,
					}, nil
				},
			}

			repo := repository.NewMemoryFuturesOrderRepository()
			log, _ := logger.NewLogger(logger.Config{Level: "info"})
			service := NewFuturesTradingService(mockClient, repo, log)

			_, err := service.OpenLongPosition(symbol, quantity, api.OrderTypeMarket, 0)

			// If no margin, order should fail
			if !hasMargin {
				return err != nil
			}
			
			// If has margin, order should succeed
			return err == nil
		},
		symbolGen(),
		gen.Float64Range(0.001, 1000.0),
		gen.Bool(),
	))

	properties.TestingRun(t)
}
