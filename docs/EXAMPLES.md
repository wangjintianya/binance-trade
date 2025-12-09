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

## 条件订单示例 / Conditional Order Examples

### 示例 23: 价格触发买入 / Example 23: Price-Triggered Buy

```go
func priceTriggeredBuy(symbol string, quantity, triggerPrice float64) {
    // 创建价格触发条件
    // Create price trigger condition
    triggerCondition := &TriggerCondition{
        Type:     TriggerTypePrice,
        Operator: OperatorGreaterEqual,
        Value:    triggerPrice,
    }
    
    // 创建条件订单请求
    // Create conditional order request
    orderReq := &ConditionalOrderRequest{
        Symbol:           symbol,
        Side:             "BUY",
        Type:             "MARKET",
        Quantity:         quantity,
        TriggerCondition: triggerCondition,
    }
    
    // 提交条件订单
    // Submit conditional order
    order, err := conditionalOrderService.CreateConditionalOrder(orderReq)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Conditional buy order created!\n")
    fmt.Printf("Order ID: %s\n", order.OrderID)
    fmt.Printf("Will trigger when %s >= $%.2f\n", symbol, triggerPrice)
}
```

### 示例 24: 涨跌幅触发卖出 / Example 24: Percentage Change Triggered Sell

```go
func percentageTriggeredSell(symbol string, quantity, basePrice, changePercent float64) {
    // 创建涨跌幅触发条件
    // Create percentage change trigger condition
    triggerCondition := &TriggerCondition{
        Type:      TriggerTypePriceChangePercent,
        Operator:  OperatorGreaterEqual,
        Value:     changePercent,
        BasePrice: basePrice,
    }
    
    orderReq := &ConditionalOrderRequest{
        Symbol:           symbol,
        Side:             "SELL",
        Type:             "MARKET",
        Quantity:         quantity,
        TriggerCondition: triggerCondition,
    }
    
    order, err := conditionalOrderService.CreateConditionalOrder(orderReq)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Percentage-triggered sell order created!\n")
    fmt.Printf("Order ID: %s\n", order.OrderID)
    fmt.Printf("Will trigger when price rises %.2f%% from $%.2f\n", changePercent, basePrice)
}
```

### 示例 25: 成交量触发订单 / Example 25: Volume-Triggered Order

```go
func volumeTriggeredOrder(symbol string, quantity, volumeThreshold float64, timeWindow time.Duration) {
    // 创建成交量触发条件
    // Create volume trigger condition
    triggerCondition := &TriggerCondition{
        Type:       TriggerTypeVolume,
        Operator:   OperatorGreaterEqual,
        Value:      volumeThreshold,
        TimeWindow: timeWindow,
    }
    
    orderReq := &ConditionalOrderRequest{
        Symbol:           symbol,
        Side:             "BUY",
        Type:             "MARKET",
        Quantity:         quantity,
        TriggerCondition: triggerCondition,
    }
    
    order, err := conditionalOrderService.CreateConditionalOrder(orderReq)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Volume-triggered order created!\n")
    fmt.Printf("Order ID: %s\n", order.OrderID)
    fmt.Printf("Will trigger when volume >= %.2f in %v\n", volumeThreshold, timeWindow)
}
```

### 示例 26: 复合条件订单 / Example 26: Composite Condition Order

```go
func compositeConditionOrder(symbol string, quantity float64) {
    // 创建复合条件：价格 > 45000 AND 成交量 > 1000
    // Create composite condition: price > 45000 AND volume > 1000
    priceCondition := &TriggerCondition{
        Type:     TriggerTypePrice,
        Operator: OperatorGreaterThan,
        Value:    45000.0,
    }
    
    volumeCondition := &TriggerCondition{
        Type:       TriggerTypeVolume,
        Operator:   OperatorGreaterThan,
        Value:      1000.0,
        TimeWindow: 1 * time.Hour,
    }
    
    compositeCondition := &TriggerCondition{
        CompositeType: LogicAND,
        SubConditions: []*TriggerCondition{priceCondition, volumeCondition},
    }
    
    orderReq := &ConditionalOrderRequest{
        Symbol:           symbol,
        Side:             "BUY",
        Type:             "MARKET",
        Quantity:         quantity,
        TriggerCondition: compositeCondition,
    }
    
    order, err := conditionalOrderService.CreateConditionalOrder(orderReq)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Composite condition order created!\n")
    fmt.Printf("Order ID: %s\n", order.OrderID)
    fmt.Println("Will trigger when BOTH conditions are met:")
    fmt.Println("  1. Price > $45000")
    fmt.Println("  2. Volume > 1000 in last hour")
}
```

### 示例 27: 查询和管理条件订单 / Example 27: Query and Manage Conditional Orders

```go
func manageConditionalOrders() {
    // 获取所有活跃的条件订单
    // Get all active conditional orders
    activeOrders, err := conditionalOrderService.GetActiveConditionalOrders()
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Active Conditional Orders: %d\n", len(activeOrders))
    fmt.Println("==================")
    
    for i, order := range activeOrders {
        fmt.Printf("[%d] Order ID: %s\n", i+1, order.OrderID)
        fmt.Printf("    Symbol: %s, Side: %s, Quantity: %.8f\n",
            order.Symbol, order.Side, order.Quantity)
        fmt.Printf("    Status: %s, Created: %s\n",
            order.Status, time.UnixMilli(order.CreatedAt).Format("2006-01-02 15:04:05"))
        
        // 显示触发条件
        // Display trigger condition
        if order.TriggerCondition != nil {
            fmt.Printf("    Trigger: %s %s %.2f\n",
                getTriggerTypeName(order.TriggerCondition.Type),
                getOperatorName(order.TriggerCondition.Operator),
                order.TriggerCondition.Value)
        }
    }
    
    // 取消特定条件订单
    // Cancel specific conditional order
    if len(activeOrders) > 0 {
        orderToCancel := activeOrders[0].OrderID
        err := conditionalOrderService.CancelConditionalOrder(orderToCancel)
        if err != nil {
            log.Printf("Failed to cancel order %s: %v\n", orderToCancel, err)
        } else {
            fmt.Printf("Cancelled conditional order: %s\n", orderToCancel)
        }
    }
}

func getTriggerTypeName(t TriggerType) string {
    switch t {
    case TriggerTypePrice:
        return "Price"
    case TriggerTypePriceChangePercent:
        return "Price Change %"
    case TriggerTypeVolume:
        return "Volume"
    default:
        return "Unknown"
    }
}

func getOperatorName(op ComparisonOperator) string {
    switch op {
    case OperatorGreaterThan:
        return ">"
    case OperatorLessThan:
        return "<"
    case OperatorGreaterEqual:
        return ">="
    case OperatorLessEqual:
        return "<="
    default:
        return "?"
    }
}
```

---

## 止损止盈示例 / Stop Loss & Take Profit Examples

### 示例 28: 设置止损 / Example 28: Set Stop Loss

```go
func setStopLossExample(symbol string, position, stopPrice float64) {
    // 为持仓设置止损
    // Set stop loss for position
    stopOrder, err := stopLossService.SetStopLoss(symbol, position, stopPrice)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Stop loss order created!\n")
    fmt.Printf("Order ID: %s\n", stopOrder.OrderID)
    fmt.Printf("Symbol: %s\n", stopOrder.Symbol)
    fmt.Printf("Position: %.8f\n", stopOrder.Position)
    fmt.Printf("Stop Price: $%.2f\n", stopOrder.StopPrice)
    fmt.Printf("Status: %s\n", stopOrder.Status)
    
    // 计算潜在损失
    // Calculate potential loss
    currentPrice, _ := marketService.GetCurrentPrice(symbol)
    potentialLoss := (currentPrice - stopPrice) * position
    lossPercent := ((currentPrice - stopPrice) / currentPrice) * 100
    
    fmt.Printf("\nRisk Analysis:\n")
    fmt.Printf("Current Price: $%.2f\n", currentPrice)
    fmt.Printf("Potential Loss: $%.2f (%.2f%%)\n", potentialLoss, lossPercent)
}
```

### 示例 29: 设置止盈 / Example 29: Set Take Profit

```go
func setTakeProfitExample(symbol string, position, targetPrice float64) {
    // 为持仓设置止盈
    // Set take profit for position
    takeProfitOrder, err := stopLossService.SetTakeProfit(symbol, position, targetPrice)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Take profit order created!\n")
    fmt.Printf("Order ID: %s\n", takeProfitOrder.OrderID)
    fmt.Printf("Symbol: %s\n", takeProfitOrder.Symbol)
    fmt.Printf("Position: %.8f\n", takeProfitOrder.Position)
    fmt.Printf("Target Price: $%.2f\n", takeProfitOrder.StopPrice)
    
    // 计算潜在利润
    // Calculate potential profit
    currentPrice, _ := marketService.GetCurrentPrice(symbol)
    potentialProfit := (targetPrice - currentPrice) * position
    profitPercent := ((targetPrice - currentPrice) / currentPrice) * 100
    
    fmt.Printf("\nProfit Target:\n")
    fmt.Printf("Current Price: $%.2f\n", currentPrice)
    fmt.Printf("Potential Profit: $%.2f (%.2f%%)\n", potentialProfit, profitPercent)
}
```

### 示例 30: 同时设置止损止盈 / Example 30: Set Both Stop-Loss and Take-Profit

```go
func setStopLossTakeProfitExample(symbol string, position, stopPrice, targetPrice float64) {
    // 同时设置止损和止盈
    // Set both stop-loss and take-profit
    orderPair, err := stopLossService.SetStopLossTakeProfit(symbol, position, stopPrice, targetPrice)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Stop-loss and take-profit orders created!\n")
    fmt.Printf("Pair ID: %s\n", orderPair.PairID)
    fmt.Printf("Symbol: %s\n", orderPair.Symbol)
    fmt.Printf("Position: %.8f\n", orderPair.Position)
    
    fmt.Printf("\nStop Loss Order:\n")
    fmt.Printf("  Order ID: %s\n", orderPair.StopLossOrder.OrderID)
    fmt.Printf("  Stop Price: $%.2f\n", orderPair.StopLossOrder.StopPrice)
    
    fmt.Printf("\nTake Profit Order:\n")
    fmt.Printf("  Order ID: %s\n", orderPair.TakeProfitOrder.OrderID)
    fmt.Printf("  Target Price: $%.2f\n", orderPair.TakeProfitOrder.StopPrice)
    
    // 计算风险回报比
    // Calculate risk-reward ratio
    currentPrice, _ := marketService.GetCurrentPrice(symbol)
    risk := (currentPrice - stopPrice) * position
    reward := (targetPrice - currentPrice) * position
    ratio := reward / risk
    
    fmt.Printf("\nRisk-Reward Analysis:\n")
    fmt.Printf("Current Price: $%.2f\n", currentPrice)
    fmt.Printf("Risk: $%.2f\n", risk)
    fmt.Printf("Reward: $%.2f\n", reward)
    fmt.Printf("Risk-Reward Ratio: 1:%.2f\n", ratio)
}
```

### 示例 31: 设置移动止损 / Example 31: Set Trailing Stop

```go
func setTrailingStopExample(symbol string, position, trailPercent float64) {
    // 设置移动止损
    // Set trailing stop
    trailingStop, err := stopLossService.SetTrailingStop(symbol, position, trailPercent)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Trailing stop order created!\n")
    fmt.Printf("Order ID: %s\n", trailingStop.OrderID)
    fmt.Printf("Symbol: %s\n", trailingStop.Symbol)
    fmt.Printf("Position: %.8f\n", trailingStop.Position)
    fmt.Printf("Trail Percent: %.2f%%\n", trailingStop.TrailPercent)
    fmt.Printf("Highest Price: $%.2f\n", trailingStop.HighestPrice)
    fmt.Printf("Current Stop Price: $%.2f\n", trailingStop.CurrentStopPrice)
    
    fmt.Printf("\nHow it works:\n")
    fmt.Println("- Stop price will adjust upward as price rises")
    fmt.Printf("- Stop price stays %.2f%% below the highest price\n", trailPercent)
    fmt.Println("- Locks in profits while allowing upside potential")
}
```

### 示例 32: 监控止损止盈订单 / Example 32: Monitor Stop Orders

```go
func monitorStopOrders(symbol string) {
    ticker := time.NewTicker(10 * time.Second)
    defer ticker.Stop()
    
    fmt.Printf("Monitoring stop orders for %s...\n", symbol)
    
    for range ticker.C {
        // 获取活跃的止损止盈订单
        // Get active stop orders
        stopOrders, err := stopLossService.GetActiveStopOrders(symbol)
        if err != nil {
            log.Printf("Error getting stop orders: %v\n", err)
            continue
        }
        
        // 获取当前价格
        // Get current price
        currentPrice, err := marketService.GetCurrentPrice(symbol)
        if err != nil {
            log.Printf("Error getting price: %v\n", err)
            continue
        }
        
        fmt.Printf("\n=== %s ===\n", time.Now().Format("15:04:05"))
        fmt.Printf("Current Price: $%.2f\n", currentPrice)
        fmt.Printf("Active Stop Orders: %d\n", len(stopOrders))
        
        for i, order := range stopOrders {
            fmt.Printf("\n[%d] %s Order (ID: %s)\n", i+1, getStopOrderTypeName(order.Type), order.OrderID)
            fmt.Printf("    Stop Price: $%.2f\n", order.StopPrice)
            
            // 计算距离触发的距离
            // Calculate distance to trigger
            var distance float64
            var direction string
            if order.Type == StopOrderTypeStopLoss {
                distance = currentPrice - order.StopPrice
                direction = "above"
            } else {
                distance = order.StopPrice - currentPrice
                direction = "below"
            }
            
            distancePercent := (distance / currentPrice) * 100
            fmt.Printf("    Distance: $%.2f (%.2f%% %s)\n", math.Abs(distance), math.Abs(distancePercent), direction)
            
            // 警告即将触发
            // Warn if close to trigger
            if math.Abs(distancePercent) < 1.0 {
                fmt.Printf("    ⚠️  WARNING: Close to trigger!\n")
            }
        }
    }
}

func getStopOrderTypeName(t StopOrderType) string {
    switch t {
    case StopOrderTypeStopLoss:
        return "Stop Loss"
    case StopOrderTypeTakeProfit:
        return "Take Profit"
    default:
        return "Unknown"
    }
}
```

### 示例 33: 动态调整移动止损 / Example 33: Dynamic Trailing Stop Adjustment

```go
func dynamicTrailingStop(symbol string, position float64) {
    // 初始设置3%的移动止损
    // Initially set 3% trailing stop
    trailingStop, err := stopLossService.SetTrailingStop(symbol, position, 3.0)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Initial trailing stop set at 3%%\n")
    fmt.Printf("Order ID: %s\n", trailingStop.OrderID)
    
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for range ticker.C {
        currentPrice, err := marketService.GetCurrentPrice(symbol)
        if err != nil {
            log.Printf("Error getting price: %v\n", err)
            continue
        }
        
        // 计算利润百分比
        // Calculate profit percentage
        entryPrice := trailingStop.HighestPrice / 1.03 // 反推入场价格 / Reverse calculate entry price
        profitPercent := ((currentPrice - entryPrice) / entryPrice) * 100
        
        fmt.Printf("\nCurrent Price: $%.2f\n", currentPrice)
        fmt.Printf("Profit: %.2f%%\n", profitPercent)
        fmt.Printf("Current Stop: $%.2f\n", trailingStop.CurrentStopPrice)
        
        // 根据利润调整移动止损百分比
        // Adjust trailing stop percentage based on profit
        var newTrailPercent float64
        if profitPercent > 20 {
            newTrailPercent = 1.0 // 利润超过20%，收紧到1% / Tighten to 1% when profit > 20%
        } else if profitPercent > 10 {
            newTrailPercent = 2.0 // 利润超过10%，收紧到2% / Tighten to 2% when profit > 10%
        } else {
            newTrailPercent = 3.0 // 保持3% / Keep at 3%
        }
        
        // 如果需要调整
        // If adjustment needed
        if newTrailPercent != trailingStop.TrailPercent {
            err := stopLossService.UpdateTrailingStop(trailingStop.OrderID, newTrailPercent)
            if err != nil {
                log.Printf("Error updating trailing stop: %v\n", err)
            } else {
                fmt.Printf("✅ Adjusted trailing stop to %.1f%%\n", newTrailPercent)
                trailingStop.TrailPercent = newTrailPercent
            }
        }
    }
}
```

---

**更多信息请参阅 / For more information, see:**
- [API Documentation](API.md)
- [README](../README.md)


---

## 合约交易示例 / Futures Trading Examples

### 示例 34: 简单合约开仓 / Example 34: Simple Futures Position Opening

```go
func simpleFuturesLong(symbol string, quantity float64) {
    // 1. 设置杠杆 / Set leverage
    err := futuresClient.SetLeverage(symbol, 5)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("Leverage set to 5x")
    
    // 2. 获取标记价格 / Get mark price
    markPrice, err := futuresClient.GetMarkPrice(symbol)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Mark Price: $%.2f\n", markPrice.MarkPrice)
    
    // 3. 开多仓 / Open long position
    order, err := futuresTradingService.OpenLongPosition(symbol, quantity, MARKET, 0)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Long position opened!\n")
    fmt.Printf("Order ID: %d\n", order.OrderID)
    fmt.Printf("Entry Price: $%.2f\n", order.AvgPrice)
    fmt.Printf("Quantity: %.8f\n", order.ExecutedQty)
    fmt.Printf("Position Value: $%.2f\n", order.AvgPrice*order.ExecutedQty)
    fmt.Printf("Margin Used: $%.2f (5x leverage)\n", (order.AvgPrice*order.ExecutedQty)/5)
}
```

### 示例 35: 合约持仓监控 / Example 35: Futures Position Monitoring

```go
func monitorFuturesPosition(symbol string) {
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()
    
    fmt.Printf("Monitoring position for %s...\n", symbol)
    
    for range ticker.C {
        // 获取持仓 / Get position
        position, err := futuresPositionMgr.GetPosition(symbol, LONG)
        if err != nil {
            log.Printf("Error getting position: %v\n", err)
            continue
        }
        
        if position.PositionAmt == 0 {
            fmt.Println("No position")
            continue
        }
        
        // 获取标记价格 / Get mark price
        markPrice, err := futuresClient.GetMarkPrice(symbol)
        if err != nil {
            log.Printf("Error getting mark price: %v\n", err)
            continue
        }
        
        // 计算盈亏百分比 / Calculate PnL percentage
        pnlPercent := (position.UnrealizedProfit / (position.EntryPrice * position.PositionAmt)) * 100
        
        // 计算距离强平的距离 / Calculate distance to liquidation
        distanceToLiq := ((markPrice.MarkPrice - position.LiquidationPrice) / markPrice.MarkPrice) * 100
        
        fmt.Printf("\n=== %s ===\n", time.Now().Format("15:04:05"))
        fmt.Printf("Position: %.8f %s\n", position.PositionAmt, symbol)
        fmt.Printf("Entry: $%.2f | Mark: $%.2f\n", position.EntryPrice, markPrice.MarkPrice)
        fmt.Printf("Unrealized PnL: $%.2f (%.2f%%)\n", position.UnrealizedProfit, pnlPercent)
        fmt.Printf("Liquidation: $%.2f (%.2f%% away)\n", position.LiquidationPrice, distanceToLiq)
        fmt.Printf("Leverage: %dx | Margin: $%.2f\n", position.Leverage, position.IsolatedMargin)
        
        // 警告强平风险 / Warn liquidation risk
        if distanceToLiq < 5.0 {
            fmt.Printf("⚠️  WARNING: Close to liquidation!\n")
        }
    }
}
```

### 示例 36: 合约网格交易策略 / Example 36: Futures Grid Trading Strategy

```go
func futuresGridTrading(symbol string, gridSize int, gridSpacing float64, leverage int) {
    // 设置杠杆 / Set leverage
    err := futuresClient.SetLeverage(symbol, leverage)
    if err != nil {
        log.Fatal(err)
    }
    
    // 获取当前标记价格 / Get current mark price
    markPrice, err := futuresClient.GetMarkPrice(symbol)
    if err != nil {
        log.Fatal(err)
    }
    
    currentPrice := markPrice.MarkPrice
    fmt.Printf("Starting futures grid trading for %s at $%.2f with %dx leverage\n", 
        symbol, currentPrice, leverage)
    
    quantity := 0.001
    
    // 创建网格订单 / Create grid orders
    for i := 1; i <= gridSize; i++ {
        // 上方卖出网格（做空）/ Upper sell grid (short)
        sellPrice := currentPrice * (1 + float64(i)*gridSpacing)
        shortOrder, err := futuresTradingService.OpenShortPosition(symbol, quantity, LIMIT, sellPrice)
        if err != nil {
            log.Printf("Error placing short order at $%.2f: %v\n", sellPrice, err)
            continue
        }
        fmt.Printf("Grid short order placed: ID=%d, Price=$%.2f\n", shortOrder.OrderID, sellPrice)
        
        // 下方买入网格（做多）/ Lower buy grid (long)
        buyPrice := currentPrice * (1 - float64(i)*gridSpacing)
        longOrder, err := futuresTradingService.OpenLongPosition(symbol, quantity, LIMIT, buyPrice)
        if err != nil {
            log.Printf("Error placing long order at $%.2f: %v\n", buyPrice, err)
            continue
        }
        fmt.Printf("Grid long order placed: ID=%d, Price=$%.2f\n", longOrder.OrderID, buyPrice)
    }
    
    fmt.Println("Futures grid trading setup complete!")
}
```

### 示例 37: 合约对冲策略 / Example 37: Futures Hedging Strategy

```go
func hedgingStrategy(symbol string, spotQuantity float64) {
    // 启用双向持仓模式 / Enable hedge mode
    err := futuresClient.SetPositionMode(true)
    if err != nil {
        log.Fatal(err)
    }
    
    // 获取现货持仓价值 / Get spot position value
    spotPrice, err := spotMarketService.GetCurrentPrice(symbol)
    if err != nil {
        log.Fatal(err)
    }
    spotValue := spotPrice * spotQuantity
    
    fmt.Printf("Hedging spot position: %.8f %s ($%.2f)\n", 
        spotQuantity, symbol, spotValue)
    
    // 在合约市场开等量空仓对冲 / Open equal short position in futures to hedge
    futuresQuantity := spotQuantity
    shortOrder, err := futuresTradingService.OpenShortPosition(symbol, futuresQuantity, MARKET, 0)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Hedge position opened!\n")
    fmt.Printf("Short Order ID: %d\n", shortOrder.OrderID)
    fmt.Printf("Entry Price: $%.2f\n", shortOrder.AvgPrice)
    fmt.Printf("Quantity: %.8f\n", shortOrder.ExecutedQty)
    
    // 监控对冲效果 / Monitor hedging effectiveness
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for range ticker.C {
        // 获取现货价格 / Get spot price
        currentSpotPrice, _ := spotMarketService.GetCurrentPrice(symbol)
        spotPnL := (currentSpotPrice - spotPrice) * spotQuantity
        
        // 获取合约持仓 / Get futures position
        futuresPosition, _ := futuresPositionMgr.GetPosition(symbol, SHORT)
        futuresPnL := futuresPosition.UnrealizedProfit
        
        // 计算总盈亏 / Calculate total PnL
        totalPnL := spotPnL + futuresPnL
        
        fmt.Printf("\n=== Hedge Status ===\n")
        fmt.Printf("Spot PnL: $%.2f\n", spotPnL)
        fmt.Printf("Futures PnL: $%.2f\n", futuresPnL)
        fmt.Printf("Total PnL: $%.2f\n", totalPnL)
        fmt.Printf("Hedge Effectiveness: %.2f%%\n", (1-(math.Abs(totalPnL)/math.Abs(spotPnL)))*100)
    }
}
```

### 示例 38: 资金费率套利 / Example 38: Funding Rate Arbitrage

```go
func fundingRateArbitrage(symbol string, quantity float64) {
    // 获取资金费率 / Get funding rate
    fundingRate, err := futuresClient.GetFundingRate(symbol)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Current Funding Rate: %.4f%%\n", fundingRate.FundingRate*100)
    fmt.Printf("Next Funding Time: %s\n", time.UnixMilli(fundingRate.FundingTime).Format("2006-01-02 15:04:05"))
    
    // 如果资金费率为正且较高，做空收取费用 / If funding rate is positive and high, short to collect fees
    if fundingRate.FundingRate > 0.001 { // 0.1%
        fmt.Println("High positive funding rate detected!")
        fmt.Println("Strategy: Open short position to collect funding fees")
        
        // 开空仓 / Open short position
        shortOrder, err := futuresTradingService.OpenShortPosition(symbol, quantity, MARKET, 0)
        if err != nil {
            log.Fatal(err)
        }
        
        fmt.Printf("Short position opened at $%.2f\n", shortOrder.AvgPrice)
        
        // 同时在现货市场买入对冲 / Simultaneously buy in spot market to hedge
        spotOrder, err := spotTradingService.PlaceMarketBuyOrder(symbol, quantity)
        if err != nil {
            log.Printf("Warning: Failed to hedge in spot market: %v\n", err)
        } else {
            fmt.Printf("Spot hedge bought at $%.2f\n", spotOrder.Price)
        }
        
        // 计算预期收益 / Calculate expected profit
        positionValue := shortOrder.AvgPrice * quantity
        expectedFunding := positionValue * fundingRate.FundingRate
        
        fmt.Printf("\nExpected funding fee per 8h: $%.2f\n", expectedFunding)
        fmt.Printf("Daily expected: $%.2f\n", expectedFunding*3)
        
    } else if fundingRate.FundingRate < -0.001 { // -0.1%
        fmt.Println("High negative funding rate detected!")
        fmt.Println("Strategy: Open long position to collect funding fees")
        
        // 开多仓 / Open long position
        longOrder, err := futuresTradingService.OpenLongPosition(symbol, quantity, MARKET, 0)
        if err != nil {
            log.Fatal(err)
        }
        
        fmt.Printf("Long position opened at $%.2f\n", longOrder.AvgPrice)
        
        // 同时在现货市场卖出对冲 / Simultaneously sell in spot market to hedge
        spotOrder, err := spotTradingService.PlaceLimitSellOrder(symbol, longOrder.AvgPrice, quantity)
        if err != nil {
            log.Printf("Warning: Failed to hedge in spot market: %v\n", err)
        } else {
            fmt.Printf("Spot hedge sold at $%.2f\n", spotOrder.Price)
        }
        
        // 计算预期收益 / Calculate expected profit
        positionValue := longOrder.AvgPrice * quantity
        expectedFunding := positionValue * math.Abs(fundingRate.FundingRate)
        
        fmt.Printf("\nExpected funding fee per 8h: $%.2f\n", expectedFunding)
        fmt.Printf("Daily expected: $%.2f\n", expectedFunding*3)
        
    } else {
        fmt.Println("Funding rate too low for arbitrage")
    }
}
```

### 示例 39: 合约止损止盈管理 / Example 39: Futures Stop Loss/Take Profit Management

```go
func manageFuturesStopOrders(symbol string, position *Position) {
    entryPrice := position.EntryPrice
    positionSide := position.PositionSide
    quantity := math.Abs(position.PositionAmt)
    
    // 计算止损止盈价格 / Calculate stop loss and take profit prices
    var stopLossPrice, takeProfitPrice float64
    
    if positionSide == LONG {
        // 多头：止损在下方，止盈在上方 / Long: stop loss below, take profit above
        stopLossPrice = entryPrice * 0.97   // 3% stop loss
        takeProfitPrice = entryPrice * 1.06 // 6% take profit
    } else {
        // 空头：止损在上方，止盈在下方 / Short: stop loss above, take profit below
        stopLossPrice = entryPrice * 1.03   // 3% stop loss
        takeProfitPrice = entryPrice * 0.94 // 6% take profit
    }
    
    // 设置止损止盈 / Set stop loss and take profit
    orderPair, err := futuresStopLossService.SetStopLossTakeProfit(
        symbol,
        positionSide,
        quantity,
        stopLossPrice,
        takeProfitPrice,
    )
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Stop Loss/Take Profit Orders Created!\n")
    fmt.Printf("Pair ID: %s\n", orderPair.PairID)
    fmt.Printf("\nStop Loss:\n")
    fmt.Printf("  Order ID: %s\n", orderPair.StopLossOrder.OrderID)
    fmt.Printf("  Price: $%.2f\n", stopLossPrice)
    fmt.Printf("  Risk: $%.2f (%.2f%%)\n", 
        math.Abs(entryPrice-stopLossPrice)*quantity,
        math.Abs((entryPrice-stopLossPrice)/entryPrice)*100)
    
    fmt.Printf("\nTake Profit:\n")
    fmt.Printf("  Order ID: %s\n", orderPair.TakeProfitOrder.OrderID)
    fmt.Printf("  Price: $%.2f\n", takeProfitPrice)
    fmt.Printf("  Reward: $%.2f (%.2f%%)\n",
        math.Abs(takeProfitPrice-entryPrice)*quantity,
        math.Abs((takeProfitPrice-entryPrice)/entryPrice)*100)
    
    // 计算风险回报比 / Calculate risk-reward ratio
    risk := math.Abs(entryPrice - stopLossPrice) * quantity
    reward := math.Abs(takeProfitPrice - entryPrice) * quantity
    ratio := reward / risk
    
    fmt.Printf("\nRisk-Reward Ratio: 1:%.2f\n", ratio)
}
```

### 示例 40: 合约移动止损 / Example 40: Futures Trailing Stop

```go
func futuresTrailingStop(symbol string, position *Position, trailPercent float64) {
    positionSide := position.PositionSide
    quantity := math.Abs(position.PositionAmt)
    
    // 设置移动止损 / Set trailing stop
    trailingStop, err := futuresStopLossService.SetTrailingStop(
        symbol,
        positionSide,
        quantity,
        trailPercent,
    )
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Trailing Stop Order Created!\n")
    fmt.Printf("Order ID: %s\n", trailingStop.OrderID)
    fmt.Printf("Trail Percent: %.2f%%\n", trailPercent)
    fmt.Printf("Initial Stop Price: $%.2f\n", trailingStop.CurrentStopPrice)
    
    // 监控移动止损 / Monitor trailing stop
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()
    
    for range ticker.C {
        // 获取最新标记价格 / Get latest mark price
        markPrice, err := futuresClient.GetMarkPrice(symbol)
        if err != nil {
            log.Printf("Error getting mark price: %v\n", err)
            continue
        }
        
        // 获取更新后的移动止损订单 / Get updated trailing stop order
        stopOrders, err := futuresStopLossService.GetActiveStopOrders(symbol)
        if err != nil {
            log.Printf("Error getting stop orders: %v\n", err)
            continue
        }
        
        // 找到我们的移动止损订单 / Find our trailing stop order
        var currentTrailingStop *TrailingStopOrder
        for _, order := range stopOrders {
            if order.OrderID == trailingStop.OrderID {
                if ts, ok := order.(*TrailingStopOrder); ok {
                    currentTrailingStop = ts
                    break
                }
            }
        }
        
        if currentTrailingStop == nil {
            fmt.Println("Trailing stop order triggered or cancelled")
            break
        }
        
        // 显示状态 / Display status
        fmt.Printf("\n=== %s ===\n", time.Now().Format("15:04:05"))
        fmt.Printf("Mark Price: $%.2f\n", markPrice.MarkPrice)
        fmt.Printf("Highest Price: $%.2f\n", currentTrailingStop.HighestPrice)
        fmt.Printf("Current Stop: $%.2f\n", currentTrailingStop.CurrentStopPrice)
        
        // 计算距离触发的距离 / Calculate distance to trigger
        var distance float64
        if positionSide == LONG {
            distance = markPrice.MarkPrice - currentTrailingStop.CurrentStopPrice
        } else {
            distance = currentTrailingStop.CurrentStopPrice - markPrice.MarkPrice
        }
        distancePercent := (distance / markPrice.MarkPrice) * 100
        
        fmt.Printf("Distance to trigger: $%.2f (%.2f%%)\n", distance, distancePercent)
        
        // 计算锁定的利润 / Calculate locked profit
        lockedProfit := math.Abs(currentTrailingStop.CurrentStopPrice - position.EntryPrice) * quantity
        lockedProfitPercent := (lockedProfit / (position.EntryPrice * quantity)) * 100
        
        fmt.Printf("Locked Profit: $%.2f (%.2f%%)\n", lockedProfit, lockedProfitPercent)
    }
}
```

### 示例 41: 合约风险监控 / Example 41: Futures Risk Monitoring

```go
func monitorFuturesRisk() {
    ticker := time.NewTicker(10 * time.Second)
    defer ticker.Stop()
    
    fmt.Println("Starting futures risk monitoring...")
    
    for range ticker.C {
        // 获取所有持仓 / Get all positions
        positions, err := futuresPositionMgr.GetAllPositions()
        if err != nil {
            log.Printf("Error getting positions: %v\n", err)
            continue
        }
        
        // 获取账户余额 / Get account balance
        balance, err := futuresClient.GetBalance()
        if err != nil {
            log.Printf("Error getting balance: %v\n", err)
            continue
        }
        
        // 计算风险指标 / Calculate risk metrics
        var totalPositionValue float64
        var totalUnrealizedPnL float64
        var totalMarginUsed float64
        positionsAtRisk := 0
        
        fmt.Printf("\n=== Risk Report %s ===\n", time.Now().Format("15:04:05"))
        fmt.Printf("Account Balance: $%.2f\n", balance.Balance)
        fmt.Printf("Available Balance: $%.2f\n", balance.AvailableBalance)
        
        for _, pos := range positions {
            if pos.PositionAmt == 0 {
                continue
            }
            
            posValue := math.Abs(pos.PositionAmt) * pos.MarkPrice
            totalPositionValue += posValue
            totalUnrealizedPnL += pos.UnrealizedProfit
            totalMarginUsed += pos.PositionInitialMargin
            
            // 检查强平风险 / Check liquidation risk
            distanceToLiq := math.Abs((pos.MarkPrice - pos.LiquidationPrice) / pos.MarkPrice)
            if distanceToLiq < 0.05 { // 5%
                positionsAtRisk++
                fmt.Printf("⚠️  %s %s: Close to liquidation (%.2f%% away)\n",
                    pos.Symbol, pos.PositionSide, distanceToLiq*100)
            }
            
            // 检查保证金率 / Check margin ratio
            marginRatio := pos.MaintenanceMargin / pos.PositionInitialMargin
            if marginRatio > 0.8 {
                fmt.Printf("⚠️  %s %s: High margin ratio (%.2f%%)\n",
                    pos.Symbol, pos.PositionSide, marginRatio*100)
            }
        }
        
        // 显示总体风险 / Display overall risk
        fmt.Printf("\nOverall Risk Metrics:\n")
        fmt.Printf("Total Position Value: $%.2f\n", totalPositionValue)
        fmt.Printf("Total Unrealized PnL: $%.2f\n", totalUnrealizedPnL)
        fmt.Printf("Total Margin Used: $%.2f\n", totalMarginUsed)
        fmt.Printf("Margin Utilization: %.2f%%\n", (totalMarginUsed/balance.Balance)*100)
        fmt.Printf("Positions at Risk: %d\n", positionsAtRisk)
        
        // 风险等级评估 / Risk level assessment
        marginUtil := (totalMarginUsed / balance.Balance) * 100
        var riskLevel string
        if marginUtil < 30 {
            riskLevel = "LOW ✅"
        } else if marginUtil < 60 {
            riskLevel = "MEDIUM ⚠️"
        } else {
            riskLevel = "HIGH 🚨"
        }
        fmt.Printf("Risk Level: %s\n", riskLevel)
        
        // 建议 / Recommendations
        if positionsAtRisk > 0 {
            fmt.Println("\n💡 Recommendations:")
            fmt.Println("  - Consider adding margin to at-risk positions")
            fmt.Println("  - Reduce position sizes")
            fmt.Println("  - Set stop loss orders")
        }
    }
}
```

### 示例 42: 合约与现货套利 / Example 42: Futures-Spot Arbitrage

```go
func futuresSpotArbitrage(symbol string, quantity float64) {
    // 获取现货价格 / Get spot price
    spotPrice, err := spotMarketService.GetCurrentPrice(symbol)
    if err != nil {
        log.Fatal(err)
    }
    
    // 获取合约标记价格 / Get futures mark price
    markPrice, err := futuresClient.GetMarkPrice(symbol)
    if err != nil {
        log.Fatal(err)
    }
    
    futuresPrice := markPrice.MarkPrice
    
    // 计算价差 / Calculate spread
    spread := futuresPrice - spotPrice
    spreadPercent := (spread / spotPrice) * 100
    
    fmt.Printf("Spot Price: $%.2f\n", spotPrice)
    fmt.Printf("Futures Price: $%.2f\n", futuresPrice)
    fmt.Printf("Spread: $%.2f (%.2f%%)\n", spread, spreadPercent)
    
    // 如果价差足够大，执行套利 / If spread is large enough, execute arbitrage
    minSpreadPercent := 0.5 // 0.5%
    
    if math.Abs(spreadPercent) > minSpreadPercent {
        fmt.Printf("\nArbitrage opportunity detected!\n")
        
        if spread > 0 {
            // 合约价格高于现货：买现货，卖合约 / Futures higher: buy spot, sell futures
            fmt.Println("Strategy: Buy spot, sell futures")
            
            // 买入现货 / Buy spot
            spotOrder, err := spotTradingService.PlaceMarketBuyOrder(symbol, quantity)
            if err != nil {
                log.Fatal(err)
            }
            fmt.Printf("Spot buy: $%.2f\n", spotOrder.Price)
            
            // 卖出合约（开空仓）/ Sell futures (open short)
            futuresOrder, err := futuresTradingService.OpenShortPosition(symbol, quantity, MARKET, 0)
            if err != nil {
                log.Fatal(err)
            }
            fmt.Printf("Futures short: $%.2f\n", futuresOrder.AvgPrice)
            
            // 计算预期利润 / Calculate expected profit
            actualSpread := futuresOrder.AvgPrice - spotOrder.Price
            expectedProfit := actualSpread * quantity
            
            fmt.Printf("\nExpected Profit: $%.2f (%.2f%%)\n", 
                expectedProfit, (expectedProfit/(spotOrder.Price*quantity))*100)
            
        } else {
            // 现货价格高于合约：卖现货，买合约 / Spot higher: sell spot, buy futures
            fmt.Println("Strategy: Sell spot, buy futures")
            
            // 卖出现货 / Sell spot
            spotOrder, err := spotTradingService.PlaceLimitSellOrder(symbol, spotPrice, quantity)
            if err != nil {
                log.Fatal(err)
            }
            fmt.Printf("Spot sell: $%.2f\n", spotOrder.Price)
            
            // 买入合约（开多仓）/ Buy futures (open long)
            futuresOrder, err := futuresTradingService.OpenLongPosition(symbol, quantity, MARKET, 0)
            if err != nil {
                log.Fatal(err)
            }
            fmt.Printf("Futures long: $%.2f\n", futuresOrder.AvgPrice)
            
            // 计算预期利润 / Calculate expected profit
            actualSpread := spotOrder.Price - futuresOrder.AvgPrice
            expectedProfit := actualSpread * quantity
            
            fmt.Printf("\nExpected Profit: $%.2f (%.2f%%)\n",
                expectedProfit, (expectedProfit/(futuresOrder.AvgPrice*quantity))*100)
        }
        
        // 监控套利平仓时机 / Monitor arbitrage closing opportunity
        fmt.Println("\nMonitoring for closing opportunity...")
        monitorArbitrageClose(symbol, quantity, spread > 0)
        
    } else {
        fmt.Println("Spread too small for arbitrage")
    }
}

func monitorArbitrageClose(symbol string, quantity float64, isShortFutures bool) {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for range ticker.C {
        // 获取当前价差 / Get current spread
        spotPrice, _ := spotMarketService.GetCurrentPrice(symbol)
        markPrice, _ := futuresClient.GetMarkPrice(symbol)
        
        currentSpread := markPrice.MarkPrice - spotPrice
        currentSpreadPercent := (currentSpread / spotPrice) * 100
        
        fmt.Printf("Current spread: $%.2f (%.2f%%)\n", currentSpread, currentSpreadPercent)
        
        // 如果价差收敛，平仓 / If spread converges, close positions
        if math.Abs(currentSpreadPercent) < 0.1 { // 0.1%
            fmt.Println("Spread converged! Closing arbitrage positions...")
            
            if isShortFutures {
                // 平空仓，卖现货 / Close short, sell spot
                futuresTradingService.ClosePosition(symbol, SHORT, quantity)
                spotTradingService.PlaceLimitSellOrder(symbol, spotPrice, quantity)
            } else {
                // 平多仓，买现货 / Close long, buy spot
                futuresTradingService.ClosePosition(symbol, LONG, quantity)
                spotTradingService.PlaceMarketBuyOrder(symbol, quantity)
            }
            
            fmt.Println("Arbitrage closed!")
            break
        }
    }
}
```

---

**更多合约交易信息请参阅 / For more futures trading information, see:**
- [Futures Quick Start Guide](FUTURES_QUICKSTART.md)
- [API Documentation](API.md)
- [README](../README.md)
