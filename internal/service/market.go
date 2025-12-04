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
}

// priceCache represents a cached price entry
type priceCache struct {
	price     float64
	timestamp time.Time
}

// marketDataService implements MarketDataService interface
type marketDataService struct {
	client       api.BinanceClient
	priceCache   map[string]*priceCache
	cacheTTL     time.Duration
	cacheMutex   sync.RWMutex
}

// NewMarketDataService creates a new market data service
func NewMarketDataService(client api.BinanceClient, cacheTTL time.Duration) MarketDataService {
	if cacheTTL <= 0 {
		cacheTTL = 1 * time.Second // Default cache TTL
	}
	
	return &marketDataService{
		client:     client,
		priceCache: make(map[string]*priceCache),
		cacheTTL:   cacheTTL,
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
