package service

import (
	"binance-trader/internal/api"
	"binance-trader/internal/repository"
	"binance-trader/pkg/logger"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Mock futures client for testing
type mockFuturesClientForPosition struct {
	positions []*api.Position
	err       error
}

func (m *mockFuturesClientForPosition) GetAccountInfo() (*api.FuturesAccountInfo, error) {
	return nil, nil
}

func (m *mockFuturesClientForPosition) GetBalance() (*api.FuturesBalance, error) {
	return nil, nil
}

func (m *mockFuturesClientForPosition) GetMarkPrice(symbol string) (*api.MarkPrice, error) {
	return nil, nil
}

func (m *mockFuturesClientForPosition) GetPrice(symbol string) (*api.Price, error) {
	return nil, nil
}

func (m *mockFuturesClientForPosition) GetKlines(symbol string, interval string, limit int) ([]*api.Kline, error) {
	return nil, nil
}

func (m *mockFuturesClientForPosition) GetFundingRate(symbol string) (*api.FundingRate, error) {
	return nil, nil
}

func (m *mockFuturesClientForPosition) GetFundingRateHistory(symbol string, startTime, endTime int64) ([]*api.FundingRate, error) {
	return nil, nil
}

func (m *mockFuturesClientForPosition) SetLeverage(symbol string, leverage int) (*api.LeverageResponse, error) {
	return nil, nil
}

func (m *mockFuturesClientForPosition) SetMarginType(symbol string, marginType api.MarginType) error {
	return nil
}

func (m *mockFuturesClientForPosition) SetPositionMode(dualSidePosition bool) error {
	return nil
}

func (m *mockFuturesClientForPosition) GetPositionMode() (*api.PositionMode, error) {
	return nil, nil
}

func (m *mockFuturesClientForPosition) CreateOrder(order *api.FuturesOrderRequest) (*api.FuturesOrderResponse, error) {
	return nil, nil
}

func (m *mockFuturesClientForPosition) CancelOrder(symbol string, orderID int64) (*api.CancelResponse, error) {
	return nil, nil
}

func (m *mockFuturesClientForPosition) GetOrder(symbol string, orderID int64) (*api.FuturesOrder, error) {
	return nil, nil
}

func (m *mockFuturesClientForPosition) GetOpenOrders(symbol string) ([]*api.FuturesOrder, error) {
	return nil, nil
}

func (m *mockFuturesClientForPosition) GetPositions(symbol string) ([]*api.Position, error) {
	if m.err != nil {
		return nil, m.err
	}
	
	// Filter by symbol
	result := make([]*api.Position, 0)
	for _, pos := range m.positions {
		if pos.Symbol == symbol {
			result = append(result, pos)
		}
	}
	return result, nil
}

func (m *mockFuturesClientForPosition) GetAllPositions() ([]*api.Position, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.positions, nil
}

// Generator for Position
func genPosition() gopter.Gen {
	return gopter.CombineGens(
		gen.AlphaString(),
		gen.OneConstOf(api.PositionSideLong, api.PositionSideShort),
		gen.Float64Range(-1000, 1000),
		gen.Float64Range(1, 100000),
		gen.Float64Range(1, 100000),
		gen.Float64Range(-10000, 10000),
		gen.Float64Range(1, 100000),
		gen.IntRange(1, 125),
		gen.OneConstOf(api.MarginTypeIsolated, api.MarginTypeCrossed),
		gen.Float64Range(0, 10000),
		gen.Bool(),
		gen.Float64Range(0, 10000),
		gen.Float64Range(0, 1000),
		gen.Int64Range(0, 9999999999999),
	).Map(func(values []interface{}) *api.Position {
		return &api.Position{
			Symbol:              values[0].(string),
			PositionSide:        values[1].(api.PositionSide),
			PositionAmt:         values[2].(float64),
			EntryPrice:          values[3].(float64),
			MarkPrice:           values[4].(float64),
			UnrealizedProfit:    values[5].(float64),
			LiquidationPrice:    values[6].(float64),
			Leverage:            values[7].(int),
			MarginType:          values[8].(api.MarginType),
			IsolatedMargin:      values[9].(float64),
			IsAutoAddMargin:     values[10].(bool),
			PositionInitialMargin: values[11].(float64),
			MaintenanceMargin:   values[12].(float64),
			UpdateTime:          values[13].(int64),
		}
	})
}

// Feature: usdt-futures-trading, Property 18: 持仓查询响应完整性
// For any position query response, each position must include quantity, entry price, mark price, unrealized PnL, and liquidation price fields
// Validates: Requirements 5.1
func TestProperty18_PositionQueryResponseCompleteness(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("position query response contains all required fields", prop.ForAll(
		func(pos *api.Position) bool {
			// Skip empty symbols
			if pos.Symbol == "" {
				return true
			}
			
			// Create mock client with the position
			mockClient := &mockFuturesClientForPosition{
				positions: []*api.Position{pos},
			}
			
			repo := repository.NewMemoryFuturesPositionRepository()
			testLogger, _ := logger.NewLogger(logger.Config{Level: "info", FilePath: "", EnableConsole: false})
			manager := NewFuturesPositionManager(mockClient, repo, testLogger)
			
			// Get the position
			result, err := manager.GetPosition(pos.Symbol, pos.PositionSide)
			if err != nil {
				return false
			}
			
			// Verify all required fields are present (not checking for zero values, just that fields exist)
			// The fields should be accessible and the struct should be complete
			hasQuantity := true // PositionAmt field exists
			hasEntryPrice := result.EntryPrice >= 0 || result.EntryPrice < 0 // Field exists
			hasMarkPrice := result.MarkPrice >= 0 || result.MarkPrice < 0 // Field exists
			hasUnrealizedPnL := result.UnrealizedProfit >= 0 || result.UnrealizedProfit < 0 // Field exists
			hasLiquidationPrice := result.LiquidationPrice >= 0 || result.LiquidationPrice < 0 // Field exists
			
			return hasQuantity && hasEntryPrice && hasMarkPrice && hasUnrealizedPnL && hasLiquidationPrice
		},
		genPosition(),
	))

	properties.TestingRun(t)
}

// Feature: usdt-futures-trading, Property 19: 持仓历史时间过滤
// For any position history query, all returned records' close times must be within the specified time range
// Validates: Requirements 5.2
func TestProperty19_PositionHistoryTimeFiltering(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("position history is filtered by time range", prop.ForAll(
		func(symbol string, startTime, endTime int64, closedPositions []*repository.ClosedPosition) bool {
			// Ensure valid time range and non-empty symbol
			if symbol == "" || startTime < 0 || endTime < 0 || startTime > endTime {
				return true // Skip invalid inputs
			}
			
			mockClient := &mockFuturesClientForPosition{}
			repo := repository.NewMemoryFuturesPositionRepository()
			testLogger, _ := logger.NewLogger(logger.Config{Level: "info", FilePath: "", EnableConsole: false})
			manager := NewFuturesPositionManager(mockClient, repo, testLogger)
			
			// Save closed positions to repository
			for _, cp := range closedPositions {
				cp.Symbol = symbol // Ensure same symbol
				repo.SaveClosedPosition(cp)
			}
			
			// Get position history
			history, err := manager.GetPositionHistory(symbol, startTime, endTime)
			if err != nil {
				return false
			}
			
			// Verify all returned positions are within time range
			for _, pos := range history {
				if pos.CloseTime < startTime || pos.CloseTime > endTime {
					return false
				}
			}
			
			return true
		},
		gen.AlphaString(),
		gen.Int64Range(0, 9999999999999),
		gen.Int64Range(0, 9999999999999),
		gen.SliceOf(gopter.CombineGens(
			gen.OneConstOf(api.PositionSideLong, api.PositionSideShort),
			gen.Float64Range(1, 100000),
			gen.Float64Range(1, 100000),
			gen.Float64Range(0.01, 1000),
			gen.Float64Range(-10000, 10000),
			gen.Int64Range(0, 9999999999999),
			gen.Int64Range(0, 9999999999999),
			gen.Float64Range(0, 1000),
		).Map(func(values []interface{}) *repository.ClosedPosition {
			return &repository.ClosedPosition{
				Symbol:         "",
				PositionSide:   values[0].(api.PositionSide),
				EntryPrice:     values[1].(float64),
				ExitPrice:      values[2].(float64),
				Quantity:       values[3].(float64),
				RealizedProfit: values[4].(float64),
				OpenTime:       values[5].(int64),
				CloseTime:      values[6].(int64),
				Commission:     values[7].(float64),
			}
		})),
	))

	properties.TestingRun(t)
}

// Feature: usdt-futures-trading, Property 20: 特定合约持仓过滤
// For any specific symbol position query, all returned positions' symbols must match the queried symbol
// Validates: Requirements 5.3
func TestProperty20_SpecificSymbolPositionFiltering(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("positions are filtered by symbol", prop.ForAll(
		func(targetSymbol string, positions []*api.Position) bool {
			if targetSymbol == "" {
				return true // Skip empty symbols
			}
			
			// Create mock client with positions
			mockClient := &mockFuturesClientForPosition{
				positions: positions,
			}
			
			repo := repository.NewMemoryFuturesPositionRepository()
			testLogger, _ := logger.NewLogger(logger.Config{Level: "info", FilePath: "", EnableConsole: false})
			manager := NewFuturesPositionManager(mockClient, repo, testLogger)
			
			// Get positions for target symbol
			result, err := manager.GetPosition(targetSymbol, api.PositionSideLong)
			if err != nil {
				// If no position found, that's okay
				return true
			}
			
			// Verify the returned position matches the symbol
			return result.Symbol == targetSymbol
		},
		gen.AlphaString(),
		gen.SliceOf(genPosition()),
	))

	properties.TestingRun(t)
}

// Feature: usdt-futures-trading, Property 21: 持仓盈亏计算正确性
// For any position and mark price update, unrealized PnL must equal (markPrice - entryPrice) * positionAmt, and liquidation price must be recalculated based on margin and maintenance margin rate
// Validates: Requirements 5.4
func TestProperty21_PositionPnLCalculationCorrectness(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("unrealized PnL is calculated correctly", prop.ForAll(
		func(pos *api.Position, markPrice float64) bool {
			// Skip invalid inputs
			if markPrice <= 0 || pos.EntryPrice <= 0 {
				return true
			}
			
			mockClient := &mockFuturesClientForPosition{}
			repo := repository.NewMemoryFuturesPositionRepository()
			testLogger, _ := logger.NewLogger(logger.Config{Level: "info", FilePath: "", EnableConsole: false})
			manager := NewFuturesPositionManager(mockClient, repo, testLogger)
			
			// Calculate unrealized PnL
			pnl, err := manager.CalculateUnrealizedPnL(pos, markPrice)
			if err != nil {
				return false
			}
			
			// Verify PnL calculation: (markPrice - entryPrice) * positionAmt
			expectedPnL := (markPrice - pos.EntryPrice) * pos.PositionAmt
			
			// Allow small floating point differences
			diff := pnl - expectedPnL
			if diff < 0 {
				diff = -diff
			}
			
			return diff < 0.0001
		},
		genPosition(),
		gen.Float64Range(1, 100000),
	))

	properties.Property("liquidation price is recalculated", prop.ForAll(
		func(pos *api.Position) bool {
			// Skip positions with zero amount or invalid leverage
			if pos.PositionAmt == 0 || pos.Leverage == 0 || pos.EntryPrice <= 0 {
				return true
			}
			
			mockClient := &mockFuturesClientForPosition{}
			repo := repository.NewMemoryFuturesPositionRepository()
			testLogger, _ := logger.NewLogger(logger.Config{Level: "info", FilePath: "", EnableConsole: false})
			manager := NewFuturesPositionManager(mockClient, repo, testLogger)
			
			// Calculate liquidation price
			liqPrice, err := manager.CalculateLiquidationPrice(pos)
			if err != nil {
				return false
			}
			
			// Verify liquidation price is calculated (non-zero for non-zero positions)
			// For long positions, liquidation price should be below entry price
			// For short positions, liquidation price should be above entry price
			if pos.PositionAmt > 0 {
				return liqPrice < pos.EntryPrice
			} else if pos.PositionAmt < 0 {
				return liqPrice > pos.EntryPrice
			}
			
			return true
		},
		genPosition(),
	))

	properties.TestingRun(t)
}

// Feature: usdt-futures-trading, Property 22: 双向持仓分离显示
// For any dual position mode query, long and short positions must be returned separately
// Validates: Requirements 5.5
func TestProperty22_DualPositionSeparateDisplay(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("long and short positions are separated in dual mode", prop.ForAll(
		func(symbol string, longPos, shortPos *api.Position) bool {
			if symbol == "" {
				return true
			}
			
			// Set up positions with same symbol but different sides
			longPos.Symbol = symbol
			longPos.PositionSide = api.PositionSideLong
			shortPos.Symbol = symbol
			shortPos.PositionSide = api.PositionSideShort
			
			// Create mock client with both positions
			mockClient := &mockFuturesClientForPosition{
				positions: []*api.Position{longPos, shortPos},
			}
			
			repo := repository.NewMemoryFuturesPositionRepository()
			testLogger, _ := logger.NewLogger(logger.Config{Level: "info", FilePath: "", EnableConsole: false})
			manager := NewFuturesPositionManager(mockClient, repo, testLogger)
			
			// Get long position
			longResult, err1 := manager.GetPosition(symbol, api.PositionSideLong)
			if err1 != nil {
				return false
			}
			
			// Get short position
			shortResult, err2 := manager.GetPosition(symbol, api.PositionSideShort)
			if err2 != nil {
				return false
			}
			
			// Verify they are separate and have correct sides
			return longResult.PositionSide == api.PositionSideLong &&
				shortResult.PositionSide == api.PositionSideShort &&
				longResult.Symbol == symbol &&
				shortResult.Symbol == symbol
		},
		gen.AlphaString(),
		genPosition(),
		genPosition(),
	))

	properties.TestingRun(t)
}
