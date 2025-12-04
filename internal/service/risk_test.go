package service

import (
	"binance-trader/internal/api"
	"binance-trader/pkg/errors"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Feature: binance-auto-trading, Property 18: 最小余额保护
// 对于任何买入订单请求，如果执行后会导致余额低于最小保留金额，系统必须拒绝该订单
// Validates: Requirements 5.3
func TestProperty_MinimumBalanceProtection(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	
	properties := gopter.NewProperties(parameters)
	
	properties.Property("balance below minimum reserve should be rejected", prop.ForAll(
		func(minReserve float64, currentBalance float64) bool {
			// Create mock client that returns the specified balance
			mockClient := &mockBinanceClient{
				getBalanceFunc: func(asset string) (*api.Balance, error) {
					return &api.Balance{
						Asset:  asset,
						Free:   currentBalance,
						Locked: 0,
					}, nil
				},
			}
			
			// Create risk manager with the specified minimum reserve
			riskMgr := NewRiskManager(&RiskLimits{
				MaxOrderAmount:    100000.0,
				MaxDailyOrders:    1000,
				MinBalanceReserve: minReserve,
				MaxAPICallsPerMin: 1000,
			}, mockClient)
			
			// Check minimum balance for USDT
			err := riskMgr.CheckMinimumBalance("USDT")
			
			// If current balance is below minimum reserve, error should be returned
			if currentBalance < minReserve {
				if err == nil {
					return false
				}
				
				// Verify it's the correct error type
				tradingErr, ok := err.(*errors.TradingError)
				if !ok {
					return false
				}
				
				if tradingErr.Type != errors.ErrRiskLimitExceeded {
					return false
				}
				
				return true
			}
			
			// If current balance is at or above minimum reserve, no error should be returned
			return err == nil
		},
		gen.Float64Range(0, 10000),   // minReserve: 0 to 10000
		gen.Float64Range(0, 10000),   // currentBalance: 0 to 10000
	))
	
	properties.TestingRun(t)
}
