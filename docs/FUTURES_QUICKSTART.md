# 合约交易快速入门指南 / Futures Trading Quick Start Guide

本指南将帮助您快速开始使用币安U本位合约交易系统。

This guide will help you quickly get started with Binance USDT-M Futures trading system.

## 目录 / Table of Contents

- [什么是合约交易 / What is Futures Trading](#什么是合约交易--what-is-futures-trading)
- [前置要求 / Prerequisites](#前置要求--prerequisites)
- [配置设置 / Configuration Setup](#配置设置--configuration-setup)
- [基础概念 / Basic Concepts](#基础概念--basic-concepts)
- [第一笔合约交易 / Your First Futures Trade](#第一笔合约交易--your-first-futures-trade)
- [风险管理 / Risk Management](#风险管理--risk-management)
- [常见问题 / FAQ](#常见问题--faq)

---

## 什么是合约交易 / What is Futures Trading

合约交易允许您使用杠杆进行加密货币交易，可以做多（预期价格上涨）或做空（预期价格下跌）。

Futures trading allows you to trade cryptocurrencies with leverage, enabling you to go long (expecting price to rise) or short (expecting price to fall).

### 关键特性 / Key Features

- **杠杆交易** / **Leveraged Trading**: 使用1x-125x杠杆放大交易规模 / Use 1x-125x leverage to amplify trading size
- **双向交易** / **Bidirectional Trading**: 可以做多或做空 / Can go long or short
- **永续合约** / **Perpetual Contracts**: 没有到期日的合约 / Contracts with no expiration date
- **资金费率** / **Funding Rate**: 多空双方之间的定期支付 / Periodic payments between long and short positions

### ⚠️ 风险警告 / Risk Warning

合约交易具有高风险，可能导致全部本金损失。请确保：

Futures trading carries high risk and can result in total loss of capital. Please ensure:

- 您完全理解杠杆交易的风险 / You fully understand the risks of leveraged trading
- 只投资您能承受损失的资金 / Only invest what you can afford to lose
- 先在测试网练习 / Practice on testnet first
- 从低杠杆开始 / Start with low leverage

---

## 前置要求 / Prerequisites

### 1. 币安账户设置 / Binance Account Setup

1. 注册币安账户 / Register Binance account: https://www.binance.com/
2. 完成KYC验证 / Complete KYC verification
3. 开通合约交易权限 / Enable futures trading permission
4. 创建API密钥 / Create API keys:
   - 登录币安 / Log in to Binance
   - 进入 API管理 / Go to API Management
   - 创建新的API密钥 / Create new API key
   - 启用"合约交易"权限 / Enable "Futures Trading" permission
   - 建议设置IP白名单 / Recommended to set IP whitelist

### 2. 系统要求 / System Requirements

- Go 1.21+ 已安装 / Go 1.21+ installed
- 稳定的网络连接 / Stable internet connection
- 足够的USDT余额 / Sufficient USDT balance

---

## 配置设置 / Configuration Setup

### 1. 环境变量 / Environment Variables

设置合约API密钥 / Set futures API keys:

**Linux/macOS:**
```bash
export BINANCE_FUTURES_API_KEY="your_futures_api_key"
export BINANCE_FUTURES_API_SECRET="your_futures_api_secret"
```

**Windows (PowerShell):**
```powershell
$env:BINANCE_FUTURES_API_KEY="your_futures_api_key"
$env:BINANCE_FUTURES_API_SECRET="your_futures_api_secret"
```

### 2. 配置文件 / Configuration File

编辑 `config.yaml` 添加合约配置 / Edit `config.yaml` to add futures configuration:

```yaml
futures:
  api_key: ${BINANCE_FUTURES_API_KEY}
  api_secret: ${BINANCE_FUTURES_API_SECRET}
  base_url: https://fapi.binance.com
  testnet: false  # 建议先设为true在测试网练习 / Recommended to set true for testnet practice
  
  # 默认设置 / Default settings
  default_leverage: 5          # 建议从低杠杆开始 / Start with low leverage
  default_margin_type: CROSSED # CROSSED (全仓) 或 ISOLATED (逐仓)
  dual_side_position: false    # false=单向持仓, true=双向持仓
  
  # 风险限制 / Risk limits
  risk:
    max_order_value: 10000.0      # 单笔最大订单价值 / Max order value
    max_position_value: 20000.0   # 最大持仓价值 / Max position value
    max_leverage: 10              # 最大杠杆 / Max leverage
    min_margin_ratio: 0.10        # 最小保证金率 / Min margin ratio
    liquidation_buffer: 0.05      # 强平缓冲区 / Liquidation buffer
```

### 3. 测试网配置 / Testnet Configuration

**强烈建议先在测试网练习！/ Highly recommended to practice on testnet first!**

1. 获取测试网API密钥 / Get testnet API keys: https://testnet.binancefuture.com/
2. 修改配置 / Update configuration:
```yaml
futures:
  base_url: https://testnet.binancefuture.com
  testnet: true
```

---

## 基础概念 / Basic Concepts

### 1. 杠杆 / Leverage

杠杆允许您用较小的资金控制较大的仓位。

Leverage allows you to control a larger position with smaller capital.

**示例 / Example:**
- 10x杠杆：用1000 USDT可以开10000 USDT的仓位 / 10x leverage: 1000 USDT can open 10000 USDT position
- 更高杠杆 = 更高风险 / Higher leverage = Higher risk

### 2. 保证金模式 / Margin Mode

**全仓保证金 (CROSSED):**
- 使用账户全部可用余额作为保证金 / Uses entire account balance as margin
- 风险分散到所有持仓 / Risk spread across all positions
- 适合经验丰富的交易者 / Suitable for experienced traders

**逐仓保证金 (ISOLATED):**
- 每个仓位使用独立的保证金 / Each position uses separate margin
- 风险隔离，最多损失该仓位的保证金 / Risk isolated, max loss is position margin
- 适合初学者 / Suitable for beginners

### 3. 仓位模式 / Position Mode

**单向持仓模式 (One-way Mode):**
- 同一合约只能持有一个方向的仓位 / Can only hold one direction per contract
- 简单直观 / Simple and intuitive
- 适合大多数交易者 / Suitable for most traders

**双向持仓模式 (Hedge Mode):**
- 同一合约可以同时持有多头和空头 / Can hold both long and short simultaneously
- 用于对冲策略 / Used for hedging strategies
- 适合高级交易者 / Suitable for advanced traders

### 4. 标记价格 / Mark Price

标记价格是用于计算未实现盈亏和强平价格的公允价格，避免市场操纵。

Mark price is the fair price used to calculate unrealized PnL and liquidation price, preventing market manipulation.

### 5. 资金费率 / Funding Rate

永续合约中多空双方之间每8小时结算一次的费用。

Fee settled every 8 hours between long and short positions in perpetual contracts.

- 正费率：多头支付给空头 / Positive rate: longs pay shorts
- 负费率：空头支付给多头 / Negative rate: shorts pay longs

### 6. 强平价格 / Liquidation Price

当标记价格达到强平价格时，仓位将被强制平仓。

When mark price reaches liquidation price, position will be forcibly closed.

---

## 第一笔合约交易 / Your First Futures Trade

### 步骤 1: 启动合约交易系统 / Step 1: Start Futures Trading System

```bash
./binance-trader.exe futures
```

### 步骤 2: 查看账户信息 / Step 2: Check Account Information

```bash
> balance
-------------------------------------------
USDT Balance
-------------------------------------------
Total Balance:      1000.00000000
Available Balance:  1000.00000000
Margin Used:        0.00000000
-------------------------------------------
```

### 步骤 3: 设置杠杆 / Step 3: Set Leverage

建议从低杠杆开始（2x-5x）/ Start with low leverage (2x-5x):

```bash
> leverage BTCUSDT 5
-------------------------------------------
Leverage Updated
-------------------------------------------
Symbol:   BTCUSDT
Leverage: 5x
-------------------------------------------
```

### 步骤 4: 查看当前价格 / Step 4: Check Current Price

```bash
> mark-price BTCUSDT
-------------------------------------------
Mark Price Information
-------------------------------------------
Symbol:          BTCUSDT
Mark Price:      45000.00000000
Index Price:     44995.50000000
Funding Rate:    0.0001
Next Funding:    2024-12-09 16:00:00
-------------------------------------------
```

### 步骤 5: 开仓 / Step 5: Open Position

**开多仓（预期价格上涨）/ Open Long (Expecting Price Rise):**

```bash
> long BTCUSDT 0.01
-------------------------------------------
Order Created Successfully
-------------------------------------------
Order ID:       123456789
Symbol:         BTCUSDT
Side:           BUY
Position Side:  LONG
Type:           MARKET
Status:         FILLED
Quantity:       0.01000000
Entry Price:    45000.00000000
Position Value: 450.00 USDT
Margin Used:    90.00 USDT (5x leverage)
-------------------------------------------
```

**开空仓（预期价格下跌）/ Open Short (Expecting Price Fall):**

```bash
> short BTCUSDT 0.01
-------------------------------------------
Order Created Successfully
-------------------------------------------
Order ID:       123456790
Symbol:         BTCUSDT
Side:           SELL
Position Side:  SHORT
Type:           MARKET
Status:         FILLED
Quantity:       0.01000000
Entry Price:    45000.00000000
Position Value: 450.00 USDT
Margin Used:    90.00 USDT (5x leverage)
-------------------------------------------
```

### 步骤 6: 查看持仓 / Step 6: View Position

```bash
> position BTCUSDT
-------------------------------------------
Position Information
-------------------------------------------
Symbol:              BTCUSDT
Position Side:       LONG
Position Amount:     0.01000000
Entry Price:         45000.00000000
Mark Price:          45100.00000000
Unrealized PnL:      +5.00 USDT (+1.11%)
Liquidation Price:   36000.00000000
Margin Type:         CROSSED
Leverage:            5x
-------------------------------------------
```

### 步骤 7: 设置止损止盈 / Step 7: Set Stop Loss and Take Profit

**设置止损（保护下行风险）/ Set Stop Loss (Protect Downside):**

```bash
> futures-stop-loss BTCUSDT LONG 0.01 44000
-------------------------------------------
Stop Loss Order Created
-------------------------------------------
Order ID:     sl_123456
Symbol:       BTCUSDT
Position:     0.01000000 LONG
Stop Price:   44000.00000000
Status:       ACTIVE
-------------------------------------------
Will automatically close position if price drops to $44000
```

**设置止盈（锁定利润）/ Set Take Profit (Lock in Profit):**

```bash
> futures-take-profit BTCUSDT LONG 0.01 46000
-------------------------------------------
Take Profit Order Created
-------------------------------------------
Order ID:     tp_123456
Symbol:       BTCUSDT
Position:     0.01000000 LONG
Target Price: 46000.00000000
Status:       ACTIVE
-------------------------------------------
Will automatically close position if price rises to $46000
```

### 步骤 8: 平仓 / Step 8: Close Position

```bash
> close-position BTCUSDT
-------------------------------------------
Position Closed Successfully
-------------------------------------------
Symbol:         BTCUSDT
Closed Amount:  0.01000000
Entry Price:    45000.00000000
Exit Price:     45100.00000000
Realized PnL:   +5.00 USDT
-------------------------------------------
```

---

## 风险管理 / Risk Management

### 1. 使用止损 / Use Stop Loss

**永远设置止损！/ Always set stop loss!**

```bash
# 设置2%的止损 / Set 2% stop loss
# 如果入场价格是45000，止损价格是44100
# If entry price is 45000, stop loss price is 44100
> futures-stop-loss BTCUSDT LONG 0.01 44100
```

### 2. 控制杠杆 / Control Leverage

| 杠杆 / Leverage | 风险等级 / Risk Level | 适合 / Suitable For |
|----------------|---------------------|-------------------|
| 1x-3x | 低 / Low | 初学者 / Beginners |
| 3x-10x | 中 / Medium | 有经验的交易者 / Experienced traders |
| 10x-20x | 高 / High | 专业交易者 / Professional traders |
| 20x+ | 极高 / Very High | 不推荐 / Not recommended |

### 3. 仓位管理 / Position Sizing

**建议规则 / Recommended Rules:**

- 单笔交易不超过账户的2-5% / Single trade should not exceed 2-5% of account
- 总持仓不超过账户的20-30% / Total positions should not exceed 20-30% of account
- 保持足够的保证金余额 / Maintain sufficient margin balance

**示例 / Example:**
```
账户余额 / Account Balance: 10000 USDT
单笔交易限制 / Single Trade Limit: 500 USDT (5%)
总持仓限制 / Total Position Limit: 3000 USDT (30%)
```

### 4. 监控强平风险 / Monitor Liquidation Risk

系统会自动监控并警告强平风险：

System automatically monitors and warns about liquidation risk:

```
⚠️  WARNING: Liquidation Risk High!
Current Price:      44500.00
Liquidation Price:  44200.00
Distance:           0.67%
Recommendation:     Add margin or reduce position
```

### 5. 资金费率管理 / Funding Rate Management

注意资金费率，避免长期持有高费率仓位：

Pay attention to funding rate, avoid holding high-rate positions long-term:

```bash
> funding-rate BTCUSDT
-------------------------------------------
Funding Rate Information
-------------------------------------------
Symbol:          BTCUSDT
Current Rate:    0.0100 (1.00%)  # 高费率！/ High rate!
Next Funding:    2024-12-09 16:00:00
Estimated Fee:   -4.50 USDT (for 0.01 BTC long position)
-------------------------------------------
```

---

## 常见问题 / FAQ

### Q1: 合约交易和现货交易有什么区别？/ What's the difference between futures and spot trading?

**现货交易 / Spot Trading:**
- 实际拥有加密货币 / Actually own the cryptocurrency
- 只能做多 / Can only go long
- 无杠杆（或低杠杆）/ No leverage (or low leverage)
- 无资金费率 / No funding rate

**合约交易 / Futures Trading:**
- 不实际拥有，只是合约 / Don't actually own, just contracts
- 可以做多或做空 / Can go long or short
- 高杠杆（1x-125x）/ High leverage (1x-125x)
- 有资金费率 / Has funding rate

### Q2: 我应该使用多少杠杆？/ How much leverage should I use?

**建议 / Recommendations:**
- 初学者：2x-3x / Beginners: 2x-3x
- 中级：5x-10x / Intermediate: 5x-10x
- 高级：根据策略 / Advanced: Based on strategy

**记住 / Remember:** 更高杠杆 = 更高风险 = 更容易爆仓 / Higher leverage = Higher risk = Easier liquidation

### Q3: 什么是强平？如何避免？/ What is liquidation? How to avoid it?

**强平 / Liquidation:** 当您的保证金不足以维持仓位时，系统会强制平仓。

Liquidation: When your margin is insufficient to maintain position, system forcibly closes it.

**避免方法 / How to Avoid:**
1. 使用低杠杆 / Use low leverage
2. 设置止损 / Set stop loss
3. 保持足够的保证金余额 / Maintain sufficient margin balance
4. 不要满仓交易 / Don't use full account balance
5. 监控强平价格 / Monitor liquidation price

### Q4: 全仓和逐仓哪个更好？/ Which is better: Cross or Isolated margin?

**全仓 (CROSSED):**
- ✅ 优点：不容易被强平 / Pros: Less likely to be liquidated
- ❌ 缺点：一个仓位爆仓可能影响全部余额 / Cons: One liquidation can affect entire balance

**逐仓 (ISOLATED):**
- ✅ 优点：风险隔离，最多损失该仓位保证金 / Pros: Risk isolated, max loss is position margin
- ❌ 缺点：更容易被强平 / Cons: Easier to be liquidated

**建议 / Recommendation:** 初学者使用逐仓 / Beginners use isolated margin

### Q5: 资金费率是什么？如何影响我？/ What is funding rate? How does it affect me?

资金费率是永续合约中多空双方之间每8小时结算一次的费用。

Funding rate is a fee settled every 8 hours between long and short positions in perpetual contracts.

**影响 / Impact:**
- 正费率：持有多头需要支付费用 / Positive rate: Long positions pay fee
- 负费率：持有空头需要支付费用 / Negative rate: Short positions pay fee
- 费率通常很小（0.01%-0.03%）/ Rate usually small (0.01%-0.03%)
- 长期持仓需要考虑累积费用 / Long-term positions need to consider accumulated fees

### Q6: 我可以同时运行现货和合约交易吗？/ Can I run spot and futures trading simultaneously?

可以！系统支持双入口点：

Yes! System supports dual entry points:

```bash
# 终端1 / Terminal 1
./binance-trader.exe spot

# 终端2 / Terminal 2
./binance-trader.exe futures
```

### Q7: 如何查看我的交易历史？/ How to view my trading history?

```bash
> positions  # 查看当前持仓 / View current positions
> orders     # 查看活跃订单 / View active orders
```

日志文件中也会记录所有交易活动：

All trading activities are also logged in log files:
- 合约日志 / Futures logs: `logs/futures_trading.log`

---

## 下一步 / Next Steps

1. **阅读完整文档 / Read Full Documentation**
   - [API文档 / API Documentation](API.md)
   - [使用示例 / Usage Examples](EXAMPLES.md)

2. **练习策略 / Practice Strategies**
   - 在测试网练习不同的交易策略 / Practice different strategies on testnet
   - 从小额开始 / Start with small amounts
   - 记录交易日志 / Keep trading journal

3. **学习高级功能 / Learn Advanced Features**
   - 条件订单 / Conditional orders
   - 移动止损 / Trailing stop
   - 风险管理工具 / Risk management tools

4. **加入社区 / Join Community**
   - 与其他交易者交流经验 / Exchange experiences with other traders
   - 学习市场分析 / Learn market analysis
   - 持续改进策略 / Continuously improve strategies

---

## 重要提醒 / Important Reminders

⚠️ **风险警告 / Risk Warning:**
- 合约交易具有高风险 / Futures trading carries high risk
- 可能损失全部本金 / May lose entire capital
- 不要投资超过您能承受损失的资金 / Don't invest more than you can afford to lose
- 先在测试网充分练习 / Practice thoroughly on testnet first

📚 **持续学习 / Continuous Learning:**
- 学习技术分析 / Learn technical analysis
- 了解市场动态 / Understand market dynamics
- 制定交易计划 / Develop trading plan
- 严格执行风险管理 / Strictly execute risk management

🎯 **交易纪律 / Trading Discipline:**
- 永远设置止损 / Always set stop loss
- 不要情绪化交易 / Don't trade emotionally
- 遵守交易计划 / Follow trading plan
- 记录和分析每笔交易 / Record and analyze every trade

---

**祝您交易顺利！/ Happy Trading!**

如有问题，请参阅完整文档或提交issue。

For questions, please refer to full documentation or submit an issue.
