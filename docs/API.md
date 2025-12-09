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
  - [ConditionalOrderService](#conditionalorderservice)
  - [StopLossService](#stoplossservice)
  - [TriggerEngine](#triggerengine)
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

### ConditionalOrderService

条件订单服务接口，管理条件触发订单。

Conditional order service interface managing condition-triggered orders.

**包路径 / Package:** `internal/service`

#### 方法 / Methods

##### CreateConditionalOrder

创建条件订单。

Create a conditional order.

```go
CreateConditionalOrder(order *ConditionalOrderRequest) (*ConditionalOrder, error)
```

**参数 / Parameters:**
- `order` (*ConditionalOrderRequest): 条件订单请求 / Conditional order request

**返回 / Returns:**
- `*ConditionalOrder`: 条件订单详情 / Conditional order details
- `error`: 错误信息 / Error if any

**示例 / Example:**
```go
triggerCondition := &TriggerCondition{
    Type:     TriggerTypePrice,
    Operator: OperatorGreaterEqual,
    Value:    45000.0,
}

orderReq := &ConditionalOrderRequest{
    Symbol:           "BTCUSDT",
    Side:             "BUY",
    Type:             "MARKET",
    Quantity:         0.001,
    TriggerCondition: triggerCondition,
}

order, err := conditionalOrderService.CreateConditionalOrder(orderReq)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Conditional Order ID: %s\n", order.OrderID)
```

---

##### CancelConditionalOrder

取消条件订单。

Cancel a conditional order.

```go
CancelConditionalOrder(orderID string) error
```

**参数 / Parameters:**
- `orderID` (string): 条件订单ID / Conditional order ID

**返回 / Returns:**
- `error`: 错误信息 / Error if any

---

##### GetActiveConditionalOrders

获取所有活跃的条件订单。

Get all active conditional orders.

```go
GetActiveConditionalOrders() ([]*ConditionalOrder, error)
```

**返回 / Returns:**
- `[]*ConditionalOrder`: 活跃条件订单列表 / List of active conditional orders
- `error`: 错误信息 / Error if any

---

##### GetConditionalOrderHistory

获取条件订单历史。

Get conditional order history.

```go
GetConditionalOrderHistory(startTime, endTime int64) ([]*ConditionalOrder, error)
```

**参数 / Parameters:**
- `startTime` (int64): 开始时间（Unix毫秒时间戳）/ Start time (Unix milliseconds)
- `endTime` (int64): 结束时间（Unix毫秒时间戳）/ End time (Unix milliseconds)

**返回 / Returns:**
- `[]*ConditionalOrder`: 条件订单历史列表 / List of historical conditional orders
- `error`: 错误信息 / Error if any

---

### StopLossService

止损止盈服务接口，管理风险控制订单。

Stop-loss service interface managing risk control orders.

**包路径 / Package:** `internal/service`

#### 方法 / Methods

##### SetStopLoss

为持仓设置止损。

Set stop loss for a position.

```go
SetStopLoss(symbol string, position float64, stopPrice float64) (*StopOrder, error)
```

**参数 / Parameters:**
- `symbol` (string): 交易对符号 / Trading pair symbol
- `position` (float64): 持仓数量 / Position quantity
- `stopPrice` (float64): 止损价格 / Stop loss price

**返回 / Returns:**
- `*StopOrder`: 止损订单详情 / Stop order details
- `error`: 错误信息 / Error if any

**示例 / Example:**
```go
stopOrder, err := stopLossService.SetStopLoss("BTCUSDT", 0.001, 42000.0)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Stop Loss Order ID: %s\n", stopOrder.OrderID)
```

---

##### SetTakeProfit

为持仓设置止盈。

Set take profit for a position.

```go
SetTakeProfit(symbol string, position float64, targetPrice float64) (*StopOrder, error)
```

**参数 / Parameters:**
- `symbol` (string): 交易对符号 / Trading pair symbol
- `position` (float64): 持仓数量 / Position quantity
- `targetPrice` (float64): 止盈价格 / Take profit price

**返回 / Returns:**
- `*StopOrder`: 止盈订单详情 / Take profit order details
- `error`: 错误信息 / Error if any

---

##### SetStopLossTakeProfit

同时设置止损和止盈（配对订单）。

Set both stop-loss and take-profit (paired orders).

```go
SetStopLossTakeProfit(symbol string, position float64, stopPrice, targetPrice float64) (*StopOrderPair, error)
```

**参数 / Parameters:**
- `symbol` (string): 交易对符号 / Trading pair symbol
- `position` (float64): 持仓数量 / Position quantity
- `stopPrice` (float64): 止损价格 / Stop loss price
- `targetPrice` (float64): 止盈价格 / Take profit price

**返回 / Returns:**
- `*StopOrderPair`: 配对订单详情 / Paired order details
- `error`: 错误信息 / Error if any

**注意 / Note:** 当任一订单触发时，另一个订单将自动取消 / When either order triggers, the other will be automatically cancelled

---

##### SetTrailingStop

设置移动止损。

Set trailing stop.

```go
SetTrailingStop(symbol string, position float64, trailPercent float64) (*TrailingStopOrder, error)
```

**参数 / Parameters:**
- `symbol` (string): 交易对符号 / Trading pair symbol
- `position` (float64): 持仓数量 / Position quantity
- `trailPercent` (float64): 移动止损百分比 / Trailing stop percentage

**返回 / Returns:**
- `*TrailingStopOrder`: 移动止损订单详情 / Trailing stop order details
- `error`: 错误信息 / Error if any

**示例 / Example:**
```go
// 设置2%的移动止损
// Set 2% trailing stop
trailingStop, err := stopLossService.SetTrailingStop("BTCUSDT", 0.001, 2.0)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Trailing Stop Order ID: %s\n", trailingStop.OrderID)
```

---

##### GetActiveStopOrders

获取指定交易对的所有活跃止损止盈订单。

Get all active stop orders for a specific symbol.

```go
GetActiveStopOrders(symbol string) ([]*StopOrder, error)
```

**参数 / Parameters:**
- `symbol` (string): 交易对符号 / Trading pair symbol

**返回 / Returns:**
- `[]*StopOrder`: 活跃止损止盈订单列表 / List of active stop orders
- `error`: 错误信息 / Error if any

---

### TriggerEngine

触发引擎接口，评估和执行触发条件。

Trigger engine interface evaluating and executing trigger conditions.

**包路径 / Package:** `internal/service`

#### 方法 / Methods

##### EvaluateCondition

评估单个触发条件。

Evaluate a single trigger condition.

```go
EvaluateCondition(condition *TriggerCondition, marketData *MarketData) (bool, error)
```

**参数 / Parameters:**
- `condition` (*TriggerCondition): 触发条件 / Trigger condition
- `marketData` (*MarketData): 市场数据 / Market data

**返回 / Returns:**
- `bool`: 条件是否满足 / Whether condition is met
- `error`: 错误信息 / Error if any

---

##### EvaluateCompositeCondition

评估复合触发条件（AND/OR逻辑）。

Evaluate composite trigger condition (AND/OR logic).

```go
EvaluateCompositeCondition(conditions []*TriggerCondition, operator LogicOperator, marketData *MarketData) (bool, error)
```

**参数 / Parameters:**
- `conditions` ([]*TriggerCondition): 触发条件列表 / List of trigger conditions
- `operator` (LogicOperator): 逻辑运算符（AND/OR）/ Logic operator (AND/OR)
- `marketData` (*MarketData): 市场数据 / Market data

**返回 / Returns:**
- `bool`: 复合条件是否满足 / Whether composite condition is met
- `error`: 错误信息 / Error if any

**示例 / Example:**
```go
// 创建复合条件：价格 > 45000 AND 成交量 > 1000
// Create composite condition: price > 45000 AND volume > 1000
conditions := []*TriggerCondition{
    {
        Type:     TriggerTypePrice,
        Operator: OperatorGreaterThan,
        Value:    45000.0,
    },
    {
        Type:       TriggerTypeVolume,
        Operator:   OperatorGreaterThan,
        Value:      1000.0,
        TimeWindow: 1 * time.Hour,
    },
}

marketData := &MarketData{
    Symbol:    "BTCUSDT",
    Price:     45500.0,
    Volume24h: 1200.0,
    Timestamp: time.Now().UnixMilli(),
}

triggered, err := triggerEngine.EvaluateCompositeCondition(conditions, LogicAND, marketData)
if err != nil {
    log.Fatal(err)
}

if triggered {
    fmt.Println("Composite condition triggered!")
}
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

### ConditionalOrderRequest

条件订单请求对象。

Conditional order request object.

```go
type ConditionalOrderRequest struct {
    Symbol           string              // 交易对 / Trading pair
    Side             string              // 方向 / Side: "BUY" or "SELL"
    Type             string              // 类型 / Type: "MARKET" or "LIMIT"
    Quantity         float64             // 交易数量 / Quantity
    Price            float64             // 价格（限价单）/ Price (for limit orders)
    TriggerCondition *TriggerCondition   // 触发条件 / Trigger condition
    TimeWindow       *TimeWindow         // 时间窗口限制（可选）/ Time window restriction (optional)
}
```

---

### ConditionalOrder

条件订单对象。

Conditional order object.

```go
type ConditionalOrder struct {
    OrderID          string                  // 条件订单ID / Conditional order ID
    Symbol           string                  // 交易对 / Trading pair
    Side             string                  // 方向 / Side
    Type             string                  // 类型 / Type
    Quantity         float64                 // 交易数量 / Quantity
    Price            float64                 // 价格 / Price
    TriggerCondition *TriggerCondition       // 触发条件 / Trigger condition
    Status           ConditionalOrderStatus  // 状态 / Status
    CreatedAt        int64                   // 创建时间 / Creation time
    TriggeredAt      int64                   // 触发时间 / Trigger time
    ExecutedOrderID  int64                   // 执行的订单ID / Executed order ID
    TimeWindow       *TimeWindow             // 时间窗口 / Time window
}
```

**状态值 / Status Values:**
- `PENDING`: 等待触发 / Pending trigger
- `TRIGGERED`: 已触发 / Triggered
- `EXECUTED`: 已执行 / Executed
- `CANCELLED`: 已取消 / Cancelled
- `EXPIRED`: 已过期 / Expired

---

### TriggerCondition

触发条件对象。

Trigger condition object.

```go
type TriggerCondition struct {
    Type          TriggerType         // 触发类型 / Trigger type
    Operator      ComparisonOperator  // 比较运算符 / Comparison operator
    Value         float64             // 触发值 / Trigger value
    BasePrice     float64             // 基准价格（用于涨跌幅）/ Base price (for percentage change)
    TimeWindow    time.Duration       // 时间窗口（用于成交量）/ Time window (for volume)
    CompositeType LogicOperator       // 复合类型（AND/OR）/ Composite type (AND/OR)
    SubConditions []*TriggerCondition // 子条件 / Sub-conditions
}
```

**触发类型 / Trigger Types:**
- `TriggerTypePrice`: 价格触发 / Price trigger
- `TriggerTypePriceChangePercent`: 涨跌幅触发 / Percentage change trigger
- `TriggerTypeVolume`: 成交量触发 / Volume trigger

**比较运算符 / Comparison Operators:**
- `OperatorGreaterThan`: 大于 / Greater than (>)
- `OperatorLessThan`: 小于 / Less than (<)
- `OperatorGreaterEqual`: 大于等于 / Greater than or equal (>=)
- `OperatorLessEqual`: 小于等于 / Less than or equal (<=)

**逻辑运算符 / Logic Operators:**
- `LogicAND`: 与逻辑 / AND logic
- `LogicOR`: 或逻辑 / OR logic

---

### StopOrder

止损止盈订单对象。

Stop-loss/take-profit order object.

```go
type StopOrder struct {
    OrderID         string          // 订单ID / Order ID
    Symbol          string          // 交易对 / Trading pair
    Position        float64         // 持仓数量 / Position quantity
    StopPrice       float64         // 止损/止盈价格 / Stop/target price
    Type            StopOrderType   // 订单类型 / Order type
    Status          StopOrderStatus // 状态 / Status
    CreatedAt       int64           // 创建时间 / Creation time
    TriggeredAt     int64           // 触发时间 / Trigger time
    ExecutedOrderID int64           // 执行的订单ID / Executed order ID
}
```

**订单类型 / Order Types:**
- `StopOrderTypeStopLoss`: 止损订单 / Stop loss order
- `StopOrderTypeTakeProfit`: 止盈订单 / Take profit order

**状态值 / Status Values:**
- `ACTIVE`: 活跃 / Active
- `TRIGGERED`: 已触发 / Triggered
- `CANCELLED`: 已取消 / Cancelled

---

### StopOrderPair

止损止盈配对订单对象。

Stop-loss/take-profit paired order object.

```go
type StopOrderPair struct {
    PairID          string      // 配对ID / Pair ID
    Symbol          string      // 交易对 / Trading pair
    Position        float64     // 持仓数量 / Position quantity
    StopLossOrder   *StopOrder  // 止损订单 / Stop loss order
    TakeProfitOrder *StopOrder  // 止盈订单 / Take profit order
    Status          string      // 状态 / Status
}
```

**状态值 / Status Values:**
- `ACTIVE`: 活跃 / Active
- `PARTIALLY_TRIGGERED`: 部分触发 / Partially triggered
- `COMPLETED`: 已完成 / Completed

---

### TrailingStopOrder

移动止损订单对象。

Trailing stop order object.

```go
type TrailingStopOrder struct {
    OrderID          string          // 订单ID / Order ID
    Symbol           string          // 交易对 / Trading pair
    Position         float64         // 持仓数量 / Position quantity
    TrailPercent     float64         // 移动止损百分比 / Trail percentage
    HighestPrice     float64         // 记录的最高价格 / Recorded highest price
    CurrentStopPrice float64         // 当前止损价格 / Current stop price
    Status           StopOrderStatus // 状态 / Status
    CreatedAt        int64           // 创建时间 / Creation time
    LastUpdatedAt    int64           // 最后更新时间 / Last update time
}
```

---

### MarketData

市场数据对象。

Market data object.

```go
type MarketData struct {
    Symbol             string  // 交易对 / Trading pair
    Price              float64 // 当前价格 / Current price
    Volume24h          float64 // 24小时成交量 / 24h volume
    Timestamp          int64   // 时间戳 / Timestamp
    PriceChange        float64 // 价格变化 / Price change
    PriceChangePercent float64 // 价格变化百分比 / Price change percentage
}
```

---

### TimeWindow

时间窗口对象。

Time window object.

```go
type TimeWindow struct {
    StartTime time.Time // 开始时间 / Start time
    EndTime   time.Time // 结束时间 / End time
}
```

---

## 错误处理 / Error Handling

### 错误类型 / Error Types

系统定义了以下错误类型 / The system defines the following error types:

```go
type ErrorType int

const (
    ErrNetwork                  ErrorType = iota  // 网络错误 / Network error
    ErrAuthentication                             // 认证错误 / Authentication error
    ErrRateLimit                                  // 速率限制 / Rate limit error
    ErrInsufficientBalance                        // 余额不足 / Insufficient balance
    ErrInvalidParameter                           // 参数无效 / Invalid parameter
    ErrOrderNotFound                              // 订单未找到 / Order not found
    ErrRiskLimitExceeded                          // 超过风险限制 / Risk limit exceeded
    ErrInvalidTriggerCondition                    // 无效触发条件 / Invalid trigger condition
    ErrConditionalOrderNotFound                   // 条件订单未找到 / Conditional order not found
    ErrStopOrderNotFound                          // 止损订单未找到 / Stop order not found
    ErrOrderAlreadyTriggered                      // 订单已触发 / Order already triggered
    ErrTimeWindowExpired                          // 时间窗口已过期 / Time window expired
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


---

## 合约交易接口 / Futures Trading Interfaces

### FuturesClient

合约API客户端接口，负责所有与币安合约交易所的通信。

Futures API client interface responsible for all communication with Binance Futures exchange.

**包路径 / Package:** `internal/api`

#### 方法 / Methods

##### GetAccountInfo

获取合约账户信息。

Get futures account information.

```go
GetAccountInfo() (*FuturesAccountInfo, error)
```

**返回 / Returns:**
- `*FuturesAccountInfo`: 合约账户信息 / Futures account info
- `error`: 错误信息 / Error if any

---

##### GetBalance

获取合约账户余额。

Get futures account balance.

```go
GetBalance() (*FuturesBalance, error)
```

**返回 / Returns:**
- `*FuturesBalance`: 合约余额信息 / Futures balance information
- `error`: 错误信息 / Error if any

---

##### GetMarkPrice

获取标记价格。

Get mark price.

```go
GetMarkPrice(symbol string) (*MarkPrice, error)
```

**参数 / Parameters:**
- `symbol` (string): 合约交易对，如 "BTCUSDT" / Futures trading pair, e.g., "BTCUSDT"

**返回 / Returns:**
- `*MarkPrice`: 标记价格信息 / Mark price information
- `error`: 错误信息 / Error if any

---

##### GetFundingRate

获取资金费率。

Get funding rate.

```go
GetFundingRate(symbol string) (*FundingRate, error)
```

**参数 / Parameters:**
- `symbol` (string): 合约交易对 / Futures trading pair

**返回 / Returns:**
- `*FundingRate`: 资金费率信息 / Funding rate information
- `error`: 错误信息 / Error if any

---

##### SetLeverage

设置杠杆倍数。

Set leverage.

```go
SetLeverage(symbol string, leverage int) error
```

**参数 / Parameters:**
- `symbol` (string): 合约交易对 / Futures trading pair
- `leverage` (int): 杠杆倍数（1-125）/ Leverage (1-125)

**返回 / Returns:**
- `error`: 错误信息 / Error if any

---

##### SetMarginType

设置保证金模式。

Set margin type.

```go
SetMarginType(symbol string, marginType MarginType) error
```

**参数 / Parameters:**
- `symbol` (string): 合约交易对 / Futures trading pair
- `marginType` (MarginType): 保证金模式 / Margin type
  - `CROSSED`: 全仓保证金 / Cross margin
  - `ISOLATED`: 逐仓保证金 / Isolated margin

**返回 / Returns:**
- `error`: 错误信息 / Error if any

---

##### CreateOrder

创建合约订单。

Create futures order.

```go
CreateOrder(order *FuturesOrderRequest) (*FuturesOrderResponse, error)
```

**参数 / Parameters:**
- `order` (*FuturesOrderRequest): 合约订单请求 / Futures order request

**返回 / Returns:**
- `*FuturesOrderResponse`: 订单响应 / Order response
- `error`: 错误信息 / Error if any

---

##### GetPositions

获取持仓信息。

Get position information.

```go
GetPositions(symbol string) ([]*Position, error)
```

**参数 / Parameters:**
- `symbol` (string): 合约交易对（空字符串表示所有）/ Futures trading pair (empty for all)

**返回 / Returns:**
- `[]*Position`: 持仓列表 / List of positions
- `error`: 错误信息 / Error if any

---

### FuturesTradingService

合约交易服务接口，实现核心合约交易逻辑。

Futures trading service interface implementing core futures trading logic.

**包路径 / Package:** `internal/service`

#### 方法 / Methods

##### OpenLongPosition

开多仓。

Open long position.

```go
OpenLongPosition(symbol string, quantity float64, orderType OrderType, price float64) (*FuturesOrder, error)
```

**参数 / Parameters:**
- `symbol` (string): 合约交易对 / Futures trading pair
- `quantity` (float64): 数量 / Quantity
- `orderType` (OrderType): 订单类型 / Order type
  - `MARKET`: 市价单 / Market order
  - `LIMIT`: 限价单 / Limit order
- `price` (float64): 价格（限价单需要）/ Price (for limit orders)

**返回 / Returns:**
- `*FuturesOrder`: 订单详情 / Order details
- `error`: 错误信息 / Error if any

**示例 / Example:**
```go
// 市价开多仓 / Market long
order, err := futuresTradingService.OpenLongPosition("BTCUSDT", 0.01, MARKET, 0)

// 限价开多仓 / Limit long
order, err := futuresTradingService.OpenLongPosition("BTCUSDT", 0.01, LIMIT, 45000.0)
```

---

##### OpenShortPosition

开空仓。

Open short position.

```go
OpenShortPosition(symbol string, quantity float64, orderType OrderType, price float64) (*FuturesOrder, error)
```

**参数 / Parameters:**
- `symbol` (string): 合约交易对 / Futures trading pair
- `quantity` (float64): 数量 / Quantity
- `orderType` (OrderType): 订单类型 / Order type
- `price` (float64): 价格（限价单需要）/ Price (for limit orders)

**返回 / Returns:**
- `*FuturesOrder`: 订单详情 / Order details
- `error`: 错误信息 / Error if any

---

##### ClosePosition

平仓。

Close position.

```go
ClosePosition(symbol string, positionSide PositionSide, quantity float64) (*FuturesOrder, error)
```

**参数 / Parameters:**
- `symbol` (string): 合约交易对 / Futures trading pair
- `positionSide` (PositionSide): 持仓方向 / Position side
  - `LONG`: 多头 / Long
  - `SHORT`: 空头 / Short
- `quantity` (float64): 平仓数量 / Quantity to close

**返回 / Returns:**
- `*FuturesOrder`: 订单详情 / Order details
- `error`: 错误信息 / Error if any

---

### FuturesPositionManager

合约持仓管理器接口。

Futures position manager interface.

**包路径 / Package:** `internal/service`

#### 方法 / Methods

##### GetPosition

获取指定持仓。

Get specific position.

```go
GetPosition(symbol string, positionSide PositionSide) (*Position, error)
```

**参数 / Parameters:**
- `symbol` (string): 合约交易对 / Futures trading pair
- `positionSide` (PositionSide): 持仓方向 / Position side

**返回 / Returns:**
- `*Position`: 持仓信息 / Position information
- `error`: 错误信息 / Error if any

---

##### CalculateUnrealizedPnL

计算未实现盈亏。

Calculate unrealized PnL.

```go
CalculateUnrealizedPnL(position *Position, markPrice float64) (float64, error)
```

**参数 / Parameters:**
- `position` (*Position): 持仓信息 / Position information
- `markPrice` (float64): 标记价格 / Mark price

**返回 / Returns:**
- `float64`: 未实现盈亏 / Unrealized PnL
- `error`: 错误信息 / Error if any

---

##### CalculateLiquidationPrice

计算强平价格。

Calculate liquidation price.

```go
CalculateLiquidationPrice(position *Position) (float64, error)
```

**参数 / Parameters:**
- `position` (*Position): 持仓信息 / Position information

**返回 / Returns:**
- `float64`: 强平价格 / Liquidation price
- `error`: 错误信息 / Error if any

---

### FuturesRiskManager

合约风险管理器接口。

Futures risk manager interface.

**包路径 / Package:** `internal/service`

#### 方法 / Methods

##### CheckLiquidationRisk

检查强平风险。

Check liquidation risk.

```go
CheckLiquidationRisk(position *Position, markPrice float64) (bool, error)
```

**参数 / Parameters:**
- `position` (*Position): 持仓信息 / Position information
- `markPrice` (float64): 标记价格 / Mark price

**返回 / Returns:**
- `bool`: 是否存在强平风险 / Whether liquidation risk exists
- `error`: 错误信息 / Error if any

---

##### CheckMarginSufficiency

检查保证金充足性。

Check margin sufficiency.

```go
CheckMarginSufficiency(symbol string, quantity float64, leverage int) error
```

**参数 / Parameters:**
- `symbol` (string): 合约交易对 / Futures trading pair
- `quantity` (float64): 数量 / Quantity
- `leverage` (int): 杠杆倍数 / Leverage

**返回 / Returns:**
- `error`: 如果保证金不足则返回错误 / Error if margin insufficient

---

## 合约数据模型 / Futures Data Models

### FuturesOrderRequest

合约订单请求对象。

Futures order request object.

```go
type FuturesOrderRequest struct {
    Symbol           string           // 合约交易对 / Futures trading pair
    Side             OrderSide        // BUY 或 SELL / BUY or SELL
    PositionSide     PositionSide     // LONG, SHORT, 或 BOTH / LONG, SHORT, or BOTH
    Type             OrderType        // MARKET, LIMIT, STOP, TAKE_PROFIT
    Quantity         float64          // 交易数量 / Quantity
    Price            float64          // 价格（限价单）/ Price (for limit orders)
    StopPrice        float64          // 触发价格（止损/止盈单）/ Trigger price (for stop orders)
    TimeInForce      string           // GTC, IOC, FOK, GTX
    ReduceOnly       bool             // 只减仓 / Reduce only
    ClosePosition    bool             // 平仓标志 / Close position flag
}
```

---

### Position

持仓对象。

Position object.

```go
type Position struct {
    Symbol              string          // 合约交易对 / Futures trading pair
    PositionSide        PositionSide    // LONG 或 SHORT / LONG or SHORT
    PositionAmt         float64         // 持仓数量 / Position amount
    EntryPrice          float64         // 开仓均价 / Entry price
    MarkPrice           float64         // 标记价格 / Mark price
    UnrealizedProfit    float64         // 未实现盈亏 / Unrealized profit
    LiquidationPrice    float64         // 强平价格 / Liquidation price
    Leverage            int             // 杠杆倍数 / Leverage
    MarginType          MarginType      // ISOLATED 或 CROSSED / ISOLATED or CROSSED
    IsolatedMargin      float64         // 逐仓保证金 / Isolated margin
    PositionInitialMargin float64       // 持仓初始保证金 / Position initial margin
    MaintenanceMargin   float64         // 维持保证金 / Maintenance margin
    UpdateTime          int64           // 更新时间 / Update time
}
```

---

### FuturesBalance

合约余额对象。

Futures balance object.

```go
type FuturesBalance struct {
    Asset                  string   // "USDT"
    Balance                float64  // 总余额 / Total balance
    AvailableBalance       float64  // 可用余额 / Available balance
    CrossWalletBalance     float64  // 全仓钱包余额 / Cross wallet balance
    CrossUnPnl             float64  // 全仓未实现盈亏 / Cross unrealized PnL
    MaxWithdrawAmount      float64  // 最大可转出余额 / Max withdraw amount
    MarginAvailable        bool     // 是否可用作保证金 / Margin available
    UpdateTime             int64    // 更新时间 / Update time
}
```

---

### MarkPrice

标记价格对象。

Mark price object.

```go
type MarkPrice struct {
    Symbol          string   // 合约交易对 / Futures trading pair
    MarkPrice       float64  // 标记价格 / Mark price
    IndexPrice      float64  // 指数价格 / Index price
    LastFundingRate float64  // 最新资金费率 / Last funding rate
    NextFundingTime int64    // 下次结算时间 / Next funding time
    Time            int64    // 时间戳 / Timestamp
}
```

---

### FundingRate

资金费率对象。

Funding rate object.

```go
type FundingRate struct {
    Symbol       string   // 合约交易对 / Futures trading pair
    FundingRate  float64  // 资金费率 / Funding rate
    FundingTime  int64    // 结算时间 / Funding time
}
```

---

## 合约使用示例 / Futures Usage Examples

### 完整合约交易流程 / Complete Futures Trading Flow

```go
package main

import (
    "fmt"
    "log"
    "time"
    
    "binance-trader/internal/api"
    "binance-trader/internal/service"
)

func main() {
    // 1. 初始化合约客户端 / Initialize futures client
    authMgr, _ := api.NewAuthManager(futuresAPIKey, futuresAPISecret)
    rateLimiter := api.NewRateLimiter(2000)
    httpClient := api.NewHTTPClient(rateLimiter, retryConfig)
    futuresClient, _ := api.NewFuturesClient(futuresBaseURL, httpClient, authMgr)
    
    // 2. 初始化服务 / Initialize services
    futuresPositionMgr := service.NewFuturesPositionManager(futuresClient, positionRepo)
    futuresRiskMgr := service.NewFuturesRiskManager(riskLimits, futuresClient)
    futuresTradingService := service.NewFuturesTradingService(
        futuresClient, 
        futuresRiskMgr, 
        futuresOrderRepo, 
        futuresPositionMgr,
        log,
    )
    
    symbol := "BTCUSDT"
    
    // 3. 设置杠杆 / Set leverage
    err := futuresClient.SetLeverage(symbol, 10)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("Leverage set to 10x")
    
    // 4. 获取标记价格 / Get mark price
    markPrice, err := futuresClient.GetMarkPrice(symbol)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Mark Price: %.2f\n", markPrice.MarkPrice)
    fmt.Printf("Funding Rate: %.4f%%\n", markPrice.LastFundingRate*100)
    
    // 5. 开多仓 / Open long position
    longOrder, err := futuresTradingService.OpenLongPosition(symbol, 0.01, MARKET, 0)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Long position opened: ID=%d, Entry=%.2f\n", 
        longOrder.OrderID, longOrder.AvgPrice)
    
    // 6. 查看持仓 / View position
    position, err := futuresPositionMgr.GetPosition(symbol, LONG)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Position: %.8f @ %.2f\n", position.PositionAmt, position.EntryPrice)
    fmt.Printf("Unrealized PnL: %.2f USDT\n", position.UnrealizedProfit)
    fmt.Printf("Liquidation Price: %.2f\n", position.LiquidationPrice)
    
    // 7. 设置止损 / Set stop loss
    stopLossPrice := position.EntryPrice * 0.98 // 2% below entry
    stopOrder, err := futuresStopLossService.SetStopLoss(
        symbol, 
        LONG, 
        position.PositionAmt, 
        stopLossPrice,
    )
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Stop loss set at %.2f\n", stopLossPrice)
    
    // 8. 等待一段时间 / Wait for some time
    time.Sleep(10 * time.Second)
    
    // 9. 平仓 / Close position
    closeOrder, err := futuresTradingService.ClosePosition(symbol, LONG, position.PositionAmt)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Position closed: Exit=%.2f, PnL=%.2f\n", 
        closeOrder.AvgPrice, closeOrder.RealizedProfit)
}
```

---

**更多合约交易示例请参阅 / For more futures trading examples, see:** [FUTURES_QUICKSTART.md](FUTURES_QUICKSTART.md)
