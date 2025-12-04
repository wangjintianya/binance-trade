package service

import (
	"binance-trader/internal/api"
	"fmt"
	"testing"
	"time"
)

// TestGetCurrentPrice_Success tests successful price retrieval
func TestGetCurrentPrice_Success(t *testing.T) {
	mockClient := &mockBinanceClient{
		getPriceFunc: func(symbol string) (*api.Price, error) {
			return &api.Price{
				Symbol: symbol,
				Price:  50000.0,
			}, nil
		},
	}
	
	service := NewMarketDataService(mockClient, 1*time.Second)
	
	price, err := service.GetCurrentPrice("BTCUSDT")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	
	if price != 50000.0 {
		t.Errorf("expected price 50000.0, got %f", price)
	}
}

// TestGetCurrentPrice_EmptySymbol tests error handling for empty symbol
func TestGetCurrentPrice_EmptySymbol(t *testing.T) {
	mockClient := &mockBinanceClient{}
	service := NewMarketDataService(mockClient, 1*time.Second)
	
	_, err := service.GetCurrentPrice("")
	if err == nil {
		t.Fatal("expected error for empty symbol, got nil")
	}
}

// TestGetCurrentPrice_APIError tests error handling when API fails
func TestGetCurrentPrice_APIError(t *testing.T) {
	mockClient := &mockBinanceClient{
		getPriceFunc: func(symbol string) (*api.Price, error) {
			return nil, fmt.Errorf("API error")
		},
	}
	
	service := NewMarketDataService(mockClient, 1*time.Second)
	
	_, err := service.GetCurrentPrice("BTCUSDT")
	if err == nil {
		t.Fatal("expected error when API fails, got nil")
	}
}

// TestGetCurrentPrice_InvalidPrice tests error handling for invalid price
func TestGetCurrentPrice_InvalidPrice(t *testing.T) {
	mockClient := &mockBinanceClient{
		getPriceFunc: func(symbol string) (*api.Price, error) {
			return &api.Price{
				Symbol: symbol,
				Price:  0,
			}, nil
		},
	}
	
	service := NewMarketDataService(mockClient, 1*time.Second)
	
	_, err := service.GetCurrentPrice("BTCUSDT")
	if err == nil {
		t.Fatal("expected error for invalid price, got nil")
	}
}

// TestGetCurrentPrice_CacheHit tests that cache is used when valid
func TestGetCurrentPrice_CacheHit(t *testing.T) {
	callCount := 0
	mockClient := &mockBinanceClient{
		getPriceFunc: func(symbol string) (*api.Price, error) {
			callCount++
			return &api.Price{
				Symbol: symbol,
				Price:  50000.0,
			}, nil
		},
	}
	
	service := NewMarketDataService(mockClient, 1*time.Second)
	
	// First call - should hit API
	price1, err := service.GetCurrentPrice("BTCUSDT")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	
	// Second call - should use cache
	price2, err := service.GetCurrentPrice("BTCUSDT")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	
	if price1 != price2 {
		t.Errorf("expected same price from cache, got %f and %f", price1, price2)
	}
	
	if callCount != 1 {
		t.Errorf("expected API to be called once, got %d calls", callCount)
	}
}

// TestGetCurrentPrice_CacheExpiry tests that cache expires after TTL
func TestGetCurrentPrice_CacheExpiry(t *testing.T) {
	callCount := 0
	mockClient := &mockBinanceClient{
		getPriceFunc: func(symbol string) (*api.Price, error) {
			callCount++
			return &api.Price{
				Symbol: symbol,
				Price:  50000.0 + float64(callCount)*100,
			}, nil
		},
	}
	
	service := NewMarketDataService(mockClient, 50*time.Millisecond)
	
	// First call
	price1, err := service.GetCurrentPrice("BTCUSDT")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	
	// Wait for cache to expire
	time.Sleep(100 * time.Millisecond)
	
	// Second call - should hit API again
	price2, err := service.GetCurrentPrice("BTCUSDT")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	
	if price1 == price2 {
		t.Errorf("expected different prices after cache expiry, got %f both times", price1)
	}
	
	if callCount != 2 {
		t.Errorf("expected API to be called twice, got %d calls", callCount)
	}
}

// TestGetCurrentPrice_MultipleConcurrentRequests tests concurrent access to cache
func TestGetCurrentPrice_MultipleConcurrentRequests(t *testing.T) {
	mockClient := &mockBinanceClient{
		getPriceFunc: func(symbol string) (*api.Price, error) {
			return &api.Price{
				Symbol: symbol,
				Price:  50000.0,
			}, nil
		},
	}
	
	service := NewMarketDataService(mockClient, 1*time.Second)
	
	// Make concurrent requests
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			_, err := service.GetCurrentPrice("BTCUSDT")
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			done <- true
		}()
	}
	
	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

// TestGetHistoricalData_Success tests successful kline retrieval
func TestGetHistoricalData_Success(t *testing.T) {
	expectedKlines := []*api.Kline{
		{
			OpenTime:  1609459200000,
			Open:      29000.0,
			High:      29500.0,
			Low:       28500.0,
			Close:     29200.0,
			Volume:    1000.0,
			CloseTime: 1609462799999,
		},
		{
			OpenTime:  1609462800000,
			Open:      29200.0,
			High:      29800.0,
			Low:       29000.0,
			Close:     29500.0,
			Volume:    1200.0,
			CloseTime: 1609466399999,
		},
	}
	
	mockClient := &mockBinanceClient{
		getKlinesFunc: func(symbol string, interval string, limit int) ([]*api.Kline, error) {
			return expectedKlines, nil
		},
	}
	
	service := NewMarketDataService(mockClient, 1*time.Second)
	
	klines, err := service.GetHistoricalData("BTCUSDT", "1h", 2)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	
	if len(klines) != 2 {
		t.Errorf("expected 2 klines, got %d", len(klines))
	}
	
	if klines[0].Open != 29000.0 {
		t.Errorf("expected first kline open 29000.0, got %f", klines[0].Open)
	}
}

// TestGetHistoricalData_EmptySymbol tests error handling for empty symbol
func TestGetHistoricalData_EmptySymbol(t *testing.T) {
	mockClient := &mockBinanceClient{}
	service := NewMarketDataService(mockClient, 1*time.Second)
	
	_, err := service.GetHistoricalData("", "1h", 10)
	if err == nil {
		t.Fatal("expected error for empty symbol, got nil")
	}
}

// TestGetHistoricalData_EmptyInterval tests error handling for empty interval
func TestGetHistoricalData_EmptyInterval(t *testing.T) {
	mockClient := &mockBinanceClient{}
	service := NewMarketDataService(mockClient, 1*time.Second)
	
	_, err := service.GetHistoricalData("BTCUSDT", "", 10)
	if err == nil {
		t.Fatal("expected error for empty interval, got nil")
	}
}

// TestGetHistoricalData_InvalidLimit tests error handling for invalid limit
func TestGetHistoricalData_InvalidLimit(t *testing.T) {
	mockClient := &mockBinanceClient{}
	service := NewMarketDataService(mockClient, 1*time.Second)
	
	_, err := service.GetHistoricalData("BTCUSDT", "1h", 0)
	if err == nil {
		t.Fatal("expected error for invalid limit, got nil")
	}
	
	_, err = service.GetHistoricalData("BTCUSDT", "1h", -1)
	if err == nil {
		t.Fatal("expected error for negative limit, got nil")
	}
}

// TestGetHistoricalData_APIError tests error handling when API fails
func TestGetHistoricalData_APIError(t *testing.T) {
	mockClient := &mockBinanceClient{
		getKlinesFunc: func(symbol string, interval string, limit int) ([]*api.Kline, error) {
			return nil, fmt.Errorf("API error")
		},
	}
	
	service := NewMarketDataService(mockClient, 1*time.Second)
	
	_, err := service.GetHistoricalData("BTCUSDT", "1h", 10)
	if err == nil {
		t.Fatal("expected error when API fails, got nil")
	}
}

// TestNewMarketDataService_DefaultCacheTTL tests default cache TTL
func TestNewMarketDataService_DefaultCacheTTL(t *testing.T) {
	mockClient := &mockBinanceClient{}
	
	// Test with zero TTL - should use default
	service := NewMarketDataService(mockClient, 0)
	if service == nil {
		t.Fatal("expected service to be created, got nil")
	}
	
	// Test with negative TTL - should use default
	service = NewMarketDataService(mockClient, -1*time.Second)
	if service == nil {
		t.Fatal("expected service to be created, got nil")
	}
}
