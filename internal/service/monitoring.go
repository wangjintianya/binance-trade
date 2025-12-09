package service

import (
	"binance-trader/internal/api"
	"binance-trader/internal/repository"
	"binance-trader/pkg/logger"
	"fmt"
	"sync"
	"time"
)

// MonitoringEngine manages the monitoring of conditional orders and trailing stops
type MonitoringEngine struct {
	repo              repository.ConditionalOrderRepository
	stopOrderRepo     repository.StopOrderRepository
	triggerEngine     TriggerEngine
	tradingService    TradingService
	marketDataService MarketDataService
	stopLossService   StopLossService
	logger            logger.Logger
	
	// Monitoring state
	mu              sync.RWMutex
	activeOrders    map[string]*repository.ConditionalOrder
	marketDataCache map[string]*MarketData
	
	// Configuration
	updateInterval time.Duration
	
	// Control channels
	stopChan chan struct{}
	doneChan chan struct{}
	
	// Status
	isRunning bool
}

// MonitoringEngineConfig holds configuration for the monitoring engine
type MonitoringEngineConfig struct {
	UpdateInterval time.Duration
}

// NewMonitoringEngine creates a new monitoring engine instance
func NewMonitoringEngine(
	repo repository.ConditionalOrderRepository,
	stopOrderRepo repository.StopOrderRepository,
	triggerEngine TriggerEngine,
	tradingService TradingService,
	marketDataService MarketDataService,
	stopLossService StopLossService,
	logger logger.Logger,
	config *MonitoringEngineConfig,
) *MonitoringEngine {
	if config == nil {
		config = &MonitoringEngineConfig{
			UpdateInterval: 1 * time.Second,
		}
	}
	
	if config.UpdateInterval <= 0 {
		config.UpdateInterval = 1 * time.Second
	}
	
	return &MonitoringEngine{
		repo:              repo,
		stopOrderRepo:     stopOrderRepo,
		triggerEngine:     triggerEngine,
		tradingService:    tradingService,
		marketDataService: marketDataService,
		stopLossService:   stopLossService,
		logger:            logger,
		activeOrders:      make(map[string]*repository.ConditionalOrder),
		marketDataCache:   make(map[string]*MarketData),
		updateInterval:    config.UpdateInterval,
		stopChan:          make(chan struct{}),
		doneChan:          make(chan struct{}),
		isRunning:         false,
	}
}

// Start starts the monitoring engine
func (me *MonitoringEngine) Start() error {
	me.mu.Lock()
	defer me.mu.Unlock()
	
	if me.isRunning {
		return fmt.Errorf("monitoring engine already running")
	}
	
	// Load active orders from repository
	if err := me.loadActiveOrders(); err != nil {
		return fmt.Errorf("failed to load active orders: %w", err)
	}
	
	me.isRunning = true
	me.stopChan = make(chan struct{})
	me.doneChan = make(chan struct{})
	
	// Start monitoring goroutine
	go me.monitoringLoop()
	
	me.logger.Info("Monitoring engine started", map[string]interface{}{
		"update_interval": me.updateInterval.String(),
		"active_orders":   len(me.activeOrders),
	})
	
	return nil
}

// Stop stops the monitoring engine gracefully
func (me *MonitoringEngine) Stop() error {
	me.mu.Lock()
	
	if !me.isRunning {
		me.mu.Unlock()
		return fmt.Errorf("monitoring engine not running")
	}
	
	me.isRunning = false
	me.mu.Unlock()
	
	// Signal stop
	close(me.stopChan)
	
	// Wait for monitoring loop to finish
	<-me.doneChan
	
	me.logger.Info("Monitoring engine stopped", nil)
	
	return nil
}

// IsRunning returns whether the monitoring engine is currently running
func (me *MonitoringEngine) IsRunning() bool {
	me.mu.RLock()
	defer me.mu.RUnlock()
	return me.isRunning
}

// loadActiveOrders loads all active orders from the repository
func (me *MonitoringEngine) loadActiveOrders() error {
	orders, err := me.repo.FindActiveOrders()
	if err != nil {
		return err
	}
	
	me.activeOrders = make(map[string]*repository.ConditionalOrder)
	for _, order := range orders {
		me.activeOrders[order.OrderID] = order
	}
	
	return nil
}

// monitoringLoop is the main monitoring loop that runs in a goroutine
func (me *MonitoringEngine) monitoringLoop() {
	defer close(me.doneChan)
	
	ticker := time.NewTicker(me.updateInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-me.stopChan:
			me.logger.Info("Monitoring loop received stop signal", nil)
			return
			
		case <-ticker.C:
			me.checkAndTriggerOrders()
		}
	}
}

// checkAndTriggerOrders checks all active orders and triggers them if conditions are met
func (me *MonitoringEngine) checkAndTriggerOrders() {
	// Reload active orders to catch any new orders
	me.mu.Lock()
	if err := me.loadActiveOrders(); err != nil {
		me.logger.Error("Failed to reload active orders", map[string]interface{}{
			"error": err.Error(),
		})
		me.mu.Unlock()
		return
	}
	
	// Create a copy of active orders to avoid holding lock during processing
	ordersCopy := make([]*repository.ConditionalOrder, 0, len(me.activeOrders))
	for _, order := range me.activeOrders {
		ordersCopy = append(ordersCopy, order)
	}
	me.mu.Unlock()
	
	// Process each conditional order
	for _, order := range ordersCopy {
		me.processOrder(order)
	}
	
	// Process trailing stop orders
	me.processTrailingStopOrders()
}

// extractValueFromMarketData extracts the appropriate value from market data based on trigger type
func (me *MonitoringEngine) extractValueFromMarketData(marketData *MarketData, condition *repository.TriggerCondition) float64 {
	switch condition.Type {
	case repository.TriggerTypePrice:
		return marketData.Price
	case repository.TriggerTypePriceChangePercent:
		if condition.BasePrice > 0 {
			return ((marketData.Price - condition.BasePrice) / condition.BasePrice) * 100.0
		}
		return 0
	case repository.TriggerTypeVolume:
		return marketData.Volume24h
	default:
		return 0
	}
}

// processOrder processes a single conditional order
func (me *MonitoringEngine) processOrder(order *repository.ConditionalOrder) {
	// Check time window if specified
	if order.TimeWindow != nil {
		currentTime := time.Now().Unix()
		tw := &TimeWindow{
			StartTime: order.TimeWindow.StartTime,
			EndTime:   order.TimeWindow.EndTime,
		}
		if !IsWithinTimeWindow(currentTime, tw) {
			return
		}
	}
	
	// Get market data
	marketData, err := me.getMarketData(order.Symbol)
	if err != nil {
		me.logger.Warn("Failed to get market data", map[string]interface{}{
			"order_id": order.OrderID,
			"symbol":   order.Symbol,
			"error":    err.Error(),
		})
		return
	}
	
	// Evaluate trigger condition
	triggerCond := me.convertToServiceTriggerCondition(order.TriggerCondition)
	currentValue := me.extractValueFromMarketData(marketData, order.TriggerCondition)
	triggered, err := me.triggerEngine.EvaluateCondition(triggerCond, currentValue)
	if err != nil {
		me.logger.Warn("Failed to evaluate trigger condition", map[string]interface{}{
			"order_id": order.OrderID,
			"error":    err.Error(),
		})
		return
	}
	
	if triggered {
		me.executeTrigger(order, marketData)
	}
}

// getMarketData retrieves market data for a symbol with caching
func (me *MonitoringEngine) getMarketData(symbol string) (*MarketData, error) {
	// Check cache first
	me.mu.RLock()
	cached, exists := me.marketDataCache[symbol]
	me.mu.RUnlock()
	
	if exists && time.Since(time.Unix(cached.Timestamp, 0)) < 1*time.Second {
		return cached, nil
	}
	
	// Fetch from market data service
	price, err := me.marketDataService.GetCurrentPrice(symbol)
	if err != nil {
		return nil, err
	}
	
	marketData := &MarketData{
		Symbol:    symbol,
		Price:     price,
		Timestamp: time.Now().Unix(),
	}
	
	// Update cache
	me.mu.Lock()
	me.marketDataCache[symbol] = marketData
	me.mu.Unlock()
	
	return marketData, nil
}

// executeTrigger executes a triggered conditional order
func (me *MonitoringEngine) executeTrigger(order *repository.ConditionalOrder, marketData *MarketData) {
	// Log trigger event with complete information
	triggerInfo := me.buildTriggerLogInfo(order, marketData)
	me.logger.Info("Trigger condition met, executing order", triggerInfo)
	
	// Update status to triggered
	triggeredAt := time.Now().Unix()
	if err := me.repo.UpdateStatus(order.OrderID, repository.ConditionalOrderStatusTriggered, triggeredAt, 0); err != nil {
		me.logger.Error("Failed to update order status to triggered", map[string]interface{}{
			"order_id": order.OrderID,
			"error":    err.Error(),
		})
		return
	}
	
	// Execute order via trading service
	executedOrder, err := me.executeOrder(order)
	if err != nil {
		me.logger.Error("Failed to execute conditional order", map[string]interface{}{
			"order_id": order.OrderID,
			"symbol":   order.Symbol,
			"error":    err.Error(),
		})
		return
	}
	
	// Update status to executed
	if err := me.repo.UpdateStatus(order.OrderID, repository.ConditionalOrderStatusExecuted, triggeredAt, executedOrder.OrderID); err != nil {
		me.logger.Error("Failed to update order status to executed", map[string]interface{}{
			"order_id": order.OrderID,
			"error":    err.Error(),
		})
		return
	}
	
	// Remove from active orders
	me.mu.Lock()
	delete(me.activeOrders, order.OrderID)
	me.mu.Unlock()
	
	// Unregister from trigger engine
	if err := me.triggerEngine.UnregisterCondition(order.OrderID); err != nil {
		me.logger.Warn("Failed to unregister condition from trigger engine", map[string]interface{}{
			"order_id": order.OrderID,
			"error":    err.Error(),
		})
	}
	
	me.logger.Info("Conditional order executed successfully", map[string]interface{}{
		"order_id":          order.OrderID,
		"executed_order_id": executedOrder.OrderID,
		"trigger_price":     marketData.Price,
	})
}

// executeOrder executes the actual order through the trading service
func (me *MonitoringEngine) executeOrder(order *repository.ConditionalOrder) (*api.Order, error) {
	// Execute based on order type and side
	var executedOrder *api.Order
	var err error
	
	switch {
	case order.Type == api.OrderTypeMarket && order.Side == api.OrderSideBuy:
		executedOrder, err = me.tradingService.PlaceMarketBuyOrder(order.Symbol, order.Quantity)
		if err != nil {
			return nil, err
		}
		
	case order.Type == api.OrderTypeLimit && order.Side == api.OrderSideSell:
		executedOrder, err = me.tradingService.PlaceLimitSellOrder(order.Symbol, order.Price, order.Quantity)
		if err != nil {
			return nil, err
		}
		
	default:
		return nil, fmt.Errorf("unsupported order type/side combination: %s/%s", order.Type, order.Side)
	}
	
	return executedOrder, err
}

// buildTriggerLogInfo builds comprehensive log information for trigger events
func (me *MonitoringEngine) buildTriggerLogInfo(order *repository.ConditionalOrder, marketData *MarketData) map[string]interface{} {
	logInfo := map[string]interface{}{
		"order_id":      order.OrderID,
		"symbol":        order.Symbol,
		"trigger_price": marketData.Price,
		"trigger_time":  time.Now().Unix(),
	}
	
	// Add trigger condition details
	if order.TriggerCondition != nil {
		me.addTriggerConditionToLog(logInfo, order.TriggerCondition, marketData)
	}
	
	return logInfo
}

// addTriggerConditionToLog adds trigger condition details to log info
func (me *MonitoringEngine) addTriggerConditionToLog(logInfo map[string]interface{}, condition *repository.TriggerCondition, marketData *MarketData) {
	// Handle composite conditions
	if len(condition.SubConditions) > 0 {
		logInfo["condition_type"] = "composite"
		logInfo["composite_operator"] = me.logicOperatorToString(condition.CompositeType)
		
		// Add details of satisfied sub-conditions
		satisfiedConditions := make([]map[string]interface{}, 0)
		for i, subCond := range condition.SubConditions {
			subCondInfo := make(map[string]interface{})
			subCondInfo["index"] = i
			me.addSimpleConditionToLog(subCondInfo, subCond, marketData)
			
			// Check if this sub-condition is satisfied
			serviceCond := me.convertToServiceTriggerCondition(subCond)
			currentValue := me.extractValueFromMarketData(marketData, subCond)
			satisfied, _ := me.triggerEngine.EvaluateCondition(serviceCond, currentValue)
			subCondInfo["satisfied"] = satisfied
			
			if satisfied {
				satisfiedConditions = append(satisfiedConditions, subCondInfo)
			}
		}
		logInfo["satisfied_conditions"] = satisfiedConditions
	} else {
		// Simple condition
		me.addSimpleConditionToLog(logInfo, condition, marketData)
	}
}

// addSimpleConditionToLog adds simple condition details to log info
func (me *MonitoringEngine) addSimpleConditionToLog(logInfo map[string]interface{}, condition *repository.TriggerCondition, marketData *MarketData) {
	logInfo["trigger_type"] = me.triggerTypeToString(condition.Type)
	logInfo["operator"] = me.operatorToString(condition.Operator)
	logInfo["trigger_value"] = condition.Value
	
	switch condition.Type {
	case repository.TriggerTypePrice:
		logInfo["current_price"] = marketData.Price
		
	case repository.TriggerTypePriceChangePercent:
		logInfo["base_price"] = condition.BasePrice
		logInfo["current_price"] = marketData.Price
		if condition.BasePrice > 0 {
			changePercent := ((marketData.Price - condition.BasePrice) / condition.BasePrice) * 100.0
			logInfo["price_change_percent"] = changePercent
		}
		
	case repository.TriggerTypeVolume:
		logInfo["current_volume"] = marketData.Volume24h
		if condition.TimeWindow > 0 {
			logInfo["time_window"] = condition.TimeWindow.String()
		}
	}
}

// Helper functions to convert enums to strings for logging
func (me *MonitoringEngine) triggerTypeToString(t repository.TriggerType) string {
	switch t {
	case repository.TriggerTypePrice:
		return "price"
	case repository.TriggerTypePriceChangePercent:
		return "price_change_percent"
	case repository.TriggerTypeVolume:
		return "volume"
	default:
		return "unknown"
	}
}

func (me *MonitoringEngine) operatorToString(op repository.ComparisonOperator) string {
	switch op {
	case repository.OperatorGreaterThan:
		return ">"
	case repository.OperatorLessThan:
		return "<"
	case repository.OperatorGreaterEqual:
		return ">="
	case repository.OperatorLessEqual:
		return "<="
	default:
		return "unknown"
	}
}

func (me *MonitoringEngine) logicOperatorToString(op repository.LogicOperator) string {
	switch op {
	case repository.LogicAND:
		return "AND"
	case repository.LogicOR:
		return "OR"
	default:
		return "unknown"
	}
}

// convertToServiceTriggerCondition converts repository trigger condition to service trigger condition
func (me *MonitoringEngine) convertToServiceTriggerCondition(repoCond *repository.TriggerCondition) *TriggerCondition {
	if repoCond == nil {
		return nil
	}
	
	serviceCond := &TriggerCondition{
		Type:          TriggerType(repoCond.Type),
		Operator:      ComparisonOperator(repoCond.Operator),
		Value:         repoCond.Value,
		BasePrice:     repoCond.BasePrice,
		TimeWindow:    repoCond.TimeWindow,
		CompositeType: LogicOperator(repoCond.CompositeType),
	}
	
	// Convert sub-conditions recursively
	if len(repoCond.SubConditions) > 0 {
		serviceCond.SubConditions = make([]*TriggerCondition, len(repoCond.SubConditions))
		for i, subCond := range repoCond.SubConditions {
			serviceCond.SubConditions[i] = me.convertToServiceTriggerCondition(subCond)
		}
	}
	
	return serviceCond
}

// processTrailingStopOrders processes all active trailing stop orders
func (me *MonitoringEngine) processTrailingStopOrders() {
	// Get all active trailing stop orders
	// We need to get all symbols first, so let's get all trailing stops
	// Since we don't have a method to get all active trailing stops across all symbols,
	// we'll need to track symbols with active trailing stops
	// For now, let's get trailing stops for symbols we're already monitoring
	
	// Collect unique symbols from market data cache
	me.mu.RLock()
	symbols := make([]string, 0, len(me.marketDataCache))
	for symbol := range me.marketDataCache {
		symbols = append(symbols, symbol)
	}
	me.mu.RUnlock()
	
	// Process trailing stops for each symbol
	for _, symbol := range symbols {
		me.processTrailingStopsForSymbol(symbol)
	}
}

// processTrailingStopsForSymbol processes trailing stop orders for a specific symbol
func (me *MonitoringEngine) processTrailingStopsForSymbol(symbol string) {
	// Get active trailing stop orders for this symbol
	trailingOrders, err := me.stopOrderRepo.FindActiveTrailingStopOrders(symbol)
	if err != nil {
		me.logger.Warn("Failed to get active trailing stop orders", map[string]interface{}{
			"symbol": symbol,
			"error":  err.Error(),
		})
		return
	}
	
	if len(trailingOrders) == 0 {
		return
	}
	
	// Get current market price
	marketData, err := me.getMarketData(symbol)
	if err != nil {
		me.logger.Warn("Failed to get market data for trailing stops", map[string]interface{}{
			"symbol": symbol,
			"error":  err.Error(),
		})
		return
	}
	
	// Update each trailing stop order
	for _, order := range trailingOrders {
		// Use the stop loss service to update the trailing stop price
		if sls, ok := me.stopLossService.(*stopLossService); ok {
			triggered, err := sls.UpdateTrailingStopPrice(order.OrderID, marketData.Price)
			if err != nil {
				me.logger.Warn("Failed to update trailing stop price", map[string]interface{}{
					"order_id": order.OrderID,
					"symbol":   symbol,
					"error":    err.Error(),
				})
				continue
			}
			
			if triggered {
				me.logger.Info("Trailing stop order triggered", map[string]interface{}{
					"order_id":      order.OrderID,
					"symbol":        symbol,
					"trigger_price": marketData.Price,
				})
			}
		}
	}
}
