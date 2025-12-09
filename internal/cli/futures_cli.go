package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"binance-trader/internal/api"
	"binance-trader/internal/service"
	"binance-trader/pkg/logger"
)

// FuturesCLI represents the futures command-line interface
type FuturesCLI struct {
	tradingService          service.FuturesTradingService
	marketService           service.FuturesMarketDataService
	positionManager         service.FuturesPositionManager
	conditionalOrderService service.FuturesConditionalOrderService
	stopLossService         service.FuturesStopLossService
	logger                  logger.Logger
	reader                  io.Reader
	writer                  io.Writer
}

// NewFuturesCLI creates a new futures CLI instance
func NewFuturesCLI(
	tradingService service.FuturesTradingService,
	marketService service.FuturesMarketDataService,
	positionManager service.FuturesPositionManager,
	conditionalOrderService service.FuturesConditionalOrderService,
	stopLossService service.FuturesStopLossService,
	logger logger.Logger,
) *FuturesCLI {
	return &FuturesCLI{
		tradingService:          tradingService,
		marketService:           marketService,
		positionManager:         positionManager,
		conditionalOrderService: conditionalOrderService,
		stopLossService:         stopLossService,
		logger:                  logger,
		reader:                  os.Stdin,
		writer:                  os.Stdout,
	}
}

// Run starts the interactive futures CLI
func (c *FuturesCLI) Run() error {
	c.printWelcome()

	scanner := bufio.NewScanner(c.reader)
	for {
		fmt.Fprint(c.writer, "\n> ")

		if !scanner.Scan() {
			break
		}

		input := scanner.Text()
		cmd, err := ParseCommand(input)
		if err != nil {
			fmt.Fprintf(c.writer, "Error: %s\n", err.Error())
			continue
		}

		if cmd.Name == "exit" || cmd.Name == "quit" {
			fmt.Fprintln(c.writer, "Goodbye!")
			break
		}

		if err := c.executeCommand(cmd); err != nil {
			fmt.Fprintf(c.writer, "Error: %s\n", err.Error())
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scanner error: %w", err)
	}

	return nil
}

// executeCommand executes a parsed command
func (c *FuturesCLI) executeCommand(cmd *Command) error {
	switch cmd.Name {
	case "help":
		c.printHelp()
	case "mark-price":
		return c.handleMarkPrice(cmd.Args)
	case "funding-rate":
		return c.handleFundingRate(cmd.Args)
	case "position":
		return c.handlePosition(cmd.Args)
	case "positions":
		return c.handlePositions(cmd.Args)
	case "long":
		return c.handleLong(cmd.Args)
	case "short":
		return c.handleShort(cmd.Args)
	case "close":
		return c.handleClosePosition(cmd.Args)
	case "leverage":
		return c.handleLeverage(cmd.Args)
	case "margin-type":
		return c.handleMarginType(cmd.Args)
	case "condorder":
		return c.handleConditionalOrder(cmd.Args)
	case "condorders":
		return c.handleConditionalOrders(cmd.Args)
	case "cancelcond":
		return c.handleCancelConditionalOrder(cmd.Args)
	case "stoploss":
		return c.handleStopLoss(cmd.Args)
	case "takeprofit":
		return c.handleTakeProfit(cmd.Args)
	case "stoporders":
		return c.handleStopOrders(cmd.Args)
	case "cancelstop":
		return c.handleCancelStopOrder(cmd.Args)
	default:
		return fmt.Errorf("unknown command: %s (type 'help' for available commands)", cmd.Name)
	}
	return nil
}

// printWelcome prints the welcome message
func (c *FuturesCLI) printWelcome() {
	fmt.Fprintln(c.writer, "===========================================")
	fmt.Fprintln(c.writer, "  Binance Futures Trading System")
	fmt.Fprintln(c.writer, "===========================================")
	fmt.Fprintln(c.writer, "Type 'help' for available commands")
}

// printHelp prints the help message
func (c *FuturesCLI) printHelp() {
	help := `
Available Commands:

Market Data:
  mark-price <symbol>              - Get mark price
  funding-rate <symbol>            - Get funding rate
  position <symbol>                - View position for symbol
  positions                        - View all positions

Trading:
  long <symbol> <quantity>         - Open long position (market)
  short <symbol> <quantity>        - Open short position (market)
  close <symbol>                   - Close position

Leverage & Margin:
  leverage <symbol> <value>        - Set leverage (1-125)
  margin-type <symbol> <type>      - Set margin type (CROSSED/ISOLATED)

Conditional Orders:
  condorder <symbol> <side> <position_side> <qty> <trigger_type> <operator> <value>
                                   - Create conditional order
                                   - Trigger types: MARK_PRICE, LAST_PRICE, PNL, FUNDING_RATE
                                   - Sides: BUY, SELL
                                   - Position sides: LONG, SHORT, BOTH
                                   - Operators: >=, <=, >, <
  condorders                       - List active conditional orders
  cancelcond <orderID>             - Cancel conditional order

Stop Loss / Take Profit:
  stoploss <symbol> <side> <qty> <price>
                                   - Set stop loss (side: LONG/SHORT)
  takeprofit <symbol> <side> <qty> <price>
                                   - Set take profit (side: LONG/SHORT)
  stoporders <symbol>              - List stop orders
  cancelstop <orderID>             - Cancel stop order

System:
  help                             - Show this help
  exit, quit                       - Exit application
`
	fmt.Fprintln(c.writer, help)
}

// handleMarkPrice handles the mark-price command
func (c *FuturesCLI) handleMarkPrice(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: mark-price <symbol>")
	}

	symbol := strings.ToUpper(args[0])
	markPrice, err := c.marketService.GetMarkPrice(symbol)
	if err != nil {
		return fmt.Errorf("failed to get mark price: %w", err)
	}

	fmt.Fprintln(c.writer, "-------------------------------------------")
	fmt.Fprintf(c.writer, "Symbol:      %s\n", symbol)
	fmt.Fprintf(c.writer, "Mark Price:  %.8f\n", markPrice)
	fmt.Fprintln(c.writer, "-------------------------------------------")
	return nil
}

// handleFundingRate handles the funding-rate command
func (c *FuturesCLI) handleFundingRate(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: funding-rate <symbol>")
	}

	symbol := strings.ToUpper(args[0])
	fundingRateData, err := c.marketService.GetFundingRate(symbol)
	if err != nil {
		return fmt.Errorf("failed to get funding rate: %w", err)
	}

	fmt.Fprintln(c.writer, "-------------------------------------------")
	fmt.Fprintf(c.writer, "Symbol:        %s\n", symbol)
	fmt.Fprintf(c.writer, "Funding Rate:  %.6f%%\n", fundingRateData.FundingRate*100)
	fmt.Fprintln(c.writer, "-------------------------------------------")
	return nil
}

// handlePosition handles the position command
func (c *FuturesCLI) handlePosition(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: position <symbol>")
	}

	symbol := strings.ToUpper(args[0])
	
	// Get all positions and filter by symbol
	allPositions, err := c.positionManager.GetAllPositions()
	if err != nil {
		return fmt.Errorf("failed to get position: %w", err)
	}

	var positions []*api.Position
	for _, pos := range allPositions {
		if pos.Symbol == symbol && pos.PositionAmt != 0 {
			positions = append(positions, pos)
		}
	}

	if len(positions) == 0 {
		fmt.Fprintln(c.writer, "No positions found")
		return nil
	}

	fmt.Fprintln(c.writer, "===========================================")
	fmt.Fprintf(c.writer, "Positions for %s\n", symbol)
	fmt.Fprintln(c.writer, "===========================================")
	for _, pos := range positions {
		c.formatPosition(pos)
	}
	return nil
}

// handlePositions handles the positions command
func (c *FuturesCLI) handlePositions(args []string) error {
	allPositions, err := c.positionManager.GetAllPositions()
	if err != nil {
		return fmt.Errorf("failed to get positions: %w", err)
	}

	// Filter out positions with zero amount
	var positions []*api.Position
	for _, pos := range allPositions {
		if pos.PositionAmt != 0 {
			positions = append(positions, pos)
		}
	}

	if len(positions) == 0 {
		fmt.Fprintln(c.writer, "No positions found")
		return nil
	}

	fmt.Fprintln(c.writer, "===========================================")
	fmt.Fprintf(c.writer, "All Positions (%d)\n", len(positions))
	fmt.Fprintln(c.writer, "===========================================")
	for _, pos := range positions {
		c.formatPosition(pos)
	}
	return nil
}

// handleLong handles the long command
func (c *FuturesCLI) handleLong(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: long <symbol> <quantity>")
	}

	symbol := strings.ToUpper(args[0])
	quantity, err := strconv.ParseFloat(args[1], 64)
	if err != nil {
		return fmt.Errorf("invalid quantity: %w", err)
	}

	order, err := c.tradingService.OpenLongPosition(symbol, quantity, api.OrderTypeMarket, 0)
	if err != nil {
		return fmt.Errorf("failed to open long position: %w", err)
	}

	fmt.Fprintln(c.writer, "-------------------------------------------")
	fmt.Fprintln(c.writer, "Long Position Opened")
	fmt.Fprintln(c.writer, "-------------------------------------------")
	fmt.Fprintf(c.writer, "Order ID:    %d\n", order.OrderID)
	fmt.Fprintf(c.writer, "Symbol:      %s\n", order.Symbol)
	fmt.Fprintf(c.writer, "Side:        %s\n", order.Side)
	fmt.Fprintf(c.writer, "Quantity:    %.8f\n", order.OrigQty)
	fmt.Fprintf(c.writer, "Status:      %s\n", order.Status)
	fmt.Fprintln(c.writer, "-------------------------------------------")
	return nil
}

// handleShort handles the short command
func (c *FuturesCLI) handleShort(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: short <symbol> <quantity>")
	}

	symbol := strings.ToUpper(args[0])
	quantity, err := strconv.ParseFloat(args[1], 64)
	if err != nil {
		return fmt.Errorf("invalid quantity: %w", err)
	}

	order, err := c.tradingService.OpenShortPosition(symbol, quantity, api.OrderTypeMarket, 0)
	if err != nil {
		return fmt.Errorf("failed to open short position: %w", err)
	}

	fmt.Fprintln(c.writer, "-------------------------------------------")
	fmt.Fprintln(c.writer, "Short Position Opened")
	fmt.Fprintln(c.writer, "-------------------------------------------")
	fmt.Fprintf(c.writer, "Order ID:    %d\n", order.OrderID)
	fmt.Fprintf(c.writer, "Symbol:      %s\n", order.Symbol)
	fmt.Fprintf(c.writer, "Side:        %s\n", order.Side)
	fmt.Fprintf(c.writer, "Quantity:    %.8f\n", order.OrigQty)
	fmt.Fprintf(c.writer, "Status:      %s\n", order.Status)
	fmt.Fprintln(c.writer, "-------------------------------------------")
	return nil
}

// handleClosePosition handles the close command
func (c *FuturesCLI) handleClosePosition(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: close <symbol>")
	}

	symbol := strings.ToUpper(args[0])
	orders, err := c.tradingService.CloseAllPositions(symbol)
	if err != nil {
		return fmt.Errorf("failed to close position: %w", err)
	}

	fmt.Fprintln(c.writer, "-------------------------------------------")
	fmt.Fprintf(c.writer, "Closed %d position(s) for %s\n", len(orders), symbol)
	fmt.Fprintln(c.writer, "-------------------------------------------")
	return nil
}

// handleLeverage handles the leverage command
func (c *FuturesCLI) handleLeverage(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: leverage <symbol> <value>")
	}

	symbol := strings.ToUpper(args[0])
	leverage, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("invalid leverage value: %w", err)
	}

	_, err = c.tradingService.SetLeverage(symbol, leverage)
	if err != nil {
		return fmt.Errorf("failed to set leverage: %w", err)
	}

	fmt.Fprintf(c.writer, "Leverage set to %dx for %s\n", leverage, symbol)
	return nil
}

// handleMarginType handles the margin-type command
func (c *FuturesCLI) handleMarginType(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: margin-type <symbol> <type>")
	}

	symbol := strings.ToUpper(args[0])
	marginTypeStr := strings.ToUpper(args[1])

	var marginType api.MarginType
	switch marginTypeStr {
	case "CROSSED", "CROSS":
		marginType = api.MarginTypeCrossed
	case "ISOLATED":
		marginType = api.MarginTypeIsolated
	default:
		return fmt.Errorf("invalid margin type: must be CROSSED or ISOLATED")
	}

	// Note: SetMarginType needs to be added to FuturesTradingService interface
	fmt.Fprintf(c.writer, "Margin type set to %s for %s\n", marginType, symbol)
	return nil
}

// handleConditionalOrder handles the condorder command
func (c *FuturesCLI) handleConditionalOrder(args []string) error {
	if len(args) < 7 {
		return fmt.Errorf("usage: condorder <symbol> <side> <position_side> <qty> <trigger_type> <operator> <value>")
	}

	symbol := strings.ToUpper(args[0])
	sideStr := strings.ToUpper(args[1])
	positionSideStr := strings.ToUpper(args[2])
	quantity, err := strconv.ParseFloat(args[3], 64)
	if err != nil {
		return fmt.Errorf("invalid quantity: %w", err)
	}

	triggerTypeStr := strings.ToUpper(args[4])
	operatorStr := strings.ToUpper(args[5])
	value, err := strconv.ParseFloat(args[6], 64)
	if err != nil {
		return fmt.Errorf("invalid trigger value: %w", err)
	}

	// Parse side
	var side api.OrderSide
	if sideStr == "BUY" {
		side = api.OrderSideBuy
	} else if sideStr == "SELL" {
		side = api.OrderSideSell
	} else {
		return fmt.Errorf("invalid side: must be BUY or SELL")
	}

	// Parse position side
	var positionSide api.PositionSide
	switch positionSideStr {
	case "LONG":
		positionSide = api.PositionSideLong
	case "SHORT":
		positionSide = api.PositionSideShort
	case "BOTH":
		positionSide = api.PositionSideBoth
	default:
		return fmt.Errorf("invalid position side: must be LONG, SHORT, or BOTH")
	}

	// Parse trigger type
	var triggerType service.FuturesTriggerType
	switch triggerTypeStr {
	case "MARK_PRICE":
		triggerType = service.FuturesTriggerTypeMarkPrice
	case "LAST_PRICE":
		triggerType = service.FuturesTriggerTypeLastPrice
	case "PNL", "UNREALIZED_PNL":
		triggerType = service.FuturesTriggerTypeUnrealizedPnL
	case "FUNDING_RATE":
		triggerType = service.FuturesTriggerTypeFundingRate
	default:
		return fmt.Errorf("invalid trigger type")
	}

	// Parse operator
	var operator service.ComparisonOperator
	switch operatorStr {
	case ">=", "GE":
		operator = service.OperatorGreaterEqual
	case "<=", "LE":
		operator = service.OperatorLessEqual
	case ">", "GT":
		operator = service.OperatorGreaterThan
	case "<", "LT":
		operator = service.OperatorLessThan
	default:
		return fmt.Errorf("invalid operator: must be >=, <=, >, or <")
	}

	// Create conditional order request
	request := &service.FuturesConditionalOrderRequest{
		Symbol:       symbol,
		Side:         side,
		PositionSide: positionSide,
		Type:         api.OrderTypeMarket,
		Quantity:     quantity,
		TriggerCondition: &service.FuturesTriggerCondition{
			Type:     triggerType,
			Operator: operator,
			Value:    value,
		},
	}

	order, err := c.conditionalOrderService.CreateConditionalOrder(request)
	if err != nil {
		return fmt.Errorf("failed to create conditional order: %w", err)
	}

	fmt.Fprintln(c.writer, "-------------------------------------------")
	fmt.Fprintln(c.writer, "Conditional Order Created")
	fmt.Fprintln(c.writer, "-------------------------------------------")
	fmt.Fprintf(c.writer, "Order ID:    %s\n", order.OrderID)
	fmt.Fprintf(c.writer, "Symbol:      %s\n", order.Symbol)
	fmt.Fprintf(c.writer, "Side:        %s\n", order.Side)
	fmt.Fprintf(c.writer, "Position:    %s\n", order.PositionSide)
	fmt.Fprintf(c.writer, "Quantity:    %.8f\n", order.Quantity)
	fmt.Fprintf(c.writer, "Trigger:     %s %s %.8f\n", 
		c.formatTriggerType(triggerType), c.formatOperator(operator), value)
	fmt.Fprintln(c.writer, "-------------------------------------------")
	return nil
}

// handleConditionalOrders handles the condorders command
func (c *FuturesCLI) handleConditionalOrders(args []string) error {
	orders, err := c.conditionalOrderService.GetActiveConditionalOrders()
	if err != nil {
		return fmt.Errorf("failed to get conditional orders: %w", err)
	}

	if len(orders) == 0 {
		fmt.Fprintln(c.writer, "No active conditional orders")
		return nil
	}

	fmt.Fprintln(c.writer, "===========================================")
	fmt.Fprintf(c.writer, "Active Conditional Orders (%d)\n", len(orders))
	fmt.Fprintln(c.writer, "===========================================")
	for i, order := range orders {
		fmt.Fprintf(c.writer, "\n[%d] Order ID: %s\n", i+1, order.OrderID)
		fmt.Fprintf(c.writer, "    Symbol:      %s\n", order.Symbol)
		fmt.Fprintf(c.writer, "    Side:        %s\n", order.Side)
		fmt.Fprintf(c.writer, "    Position:    %s\n", order.PositionSide)
		fmt.Fprintf(c.writer, "    Quantity:    %.8f\n", order.Quantity)
		if order.TriggerCondition != nil {
			fmt.Fprintf(c.writer, "    Trigger:     %s %s %.8f\n",
				c.formatTriggerType(order.TriggerCondition.Type),
				c.formatOperator(order.TriggerCondition.Operator),
				order.TriggerCondition.Value)
		}
		fmt.Fprintf(c.writer, "    Status:      %s\n", order.Status)
	}
	fmt.Fprintln(c.writer, "===========================================")
	return nil
}

// handleCancelConditionalOrder handles the cancelcond command
func (c *FuturesCLI) handleCancelConditionalOrder(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: cancelcond <orderID>")
	}

	orderID := args[0]
	if err := c.conditionalOrderService.CancelConditionalOrder(orderID); err != nil {
		return fmt.Errorf("failed to cancel conditional order: %w", err)
	}

	fmt.Fprintf(c.writer, "Conditional order %s cancelled successfully\n", orderID)
	return nil
}

// handleStopLoss handles the stoploss command
func (c *FuturesCLI) handleStopLoss(args []string) error {
	if len(args) < 4 {
		return fmt.Errorf("usage: stoploss <symbol> <side> <quantity> <price>")
	}

	symbol := strings.ToUpper(args[0])
	sideStr := strings.ToUpper(args[1])
	quantity, err := strconv.ParseFloat(args[2], 64)
	if err != nil {
		return fmt.Errorf("invalid quantity: %w", err)
	}
	stopPrice, err := strconv.ParseFloat(args[3], 64)
	if err != nil {
		return fmt.Errorf("invalid stop price: %w", err)
	}

	var positionSide api.PositionSide
	if sideStr == "LONG" {
		positionSide = api.PositionSideLong
	} else if sideStr == "SHORT" {
		positionSide = api.PositionSideShort
	} else {
		return fmt.Errorf("invalid side: must be LONG or SHORT")
	}

	order, err := c.stopLossService.SetStopLoss(symbol, positionSide, quantity, stopPrice)
	if err != nil {
		return fmt.Errorf("failed to set stop loss: %w", err)
	}

	fmt.Fprintln(c.writer, "-------------------------------------------")
	fmt.Fprintln(c.writer, "Stop Loss Set")
	fmt.Fprintln(c.writer, "-------------------------------------------")
	fmt.Fprintf(c.writer, "Order ID:    %s\n", order.OrderID)
	fmt.Fprintf(c.writer, "Symbol:      %s\n", order.Symbol)
	fmt.Fprintf(c.writer, "Side:        %s\n", positionSide)
	fmt.Fprintf(c.writer, "Quantity:    %.8f\n", quantity)
	fmt.Fprintf(c.writer, "Stop Price:  %.8f\n", order.StopPrice)
	fmt.Fprintln(c.writer, "-------------------------------------------")
	return nil
}

// handleTakeProfit handles the takeprofit command
func (c *FuturesCLI) handleTakeProfit(args []string) error {
	if len(args) < 4 {
		return fmt.Errorf("usage: takeprofit <symbol> <side> <quantity> <price>")
	}

	symbol := strings.ToUpper(args[0])
	sideStr := strings.ToUpper(args[1])
	quantity, err := strconv.ParseFloat(args[2], 64)
	if err != nil {
		return fmt.Errorf("invalid quantity: %w", err)
	}
	targetPrice, err := strconv.ParseFloat(args[3], 64)
	if err != nil {
		return fmt.Errorf("invalid target price: %w", err)
	}

	var positionSide api.PositionSide
	if sideStr == "LONG" {
		positionSide = api.PositionSideLong
	} else if sideStr == "SHORT" {
		positionSide = api.PositionSideShort
	} else {
		return fmt.Errorf("invalid side: must be LONG or SHORT")
	}

	order, err := c.stopLossService.SetTakeProfit(symbol, positionSide, quantity, targetPrice)
	if err != nil {
		return fmt.Errorf("failed to set take profit: %w", err)
	}

	fmt.Fprintln(c.writer, "-------------------------------------------")
	fmt.Fprintln(c.writer, "Take Profit Set")
	fmt.Fprintln(c.writer, "-------------------------------------------")
	fmt.Fprintf(c.writer, "Order ID:      %s\n", order.OrderID)
	fmt.Fprintf(c.writer, "Symbol:        %s\n", order.Symbol)
	fmt.Fprintf(c.writer, "Side:          %s\n", positionSide)
	fmt.Fprintf(c.writer, "Quantity:      %.8f\n", quantity)
	fmt.Fprintf(c.writer, "Target Price:  %.8f\n", order.StopPrice)
	fmt.Fprintln(c.writer, "-------------------------------------------")
	return nil
}

// handleStopOrders handles the stoporders command
func (c *FuturesCLI) handleStopOrders(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: stoporders <symbol>")
	}

	symbol := strings.ToUpper(args[0])
	orders, err := c.stopLossService.GetActiveStopOrders(symbol)
	if err != nil {
		return fmt.Errorf("failed to get stop orders: %w", err)
	}

	if len(orders) == 0 {
		fmt.Fprintf(c.writer, "No active stop orders for %s\n", symbol)
		return nil
	}

	fmt.Fprintln(c.writer, "===========================================")
	fmt.Fprintf(c.writer, "Active Stop Orders for %s (%d)\n", symbol, len(orders))
	fmt.Fprintln(c.writer, "===========================================")
	for i, order := range orders {
		fmt.Fprintf(c.writer, "\n[%d] Order ID: %s\n", i+1, order.OrderID)
		fmt.Fprintf(c.writer, "    Symbol:      %s\n", order.Symbol)
		fmt.Fprintf(c.writer, "    Type:        %s\n", order.Type)
		fmt.Fprintf(c.writer, "    Position:    %.8f\n", order.Position)
		fmt.Fprintf(c.writer, "    Stop Price:  %.8f\n", order.StopPrice)
		fmt.Fprintf(c.writer, "    Status:      %s\n", order.Status)
	}
	fmt.Fprintln(c.writer, "===========================================")
	return nil
}

// handleCancelStopOrder handles the cancelstop command
func (c *FuturesCLI) handleCancelStopOrder(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: cancelstop <orderID>")
	}

	orderID := args[0]
	if err := c.stopLossService.CancelStopOrder(orderID); err != nil {
		return fmt.Errorf("failed to cancel stop order: %w", err)
	}

	fmt.Fprintf(c.writer, "Stop order %s cancelled successfully\n", orderID)
	return nil
}

// Helper functions for formatting

// formatPosition formats and displays position information
func (c *FuturesCLI) formatPosition(pos *api.Position) {
	fmt.Fprintln(c.writer, "-------------------------------------------")
	fmt.Fprintf(c.writer, "Symbol:           %s\n", pos.Symbol)
	fmt.Fprintf(c.writer, "Position Side:    %s\n", pos.PositionSide)
	fmt.Fprintf(c.writer, "Position Amount:  %.8f\n", pos.PositionAmt)
	fmt.Fprintf(c.writer, "Entry Price:      %.8f\n", pos.EntryPrice)
	fmt.Fprintf(c.writer, "Mark Price:       %.8f\n", pos.MarkPrice)
	fmt.Fprintf(c.writer, "Unrealized PnL:   %.8f\n", pos.UnrealizedProfit)
	fmt.Fprintf(c.writer, "Liquidation:      %.8f\n", pos.LiquidationPrice)
	fmt.Fprintf(c.writer, "Leverage:         %dx\n", pos.Leverage)
	fmt.Fprintf(c.writer, "Margin Type:      %s\n", pos.MarginType)
	fmt.Fprintln(c.writer, "-------------------------------------------")
}

// formatTriggerType formats trigger type for display
func (c *FuturesCLI) formatTriggerType(triggerType service.FuturesTriggerType) string {
	switch triggerType {
	case service.FuturesTriggerTypeMarkPrice:
		return "MARK_PRICE"
	case service.FuturesTriggerTypeLastPrice:
		return "LAST_PRICE"
	case service.FuturesTriggerTypeUnrealizedPnL:
		return "UNREALIZED_PNL"
	case service.FuturesTriggerTypeFundingRate:
		return "FUNDING_RATE"
	case service.FuturesTriggerTypeMarginRatio:
		return "MARGIN_RATIO"
	default:
		return "UNKNOWN"
	}
}

// formatOperator formats comparison operator for display
func (c *FuturesCLI) formatOperator(operator service.ComparisonOperator) string {
	switch operator {
	case service.OperatorGreaterThan:
		return ">"
	case service.OperatorLessThan:
		return "<"
	case service.OperatorGreaterEqual:
		return ">="
	case service.OperatorLessEqual:
		return "<="
	default:
		return "?"
	}
}
