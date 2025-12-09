# 命令使用指南 / Command Usage Guide

本文档详细说明各种交易命令的工作原理和使用场景。

This document explains how different trading commands work and their use cases.

---

## 目录 / Table of Contents

1. [命令分类 / Command Categories](#命令分类--command-categories)
2. [普通订单 vs 条件订单 / Regular Orders vs Conditional Orders](#普通订单-vs-条件订单--regular-orders-vs-conditional-orders)
3. [详细命令说明 / Detailed Command Explanations](#详细命令说明--detailed-command-explanations)
4. [使用场景对比 / Use Case Comparison](#使用场景对比--use-case-comparison)

---

## 命令分类 / Command Categories

### 1. 立即执行订单 (Immediate Execution Orders)

这些命令会**立即**向币安交易所发送订单请求：

These commands **immediately** send order requests to Binance exchange:

| 命令 | 说明 | 执行时机 |
|------|------|----------|
| `buy <symbol> <quantity>` | 市价买入 | 立即执行 |
| `sell <symbol> <price> <quantity>` | 限价卖出 | 立即提交到交易所 |

### 2. 条件触发订单 (Conditional Trigger Orders)

这些命令会在系统中**创建监控任务**，等待条件满足后才执行：

These commands **create monitoring tasks** in the system and execute only when conditions are met:

| 命令 | 说明 | 执行时机 |
|------|------|----------|
| `condorder <symbol> <side> <qty> <trigger_type> <operator> <value>` | 条件订单 | 条件满足时执行 |
| `stoploss <symbol> <position> <stop_price>` | 止损单 | 价格触及止损价时执行 |
| `takeprofit <symbol> <position> <target_price>` | 止盈单 | 价格触及目标价时执行 |

---

## 普通订单 vs 条件订单 / Regular Orders vs Conditional Orders

### 普通限价单 (Regular Limit Order)

```bash
# 命令：立即下限价卖单
> sell BTCUSDT 50000 0.001
```

**工作流程：**
```
用户输入命令
    ↓
立即调用 tradingService.PlaceLimitSellOrder()
    ↓
直接发送到币安交易所
    ↓
订单挂在交易所订单簿上
    ↓
等待市场价格达到 50000 时成交（由交易所处理）
```

**特点：**
- ✅ 订单立即提交到交易所
- ✅ 订单显示在交易所的订单簿中
- ✅ 由交易所撮合引擎处理
- ❌ 无法设置复杂触发条件（只能设置价格）
- ❌ 占用交易所订单数量限制

### 条件订单 (Conditional Order)

```bash
# 命令：创建价格触发的买单
> condorder BTCUSDT BUY 0.001 PRICE >= 50000
```

**工作流程：**
```
用户输入命令
    ↓
创建条件订单记录（保存在本地）
    ↓
启动监控任务（每秒检查一次）
    ↓
持续监控市场价格
    ↓
当价格 >= 50000 时触发
    ↓
自动调用 tradingService.PlaceMarketBuyOrder()
    ↓
发送市价单到交易所
```

**特点：**
- ✅ 支持复杂触发条件（价格、成交量、涨跌幅等）
- ✅ 不占用交易所订单数量
- ✅ 可以随时修改或取消（在触发前）
- ✅ 触发后执行市价单，保证成交
- ❌ 需要系统持续运行监控
- ❌ 订单不在交易所订单簿中

---

## 详细命令说明 / Detailed Command Explanations

### 1. buy - 市价买入 (Market Buy)

```bash
> buy BTCUSDT 0.001
```

**代码路径：**
```
CLI.handleBuy()
  → TradingService.PlaceMarketBuyOrder()
    → BinanceClient.CreateOrder()
      → 发送到币安 API
```

**执行逻辑：**
1. 立即以当前市场价格买入
2. 订单类型：MARKET
3. 保证成交（除非余额不足）

**适用场景：**
- 需要立即建仓
- 不在乎短期价格波动
- 追涨买入

---

### 2. sell - 限价卖出 (Limit Sell)

```bash
> sell BTCUSDT 50000 0.001
```

**代码路径：**
```
CLI.handleSell()
  → TradingService.PlaceLimitSellOrder()
    → BinanceClient.CreateOrder()
      → 发送到币安 API
```

**执行逻辑：**
1. 立即提交限价卖单到交易所
2. 订单类型：LIMIT
3. 挂在订单簿上，等待买家
4. 只有当市场价格 >= 50000 时才会成交

**适用场景：**
- 想以特定价格卖出
- 不急于成交
- 设置止盈价位

**注意：**
- 订单会一直挂在交易所，直到成交或取消
- 如果价格一直不到 50000，订单永远不会成交

---

### 3. condorder - 条件订单 (Conditional Order)

```bash
# 示例1：价格触发
> condorder BTCUSDT BUY 0.001 PRICE >= 50000

# 示例2：涨幅触发
> condorder ETHUSDT BUY 0.01 PRICE_CHANGE >= 5.0

# 示例3：成交量触发
> condorder BNBUSDT SELL 1.0 VOLUME >= 1000000
```

**代码路径：**
```
CLI.handleConditionalOrder()
  → ConditionalOrderService.CreateConditionalOrder()
    → 保存到 ConditionalOrderRepository
    → TriggerEngine.RegisterCondition()
      → 启动监控循环（每秒检查）
        → 当条件满足时：
          → TradingService.PlaceMarketBuyOrder()
            → 发送到币安 API
```

**执行逻辑：**
1. **创建阶段**：
   - 解析触发条件
   - 保存订单信息到本地数据库
   - 注册到触发引擎
   - 状态：PENDING

2. **监控阶段**：
   - 每秒（可配置）检查一次市场数据
   - 评估触发条件是否满足
   - 记录监控日志

3. **触发阶段**：
   - 条件满足时，立即执行市价单
   - 更新订单状态：PENDING → TRIGGERED → EXECUTED
   - 记录执行日志

**支持的触发类型：**

| 触发类型 | 说明 | 示例 |
|----------|------|------|
| `PRICE` | 价格达到指定值 | `PRICE >= 50000` |
| `PRICE_CHANGE` | 价格涨跌幅 | `PRICE_CHANGE >= 5.0` (涨5%) |
| `VOLUME` | 成交量达到指定值 | `VOLUME >= 1000000` |

**支持的操作符：**
- `>=` (GE) - 大于等于
- `<=` (LE) - 小于等于
- `>` (GT) - 大于
- `<` (LT) - 小于

**适用场景：**
- 突破买入：价格突破关键阻力位时买入
- 回调买入：价格回调到支撑位时买入
- 放量买入：成交量放大时买入
- 涨幅追踪：涨幅达到一定比例时买入

---

### 4. stoploss - 止损单 (Stop Loss)

```bash
> stoploss BTCUSDT 0.001 49000
```

**代码路径：**
```
CLI.handleStopLoss()
  → StopLossService.SetStopLoss()
    → 保存到 StopOrderRepository
    → TriggerEngine.RegisterCondition()
      → 监控价格
        → 当价格 <= 49000 时：
          → TradingService.PlaceMarketSellOrder()
```

**执行逻辑：**
1. 创建止损监控任务
2. 持续监控当前价格
3. 当价格跌破止损价时，立即市价卖出
4. 目的：限制损失

**适用场景：**
- 保护利润：买入后设置止损，防止回撤
- 限制损失：设置最大可接受亏损
- 自动风控：无需人工盯盘

**示例：**
```bash
# 场景：以 50000 买入 BTC，设置 2% 止损
> buy BTCUSDT 0.001
Order placed at $50000

> stoploss BTCUSDT 0.001 49000
Stop loss set at $49000 (2% below entry)

# 如果价格跌到 49000，自动卖出
# 最大损失：$1000
```

---

### 5. takeprofit - 止盈单 (Take Profit)

```bash
> takeprofit BTCUSDT 0.001 51000
```

**代码路径：**
```
CLI.handleTakeProfit()
  → StopLossService.SetTakeProfit()
    → 保存到 StopOrderRepository
    → TriggerEngine.RegisterCondition()
      → 监控价格
        → 当价格 >= 51000 时：
          → TradingService.PlaceMarketSellOrder()
```

**执行逻辑：**
1. 创建止盈监控任务
2. 持续监控当前价格
3. 当价格达到目标价时，立即市价卖出
4. 目的：锁定利润

**适用场景：**
- 自动获利：达到目标价位自动卖出
- 避免贪婪：防止利润回吐
- 无人值守：自动执行交易计划

---

## 使用场景对比 / Use Case Comparison

### 场景 1：我想在 BTC 价格到 50000 时卖出

#### 方案 A：使用限价单 (Limit Order)
```bash
> sell BTCUSDT 50000 0.001
```

**优点：**
- 订单立即提交到交易所
- 可能以更好的价格成交（如果有人出价更高）
- 不需要系统持续运行

**缺点：**
- 如果价格快速突破 50000 后回落，可能不成交
- 占用交易所订单数量限制
- 只能设置固定价格

**适合：**
- 不急于成交
- 希望以特定价格或更好价格卖出
- 系统可能会关闭

#### 方案 B：使用条件订单 (Conditional Order)
```bash
> condorder BTCUSDT SELL 0.001 PRICE >= 50000
```

**优点：**
- 价格达到 50000 时立即市价卖出，保证成交
- 可以设置复杂条件（如涨幅、成交量等）
- 不占用交易所订单数量

**缺点：**
- 需要系统持续运行
- 以市价成交，可能有滑点
- 订单不在交易所订单簿中

**适合：**
- 需要保证成交
- 需要复杂触发条件
- 系统会持续运行

---

### 场景 2：我想在 BTC 跌破 49000 时止损

#### 唯一方案：使用止损单 (Stop Loss)
```bash
> stoploss BTCUSDT 0.001 49000
```

**为什么不能用限价单？**
- 限价单是"卖出价格 >= 49000"
- 止损需要"卖出价格 <= 49000"
- 限价单无法实现止损逻辑

**止损单的工作原理：**
1. 监控价格
2. 当价格 <= 49000 时触发
3. 立即市价卖出
4. 保证成交，限制损失

---

### 场景 3：我想在 BTC 突破 52000 后追涨买入

#### 方案 A：使用限价单 (不推荐)
```bash
> buy BTCUSDT 0.001
# 问题：只能市价买入，无法等待突破
```

#### 方案 B：使用条件订单 (推荐)
```bash
> condorder BTCUSDT BUY 0.001 PRICE >= 52000
```

**优势：**
- 等待价格突破 52000
- 突破后立即买入
- 避免假突破（可以设置更复杂的条件）

---

## 命令对应的代码模块 / Code Modules for Each Command

### 立即执行订单 (Immediate Orders)

```
buy/sell 命令
    ↓
internal/cli/cli.go (handleBuy/handleSell)
    ↓
internal/service/trading.go (PlaceMarketBuyOrder/PlaceLimitSellOrder)
    ↓
internal/api/client.go (CreateOrder)
    ↓
币安 API
```

### 条件订单 (Conditional Orders)

```
condorder 命令
    ↓
internal/cli/cli.go (handleConditionalOrder)
    ↓
internal/service/conditional_order.go (CreateConditionalOrder)
    ↓
internal/repository/conditional_order.go (Save)
    ↓
internal/service/trigger.go (RegisterCondition)
    ↓
[监控循环]
    ↓
条件满足时 → internal/service/trading.go (PlaceMarketBuyOrder)
    ↓
币安 API
```

### 止损止盈订单 (Stop Loss/Take Profit)

```
stoploss/takeprofit 命令
    ↓
internal/cli/cli.go (handleStopLoss/handleTakeProfit)
    ↓
internal/service/stop_loss.go (SetStopLoss/SetTakeProfit)
    ↓
internal/repository/stop_order.go (Save)
    ↓
internal/service/trigger.go (RegisterCondition)
    ↓
[监控循环]
    ↓
条件满足时 → internal/service/trading.go (PlaceMarketSellOrder)
    ↓
币安 API
```

---

## 配置参数 / Configuration Parameters

### 条件订单监控间隔

```yaml
conditional_orders:
  monitoring_interval_ms: 1000  # 每秒检查一次
  max_active_orders: 100        # 最多100个活跃条件订单
  trigger_execution_timeout_ms: 3000  # 触发后3秒内必须执行
```

### 止损止盈监控间隔

```yaml
stop_loss:
  update_interval_ms: 1000      # 每秒检查一次
  default_trail_percent: 1.0    # 默认移动止损幅度 1%
```

---

## 总结 / Summary

### 何时使用限价单 (When to Use Limit Orders)
- ✅ 不急于成交
- ✅ 想以特定价格或更好价格交易
- ✅ 系统可能会关闭
- ✅ 只需要简单的价格条件

### 何时使用条件订单 (When to Use Conditional Orders)
- ✅ 需要保证成交
- ✅ 需要复杂触发条件（涨跌幅、成交量等）
- ✅ 系统会持续运行
- ✅ 需要灵活的策略

### 何时使用止损止盈 (When to Use Stop Loss/Take Profit)
- ✅ 需要自动风险管理
- ✅ 无法实时盯盘
- ✅ 需要严格执行交易计划
- ✅ 保护利润或限制损失

---

## 常见问题 / FAQ

### Q1: 限价单和条件订单有什么区别？

**A:** 
- **限价单**：立即提交到交易所，挂在订单簿上，由交易所撮合
- **条件订单**：保存在本地，由系统监控，条件满足时才发送到交易所

### Q2: 条件订单会占用交易所的订单数量限制吗？

**A:** 不会。条件订单在触发前不会提交到交易所，因此不占用订单数量限制。

### Q3: 如果系统关闭，条件订单还会执行吗？

**A:** 不会。条件订单需要系统持续运行来监控市场。如果系统关闭，监控会停止。

### Q4: 条件订单触发后是市价单还是限价单？

**A:** 默认是市价单，保证立即成交。这样可以确保在条件满足时不会错过机会。

### Q5: 可以同时设置止损和止盈吗？

**A:** 可以。系统支持同时设置止损和止盈，任一条件触发后，另一个会自动取消。

### Q6: 条件订单的监控频率是多少？

**A:** 默认每秒检查一次，可以在配置文件中调整 `monitoring_interval_ms` 参数。

---

## 实战示例 / Practical Examples

### 示例 1：完整的交易流程

```bash
# 1. 查询当前价格
> price BTCUSDT
Current price: $48000

# 2. 市价买入
> buy BTCUSDT 0.001
Order placed: ID=12345, Price=$48000, Status=FILLED

# 3. 设置止损（2%）
> stoploss BTCUSDT 0.001 47040
Stop loss set at $47040

# 4. 设置止盈（5%）
> takeprofit BTCUSDT 0.001 50400
Take profit set at $50400

# 5. 查看活跃的止损止盈订单
> stoporders BTCUSDT
Active Stop Orders (2)
[1] Stop Loss at $47040
[2] Take Profit at $50400
```

### 示例 2：突破策略

```bash
# 1. 查询当前价格
> price BTCUSDT
Current price: $49500

# 2. 设置突破买入（突破50000）
> condorder BTCUSDT BUY 0.001 PRICE >= 50000
Conditional order created: ID=cond-001

# 3. 查看条件订单
> condorders
Active Conditional Orders (1)
[1] ID=cond-001, Symbol=BTCUSDT, Trigger=PRICE >= 50000

# 当价格突破50000时，系统自动执行：
# → Market buy order placed at $50000
```

### 示例 3：网格交易

```bash
# 设置多个限价卖单（网格）
> sell BTCUSDT 50000 0.001
> sell BTCUSDT 51000 0.001
> sell BTCUSDT 52000 0.001
> sell BTCUSDT 53000 0.001

# 查看活跃订单
> orders
Active Orders (4)
[1] Sell at $50000
[2] Sell at $51000
[3] Sell at $52000
[4] Sell at $53000
```

---

**提示：** 建议先在测试网环境熟悉各种命令的使用，再在实盘环境操作。

**Tip:** It's recommended to practice with testnet environment before trading on mainnet.
