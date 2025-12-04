package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"binance-trader/internal/api"
	"binance-trader/internal/cli"
	"binance-trader/internal/config"
	"binance-trader/internal/repository"
	"binance-trader/internal/service"
	"binance-trader/pkg/logger"
)

// Application holds all application dependencies
type Application struct {
	config         *config.Config
	logger         logger.Logger
	binanceClient  api.BinanceClient
	tradingService service.TradingService
	marketService  service.MarketDataService
	orderRepo      repository.OrderRepository
	riskMgr        service.RiskManager
	cli            *cli.CLI
}

func main() {
	// Set up error recovery
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "Fatal error: %v\n", r)
			os.Exit(1)
		}
	}()

	// Run application with context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Initialize application
	app, err := initializeApplication()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize application: %v\n", err)
		os.Exit(1)
	}

	app.logger.Info("System initialized successfully", nil)

	// Run application in goroutine
	errChan := make(chan error, 1)
	go func() {
		errChan <- app.run(ctx)
	}()

	// Wait for shutdown signal or error
	select {
	case sig := <-sigChan:
		app.logger.Info("Received shutdown signal", map[string]interface{}{
			"signal": sig.String(),
		})
		cancel()
		
		// Wait for graceful shutdown with timeout
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()
		
		if err := app.shutdown(shutdownCtx); err != nil {
			app.logger.Error("Error during shutdown", map[string]interface{}{
				"error": err.Error(),
			})
			os.Exit(1)
		}
		
	case err := <-errChan:
		if err != nil {
			app.logger.Fatal("Application error", map[string]interface{}{
				"error": err.Error(),
			})
			os.Exit(1)
		}
	}

	app.logger.Info("System shutdown complete", nil)
}

// initializeApplication initializes all application components with dependency injection
func initializeApplication() (*Application, error) {
	// Get config file path from environment or use default
	configPath := os.Getenv("CONFIG_FILE")
	if configPath == "" {
		configPath = "config.yaml"
	}

	// Load configuration
	configMgr := config.NewConfigManager()
	cfg, err := configMgr.Load(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Initialize logger
	loggerConfig := logger.Config{
		Level:         cfg.Logging.Level,
		FilePath:      cfg.Logging.File,
		MaxSizeMB:     int64(cfg.Logging.MaxSizeMB),
		MaxBackups:    cfg.Logging.MaxBackups,
		EnableConsole: true,
	}
	log, err := logger.NewLogger(loggerConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	log.Info("Starting Binance Auto-Trading System", map[string]interface{}{
		"config_file": configPath,
		"base_url":    cfg.Binance.BaseURL,
		"testnet":     cfg.Binance.Testnet,
	})

	// Initialize authentication manager
	authMgr, err := api.NewAuthManager(cfg.Binance.APIKey, cfg.Binance.APISecret)
	if err != nil {
		log.Error("Failed to initialize auth manager", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to initialize auth manager: %w", err)
	}

	// Initialize rate limiter
	rateLimiter := api.NewRateLimiter(cfg.Risk.MaxAPICallsPerMin)

	// Initialize HTTP client with retry configuration
	retryConfig := api.RetryConfig{
		MaxAttempts:       cfg.Retry.MaxAttempts,
		InitialDelayMs:    cfg.Retry.InitialDelayMs,
		BackoffMultiplier: cfg.Retry.BackoffMultiplier,
	}
	httpClient := api.NewHTTPClient(rateLimiter, retryConfig)

	// Initialize Binance API client
	binanceClient, err := api.NewBinanceClient(cfg.Binance.BaseURL, httpClient, authMgr)
	if err != nil {
		log.Error("Failed to initialize Binance client", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to initialize Binance client: %w", err)
	}

	// Initialize order repository
	orderRepo := repository.NewMemoryOrderRepository()

	// Initialize risk manager
	riskLimits := &service.RiskLimits{
		MaxOrderAmount:    cfg.Risk.MaxOrderAmount,
		MaxDailyOrders:    cfg.Risk.MaxDailyOrders,
		MinBalanceReserve: cfg.Risk.MinBalanceReserve,
		MaxAPICallsPerMin: cfg.Risk.MaxAPICallsPerMin,
	}
	riskMgr := service.NewRiskManager(riskLimits, binanceClient)

	// Initialize trading service
	tradingService := service.NewTradingService(binanceClient, riskMgr, orderRepo, log)

	// Initialize market data service with 1 second cache TTL
	marketService := service.NewMarketDataService(binanceClient, 1*time.Second)

	// Initialize CLI
	cliApp := cli.NewCLI(tradingService, marketService, log)

	return &Application{
		config:         cfg,
		logger:         log,
		binanceClient:  binanceClient,
		tradingService: tradingService,
		marketService:  marketService,
		orderRepo:      orderRepo,
		riskMgr:        riskMgr,
		cli:            cliApp,
	}, nil
}

// run starts the application
func (app *Application) run(ctx context.Context) error {
	// Set up panic recovery for the CLI
	defer func() {
		if r := recover(); r != nil {
			app.logger.Error("Panic recovered in application run", map[string]interface{}{
				"panic": fmt.Sprintf("%v", r),
			})
		}
	}()

	// Run CLI
	if err := app.cli.Run(); err != nil {
		return fmt.Errorf("CLI error: %w", err)
	}

	return nil
}

// shutdown performs graceful shutdown of all components
func (app *Application) shutdown(ctx context.Context) error {
	app.logger.Info("Starting graceful shutdown", nil)

	// Create a channel to signal completion
	done := make(chan error, 1)

	go func() {
		// Perform cleanup operations
		var shutdownErr error

		// Log final statistics
		app.logger.Info("Shutdown: Logging final statistics", nil)

		// Close any open resources
		// Note: In this implementation, most resources are automatically cleaned up
		// In a production system, you might need to:
		// - Close database connections
		// - Flush pending logs
		// - Cancel pending API requests
		// - Save state to disk

		app.logger.Info("Shutdown: All resources cleaned up", nil)

		done <- shutdownErr
	}()

	// Wait for shutdown to complete or timeout
	select {
	case err := <-done:
		if err != nil {
			return fmt.Errorf("error during shutdown: %w", err)
		}
		app.logger.Info("Graceful shutdown completed", nil)
		return nil
	case <-ctx.Done():
		return fmt.Errorf("shutdown timeout exceeded")
	}
}
