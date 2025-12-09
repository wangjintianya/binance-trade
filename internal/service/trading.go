package service

import (
	"binance-trader/internal/api"
	"binance-trader/internal/repository"
	"binance-trader/pkg/logger"
)

// OrderStatus represents the status of an order
type OrderStatus struct {
	OrderID     int64
	Symbol      string
	Status      api.OrderStatus
	ExecutedQty float64
	Price       float64
}

// TradingService is an alias for SpotTradingService for backward compatibility
// Deprecated: Use SpotTradingService instead
type TradingService = SpotTradingService

// NewTradingService creates a new trading service instance
// Deprecated: Use NewSpotTradingService instead
func NewTradingService(
	client api.BinanceClient,
	riskMgr RiskManager,
	orderRepo repository.OrderRepository,
	log logger.Logger,
) TradingService {
	return NewSpotTradingService(client, riskMgr, orderRepo, log)
}
