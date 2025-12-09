package service

import (
	"binance-trader/internal/api"
	"binance-trader/pkg/logger"
	"fmt"
	"sync"
	"time"
)

// FundingFeeSettlement represents a funding fee settlement record
type FundingFeeSettlement struct {
	Symbol       string
	PositionSide api.PositionSide
	PositionAmt  float64
	FundingRate  float64
	FundingFee   float64
	SettleTime   int64
}

// FuturesFundingService defines the interface for funding rate processing
type FuturesFundingService interface {
	// Query funding rate at settlement time
	QueryFundingRateAtSettlement(symbol string, settleTime int64) (*api.FundingRate, error)
	
	// Calculate funding fee for a position
	CalculateFundingFee(position *api.Position, fundingRate float64) (float64, error)
	
	// Process funding fee settlement
	ProcessFundingSettlement(position *api.Position, fundingRate *api.FundingRate) (*FundingFeeSettlement, error)
	
	// Update account balance and position cost after settlement
	UpdateAfterSettlement(settlement *FundingFeeSettlement, currentBalance float64, currentCost float64) (newBalance float64, newCost float64, error error)
	
	// Get funding rate history
	GetFundingRateHistory(symbol string, startTime, endTime int64) ([]*api.FundingRate, error)
	
	// Start automatic funding rate monitoring
	StartMonitoring(checkInterval time.Duration) error
	
	// Stop monitoring
	StopMonitoring() error
}

// futuresFundingService implements FuturesFundingService
type futuresFundingService struct {
	marketService FuturesMarketDataService
	logger        logger.Logger
	
	// Monitoring
	stopChan      chan struct{}
	monitoringMu  sync.Mutex
	isMonitoring  bool
	
	// Settlement records
	settlements   []*FundingFeeSettlement
	settlementsMu sync.RWMutex
}

// NewFuturesFundingService creates a new futures funding service
func NewFuturesFundingService(marketService FuturesMarketDataService, logger logger.Logger) FuturesFundingService {
	return &futuresFundingService{
		marketService: marketService,
		logger:        logger,
		settlements:   make([]*FundingFeeSettlement, 0),
	}
}

// QueryFundingRateAtSettlement queries the funding rate at a specific settlement time
func (s *futuresFundingService) QueryFundingRateAtSettlement(symbol string, settleTime int64) (*api.FundingRate, error) {
	if symbol == "" {
		return nil, fmt.Errorf("symbol cannot be empty")
	}
	
	if settleTime <= 0 {
		return nil, fmt.Errorf("settle time must be positive")
	}
	
	// Check if current time is at or past settlement time
	currentTime := time.Now().Unix() * 1000 // Convert to milliseconds
	if currentTime < settleTime {
		return nil, fmt.Errorf("settlement time has not arrived yet")
	}
	
	// Query funding rate from market service
	fundingRate, err := s.marketService.GetFundingRate(symbol)
	if err != nil {
		s.logger.Error("Failed to query funding rate at settlement", map[string]interface{}{
			"symbol":      symbol,
			"settle_time": settleTime,
			"error":       err.Error(),
		})
		return nil, fmt.Errorf("failed to query funding rate: %w", err)
	}
	
	s.logger.Info("Queried funding rate at settlement", map[string]interface{}{
		"symbol":       symbol,
		"settle_time":  settleTime,
		"funding_rate": fundingRate.FundingRate,
	})
	
	return fundingRate, nil
}

// CalculateFundingFee calculates the funding fee for a position
func (s *futuresFundingService) CalculateFundingFee(position *api.Position, fundingRate float64) (float64, error) {
	if position == nil {
		return 0, fmt.Errorf("position cannot be nil")
	}
	
	if position.PositionAmt == 0 {
		return 0, nil // No position, no fee
	}
	
	// Calculate notional value
	notionalValue := position.PositionAmt * position.MarkPrice
	if notionalValue < 0 {
		notionalValue = -notionalValue
	}
	
	// Calculate funding fee
	// For long positions with positive funding rate: pay fee (negative)
	// For short positions with positive funding rate: receive fee (positive)
	var fundingFee float64
	if position.PositionSide == api.PositionSideLong {
		// Long position pays when funding rate is positive
		fundingFee = -notionalValue * fundingRate
	} else if position.PositionSide == api.PositionSideShort {
		// Short position receives when funding rate is positive
		fundingFee = notionalValue * fundingRate
	} else {
		return 0, fmt.Errorf("invalid position side: %s", position.PositionSide)
	}
	
	s.logger.Debug("Calculated funding fee", map[string]interface{}{
		"symbol":          position.Symbol,
		"position_side":   position.PositionSide,
		"position_amt":    position.PositionAmt,
		"mark_price":      position.MarkPrice,
		"notional_value":  notionalValue,
		"funding_rate":    fundingRate,
		"funding_fee":     fundingFee,
	})
	
	return fundingFee, nil
}

// ProcessFundingSettlement processes a funding fee settlement
func (s *futuresFundingService) ProcessFundingSettlement(position *api.Position, fundingRate *api.FundingRate) (*FundingFeeSettlement, error) {
	if position == nil {
		return nil, fmt.Errorf("position cannot be nil")
	}
	
	if fundingRate == nil {
		return nil, fmt.Errorf("funding rate cannot be nil")
	}
	
	// Calculate funding fee
	fundingFee, err := s.CalculateFundingFee(position, fundingRate.FundingRate)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate funding fee: %w", err)
	}
	
	// Create settlement record
	settlement := &FundingFeeSettlement{
		Symbol:       position.Symbol,
		PositionSide: position.PositionSide,
		PositionAmt:  position.PositionAmt,
		FundingRate:  fundingRate.FundingRate,
		FundingFee:   fundingFee,
		SettleTime:   fundingRate.FundingTime,
	}
	
	// Store settlement record
	s.settlementsMu.Lock()
	s.settlements = append(s.settlements, settlement)
	s.settlementsMu.Unlock()
	
	// Log settlement
	s.logger.Info("Processed funding fee settlement", map[string]interface{}{
		"symbol":        settlement.Symbol,
		"position_side": settlement.PositionSide,
		"position_amt":  settlement.PositionAmt,
		"funding_rate":  settlement.FundingRate,
		"funding_fee":   settlement.FundingFee,
		"settle_time":   settlement.SettleTime,
	})
	
	return settlement, nil
}

// UpdateAfterSettlement updates account balance and position cost after settlement
func (s *futuresFundingService) UpdateAfterSettlement(settlement *FundingFeeSettlement, currentBalance float64, currentCost float64) (newBalance float64, newCost float64, error error) {
	if settlement == nil {
		return 0, 0, fmt.Errorf("settlement cannot be nil")
	}
	
	// Update balance: add funding fee (positive if received, negative if paid)
	newBalance = currentBalance + settlement.FundingFee
	
	// Update position cost: subtract funding fee from cost
	// If we paid fee (negative), cost increases
	// If we received fee (positive), cost decreases
	newCost = currentCost - settlement.FundingFee
	
	s.logger.Info("Updated balance and cost after settlement", map[string]interface{}{
		"symbol":          settlement.Symbol,
		"funding_fee":     settlement.FundingFee,
		"old_balance":     currentBalance,
		"new_balance":     newBalance,
		"old_cost":        currentCost,
		"new_cost":        newCost,
	})
	
	return newBalance, newCost, nil
}

// GetFundingRateHistory retrieves funding rate history
func (s *futuresFundingService) GetFundingRateHistory(symbol string, startTime, endTime int64) ([]*api.FundingRate, error) {
	if symbol == "" {
		return nil, fmt.Errorf("symbol cannot be empty")
	}
	
	if startTime < 0 || endTime < 0 {
		return nil, fmt.Errorf("time values cannot be negative")
	}
	
	if startTime > endTime {
		return nil, fmt.Errorf("start time cannot be after end time")
	}
	
	// Query from market service
	rates, err := s.marketService.GetFundingRateHistory(symbol, startTime, endTime)
	if err != nil {
		s.logger.Error("Failed to get funding rate history", map[string]interface{}{
			"symbol":     symbol,
			"start_time": startTime,
			"end_time":   endTime,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("failed to get funding rate history: %w", err)
	}
	
	// Filter by time range
	filtered := make([]*api.FundingRate, 0)
	for _, rate := range rates {
		if rate.FundingTime >= startTime && rate.FundingTime <= endTime {
			filtered = append(filtered, rate)
		}
	}
	
	s.logger.Debug("Retrieved funding rate history", map[string]interface{}{
		"symbol":      symbol,
		"start_time":  startTime,
		"end_time":    endTime,
		"total_count": len(rates),
		"filtered_count": len(filtered),
	})
	
	return filtered, nil
}

// StartMonitoring starts automatic funding rate monitoring
func (s *futuresFundingService) StartMonitoring(checkInterval time.Duration) error {
	s.monitoringMu.Lock()
	defer s.monitoringMu.Unlock()
	
	if s.isMonitoring {
		return fmt.Errorf("monitoring is already running")
	}
	
	if checkInterval <= 0 {
		return fmt.Errorf("check interval must be positive")
	}
	
	s.stopChan = make(chan struct{})
	s.isMonitoring = true
	
	go s.monitoringLoop(checkInterval)
	
	s.logger.Info("Started funding rate monitoring", map[string]interface{}{
		"check_interval": checkInterval.String(),
	})
	
	return nil
}

// StopMonitoring stops the monitoring
func (s *futuresFundingService) StopMonitoring() error {
	s.monitoringMu.Lock()
	defer s.monitoringMu.Unlock()
	
	if !s.isMonitoring {
		return fmt.Errorf("monitoring is not running")
	}
	
	close(s.stopChan)
	s.isMonitoring = false
	
	s.logger.Info("Stopped funding rate monitoring", nil)
	
	return nil
}

// monitoringLoop is the main monitoring loop
func (s *futuresFundingService) monitoringLoop(checkInterval time.Duration) {
	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			// This is a placeholder for actual monitoring logic
			// In a real implementation, this would check for settlement times
			// and trigger settlements automatically
			s.logger.Debug("Funding rate monitoring tick", nil)
		}
	}
}
