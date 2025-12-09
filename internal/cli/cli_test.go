package cli

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"binance-trader/internal/api"
	"binance-trader/internal/repository"
	"binance-trader/internal/service"
)

// mockTradingService is a mock implementation of TradingService
type mockTradingService struct {
	placeMarketBuyOrderFunc  func(symbol string, quantity float64) (*api.Order, error)
	placeMarketSellOrderFunc func(symbol string, quantity float64) (*api.Order, error)
	placeLimitSellOrderFunc  func(symbol string, price, quantity float64) (*api.Order, error)
	cancelOrderFunc          func(orderID int64) error
	getOrderStatusFunc       func(orderID int64) (*service.OrderStatus, error)
	getActiveOrdersFunc      func() ([]*api.Order, error)
}

func (m *mockTradingService) PlaceMarketBuyOrder(symbol string, quantity float64) (*api.Order, error) {
	if m.placeMarketBuyOrderFunc != nil {
		return m.placeMarketBuyOrderFunc(symbol, quantity)
	}
	return nil, nil
}

func (m *mockTradingService) PlaceMarketSellOrder(symbol string, quantity float64) (*api.Order, error) {
	if m.placeMarketSellOrderFunc != nil {
		return m.placeMarketSellOrderFunc(symbol, quantity)
	}
	return nil, nil
}

func (m *mockTradingService) PlaceLimitSellOrder(symbol string, price, quantity float64) (*api.Order, error) {
	if m.placeLimitSellOrderFunc != nil {
		return m.placeLimitSellOrderFunc(symbol, price, quantity)
	}
	return nil, nil
}

func (m *mockTradingService) CancelOrder(orderID int64) error {
	if m.cancelOrderFunc != nil {
		return m.cancelOrderFunc(orderID)
	}
	return nil
}

func (m *mockTradingService) GetOrderStatus(orderID int64) (*service.OrderStatus, error) {
	if m.getOrderStatusFunc != nil {
		return m.getOrderStatusFunc(orderID)
	}
	return nil, nil
}

func (m *mockTradingService) GetActiveOrders() ([]*api.Order, error) {
	if m.getActiveOrdersFunc != nil {
		return m.getActiveOrdersFunc()
	}
	return nil, nil
}

// mockMarketDataService is a mock implementation of MarketDataService
type mockMarketDataService struct {
	getCurrentPriceFunc     func(symbol string) (float64, error)
	getHistoricalDataFunc   func(symbol string, interval string, limit int) ([]*api.Kline, error)
	subscribeToPriceFunc    func(symbol string, callback func(float64)) error
	getVolumeFunc           func(symbol string, timeWindow time.Duration) (float64, error)
}

func (m *mockMarketDataService) GetCurrentPrice(symbol string) (float64, error) {
	if m.getCurrentPriceFunc != nil {
		return m.getCurrentPriceFunc(symbol)
	}
	return 0, nil
}

func (m *mockMarketDataService) GetHistoricalData(symbol string, interval string, limit int) ([]*api.Kline, error) {
	if m.getHistoricalDataFunc != nil {
		return m.getHistoricalDataFunc(symbol, interval, limit)
	}
	return nil, nil
}

func (m *mockMarketDataService) SubscribeToPrice(symbol string, callback func(float64)) error {
	if m.subscribeToPriceFunc != nil {
		return m.subscribeToPriceFunc(symbol, callback)
	}
	return nil
}

func (m *mockMarketDataService) GetVolume(symbol string, timeWindow time.Duration) (float64, error) {
	if m.getVolumeFunc != nil {
		return m.getVolumeFunc(symbol, timeWindow)
	}
	return 0, nil
}

// mockLogger is a mock implementation of Logger
type mockLogger struct{}

func (m *mockLogger) Debug(msg string, fields map[string]interface{})                      {}
func (m *mockLogger) Info(msg string, fields map[string]interface{})                       {}
func (m *mockLogger) Warn(msg string, fields map[string]interface{})                       {}
func (m *mockLogger) Error(msg string, fields map[string]interface{})                      {}
func (m *mockLogger) Fatal(msg string, fields map[string]interface{})                      {}
func (m *mockLogger) LogAPIOperation(operation string, result string, fields map[string]interface{}) {}
func (m *mockLogger) LogOrderEvent(event string, orderID int64, symbol, side, orderType string, quantity float64, fields map[string]interface{}) {}
func (m *mockLogger) LogError(err error, fields map[string]interface{})                    {}
func (m *mockLogger) LogFuturesAPIOperation(operationType string, result string, fields map[string]interface{}) {}
func (m *mockLogger) LogFuturesOrderEvent(eventType string, orderID int64, symbol, side, orderType string, quantity float64, positionChange map[string]interface{}, fields map[string]interface{}) {}
func (m *mockLogger) LogLiquidationEvent(symbol string, positionSide string, liquidationPrice float64, lossAmount float64, reason string, fields map[string]interface{}) {}
func (m *mockLogger) LogFundingRateSettlement(symbol string, fundingFee float64, fundingRate float64, positionSize float64, fields map[string]interface{}) {}
func (m *mockLogger) SetTradingType(tradingType string)                                    {}

// TestParseCommand tests command parsing
func TestParseCommand(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantName    string
		wantArgs    []string
		wantErr     bool
	}{
		{
			name:     "simple command",
			input:    "help",
			wantName: "help",
			wantArgs: []string{},
			wantErr:  false,
		},
		{
			name:     "command with one arg",
			input:    "price BTCUSDT",
			wantName: "price",
			wantArgs: []string{"BTCUSDT"},
			wantErr:  false,
		},
		{
			name:     "command with multiple args",
			input:    "buy BTCUSDT 0.001",
			wantName: "buy",
			wantArgs: []string{"BTCUSDT", "0.001"},
			wantErr:  false,
		},
		{
			name:     "command with extra spaces",
			input:    "  sell   BTCUSDT   50000   0.001  ",
			wantName: "sell",
			wantArgs: []string{"BTCUSDT", "50000", "0.001"},
			wantErr:  false,
		},
		{
			name:     "empty command",
			input:    "",
			wantErr:  true,
		},
		{
			name:     "whitespace only",
			input:    "   ",
			wantErr:  true,
		},
		{
			name:     "case insensitive",
			input:    "HELP",
			wantName: "help",
			wantArgs: []string{},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := ParseCommand(tt.input)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseCommand() expected error, got nil")
				}
				return
			}
			
			if err != nil {
				t.Errorf("ParseCommand() unexpected error: %v", err)
				return
			}
			
			if cmd.Name != tt.wantName {
				t.Errorf("ParseCommand() name = %v, want %v", cmd.Name, tt.wantName)
			}
			
			if len(cmd.Args) != len(tt.wantArgs) {
				t.Errorf("ParseCommand() args length = %v, want %v", len(cmd.Args), len(tt.wantArgs))
				return
			}
			
			for i, arg := range cmd.Args {
				if arg != tt.wantArgs[i] {
					t.Errorf("ParseCommand() args[%d] = %v, want %v", i, arg, tt.wantArgs[i])
				}
			}
		})
	}
}

// TestFormatPrice tests price formatting
func TestFormatPrice(t *testing.T) {
	mockTrading := &mockTradingService{}
	mockMarket := &mockMarketDataService{}
	mockLog := &mockLogger{}
	
	cli := NewCLI(mockTrading, mockMarket, &mockConditionalOrderService{}, &mockStopLossService{}, mockLog)
	
	var buf bytes.Buffer
	cli.writer = &buf
	
	cli.formatPrice("BTCUSDT", 50000.12345678)
	
	output := buf.String()
	
	if !strings.Contains(output, "BTCUSDT") {
		t.Errorf("formatPrice() output should contain symbol")
	}
	
	if !strings.Contains(output, "50000.12345678") {
		t.Errorf("formatPrice() output should contain price")
	}
}

// TestFormatOrder tests order formatting
func TestFormatOrder(t *testing.T) {
	mockTrading := &mockTradingService{}
	mockMarket := &mockMarketDataService{}
	mockLog := &mockLogger{}
	
	cli := NewCLI(mockTrading, mockMarket, &mockConditionalOrderService{}, &mockStopLossService{}, mockLog)
	
	var buf bytes.Buffer
	cli.writer = &buf
	
	order := &api.Order{
		OrderID:             12345,
		Symbol:              "BTCUSDT",
		Side:                api.OrderSideBuy,
		Type:                api.OrderTypeMarket,
		Status:              api.OrderStatusFilled,
		Price:               50000.0,
		OrigQty:             0.001,
		ExecutedQty:         0.001,
		CummulativeQuoteQty: 50.0,
	}
	
	cli.formatOrder(order)
	
	output := buf.String()
	
	// Check that all important fields are present
	expectedFields := []string{
		"12345",
		"BTCUSDT",
		"BUY",
		"MARKET",
		"FILLED",
		"50000",
		"0.001",
	}
	
	for _, field := range expectedFields {
		if !strings.Contains(output, field) {
			t.Errorf("formatOrder() output should contain %s", field)
		}
	}
}

// TestFormatOrderStatus tests order status formatting
func TestFormatOrderStatus(t *testing.T) {
	mockTrading := &mockTradingService{}
	mockMarket := &mockMarketDataService{}
	mockLog := &mockLogger{}
	
	cli := NewCLI(mockTrading, mockMarket, &mockConditionalOrderService{}, &mockStopLossService{}, mockLog)
	
	var buf bytes.Buffer
	cli.writer = &buf
	
	status := &service.OrderStatus{
		OrderID:     12345,
		Symbol:      "BTCUSDT",
		Status:      api.OrderStatusFilled,
		ExecutedQty: 0.001,
		Price:       50000.0,
	}
	
	cli.formatOrderStatus(status)
	
	output := buf.String()
	
	expectedFields := []string{
		"12345",
		"BTCUSDT",
		"FILLED",
		"0.001",
		"50000",
	}
	
	for _, field := range expectedFields {
		if !strings.Contains(output, field) {
			t.Errorf("formatOrderStatus() output should contain %s", field)
		}
	}
}

// TestFormatOrderList tests order list formatting
func TestFormatOrderList(t *testing.T) {
	mockTrading := &mockTradingService{}
	mockMarket := &mockMarketDataService{}
	mockLog := &mockLogger{}
	
	cli := NewCLI(mockTrading, mockMarket, &mockConditionalOrderService{}, &mockStopLossService{}, mockLog)
	
	t.Run("empty list", func(t *testing.T) {
		var buf bytes.Buffer
		cli.writer = &buf
		
		cli.formatOrderList([]*api.Order{})
		
		output := buf.String()
		if !strings.Contains(output, "No active orders") {
			t.Errorf("formatOrderList() should show 'No active orders' for empty list")
		}
	})
	
	t.Run("with orders", func(t *testing.T) {
		var buf bytes.Buffer
		cli.writer = &buf
		
		orders := []*api.Order{
			{
				OrderID: 12345,
				Symbol:  "BTCUSDT",
				Side:    api.OrderSideBuy,
				Type:    api.OrderTypeMarket,
				Status:  api.OrderStatusNew,
			},
			{
				OrderID: 67890,
				Symbol:  "ETHUSDT",
				Side:    api.OrderSideSell,
				Type:    api.OrderTypeLimit,
				Status:  api.OrderStatusNew,
			},
		}
		
		cli.formatOrderList(orders)
		
		output := buf.String()
		
		if !strings.Contains(output, "12345") {
			t.Errorf("formatOrderList() should contain first order ID")
		}
		
		if !strings.Contains(output, "67890") {
			t.Errorf("formatOrderList() should contain second order ID")
		}
		
		if !strings.Contains(output, "BTCUSDT") {
			t.Errorf("formatOrderList() should contain first symbol")
		}
		
		if !strings.Contains(output, "ETHUSDT") {
			t.Errorf("formatOrderList() should contain second symbol")
		}
	})
}

// TestFormatKlines tests kline data formatting
func TestFormatKlines(t *testing.T) {
	mockTrading := &mockTradingService{}
	mockMarket := &mockMarketDataService{}
	mockLog := &mockLogger{}
	
	cli := NewCLI(mockTrading, mockMarket, &mockConditionalOrderService{}, &mockStopLossService{}, mockLog)
	
	t.Run("empty klines", func(t *testing.T) {
		var buf bytes.Buffer
		cli.writer = &buf
		
		cli.formatKlines("BTCUSDT", "1h", []*api.Kline{})
		
		output := buf.String()
		if !strings.Contains(output, "No kline data available") {
			t.Errorf("formatKlines() should show 'No kline data available' for empty list")
		}
	})
	
	t.Run("with klines", func(t *testing.T) {
		var buf bytes.Buffer
		cli.writer = &buf
		
		klines := []*api.Kline{
			{
				OpenTime:  1609459200000,
				Open:      50000.0,
				High:      51000.0,
				Low:       49000.0,
				Close:     50500.0,
				Volume:    100.5,
				CloseTime: 1609462800000,
			},
		}
		
		cli.formatKlines("BTCUSDT", "1h", klines)
		
		output := buf.String()
		
		if !strings.Contains(output, "BTCUSDT") {
			t.Errorf("formatKlines() should contain symbol")
		}
		
		if !strings.Contains(output, "1h") {
			t.Errorf("formatKlines() should contain interval")
		}
		
		// Check for price values
		if !strings.Contains(output, "50000") {
			t.Errorf("formatKlines() should contain open price")
		}
	})
}

// TestHandlePrice tests the price command handler
func TestHandlePrice(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockMarket := &mockMarketDataService{
			getCurrentPriceFunc: func(symbol string) (float64, error) {
				if symbol == "BTCUSDT" {
					return 50000.0, nil
				}
				return 0, nil
			},
		}
		
		cli := NewCLI(&mockTradingService{}, mockMarket, &mockConditionalOrderService{}, &mockStopLossService{}, &mockLogger{})
		
		var buf bytes.Buffer
		cli.writer = &buf
		
		err := cli.handlePrice([]string{"BTCUSDT"})
		if err != nil {
			t.Errorf("handlePrice() unexpected error: %v", err)
		}
		
		output := buf.String()
		if !strings.Contains(output, "BTCUSDT") {
			t.Errorf("handlePrice() output should contain symbol")
		}
	})
	
	t.Run("missing argument", func(t *testing.T) {
		cli := NewCLI(&mockTradingService{}, &mockMarketDataService{}, &mockConditionalOrderService{}, &mockStopLossService{}, &mockLogger{})
		
		err := cli.handlePrice([]string{})
		if err == nil {
			t.Errorf("handlePrice() expected error for missing argument")
		}
	})
}

// TestHandleBuy tests the buy command handler
func TestHandleBuy(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockTrading := &mockTradingService{
			placeMarketBuyOrderFunc: func(symbol string, quantity float64) (*api.Order, error) {
				return &api.Order{
					OrderID: 12345,
					Symbol:  symbol,
					Side:    api.OrderSideBuy,
					Type:    api.OrderTypeMarket,
					Status:  api.OrderStatusFilled,
					OrigQty: quantity,
				}, nil
			},
		}
		
		cli := NewCLI(mockTrading, &mockMarketDataService{}, &mockConditionalOrderService{}, &mockStopLossService{}, &mockLogger{})
		
		var buf bytes.Buffer
		cli.writer = &buf
		
		err := cli.handleBuy([]string{"BTCUSDT", "0.001"})
		if err != nil {
			t.Errorf("handleBuy() unexpected error: %v", err)
		}
		
		output := buf.String()
		if !strings.Contains(output, "12345") {
			t.Errorf("handleBuy() output should contain order ID")
		}
	})
	
	t.Run("missing arguments", func(t *testing.T) {
		cli := NewCLI(&mockTradingService{}, &mockMarketDataService{}, &mockConditionalOrderService{}, &mockStopLossService{}, &mockLogger{})
		
		err := cli.handleBuy([]string{"BTCUSDT"})
		if err == nil {
			t.Errorf("handleBuy() expected error for missing argument")
		}
	})
	
	t.Run("invalid quantity", func(t *testing.T) {
		cli := NewCLI(&mockTradingService{}, &mockMarketDataService{}, &mockConditionalOrderService{}, &mockStopLossService{}, &mockLogger{})
		
		err := cli.handleBuy([]string{"BTCUSDT", "invalid"})
		if err == nil {
			t.Errorf("handleBuy() expected error for invalid quantity")
		}
	})
}

// TestHandleSell tests the sell command handler
func TestHandleSell(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockTrading := &mockTradingService{
			placeLimitSellOrderFunc: func(symbol string, price, quantity float64) (*api.Order, error) {
				return &api.Order{
					OrderID: 12345,
					Symbol:  symbol,
					Side:    api.OrderSideSell,
					Type:    api.OrderTypeLimit,
					Status:  api.OrderStatusNew,
					Price:   price,
					OrigQty: quantity,
				}, nil
			},
		}
		
		cli := NewCLI(mockTrading, &mockMarketDataService{}, &mockConditionalOrderService{}, &mockStopLossService{}, &mockLogger{})
		
		var buf bytes.Buffer
		cli.writer = &buf
		
		err := cli.handleSell([]string{"BTCUSDT", "50000", "0.001"})
		if err != nil {
			t.Errorf("handleSell() unexpected error: %v", err)
		}
		
		output := buf.String()
		if !strings.Contains(output, "12345") {
			t.Errorf("handleSell() output should contain order ID")
		}
	})
	
	t.Run("missing arguments", func(t *testing.T) {
		cli := NewCLI(&mockTradingService{}, &mockMarketDataService{}, &mockConditionalOrderService{}, &mockStopLossService{}, &mockLogger{})
		
		err := cli.handleSell([]string{"BTCUSDT", "50000"})
		if err == nil {
			t.Errorf("handleSell() expected error for missing argument")
		}
	})
}

// TestHandleCancel tests the cancel command handler
func TestHandleCancel(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockTrading := &mockTradingService{
			cancelOrderFunc: func(orderID int64) error {
				return nil
			},
		}
		
		cli := NewCLI(mockTrading, &mockMarketDataService{}, &mockConditionalOrderService{}, &mockStopLossService{}, &mockLogger{})
		
		var buf bytes.Buffer
		cli.writer = &buf
		
		err := cli.handleCancel([]string{"12345"})
		if err != nil {
			t.Errorf("handleCancel() unexpected error: %v", err)
		}
		
		output := buf.String()
		if !strings.Contains(output, "canceled successfully") {
			t.Errorf("handleCancel() output should contain success message")
		}
	})
	
	t.Run("invalid order ID", func(t *testing.T) {
		cli := NewCLI(&mockTradingService{}, &mockMarketDataService{}, &mockConditionalOrderService{}, &mockStopLossService{}, &mockLogger{})
		
		err := cli.handleCancel([]string{"invalid"})
		if err == nil {
			t.Errorf("handleCancel() expected error for invalid order ID")
		}
	})
}

// mockConditionalOrderService is a mock implementation of ConditionalOrderService
type mockConditionalOrderService struct {
	createConditionalOrderFunc       func(request *repository.ConditionalOrderRequest) (*repository.ConditionalOrder, error)
	cancelConditionalOrderFunc       func(orderID string) error
	getActiveConditionalOrdersFunc   func() ([]*repository.ConditionalOrder, error)
}

func (m *mockConditionalOrderService) CreateConditionalOrder(request *repository.ConditionalOrderRequest) (*repository.ConditionalOrder, error) {
	if m.createConditionalOrderFunc != nil {
		return m.createConditionalOrderFunc(request)
	}
	return nil, nil
}

func (m *mockConditionalOrderService) CancelConditionalOrder(orderID string) error {
	if m.cancelConditionalOrderFunc != nil {
		return m.cancelConditionalOrderFunc(orderID)
	}
	return nil
}

func (m *mockConditionalOrderService) UpdateConditionalOrder(orderID string, updates *service.ConditionalOrderUpdate) error {
	return nil
}

func (m *mockConditionalOrderService) GetConditionalOrder(orderID string) (*repository.ConditionalOrder, error) {
	return nil, nil
}

func (m *mockConditionalOrderService) GetActiveConditionalOrders() ([]*repository.ConditionalOrder, error) {
	if m.getActiveConditionalOrdersFunc != nil {
		return m.getActiveConditionalOrdersFunc()
	}
	return nil, nil
}

func (m *mockConditionalOrderService) GetConditionalOrderHistory(startTime, endTime int64) ([]*repository.ConditionalOrder, error) {
	return nil, nil
}

func (m *mockConditionalOrderService) StartMonitoring() error {
	return nil
}

func (m *mockConditionalOrderService) StopMonitoring() error {
	return nil
}

// mockStopLossService is a mock implementation of StopLossService
type mockStopLossService struct {
	setStopLossFunc         func(symbol string, position float64, stopPrice float64) (*repository.StopOrder, error)
	setTakeProfitFunc       func(symbol string, position float64, targetPrice float64) (*repository.StopOrder, error)
	cancelStopOrderFunc     func(orderID string) error
	getActiveStopOrdersFunc func(symbol string) ([]*repository.StopOrder, error)
}

func (m *mockStopLossService) SetStopLoss(symbol string, position float64, stopPrice float64) (*repository.StopOrder, error) {
	if m.setStopLossFunc != nil {
		return m.setStopLossFunc(symbol, position, stopPrice)
	}
	return nil, nil
}

func (m *mockStopLossService) SetTakeProfit(symbol string, position float64, targetPrice float64) (*repository.StopOrder, error) {
	if m.setTakeProfitFunc != nil {
		return m.setTakeProfitFunc(symbol, position, targetPrice)
	}
	return nil, nil
}

func (m *mockStopLossService) SetStopLossTakeProfit(symbol string, position float64, stopPrice, targetPrice float64) (*repository.StopOrderPair, error) {
	return nil, nil
}

func (m *mockStopLossService) SetTrailingStop(symbol string, position float64, trailPercent float64) (*repository.TrailingStopOrder, error) {
	return nil, nil
}

func (m *mockStopLossService) CancelStopOrder(orderID string) error {
	if m.cancelStopOrderFunc != nil {
		return m.cancelStopOrderFunc(orderID)
	}
	return nil
}

func (m *mockStopLossService) GetActiveStopOrders(symbol string) ([]*repository.StopOrder, error) {
	if m.getActiveStopOrdersFunc != nil {
		return m.getActiveStopOrdersFunc(symbol)
	}
	return nil, nil
}

func (m *mockStopLossService) UpdateTrailingStop(orderID string, newTrailPercent float64) error {
	return nil
}

// TestHandleConditionalOrder tests the condorder command handler
func TestHandleConditionalOrder(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockCondService := &mockConditionalOrderService{
			createConditionalOrderFunc: func(request *repository.ConditionalOrderRequest) (*repository.ConditionalOrder, error) {
				return &repository.ConditionalOrder{
					OrderID:  "cond-12345",
					Symbol:   request.Symbol,
					Side:     request.Side,
					Type:     request.Type,
					Quantity: request.Quantity,
					Status:   repository.ConditionalOrderStatusPending,
					TriggerCondition: request.TriggerCondition,
				}, nil
			},
		}

		cli := NewCLI(&mockTradingService{}, &mockMarketDataService{}, mockCondService, &mockStopLossService{}, &mockLogger{})

		var buf bytes.Buffer
		cli.writer = &buf

		err := cli.handleConditionalOrder([]string{"BTCUSDT", "BUY", "0.001", "PRICE", ">=", "50000"})
		if err != nil {
			t.Errorf("handleConditionalOrder() unexpected error: %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "cond-12345") {
			t.Errorf("handleConditionalOrder() output should contain order ID")
		}
	})

	t.Run("missing arguments", func(t *testing.T) {
		cli := NewCLI(&mockTradingService{}, &mockMarketDataService{}, &mockConditionalOrderService{}, &mockStopLossService{}, &mockLogger{})

		err := cli.handleConditionalOrder([]string{"BTCUSDT", "BUY"})
		if err == nil {
			t.Errorf("handleConditionalOrder() expected error for missing arguments")
		}
	})

	t.Run("invalid quantity", func(t *testing.T) {
		cli := NewCLI(&mockTradingService{}, &mockMarketDataService{}, &mockConditionalOrderService{}, &mockStopLossService{}, &mockLogger{})

		err := cli.handleConditionalOrder([]string{"BTCUSDT", "BUY", "invalid", "PRICE", ">=", "50000"})
		if err == nil {
			t.Errorf("handleConditionalOrder() expected error for invalid quantity")
		}
	})

	t.Run("invalid trigger value", func(t *testing.T) {
		cli := NewCLI(&mockTradingService{}, &mockMarketDataService{}, &mockConditionalOrderService{}, &mockStopLossService{}, &mockLogger{})

		err := cli.handleConditionalOrder([]string{"BTCUSDT", "BUY", "0.001", "PRICE", ">=", "invalid"})
		if err == nil {
			t.Errorf("handleConditionalOrder() expected error for invalid trigger value")
		}
	})
}

// TestHandleConditionalOrders tests the condorders command handler
func TestHandleConditionalOrders(t *testing.T) {
	t.Run("success with orders", func(t *testing.T) {
		mockCondService := &mockConditionalOrderService{
			getActiveConditionalOrdersFunc: func() ([]*repository.ConditionalOrder, error) {
				return []*repository.ConditionalOrder{
					{
						OrderID:  "cond-1",
						Symbol:   "BTCUSDT",
						Side:     api.OrderSideBuy,
						Type:     api.OrderTypeMarket,
						Quantity: 0.001,
						Status:   repository.ConditionalOrderStatusPending,
						TriggerCondition: &repository.TriggerCondition{
							Type:     repository.TriggerTypePrice,
							Operator: repository.OperatorGreaterEqual,
							Value:    50000,
						},
					},
				}, nil
			},
		}

		cli := NewCLI(&mockTradingService{}, &mockMarketDataService{}, mockCondService, &mockStopLossService{}, &mockLogger{})

		var buf bytes.Buffer
		cli.writer = &buf

		err := cli.handleConditionalOrders([]string{})
		if err != nil {
			t.Errorf("handleConditionalOrders() unexpected error: %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "cond-1") {
			t.Errorf("handleConditionalOrders() output should contain order ID")
		}
		if !strings.Contains(output, "BTCUSDT") {
			t.Errorf("handleConditionalOrders() output should contain symbol")
		}
	})

	t.Run("empty list", func(t *testing.T) {
		mockCondService := &mockConditionalOrderService{
			getActiveConditionalOrdersFunc: func() ([]*repository.ConditionalOrder, error) {
				return []*repository.ConditionalOrder{}, nil
			},
		}

		cli := NewCLI(&mockTradingService{}, &mockMarketDataService{}, mockCondService, &mockStopLossService{}, &mockLogger{})

		var buf bytes.Buffer
		cli.writer = &buf

		err := cli.handleConditionalOrders([]string{})
		if err != nil {
			t.Errorf("handleConditionalOrders() unexpected error: %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "No active conditional orders") {
			t.Errorf("handleConditionalOrders() should show 'No active conditional orders' for empty list")
		}
	})
}

// TestHandleCancelConditionalOrder tests the cancelcond command handler
func TestHandleCancelConditionalOrder(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockCondService := &mockConditionalOrderService{
			cancelConditionalOrderFunc: func(orderID string) error {
				return nil
			},
		}

		cli := NewCLI(&mockTradingService{}, &mockMarketDataService{}, mockCondService, &mockStopLossService{}, &mockLogger{})

		var buf bytes.Buffer
		cli.writer = &buf

		err := cli.handleCancelConditionalOrder([]string{"cond-12345"})
		if err != nil {
			t.Errorf("handleCancelConditionalOrder() unexpected error: %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "canceled successfully") {
			t.Errorf("handleCancelConditionalOrder() output should contain success message")
		}
	})

	t.Run("missing argument", func(t *testing.T) {
		cli := NewCLI(&mockTradingService{}, &mockMarketDataService{}, &mockConditionalOrderService{}, &mockStopLossService{}, &mockLogger{})

		err := cli.handleCancelConditionalOrder([]string{})
		if err == nil {
			t.Errorf("handleCancelConditionalOrder() expected error for missing argument")
		}
	})
}

// TestHandleStopLoss tests the stoploss command handler
func TestHandleStopLoss(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockStopService := &mockStopLossService{
			setStopLossFunc: func(symbol string, position float64, stopPrice float64) (*repository.StopOrder, error) {
				return &repository.StopOrder{
					OrderID:   "sl-12345",
					Symbol:    symbol,
					Position:  position,
					StopPrice: stopPrice,
					Type:      repository.StopOrderTypeStopLoss,
					Status:    repository.StopOrderStatusActive,
				}, nil
			},
		}

		cli := NewCLI(&mockTradingService{}, &mockMarketDataService{}, &mockConditionalOrderService{}, mockStopService, &mockLogger{})

		var buf bytes.Buffer
		cli.writer = &buf

		err := cli.handleStopLoss([]string{"BTCUSDT", "0.001", "49000"})
		if err != nil {
			t.Errorf("handleStopLoss() unexpected error: %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "sl-12345") {
			t.Errorf("handleStopLoss() output should contain order ID")
		}
		if !strings.Contains(output, "STOP_LOSS") {
			t.Errorf("handleStopLoss() output should contain order type")
		}
	})

	t.Run("missing arguments", func(t *testing.T) {
		cli := NewCLI(&mockTradingService{}, &mockMarketDataService{}, &mockConditionalOrderService{}, &mockStopLossService{}, &mockLogger{})

		err := cli.handleStopLoss([]string{"BTCUSDT"})
		if err == nil {
			t.Errorf("handleStopLoss() expected error for missing arguments")
		}
	})

	t.Run("invalid position", func(t *testing.T) {
		cli := NewCLI(&mockTradingService{}, &mockMarketDataService{}, &mockConditionalOrderService{}, &mockStopLossService{}, &mockLogger{})

		err := cli.handleStopLoss([]string{"BTCUSDT", "invalid", "49000"})
		if err == nil {
			t.Errorf("handleStopLoss() expected error for invalid position")
		}
	})
}

// TestHandleTakeProfit tests the takeprofit command handler
func TestHandleTakeProfit(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockStopService := &mockStopLossService{
			setTakeProfitFunc: func(symbol string, position float64, targetPrice float64) (*repository.StopOrder, error) {
				return &repository.StopOrder{
					OrderID:   "tp-12345",
					Symbol:    symbol,
					Position:  position,
					StopPrice: targetPrice,
					Type:      repository.StopOrderTypeTakeProfit,
					Status:    repository.StopOrderStatusActive,
				}, nil
			},
		}

		cli := NewCLI(&mockTradingService{}, &mockMarketDataService{}, &mockConditionalOrderService{}, mockStopService, &mockLogger{})

		var buf bytes.Buffer
		cli.writer = &buf

		err := cli.handleTakeProfit([]string{"BTCUSDT", "0.001", "51000"})
		if err != nil {
			t.Errorf("handleTakeProfit() unexpected error: %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "tp-12345") {
			t.Errorf("handleTakeProfit() output should contain order ID")
		}
		if !strings.Contains(output, "TAKE_PROFIT") {
			t.Errorf("handleTakeProfit() output should contain order type")
		}
	})

	t.Run("missing arguments", func(t *testing.T) {
		cli := NewCLI(&mockTradingService{}, &mockMarketDataService{}, &mockConditionalOrderService{}, &mockStopLossService{}, &mockLogger{})

		err := cli.handleTakeProfit([]string{"BTCUSDT", "0.001"})
		if err == nil {
			t.Errorf("handleTakeProfit() expected error for missing arguments")
		}
	})
}

// TestHandleStopOrders tests the stoporders command handler
func TestHandleStopOrders(t *testing.T) {
	t.Run("success with orders", func(t *testing.T) {
		mockStopService := &mockStopLossService{
			getActiveStopOrdersFunc: func(symbol string) ([]*repository.StopOrder, error) {
				return []*repository.StopOrder{
					{
						OrderID:   "sl-1",
						Symbol:    symbol,
						Position:  0.001,
						StopPrice: 49000,
						Type:      repository.StopOrderTypeStopLoss,
						Status:    repository.StopOrderStatusActive,
					},
					{
						OrderID:   "tp-1",
						Symbol:    symbol,
						Position:  0.001,
						StopPrice: 51000,
						Type:      repository.StopOrderTypeTakeProfit,
						Status:    repository.StopOrderStatusActive,
					},
				}, nil
			},
		}

		cli := NewCLI(&mockTradingService{}, &mockMarketDataService{}, &mockConditionalOrderService{}, mockStopService, &mockLogger{})

		var buf bytes.Buffer
		cli.writer = &buf

		err := cli.handleStopOrders([]string{"BTCUSDT"})
		if err != nil {
			t.Errorf("handleStopOrders() unexpected error: %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "sl-1") {
			t.Errorf("handleStopOrders() output should contain stop loss order ID")
		}
		if !strings.Contains(output, "tp-1") {
			t.Errorf("handleStopOrders() output should contain take profit order ID")
		}
	})

	t.Run("empty list", func(t *testing.T) {
		mockStopService := &mockStopLossService{
			getActiveStopOrdersFunc: func(symbol string) ([]*repository.StopOrder, error) {
				return []*repository.StopOrder{}, nil
			},
		}

		cli := NewCLI(&mockTradingService{}, &mockMarketDataService{}, &mockConditionalOrderService{}, mockStopService, &mockLogger{})

		var buf bytes.Buffer
		cli.writer = &buf

		err := cli.handleStopOrders([]string{"BTCUSDT"})
		if err != nil {
			t.Errorf("handleStopOrders() unexpected error: %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "No active stop orders") {
			t.Errorf("handleStopOrders() should show 'No active stop orders' for empty list")
		}
	})

	t.Run("missing argument", func(t *testing.T) {
		cli := NewCLI(&mockTradingService{}, &mockMarketDataService{}, &mockConditionalOrderService{}, &mockStopLossService{}, &mockLogger{})

		err := cli.handleStopOrders([]string{})
		if err == nil {
			t.Errorf("handleStopOrders() expected error for missing argument")
		}
	})
}

// TestHandleCancelStopOrder tests the cancelstop command handler
func TestHandleCancelStopOrder(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockStopService := &mockStopLossService{
			cancelStopOrderFunc: func(orderID string) error {
				return nil
			},
		}

		cli := NewCLI(&mockTradingService{}, &mockMarketDataService{}, &mockConditionalOrderService{}, mockStopService, &mockLogger{})

		var buf bytes.Buffer
		cli.writer = &buf

		err := cli.handleCancelStopOrder([]string{"sl-12345"})
		if err != nil {
			t.Errorf("handleCancelStopOrder() unexpected error: %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "canceled successfully") {
			t.Errorf("handleCancelStopOrder() output should contain success message")
		}
	})

	t.Run("missing argument", func(t *testing.T) {
		cli := NewCLI(&mockTradingService{}, &mockMarketDataService{}, &mockConditionalOrderService{}, &mockStopLossService{}, &mockLogger{})

		err := cli.handleCancelStopOrder([]string{})
		if err == nil {
			t.Errorf("handleCancelStopOrder() expected error for missing argument")
		}
	})
}

// TestFormatConditionalOrder tests conditional order formatting
func TestFormatConditionalOrder(t *testing.T) {
	cli := NewCLI(&mockTradingService{}, &mockMarketDataService{}, &mockConditionalOrderService{}, &mockStopLossService{}, &mockLogger{})

	var buf bytes.Buffer
	cli.writer = &buf

	order := &repository.ConditionalOrder{
		OrderID:  "cond-12345",
		Symbol:   "BTCUSDT",
		Side:     api.OrderSideBuy,
		Type:     api.OrderTypeMarket,
		Quantity: 0.001,
		Status:   repository.ConditionalOrderStatusPending,
		TriggerCondition: &repository.TriggerCondition{
			Type:     repository.TriggerTypePrice,
			Operator: repository.OperatorGreaterEqual,
			Value:    50000,
		},
	}

	cli.formatConditionalOrder(order)

	output := buf.String()

	expectedFields := []string{
		"cond-12345",
		"BTCUSDT",
		"BUY",
		"MARKET",
		"0.001",
		"PENDING",
		"PRICE",
		">=",
		"50000",
	}

	for _, field := range expectedFields {
		if !strings.Contains(output, field) {
			t.Errorf("formatConditionalOrder() output should contain %s", field)
		}
	}
}

// TestFormatStopOrder tests stop order formatting
func TestFormatStopOrder(t *testing.T) {
	cli := NewCLI(&mockTradingService{}, &mockMarketDataService{}, &mockConditionalOrderService{}, &mockStopLossService{}, &mockLogger{})

	var buf bytes.Buffer
	cli.writer = &buf

	order := &repository.StopOrder{
		OrderID:   "sl-12345",
		Symbol:    "BTCUSDT",
		Position:  0.001,
		StopPrice: 49000,
		Type:      repository.StopOrderTypeStopLoss,
		Status:    repository.StopOrderStatusActive,
	}

	cli.formatStopOrder(order)

	output := buf.String()

	expectedFields := []string{
		"sl-12345",
		"BTCUSDT",
		"STOP_LOSS",
		"0.001",
		"49000",
		"ACTIVE",
	}

	for _, field := range expectedFields {
		if !strings.Contains(output, field) {
			t.Errorf("formatStopOrder() output should contain %s", field)
		}
	}
}
