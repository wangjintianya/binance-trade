# API Documentation / API文档

本文档详细描述了币安自动交易系统的核心接口和数据模型。

This document provides detailed descriptions of the core interfaces and data models of the Binance Auto-Trading System.

## 目录 / Table of Contents

- [核心接口 / Core Interfaces](#核心接口--core-interfaces)
  - [BinanceClient](#binanceclient)
  - [TradingService](#tradingservice)
  - [RiskManager](#riskmanager)
  - [MarketDataService](#marketdataservice)
  - [OrderRepository](#orderrepository)
- [数据模型 / Data Models](#数据模型--data-models)
- [错误处理 / Error Handling](#错误处理--error-handling)
- [使用示例 / Usage Examples](#使用示例--usage-examples)

---

## 核心接口 / Core Interfaces

### BinanceClient

币安API客户端接口，负责所有与币安交易所的通信。

Binance API client interface responsible for all communication with Binance exchange.

**包路径 / Package:** `internal/api`

#### 方法 / Methods

##### GetAccountInfo

获取账户信息。

Get account information.

```go
GetAccountInfo() (*AccountInfo, error)
```

**返回 / Returns:**
- `*AccountInfo`: 账户信息，包含余额和权限 / Account info including balances and permissions
- `error`: 错误信息 / Error if any

**示例 / Example:**
```go
accountInfo, err := client.GetAccountInfo()
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Account Type: %s\n", accountInfo.AccountType)
```

---

##### GetBalance

获取指定资产的余额。

Get balance for a specific asset.

```go
GetBalance(asset string) (*Balance, error)
```

**参数 / Parameters:**
- `asset` (string): 资产符号，如 "BTC", "USDT" / Asset symbol, e.g., "BTC", "USDT"

**返回 / Returns:**
- `*Balance`: 余额信息 / Balance information
- `error`: 错误信息 / Error if any

**示例 / Example:**
```go
balance, err := client.GetBalance("USDT")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Free: %.8f, Locked: %.8f\n", balance.Free, balance.Locked)
```

---

##### GetPrice

获取指定交易对的当前价格。

Get current price for a trading pair.

```go
GetPrice(symbol string) (*Price, error)
```

**参数 / Parameters:**
- `symbol` (string): 交易对符号，如 "BTCUSDT" / Trading pair symbol, e.g., "BTCUSDT"

**返回 / Returns:**
- `*Price`: 价格信息 / Price information
- `error`: 错误信息 / Error if any

**示例 / Example:**
```go
price, err := client.GetPrice("BTCUSDT")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Current Price: %.8f\n", price.Price)
```

---

##### GetKlines

获取K线（蜡烛图）历史数据。

Get kline (candlestick) historical data.

```go
GetKlines(symbol string, interval string, limit int) ([]*Kline, error)
```

**参数 / Parameters:**
- `symbol` (string): 交易对符号 / Trading pair symbol
- `interval` (string): 时间间隔 / Time interval
  - 支持的值 / Supported values: `1m`, `5m`, `15m`, `30m`, `1h`, `4h`, `1d`, `1w`, `1M`
- `limit` (int): 返回的K线数量（最大1000）/ Number of klines to return (max 1000)

**返回 / Returns:**
- `[]*Kline`: K线数据数组 / Array of kline data
- `error`: 错误信息 / Error if any

**示例 / Example:**
```go
klines, err := client.GetKlines("BTCUSDT", "1h", 24)
if err != nil {
    log.Fatal(err)
}
for _, k := range klines {
    fmt.Printf("Time: %d, Open: %.2f, Close: %.2f\n", k.OpenTime, k.Open, k.Close)
}
```

---

##### CreateOrder

创建新订单。

Create a new order.

```go
CreateOrder(order *OrderRequest) (*OrderResponse, error)
```

**参数 / Parameters:**
- `order` (*OrderRequest): 订单请求对象 / Order request object

**返回 / Returns:**
- `*OrderResponse`: 订单响应，包含订单ID和状态 / Order response with order ID and status
- `error`: 错误信息 / Error if any

**示例 / Example:**
```go
orderReq := &OrderRequest{
    Symbol:   "BTCUSDT",
    Side:     "BUY",
    Type:     "MARKET",
    Quantity: 0.001,
}
orderResp, err := client.CreateOrder(orderReq)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Order ID: %d, Status: %s\n", orderResp.OrderID, orderResp.Status)
```

---

##### CancelOrder

取消指定订单。

Cancel a specific order.

```go
CancelOrder(symbol string, orderID int64) (*CancelResponse, error)
```

**参数 / Parameters:**
- `symbol` (string): 交易对符号 / Trading pair symbol
- `orderID` (int64): 订单ID / Order ID

**返回 / Returns:**
- `*CancelResponse`: 取消响应 / Cancel response
- `error`: 错误信息 / Error if any

**示例 / Example:**
```go
cancelResp, err := client.CancelOrder("BTCUSDT", 12345)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Cancelled Order ID: %d\n", cancelResp.OrderID)
```

---

##### GetOrder

获取指定订单的详细信息。

Get detailed information for a specific order.

```go
GetOrder(symbol string, orderID int64) (*Order, error)
```

**参数 / Parameters:**
- `symbol` (string): 交易对符号 / Trading pair symbol
- `orderID` (int64): 订单ID / Order ID

**返回 / Returns:**
- `*Order`: 订单详情 / Order details
- `error`: 错误信息 / Error if any

---

##### GetOpenOrders

获取所有未完成订单。

Get all open orders.

```go
GetOpenOrders(symbol string) ([]*Order, error)
```

**参数 / Parameters:**
- `symbol` (string): 交易对符号（空字符串表示所有交易对）/ Trading pair symbol (empty for all pairs)

**返回 / Returns:**
- `[]*Order`: 未完成订单列表 / List of open orders
- `error`: 错误信息 / Error if any

---

##### GetHistoricalOrders

获取历史订单。

Get historical orders.

```go
GetHistoricalOrders(symbol string, startTime, endTime int64) ([]*Order, error)
```

**参数 / Parameters:**
- `symbol` (string): 交易对符号 / Trading pair symbol
- `startTime` (int64): 开始时间（Unix毫秒时间戳）/ Start time (Unix milliseconds)
- `endTime` (int64): 结束时间（Unix毫秒时间戳）/ End time (Unix milliseconds)

**返回 / Returns:**
- `[]*Order`: 历史订单列表 / List of historical orders
- `error`: 错误信息 / Error if any

---

### TradingService

交易服务接口，实现核心交易逻辑。

Trading service interface implementing core trading logic.

**包路径 / Package:** `internal/service`

#### 方法 / Methods

##### PlaceMarketBuyOrder

下市价买入订单。

Place a market buy order.

```go
PlaceMarketBuyOrder(symbol string, quantity float64) (*Order, error)
```

**参数 / Parameters:**
- `symbol` (string): 交易对符号 / Trading pair symbol
- `quantity` (float64): 购买数量 / Quantity to buy

**返回 / Returns:**
- `*Order`: 订单详情 / Order details
- `error`: 错误信息 / Error if any

**风险检查 / Risk Checks:**
- 验证订单金额是否超过限制 / Validates order amount against limits
- 检查账户余额是否充足 / Checks account balance sufficiency
- 验证是否超过每日订单限制 / Validates daily order limit

**示例 / Example:**
```go
order, err := tradingService.PlaceMarketBuyOrder("BTCUSDT", 0.001)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Buy Order ID: %d\n", order.OrderID)
```

---

##### PlaceLimitSellOrder

下限价卖出订单。

Place a limit sell order.

```go
PlaceLimitSellOrder(symbol string, price, quantity float64) (*Order, error)
```

**参数 / Parameters:**
- `symbol` (string): 交易对符号 / Trading pair symbol
- `price` (float64): 限价价格 / Limit price
- `quantity` (float64): 卖出数量 / Quantity to sell

**返回 / Returns:**
- `*Order`: 订单详情 / Order details
- `error`: 错误信息 / Error if any

**风险检查 / Risk Checks:**
- 验证订单金额是否超过限制 / Validates order amount against limits
- 验证是否超过每日订单限制 / Validates daily order limit

**示例 / Example:**
```go
order, err := tradingService.PlaceLimitSellOrder("BTCUSDT", 50000.0, 0.001)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Sell Order ID: %d\n", order.OrderID)
```

---

##### CancelOrder

取消订单。

Cancel an order.

```go
CancelOrder(orderID int64) error
```

**参数 / Parameters:**
- `orderID` (int64): 订单ID / Order ID

**返回 / Returns:**
- `error`: 错误信息 / Error if any

---

##### GetOrderStatus

获取订单状态。

Get order status.

```go
GetOrderStatus(orderID int64) (*OrderStatus, error)
```

**参数 / Parameters:**
- `orderID` (int64): 订单ID / Order ID

**返回 / Returns:**
- `*OrderStatus`: 订单状态信息 / Order status information
- `error`: 错误信息 / Error if any

---

##### GetActiveOrders

获取所有活跃订单。

Get all active orders.

```go
GetActiveOrders() ([]*Order, error)
```

**返回 / Returns:**
- `[]*Order`: 活跃订单列表 / List of active orders
- `error`: 错误信息 / Error if any

---

### RiskManager

风险管理器接口，执行风险控制规则。

Risk manager interface executing risk control rules.

**包路径 / Package:** `internal/service`

#### 方法 / Methods

##### ValidateOrder

验证订单是否符合风险控制规则。

Validate if an order complies with risk control rules.

```go
ValidateOrder(order *OrderRequest) error
```

**参数 / Parameters:**
- `order` (*OrderRequest): 订单请求 / Order request

**返回 / Returns:**
- `error`: 如果订单违反风险规则则返回错误 / Error if order violates risk rules

**检查项 / Checks:**
- 订单金额是否超过最大限额 / Order amount exceeds max limit
- 是否超过每日订单数量限制 / Exceeds daily order count limit
- 买入订单是否会导致余额低于最小保留 / Buy order would cause balance below minimum reserve

---

##### CheckDailyLimit

检查是否超过每日订单限制。

Check if daily order limit is exceeded.

```go
CheckDailyLimit() error
```

**返回 / Returns:**
- `error`: 如果超过限制则返回错误 / Error if limit exceeded

---

##### CheckMinimumBalance

检查指定资产的余额是否满足最小保留要求。

Check if balance for specified asset meets minimum reserve requirement.

```go
CheckMinimumBalance(asset string) error
```

**参数 / Parameters:**
- `asset` (string): 资产符号 / Asset symbol

**返回 / Returns:**
- `error`: 如果余额不足则返回错误 / Error if balance insufficient

---

##### UpdateLimits

更新风险限制配置。

Update risk limit configuration.

```go
UpdateLimits(limits *RiskLimits) error
```

**参数 / Parameters:**
- `limits` (*RiskLimits): 新的风险限制 / New risk limits

**返回 / Returns:**
- `error`: 错误信息 / Error if any

---

##### GetCurrentLimits

获取当前风险限制配置。

Get current risk limit configuration.

```go
GetCurrentLimits() *RiskLimits
```

**返回 / Returns:**
- `*RiskLimits`: 当前风险限制 / Current risk limits

---

### MarketDataService

市场数据服务接口，提供市场信息。

Market data service interface providing market information.

**包路径 / Package:** `internal/service`

#### 方法 / Methods

##### GetCurrentPrice

获取当前价格（带缓存）。

Get current price (with caching).

```go
GetCurrentPrice(symbol string) (float64, error)
```

**参数 / Parameters:**
- `symbol` (string): 交易对符号 / Trading pair symbol

**返回 / Returns:**
- `float64`: 当前价格 / Current price
- `error`: 错误信息 / Error if any

**注意 / Note:** 价格数据会缓存1秒以减少API调用 / Price data is cached for 1 second to reduce API calls

---

##### GetHistoricalData

获取历史K线数据。

Get historical kline data.

```go
GetHistoricalData(symbol string, interval string, limit int) ([]*Kline, error)
```

**参数 / Parameters:**
- `symbol` (string): 交易对符号 / Trading pair symbol
- `interval` (string): 时间间隔 / Time interval
- `limit` (int): 数据数量 / Data count

**返回 / Returns:**
- `[]*Kline`: K线数据 / Kline data
- `error`: 错误信息 / Error if any

---

### OrderRepository

订单仓储接口，管理订单数据持久化。

Order repository interface managing order data persistence.

**包路径 / Package:** `internal/repository`

#### 方法 / Methods

##### Save

保存订单。

Save an order.

```go
Save(order *Order) error
```

---

##### GetByID

根据ID获取订单。

Get order by ID.

```go
GetByID(orderID int64) (*Order, error)
```

---

##### GetAll

获取所有订单。

Get all orders.

```go
GetAll() ([]*Order, error)
```

---

##### GetBySymbol

获取指定交易对的所有订单。

Get all orders for a specific symbol.

```go
GetBySymbol(symbol string) ([]*Order, error)
```

---

##### GetOpenOrders

获取所有未完成订单。

Get all open orders.

```go
GetOpenOrders() ([]*Order, error)
```

---

##### Update

更新订单。

Update an order.

```go
Update(order *Order) error
```

---

##### Delete

删除订单。

Delete an order.

```go
Delete(orderID int64) error
```

---

## 数据模型 / Data Models

### OrderRequest

订单请求对象。

Order request object.

```go
type OrderRequest struct {
    Symbol      string      // 交易对 / Trading pair, e.g., "BTCUSDT"
    Side        string      // 方向 / Side: "BUY" or "SELL"
    Type        string      // 类型 / Type: "MARKET" or "LIMIT"
    Quantity    float64     // 数量 / Quantity
    Price       float64     // 价格（限价单）/ Price (for limit orders)
    TimeInForce string      // 有效期 / Time in force: "GTC", "IOC", "FOK"
}
```

**字段说明 / Field Descriptions:**

- **Symbol**: 交易对符号，如 "BTCUSDT" / Trading pair symbol, e.g., "BTCUSDT"
- **Side**: 订单方向 / Order side
  - `BUY`: 买入 / Buy
  - `SELL`: 卖出 / Sell
- **Type**: 订单类型 / Order type
  - `MARKET`: 市价单 / Market order
  - `LIMIT`: 限价单 / Limit order
- **Quantity**: 交易数量 / Trading quantity
- **Price**: 价格（仅限价单需要）/ Price (required for limit orders only)
- **TimeInForce**: 订单有效期 / Order time in force
  - `GTC` (Good Till Cancel): 一直有效直到取消 / Valid until cancelled
  - `IOC` (Immediate Or Cancel): 立即成交或取消 / Immediate or cancel
  - `FOK` (Fill Or Kill): 全部成交或取消 / Fill or kill

---

### Order

订单对象。

Order object.

```go
type Order struct {
    OrderID              int64       // 订单ID / Order ID
    Symbol               string      // 交易对 / Trading pair
    Side                 string      // 方向 / Side
    Type                 string      // 类型 / Type
    Status               string      // 状态 / Status
    Price                float64     // 价格 / Price
    OrigQty              float64     // 原始数量 / Original quantity
    ExecutedQty          float64     // 已执行数量 / Executed quantity
    CummulativeQuoteQty  float64     // 累计成交金额 / Cumulative quote quantity
    Time                 int64       // 创建时间 / Creation time
    UpdateTime           int64       // 更新时间 / Update time
}
```

**状态值 / Status Values:**
- `NEW`: 新订单 / New order
- `PARTIALLY_FILLED`: 部分成交 / Partially filled
- `FILLED`: 完全成交 / Filled
- `CANCELED`: 已取消 / Canceled
- `REJECTED`: 已拒绝 / Rejected
- `EXPIRED`: 已过期 / Expired

---

### Balance

余额对象。

Balance object.

```go
type Balance struct {
    Asset  string   // 资产符号 / Asset symbol
    Free   float64  // 可用余额 / Free balance
    Locked float64  // 锁定余额 / Locked balance
}
```

---

### Kline

K线数据对象。

Kline data object.

```go
type Kline struct {
    OpenTime  int64    // 开盘时间 / Open time
    Open      float64  // 开盘价 / Open price
    High      float64  // 最高价 / High price
    Low       float64  // 最低价 / Low price
    Close     float64  // 收盘价 / Close price
    Volume    float64  // 成交量 / Volume
    CloseTime int64    // 收盘时间 / Close time
}
```

---

### RiskLimits

风险限制配置。

Risk limits configuration.

```go
type RiskLimits struct {
    MaxOrderAmount    float64  // 单笔最大金额 / Max order amount
    MaxDailyOrders    int      // 每日最大订单数 / Max daily orders
    MinBalanceReserve float64  // 最小保留余额 / Min balance reserve
    MaxAPICallsPerMin int      // 每分钟最大API调用 / Max API calls per minute
}
```

---

## 错误处理 / Error Handling

### 错误类型 / Error Types

系统定义了以下错误类型 / The system defines the following error types:

```go
type ErrorType int

const (
    ErrNetwork           ErrorType = iota  // 网络错误 / Network error
    ErrAuthentication                      // 认证错误 / Authentication error
    ErrRateLimit                           // 速率限制 / Rate limit error
    ErrInsufficientBalance                 // 余额不足 / Insufficient balance
    ErrInvalidParameter                    // 参数无效 / Invalid parameter
    ErrOrderNotFound                       // 订单未找到 / Order not found
    ErrRiskLimitExceeded                   // 超过风险限制 / Risk limit exceeded
)
```

### TradingError

交易错误对象。

Trading error object.

```go
type TradingError struct {
    Type    ErrorType  // 错误类型 / Error type
    Message string     // 错误消息 / Error message
    Code    int        // 错误代码 / Error code
    Cause   error      // 原始错误 / Original error
}
```

### 错误处理示例 / Error Handling Example

```go
order, err := tradingService.PlaceMarketBuyOrder("BTCUSDT", 0.001)
if err != nil {
    if tradingErr, ok := err.(*TradingError); ok {
        switch tradingErr.Type {
        case ErrInsufficientBalance:
            fmt.Println("Insufficient balance, please deposit funds")
        case ErrRiskLimitExceeded:
            fmt.Println("Order exceeds risk limits")
        case ErrRateLimit:
            fmt.Println("Rate limit exceeded, please wait")
        default:
            fmt.Printf("Trading error: %s\n", tradingErr.Message)
        }
    } else {
        fmt.Printf("Unknown error: %v\n", err)
    }
    return
}
```

---

## 使用示例 / Usage Examples

### 完整交易流程示例 / Complete Trading Flow Example

```go
package main

import (
    "fmt"
    "log"
    "time"
    
    "binance-trader/internal/api"
    "binance-trader/internal/service"
    "binance-trader/internal/repository"
    "binance-trader/pkg/logger"
)

func main() {
    // 1. 初始化组件 / Initialize components
    authMgr, _ := api.NewAuthManager(apiKey, apiSecret)
    rateLimiter := api.NewRateLimiter(1000)
    httpClient := api.NewHTTPClient(rateLimiter, retryConfig)
    binanceClient, _ := api.NewBinanceClient(baseURL, httpClient, authMgr)
    
    orderRepo := repository.NewMemoryOrderRepository()
    riskMgr := service.NewRiskManager(riskLimits, binanceClient)
    tradingService := service.NewTradingService(binanceClient, riskMgr, orderRepo, log)
    marketService := service.NewMarketDataService(binanceClient, 1*time.Second)
    
    // 2. 查询当前价格 / Query current price
    price, err := marketService.GetCurrentPrice("BTCUSDT")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Current BTC price: %.2f USDT\n", price)
    
    // 3. 下市价买单 / Place market buy order
    buyOrder, err := tradingService.PlaceMarketBuyOrder("BTCUSDT", 0.001)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Buy order placed: ID=%d, Status=%s\n", buyOrder.OrderID, buyOrder.Status)
    
    // 4. 等待订单成交 / Wait for order to fill
    time.Sleep(2 * time.Second)
    
    // 5. 查询订单状态 / Query order status
    status, err := tradingService.GetOrderStatus(buyOrder.OrderID)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Order status: %s, Executed: %.8f\n", status.Status, status.ExecutedQty)
    
    // 6. 下限价卖单 / Place limit sell order
    sellPrice := price * 1.02  // 2% higher
    sellOrder, err := tradingService.PlaceLimitSellOrder("BTCUSDT", sellPrice, 0.001)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Sell order placed: ID=%d, Price=%.2f\n", sellOrder.OrderID, sellPrice)
    
    // 7. 查看活跃订单 / View active orders
    activeOrders, err := tradingService.GetActiveOrders()
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Active orders: %d\n", len(activeOrders))
    
    // 8. 取消订单（如果需要）/ Cancel order (if needed)
    // err = tradingService.CancelOrder(sellOrder.OrderID)
}
```

### 市场数据分析示例 / Market Data Analysis Example

```go
// 获取24小时K线数据并计算平均价格
// Get 24-hour kline data and calculate average price
klines, err := marketService.GetHistoricalData("BTCUSDT", "1h", 24)
if err != nil {
    log.Fatal(err)
}

var sum float64
for _, k := range klines {
    sum += k.Close
}
avgPrice := sum / float64(len(klines))
fmt.Printf("24h average price: %.2f\n", avgPrice)

// 找出最高价和最低价
// Find highest and lowest prices
var high, low float64 = klines[0].High, klines[0].Low
for _, k := range klines {
    if k.High > high {
        high = k.High
    }
    if k.Low < low {
        low = k.Low
    }
}
fmt.Printf("24h High: %.2f, Low: %.2f\n", high, low)
```

### 风险管理示例 / Risk Management Example

```go
// 检查当前风险限制
// Check current risk limits
limits := riskMgr.GetCurrentLimits()
fmt.Printf("Max order amount: %.2f USDT\n", limits.MaxOrderAmount)
fmt.Printf("Max daily orders: %d\n", limits.MaxDailyOrders)

// 更新风险限制
// Update risk limits
newLimits := &service.RiskLimits{
    MaxOrderAmount:    5000.0,
    MaxDailyOrders:    50,
    MinBalanceReserve: 200.0,
    MaxAPICallsPerMin: 800,
}
err := riskMgr.UpdateLimits(newLimits)
if err != nil {
    log.Fatal(err)
}

// 验证订单
// Validate order
orderReq := &OrderRequest{
    Symbol:   "BTCUSDT",
    Side:     "BUY",
    Type:     "MARKET",
    Quantity: 0.1,
}
err = riskMgr.ValidateOrder(orderReq)
if err != nil {
    fmt.Printf("Order validation failed: %v\n", err)
}
```

---

## 最佳实践 / Best Practices

1. **错误处理** / **Error Handling**
   - 始终检查错误返回值 / Always check error return values
   - 使用类型断言处理特定错误类型 / Use type assertions for specific error types
   - 记录详细的错误日志 / Log detailed error information

2. **资源管理** / **Resource Management**
   - 使用defer确保资源释放 / Use defer to ensure resource cleanup
   - 避免goroutine泄漏 / Avoid goroutine leaks
   - 正确处理context取消 / Properly handle context cancellation

3. **并发安全** / **Concurrency Safety**
   - 使用互斥锁保护共享数据 / Use mutexes to protect shared data
   - 避免数据竞争 / Avoid data races
   - 使用channel进行goroutine通信 / Use channels for goroutine communication

4. **性能优化** / **Performance Optimization**
   - 使用缓存减少API调用 / Use caching to reduce API calls
   - 批量处理订单查询 / Batch order queries
   - 合理设置超时时间 / Set reasonable timeout values

5. **安全性** / **Security**
   - 永远不要记录敏感信息 / Never log sensitive information
   - 使用HTTPS进行所有通信 / Use HTTPS for all communication
   - 定期轮换API密钥 / Rotate API keys regularly

---

**更多示例请参阅 / For more examples, see:** [EXAMPLES.md](EXAMPLES.md)
