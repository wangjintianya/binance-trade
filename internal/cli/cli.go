package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"binance-trader/internal/api"
	"binance-trader/internal/repository"
	"binance-trader/internal/service"
	"binance-trader/pkg/logger"
)

// CLI represents the command-line interface
type CLI struct {
	tradingService          service.TradingService
	marketService           service.MarketDataService
	conditionalOrderService service.ConditionalOrderService
	stopLossService         service.StopLossService
	logger                  logger.Logger
	reader                  io.Reader
	writer                  io.Writer
}

// NewCLI creates a new CLI instance
func NewCLI(
	tradingService service.TradingService,
	marketService service.MarketDataService,
	conditionalOrderService service.ConditionalOrderService,
	stopLossService service.StopLossService,
	logger logger.Logger,
) *CLI {
	return &CLI{
		tradingService:          tradingService,
		marketService:           marketService,
		conditionalOrderService: conditionalOrderService,
		stopLossService:         stopLossService,
		logger:                  logger,
		reader:                  os.Stdin,
		writer:                  os.Stdout,
	}
}

// Command represents a parsed command
type Command struct {
	Name string
	Args []string
}

// ParseCommand parses a command string into a Command struct
func ParseCommand(input string) (*Command, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, fmt.Errorf("empty command")
	}

	parts := strings.Fields(input)
	if len(parts) == 0 {
		return nil, fmt.Errorf("empty command")
	}

	return &Command{
		Name: strings.ToLower(parts[0]),
		Args: parts[1:],
	}, nil
}

// Run starts the interactive CLI
func (c *CLI) Run() error {
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
func (c *CLI) executeCommand(cmd *Command) error {
	switch cmd.Name {
	case "help":
		c.printHelp()
	case "price":
		return c.handlePrice(cmd.Args)
	case "balance":
		return c.handleBalance(cmd.Args)
	case "buy":
		return c.handleBuy(cmd.Args)
	case "sell":
		return c.handleSell(cmd.Args)
	case "cancel":
		return c.handleCancel(cmd.Args)
	case "status":
		return c.handleStatus(cmd.Args)
	case "orders":
		return c.handleOrders(cmd.Args)
	case "history":
		return c.handleHistory(cmd.Args)
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
func (c *CLI) printWelcome() {
	fmt.Fprintln(c.writer, "===========================================")
	fmt.Fprintln(c.writer, "  Binance Auto-Trading System")
	fmt.Fprintln(c.writer, "===========================================")
	fmt.Fprintln(c.writer, "Type 'help' for available commands")
}

// printHelp prints the help message
func (c *CLI) printHelp() {
	help := `
Available Commands:
  help                          - Show this help message
  price <symbol>                - Get current price for a symbol (e.g., price BTCUSDT)
  balance <asset>               - Get balance for an asset (e.g., balance USDT)
  buy <symbol> <quantity>       - Place market buy order (e.g., buy BTCUSDT 0.001)
  sell <symbol> <price> <qty>   - Place limit sell order (e.g., sell BTCUSDT 50000 0.001)
  cancel <orderID>              - Cancel an order (e.g., cancel 12345)
  status <orderID>              - Get order status (e.g., status 12345)
  orders                        - List all active orders
  history <symbol> <interval> <limit> - Get historical kline data (e.g., history BTCUSDT 1h 10)
  
  Conditional Orders:
  condorder <symbol> <side> <qty> <trigger_type> <operator> <value>
                                - Create conditional order (e.g., condorder BTCUSDT BUY 0.001 PRICE >= 50000)
  condorders                    - List all active conditional orders
  cancelcond <orderID>          - Cancel a conditional order
  
  Stop Loss / Take Profit:
  stoploss <symbol> <position> <stop_price>
                                - Set stop loss (e.g., stoploss BTCUSDT 0.001 49000)
  takeprofit <symbol> <position> <target_price>
                                - Set take profit (e.g., takeprofit BTCUSDT 0.001 51000)
  stoporders <symbol>           - List all active stop orders for a symbol
  cancelstop <orderID>          - Cancel a stop order
  
  exit, quit                    - Exit the application
`
	fmt.Fprintln(c.writer, help)
}

// handlePrice handles the price command
func (c *CLI) handlePrice(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: price <symbol>")
	}

	symbol := strings.ToUpper(args[0])
	price, err := c.marketService.GetCurrentPrice(symbol)
	if err != nil {
		return fmt.Errorf("failed to get price: %w", err)
	}

	c.formatPrice(symbol, price)
	return nil
}

// handleBalance handles the balance command
func (c *CLI) handleBalance(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: balance <asset>")
	}

	// This would require access to the API client
	// For now, return a message
	return fmt.Errorf("balance command requires direct API access (not yet implemented in CLI)")
}

// handleBuy handles the buy command
func (c *CLI) handleBuy(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: buy <symbol> <quantity>")
	}

	symbol := strings.ToUpper(args[0])
	quantity, err := strconv.ParseFloat(args[1], 64)
	if err != nil {
		return fmt.Errorf("invalid quantity: %w", err)
	}

	order, err := c.tradingService.PlaceMarketBuyOrder(symbol, quantity)
	if err != nil {
		return fmt.Errorf("failed to place buy order: %w", err)
	}

	c.formatOrder(order)
	return nil
}

// handleSell handles the sell command
func (c *CLI) handleSell(args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("usage: sell <symbol> <price> <quantity>")
	}

	symbol := strings.ToUpper(args[0])
	price, err := strconv.ParseFloat(args[1], 64)
	if err != nil {
		return fmt.Errorf("invalid price: %w", err)
	}

	quantity, err := strconv.ParseFloat(args[2], 64)
	if err != nil {
		return fmt.Errorf("invalid quantity: %w", err)
	}

	order, err := c.tradingService.PlaceLimitSellOrder(symbol, price, quantity)
	if err != nil {
		return fmt.Errorf("failed to place sell order: %w", err)
	}

	c.formatOrder(order)
	return nil
}

// handleCancel handles the cancel command
func (c *CLI) handleCancel(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: cancel <orderID>")
	}

	orderID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid order ID: %w", err)
	}

	if err := c.tradingService.CancelOrder(orderID); err != nil {
		return fmt.Errorf("failed to cancel order: %w", err)
	}

	fmt.Fprintf(c.writer, "Order %d canceled successfully\n", orderID)
	return nil
}

// handleStatus handles the status command
func (c *CLI) handleStatus(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: status <orderID>")
	}

	orderID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid order ID: %w", err)
	}

	status, err := c.tradingService.GetOrderStatus(orderID)
	if err != nil {
		return fmt.Errorf("failed to get order status: %w", err)
	}

	c.formatOrderStatus(status)
	return nil
}

// handleOrders handles the orders command
func (c *CLI) handleOrders(args []string) error {
	orders, err := c.tradingService.GetActiveOrders()
	if err != nil {
		return fmt.Errorf("failed to get active orders: %w", err)
	}

	c.formatOrderList(orders)
	return nil
}

// handleHistory handles the history command
func (c *CLI) handleHistory(args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("usage: history <symbol> <interval> <limit>")
	}

	symbol := strings.ToUpper(args[0])
	interval := args[1]
	limit, err := strconv.Atoi(args[2])
	if err != nil {
		return fmt.Errorf("invalid limit: %w", err)
	}

	klines, err := c.marketService.GetHistoricalData(symbol, interval, limit)
	if err != nil {
		return fmt.Errorf("failed to get historical data: %w", err)
	}

	c.formatKlines(symbol, interval, klines)
	return nil
}

// formatPrice formats and displays price information
func (c *CLI) formatPrice(symbol string, price float64) {
	fmt.Fprintln(c.writer, "-------------------------------------------")
	fmt.Fprintf(c.writer, "Symbol: %s\n", symbol)
	fmt.Fprintf(c.writer, "Price:  %.8f\n", price)
	fmt.Fprintln(c.writer, "-------------------------------------------")
}

// formatOrder formats and displays order information
func (c *CLI) formatOrder(order *api.Order) {
	fmt.Fprintln(c.writer, "-------------------------------------------")
	fmt.Fprintln(c.writer, "Order Created Successfully")
	fmt.Fprintln(c.writer, "-------------------------------------------")
	fmt.Fprintf(c.writer, "Order ID:       %d\n", order.OrderID)
	fmt.Fprintf(c.writer, "Symbol:         %s\n", order.Symbol)
	fmt.Fprintf(c.writer, "Side:           %s\n", order.Side)
	fmt.Fprintf(c.writer, "Type:           %s\n", order.Type)
	fmt.Fprintf(c.writer, "Status:         %s\n", order.Status)
	fmt.Fprintf(c.writer, "Price:          %.8f\n", order.Price)
	fmt.Fprintf(c.writer, "Quantity:       %.8f\n", order.OrigQty)
	fmt.Fprintf(c.writer, "Executed Qty:   %.8f\n", order.ExecutedQty)
	fmt.Fprintf(c.writer, "Quote Qty:      %.8f\n", order.CummulativeQuoteQty)
	fmt.Fprintln(c.writer, "-------------------------------------------")
}

// formatOrderStatus formats and displays order status information
func (c *CLI) formatOrderStatus(status *service.OrderStatus) {
	fmt.Fprintln(c.writer, "-------------------------------------------")
	fmt.Fprintln(c.writer, "Order Status")
	fmt.Fprintln(c.writer, "-------------------------------------------")
	fmt.Fprintf(c.writer, "Order ID:       %d\n", status.OrderID)
	fmt.Fprintf(c.writer, "Symbol:         %s\n", status.Symbol)
	fmt.Fprintf(c.writer, "Status:         %s\n", status.Status)
	fmt.Fprintf(c.writer, "Executed Qty:   %.8f\n", status.ExecutedQty)
	fmt.Fprintf(c.writer, "Price:          %.8f\n", status.Price)
	fmt.Fprintln(c.writer, "-------------------------------------------")
}

// formatOrderList formats and displays a list of orders
func (c *CLI) formatOrderList(orders []*api.Order) {
	if len(orders) == 0 {
		fmt.Fprintln(c.writer, "No active orders")
		return
	}

	fmt.Fprintln(c.writer, "===========================================")
	fmt.Fprintf(c.writer, "Active Orders (%d)\n", len(orders))
	fmt.Fprintln(c.writer, "===========================================")

	for i, order := range orders {
		fmt.Fprintf(c.writer, "\n[%d] Order ID: %d\n", i+1, order.OrderID)
		fmt.Fprintf(c.writer, "    Symbol:       %s\n", order.Symbol)
		fmt.Fprintf(c.writer, "    Side:         %s\n", order.Side)
		fmt.Fprintf(c.writer, "    Type:         %s\n", order.Type)
		fmt.Fprintf(c.writer, "    Status:       %s\n", order.Status)
		fmt.Fprintf(c.writer, "    Price:        %.8f\n", order.Price)
		fmt.Fprintf(c.writer, "    Quantity:     %.8f\n", order.OrigQty)
		fmt.Fprintf(c.writer, "    Executed:     %.8f\n", order.ExecutedQty)
	}

	fmt.Fprintln(c.writer, "===========================================")
}

// formatKlines formats and displays kline data
func (c *CLI) formatKlines(symbol, interval string, klines []*api.Kline) {
	if len(klines) == 0 {
		fmt.Fprintln(c.writer, "No kline data available")
		return
	}

	fmt.Fprintln(c.writer, "===========================================")
	fmt.Fprintf(c.writer, "Historical Data: %s (%s)\n", symbol, interval)
	fmt.Fprintln(c.writer, "===========================================")
	fmt.Fprintln(c.writer, "Time                Open        High        Low         Close       Volume")
	fmt.Fprintln(c.writer, "-------------------------------------------")

	for _, kline := range klines {
		fmt.Fprintf(c.writer, "%-19d %-11.8f %-11.8f %-11.8f %-11.8f %.2f\n",
			kline.OpenTime,
			kline.Open,
			kline.High,
			kline.Low,
			kline.Close,
			kline.Volume,
		)
	}

	fmt.Fprintln(c.writer, "===========================================")
}

// handleConditionalOrder handles the condorder command
func (c *CLI) handleConditionalOrder(args []string) error {
	if len(args) < 6 {
		return fmt.Errorf("usage: condorder <symbol> <side> <quantity> <trigger_type> <operator> <value>")
	}

	symbol := strings.ToUpper(args[0])
	side := strings.ToUpper(args[1])
	quantity, err := strconv.ParseFloat(args[2], 64)
	if err != nil {
		return fmt.Errorf("invalid quantity: %w", err)
	}

	triggerType := strings.ToUpper(args[3])
	operator := strings.ToUpper(args[4])
	value, err := strconv.ParseFloat(args[5], 64)
	if err != nil {
		return fmt.Errorf("invalid trigger value: %w", err)
	}

	// Parse side
	var orderSide api.OrderSide
	if side == "BUY" {
		orderSide = api.OrderSideBuy
	} else if side == "SELL" {
		orderSide = api.OrderSideSell
	} else {
		return fmt.Errorf("invalid side: must be BUY or SELL")
	}

	// Parse trigger type
	var trigType service.TriggerType
	switch triggerType {
	case "PRICE":
		trigType = service.TriggerTypePrice
	case "PRICE_CHANGE":
		trigType = service.TriggerTypePriceChangePercent
	case "VOLUME":
		trigType = service.TriggerTypeVolume
	default:
		return fmt.Errorf("invalid trigger type: must be PRICE, PRICE_CHANGE, or VOLUME")
	}

	// Parse operator
	var op service.ComparisonOperator
	switch operator {
	case ">=", "GE":
		op = service.OperatorGreaterEqual
	case "<=", "LE":
		op = service.OperatorLessEqual
	case ">", "GT":
		op = service.OperatorGreaterThan
	case "<", "LT":
		op = service.OperatorLessThan
	default:
		return fmt.Errorf("invalid operator: must be >=, <=, >, or <")
	}

	// Create trigger condition (using repository types)
	triggerCondition := &repository.TriggerCondition{
		Type:     repository.TriggerType(trigType),
		Operator: repository.ComparisonOperator(op),
		Value:    value,
	}

	// Create conditional order request
	request := &repository.ConditionalOrderRequest{
		Symbol:           symbol,
		Side:             orderSide,
		Type:             api.OrderTypeMarket,
		Quantity:         quantity,
		TriggerCondition: triggerCondition,
	}

	order, err := c.conditionalOrderService.CreateConditionalOrder(request)
	if err != nil {
		return fmt.Errorf("failed to create conditional order: %w", err)
	}

	c.formatConditionalOrder(order)
	return nil
}

// handleConditionalOrders handles the condorders command
func (c *CLI) handleConditionalOrders(args []string) error {
	orders, err := c.conditionalOrderService.GetActiveConditionalOrders()
	if err != nil {
		return fmt.Errorf("failed to get conditional orders: %w", err)
	}

	c.formatConditionalOrderList(orders)
	return nil
}

// handleCancelConditionalOrder handles the cancelcond command
func (c *CLI) handleCancelConditionalOrder(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: cancelcond <orderID>")
	}

	orderID := args[0]

	if err := c.conditionalOrderService.CancelConditionalOrder(orderID); err != nil {
		return fmt.Errorf("failed to cancel conditional order: %w", err)
	}

	fmt.Fprintf(c.writer, "Conditional order %s canceled successfully\n", orderID)
	return nil
}

// handleStopLoss handles the stoploss command
func (c *CLI) handleStopLoss(args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("usage: stoploss <symbol> <position> <stop_price>")
	}

	symbol := strings.ToUpper(args[0])
	position, err := strconv.ParseFloat(args[1], 64)
	if err != nil {
		return fmt.Errorf("invalid position: %w", err)
	}

	stopPrice, err := strconv.ParseFloat(args[2], 64)
	if err != nil {
		return fmt.Errorf("invalid stop price: %w", err)
	}

	order, err := c.stopLossService.SetStopLoss(symbol, position, stopPrice)
	if err != nil {
		return fmt.Errorf("failed to set stop loss: %w", err)
	}

	c.formatStopOrder(order)
	return nil
}

// handleTakeProfit handles the takeprofit command
func (c *CLI) handleTakeProfit(args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("usage: takeprofit <symbol> <position> <target_price>")
	}

	symbol := strings.ToUpper(args[0])
	position, err := strconv.ParseFloat(args[1], 64)
	if err != nil {
		return fmt.Errorf("invalid position: %w", err)
	}

	targetPrice, err := strconv.ParseFloat(args[2], 64)
	if err != nil {
		return fmt.Errorf("invalid target price: %w", err)
	}

	order, err := c.stopLossService.SetTakeProfit(symbol, position, targetPrice)
	if err != nil {
		return fmt.Errorf("failed to set take profit: %w", err)
	}

	c.formatStopOrder(order)
	return nil
}

// handleStopOrders handles the stoporders command
func (c *CLI) handleStopOrders(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: stoporders <symbol>")
	}

	symbol := strings.ToUpper(args[0])

	orders, err := c.stopLossService.GetActiveStopOrders(symbol)
	if err != nil {
		return fmt.Errorf("failed to get stop orders: %w", err)
	}

	c.formatStopOrderList(orders)
	return nil
}

// handleCancelStopOrder handles the cancelstop command
func (c *CLI) handleCancelStopOrder(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: cancelstop <orderID>")
	}

	orderID := args[0]

	if err := c.stopLossService.CancelStopOrder(orderID); err != nil {
		return fmt.Errorf("failed to cancel stop order: %w", err)
	}

	fmt.Fprintf(c.writer, "Stop order %s canceled successfully\n", orderID)
	return nil
}

// formatConditionalOrder formats and displays conditional order information
func (c *CLI) formatConditionalOrder(order *repository.ConditionalOrder) {
	fmt.Fprintln(c.writer, "-------------------------------------------")
	fmt.Fprintln(c.writer, "Conditional Order Created Successfully")
	fmt.Fprintln(c.writer, "-------------------------------------------")
	fmt.Fprintf(c.writer, "Order ID:       %s\n", order.OrderID)
	fmt.Fprintf(c.writer, "Symbol:         %s\n", order.Symbol)
	fmt.Fprintf(c.writer, "Side:           %s\n", order.Side)
	fmt.Fprintf(c.writer, "Type:           %s\n", order.Type)
	fmt.Fprintf(c.writer, "Quantity:       %.8f\n", order.Quantity)
	fmt.Fprintf(c.writer, "Status:         %s\n", order.Status)
	if order.TriggerCondition != nil {
		fmt.Fprintf(c.writer, "Trigger:        %s %s %.8f\n",
			c.formatTriggerType(order.TriggerCondition.Type),
			c.formatOperator(order.TriggerCondition.Operator),
			order.TriggerCondition.Value)
	}
	fmt.Fprintln(c.writer, "-------------------------------------------")
}

// formatConditionalOrderList formats and displays a list of conditional orders
func (c *CLI) formatConditionalOrderList(orders []*repository.ConditionalOrder) {
	if len(orders) == 0 {
		fmt.Fprintln(c.writer, "No active conditional orders")
		return
	}

	fmt.Fprintln(c.writer, "===========================================")
	fmt.Fprintf(c.writer, "Active Conditional Orders (%d)\n", len(orders))
	fmt.Fprintln(c.writer, "===========================================")

	for i, order := range orders {
		fmt.Fprintf(c.writer, "\n[%d] Order ID: %s\n", i+1, order.OrderID)
		fmt.Fprintf(c.writer, "    Symbol:       %s\n", order.Symbol)
		fmt.Fprintf(c.writer, "    Side:         %s\n", order.Side)
		fmt.Fprintf(c.writer, "    Type:         %s\n", order.Type)
		fmt.Fprintf(c.writer, "    Quantity:     %.8f\n", order.Quantity)
		fmt.Fprintf(c.writer, "    Status:       %s\n", order.Status)
		if order.TriggerCondition != nil {
			fmt.Fprintf(c.writer, "    Trigger:      %s %s %.8f\n",
				c.formatTriggerType(order.TriggerCondition.Type),
				c.formatOperator(order.TriggerCondition.Operator),
				order.TriggerCondition.Value)
		}
	}

	fmt.Fprintln(c.writer, "===========================================")
}

// formatStopOrder formats and displays stop order information
func (c *CLI) formatStopOrder(order *repository.StopOrder) {
	fmt.Fprintln(c.writer, "-------------------------------------------")
	fmt.Fprintln(c.writer, "Stop Order Created Successfully")
	fmt.Fprintln(c.writer, "-------------------------------------------")
	fmt.Fprintf(c.writer, "Order ID:       %s\n", order.OrderID)
	fmt.Fprintf(c.writer, "Symbol:         %s\n", order.Symbol)
	fmt.Fprintf(c.writer, "Type:           %s\n", c.formatStopOrderType(order.Type))
	fmt.Fprintf(c.writer, "Position:       %.8f\n", order.Position)
	fmt.Fprintf(c.writer, "Stop Price:     %.8f\n", order.StopPrice)
	fmt.Fprintf(c.writer, "Status:         %s\n", order.Status)
	fmt.Fprintln(c.writer, "-------------------------------------------")
}

// formatStopOrderList formats and displays a list of stop orders
func (c *CLI) formatStopOrderList(orders []*repository.StopOrder) {
	if len(orders) == 0 {
		fmt.Fprintln(c.writer, "No active stop orders")
		return
	}

	fmt.Fprintln(c.writer, "===========================================")
	fmt.Fprintf(c.writer, "Active Stop Orders (%d)\n", len(orders))
	fmt.Fprintln(c.writer, "===========================================")

	for i, order := range orders {
		fmt.Fprintf(c.writer, "\n[%d] Order ID: %s\n", i+1, order.OrderID)
		fmt.Fprintf(c.writer, "    Symbol:       %s\n", order.Symbol)
		fmt.Fprintf(c.writer, "    Type:         %s\n", c.formatStopOrderType(order.Type))
		fmt.Fprintf(c.writer, "    Position:     %.8f\n", order.Position)
		fmt.Fprintf(c.writer, "    Stop Price:   %.8f\n", order.StopPrice)
		fmt.Fprintf(c.writer, "    Status:       %s\n", order.Status)
	}

	fmt.Fprintln(c.writer, "===========================================")
}

// formatTriggerType formats trigger type for display
func (c *CLI) formatTriggerType(triggerType repository.TriggerType) string {
	switch triggerType {
	case repository.TriggerTypePrice:
		return "PRICE"
	case repository.TriggerTypePriceChangePercent:
		return "PRICE_CHANGE"
	case repository.TriggerTypeVolume:
		return "VOLUME"
	default:
		return "UNKNOWN"
	}
}

// formatOperator formats comparison operator for display
func (c *CLI) formatOperator(operator repository.ComparisonOperator) string {
	switch operator {
	case repository.OperatorGreaterThan:
		return ">"
	case repository.OperatorLessThan:
		return "<"
	case repository.OperatorGreaterEqual:
		return ">="
	case repository.OperatorLessEqual:
		return "<="
	default:
		return "?"
	}
}

// formatStopOrderType formats stop order type for display
func (c *CLI) formatStopOrderType(orderType repository.StopOrderType) string {
	switch orderType {
	case repository.StopOrderTypeStopLoss:
		return "STOP_LOSS"
	case repository.StopOrderTypeTakeProfit:
		return "TAKE_PROFIT"
	default:
		return "UNKNOWN"
	}
}
