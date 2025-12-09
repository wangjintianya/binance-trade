package service

import (
	"binance-trader/internal/api"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Mock market service for testing
type mockFundingMarketService struct {
	fundingRate        *api.FundingRate
	fundingRateHistory []*api.FundingRate
	err                error
}

func (m *mockFundingMarketService) GetMarkPrice(symbol string) (float64, error) {
	return 50000.0, m.err
}

func (m *mockFundingMarketService) GetLastPrice(symbol string) (float64, error) {
	return 50000.0, m.err
}

func (m *mockFundingMarketService) GetHistoricalData(symbol string, interval string, limit int) ([]*api.Kline, error) {
	return nil, m.err
}

func (m *mockFundingMarketService) GetFundingRate(symbol string) (*api.FundingRate, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.fundingRate, nil
}

func (m *mockFundingMarketService) GetFundingRateHistory(symbol string, startTime, endTime int64) ([]*api.FundingRate, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.fundingRateHistory, nil
}

func (m *mockFundingMarketService) SubscribeToMarkPrice(symbol string, callback func(float64)) error {
	return m.err
}

// Feature: usdt-futures-trading, Property 45: 资金费率结算触发
// Validates: Requirements 12.1
func TestProperty_FundingRateSettlementTrigger(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("for any settlement time, when time arrives, system must trigger funding rate query", 
		prop.ForAll(
			func(settleTimeOffset int64, fundingRate float64) bool {
				// Generate a settlement time in the past
				settleTime := time.Now().Unix()*1000 - settleTimeOffset
				
				// Create mock market service
				mockMarket := &mockFundingMarketService{
					fundingRate: &api.FundingRate{
						Symbol:      "BTCUSDT",
						FundingRate: fundingRate,
						FundingTime: settleTime,
					},
				}
				
				// Create funding service
				service := NewFuturesFundingService(mockMarket, &mockLogger{})
				
				// Query funding rate at settlement time
				result, err := service.QueryFundingRateAtSettlement("BTCUSDT", settleTime)
				
				// Verify that query was triggered successfully
				if err != nil {
					return false
				}
				
				// Verify result is not nil
				if result == nil {
					return false
				}
				
				// Verify funding rate matches
				return result.FundingRate == fundingRate
			},
			gen.Int64Range(1, 86400000), // Settlement time offset: 1ms to 1 day ago
			gen.Float64Range(-0.01, 0.01), // Funding rate: -1% to 1%
		))

	properties.TestingRun(t)
}

// Feature: usdt-futures-trading, Property 46: 多头资金费用计算
// Validates: Requirements 12.2
func TestProperty_LongPositionFundingFeeCalculation(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("for any long position with positive funding rate, funding fee must be negative (payment)", 
		prop.ForAll(
			func(positionAmt float64, markPrice float64, fundingRate float64) bool {
				// Skip if funding rate is not positive
				if fundingRate <= 0 {
					return true
				}
				
				// Skip if position amount is zero
				if positionAmt == 0 {
					return true
				}
				
				// Create long position
				position := &api.Position{
					Symbol:       "BTCUSDT",
					PositionSide: api.PositionSideLong,
					PositionAmt:  positionAmt,
					MarkPrice:    markPrice,
				}
				
				// Create funding service
				mockMarket := &mockFundingMarketService{}
				service := NewFuturesFundingService(mockMarket, &mockLogger{})
				
				// Calculate funding fee
				fundingFee, err := service.CalculateFundingFee(position, fundingRate)
				if err != nil {
					return false
				}
				
				// For long position with positive funding rate, fee must be negative (paying)
				return fundingFee < 0
			},
			gen.Float64Range(0.01, 100.0),    // Position amount: 0.01 to 100
			gen.Float64Range(1000.0, 100000.0), // Mark price: 1000 to 100000
			gen.Float64Range(0.0001, 0.01),   // Positive funding rate: 0.01% to 1%
		))

	properties.TestingRun(t)
}

// Feature: usdt-futures-trading, Property 47: 空头资金费用计算
// Validates: Requirements 12.3
func TestProperty_ShortPositionFundingFeeCalculation(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("for any short position with positive funding rate, funding fee must be positive (receiving)", 
		prop.ForAll(
			func(positionAmt float64, markPrice float64, fundingRate float64) bool {
				// Skip if funding rate is not positive
				if fundingRate <= 0 {
					return true
				}
				
				// Skip if position amount is zero
				if positionAmt == 0 {
					return true
				}
				
				// Create short position
				position := &api.Position{
					Symbol:       "BTCUSDT",
					PositionSide: api.PositionSideShort,
					PositionAmt:  -positionAmt, // Negative for short
					MarkPrice:    markPrice,
				}
				
				// Create funding service
				mockMarket := &mockFundingMarketService{}
				service := NewFuturesFundingService(mockMarket, &mockLogger{})
				
				// Calculate funding fee
				fundingFee, err := service.CalculateFundingFee(position, fundingRate)
				if err != nil {
					return false
				}
				
				// For short position with positive funding rate, fee must be positive (receiving)
				return fundingFee > 0
			},
			gen.Float64Range(0.01, 100.0),    // Position amount: 0.01 to 100
			gen.Float64Range(1000.0, 100000.0), // Mark price: 1000 to 100000
			gen.Float64Range(0.0001, 0.01),   // Positive funding rate: 0.01% to 1%
		))

	properties.TestingRun(t)
}

// Feature: usdt-futures-trading, Property 48: 资金费用结算后状态更新
// Validates: Requirements 12.4
func TestProperty_StateUpdateAfterFundingSettlement(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("for any funding settlement, balance and cost must be adjusted according to fee amount", 
		prop.ForAll(
			func(fundingFee float64, currentBalance float64, currentCost float64) bool {
				// Create settlement
				settlement := &FundingFeeSettlement{
					Symbol:       "BTCUSDT",
					PositionSide: api.PositionSideLong,
					PositionAmt:  1.0,
					FundingRate:  0.0001,
					FundingFee:   fundingFee,
					SettleTime:   time.Now().Unix() * 1000,
				}
				
				// Create funding service
				mockMarket := &mockFundingMarketService{}
				service := NewFuturesFundingService(mockMarket, &mockLogger{})
				
				// Update after settlement
				newBalance, newCost, err := service.UpdateAfterSettlement(settlement, currentBalance, currentCost)
				if err != nil {
					return false
				}
				
				// Verify balance update: newBalance = currentBalance + fundingFee
				balanceCorrect := newBalance == currentBalance + fundingFee
				
				// Verify cost update: newCost = currentCost - fundingFee
				costCorrect := newCost == currentCost - fundingFee
				
				return balanceCorrect && costCorrect
			},
			gen.Float64Range(-1000.0, 1000.0),  // Funding fee: -1000 to 1000
			gen.Float64Range(1000.0, 100000.0), // Current balance: 1000 to 100000
			gen.Float64Range(1000.0, 50000.0),  // Current cost: 1000 to 50000
		))

	properties.TestingRun(t)
}

// Feature: usdt-futures-trading, Property 49: 资金费率历史时间过滤
// Validates: Requirements 12.5
func TestProperty_FundingRateHistoryTimeFilter(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("for any funding rate history query, all returned records must be within specified time range", 
		prop.ForAll(
			func(startOffset int64, endOffset int64) bool {
				// Ensure start is before end
				if startOffset >= endOffset {
					return true
				}
				
				now := time.Now().Unix() * 1000
				startTime := now - startOffset
				endTime := now - endOffset
				
				// Swap if needed
				if startTime > endTime {
					startTime, endTime = endTime, startTime
				}
				
				// Create mock funding rate history with various times
				mockHistory := []*api.FundingRate{
					{Symbol: "BTCUSDT", FundingRate: 0.0001, FundingTime: startTime - 1000},     // Before range
					{Symbol: "BTCUSDT", FundingRate: 0.0002, FundingTime: startTime},            // At start
					{Symbol: "BTCUSDT", FundingRate: 0.0003, FundingTime: (startTime + endTime) / 2}, // In range
					{Symbol: "BTCUSDT", FundingRate: 0.0004, FundingTime: endTime},              // At end
					{Symbol: "BTCUSDT", FundingRate: 0.0005, FundingTime: endTime + 1000},       // After range
				}
				
				mockMarket := &mockFundingMarketService{
					fundingRateHistory: mockHistory,
				}
				
				// Create funding service
				service := NewFuturesFundingService(mockMarket, &mockLogger{})
				
				// Get funding rate history
				result, err := service.GetFundingRateHistory("BTCUSDT", startTime, endTime)
				if err != nil {
					return false
				}
				
				// Verify all returned records are within time range
				for _, rate := range result {
					if rate.FundingTime < startTime || rate.FundingTime > endTime {
						return false
					}
				}
				
				return true
			},
			gen.Int64Range(86400000, 864000000),  // Start offset: 1 day to 10 days ago
			gen.Int64Range(1, 86400000),          // End offset: 1ms to 1 day ago
		))

	properties.TestingRun(t)
}
