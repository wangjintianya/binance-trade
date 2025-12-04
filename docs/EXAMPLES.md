# 使用示例 / Usage Examples

本文档提供币安自动交易系统的实际使用示例。

This document provides practical usage examples for the Binance Auto-Trading System.

## 目录 / Table of Contents

- [基础示例 / Basic Examples](#基础示例--basic-examples)
- [高级示例 / Advanced Examples](#高级示例--advanced-examples)
- [策略示例 / Strategy Examples](#策略示例--strategy-examples)
- [错误处理示例 / Error Handling Examples](#错误处理示例--error-handling-examples)

---

## 基础示例 / Basic Examples

### 示例 1: 查询价格 / Example 1: Query Price

```go
package main

import (
    "fmt"
    "log"
    "binance-trader/internal/service"
)

func main() {
    // 假设已初始化 marketService
    // Assume marketService is already initialized
    
    // 查询单个交易对价格
    // Query single trading pair price
    price, err := marketService.GetCurrentPrice("BTCUSDT")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("BTC/USDT: $%.2f\n", price)
    
    // 查询多个交易对
    // Query multiple trading pairs
    symbols := []string{"BTCUSDT", "ETHUSDT", "BNBUSDT"}
    for _, symbol := range symbols {
        price, err := marketService.GetCurrentPrice(symbol)
        if err != nil {
            log.Printf("Error getting price for %s: %v\n", symbol, err)
            continue
        }
        fmt.Printf("%s: $%.2f\n", symbol, price)
    }
}
```


### 示例 2: 简单买入 / Example 2: Simple Buy

```go
package main

import (
    "fmt"
    "log"
)

func simpleBuy() {
    symbol := "BTCUSDT"
    quantity := 0.001  // 买入 0.001 BTC
    
    // 1. 先查询当前价格
    // 1. First query current price
    price, err := marketService.GetCurrentPrice(symbol)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Current price: $%.2f\n", price)
    
    // 2. 计算预估成本
    // 2. Calculate estimated cost
    estimatedCost := price * quantity
    fmt.Printf("Estimated cost: $%.2f\n", estimatedCost)
    
    // 3. 下市价买单
    // 3. Place market buy order
    order, err := tradingService.PlaceMarketBuyOrder(symbol, quantity)
    if err != nil {
        log.Fatal(err)
    }
    
    // 4. 显示订单信息
    // 4. Display order information
    fmt.Printf("Order placed successfully!\n")
    fmt.Printf("Order ID: %d\n", order.OrderID)
    fmt.Printf("Status: %s\n", order.Status)
    fmt.Printf("Executed Qty: %.8f\n", order.ExecutedQty)
    fmt.Printf("Total Cost: $%.2f\n", order.CummulativeQuoteQty)
}
```

### 示例 3: 限价卖出 / Example 3: Limit Sell

```go
func limitSell() {
    symbol := "BTCUSDT"
    quantity := 0.001
    
    // 1. 获取当前价格
    // 1. Get current price
    currentPrice, err := marketService.GetCurrentPrice(symbol)
    if err != nil {
        log.Fatal(err)
    }
    
    // 2. 设置卖出价格为当前价格的 102%
    // 2. Set sell price to 102% of current price
    sellPrice := currentPrice * 1.02
    fmt.Printf("Current: $%.2f, Sell at: $%.2f\n", currentPrice, sellPrice)
    
    // 3. 下限价卖单
    // 3. Place limit sell order
    order, err := tradingService.PlaceLimitSellOrder(symbol, sellPrice, quantity)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Limit sell order placed!\n")
    fmt.Printf("Order ID: %d\n", order.OrderID)
    fmt.Printf("Price: $%.2f\n", order.Price)
    fmt.Printf("Quantity: %.8f\n", order.OrigQty)
}
```

### 示例 4: 查看订单状态 / Example 4: Check Order Status

```go
func checkOrderStatus(orderID int64) {
    // 查询订单状态
    // Query order status
    status, err := tradingService.GetOrderStatus(orderID)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Order Status Report\n")
    fmt.Printf("==================\n")
    fmt.Printf("Order ID: %d\n", status.OrderID)
    fmt.Printf("Symbol: %s\n", status.Symbol)
    fmt.Printf("Status: %s\n", status.Status)
    fmt.Printf("Original Qty: %.8f\n", status.OrigQty)
    fmt.Printf("Executed Qty: %.8f\n", status.ExecutedQty)
    
    // 计算完成百分比
    // Calculate completion percentage
    if status.OrigQty > 0 {
        percentage := (status.ExecutedQty / status.OrigQty) * 100
        fmt.Printf("Completion: %.2f%%\n", percentage)
    }
}
```

### 示例 5: 取消订单 / Example 5: Cancel Order

```go
func cancelOrderExample(orderID int64) {
    // 1. 先查询订单状态
    // 1. First query order status
    status, err := tradingService.GetOrderStatus(orderID)
    if err != nil {
        log.Fatal(err)
    }
    
    // 2. 检查订单是否可以取消
    // 2. Check if order can be cancelled
    if status.Status == "FILLED" {
        fmt.Println("Order already filled, cannot cancel")
        return
    }
    
    if status.Status == "CANCELED" {
        fmt.Println("Order already cancelled")
        return
    }
    
    // 3. 取消订单
    // 3. Cancel order
    err = tradingService.CancelOrder(orderID)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Order %d cancelled successfully\n", orderID)
}
```

---

## 高级示例 / Advanced Examples

### 示例 6: 批量查询订单 / Example 6: Batch Query Orders

```go
func batchQueryOrders() {
    // 获取所有活跃订单
    // Get all active orders
    activeOrders, err := tradingService.GetActiveOrders()
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Active Orders: %d\n", len(activeOrders))
    fmt.Println("==================")
    
    for i, order := range activeOrders {
        fmt.Printf("[%d] ID: %d, Symbol: %s, Side: %s, Status: %s\n",
            i+1, order.OrderID, order.Symbol, order.Side, order.Status)
        fmt.Printf("    Price: %.8f, Qty: %.8f, Executed: %.8f\n",
            order.Price, order.OrigQty, order.ExecutedQty)
    }
}
```

### 示例 7: 历史数据分析 / Example 7: Historical Data Analysis

```go
func analyzeHistoricalData(symbol string) {
    // 获取最近24小时的K线数据
    // Get last 24 hours of kline data
    klines, err := marketService.GetHistoricalData(symbol, "1h", 24)
    if err != nil {
        log.Fatal(err)
    }
    
    // 计算统计数据
    // Calculate statistics
    var high, low, totalVolume float64
    high = klines[0].High
    low = klines[0].Low
    
    for _, k := range klines {
        if k.High > high {
            high = k.High
        }
        if k.Low < low {
            low = k.Low
        }
        totalVolume += k.Volume
    }
    
    // 计算平均价格
    // Calculate average price
    var sum float64
    for _, k := range klines {
        sum += k.Close
    }
    avgPrice := sum / float64(len(klines))
    
    // 显示分析结果
    // Display analysis results
    fmt.Printf("24h Analysis for %s\n", symbol)
    fmt.Println("==================")
    fmt.Printf("High: $%.2f\n", high)
    fmt.Printf("Low: $%.2f\n", low)
    fmt.Printf("Average: $%.2f\n", avgPrice)
    fmt.Printf("Total Volume: %.2f\n", totalVolume)
    fmt.Printf("Price Range: $%.2f (%.2f%%)\n", high-low, ((high-low)/low)*100)
}
```

### 示例 8: 风险检查 / Example 8: Risk Checking

```go
func checkRisksBeforeTrading(symbol string, quantity float64) bool {
    // 1. 获取当前价格
    // 1. Get current price
    price, err := marketService.GetCurrentPrice(symbol)
    if err != nil {
        log.Printf("Error getting price: %v\n", err)
        return false
    }
    
    // 2. 计算订单金额
    // 2. Calculate order amount
    orderAmount := price * quantity
    
    // 3. 获取风险限制
    // 3. Get risk limits
    limits := riskMgr.GetCurrentLimits()
    
    // 4. 检查订单金额
    // 4. Check order amount
    if orderAmount > limits.MaxOrderAmount {
        fmt.Printf("Order amount $%.2f exceeds limit $%.2f\n",
            orderAmount, limits.MaxOrderAmount)
        return false
    }
    
    // 5. 检查每日订单限制
    // 5. Check daily order limit
    err = riskMgr.CheckDailyLimit()
    if err != nil {
        fmt.Printf("Daily limit exceeded: %v\n", err)
        return false
    }
    
    // 6. 检查最小余额
    // 6. Check minimum balance
    err = riskMgr.CheckMinimumBalance("USDT")
    if err != nil {
        fmt.Printf("Insufficient balance: %v\n", err)
        return false
    }
    
    fmt.Println("All risk checks passed!")
    return true
}
```

---

## 策略示例 / Strategy Examples

### 示例 9: 简单网格交易策略 / Example 9: Simple Grid Trading Strategy

```go
func gridTradingStrategy(symbol string, gridSize int, gridSpacing float64) {
    // 获取当前价格
    // Get current price
    currentPrice, err := marketService.GetCurrentPrice(symbol)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Starting grid trading for %s at $%.2f\n", symbol, currentPrice)
    
    // 创建网格订单
    // Create grid orders
    quantity := 0.001
    
    for i := 1; i <= gridSize; i++ {
        // 计算卖出价格（高于当前价格）
        // Calculate sell price (above current price)
        sellPrice := currentPrice * (1 + float64(i)*gridSpacing)
        
        // 下限价卖单
        // Place limit sell order
        order, err := tradingService.PlaceLimitSellOrder(symbol, sellPrice, quantity)
        if err != nil {
            log.Printf("Error placing sell order at $%.2f: %v\n", sellPrice, err)
            continue
        }
        
        fmt.Printf("Grid sell order placed: ID=%d, Price=$%.2f\n",
            order.OrderID, sellPrice)
    }
    
    fmt.Println("Grid trading setup complete!")
}
```

### 示例 10: 止损策略 / Example 10: Stop Loss Strategy

```go
func stopLossMonitor(symbol string, buyPrice, stopLossPercent float64) {
    fmt.Printf("Monitoring %s for stop loss at %.2f%% below $%.2f\n",
        symbol, stopLossPercent*100, buyPrice)
    
    stopLossPrice := buyPrice * (1 - stopLossPercent)
    
    // 持续监控价格
    // Continuously monitor price
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()
    
    for range ticker.C {
        currentPrice, err := marketService.GetCurrentPrice(symbol)
        if err != nil {
            log.Printf("Error getting price: %v\n", err)
            continue
        }
        
        fmt.Printf("Current: $%.2f, Stop Loss: $%.2f\n", currentPrice, stopLossPrice)
        
        // 检查是否触发止损
        // Check if stop loss triggered
        if currentPrice <= stopLossPrice {
            fmt.Println("Stop loss triggered! Selling...")
            
            // 执行市价卖出
            // Execute market sell
            order, err := tradingService.PlaceMarketSellOrder(symbol, 0.001)
            if err != nil {
                log.Printf("Error placing stop loss order: %v\n", err)
                continue
            }
            
            fmt.Printf("Stop loss order executed: ID=%d\n", order.OrderID)
            break
        }
    }
}
```

### 示例 11: 定投策略 / Example 11: Dollar Cost Averaging (DCA)

```go
func dollarCostAveraging(symbol string, amountUSDT float64, intervalHours int) {
    fmt.Printf("Starting DCA for %s: $%.2f every %d hours\n",
        symbol, amountUSDT, intervalHours)
    
    ticker := time.NewTicker(time.Duration(intervalHours) * time.Hour)
    defer ticker.Stop()
    
    for range ticker.C {
        // 获取当前价格
        // Get current price
        price, err := marketService.GetCurrentPrice(symbol)
        if err != nil {
            log.Printf("Error getting price: %v\n", err)
            continue
        }
        
        // 计算购买数量
        // Calculate quantity to buy
        quantity := amountUSDT / price
        
        fmt.Printf("DCA Buy: %.8f %s at $%.2f (Total: $%.2f)\n",
            quantity, symbol, price, amountUSDT)
        
        // 执行买入
        // Execute buy
        order, err := tradingService.PlaceMarketBuyOrder(symbol, quantity)
        if err != nil {
            log.Printf("Error placing DCA order: %v\n", err)
            continue
        }
        
        fmt.Printf("DCA order executed: ID=%d, Cost=$%.2f\n",
            order.OrderID, order.CummulativeQuoteQty)
    }
}
```

### 示例 12: 价格提醒 / Example 12: Price Alert

```go
func priceAlert(symbol string, targetPrice float64, alertType string) {
    fmt.Printf("Setting %s alert for %s at $%.2f\n", alertType, symbol, targetPrice)
    
    ticker := time.NewTicker(10 * time.Second)
    defer ticker.Stop()
    
    for range ticker.C {
        currentPrice, err := marketService.GetCurrentPrice(symbol)
        if err != nil {
            log.Printf("Error getting price: %v\n", err)
            continue
        }
        
        triggered := false
        if alertType == "above" && currentPrice >= targetPrice {
            triggered = true
        } else if alertType == "below" && currentPrice <= targetPrice {
            triggered = true
        }
        
        if triggered {
            fmt.Printf("🔔 ALERT! %s price is $%.2f (target: $%.2f)\n",
                symbol, currentPrice, targetPrice)
            break
        }
        
        fmt.Printf("Monitoring: %s = $%.2f (target: $%.2f)\n",
            symbol, currentPrice, targetPrice)
    }
}
```

### 示例 13: 余额监控 / Example 13: Balance Monitoring

```go
func monitorBalance(asset string, minThreshold float64) {
    fmt.Printf("Monitoring %s balance (threshold: %.2f)\n", asset, minThreshold)
    
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for range ticker.C {
        balance, err := binanceClient.GetBalance(asset)
        if err != nil {
            log.Printf("Error getting balance: %v\n", err)
            continue
        }
        
        totalBalance := balance.Free + balance.Locked
        fmt.Printf("%s Balance: Free=%.2f, Locked=%.2f, Total=%.2f\n",
            asset, balance.Free, balance.Locked, totalBalance)
        
        if totalBalance < minThreshold {
            fmt.Printf("⚠️  WARNING: %s balance (%.2f) below threshold (%.2f)\n",
                asset, totalBalance, minThreshold)
        }
    }
}
```

### 示例 14: 订单历史导出 / Example 14: Export Order History

```go
func exportOrderHistory(symbol string, startTime, endTime time.Time) {
    // 获取历史订单
    // Get historical orders
    orders, err := binanceClient.GetHistoricalOrders(
        symbol,
        startTime.UnixMilli(),
        endTime.UnixMilli(),
    )
    if err != nil {
        log.Fatal(err)
    }
    
    // 创建CSV文件
    // Create CSV file
    filename := fmt.Sprintf("orders_%s_%s.csv", symbol, time.Now().Format("20060102"))
    file, err := os.Create(filename)
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()
    
    writer := csv.NewWriter(file)
    defer writer.Flush()
    
    // 写入表头
    // Write header
    header := []string{"OrderID", "Symbol", "Side", "Type", "Status", "Price", "Quantity", "Executed", "Time"}
    writer.Write(header)
    
    // 写入订单数据
    // Write order data
    for _, order := range orders {
        record := []string{
            fmt.Sprintf("%d", order.OrderID),
            order.Symbol,
            order.Side,
            order.Type,
            order.Status,
            fmt.Sprintf("%.8f", order.Price),
            fmt.Sprintf("%.8f", order.OrigQty),
            fmt.Sprintf("%.8f", order.ExecutedQty),
            time.UnixMilli(order.Time).Format("2006-01-02 15:04:05"),
        }
        writer.Write(record)
    }
    
    fmt.Printf("Exported %d orders to %s\n", len(orders), filename)
}
```

---

## 错误处理示例 / Error Handling Examples

### 示例 15: 完整的错误处理 / Example 15: Complete Error Handling

```go
func robustTrading(symbol string, quantity float64) {
    // 使用defer捕获panic
    // Use defer to catch panics
    defer func() {
        if r := recover(); r != nil {
            log.Printf("Recovered from panic: %v\n", r)
        }
    }()
    
    // 尝试下单，带完整错误处理
    // Try to place order with complete error handling
    order, err := tradingService.PlaceMarketBuyOrder(symbol, quantity)
    if err != nil {
        // 类型断言检查错误类型
        // Type assertion to check error type
        if tradingErr, ok := err.(*TradingError); ok {
            switch tradingErr.Type {
            case ErrInsufficientBalance:
                fmt.Println("❌ Insufficient balance")
                fmt.Println("💡 Please deposit more funds")
                
            case ErrRiskLimitExceeded:
                fmt.Println("❌ Risk limit exceeded")
                fmt.Println("💡 Try reducing order size or wait for daily limit reset")
                
            case ErrRateLimit:
                fmt.Println("❌ Rate limit exceeded")
                fmt.Println("💡 Waiting 60 seconds before retry...")
                time.Sleep(60 * time.Second)
                // 重试
                // Retry
                return robustTrading(symbol, quantity)
                
            case ErrNetwork:
                fmt.Println("❌ Network error")
                fmt.Println("💡 Check your internet connection")
                
            case ErrAuthentication:
                fmt.Println("❌ Authentication failed")
                fmt.Println("💡 Check your API keys")
                
            default:
                fmt.Printf("❌ Trading error: %s\n", tradingErr.Message)
            }
        } else {
            fmt.Printf("❌ Unknown error: %v\n", err)
        }
        return
    }
    
    // 成功
    // Success
    fmt.Println("✅ Order placed successfully!")
    fmt.Printf("Order ID: %d\n", order.OrderID)
}
```


### 示例 16: 重试机制示例 / Example 16: Retry Mechanism Example

```go
func retryableOperation(symbol string, quantity float64, maxRetries int) (*Order, error) {
    var lastErr error
    
    for attempt := 1; attempt <= maxRetries; attempt++ {
        fmt.Printf("Attempt %d/%d: Placing order...\n", attempt, maxRetries)
        
        order, err := tradingService.PlaceMarketBuyOrder(symbol, quantity)
        if err == nil {
            fmt.Println("✅ Order placed successfully!")
            return order, nil
        }
        
        lastErr = err
        
        // 检查是否应该重试
        // Check if should retry
        if tradingErr, ok := err.(*TradingError); ok {
            if tradingErr.Type == ErrAuthentication {
                // 认证错误不重试
                // Don't retry authentication errors
                fmt.Println("❌ Authentication error, not retrying")
                return nil, err
            }
        }
        
        if attempt < maxRetries {
            // 指数退避
            // Exponential backoff
            waitTime := time.Duration(attempt*attempt) * time.Second
            fmt.Printf("⏳ Waiting %v before retry...\n", waitTime)
            time.Sleep(waitTime)
        }
    }
    
    fmt.Printf("❌ Failed after %d attempts\n", maxRetries)
    return nil, lastErr
}
```

---

## 集成示例 / Integration Examples

### 示例 17: 完整的交易机器人 / Example 17: Complete Trading Bot

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"
    
    "binance-trader/internal/api"
    "binance-trader/internal/config"
    "binance-trader/internal/repository"
    "binance-trader/internal/service"
    "binance-trader/pkg/logger"
)

type TradingBot struct {
    config         *config.Config
    binanceClient  api.BinanceClient
    tradingService service.TradingService
    marketService  service.MarketDataService
    riskManager    service.RiskManager
    logger         *logger.Logger
    ctx            context.Context
    cancel         context.CancelFunc
}

func NewTradingBot(configPath string) (*TradingBot, error) {
    // 加载配置
    // Load configuration
    cfg, err := config.LoadConfig(configPath)
    if err != nil {
        return nil, fmt.Errorf("failed to load config: %w", err)
    }
    
    // 初始化日志
    // Initialize logger
    log, err := logger.NewLogger(cfg.Logging)
    if err != nil {
        return nil, fmt.Errorf("failed to initialize logger: %w", err)
    }
    
    // 初始化API客户端
    // Initialize API client
    authMgr, err := api.NewAuthManager(cfg.Binance.APIKey, cfg.Binance.APISecret)
    if err != nil {
        return nil, fmt.Errorf("failed to create auth manager: %w", err)
    }
    
    rateLimiter := api.NewRateLimiter(cfg.Risk.MaxAPICallsPerMin)
    httpClient := api.NewHTTPClient(rateLimiter, cfg.Retry)
    binanceClient, err := api.NewBinanceClient(cfg.Binance.BaseURL, httpClient, authMgr)
    if err != nil {
        return nil, fmt.Errorf("failed to create binance client: %w", err)
    }
    
    // 初始化服务
    // Initialize services
    orderRepo := repository.NewMemoryOrderRepository()
    riskMgr := service.NewRiskManager(cfg.Risk, binanceClient)
    tradingService := service.NewTradingService(binanceClient, riskMgr, orderRepo, log)
    marketService := service.NewMarketDataService(binanceClient, 1*time.Second)
    
    ctx, cancel := context.WithCancel(context.Background())
    
    return &TradingBot{
        config:         cfg,
        binanceClient:  binanceClient,
        tradingService: tradingService,
        marketService:  marketService,
        riskManager:    riskMgr,
        logger:         log,
        ctx:            ctx,
        cancel:         cancel,
    }, nil
}

func (bot *TradingBot) Start() error {
    bot.logger.Info("Trading bot starting...")
    
    // 设置信号处理
    // Setup signal handling
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    
    // 启动监控goroutine
    // Start monitoring goroutines
    go bot.monitorPrices()
    go bot.monitorOrders()
    
    // 等待退出信号
    // Wait for exit signal
    <-sigChan
    bot.logger.Info("Shutdown signal received")
    bot.Stop()
    
    return nil
}

func (bot *TradingBot) Stop() {
    bot.logger.Info("Trading bot stopping...")
    bot.cancel()
    
    // 取消所有活跃订单
    // Cancel all active orders
    orders, err := bot.tradingService.GetActiveOrders()
    if err != nil {
        bot.logger.Error("Failed to get active orders", "error", err)
        return
    }
    
    for _, order := range orders {
        err := bot.tradingService.CancelOrder(order.OrderID)
        if err != nil {
            bot.logger.Error("Failed to cancel order", "orderID", order.OrderID, "error", err)
        } else {
            bot.logger.Info("Cancelled order", "orderID", order.OrderID)
        }
    }
    
    bot.logger.Info("Trading bot stopped")
}

func (bot *TradingBot) monitorPrices() {
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-bot.ctx.Done():
            return
        case <-ticker.C:
            price, err := bot.marketService.GetCurrentPrice("BTCUSDT")
            if err != nil {
                bot.logger.Error("Failed to get price", "error", err)
                continue
            }
            bot.logger.Debug("Current BTC price", "price", price)
        }
    }
}

func (bot *TradingBot) monitorOrders() {
    ticker := time.NewTicker(10 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-bot.ctx.Done():
            return
        case <-ticker.C:
            orders, err := bot.tradingService.GetActiveOrders()
            if err != nil {
                bot.logger.Error("Failed to get active orders", "error", err)
                continue
            }
            bot.logger.Info("Active orders", "count", len(orders))
        }
    }
}

func main() {
    bot, err := NewTradingBot("config.yaml")
    if err != nil {
        log.Fatal(err)
    }
    
    if err := bot.Start(); err != nil {
        log.Fatal(err)
    }
}
```

---

## 测试示例 / Testing Examples

### 示例 18: 单元测试示例 / Example 18: Unit Test Example

```go
package service

import (
    "testing"
    "binance-trader/internal/api"
)

func TestPlaceMarketBuyOrder(t *testing.T) {
    // 创建模拟客户端
    // Create mock client
    mockClient := &MockBinanceClient{
        GetPriceFunc: func(symbol string) (*api.Price, error) {
            return &api.Price{Symbol: symbol, Price: 50000.0}, nil
        },
        CreateOrderFunc: func(req *api.OrderRequest) (*api.OrderResponse, error) {
            return &api.OrderResponse{
                OrderID: 12345,
                Status:  "FILLED",
                Price:   50000.0,
            }, nil
        },
    }
    
    // 创建服务
    // Create service
    orderRepo := repository.NewMemoryOrderRepository()
    riskMgr := &MockRiskManager{}
    logger := &MockLogger{}
    
    service := NewTradingService(mockClient, riskMgr, orderRepo, logger)
    
    // 测试
    // Test
    order, err := service.PlaceMarketBuyOrder("BTCUSDT", 0.001)
    
    // 断言
    // Assertions
    if err != nil {
        t.Fatalf("Expected no error, got %v", err)
    }
    
    if order.OrderID != 12345 {
        t.Errorf("Expected order ID 12345, got %d", order.OrderID)
    }
    
    if order.Status != "FILLED" {
        t.Errorf("Expected status FILLED, got %s", order.Status)
    }
}
```

### 示例 19: 属性测试示例 / Example 19: Property-Based Test Example

```go
package service

import (
    "testing"
    "github.com/leanovate/gopter"
    "github.com/leanovate/gopter/gen"
    "github.com/leanovate/gopter/prop"
)

// Feature: binance-auto-trading, Property 16: 订单金额限制检查
func TestProperty_OrderAmountLimit(t *testing.T) {
    properties := gopter.NewProperties(nil)
    
    properties.Property("orders exceeding max amount are rejected", prop.ForAll(
        func(price float64, quantity float64) bool {
            // 设置最大金额限制
            // Set max amount limit
            maxAmount := 10000.0
            orderAmount := price * quantity
            
            // 创建风险管理器
            // Create risk manager
            limits := &RiskLimits{
                MaxOrderAmount: maxAmount,
            }
            riskMgr := NewRiskManager(limits, nil)
            
            // 创建订单请求
            // Create order request
            orderReq := &OrderRequest{
                Symbol:   "BTCUSDT",
                Side:     "BUY",
                Type:     "MARKET",
                Quantity: quantity,
            }
            
            // 验证订单
            // Validate order
            err := riskMgr.ValidateOrder(orderReq)
            
            // 属性：如果订单金额超过限制，应该返回错误
            // Property: if order amount exceeds limit, should return error
            if orderAmount > maxAmount {
                return err != nil
            }
            return err == nil
        },
        gen.Float64Range(1000, 100000),  // price range
        gen.Float64Range(0.001, 10),     // quantity range
    ))
    
    properties.TestingRun(t, gopter.ConsoleReporter(false))
}
```

### 示例 20: 集成测试示例 / Example 20: Integration Test Example

```go
// +build integration

package integration

import (
    "testing"
    "time"
    "binance-trader/internal/api"
    "binance-trader/internal/config"
)

func TestIntegration_CompleteTradeFlow(t *testing.T) {
    // 加载测试网配置
    // Load testnet configuration
    cfg, err := config.LoadConfig("../config.testnet.yaml")
    if err != nil {
        t.Fatal(err)
    }
    
    // 初始化客户端
    // Initialize client
    authMgr, _ := api.NewAuthManager(cfg.Binance.APIKey, cfg.Binance.APISecret)
    rateLimiter := api.NewRateLimiter(1000)
    httpClient := api.NewHTTPClient(rateLimiter, cfg.Retry)
    client, err := api.NewBinanceClient(cfg.Binance.BaseURL, httpClient, authMgr)
    if err != nil {
        t.Fatal(err)
    }
    
    // 测试获取价格
    // Test get price
    t.Run("GetPrice", func(t *testing.T) {
        price, err := client.GetPrice("BTCUSDT")
        if err != nil {
            t.Fatalf("Failed to get price: %v", err)
        }
        if price.Price <= 0 {
            t.Errorf("Invalid price: %f", price.Price)
        }
        t.Logf("Current BTC price: $%.2f", price.Price)
    })
    
    // 测试创建订单
    // Test create order
    t.Run("CreateOrder", func(t *testing.T) {
        orderReq := &api.OrderRequest{
            Symbol:   "BTCUSDT",
            Side:     "BUY",
            Type:     "MARKET",
            Quantity: 0.001,
        }
        
        order, err := client.CreateOrder(orderReq)
        if err != nil {
            t.Fatalf("Failed to create order: %v", err)
        }
        
        if order.OrderID == 0 {
            t.Error("Invalid order ID")
        }
        
        t.Logf("Order created: ID=%d, Status=%s", order.OrderID, order.Status)
        
        // 等待订单成交
        // Wait for order to fill
        time.Sleep(2 * time.Second)
        
        // 查询订单状态
        // Query order status
        orderStatus, err := client.GetOrder("BTCUSDT", order.OrderID)
        if err != nil {
            t.Fatalf("Failed to get order: %v", err)
        }
        
        t.Logf("Order status: %s, Executed: %.8f", orderStatus.Status, orderStatus.ExecutedQty)
    })
}
```

---

## 性能优化示例 / Performance Optimization Examples

### 示例 21: 批量操作 / Example 21: Batch Operations

```go
func batchGetPrices(symbols []string) map[string]float64 {
    results := make(map[string]float64)
    resultChan := make(chan struct {
        symbol string
        price  float64
        err    error
    }, len(symbols))
    
    // 并发获取价格
    // Concurrent price fetching
    for _, symbol := range symbols {
        go func(sym string) {
            price, err := marketService.GetCurrentPrice(sym)
            resultChan <- struct {
                symbol string
                price  float64
                err    error
            }{sym, price, err}
        }(symbol)
    }
    
    // 收集结果
    // Collect results
    for i := 0; i < len(symbols); i++ {
        result := <-resultChan
        if result.err == nil {
            results[result.symbol] = result.price
        } else {
            log.Printf("Error getting price for %s: %v", result.symbol, result.err)
        }
    }
    
    return results
}
```

### 示例 22: 缓存使用 / Example 22: Cache Usage

```go
type PriceCache struct {
    cache map[string]cachedPrice
    mu    sync.RWMutex
    ttl   time.Duration
}

type cachedPrice struct {
    price     float64
    timestamp time.Time
}

func NewPriceCache(ttl time.Duration) *PriceCache {
    return &PriceCache{
        cache: make(map[string]cachedPrice),
        ttl:   ttl,
    }
}

func (pc *PriceCache) Get(symbol string) (float64, bool) {
    pc.mu.RLock()
    defer pc.mu.RUnlock()
    
    cached, exists := pc.cache[symbol]
    if !exists {
        return 0, false
    }
    
    // 检查是否过期
    // Check if expired
    if time.Since(cached.timestamp) > pc.ttl {
        return 0, false
    }
    
    return cached.price, true
}

func (pc *PriceCache) Set(symbol string, price float64) {
    pc.mu.Lock()
    defer pc.mu.Unlock()
    
    pc.cache[symbol] = cachedPrice{
        price:     price,
        timestamp: time.Now(),
    }
}

func getPriceWithCache(symbol string, cache *PriceCache) (float64, error) {
    // 先检查缓存
    // Check cache first
    if price, ok := cache.Get(symbol); ok {
        return price, nil
    }
    
    // 缓存未命中，从API获取
    // Cache miss, fetch from API
    price, err := marketService.GetCurrentPrice(symbol)
    if err != nil {
        return 0, err
    }
    
    // 更新缓存
    // Update cache
    cache.Set(symbol, price)
    
    return price, nil
}
```

---

**更多信息请参阅 / For more information, see:**
- [API Documentation](API.md)
- [README](../README.md)
