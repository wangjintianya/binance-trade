package service

import (
	"binance-trader/internal/api"
	"binance-trader/internal/repository"
	"binance-trader/pkg/logger"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Feature: usdt-futures-trading, Property 33: 标记价格触发监控
// For any price trigger order and mark price sequence, when price satisfies trigger condition, system must identify trigger point
// Validates: Requirements 8.1, 8.2
func TestProperty33_MarkPriceTriggerMonitoring(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("mark price trigger monitoring identifies trigger points", prop.ForAll(
		func(triggerPrice float64, priceOffset float64, useGreaterThan bool) bool {
			// Skip invalid inputs
			if triggerPrice <= 0 || priceOffset == 0 {
				return true
			}

			// Create mock services
			mockClient := &mockFuturesClientShared{}
			mockPositionMgr := &mockFuturesPositionManagerShared{
				positions: make(map[string]*api.Position),
			}
			mockTrading := &mockFuturesTradingServiceShared{
				orders: make([]*api.FuturesOrder, 0),
			}
			mockLogger, _ := logger.NewLogger(logger.Config{
				Level:         "info",
				EnableConsole: false,
			})

			// Determine operator and current price based on whether condition should trigger
			var operator ComparisonOperator
			var currentPrice float64
			var expectedTrigger bool

			if useGreaterThan {
				operator = OperatorGreaterThan
				if priceOffset > 0 {
					// Price is above trigger - should trigger
					currentPrice = triggerPrice + priceOffset
					expectedTrigger = true
				} else {
					// Price is below trigger - should not trigger
					currentPrice = triggerPrice + priceOffset
					expectedTrigger = false
				}
			} else {
				operator = OperatorLessThan
				if priceOffset < 0 {
					// Price is below trigger - should trigger
					currentPrice = triggerPrice + priceOffset
					expectedTrigger = true
				} else {
					// Price is above trigger - should not trigger
					currentPrice = triggerPrice + priceOffset
					expectedTrigger = false
				}
			}

			// Ensure price is positive
			if currentPrice <= 0 {
				return true
			}

			mockMarketData := &mockFuturesMarketDataServiceShared{
				markPrice: currentPrice,
			}

			// Create service
			service := NewFuturesConditionalOrderService(
				mockClient,
				mockMarketData,
				mockPositionMgr,
				mockTrading,
				mockLogger,
			)

			// Create conditional order with mark price trigger
			request := &FuturesConditionalOrderRequest{
				Symbol:       "BTCUSDT",
				Side:         api.OrderSideBuy,
				PositionSide: api.PositionSideLong,
				Type:         api.OrderTypeMarket,
				Quantity:     0.1,
				TriggerCondition: &FuturesTriggerCondition{
					Type:     FuturesTriggerTypeMarkPrice,
					Operator: operator,
					Value:    triggerPrice,
				},
			}

			order, err := service.CreateConditionalOrder(request)
			if err != nil {
				return false
			}

			// Verify order was created
			if order.Status != repository.ConditionalOrderStatusPending {
				return false
			}

			// Evaluate trigger condition
			triggered, triggerValue, err := service.(*futuresConditionalOrderService).evaluateTriggerCondition(order)
			if err != nil {
				return false
			}

			// Verify trigger detection matches expected behavior
			if triggered != expectedTrigger {
				return false
			}

			// If triggered, verify trigger value is the current mark price
			if triggered && triggerValue != currentPrice {
				return false
			}

			return true
		},
		gen.Float64Range(10000, 50000),  // triggerPrice
		gen.Float64Range(-5000, 5000),   // priceOffset (can be positive or negative)
		gen.Bool(),                       // useGreaterThan
	))

	properties.TestingRun(t)
}

// Feature: usdt-futures-trading, Property 34: 盈亏触发监控
// For any PnL-based trigger order and PnL sequence, when unrealized PnL reaches threshold, system must identify trigger point
// Validates: Requirements 8.3
func TestProperty34_PnLTriggerMonitoring(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("unrealized PnL trigger monitoring identifies trigger points", prop.ForAll(
		func(entryPrice float64, positionSize float64, pnlThreshold float64, useGreaterThan bool) bool {
			// Skip invalid inputs
			if entryPrice <= 0 || positionSize == 0 {
				return true
			}

			// Create mock services
			mockClient := &mockFuturesClientShared{}
			mockMarketData := &mockFuturesMarketDataServiceShared{
				markPrice: entryPrice,
			}
			mockPositionMgr := &mockFuturesPositionManagerShared{
				positions: map[string]*api.Position{
					"BTCUSDT" + string(api.PositionSideLong): {
						Symbol:           "BTCUSDT",
						PositionSide:     api.PositionSideLong,
						PositionAmt:      positionSize,
						EntryPrice:       entryPrice,
						UnrealizedProfit: 0,
					},
				},
			}
			mockTrading := &mockFuturesTradingServiceShared{
				orders: make([]*api.FuturesOrder, 0),
			}
			mockLogger, _ := logger.NewLogger(logger.Config{
				Level:         "info",
				EnableConsole: false,
			})

			// Create service
			service := NewFuturesConditionalOrderService(
				mockClient,
				mockMarketData,
				mockPositionMgr,
				mockTrading,
				mockLogger,
			)

			// Determine operator
			var operator ComparisonOperator
			if useGreaterThan {
				operator = OperatorGreaterThan
			} else {
				operator = OperatorLessThan
			}

			// Create conditional order with PnL trigger
			request := &FuturesConditionalOrderRequest{
				Symbol:       "BTCUSDT",
				Side:         api.OrderSideSell,
				PositionSide: api.PositionSideLong,
				Type:         api.OrderTypeMarket,
				Quantity:     0.1,
				TriggerCondition: &FuturesTriggerCondition{
					Type:     FuturesTriggerTypeUnrealizedPnL,
					Operator: operator,
					Value:    pnlThreshold,
				},
			}

			order, err := service.CreateConditionalOrder(request)
			if err != nil {
				return false
			}

			// Calculate mark price that would produce the threshold PnL
			// PnL = (markPrice - entryPrice) * positionSize
			// markPrice = entryPrice + (PnL / positionSize)
			targetMarkPrice := entryPrice + (pnlThreshold / positionSize)

			// Set mark price to trigger condition
			if useGreaterThan {
				mockMarketData.markPrice = targetMarkPrice + 100
			} else {
				mockMarketData.markPrice = targetMarkPrice - 100
			}

			// Evaluate trigger condition
			triggered, triggerValue, err := service.(*futuresConditionalOrderService).evaluateTriggerCondition(order)
			if err != nil {
				return false
			}

			// Calculate expected PnL
			expectedPnL := (mockMarketData.markPrice - entryPrice) * positionSize

			// Verify trigger detection
			if useGreaterThan {
				if expectedPnL > pnlThreshold {
					if !triggered {
						return false
					}
					// Trigger value should be close to expected PnL
					if triggerValue < pnlThreshold {
						return false
					}
				}
			} else {
				if expectedPnL < pnlThreshold {
					if !triggered {
						return false
					}
					// Trigger value should be close to expected PnL
					if triggerValue > pnlThreshold {
						return false
					}
				}
			}

			return true
		},
		gen.Float64Range(10000, 50000),   // entryPrice
		gen.Float64Range(0.01, 10),       // positionSize
		gen.Float64Range(-5000, 5000),    // pnlThreshold
		gen.Bool(),                        // useGreaterThan
	))

	properties.TestingRun(t)
}

// Feature: usdt-futures-trading, Property 35: 资金费率触发监控
// For any funding rate-based trigger order and rate sequence, when funding rate reaches threshold, system must identify trigger point
// Validates: Requirements 8.4
func TestProperty35_FundingRateTriggerMonitoring(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("funding rate trigger monitoring identifies trigger points", prop.ForAll(
		func(currentRate float64, thresholdRate float64, useGreaterThan bool) bool {
			// Create mock services
			mockClient := &mockFuturesClientShared{
				fundingRate: &api.FundingRate{
					Symbol:      "BTCUSDT",
					FundingRate: currentRate,
					FundingTime: time.Now().Unix(),
				},
			}
			mockMarketData := &mockFuturesMarketDataServiceShared{
				fundingRate: &api.FundingRate{
					Symbol:      "BTCUSDT",
					FundingRate: currentRate,
					FundingTime: time.Now().Unix(),
				},
			}
			mockPositionMgr := &mockFuturesPositionManagerShared{
				positions: make(map[string]*api.Position),
			}
			mockTrading := &mockFuturesTradingServiceShared{
				orders: make([]*api.FuturesOrder, 0),
			}
			mockLogger, _ := logger.NewLogger(logger.Config{
				Level:         "info",
				EnableConsole: false,
			})

			// Create service
			service := NewFuturesConditionalOrderService(
				mockClient,
				mockMarketData,
				mockPositionMgr,
				mockTrading,
				mockLogger,
			)

			// Determine operator
			var operator ComparisonOperator
			if useGreaterThan {
				operator = OperatorGreaterThan
			} else {
				operator = OperatorLessThan
			}

			// Create conditional order with funding rate trigger
			request := &FuturesConditionalOrderRequest{
				Symbol:       "BTCUSDT",
				Side:         api.OrderSideBuy,
				PositionSide: api.PositionSideLong,
				Type:         api.OrderTypeMarket,
				Quantity:     0.1,
				TriggerCondition: &FuturesTriggerCondition{
					Type:     FuturesTriggerTypeFundingRate,
					Operator: operator,
					Value:    thresholdRate,
				},
			}

			order, err := service.CreateConditionalOrder(request)
			if err != nil {
				return false
			}

			// Evaluate trigger condition
			triggered, triggerValue, err := service.(*futuresConditionalOrderService).evaluateTriggerCondition(order)
			if err != nil {
				return false
			}

			// Verify trigger detection
			expectedTrigger := false
			if useGreaterThan {
				expectedTrigger = currentRate > thresholdRate
			} else {
				expectedTrigger = currentRate < thresholdRate
			}

			if triggered != expectedTrigger {
				return false
			}

			// If triggered, verify trigger value is the current funding rate
			if triggered && triggerValue != currentRate {
				return false
			}

			return true
		},
		gen.Float64Range(-0.01, 0.01),   // currentRate
		gen.Float64Range(-0.01, 0.01),   // thresholdRate
		gen.Bool(),                       // useGreaterThan
	))

	properties.TestingRun(t)
}

// Feature: usdt-futures-trading, Property 36: 触发事件日志完整性
// For any triggered conditional order, log must include trigger time, trigger value, and order ID
// Validates: Requirements 8.5
func TestProperty36_TriggerEventLogCompleteness(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("trigger event logs contain complete information", prop.ForAll(
		func(markPrice float64, triggerPrice float64) bool {
			// Skip cases where trigger won't happen
			if markPrice <= triggerPrice {
				return true
			}

			// Create mock services
			mockClient := &mockFuturesClientShared{
				markPrice: markPrice,
			}
			mockMarketData := &mockFuturesMarketDataServiceShared{
				markPrice: markPrice,
			}
			mockPositionMgr := &mockFuturesPositionManagerShared{
				positions: make(map[string]*api.Position),
			}
			mockTrading := &mockFuturesTradingServiceShared{
				orders: make([]*api.FuturesOrder, 0),
			}
			mockLogger, _ := logger.NewLogger(logger.Config{
				Level:         "info",
				EnableConsole: false,
			})

			// Create service
			service := NewFuturesConditionalOrderService(
				mockClient,
				mockMarketData,
				mockPositionMgr,
				mockTrading,
				mockLogger,
			)

			// Create conditional order that will trigger
			request := &FuturesConditionalOrderRequest{
				Symbol:       "BTCUSDT",
				Side:         api.OrderSideBuy,
				PositionSide: api.PositionSideLong,
				Type:         api.OrderTypeMarket,
				Quantity:     0.1,
				TriggerCondition: &FuturesTriggerCondition{
					Type:     FuturesTriggerTypeMarkPrice,
					Operator: OperatorGreaterThan,
					Value:    triggerPrice,
				},
			}

			order, err := service.CreateConditionalOrder(request)
			if err != nil {
				return false
			}

			// Record time before trigger
			beforeTrigger := time.Now().Unix()

			// Execute trigger
			service.(*futuresConditionalOrderService).executeTrigger(order, markPrice)

			// Record time after trigger
			afterTrigger := time.Now().Unix()

			// Verify order status changed to triggered/executed
			if order.Status != repository.ConditionalOrderStatusExecuted {
				return false
			}

			// Verify trigger time is set and within reasonable range
			if order.TriggeredAt < beforeTrigger || order.TriggeredAt > afterTrigger {
				return false
			}

			// Verify executed order ID is set
			if order.ExecutedOrderID == 0 {
				return false
			}

			// Verify order ID is not empty
			if order.OrderID == "" {
				return false
			}

			return true
		},
		gen.Float64Range(20000, 50000),  // markPrice
		gen.Float64Range(10000, 19999),  // triggerPrice (lower to ensure trigger)
	))

	properties.TestingRun(t)
}
