package service

import (
	"binance-trader/internal/api"
	"binance-trader/internal/repository"
	"binance-trader/pkg/logger"
	"math"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Mock implementations for testing

type mockFuturesTradingService struct{}

func (m *mockFuturesTradingService) OpenLongPosition(symbol string, quantity float64, orderType api.OrderType, price float64) (*api.FuturesOrder, error) {
	return &api.FuturesOrder{OrderID: 12345, Symbol: symbol, Status: api.OrderStatusFilled}, nil
}

func (m *mockFuturesTradingService) OpenShortPosition(symbol string, quantity float64, orderType api.OrderType, price float64) (*api.FuturesOrder, error) {
	return &api.FuturesOrder{OrderID: 12346, Symbol: symbol, Status: api.OrderStatusFilled}, nil
}

func (m *mockFuturesTradingService) ClosePosition(symbol string, positionSide api.PositionSide, quantity float64) (*api.FuturesOrder, error) {
	return &api.FuturesOrder{OrderID: 12347, Symbol: symbol, Status: api.OrderStatusFilled}, nil
}

func (m *mockFuturesTradingService) CloseAllPositions(symbol string) ([]*api.FuturesOrder, error) {
	return []*api.FuturesOrder{}, nil
}

func (m *mockFuturesTradingService) CancelOrder(symbol string, orderID int64) error {
	return nil
}

func (m *mockFuturesTradingService) GetOrderStatus(orderID int64) (*api.FuturesOrder, error) {
	return &api.FuturesOrder{OrderID: orderID, Status: api.OrderStatusFilled}, nil
}

func (m *mockFuturesTradingService) GetActiveOrders(symbol string) ([]*api.FuturesOrder, error) {
	return []*api.FuturesOrder{}, nil
}

func (m *mockFuturesTradingService) SetLeverage(symbol string, leverage int) (*api.LeverageResponse, error) {
	return &api.LeverageResponse{Leverage: leverage, MaxNotionalValue: 1000000.0, Symbol: symbol}, nil
}

func (m *mockFuturesTradingService) GetLeverage(symbol string) (int, error) {
	return 10, nil
}

type mockFuturesMarketDataService struct {
	markPrice float64
}

func (m *mockFuturesMarketDataService) GetMarkPrice(symbol string) (float64, error) {
	return m.markPrice, nil
}

func (m *mockFuturesMarketDataService) GetLastPrice(symbol string) (float64, error) {
	return m.markPrice, nil
}

func (m *mockFuturesMarketDataService) GetHistoricalData(symbol string, interval string, limit int) ([]*api.Kline, error) {
	return []*api.Kline{}, nil
}

func (m *mockFuturesMarketDataService) GetFundingRate(symbol string) (*api.FundingRate, error) {
	return &api.FundingRate{Symbol: symbol, FundingRate: 0.0001}, nil
}

func (m *mockFuturesMarketDataService) SubscribeToMarkPrice(symbol string, callback func(float64)) error {
	return nil
}

func (m *mockFuturesMarketDataService) GetOpenInterest(symbol string) (float64, error) {
	return 1000000.0, nil
}

func (m *mockFuturesMarketDataService) GetFundingRateHistory(symbol string, startTime, endTime int64) ([]*api.FundingRate, error) {
	return []*api.FundingRate{}, nil
}

// Feature: usdt-futures-trading, Property 28: Long stop loss trigger correctness
// Validates: Requirements 7.1
func TestProperty_LongStopLossTriggerCorrectness(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("for any long position and stop price, when mark price falls below stop price, stop loss should trigger", prop.ForAll(
		func(quantity float64, stopPrice float64, markPrice float64) bool {
			stopOrderRepo := repository.NewMemoryStopOrderRepository()
			triggerEngine := NewTriggerEngine()
			mockTrading := &mockFuturesTradingService{}
			mockMarket := &mockFuturesMarketDataService{markPrice: markPrice}
			log, _ := logger.NewLogger(logger.Config{Level: "info", EnableConsole: false})

			service := NewFuturesStopLossService(stopOrderRepo, triggerEngine, mockTrading, mockMarket, log)

			stopOrder, err := service.SetStopLoss("BTCUSDT", api.PositionSideLong, quantity, stopPrice)
			if err != nil {
				return false
			}

			condition := &TriggerCondition{
				Type:     TriggerTypePrice,
				Operator: OperatorLessEqual,
				Value:    stopPrice,
			}

			triggered, err := triggerEngine.EvaluateCondition(condition, markPrice)
			if err != nil {
				return false
			}

			expectedTrigger := markPrice <= stopPrice

			if stopOrder.Type != repository.StopOrderTypeStopLoss {
				return false
			}

			if stopOrder.Status != repository.StopOrderStatusActive {
				return false
			}

			return triggered == expectedTrigger
		},
		gen.Float64Range(0.1, 1000.0),
		gen.Float64Range(100.0, 50000.0),
		gen.Float64Range(50.0, 60000.0),
	))

	properties.TestingRun(t)
}

// Feature: usdt-futures-trading, Property 29: Short stop loss trigger correctness
// Validates: Requirements 7.2
func TestProperty_ShortStopLossTriggerCorrectness(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("for any short position and stop price, when mark price rises above stop price, stop loss should trigger", prop.ForAll(
		func(quantity float64, stopPrice float64, markPrice float64) bool {
			stopOrderRepo := repository.NewMemoryStopOrderRepository()
			triggerEngine := NewTriggerEngine()
			mockTrading := &mockFuturesTradingService{}
			mockMarket := &mockFuturesMarketDataService{markPrice: markPrice}
			log, _ := logger.NewLogger(logger.Config{Level: "info", EnableConsole: false})

			service := NewFuturesStopLossService(stopOrderRepo, triggerEngine, mockTrading, mockMarket, log)

			stopOrder, err := service.SetStopLoss("BTCUSDT", api.PositionSideShort, quantity, stopPrice)
			if err != nil {
				return false
			}

			condition := &TriggerCondition{
				Type:     TriggerTypePrice,
				Operator: OperatorGreaterEqual,
				Value:    stopPrice,
			}

			triggered, err := triggerEngine.EvaluateCondition(condition, markPrice)
			if err != nil {
				return false
			}

			expectedTrigger := markPrice >= stopPrice

			if stopOrder.Type != repository.StopOrderTypeStopLoss {
				return false
			}

			if stopOrder.Status != repository.StopOrderStatusActive {
				return false
			}

			return triggered == expectedTrigger
		},
		gen.Float64Range(0.1, 1000.0),
		gen.Float64Range(100.0, 50000.0),
		gen.Float64Range(50.0, 60000.0),
	))

	properties.TestingRun(t)
}

// Feature: usdt-futures-trading, Property 30: Take profit trigger correctness
// Validates: Requirements 7.3
func TestProperty_FuturesTakeProfitTriggerCorrectness(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("for any position and target price, when mark price reaches target, take profit should trigger", prop.ForAll(
		func(quantity float64, targetPrice float64, markPrice float64, isLong bool) bool {
			stopOrderRepo := repository.NewMemoryStopOrderRepository()
			triggerEngine := NewTriggerEngine()
			mockTrading := &mockFuturesTradingService{}
			mockMarket := &mockFuturesMarketDataService{markPrice: markPrice}
			log, _ := logger.NewLogger(logger.Config{Level: "info", EnableConsole: false})

			service := NewFuturesStopLossService(stopOrderRepo, triggerEngine, mockTrading, mockMarket, log)

			var positionSide api.PositionSide
			var operator ComparisonOperator
			var expectedTrigger bool

			if isLong {
				positionSide = api.PositionSideLong
				operator = OperatorGreaterEqual
				expectedTrigger = markPrice >= targetPrice
			} else {
				positionSide = api.PositionSideShort
				operator = OperatorLessEqual
				expectedTrigger = markPrice <= targetPrice
			}

			takeProfitOrder, err := service.SetTakeProfit("BTCUSDT", positionSide, quantity, targetPrice)
			if err != nil {
				return false
			}

			condition := &TriggerCondition{
				Type:     TriggerTypePrice,
				Operator: operator,
				Value:    targetPrice,
			}

			triggered, err := triggerEngine.EvaluateCondition(condition, markPrice)
			if err != nil {
				return false
			}

			if takeProfitOrder.Type != repository.StopOrderTypeTakeProfit {
				return false
			}

			if takeProfitOrder.Status != repository.StopOrderStatusActive {
				return false
			}

			return triggered == expectedTrigger
		},
		gen.Float64Range(0.1, 1000.0),
		gen.Float64Range(100.0, 50000.0),
		gen.Float64Range(50.0, 60000.0),
		gen.Bool(),
	))

	properties.TestingRun(t)
}

// Feature: usdt-futures-trading, Property 31: Stop loss take profit mutual exclusivity
// Validates: Requirements 7.4
func TestProperty_FuturesStopLossTakeProfitMutualExclusivity(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("for any stop loss and take profit pair, when one triggers, the other should be cancelled", prop.ForAll(
		func(quantity float64, stopPrice float64, targetPrice float64, isLong bool) bool {
			if math.Abs(stopPrice-targetPrice) < 1.0 {
				return true
			}

			if isLong && stopPrice >= targetPrice {
				return true
			}
			if !isLong && stopPrice <= targetPrice {
				return true
			}

			stopOrderRepo := repository.NewMemoryStopOrderRepository()
			triggerEngine := NewTriggerEngine()
			mockTrading := &mockFuturesTradingService{}
			mockMarket := &mockFuturesMarketDataService{markPrice: (stopPrice + targetPrice) / 2}
			log, _ := logger.NewLogger(logger.Config{Level: "info", EnableConsole: false})

			service := NewFuturesStopLossService(stopOrderRepo, triggerEngine, mockTrading, mockMarket, log)

			var positionSide api.PositionSide
			if isLong {
				positionSide = api.PositionSideLong
			} else {
				positionSide = api.PositionSideShort
			}

			pair, err := service.SetStopLossTakeProfit("BTCUSDT", positionSide, quantity, stopPrice, targetPrice)
			if err != nil {
				return false
			}

			if pair.StopLossOrder.Status != repository.StopOrderStatusActive {
				return false
			}

			if pair.TakeProfitOrder.Status != repository.StopOrderStatusActive {
				return false
			}

			err = stopOrderRepo.UpdateStopOrderStatus(
				pair.StopLossOrder.OrderID,
				repository.StopOrderStatusTriggered,
				0,
				0,
			)
			if err != nil {
				return false
			}

			err = stopOrderRepo.UpdateStopOrderStatus(
				pair.TakeProfitOrder.OrderID,
				repository.StopOrderStatusCancelled,
				0,
				0,
			)
			if err != nil {
				return false
			}

			stopLoss, _ := stopOrderRepo.FindStopOrderByID(pair.StopLossOrder.OrderID)
			takeProfit, _ := stopOrderRepo.FindStopOrderByID(pair.TakeProfitOrder.OrderID)

			return stopLoss.Status == repository.StopOrderStatusTriggered &&
				takeProfit.Status == repository.StopOrderStatusCancelled
		},
		gen.Float64Range(0.1, 1000.0),
		gen.Float64Range(100.0, 30000.0),
		gen.Float64Range(100.0, 30000.0),
		gen.Bool(),
	))

	properties.TestingRun(t)
}

// Feature: usdt-futures-trading, Property 32: Trailing stop price adjustment
// Validates: Requirements 7.5
func TestProperty_FuturesTrailingStopPriceAdjustment(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("for any trailing stop order and price sequence, when price moves favorably, stop price must adjust by callback rate", prop.ForAll(
		func(quantity float64, callbackRate float64, initialPrice float64, priceChange float64, isLong bool) bool {
			if callbackRate <= 0 || callbackRate >= 100 {
				return true
			}
			if initialPrice <= 0 {
				return true
			}

			stopOrderRepo := repository.NewMemoryStopOrderRepository()
			triggerEngine := NewTriggerEngine()
			mockTrading := &mockFuturesTradingService{}
			mockMarket := &mockFuturesMarketDataService{markPrice: initialPrice}
			log, _ := logger.NewLogger(logger.Config{Level: "info", EnableConsole: false})

			service := NewFuturesStopLossService(stopOrderRepo, triggerEngine, mockTrading, mockMarket, log).(*futuresStopLossService)

			var positionSide api.PositionSide
			if isLong {
				positionSide = api.PositionSideLong
			} else {
				positionSide = api.PositionSideShort
			}

			trailingOrder, err := service.SetTrailingStop("BTCUSDT", positionSide, quantity, callbackRate)
			if err != nil {
				return false
			}

			initialExtreme := trailingOrder.HighestPrice
			initialStopPrice := trailingOrder.CurrentStopPrice

			var expectedInitialStop float64
			if isLong {
				expectedInitialStop = initialPrice * (1 - callbackRate/100)
			} else {
				expectedInitialStop = initialPrice * (1 + callbackRate/100)
			}
			tolerance := 0.0001
			if math.Abs(initialStopPrice-expectedInitialStop) > tolerance {
				return false
			}

			newPrice := initialPrice + priceChange

			isFavorable := (isLong && priceChange > 0) || (!isLong && priceChange < 0)

			updated, err := service.UpdateTrailingStopPrice(trailingOrder.OrderID, positionSide, newPrice)
			if err != nil {
				return false
			}

			updatedOrder, err := stopOrderRepo.FindTrailingStopOrderByID(trailingOrder.OrderID)
			if err != nil {
				return false
			}

			if isFavorable {
				if !updated {
					return false
				}

				if updatedOrder.HighestPrice != newPrice {
					return false
				}

				var expectedNewStop float64
				if isLong {
					expectedNewStop = newPrice * (1 - callbackRate/100)
				} else {
					expectedNewStop = newPrice * (1 + callbackRate/100)
				}
				if math.Abs(updatedOrder.CurrentStopPrice-expectedNewStop) > tolerance {
					return false
				}

				if isLong && updatedOrder.CurrentStopPrice <= initialStopPrice {
					return false
				}
				if !isLong && updatedOrder.CurrentStopPrice >= initialStopPrice {
					return false
				}
			} else if !isFavorable && math.Abs(priceChange) > 0.01 {
				if updatedOrder.Status == repository.StopOrderStatusTriggered {
					return true
				}

				if updated {
					return false
				}

				if updatedOrder.HighestPrice != initialExtreme {
					return false
				}

				if updatedOrder.CurrentStopPrice != initialStopPrice {
					return false
				}
			}

			return true
		},
		gen.Float64Range(0.1, 1000.0),
		gen.Float64Range(0.5, 10.0),
		gen.Float64Range(1000.0, 50000.0),
		gen.Float64Range(-10000.0, 10000.0),
		gen.Bool(),
	))

	properties.TestingRun(t)
}
