package service

import (
	"binance-trader/internal/api"
	"binance-trader/pkg/logger"
	"fmt"
	"sync"
	"time"
)

// FuturesMarketDataService defines the interface for futures market data operations
type FuturesMarketDataService interface {
	GetMarkPrice(symbol string) (float64, error)
	GetLastPrice(symbol string) (float64, error)
	GetHistoricalData(symbol string, interval string, limit int) ([]*api.Kline, error)
	GetFundingRate(symbol string) (*api.FundingRate, error)
	GetFundingRateHistory(symbol string, startTime, endTime int64) ([]*api.FundingRate, error)
	SubscribeToMarkPrice(symbol string, callback func(float64)) error
}

// futuresMarketCache represents cached market data
type futuresMarketCache struct {
	markPrice     *markPriceCache
	lastPrice     *lastPriceCache
	fundingRate   *fundingRateCache
	cacheMutex    sync.RWMutex
}

type markPriceCache struct {
	price     float64
	timestamp time.Time
}

type lastPriceCache struct {
	price     float64
	timestamp time.Time
}

type fundingRateCache struct {
	rate      *api.FundingRate
	timestamp time.Time
}

// futuresMarketDataService implements FuturesMarketDataService interface
type futuresMarketDataService struct {
	client       api.FuturesClient
	logger       logger.Logger
	cache        map[string]*futuresMarketCache
	cacheTTL     map[string]time.Duration
	cacheMutex   sync.RWMutex
}

// NewFuturesMarketDataService creates a new futures market data service
func NewFuturesMarketDataService(client api.FuturesClient, logger logger.Logger) FuturesMarketDataService {
	return &futuresMarketDataService{
		client: client,
		logger: logger,
		cache:  make(map[string]*futuresMarketCache),
		cacheTTL: map[string]time.Duration{
			"markPrice":   1 * time.Second,
			"lastPrice":   1 * time.Second,
			"fundingRate": 1 * time.Minute,
		},
	}
}

// getOrCreateCache gets or creates cache for a symbol
func (s *futuresMarketDataService) getOrCreateCache(symbol string) *futuresMarketCache {
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()
	
	if cache, exists := s.cache[symbol]; exists {
		return cache
	}
	
	cache := &futuresMarketCache{}
	s.cache[symbol] = cache
	return cache
}

// GetMarkPrice retrieves the mark price for a symbol with caching
func (s *futuresMarketDataService) GetMarkPrice(symbol string) (float64, error) {
	if symbol == "" {
		return 0, fmt.Errorf("symbol cannot be empty")
	}
	
	cache := s.getOrCreateCache(symbol)
	
	// Check cache first
	cache.cacheMutex.RLock()
	if cache.markPrice != nil && time.Since(cache.markPrice.timestamp) < s.cacheTTL["markPrice"] {
		price := cache.markPrice.price
		cache.cacheMutex.RUnlock()
		return price, nil
	}
	cache.cacheMutex.RUnlock()
	
	// Fetch from API with retry
	var markPriceData *api.MarkPrice
	var err error
	
	for attempt := 1; attempt <= 3; attempt++ {
		markPriceData, err = s.client.GetMarkPrice(symbol)
		if err == nil {
			break
		}
		
		if attempt < 3 {
			delay := time.Duration(attempt) * time.Second
			s.logger.Warn("Failed to get mark price, retrying", map[string]interface{}{
				"symbol":  symbol,
				"attempt": attempt,
				"delay":   delay.String(),
				"error":   err.Error(),
			})
			time.Sleep(delay)
		}
	}
	
	if err != nil {
		s.logger.Error("Failed to get mark price after retries", map[string]interface{}{
			"symbol": symbol,
			"error":  err.Error(),
		})
		return 0, fmt.Errorf("failed to get mark price for %s: %w", symbol, err)
	}
	
	if markPriceData.MarkPrice <= 0 {
		return 0, fmt.Errorf("invalid mark price received: %f", markPriceData.MarkPrice)
	}
	
	// Update cache
	cache.cacheMutex.Lock()
	cache.markPrice = &markPriceCache{
		price:     markPriceData.MarkPrice,
		timestamp: time.Now(),
	}
	cache.cacheMutex.Unlock()
	
	s.logger.Debug("Retrieved mark price", map[string]interface{}{
		"symbol":     symbol,
		"mark_price": markPriceData.MarkPrice,
	})
	
	return markPriceData.MarkPrice, nil
}

// GetLastPrice retrieves the last traded price for a symbol with caching
func (s *futuresMarketDataService) GetLastPrice(symbol string) (float64, error) {
	if symbol == "" {
		return 0, fmt.Errorf("symbol cannot be empty")
	}
	
	cache := s.getOrCreateCache(symbol)
	
	// Check cache first
	cache.cacheMutex.RLock()
	if cache.lastPrice != nil && time.Since(cache.lastPrice.timestamp) < s.cacheTTL["lastPrice"] {
		price := cache.lastPrice.price
		cache.cacheMutex.RUnlock()
		return price, nil
	}
	cache.cacheMutex.RUnlock()
	
	// Fetch from API with retry
	var priceData *api.Price
	var err error
	
	for attempt := 1; attempt <= 3; attempt++ {
		priceData, err = s.client.GetPrice(symbol)
		if err == nil {
			break
		}
		
		if attempt < 3 {
			delay := time.Duration(attempt) * time.Second
			s.logger.Warn("Failed to get last price, retrying", map[string]interface{}{
				"symbol":  symbol,
				"attempt": attempt,
				"delay":   delay.String(),
				"error":   err.Error(),
			})
			time.Sleep(delay)
		}
	}
	
	if err != nil {
		s.logger.Error("Failed to get last price after retries", map[string]interface{}{
			"symbol": symbol,
			"error":  err.Error(),
		})
		return 0, fmt.Errorf("failed to get last price for %s: %w", symbol, err)
	}
	
	if priceData.Price <= 0 {
		return 0, fmt.Errorf("invalid last price received: %f", priceData.Price)
	}
	
	// Update cache
	cache.cacheMutex.Lock()
	cache.lastPrice = &lastPriceCache{
		price:     priceData.Price,
		timestamp: time.Now(),
	}
	cache.cacheMutex.Unlock()
	
	s.logger.Debug("Retrieved last price", map[string]interface{}{
		"symbol":     symbol,
		"last_price": priceData.Price,
	})
	
	return priceData.Price, nil
}

// GetHistoricalData retrieves historical kline data for a symbol
func (s *futuresMarketDataService) GetHistoricalData(symbol string, interval string, limit int) ([]*api.Kline, error) {
	if symbol == "" {
		return nil, fmt.Errorf("symbol cannot be empty")
	}
	
	if interval == "" {
		return nil, fmt.Errorf("interval cannot be empty")
	}
	
	if limit <= 0 {
		return nil, fmt.Errorf("limit must be greater than 0")
	}
	
	// Fetch from API with retry
	var klines []*api.Kline
	var err error
	
	for attempt := 1; attempt <= 3; attempt++ {
		klines, err = s.client.GetKlines(symbol, interval, limit)
		if err == nil {
			break
		}
		
		if attempt < 3 {
			delay := time.Duration(attempt) * time.Second
			s.logger.Warn("Failed to get klines, retrying", map[string]interface{}{
				"symbol":   symbol,
				"interval": interval,
				"limit":    limit,
				"attempt":  attempt,
				"delay":    delay.String(),
				"error":    err.Error(),
			})
			time.Sleep(delay)
		}
	}
	
	if err != nil {
		s.logger.Error("Failed to get klines after retries", map[string]interface{}{
			"symbol":   symbol,
			"interval": interval,
			"limit":    limit,
			"error":    err.Error(),
		})
		return nil, fmt.Errorf("failed to get historical data for %s: %w", symbol, err)
	}
	
	s.logger.Debug("Retrieved historical data", map[string]interface{}{
		"symbol":       symbol,
		"interval":     interval,
		"limit":        limit,
		"klines_count": len(klines),
	})
	
	return klines, nil
}

// GetFundingRate retrieves the current funding rate for a symbol with caching
func (s *futuresMarketDataService) GetFundingRate(symbol string) (*api.FundingRate, error) {
	if symbol == "" {
		return nil, fmt.Errorf("symbol cannot be empty")
	}
	
	cache := s.getOrCreateCache(symbol)
	
	// Check cache first
	cache.cacheMutex.RLock()
	if cache.fundingRate != nil && time.Since(cache.fundingRate.timestamp) < s.cacheTTL["fundingRate"] {
		rate := cache.fundingRate.rate
		cache.cacheMutex.RUnlock()
		return rate, nil
	}
	cache.cacheMutex.RUnlock()
	
	// Fetch from API with retry
	var fundingRate *api.FundingRate
	var err error
	
	for attempt := 1; attempt <= 3; attempt++ {
		fundingRate, err = s.client.GetFundingRate(symbol)
		if err == nil {
			break
		}
		
		if attempt < 3 {
			delay := time.Duration(attempt) * time.Second
			s.logger.Warn("Failed to get funding rate, retrying", map[string]interface{}{
				"symbol":  symbol,
				"attempt": attempt,
				"delay":   delay.String(),
				"error":   err.Error(),
			})
			time.Sleep(delay)
		}
	}
	
	if err != nil {
		s.logger.Error("Failed to get funding rate after retries", map[string]interface{}{
			"symbol": symbol,
			"error":  err.Error(),
		})
		return nil, fmt.Errorf("failed to get funding rate for %s: %w", symbol, err)
	}
	
	// Update cache
	cache.cacheMutex.Lock()
	cache.fundingRate = &fundingRateCache{
		rate:      fundingRate,
		timestamp: time.Now(),
	}
	cache.cacheMutex.Unlock()
	
	s.logger.Debug("Retrieved funding rate", map[string]interface{}{
		"symbol":        symbol,
		"funding_rate":  fundingRate.FundingRate,
		"funding_time":  fundingRate.FundingTime,
	})
	
	return fundingRate, nil
}

// GetFundingRateHistory retrieves funding rate history for a symbol
func (s *futuresMarketDataService) GetFundingRateHistory(symbol string, startTime, endTime int64) ([]*api.FundingRate, error) {
	if symbol == "" {
		return nil, fmt.Errorf("symbol cannot be empty")
	}
	
	// Fetch from API with retry
	var rates []*api.FundingRate
	var err error
	
	for attempt := 1; attempt <= 3; attempt++ {
		rates, err = s.client.GetFundingRateHistory(symbol, startTime, endTime)
		if err == nil {
			break
		}
		
		if attempt < 3 {
			delay := time.Duration(attempt) * time.Second
			s.logger.Warn("Failed to get funding rate history, retrying", map[string]interface{}{
				"symbol":     symbol,
				"start_time": startTime,
				"end_time":   endTime,
				"attempt":    attempt,
				"delay":      delay.String(),
				"error":      err.Error(),
			})
			time.Sleep(delay)
		}
	}
	
	if err != nil {
		s.logger.Error("Failed to get funding rate history after retries", map[string]interface{}{
			"symbol":     symbol,
			"start_time": startTime,
			"end_time":   endTime,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("failed to get funding rate history for %s: %w", symbol, err)
	}
	
	s.logger.Debug("Retrieved funding rate history", map[string]interface{}{
		"symbol":      symbol,
		"start_time":  startTime,
		"end_time":    endTime,
		"rates_count": len(rates),
	})
	
	return rates, nil
}

// SubscribeToMarkPrice subscribes to mark price updates for a symbol
func (s *futuresMarketDataService) SubscribeToMarkPrice(symbol string, callback func(float64)) error {
	// This is a placeholder implementation
	// In a real implementation, this would use WebSocket connections
	return fmt.Errorf("mark price subscription not implemented yet")
}
