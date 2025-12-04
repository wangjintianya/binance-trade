# Project Structure

## Architecture Pattern

Three-layer architecture with clear separation of concerns:

1. **API Client Layer** - External communication (BinanceClient, AuthManager, RateLimiter)
2. **Business Logic Layer** - Core functionality (TradingService, RiskManager, MarketDataService)
3. **Data Layer** - Persistence and configuration (OrderRepository, ConfigManager, Logger)

## Directory Layout

```
binance-trader/
├── cmd/                    # Application entry points
│   └── main.go            # Main application
├── internal/              # Private application code
│   ├── api/              # API client implementation
│   │   ├── client.go     # BinanceClient interface & implementation
│   │   ├── auth.go       # Authentication and signing
│   │   └── ratelimit.go  # Rate limiting logic
│   ├── service/          # Business logic services
│   │   ├── trading.go    # TradingService implementation
│   │   ├── risk.go       # RiskManager implementation
│   │   └── market.go     # MarketDataService implementation
│   ├── repository/       # Data persistence
│   │   └── order.go      # OrderRepository implementation
│   └── config/           # Configuration management
│       └── config.go     # ConfigManager implementation
├── pkg/                   # Public/reusable packages
│   ├── models/           # Data models (Order, Balance, Kline, etc.)
│   ├── logger/           # Logging utilities
│   └── errors/           # Error types and handling
├── config.yaml           # Configuration file
├── logs/                 # Log files directory
├── go.mod                # Go module definition
└── go.sum                # Dependency checksums
```

## Module Organization

- **cmd/** - Keep minimal, only initialization and dependency injection
- **internal/** - All private application code, not importable by other projects
- **pkg/** - Reusable packages that could be extracted or shared
- Each package should have clear interfaces defined
- Tests should be co-located with implementation files (e.g., `client.go` and `client_test.go`)

## Interface-Driven Design

All major components are defined as interfaces to enable:
- Easy mocking for unit tests
- Swappable implementations
- Clear contracts between layers
- Independent module development

## Testing Organization

- Unit tests: `*_test.go` files alongside implementation
- Property tests: Use comment format `// Feature: binance-auto-trading, Property X: [description]`
- Integration tests: Separate `*_integration_test.go` files (use build tags if needed)
- Test data/mocks: `testdata/` subdirectories within each package

## Naming Conventions

- Interfaces: Descriptive names (e.g., `BinanceClient`, `TradingService`)
- Implementations: Often same name as interface or with suffix (e.g., `binanceClient`, `tradingService`)
- Test files: `<filename>_test.go`
- Property tests: Must reference property number in comments
