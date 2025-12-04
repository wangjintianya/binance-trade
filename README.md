# Binance Auto-Trading System

币安自动交易系统 - 使用Go语言开发的自动化加密货币交易应用程序

An automated cryptocurrency trading application that integrates with the Binance exchange API to execute trades programmatically with built-in risk management and comprehensive logging.

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
- ✅ **安全的API集成** / **Secure API Integration** - HMAC SHA256 authentication with automatic request signing
- 📊 **实时市场数据** / **Real-time Market Data** - Prices, K-lines, and account balances
- 🤖 **自动化订单管理** / **Automated Order Management** - Market orders and limit orders
- 🛡️ **风险控制机制** / **Risk Control** - Order limits, balance protection, and rate limiting
- 📝 **完整的日志记录** / **Comprehensive Logging** - Structured logging with sensitive data masking
- 💻 **交互式CLI** / **Interactive CLI** - User-friendly command-line interface
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

```bash
./binance-trader.exe
```

## 配置 / Configuration

### 配置文件 / Configuration File

创建或编辑 `config.yaml` 文件（参考 `config.example.yaml`）：

Create or edit `config.yaml` file (see `config.example.yaml` for reference):

```yaml
binance:
  api_key: ${BINANCE_API_KEY}        # 从环境变量读取 / Read from environment
  api_secret: ${BINANCE_API_SECRET}  # 从环境变量读取 / Read from environment
  base_url: https://api.binance.com  # 生产环境 / Production
  # base_url: https://testnet.binance.vision  # 测试网 / Testnet
  testnet: false                     # 设为true使用测试网 / Set to true for testnet

risk:
  max_order_amount: 10000.0          # 单笔最大金额(USDT) / Max order amount (USDT)
  max_daily_orders: 100              # 每日最大订单数 / Max daily orders
  min_balance_reserve: 100.0         # 最小保留余额(USDT) / Min balance reserve (USDT)
  max_api_calls_per_min: 1000        # 每分钟最大API调用 / Max API calls per minute

logging:
  level: info                        # 日志级别: debug, info, warn, error / Log level
  file: logs/trading.log             # 日志文件路径 / Log file path
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
| `BINANCE_API_KEY` | ✅ Yes | 币安API密钥 / Binance API key |
| `BINANCE_API_SECRET` | ✅ Yes | 币安API密钥 / Binance API secret |
| `CONFIG_FILE` | ❌ No | 配置文件路径 / Config file path (default: `config.yaml`) |
| `LOG_LEVEL` | ❌ No | 日志级别 / Log level (default: `info`) |

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

```bash
./binance-trader.exe
```

应用启动后会显示欢迎界面和命令提示符 / After starting, you'll see a welcome screen and command prompt.

### 可用命令 / Available Commands

#### 市场数据命令 / Market Data Commands

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

## API文档 / API Documentation

详细的API文档请参阅 [API.md](docs/API.md)

For detailed API documentation, see [API.md](docs/API.md)

### 核心接口 / Core Interfaces

- **BinanceClient** - 币安API客户端接口 / Binance API client interface
- **TradingService** - 交易服务接口 / Trading service interface
- **RiskManager** - 风险管理接口 / Risk management interface
- **MarketDataService** - 市场数据服务接口 / Market data service interface
- **OrderRepository** - 订单仓储接口 / Order repository interface

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
