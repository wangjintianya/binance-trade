package service

import (
	"binance-trader/internal/api"
	"fmt"
	"sync"
	"time"
)

// MarketDataService defines the interface for market data operations
type MarketDataService interface {
	GetCurrentPrice(symbol string) (float64, error)
	GetHistoricalData(symbol string, interval string, limit int) ([]*api.Kline, error)
	SubscribeToPrice(symbol string, callback func(float64)) error
	GetVolume(symbol string, timeWindow time.Duration) (float64, error)
}

// priceCache represents a cached price entry
type priceCache struct {
	price     float64
	timestamp time.Time
}

// volumeCache represents a cached volume entry
type volumeCache struct {
	volume     float64
	timeWindow time.Duration
	timestamp  time.Time
}

// marketDataService implements MarketDataService interface
type marketDataService struct {
	client       api.BinanceClient
	priceCache   map[string]*priceCache
	volumeCache  map[string]*volumeCache
	cacheTTL     time.Duration
	cacheMutex   sync.RWMutex
}

// NewMarketDataService creates a new market data service
func NewMarketDataService(client api.BinanceClient, cacheTTL time.Duration) MarketDataService {
	if cacheTTL <= 0 {
		cacheTTL = 1 * time.Second // Default cache TTL
	}
	
	return &marketDataService{
		client:      client,
		priceCache:  make(map[string]*priceCache),
		volumeCache: make(map[string]*volumeCache),
		cacheTTL:    cacheTTL,
	}
}

// GetCurrentPrice retrieves the current price for a symbol with caching
func (s *marketDataService) GetCurrentPrice(symbol string) (float64, error) {
	if symbol == "" {
		return 0, fmt.Errorf("symbol cannot be empty")
	}
	
	// Check cache first
	s.cacheMutex.RLock()
	cached, exists := s.priceCache[symbol]
	s.cacheMutex.RUnlock()
	
	if exists && time.Since(cached.timestamp) < s.cacheTTL {
		return cached.price, nil
	}
	
	// Fetch from API
	priceData, err := s.client.GetPrice(symbol)
	if err != nil {
		return 0, fmt.Errorf("failed to get price for %s: %w", symbol, err)
	}
	
	if priceData.Price <= 0 {
		return 0, fmt.Errorf("invalid price received: %f", priceData.Price)
	}
	
	// Update cache
	s.cacheMutex.Lock()
	s.priceCache[symbol] = &priceCache{
		price:     priceData.Price,
		timestamp: time.Now(),
	}
	s.cacheMutex.Unlock()
	
	return priceData.Price, nil
}

// GetHistoricalData retrieves historical kline data for a symbol
func (s *marketDataService) GetHistoricalData(symbol string, interval string, limit int) ([]*api.Kline, error) {
	if symbol == "" {
		return nil, fmt.Errorf("symbol cannot be empty")
	}
	
	if interval == "" {
		return nil, fmt.Errorf("interval cannot be empty")
	}
	
	if limit <= 0 {
		return nil, fmt.Errorf("limit must be greater than 0")
	}
	
	klines, err := s.client.GetKlines(symbol, interval, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get historical data for %s: %w", symbol, err)
	}
	
	return klines, nil
}

// SubscribeToPrice subscribes to price updates for a symbol
func (s *marketDataService) SubscribeToPrice(symbol string, callback func(float64)) error {
	// This is a placeholder implementation
	// In a real implementation, this would use WebSocket connections
	return fmt.Errorf("price subscription not implemented yet")
}

// GetVolume retrieves the cumulative volume for a symbol within a time window
func (s *marketDataService) GetVolume(symbol string, timeWindow time.Duration) (float64, error) {
	if symbol == "" {
		return 0, fmt.Errorf("symbol cannot be empty")
	}
	
	if timeWindow <= 0 {
		return 0, fmt.Errorf("time window must be greater than 0")
	}
	
	// Create cache key based on symbol and time window
	cacheKey := fmt.Sprintf("%s_%d", symbol, timeWindow)
	
	// Check cache first
	s.cacheMutex.RLock()
	cached, exists := s.volumeCache[cacheKey]
	s.cacheMutex.RUnlock()
	
	if exists && time.Since(cached.timestamp) < s.cacheTTL && cached.timeWindow == timeWindow {
		return cached.volume, nil
	}
	
	// Calculate how many klines we need based on time window
	// Use 1-minute intervals for granular volume data
	interval := "1m"
	limit := int(timeWindow.Minutes())
	
	// Binance API has a maximum limit of 1000 klines
	if limit > 1000 {
		limit = 1000
	}
	
	// If time window is less than 1 minute, use at least 1 kline
	if limit < 1 {
		limit = 1
	}
	
	// Fetch klines from API
	klines, err := s.client.GetKlines(symbol, interval, limit)
	if err != nil {
		return 0, fmt.Errorf("failed to get klines for volume calculation: %w", err)
	}
	
	if len(klines) == 0 {
		return 0, fmt.Errorf("no kline data available for %s", symbol)
	}
	
	// Calculate cumulative volume within the time window
	now := time.Now().UnixMilli()
	windowStart := now - timeWindow.Milliseconds()
	
	var totalVolume float64
	for _, kline := range klines {
		// Only include klines within the time window
		if kline.OpenTime >= windowStart {
			totalVolume += kline.Volume
		}
	}
	
	// Update cache
	s.cacheMutex.Lock()
	s.volumeCache[cacheKey] = &volumeCache{
		volume:     totalVolume,
		timeWindow: timeWindow,
		timestamp:  time.Now(),
	}
	s.cacheMutex.Unlock()
	
	return totalVolume, nil
}
