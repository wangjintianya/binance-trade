# 合约CLI命令参考 / Futures CLI Command Reference

快速参考指南，列出所有可用的合约交易CLI命令。

Quick reference guide for all available futures trading CLI commands.

---

## 🚀 启动 / Starting

```bash
# 设置环境变量 / Set environment variables
export BINANCE_FUTURES_API_KEY="your_api_key"
export BINANCE_FUTURES_API_SECRET="your_api_secret"

# 启动合约交易系统 / Start futures trading system
./binance-trader.exe futures
```

---

## 📊 市场数据命令 / Market Data Commands

### mark-price - 查询标记价格
```bash
> mark-price BTCUSDT
Symbol:      BTCUSDT
Mark Price:  50000.12345678
```

### funding-rate - 查询资金费率
```bash
> funding-rate BTCUSDT
Symbol:        BTCUSDT
Funding Rate:  0.010000%
```

### position - 查看指定合约持仓
```bash
> position BTCUSDT
Positions for BTCUSDT
Symbol:           BTCUSDT
Position Side:    LONG
Position Amount:  0.00100000
Entry Price:      50000.00000000
Mark Price:       50100.00000000
Unrealized PnL:   0.10000000
Liquidation:      45000.00000000
Leverage:         10x
Margin Type:      CROSSED
```

### positions - 查看所有持仓
```bash
> positions
All Positions (2)
[1] BTCUSDT LONG: 0.001 @ 50000
[2] ETHUSDT SHORT: 0.01 @ 2000
```

---

## 💰 交易命令 / Trading Commands

### long - 开多仓（市价）
```bash
> long BTCUSDT 0.001
Long Position Opened
Order ID:    12345
Symbol:      BTCUSDT
Side:        BUY
Quantity:    0.00100000
Status:      FILLED
```

### short - 开空仓（市价）
```bash
> short BTCUSDT 0.001
Short Position Opened
Order ID:    12346
Symbol:      BTCUSDT
Side:        SELL
Quantity:    0.00100000
Status:      FILLED
```

### close - 平仓
```bash
> close BTCUSDT
Closed 1 position(s) for BTCUSDT
```

---

## ⚖️ 杠杆和保证金命令 / Leverage & Margin Commands

### leverage - 设置杠杆
```bash
> leverage BTCUSDT 10
Leverage set to 10x for BTCUSDT
```

**注意：** 杠杆范围 1-125x / Note: Leverage range 1-125x

### margin-type - 设置保证金模式
```bash
> margin-type BTCUSDT CROSSED
Margin type set to CROSSED for BTCUSDT

> margin-type BTCUSDT ISOLATED
Margin type set to ISOLATED for BTCUSDT
```

**选项 / Options:**
- `CROSSED` / `CROSS` - 全仓模式
- `ISOLATED` - 逐仓模式

---

## 🎯 条件订单命令 / Conditional Order Commands

### condorder - 创建条件订单
```bash
# 语法 / Syntax
condorder <symbol> <side> <position_side> <quantity> <trigger_type> <operator> <value>

# 示例 1: 标记价格突破开多 / Mark price breakout long
> condorder BTCUSDT BUY LONG 0.001 MARK_PRICE >= 51000
Conditional Order Created
Order ID:    cond-001
Symbol:      BTCUSDT
Side:        BUY
Position:    LONG
Quantity:    0.00100000
Trigger:     MARK_PRICE >= 51000.00000000

# 示例 2: 盈亏止盈 / PnL take profit
> condorder BTCUSDT SELL LONG 0.001 PNL >= 1000
Conditional Order Created - Will close when PnL >= 1000 USDT

# 示例 3: 资金费率触发 / Funding rate trigger
> condorder BTCUSDT SELL SHORT 0.001 FUNDING_RATE >= 0.0001
Conditional Order Created - Will open short when funding rate >= 0.01%
```

**参数说明 / Parameters:**
- `<symbol>` - 交易对，如 BTCUSDT
- `<side>` - BUY 或 SELL
- `<position_side>` - LONG, SHORT, 或 BOTH
- `<quantity>` - 数量
- `<trigger_type>` - 触发类型：
  - `MARK_PRICE` - 标记价格
  - `LAST_PRICE` - 最新价格
  - `PNL` / `UNREALIZED_PNL` - 未实现盈亏
  - `FUNDING_RATE` - 资金费率
- `<operator>` - 操作符：`>=`, `<=`, `>`, `<`
- `<value>` - 触发值

### condorders - 查看活跃条件订单
```bash
> condorders
Active Conditional Orders (2)
[1] Order ID: cond-001
    Symbol:      BTCUSDT
    Side:        BUY
    Position:    LONG
    Quantity:    0.00100000
    Trigger:     MARK_PRICE >= 51000.00000000
    Status:      PENDING

[2] Order ID: cond-002
    Symbol:      ETHUSDT
    Side:        SELL
    Position:    LONG
    Quantity:    0.01000000
    Trigger:     PNL >= 500.00000000
    Status:      PENDING
```

### cancelcond - 取消条件订单
```bash
> cancelcond cond-001
Conditional order cond-001 cancelled successfully
```

---

## 🛑 止损止盈命令 / Stop Loss/Take Profit Commands

### stoploss - 设置止损
```bash
# 语法 / Syntax
stoploss <symbol> <side> <quantity> <price>

# 示例: 为多头设置止损 / Set stop loss for long position
> stoploss BTCUSDT LONG 0.001 49000
Stop Loss Set
Order ID:    stop-001
Symbol:      BTCUSDT
Side:        LONG
Quantity:    0.00100000
Stop Price:  49000.00000000

# 示例: 为空头设置止损 / Set stop loss for short position
> stoploss BTCUSDT SHORT 0.001 51000
Stop Loss Set (will close short if price rises to 51000)
```

### takeprofit - 设置止盈
```bash
# 语法 / Syntax
takeprofit <symbol> <side> <quantity> <price>

# 示例: 为多头设置止盈 / Set take profit for long position
> takeprofit BTCUSDT LONG 0.001 52000
Take Profit Set
Order ID:      tp-001
Symbol:        BTCUSDT
Side:          LONG
Quantity:      0.00100000
Target Price:  52000.00000000
```

### stoporders - 查看止损止盈订单
```bash
> stoporders BTCUSDT
Active Stop Orders for BTCUSDT (2)
[1] Order ID: stop-001
    Symbol:      BTCUSDT
    Type:        STOP_LOSS
    Position:    0.00100000
    Stop Price:  49000.00000000
    Status:      ACTIVE

[2] Order ID: tp-001
    Symbol:      BTCUSDT
    Type:        TAKE_PROFIT
    Position:    0.00100000
    Stop Price:  52000.00000000
    Status:      ACTIVE
```

### cancelstop - 取消止损止盈订单
```bash
> cancelstop stop-001
Stop order stop-001 cancelled successfully
```

---

## 🔧 系统命令 / System Commands

### help - 显示帮助
```bash
> help
Available Commands:
  [显示所有可用命令列表]
```

### exit / quit - 退出程序
```bash
> exit
Goodbye!
```

---

## 💡 使用技巧 / Usage Tips

### 1. 完整交易流程示例 / Complete Trading Flow Example

```bash
# 1. 查询当前价格
> mark-price BTCUSDT
Mark Price: 50000.00

# 2. 设置杠杆
> leverage BTCUSDT 10
Leverage set to 10x

# 3. 开多仓
> long BTCUSDT 0.001
Long Position Opened

# 4. 设置止损和止盈
> stoploss BTCUSDT LONG 0.001 49000
Stop Loss Set

> takeprofit BTCUSDT LONG 0.001 52000
Take Profit Set

# 5. 查看持仓
> position BTCUSDT
Position: LONG 0.001 @ 50000
Unrealized PnL: +10.00 USDT

# 6. 平仓（可选）
> close BTCUSDT
Position closed
```

### 2. 条件订单策略示例 / Conditional Order Strategy Example

```bash
# 突破策略：价格突破51000时开多
> condorder BTCUSDT BUY LONG 0.001 MARK_PRICE >= 51000

# 回调策略：价格回调到49000时加仓
> condorder BTCUSDT BUY LONG 0.001 MARK_PRICE <= 49000

# 盈亏管理：盈利达到1000 USDT时平仓
> condorder BTCUSDT SELL LONG 0.001 PNL >= 1000

# 查看所有条件订单
> condorders
```

### 3. 资金费率套利示例 / Funding Rate Arbitrage Example

```bash
# 当资金费率过高时开空仓收取资金费
> condorder BTCUSDT SELL SHORT 0.001 FUNDING_RATE >= 0.0001

# 查询当前资金费率
> funding-rate BTCUSDT
Funding Rate: 0.0150%  (高资金费率)
```

---

## ⚠️ 重要提示 / Important Notes

1. **杠杆风险** / **Leverage Risk**
   - 高杠杆意味着高风险
   - 建议新手使用低杠杆（1-5x）
   - 始终设置止损

2. **保证金模式** / **Margin Mode**
   - 全仓模式：使用账户全部余额作为保证金
   - 逐仓模式：只使用分配给该仓位的保证金
   - 有持仓时无法切换模式

3. **条件订单监控** / **Conditional Order Monitoring**
   - 条件订单需要系统持续运行
   - 系统每秒检查一次触发条件
   - 条件满足时立即执行市价单

4. **止损止盈** / **Stop Loss/Take Profit**
   - 建议每个持仓都设置止损
   - 可以同时设置止损和止盈
   - 任一触发后，另一个自动取消

---

## 📚 相关文档 / Related Documentation

- [README.md](../README.md) - 项目总览
- [API.md](API.md) - API文档
- [EXAMPLES.md](EXAMPLES.md) - 使用示例
- [COMMAND_GUIDE.md](COMMAND_GUIDE.md) - 命令详细指南
- [CLI_STATUS.md](CLI_STATUS.md) - CLI实现状态
- [FUTURES_QUICKSTART.md](FUTURES_QUICKSTART.md) - 合约快速入门

---

**最后更新 / Last Updated:** 2024-12-09

**版本 / Version:** 1.0.0 - 合约CLI完整实现
