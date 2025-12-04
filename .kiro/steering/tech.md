# Technology Stack

## Language & Runtime

- **Go 1.21+** - Primary programming language

## Key Dependencies

- `github.com/adshao/go-binance/v2` - Binance Go SDK for API integration
- `gopkg.in/yaml.v3` - YAML configuration parsing
- `github.com/sirupsen/logrus` - Structured logging
- `github.com/leanovate/gopter` - Property-based testing framework

## Testing Framework

- Go standard `testing` package for unit tests
- `gopter` for property-based testing (minimum 100 iterations per property)
- `httptest` for mocking HTTP responses
- Target coverage: 80%+ overall, 90%+ for core business logic

## Configuration

- YAML-based configuration files
- Environment variables for sensitive data (API keys)
- Configuration file location: `config.yaml` (customizable via `CONFIG_FILE` env var)

## Common Commands

```bash
# Initialize module
go mod init binance-trader

# Install dependencies
go mod tidy

# Build
go build -o binance-trader cmd/main.go

# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific test
go test -v ./internal/...

# Build for production
go build -ldflags="-s -w" -o binance-trader cmd/main.go
```

## Environment Variables

- `BINANCE_API_KEY` - Binance API key (required)
- `BINANCE_API_SECRET` - Binance API secret (required)
- `CONFIG_FILE` - Path to config file (default: config.yaml)
- `LOG_LEVEL` - Logging level (default: info)

## Security Requirements

- All API requests must use HTTPS
- HMAC SHA256 signing for all private API requests
- Never hardcode credentials in source code
- Mask sensitive information in logs (show only first 4 and last 4 characters)
