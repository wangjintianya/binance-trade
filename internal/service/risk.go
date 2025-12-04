package service

import (
	"binance-trader/internal/api"
	"binance-trader/pkg/errors"
	"fmt"
	"sync"
	"time"
)

// RiskLimits defines risk control parameters
type RiskLimits struct {
	MaxOrderAmount    float64 // Maximum amount per order
	MaxDailyOrders    int     // Maximum orders per day
	MinBalanceReserve float64 // Minimum balance to keep
	MaxAPICallsPerMin int     // Maximum API calls per minute
}

// RiskManager defines the interface for risk management
type RiskManager interface {
	// Risk checks
	ValidateOrder(order *api.OrderRequest) error
	CheckDailyLimit() error
	CheckMinimumBalance(asset string) error

	// Limit management
	UpdateLimits(limits *RiskLimits) error
	GetCurrentLimits() *RiskLimits
}

// riskManager implements the RiskManager interface
type riskManager struct {
	limits        *RiskLimits
	client        api.BinanceClient
	orderHistory  []orderRecord
	mu            sync.RWMutex
}

// orderRecord tracks order creation time for frequency limiting
type orderRecord struct {
	timestamp time.Time
	amount    float64
}

// NewRiskManager creates a new RiskManager instance
func NewRiskManager(limits *RiskLimits, client api.BinanceClient) RiskManager {
	if limits == nil {
		limits = &RiskLimits{
			MaxOrderAmount:    10000.0,
			MaxDailyOrders:    100,
			MinBalanceReserve: 100.0,
			MaxAPICallsPerMin: 1000,
		}
	}
	
	return &riskManager{
		limits:       limits,
		client:       client,
		orderHistory: make([]orderRecord, 0),
	}
}

// ValidateOrder validates an order against risk limits
func (rm *riskManager) ValidateOrder(order *api.OrderRequest) error {
	if order == nil {
		return errors.NewTradingError(
			errors.ErrInvalidParameter,
			"order request cannot be nil",
			0,
			nil,
		)
	}
	
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	
	// Calculate order amount
	var orderAmount float64
	if order.Type == api.OrderTypeMarket {
		// For market orders, we need to get current price
		price, err := rm.client.GetPrice(order.Symbol)
		if err != nil {
			return errors.NewTradingError(
				errors.ErrInvalidParameter,
				fmt.Sprintf("failed to get price for amount validation: %v", err),
				0,
				err,
			)
		}
		orderAmount = order.Quantity * price.Price
	} else {
		// For limit orders, use the specified price
		orderAmount = order.Quantity * order.Price
	}
	
	// Check order amount limit
	if orderAmount > rm.limits.MaxOrderAmount {
		return errors.NewTradingError(
			errors.ErrRiskLimitExceeded,
			fmt.Sprintf("order amount %.2f exceeds maximum limit %.2f", orderAmount, rm.limits.MaxOrderAmount),
			0,
			nil,
		)
	}
	
	return nil
}

// CheckDailyLimit checks if the daily order limit has been reached
func (rm *riskManager) CheckDailyLimit() error {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	
	// Get current time
	now := time.Now()
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	
	// Count orders created today
	todayOrders := 0
	for _, record := range rm.orderHistory {
		if record.timestamp.After(dayStart) || record.timestamp.Equal(dayStart) {
			todayOrders++
		}
	}
	
	// Check if limit exceeded
	if todayOrders >= rm.limits.MaxDailyOrders {
		return errors.NewTradingError(
			errors.ErrRiskLimitExceeded,
			fmt.Sprintf("daily order limit reached: %d/%d", todayOrders, rm.limits.MaxDailyOrders),
			0,
			nil,
		)
	}
	
	return nil
}

// CheckMinimumBalance checks if placing an order would violate minimum balance requirements
func (rm *riskManager) CheckMinimumBalance(asset string) error {
	if asset == "" {
		return errors.NewTradingError(
			errors.ErrInvalidParameter,
			"asset cannot be empty",
			0,
			nil,
		)
	}
	
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	
	// Get current balance
	balance, err := rm.client.GetBalance(asset)
	if err != nil {
		return errors.NewTradingError(
			errors.ErrInvalidParameter,
			fmt.Sprintf("failed to get balance for %s: %v", asset, err),
			0,
			err,
		)
	}
	
	// Check if free balance is below minimum reserve
	if balance.Free < rm.limits.MinBalanceReserve {
		return errors.NewTradingError(
			errors.ErrRiskLimitExceeded,
			fmt.Sprintf("balance %.2f is below minimum reserve %.2f for asset %s", balance.Free, rm.limits.MinBalanceReserve, asset),
			0,
			nil,
		)
	}
	
	return nil
}

// UpdateLimits updates the risk limits
func (rm *riskManager) UpdateLimits(limits *RiskLimits) error {
	if limits == nil {
		return errors.NewTradingError(
			errors.ErrInvalidParameter,
			"limits cannot be nil",
			0,
			nil,
		)
	}
	
	// Validate limits
	if limits.MaxOrderAmount <= 0 {
		return errors.NewTradingError(
			errors.ErrInvalidParameter,
			"max order amount must be greater than 0",
			0,
			nil,
		)
	}
	if limits.MaxDailyOrders <= 0 {
		return errors.NewTradingError(
			errors.ErrInvalidParameter,
			"max daily orders must be greater than 0",
			0,
			nil,
		)
	}
	if limits.MinBalanceReserve < 0 {
		return errors.NewTradingError(
			errors.ErrInvalidParameter,
			"min balance reserve cannot be negative",
			0,
			nil,
		)
	}
	
	rm.mu.Lock()
	defer rm.mu.Unlock()
	
	rm.limits = limits
	return nil
}

// GetCurrentLimits returns the current risk limits
func (rm *riskManager) GetCurrentLimits() *RiskLimits {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	
	// Return a copy to prevent external modification
	return &RiskLimits{
		MaxOrderAmount:    rm.limits.MaxOrderAmount,
		MaxDailyOrders:    rm.limits.MaxDailyOrders,
		MinBalanceReserve: rm.limits.MinBalanceReserve,
		MaxAPICallsPerMin: rm.limits.MaxAPICallsPerMin,
	}
}

// RecordOrder records an order for frequency tracking (internal method)
func (rm *riskManager) RecordOrder(amount float64) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	
	rm.orderHistory = append(rm.orderHistory, orderRecord{
		timestamp: time.Now(),
		amount:    amount,
	})
	
	// Clean up old records (older than 24 hours)
	cutoff := time.Now().Add(-24 * time.Hour)
	newHistory := make([]orderRecord, 0)
	for _, record := range rm.orderHistory {
		if record.timestamp.After(cutoff) {
			newHistory = append(newHistory, record)
		}
	}
	rm.orderHistory = newHistory
}
