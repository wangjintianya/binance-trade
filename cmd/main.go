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
	config      *config.Config
	logger      logger.Logger
	tradingType config.TradingType
	
	// Spot-specific components
	spotClient              api.BinanceClient
	spotTradingService      service.TradingService
	spotMarketService       service.MarketDataService
	spotOrderRepo           repository.OrderRepository
	spotRiskMgr             service.RiskManager
	spotConditionalOrderSvc service.ConditionalOrderService
	spotStopLossSvc         service.StopLossService
	
	// Futures-specific components
	futuresClient              api.FuturesClient
	futuresTradingService      service.FuturesTradingService
	futuresMarketService       service.FuturesMarketDataService
	futuresPositionManager     service.FuturesPositionManager
	futuresRiskManager         service.FuturesRiskManager
	futuresConditionalOrderSvc service.FuturesConditionalOrderService
	futuresStopLossSvc         service.FuturesStopLossService
	futuresFundingService      service.FuturesFundingService
	
	cli *cli.CLI
}

func main() {
	// Set up error recovery
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "Fatal error: %v\n", r)
			os.Exit(1)
		}
	}()

	// Determine trading type from command line arguments
	tradingType := config.TradingTypeSpot // Default to spot
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "spot":
			tradingType = config.TradingTypeSpot
		case "futures":
			tradingType = config.TradingTypeFutures
		default:
			fmt.Fprintf(os.Stderr, "Usage: %s [spot|futures]\n", os.Args[0])
			fmt.Fprintf(os.Stderr, "  spot    - Run spot trading system (default)\n")
			fmt.Fprintf(os.Stderr, "  futures - Run futures trading system\n")
			os.Exit(1)
		}
	}

	// Run application with context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Initialize application
	app, err := initializeApplication(tradingType)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize application: %v\n", err)
		os.Exit(1)
	}

	app.logger.Info("System initialized successfully", map[string]interface{}{
		"trading_type": string(tradingType),
	})

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
func initializeApplication(tradingType config.TradingType) (*Application, error) {
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
	log, err := initializeLogger(cfg, tradingType)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	log.Info("Starting Binance Auto-Trading System", map[string]interface{}{
		"config_file":  configPath,
		"trading_type": string(tradingType),
	})

	// Initialize application based on trading type
	app := &Application{
		config:      cfg,
		logger:      log,
		tradingType: tradingType,
	}

	switch tradingType {
	case config.TradingTypeSpot:
		if err := initializeSpotComponents(app, cfg, log); err != nil {
			return nil, fmt.Errorf("failed to initialize spot components: %w", err)
		}
	case config.TradingTypeFutures:
		if err := initializeFuturesComponents(app, cfg, log); err != nil {
			return nil, fmt.Errorf("failed to initialize futures components: %w", err)
		}
	default:
		return nil, fmt.Errorf("unknown trading type: %s", tradingType)
	}

	return app, nil
}

// initializeLogger creates and configures the logger based on trading type
func initializeLogger(cfg *config.Config, tradingType config.TradingType) (logger.Logger, error) {
	var logFile string
	
	switch tradingType {
	case config.TradingTypeSpot:
		if cfg.Logging.SpotFile != "" {
			logFile = cfg.Logging.SpotFile
		} else {
			logFile = cfg.Logging.File
		}
	case config.TradingTypeFutures:
		if cfg.Logging.FuturesFile != "" {
			logFile = cfg.Logging.FuturesFile
		} else {
			logFile = cfg.Logging.File
		}
	default:
		return nil, fmt.Errorf("unknown trading type: %s", tradingType)
	}

	loggerConfig := logger.Config{
		Level:         cfg.Logging.Level,
		FilePath:      logFile,
		MaxSizeMB:     int64(cfg.Logging.MaxSizeMB),
		MaxBackups:    cfg.Logging.MaxBackups,
		EnableConsole: true,
		TradingType:   string(tradingType),
	}
	
	return logger.NewLogger(loggerConfig)
}

// initializeSpotComponents initializes all spot trading components
func initializeSpotComponents(app *Application, cfg *config.Config, log logger.Logger) error {
	// Determine which config to use
	var binanceConfig *config.BinanceConfig
	if cfg.Spot != nil {
		binanceConfig = cfg.Spot
	} else {
		binanceConfig = &cfg.Binance
	}

	log.Info("Initializing spot trading components", map[string]interface{}{
		"base_url": binanceConfig.BaseURL,
		"testnet":  binanceConfig.Testnet,
	})

	// Initialize authentication manager
	authMgr, err := api.NewAuthManager(binanceConfig.APIKey, binanceConfig.APISecret)
	if err != nil {
		return fmt.Errorf("failed to initialize auth manager: %w", err)
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

	// Initialize Binance spot client
	spotClient, err := api.NewSpotClient(binanceConfig.BaseURL, httpClient, authMgr)
	if err != nil {
		return fmt.Errorf("failed to initialize spot client: %w", err)
	}
	app.spotClient = spotClient

	// Initialize order repository
	app.spotOrderRepo = repository.NewMemoryOrderRepository()

	// Initialize risk manager
	riskLimits := &service.RiskLimits{
		MaxOrderAmount:    cfg.Risk.MaxOrderAmount,
		MaxDailyOrders:    cfg.Risk.MaxDailyOrders,
		MinBalanceReserve: cfg.Risk.MinBalanceReserve,
		MaxAPICallsPerMin: cfg.Risk.MaxAPICallsPerMin,
	}
	app.spotRiskMgr = service.NewRiskManager(riskLimits, spotClient)

	// Initialize trading service
	app.spotTradingService = service.NewSpotTradingService(spotClient, app.spotRiskMgr, app.spotOrderRepo, log)

	// Initialize market data service
	app.spotMarketService = service.NewMarketDataService(spotClient, 1*time.Second)

	// Initialize conditional order repository
	conditionalOrderRepo := repository.NewMemoryConditionalOrderRepository()

	// Initialize stop order repository
	stopOrderRepo := repository.NewMemoryStopOrderRepository()

	// Initialize trigger engine
	triggerEngine := service.NewTriggerEngine()

	// Initialize stop loss service
	app.spotStopLossSvc = service.NewStopLossService(
		stopOrderRepo,
		triggerEngine,
		app.spotTradingService,
		app.spotMarketService,
		log,
	)

	// Initialize conditional order service
	app.spotConditionalOrderSvc = service.NewConditionalOrderService(
		conditionalOrderRepo,
		stopOrderRepo,
		triggerEngine,
		app.spotTradingService,
		app.spotMarketService,
		app.spotStopLossSvc,
		log,
	)

	// Initialize CLI
	app.cli = cli.NewCLI(app.spotTradingService, app.spotMarketService, app.spotConditionalOrderSvc, app.spotStopLossSvc, log)

	log.Info("Spot trading components initialized successfully", nil)
	return nil
}

// initializeFuturesComponents initializes all futures trading components
func initializeFuturesComponents(app *Application, cfg *config.Config, log logger.Logger) error {
	// Futures config must be present
	if cfg.Futures == nil {
		return fmt.Errorf("futures configuration not found in config file")
	}

	log.Info("Initializing futures trading components", map[string]interface{}{
		"base_url": cfg.Futures.BaseURL,
		"testnet":  cfg.Futures.Testnet,
	})

	// Initialize authentication manager
	authMgr, err := api.NewAuthManager(cfg.Futures.APIKey, cfg.Futures.APISecret)
	if err != nil {
		return fmt.Errorf("failed to initialize auth manager: %w", err)
	}

	// Initialize rate limiter
	rateLimiter := api.NewRateLimiter(cfg.Futures.Risk.MaxAPICallsPerMin)

	// Initialize HTTP client with retry configuration
	retryConfig := api.RetryConfig{
		MaxAttempts:       cfg.Retry.MaxAttempts,
		InitialDelayMs:    cfg.Retry.InitialDelayMs,
		BackoffMultiplier: cfg.Retry.BackoffMultiplier,
	}
	httpClient := api.NewHTTPClient(rateLimiter, retryConfig)

	// Initialize Binance futures client
	futuresClient, err := api.NewFuturesClient(cfg.Futures.BaseURL, httpClient, authMgr)
	if err != nil {
		return fmt.Errorf("failed to initialize futures client: %w", err)
	}
	app.futuresClient = futuresClient

	// Initialize futures order repository
	futuresOrderRepo := repository.NewMemoryFuturesOrderRepository()

	// Initialize futures position repository
	futuresPositionRepo := repository.NewMemoryFuturesPositionRepository()

	// Initialize futures market data service
	app.futuresMarketService = service.NewFuturesMarketDataService(futuresClient, log)

	// Initialize futures position manager
	app.futuresPositionManager = service.NewFuturesPositionManager(
		futuresClient,
		futuresPositionRepo,
		log,
	)

	// Initialize futures risk manager
	app.futuresRiskManager = service.NewFuturesRiskManager(
		&cfg.Futures.Risk,
		futuresClient,
		app.futuresPositionManager,
		log,
	)

	// Initialize futures trading service
	app.futuresTradingService = service.NewFuturesTradingService(
		futuresClient,
		futuresOrderRepo,
		log,
	)

	// Initialize trigger engine
	triggerEngine := service.NewTriggerEngine()

	// Initialize stop order repository
	stopOrderRepo := repository.NewMemoryStopOrderRepository()

	// Initialize futures stop loss service
	app.futuresStopLossSvc = service.NewFuturesStopLossService(
		stopOrderRepo,
		triggerEngine,
		app.futuresTradingService,
		app.futuresMarketService,
		log,
	)

	// Initialize futures conditional order service
	app.futuresConditionalOrderSvc = service.NewFuturesConditionalOrderService(
		futuresClient,
		app.futuresMarketService,
		app.futuresPositionManager,
		app.futuresTradingService,
		log,
	)

	// Initialize futures funding service
	app.futuresFundingService = service.NewFuturesFundingService(
		app.futuresMarketService,
		log,
	)

	// Initialize CLI (futures CLI would need to be implemented separately)
	// For now, we'll use a placeholder or the spot CLI
	// TODO: Implement futures-specific CLI
	app.cli = nil // Futures CLI not yet implemented

	log.Info("Futures trading components initialized successfully", nil)
	return nil
}

// run starts the application
func (app *Application) run(ctx context.Context) error {
	// Set up panic recovery
	defer func() {
		if r := recover(); r != nil {
			app.logger.Error("Panic recovered in application run", map[string]interface{}{
				"panic": fmt.Sprintf("%v", r),
			})
		}
	}()

	switch app.tradingType {
	case config.TradingTypeSpot:
		return app.runSpot(ctx)
	case config.TradingTypeFutures:
		return app.runFutures(ctx)
	default:
		return fmt.Errorf("unknown trading type: %s", app.tradingType)
	}
}

// runSpot runs the spot trading application
func (app *Application) runSpot(ctx context.Context) error {
	// Start monitoring engine for conditional orders
	app.logger.Info("Starting spot conditional order monitoring", nil)
	if err := app.spotConditionalOrderSvc.StartMonitoring(); err != nil {
		return fmt.Errorf("failed to start conditional order monitoring: %w", err)
	}

	// Run CLI
	if app.cli != nil {
		if err := app.cli.Run(); err != nil {
			return fmt.Errorf("CLI error: %w", err)
		}
	}

	return nil
}

// runFutures runs the futures trading application
func (app *Application) runFutures(ctx context.Context) error {
	app.logger.Info("Starting futures trading system", nil)
	
	// Start monitoring for futures conditional orders
	if app.futuresConditionalOrderSvc != nil {
		app.logger.Info("Starting futures conditional order monitoring", nil)
		if err := app.futuresConditionalOrderSvc.StartMonitoring(); err != nil {
			return fmt.Errorf("failed to start futures conditional order monitoring: %w", err)
		}
	}

	// Start funding rate monitoring
	if app.futuresFundingService != nil {
		app.logger.Info("Starting funding rate monitoring", nil)
		checkInterval := time.Duration(app.config.Futures.Monitoring.FundingRateCheckIntervalMs) * time.Millisecond
		if err := app.futuresFundingService.StartMonitoring(checkInterval); err != nil {
			return fmt.Errorf("failed to start funding rate monitoring: %w", err)
		}
	}

	// Run CLI if available
	if app.cli != nil {
		if err := app.cli.Run(); err != nil {
			return fmt.Errorf("CLI error: %w", err)
		}
	} else {
		// If no CLI, just wait for context cancellation
		app.logger.Info("Futures system running (CLI not implemented yet)", nil)
		<-ctx.Done()
	}

	return nil
}

// shutdown performs graceful shutdown of all components
func (app *Application) shutdown(ctx context.Context) error {
	app.logger.Info("Starting graceful shutdown", nil)

	// Create a channel to signal completion
	done := make(chan error, 1)

	go func() {
		var shutdownErr error

		switch app.tradingType {
		case config.TradingTypeSpot:
			shutdownErr = app.shutdownSpot()
		case config.TradingTypeFutures:
			shutdownErr = app.shutdownFutures()
		}

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

// shutdownSpot performs graceful shutdown of spot components
func (app *Application) shutdownSpot() error {
	app.logger.Info("Shutdown: Stopping spot conditional order monitoring", nil)
	
	if app.spotConditionalOrderSvc != nil {
		if err := app.spotConditionalOrderSvc.StopMonitoring(); err != nil {
			if err.Error() == "monitoring engine not running" {
				app.logger.Debug("Spot monitoring was not running during shutdown", nil)
			} else {
				app.logger.Error("Error stopping spot conditional order monitoring", map[string]interface{}{
					"error": err.Error(),
				})
				return err
			}
		}
	}

	return nil
}

// shutdownFutures performs graceful shutdown of futures components
func (app *Application) shutdownFutures() error {
	app.logger.Info("Shutdown: Stopping futures services", nil)
	
	// Stop conditional order monitoring
	if app.futuresConditionalOrderSvc != nil {
		if err := app.futuresConditionalOrderSvc.StopMonitoring(); err != nil {
			if err.Error() != "monitoring engine not running" {
				app.logger.Error("Error stopping futures conditional order monitoring", map[string]interface{}{
					"error": err.Error(),
				})
				return err
			}
		}
	}

	// Stop funding rate monitoring
	if app.futuresFundingService != nil {
		if err := app.futuresFundingService.StopMonitoring(); err != nil {
			if err.Error() != "monitoring not running" {
				app.logger.Error("Error stopping funding rate monitoring", map[string]interface{}{
					"error": err.Error(),
				})
				return err
			}
		}
	}

	return nil
}
