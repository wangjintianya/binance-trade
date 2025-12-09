# Binance Auto-Trading System

币安自动交易系统 - 使用Go语言开发的自动化加密货币交易应用程序，支持现货和U本位合约交易

An automated cryptocurrency trading application that integrates with the Binance exchange API to execute spot and USDT-M futures trades programmatically with built-in risk management and comprehensive logging.

## 📋 目录 / Table of Contents

- [功能特性 / Features](#功能特性--features)
- [前置要求 / Prerequisites](#前置要求--prerequisites)
- [快速开始 / Quick Start](#快速开始--quick-start)
- [配置 / Configuration](#配置--configuration)
- [使用方法 / Usage](#使用方法--usage)
- [API文档 / API Documentation](#api文档--api-documentation)
- [测试 / Testing](#测试--testing)
- [项目结构 / Project Structure](#项目结构--project-structure)
- [安全性 / Security](#安全性--security)
- [风险管理 / Risk Management](#风险管理--risk-management)
- [故障排除 / Troubleshooting](#故障排除--troubleshooting)
- [贡献 / Contributing](#贡献--contributing)
- [许可证 / License](#许可证--license)

## 功能特性 / Features

### 核心功能 / Core Features

#### 现货交易 / Spot Trading
- ✅ **安全的API集成** / **Secure API Integration** - HMAC SHA256 authentication with automatic request signing
- 📊 **实时市场数据** / **Real-time Market Data** - Prices, K-lines, and account balances
- 🤖 **自动化订单管理** / **Automated Order Management** - Market orders and limit orders
- 🎯 **条件订单** / **Conditional Orders** - Trigger orders based on price, volume, or percentage changes
- 🛑 **止损止盈** / **Stop Loss & Take Profit** - Automatic position protection with stop-loss and take-profit orders
- 📈 **移动止损** / **Trailing Stop** - Dynamic stop-loss that adjusts with favorable price movements
- 🔀 **复合触发条件** / **Composite Triggers** - Combine multiple conditions with AND/OR logic
- 🛡️ **风险控制机制** / **Risk Control** - Order limits, balance protection, and rate limiting

#### U本位合约交易 / USDT-M Futures Trading
- 🚀 **合约交易支持** / **Futures Trading Support** - Full support for USDT-margined perpetual and delivery contracts
- 📊 **合约市场数据** / **Futures Market Data** - Mark price, funding rate, and position information
- ⚖️ **杠杆管理** / **Leverage Management** - Adjustable leverage from 1x to 125x
- 💰 **保证金模式** / **Margin Modes** - Support for cross margin and isolated margin
- 📍 **持仓管理** / **Position Management** - Real-time position tracking with PnL calculation
- 🔄 **双向持仓** / **Hedge Mode** - Support for both one-way and hedge position modes
- 💸 **资金费率处理** / **Funding Rate Management** - Automatic funding fee tracking and settlement
- 🛡️ **合约风险控制** / **Futures Risk Control** - Liquidation monitoring, margin ratio alerts, and position limits
- 🎯 **合约条件订单** / **Futures Conditional Orders** - Trigger based on mark price, PnL, or funding rate
- 🛑 **合约止损止盈** / **Futures Stop Loss/Take Profit** - Advanced stop orders for futures positions

#### 通用功能 / General Features
- 📝 **完整的日志记录** / **Comprehensive Logging** - Structured logging with sensitive data masking
- 💻 **双入口点系统** / **Dual Entry Points** - Separate commands for spot and futures trading
- 🔄 **自动重试机制** / **Automatic Retry** - Exponential backoff for failed requests
- ⚡ **速率限制管理** / **Rate Limit Management** - Automatic API rate limiting to prevent throttling

### 技术特性 / Technical Features
- 模块化三层架构 / Modular three-layer architecture
- 接口驱动设计便于测试 / Interface-driven design for easy testing
- 属性测试确保正确性 / Property-based testing for correctness
- 优雅关闭机制 / Graceful shutdown mechanism
- 环境变量配置 / Environment variable configuration
- 结构化日志与日志轮转 / Structured logging with rotation

## 前置要求 / Prerequisites

- **Go 1.21+** - [下载安装 / Download](https://golang.org/dl/)
- **币安账户** / **Binance Account** - [注册 / Sign up](https://www.binance.com/)
- **API密钥** / **API Keys** - [创建API密钥 / Create API Keys](https://www.binance.com/en/my/settings/api-management)
  - 需要启用"现货和杠杆交易"权限 / Enable "Spot & Margin Trading" permission
  - 建议使用IP白名单提高安全性 / Recommended to use IP whitelist for security

## 快速开始 / Quick Start

### 1. 克隆仓库 / Clone Repository

```bash
git clone <repository-url>
cd binance-trader
```

### 2. 安装依赖 / Install Dependencies

```bash
go mod download
```

### 3. 配置环境变量 / Configure Environment Variables

**Linux/macOS:**
```bash
export BINANCE_API_KEY="your_api_key_here"
export BINANCE_API_SECRET="your_api_secret_here"
```

**Windows (PowerShell):**
```powershell
$env:BINANCE_API_KEY="your_api_key_here"
$env:BINANCE_API_SECRET="your_api_secret_here"
```

**Windows (CMD):**
```cmd
set BINANCE_API_KEY=your_api_key_here
set BINANCE_API_SECRET=your_api_secret_here
```

### 4. 构建应用程序 / Build Application

```bash
go build -o binance-trader.exe cmd/main.go
```

### 5. 运行 / Run

**现货交易 / Spot Trading:**
```bash
./binance-trader.exe spot
```

**合约交易 / Futures Trading:**
```bash
# 设置合约API密钥（可与现货相同）/ Set futures API keys (can be same as spot)
export BINANCE_FUTURES_API_KEY="your_futures_api_key_here"
export BINANCE_FUTURES_API_SECRET="your_futures_api_secret_here"

./binance-trader.exe futures
```

**同时运行现货和合约 / Run Both Spot and Futures:**
```bash
# 在不同终端窗口 / In different terminal windows
./binance-trader.exe spot &
./binance-trader.exe futures &
```

## 配置 / Configuration

### 配置文件 / Configuration File

创建或编辑 `config.yaml` 文件（参考 `config.example.yaml`）：

Create or edit `config.yaml` file (see `config.example.yaml` for reference):

```yaml
# 现货交易配置 / Spot Trading Configuration
spot:
  api_key: ${BINANCE_API_KEY}        # 从环境变量读取 / Read from environment
  api_secret: ${BINANCE_API_SECRET}  # 从环境变量读取 / Read from environment
  base_url: https://api.binance.com  # 生产环境 / Production
  testnet: false                     # 设为true使用测试网 / Set to true for testnet

  risk:
    max_order_amount: 10000.0          # 单笔最大金额(USDT) / Max order amount (USDT)
    max_daily_orders: 100              # 每日最大订单数 / Max daily orders
    min_balance_reserve: 100.0         # 最小保留余额(USDT) / Min balance reserve (USDT)
    max_api_calls_per_min: 1000        # 每分钟最大API调用 / Max API calls per minute

# 合约交易配置 / Futures Trading Configuration
futures:
  api_key: ${BINANCE_FUTURES_API_KEY}        # 合约API密钥 / Futures API key
  api_secret: ${BINANCE_FUTURES_API_SECRET}  # 合约API密钥 / Futures API secret
  base_url: https://fapi.binance.com         # 合约API端点 / Futures API endpoint
  testnet: false                             # 设为true使用测试网 / Set to true for testnet
  
  default_leverage: 10                       # 默认杠杆倍数 / Default leverage
  default_margin_type: CROSSED               # 默认保证金模式: CROSSED/ISOLATED
  dual_side_position: false                  # 双向持仓模式 / Hedge mode
  
  risk:
    max_order_value: 50000.0                 # 单笔最大订单价值 / Max order value
    max_position_value: 100000.0             # 最大持仓价值 / Max position value
    max_leverage: 20                         # 最大杠杆倍数 / Max leverage
    min_margin_ratio: 0.05                   # 最小保证金率 / Min margin ratio
    liquidation_buffer: 0.02                 # 强平缓冲区 / Liquidation buffer
    max_daily_orders: 200                    # 每日最大订单数 / Max daily orders
    max_api_calls_per_min: 2000              # 每分钟最大API调用 / Max API calls per minute
  
  monitoring:
    position_update_interval_ms: 5000        # 持仓更新间隔 / Position update interval
    conditional_order_interval_ms: 1000      # 条件订单检查间隔 / Conditional order check interval
    funding_rate_check_interval_ms: 60000    # 资金费率检查间隔 / Funding rate check interval

# 共享配置 / Shared Configuration
logging:
  level: info                        # 日志级别: debug, info, warn, error / Log level
  spot_file: logs/spot_trading.log   # 现货日志文件 / Spot log file
  futures_file: logs/futures_trading.log  # 合约日志文件 / Futures log file
  max_size_mb: 100                   # 单个日志文件最大大小 / Max log file size
  max_backups: 5                     # 保留的日志文件数 / Number of log files to keep

retry:
  max_attempts: 3                    # 最大重试次数 / Max retry attempts
  initial_delay_ms: 1000             # 初始延迟(毫秒) / Initial delay (ms)
  backoff_multiplier: 2.0            # 退避倍数 / Backoff multiplier
```

### 环境变量 / Environment Variables

| 变量名 / Variable | 必需 / Required | 说明 / Description |
|------------------|----------------|-------------------|
| `BINANCE_API_KEY` | ✅ Yes (Spot) | 币安现货API密钥 / Binance spot API key |
| `BINANCE_API_SECRET` | ✅ Yes (Spot) | 币安现货API密钥 / Binance spot API secret |
| `BINANCE_FUTURES_API_KEY` | ✅ Yes (Futures) | 币安合约API密钥 / Binance futures API key |
| `BINANCE_FUTURES_API_SECRET` | ✅ Yes (Futures) | 币安合约API密钥 / Binance futures API secret |
| `CONFIG_FILE` | ❌ No | 配置文件路径 / Config file path (default: `config.yaml`) |
| `LOG_LEVEL` | ❌ No | 日志级别 / Log level (default: `info`) |

**注意 / Note:** 现货和合约可以使用相同的API密钥，但需要确保API密钥有相应的权限。/ Spot and futures can use the same API keys, but ensure the keys have appropriate permissions.

### 测试网配置 / Testnet Configuration

建议先在测试网测试 / It's recommended to test on testnet first:

1. 获取测试网API密钥 / Get testnet API keys: https://testnet.binance.vision/
2. 修改配置 / Update configuration:
```yaml
binance:
  base_url: https://testnet.binance.vision
  testnet: true
```

## 使用方法 / Usage

### 启动应用 / Starting the Application

**现货交易 / Spot Trading:**
```bash
./binance-trader.exe spot
```

**合约交易 / Futures Trading:**
```bash
./binance-trader.exe futures
```

应用启动后会显示欢迎界面和命令提示符 / After starting, you'll see a welcome screen and command prompt.

### 快速入门指南 / Quick Start Guide

详细的合约交易快速入门指南请参阅 / For detailed futures trading quick start guide, see: [docs/FUTURES_QUICKSTART.md](docs/FUTURES_QUICKSTART.md)

### 可用命令 / Available Commands

#### 现货交易命令 / Spot Trading Commands

##### 市场数据命令 / Market Data Commands

| 命令 / Command | 说明 / Description | 示例 / Example |
|---------------|-------------------|---------------|
| `price <symbol>` | 获取当前价格 / Get current price | `price BTCUSDT` |
| `history <symbol> <interval> <limit>` | 获取历史K线 / Get historical klines | `history BTCUSDT 1h 10` |

**支持的时间间隔 / Supported Intervals:** `1m`, `5m`, `15m`, `30m`, `1h`, `4h`, `1d`, `1w`

#### 交易命令 / Trading Commands

| 命令 / Command | 说明 / Description | 示例 / Example |
|---------------|-------------------|---------------|
| `buy <symbol> <quantity>` | 市价买入 / Market buy order | `buy BTCUSDT 0.001` |
| `sell <symbol> <price> <quantity>` | 限价卖出 / Limit sell order | `sell BTCUSDT 50000 0.001` |
| `cancel <orderID>` | 取消订单 / Cancel order | `cancel 12345` |
| `status <orderID>` | 查询订单状态 / Get order status | `status 12345` |
| `orders` | 列出活跃订单 / List active orders | `orders` |

#### 条件订单命令 / Conditional Order Commands

| 命令 / Command | 说明 / Description | 示例 / Example |
|---------------|-------------------|---------------|
| `conditional-buy <symbol> <quantity> <trigger_price>` | 创建价格触发买单 / Create price-triggered buy order | `conditional-buy BTCUSDT 0.001 45000` |
| `conditional-sell <symbol> <quantity> <trigger_price>` | 创建价格触发卖单 / Create price-triggered sell order | `conditional-sell BTCUSDT 0.001 50000` |
| `conditional-orders` | 列出活跃条件订单 / List active conditional orders | `conditional-orders` |
| `cancel-conditional <orderID>` | 取消条件订单 / Cancel conditional order | `cancel-conditional abc123` |

#### 止损止盈命令 / Stop Loss & Take Profit Commands

| 命令 / Command | 说明 / Description | 示例 / Example |
|---------------|-------------------|---------------|
| `stop-loss <symbol> <position> <stop_price>` | 设置止损 / Set stop loss | `stop-loss BTCUSDT 0.001 42000` |
| `take-profit <symbol> <position> <target_price>` | 设置止盈 / Set take profit | `take-profit BTCUSDT 0.001 48000` |
| `stop-loss-take-profit <symbol> <position> <stop> <target>` | 同时设置止损止盈 / Set both stop-loss and take-profit | `stop-loss-take-profit BTCUSDT 0.001 42000 48000` |
| `trailing-stop <symbol> <position> <trail_percent>` | 设置移动止损 / Set trailing stop | `trailing-stop BTCUSDT 0.001 2.0` |
| `stop-orders` | 列出活跃止损止盈订单 / List active stop orders | `stop-orders` |

#### 合约交易命令 / Futures Trading Commands

##### 合约市场数据 / Futures Market Data

| 命令 / Command | 说明 / Description | 示例 / Example |
|---------------|-------------------|---------------|
| `mark-price <symbol>` | 获取标记价格 / Get mark price | `mark-price BTCUSDT` |
| `funding-rate <symbol>` | 获取资金费率 / Get funding rate | `funding-rate BTCUSDT` |
| `position <symbol>` | 查看持仓 / View position | `position BTCUSDT` |
| `positions` | 查看所有持仓 / View all positions | `positions` |

##### 合约交易 / Futures Trading

| 命令 / Command | 说明 / Description | 示例 / Example |
|---------------|-------------------|---------------|
| `long <symbol> <quantity>` | 开多仓（市价）/ Open long position (market) | `long BTCUSDT 0.001` |
| `short <symbol> <quantity>` | 开空仓（市价）/ Open short position (market) | `short BTCUSDT 0.001` |
| `long-limit <symbol> <price> <quantity>` | 开多仓（限价）/ Open long position (limit) | `long-limit BTCUSDT 45000 0.001` |
| `short-limit <symbol> <price> <quantity>` | 开空仓（限价）/ Open short position (limit) | `short-limit BTCUSDT 50000 0.001` |
| `close-position <symbol>` | 平仓 / Close position | `close-position BTCUSDT` |

##### 杠杆和保证金 / Leverage and Margin

| 命令 / Command | 说明 / Description | 示例 / Example |
|---------------|-------------------|---------------|
| `leverage <symbol> <value>` | 设置杠杆 / Set leverage | `leverage BTCUSDT 10` |
| `margin-type <symbol> <type>` | 设置保证金模式 / Set margin type | `margin-type BTCUSDT CROSSED` |
| `position-mode <mode>` | 设置仓位模式 / Set position mode | `position-mode true` |

##### 合约止损止盈 / Futures Stop Loss/Take Profit

| 命令 / Command | 说明 / Description | 示例 / Example |
|---------------|-------------------|---------------|
| `futures-stop-loss <symbol> <side> <quantity> <price>` | 设置止损 / Set stop loss | `futures-stop-loss BTCUSDT LONG 0.001 42000` |
| `futures-take-profit <symbol> <side> <quantity> <price>` | 设置止盈 / Set take profit | `futures-take-profit BTCUSDT LONG 0.001 48000` |

#### 系统命令 / System Commands

| 命令 / Command | 说明 / Description |
|---------------|-------------------|
| `help` | 显示帮助信息 / Show help |
| `exit` 或 `quit` | 退出程序 / Exit application |

### 使用示例 / Usage Examples

#### 示例 1: 查询价格并买入 / Example 1: Check Price and Buy

```
> price BTCUSDT
-------------------------------------------
Symbol: BTCUSDT
Price:  43250.50000000
-------------------------------------------

> buy BTCUSDT 0.001
-------------------------------------------
Order Created Successfully
-------------------------------------------
Order ID:       987654321
Symbol:         BTCUSDT
Side:           BUY
Type:           MARKET
Status:         FILLED
Price:          43250.50000000
Quantity:       0.00100000
Executed Qty:   0.00100000
Quote Qty:      43.25050000
-------------------------------------------
```

#### 示例 2: 设置限价卖出 / Example 2: Set Limit Sell Order

```
> sell BTCUSDT 45000 0.001
-------------------------------------------
Order Created Successfully
-------------------------------------------
Order ID:       987654322
Symbol:         BTCUSDT
Side:           SELL
Type:           LIMIT
Status:         NEW
Price:          45000.00000000
Quantity:       0.00100000
Executed Qty:   0.00000000
Quote Qty:      0.00000000
-------------------------------------------
```

#### 示例 3: 查看活跃订单 / Example 3: View Active Orders

```
> orders
===========================================
Active Orders (1)
===========================================

[1] Order ID: 987654322
    Symbol:       BTCUSDT
    Side:         SELL
    Type:         LIMIT
    Status:       NEW
    Price:        45000.00000000
    Quantity:     0.00100000
    Executed:     0.00000000
===========================================
```

#### 示例 4: 查看历史K线 / Example 4: View Historical Klines

```
> history BTCUSDT 1h 5
===========================================
Historical Klines for BTCUSDT (1h)
===========================================

[1] Time: 2024-12-04 10:00:00
    Open:   43100.00  High:   43300.00
    Low:    43050.00  Close:  43250.00
    Volume: 125.45

[2] Time: 2024-12-04 11:00:00
    Open:   43250.00  High:   43400.00
    Low:    43200.00  Close:  43350.00
    Volume: 98.32
...
===========================================
```

### 完整会话示例 / Complete Session Example

```
===========================================
  Binance Auto-Trading System
===========================================
Type 'help' for available commands

> help
Available commands:
  price <symbol>                    - Get current price
  buy <symbol> <quantity>           - Place market buy order
  sell <symbol> <price> <quantity>  - Place limit sell order
  cancel <orderID>                  - Cancel an order
  status <orderID>                  - Get order status
  orders                            - List all active orders
  history <symbol> <interval> <limit> - Get historical kline data
  help                              - Show this help message
  exit, quit                        - Exit the application

> price ETHUSDT
-------------------------------------------
Symbol: ETHUSDT
Price:  2250.75000000
-------------------------------------------

> buy ETHUSDT 0.01
-------------------------------------------
Order Created Successfully
-------------------------------------------
Order ID:       123456789
Symbol:         ETHUSDT
Side:           BUY
Type:           MARKET
Status:         FILLED
Price:          2250.75000000
Quantity:       0.01000000
Executed Qty:   0.01000000
Quote Qty:      22.50750000
-------------------------------------------

> orders
===========================================
Active Orders (0)
===========================================
No active orders
===========================================

> exit
Goodbye!
```

## 高级功能 / Advanced Features

### 条件订单 / Conditional Orders

条件订单允许您设置在满足特定市场条件时自动执行的订单。与普通限价单不同，条件订单在本地系统中监控，当条件满足时自动发送市价单到交易所，保证成交。

Conditional orders allow you to set orders that execute automatically when specific market conditions are met. Unlike regular limit orders, conditional orders are monitored locally and automatically send market orders to the exchange when conditions are met, ensuring execution.

#### 🔄 条件单 vs 限价单 / Conditional Orders vs Limit Orders

| 特性 / Feature | 限价单 / Limit Order | 条件单 / Conditional Order |
|---------------|---------------------|---------------------------|
| **提交时机** / **Submission** | 立即提交到交易所 / Immediately to exchange | 条件满足时提交 / When condition met |
| **订单类型** / **Order Type** | 限价单 / Limit | 市价单 / Market |
| **成交保证** / **Execution** | 不保证成交 / Not guaranteed | 保证成交 / Guaranteed |
| **监控位置** / **Monitoring** | 交易所 / Exchange | 本地系统 / Local system |
| **触发条件** / **Triggers** | 仅价格 / Price only | 价格、涨跌幅、成交量等 / Price, %, volume, etc. |
| **系统要求** / **System** | 可关闭 / Can shutdown | 必须运行 / Must run |

#### 📊 现货条件单 / Spot Conditional Orders

**命令行支持：** ✅ 已实现 / CLI Support: ✅ Implemented

##### 支持的触发条件 / Supported Trigger Conditions

1. **价格触发** / **Price Trigger**
   - 当价格达到指定水平时触发 / Triggers when price reaches specified level
   - 示例 / Example: 价格 >= 50000 / Price >= 50000

2. **涨跌幅触发** / **Percentage Change Trigger**
   - 当价格变化达到指定百分比时触发 / Triggers when price changes by specified percentage
   - 示例 / Example: 涨幅 >= 5% / Rise >= 5%

3. **成交量触发** / **Volume Trigger**
   - 当成交量达到指定阈值时触发 / Triggers when volume reaches specified threshold
   - 示例 / Example: 成交量 >= 1000000 / Volume >= 1000000

4. **复合条件** / **Composite Conditions**
   - 使用AND/OR逻辑组合多个条件 / Combine multiple conditions with AND/OR logic
   - 示例 / Example: (价格 >= 50000) AND (成交量 >= 100000)

##### 使用示例 / Usage Examples

**示例 1: 突破买入 / Breakout Buy**
```bash
# 当BTC价格突破50000时买入（突破策略）
# Buy BTC when price breaks above 50000 (breakout strategy)
> conditional-buy BTCUSDT 0.001 50000

# 系统会：
# 1. 保存条件订单到本地
# 2. 每秒监控BTC价格
# 3. 当价格 >= 50000 时，自动发送市价买单
# 4. 保证成交
```

**示例 2: 回调买入 / Pullback Buy**
```bash
# 当BTC价格回调到45000时买入（回调策略）
# Buy BTC when price pulls back to 45000 (pullback strategy)
> conditional-buy BTCUSDT 0.001 45000
```

**示例 3: 止损卖出 / Stop Loss Sell**
```bash
# 当BTC价格跌破42000时卖出（止损）
# Sell BTC when price drops below 42000 (stop loss)
> conditional-sell BTCUSDT 0.001 42000
```

**示例 4: 查看和管理条件订单 / View and Manage**
```bash
# 查看所有活跃的条件订单
# View all active conditional orders
> conditional-orders

Active Conditional Orders (2)
[1] ID=cond-001, Symbol=BTCUSDT, Side=BUY, Trigger=PRICE >= 50000
[2] ID=cond-002, Symbol=ETHUSDT, Side=SELL, Trigger=PRICE <= 2000

# 取消条件订单
# Cancel conditional order
> cancel-conditional cond-001
Conditional order cond-001 cancelled successfully
```

#### 🚀 合约条件单 / Futures Conditional Orders

**命令行支持：** ✅ 已实现 / CLI Support: ✅ Implemented

合约条件单支持更多触发类型，适用于合约交易的特殊需求。

Futures conditional orders support more trigger types for specific futures trading needs.

##### 支持的触发类型 / Supported Trigger Types

1. **标记价格触发** / **Mark Price Trigger**
   - 基于标记价格（更稳定，防止操纵）/ Based on mark price (more stable, manipulation-resistant)
   - 示例 / Example: 标记价 >= 50000 / Mark price >= 50000

2. **最新价格触发** / **Last Price Trigger**
   - 基于最新成交价 / Based on last traded price
   - 示例 / Example: 最新价 >= 50000 / Last price >= 50000

3. **未实现盈亏触发** / **Unrealized PnL Trigger**
   - 基于持仓的未实现盈亏 / Based on position's unrealized profit/loss
   - 示例 / Example: 盈亏 >= 1000 USDT / PnL >= 1000 USDT

4. **资金费率触发** / **Funding Rate Trigger**
   - 基于资金费率水平 / Based on funding rate level
   - 示例 / Example: 费率 >= 0.01% / Rate >= 0.01%

5. **保证金率触发** / **Margin Ratio Trigger**
   - 基于账户保证金率 / Based on account margin ratio
   - 示例 / Example: 保证金率 <= 10% / Margin ratio <= 10%

##### 合约条件单特性 / Futures Conditional Features

- ✅ **仓位方向控制** / **Position Side Control**: 支持 LONG/SHORT/BOTH
- ✅ **只减仓模式** / **Reduce Only Mode**: 只允许平仓，不开新仓
- ✅ **标记价格保护** / **Mark Price Protection**: 使用标记价格防止价格操纵
- ✅ **盈亏自动管理** / **PnL Auto Management**: 基于盈亏自动平仓

##### 使用示例 / Usage Examples

**示例 1: 标记价格突破开多 / Mark Price Breakout Long**
```bash
# CLI命令 / CLI Command
./binance-trader.exe futures

> condorder BTCUSDT BUY LONG 0.001 MARK_PRICE >= 50000
Conditional Order Created
Order ID:    cond-001
Symbol:      BTCUSDT
Side:        BUY
Position:    LONG
Quantity:    0.00100000
Trigger:     MARK_PRICE >= 50000.00000000
```

**代码示例 / Code Example:**
```go
request := &FuturesConditionalOrderRequest{
    Symbol:       "BTCUSDT",
    Side:         api.OrderSideBuy,
    PositionSide: api.PositionSideLong,
    Type:         api.OrderTypeMarket,
    Quantity:     0.001,
    TriggerCondition: &FuturesTriggerCondition{
        Type:      FuturesTriggerTypeMarkPrice,
        Operator:  OperatorGreaterEqual,
        Value:     50000.0,
        PriceType: api.PriceTypeMark,
    },
}

order, err := futuresConditionalService.CreateConditionalOrder(request)
```

**示例 2: 盈亏止盈 / PnL Take Profit**
```bash
# CLI命令 / CLI Command
> condorder BTCUSDT SELL LONG 0.001 PNL >= 1000
Conditional Order Created - Will close position when PnL reaches 1000 USDT
```

**代码示例 / Code Example:**
```go
// 当未实现盈亏达到1000 USDT时自动平仓
// Auto close position when unrealized PnL reaches 1000 USDT
request := &FuturesConditionalOrderRequest{
    Symbol:       "BTCUSDT",
    Side:         api.OrderSideSell,
    PositionSide: api.PositionSideLong,
    Type:         api.OrderTypeMarket,
    Quantity:     0.001,
    ReduceOnly:   true,  // 只减仓
    TriggerCondition: &FuturesTriggerCondition{
        Type:     FuturesTriggerTypeUnrealizedPnL,
        Operator: OperatorGreaterEqual,
        Value:    1000.0,
    },
}
```

**示例 3: 资金费率套利 / Funding Rate Arbitrage**
```bash
# CLI命令 / CLI Command
> condorder BTCUSDT SELL SHORT 0.001 FUNDING_RATE >= 0.0001
Conditional Order Created - Will open short when funding rate >= 0.01%
```

**代码示例 / Code Example:**
```go
// 当资金费率超过0.01%时开空仓（套利策略）
// Open short when funding rate exceeds 0.01% (arbitrage strategy)
request := &FuturesConditionalOrderRequest{
    Symbol:       "BTCUSDT",
    Side:         api.OrderSideSell,
    PositionSide: api.PositionSideShort,
    Type:         api.OrderTypeMarket,
    Quantity:     0.001,
    TriggerCondition: &FuturesTriggerCondition{
        Type:     FuturesTriggerTypeFundingRate,
        Operator: OperatorGreaterEqual,
        Value:    0.0001, // 0.01%
    },
}
```

#### ⚙️ 监控机制 / Monitoring Mechanism

条件订单通过后台监控引擎持续监控市场数据。

Conditional orders are continuously monitored by a background monitoring engine.

**监控流程 / Monitoring Flow:**
```
创建条件订单
  ↓
保存到本地数据库
  ↓
注册到监控引擎
  ↓
[每秒检查一次]
  ↓
评估触发条件
  ↓
条件满足？
  ├─ 是 → 发送市价单 → 更新状态为已执行
  └─ 否 → 继续监控
```

**配置参数 / Configuration:**
```yaml
# 现货条件订单配置 / Spot conditional orders
conditional_orders:
  monitoring_interval_ms: 1000      # 监控间隔（毫秒）/ Monitoring interval (ms)
  max_active_orders: 100            # 最大活跃订单数 / Max active orders
  trigger_execution_timeout_ms: 3000 # 触发执行超时 / Trigger timeout

# 合约条件订单配置 / Futures conditional orders
futures:
  monitoring:
    conditional_order_interval_ms: 1000  # 监控间隔 / Monitoring interval
```

#### 💡 使用建议 / Usage Tips

**何时使用条件单 / When to Use Conditional Orders:**
- ✅ 需要保证成交（条件满足时立即市价成交）
- ✅ 需要复杂触发条件（涨跌幅、成交量、盈亏等）
- ✅ 实施突破策略、回调策略
- ✅ 系统会持续运行

**何时使用限价单 / When to Use Limit Orders:**
- ✅ 不急于成交
- ✅ 希望以特定价格或更好价格交易
- ✅ 系统可能会关闭
- ✅ 只需要简单的价格条件

**最佳实践 / Best Practices:**
1. 先在测试网测试条件订单 / Test conditional orders on testnet first
2. 设置合理的触发条件，避免频繁触发 / Set reasonable triggers to avoid frequent execution
3. 监控系统日志，确保监控引擎正常运行 / Monitor logs to ensure monitoring engine runs properly
4. 定期检查活跃条件订单 / Regularly check active conditional orders
5. 为重要策略设置备用条件订单 / Set backup conditional orders for important strategies

#### 📝 日志和监控 / Logging and Monitoring

条件订单的所有活动都会被记录：

All conditional order activities are logged:

```
# 创建条件订单
{"level":"info","message":"Conditional order created","order_id":"cond-001","symbol":"BTCUSDT","trigger":"PRICE >= 50000"}

# 监控中
{"level":"debug","message":"Evaluating trigger condition","order_id":"cond-001","current_price":49500,"trigger_price":50000}

# 触发执行
{"level":"info","message":"Conditional order triggered","order_id":"cond-001","trigger_value":50100}
{"level":"info","message":"Market order placed","order_id":"12345","symbol":"BTCUSDT","side":"BUY"}

# 执行完成
{"level":"info","message":"Conditional order executed","order_id":"cond-001","executed_order_id":12345}
```

#### 🔗 相关文档 / Related Documentation

- 详细命令指南 / Detailed command guide: [docs/COMMAND_GUIDE.md](docs/COMMAND_GUIDE.md)
- API文档 / API documentation: [docs/API.md](docs/API.md)
- 使用示例 / Usage examples: [docs/EXAMPLES.md](docs/EXAMPLES.md)
- 合约快速入门 / Futures quick start: [docs/FUTURES_QUICKSTART.md](docs/FUTURES_QUICKSTART.md)

### 止损止盈 / Stop Loss & Take Profit

止损止盈功能帮助您自动保护利润和限制损失。

Stop-loss and take-profit features help you automatically protect profits and limit losses.

#### 功能特性 / Features

1. **止损订单** / **Stop Loss** - 当价格向不利方向移动时自动平仓 / Automatically close position when price moves unfavorably
2. **止盈订单** / **Take Profit** - 当价格达到目标利润时自动平仓 / Automatically close position when target profit is reached
3. **配对订单** / **Paired Orders** - 同时设置止损和止盈，任一触发后取消另一个 / Set both stop-loss and take-profit, cancel one when other triggers
4. **移动止损** / **Trailing Stop** - 随价格有利变动自动调整止损价格 / Automatically adjust stop price with favorable price movements

#### 使用示例 / Usage Example

```bash
# 为持仓设置止损
# Set stop loss for position
> stop-loss BTCUSDT 0.001 42000

# 为持仓设置止盈
# Set take profit for position
> take-profit BTCUSDT 0.001 48000

# 同时设置止损和止盈
# Set both stop-loss and take-profit
> stop-loss-take-profit BTCUSDT 0.001 42000 48000

# 设置2%的移动止损
# Set 2% trailing stop
> trailing-stop BTCUSDT 0.001 2.0
```

### 监控引擎 / Monitoring Engine

系统包含后台监控引擎，持续监控市场数据并评估触发条件。

The system includes a background monitoring engine that continuously monitors market data and evaluates trigger conditions.

- 默认监控间隔：1秒 / Default monitoring interval: 1 second
- 支持智能轮询优化 / Supports smart polling optimization
- 自动执行触发的订单 / Automatically executes triggered orders
- 完整的触发事件日志 / Complete trigger event logging

## API文档 / API Documentation

详细的API文档请参阅 [API.md](docs/API.md)

For detailed API documentation, see [API.md](docs/API.md)

### 核心接口 / Core Interfaces

- **BinanceClient** - 币安API客户端接口 / Binance API client interface
- **TradingService** - 交易服务接口 / Trading service interface
- **RiskManager** - 风险管理接口 / Risk management interface
- **MarketDataService** - 市场数据服务接口 / Market data service interface
- **OrderRepository** - 订单仓储接口 / Order repository interface
- **ConditionalOrderService** - 条件订单服务接口 / Conditional order service interface
- **StopLossService** - 止损止盈服务接口 / Stop-loss service interface
- **TriggerEngine** - 触发引擎接口 / Trigger engine interface
- **MonitoringEngine** - 监控引擎接口 / Monitoring engine interface

## 测试 / Testing

### 运行测试 / Running Tests

```bash
# 运行所有测试 / Run all tests
go test ./...

# 运行带覆盖率的测试 / Run tests with coverage
go test -cover ./...

# 运行详细测试 / Run tests with verbose output
go test -v ./...

# 运行特定包的测试 / Run tests for specific package
go test -v ./internal/api/...

# 生成覆盖率报告 / Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

### 测试类型 / Test Types

#### 单元测试 / Unit Tests
- 测试单个函数和方法 / Test individual functions and methods
- 使用模拟数据和httptest / Use mock data and httptest
- 文件命名: `*_test.go`

#### 属性测试 / Property-Based Tests
- 使用gopter框架 / Using gopter framework
- 每个测试运行100+次迭代 / Each test runs 100+ iterations
- 验证系统正确性属性 / Verify system correctness properties
- 标记格式: `// Feature: binance-auto-trading, Property X: [description]`

#### 集成测试 / Integration Tests
- 测试组件间交互 / Test component interactions
- 可使用币安测试网 / Can use Binance testnet
- 文件命名: `*_integration_test.go`

### 测试覆盖率目标 / Test Coverage Goals

- 总体覆盖率: 80%+ / Overall coverage: 80%+
- 核心业务逻辑: 90%+ / Core business logic: 90%+
- 所有正确性属性都有对应测试 / All correctness properties have corresponding tests

## 项目结构 / Project Structure

```
binance-trader/
├── cmd/                          # 应用程序入口点 / Application entry points
│   └── main.go                  # 主程序，依赖注入 / Main program, dependency injection
│
├── internal/                     # 私有应用程序代码 / Private application code
│   ├── api/                     # API客户端层 / API client layer
│   │   ├── client.go           # BinanceClient接口和实现 / BinanceClient interface & impl
│   │   ├── client_test.go      # 客户端测试 / Client tests
│   │   ├── auth.go             # 认证和签名 / Authentication and signing
│   │   ├── auth_test.go        # 认证测试 / Auth tests
│   │   ├── http_client.go      # HTTP客户端封装 / HTTP client wrapper
│   │   ├── http_client_test.go # HTTP客户端测试 / HTTP client tests
│   │   └── ratelimit.go        # 速率限制器 / Rate limiter
│   │
│   ├── cli/                     # 命令行界面 / Command-line interface
│   │   ├── cli.go              # CLI实现 / CLI implementation
│   │   └── cli_test.go         # CLI测试 / CLI tests
│   │
│   ├── config/                  # 配置管理 / Configuration management
│   │   ├── config.go           # 配置加载和验证 / Config loading & validation
│   │   └── config_test.go      # 配置测试 / Config tests
│   │
│   ├── repository/              # 数据持久化层 / Data persistence layer
│   │   ├── order.go            # 订单仓储接口和实现 / Order repository interface & impl
│   │   └── order_test.go       # 订单仓储测试 / Order repository tests
│   │
│   └── service/                 # 业务逻辑层 / Business logic layer
│       ├── trading.go          # 交易服务 / Trading service
│       ├── trading_test.go     # 交易服务测试 / Trading service tests
│       ├── risk.go             # 风险管理器 / Risk manager
│       ├── market.go           # 市场数据服务 / Market data service
│       ├── market_test.go      # 市场数据测试 / Market data tests
│       └── test_helpers.go     # 测试辅助函数 / Test helper functions
│
├── pkg/                         # 公共/可重用包 / Public/reusable packages
│   ├── errors/                 # 错误类型 / Error types
│   │   ├── errors.go          # 错误定义 / Error definitions
│   │   └── errors_test.go     # 错误测试 / Error tests
│   │
│   └── logger/                 # 日志工具 / Logging utilities
│       ├── logger.go          # 日志实现 / Logger implementation
│       └── logger_test.go     # 日志测试 / Logger tests
│
├── docs/                        # 文档 / Documentation
│   ├── API.md                  # API文档 / API documentation
│   └── EXAMPLES.md             # 使用示例 / Usage examples
│
├── logs/                        # 日志文件目录 / Log files directory
│   └── trading.log             # 交易日志 / Trading logs
│
├── .kiro/                       # Kiro规范文件 / Kiro spec files
│   └── specs/                  # 功能规范 / Feature specs
│       └── binance-auto-trading/
│           ├── requirements.md # 需求文档 / Requirements
│           ├── design.md       # 设计文档 / Design
│           └── tasks.md        # 任务列表 / Task list
│
├── config.yaml                  # 配置文件 / Configuration file
├── config.example.yaml          # 配置示例 / Configuration example
├── go.mod                       # Go模块定义 / Go module definition
├── go.sum                       # 依赖校验和 / Dependency checksums
└── README.md                    # 本文件 / This file
```

### 架构说明 / Architecture Overview

系统采用三层架构 / The system uses a three-layer architecture:

1. **API客户端层** / **API Client Layer** (`internal/api/`)
   - 处理与币安API的通信 / Handles communication with Binance API
   - 实现认证、签名、速率限制 / Implements auth, signing, rate limiting
   - 管理HTTP请求和重试 / Manages HTTP requests and retries

2. **业务逻辑层** / **Business Logic Layer** (`internal/service/`)
   - 实现交易策略和订单管理 / Implements trading strategies and order management
   - 执行风险控制规则 / Executes risk control rules
   - 处理市场数据缓存 / Handles market data caching

3. **数据层** / **Data Layer** (`internal/repository/`, `internal/config/`)
   - 管理订单数据持久化 / Manages order data persistence
   - 处理配置加载和验证 / Handles config loading and validation
   - 提供统一的日志记录 / Provides unified logging

## 安全性 / Security

### 🔐 API密钥管理 / API Key Management

- ❌ **永远不要**将API密钥提交到版本控制 / **Never** commit API keys to version control
- ✅ 始终使用环境变量存储敏感数据 / Always use environment variables for sensitive data
- ✅ API密钥在日志中自动屏蔽 / API keys are automatically masked in logs
- ✅ 建议使用IP白名单限制API访问 / Recommended to use IP whitelist for API access

### 🔒 通信安全 / Communication Security

- 所有API通信强制使用HTTPS / All API communication enforces HTTPS
- 使用HMAC SHA256签名所有私有请求 / Uses HMAC SHA256 to sign all private requests
- 包含时间戳防止重放攻击 / Includes timestamp to prevent replay attacks

### 📝 日志安全 / Logging Security

敏感信息自动屏蔽 / Sensitive information is automatically masked:
- API密钥只显示前4位和后4位 / API keys show only first 4 and last 4 characters
- API密钥完全隐藏 / API secrets are completely hidden
- 示例 / Example: `abcd****xyz123`

### 🛡️ 最佳实践 / Best Practices

1. 使用只读API密钥进行测试 / Use read-only API keys for testing
2. 为生产环境创建单独的API密钥 / Create separate API keys for production
3. 定期轮换API密钥 / Rotate API keys regularly
4. 启用IP白名单 / Enable IP whitelist
5. 监控API使用情况 / Monitor API usage
6. 先在测试网测试 / Test on testnet first

## 风险管理 / Risk Management

系统包含多层风险控制机制 / The system includes multi-layer risk control mechanisms:

### 💰 订单限制 / Order Limits

```yaml
risk:
  max_order_amount: 10000.0      # 单笔最大金额 / Max single order amount
  max_daily_orders: 100          # 每日最大订单数 / Max daily orders
  min_balance_reserve: 100.0     # 最小保留余额 / Min balance reserve
```

### 🚦 速率限制 / Rate Limiting

- 自动管理API调用频率 / Automatically manages API call frequency
- 防止超过币安速率限制 / Prevents exceeding Binance rate limits
- 检测到限制时自动降速 / Automatically slows down when limits detected

### 🔄 错误处理 / Error Handling

- 网络错误自动重试（指数退避）/ Network errors auto-retry (exponential backoff)
- 余额不足自动拒绝订单 / Insufficient balance auto-rejects orders
- 详细的错误日志便于调试 / Detailed error logs for debugging

### ⚠️ 风险提示 / Risk Warnings

- ⚠️ 加密货币交易存在高风险 / Cryptocurrency trading involves high risk
- ⚠️ 仅投资您能承受损失的资金 / Only invest what you can afford to lose
- ⚠️ 先在测试网充分测试 / Test thoroughly on testnet first
- ⚠️ 从小额订单开始 / Start with small orders
- ⚠️ 持续监控系统运行 / Continuously monitor system operation

## 故障排除 / Troubleshooting

### 常见问题 / Common Issues

#### 1. 认证失败 / Authentication Failed

**问题 / Problem:** `Authentication failed: invalid signature`

**解决方案 / Solution:**
- 检查API密钥和密钥是否正确 / Check if API key and secret are correct
- 确保系统时间同步 / Ensure system time is synchronized
- 验证API密钥权限 / Verify API key permissions

```bash
# 同步系统时间 / Sync system time (Linux)
sudo ntpdate -s time.nist.gov
```

#### 2. 速率限制错误 / Rate Limit Error

**问题 / Problem:** `Rate limit exceeded`

**解决方案 / Solution:**
- 降低 `max_api_calls_per_min` 配置 / Lower `max_api_calls_per_min` config
- 等待速率限制窗口重置 / Wait for rate limit window to reset
- 检查是否有其他程序使用同一API密钥 / Check if other programs use same API key

#### 3. 余额不足 / Insufficient Balance

**问题 / Problem:** `Insufficient balance`

**解决方案 / Solution:**
- 检查账户余额 / Check account balance
- 降低订单数量 / Reduce order quantity
- 调整 `min_balance_reserve` 配置 / Adjust `min_balance_reserve` config

#### 4. 网络连接问题 / Network Connection Issues

**问题 / Problem:** `Connection timeout` 或 `Network error`

**解决方案 / Solution:**
- 检查网络连接 / Check network connection
- 验证防火墙设置 / Verify firewall settings
- 尝试使用VPN（如果币安在您的地区受限）/ Try using VPN (if Binance is restricted in your region)

#### 5. 配置文件错误 / Configuration File Error

**问题 / Problem:** `Failed to load configuration`

**解决方案 / Solution:**
- 验证YAML语法 / Verify YAML syntax
- 检查文件路径 / Check file path
- 确保环境变量已设置 / Ensure environment variables are set

### 调试模式 / Debug Mode

启用详细日志 / Enable verbose logging:

```yaml
logging:
  level: debug  # 改为debug获取更多信息 / Change to debug for more info
```

或使用环境变量 / Or use environment variable:

```bash
export LOG_LEVEL=debug
```

### 获取帮助 / Getting Help

如果问题仍未解决 / If issues persist:

1. 查看日志文件 `logs/trading.log` / Check log file `logs/trading.log`
2. 启用debug日志级别 / Enable debug log level
3. 在测试网重现问题 / Reproduce issue on testnet
4. 提交issue并附上日志（记得屏蔽敏感信息）/ Submit issue with logs (mask sensitive info)

## 贡献 / Contributing

欢迎贡献！请遵循以下步骤 / Contributions are welcome! Please follow these steps:

1. Fork本仓库 / Fork the repository
2. 创建功能分支 / Create a feature branch (`git checkout -b feature/amazing-feature`)
3. 提交更改 / Commit your changes (`git commit -m 'Add amazing feature'`)
4. 推送到分支 / Push to the branch (`git push origin feature/amazing-feature`)
5. 开启Pull Request / Open a Pull Request

### 开发指南 / Development Guidelines

- 遵循Go代码规范 / Follow Go code conventions
- 为新功能编写测试 / Write tests for new features
- 更新文档 / Update documentation
- 确保所有测试通过 / Ensure all tests pass
- 保持代码覆盖率 / Maintain code coverage

## 依赖项 / Dependencies

### 核心依赖 / Core Dependencies

- **Go 1.21+** - 编程语言 / Programming language
- **github.com/adshao/go-binance/v2** - 币安Go SDK / Binance Go SDK
- **gopkg.in/yaml.v3** - YAML配置解析 / YAML configuration parsing
- **github.com/sirupsen/logrus** - 结构化日志 / Structured logging
- **github.com/leanovate/gopter** - 属性测试框架 / Property-based testing framework

### 开发依赖 / Development Dependencies

- Go标准库 `testing` - 单元测试 / Unit testing
- `httptest` - HTTP模拟 / HTTP mocking

## 许可证 / License

MIT License - 详见 [LICENSE](LICENSE) 文件 / See [LICENSE](LICENSE) file for details

## 免责声明 / Disclaimer

本软件仅供教育和研究目的。使用本软件进行实际交易需自行承担风险。作者不对任何交易损失负责。

This software is for educational and research purposes only. Use this software for actual trading at your own risk. The authors are not responsible for any trading losses.

---

**⭐ 如果这个项目对您有帮助，请给个星标！/ If this project helps you, please give it a star!**
