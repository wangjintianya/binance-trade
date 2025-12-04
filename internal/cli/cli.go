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

// CLI represents the command-line interface
type CLI struct {
	tradingService service.TradingService
	marketService  service.MarketDataService
	logger         logger.Logger
	reader         io.Reader
	writer         io.Writer
}

// NewCLI creates a new CLI instance
func NewCLI(
	tradingService service.TradingService,
	marketService service.MarketDataService,
	logger logger.Logger,
) *CLI {
	return &CLI{
		tradingService: tradingService,
		marketService:  marketService,
		logger:         logger,
		reader:         os.Stdin,
		writer:         os.Stdout,
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
