package repository

import (
	"binance-trader/internal/api"
	"binance-trader/pkg/errors"
	"fmt"
	"sync"
)

// ClosedPosition represents a closed futures position
type ClosedPosition struct {
	Symbol           string
	PositionSide     api.PositionSide
	EntryPrice       float64
	ExitPrice        float64
	Quantity         float64
	RealizedProfit   float64
	OpenTime         int64
	CloseTime        int64
	Commission       float64
}

// FuturesPositionRepository defines the interface for futures position data persistence
type FuturesPositionRepository interface {
	// Position operations
	SavePosition(position *api.Position) error
	GetPosition(symbol string, positionSide api.PositionSide) (*api.Position, error)
	GetAllPositions() ([]*api.Position, error)
	GetPositionsBySymbol(symbol string) ([]*api.Position, error)
	DeletePosition(symbol string, positionSide api.PositionSide) error
	
	// Closed position operations
	SaveClosedPosition(closedPosition *ClosedPosition) error
	GetPositionHistory(symbol string, startTime, endTime int64) ([]*ClosedPosition, error)
}

// memoryFuturesPositionRepository implements FuturesPositionRepository using in-memory storage
type memoryFuturesPositionRepository struct {
	mu              sync.RWMutex
	positions       map[string]*api.Position // key: symbol_positionSide
	closedPositions []*ClosedPosition
}

// NewMemoryFuturesPositionRepository creates a new in-memory futures position repository
func NewMemoryFuturesPositionRepository() FuturesPositionRepository {
	return &memoryFuturesPositionRepository{
		positions:       make(map[string]*api.Position),
		closedPositions: make([]*ClosedPosition, 0),
	}
}

// positionKey generates a unique key for a position
func positionKey(symbol string, positionSide api.PositionSide) string {
	return fmt.Sprintf("%s_%s", symbol, positionSide)
}

// SavePosition stores or updates a futures position
func (r *memoryFuturesPositionRepository) SavePosition(position *api.Position) error {
	if position == nil {
		return errors.NewTradingError(errors.ErrInvalidParameter, "position cannot be nil", 0, nil)
	}
	
	if position.Symbol == "" {
		return errors.NewTradingError(errors.ErrInvalidParameter, "symbol cannot be empty", 0, nil)
	}
	
	r.mu.Lock()
	defer r.mu.Unlock()
	
	// Create a copy to avoid external modifications
	positionCopy := *position
	key := positionKey(position.Symbol, position.PositionSide)
	r.positions[key] = &positionCopy
	
	return nil
}

// GetPosition retrieves a specific futures position
func (r *memoryFuturesPositionRepository) GetPosition(symbol string, positionSide api.PositionSide) (*api.Position, error) {
	if symbol == "" {
		return nil, errors.NewTradingError(errors.ErrInvalidParameter, "symbol cannot be empty", 0, nil)
	}
	
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	key := positionKey(symbol, positionSide)
	position, exists := r.positions[key]
	if !exists {
		return nil, errors.NewTradingError(errors.ErrPositionNotFound, "position not found", 0, nil)
	}
	
	// Return a copy to prevent external modifications
	positionCopy := *position
	return &positionCopy, nil
}

// GetAllPositions retrieves all futures positions
func (r *memoryFuturesPositionRepository) GetAllPositions() ([]*api.Position, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	result := make([]*api.Position, 0, len(r.positions))
	for _, position := range r.positions {
		positionCopy := *position
		result = append(result, &positionCopy)
	}
	
	return result, nil
}

// GetPositionsBySymbol retrieves all positions for a specific symbol
func (r *memoryFuturesPositionRepository) GetPositionsBySymbol(symbol string) ([]*api.Position, error) {
	if symbol == "" {
		return nil, errors.NewTradingError(errors.ErrInvalidParameter, "symbol cannot be empty", 0, nil)
	}
	
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	result := make([]*api.Position, 0)
	for _, position := range r.positions {
		if position.Symbol == symbol {
			positionCopy := *position
			result = append(result, &positionCopy)
		}
	}
	
	return result, nil
}

// DeletePosition removes a futures position
func (r *memoryFuturesPositionRepository) DeletePosition(symbol string, positionSide api.PositionSide) error {
	if symbol == "" {
		return errors.NewTradingError(errors.ErrInvalidParameter, "symbol cannot be empty", 0, nil)
	}
	
	r.mu.Lock()
	defer r.mu.Unlock()
	
	key := positionKey(symbol, positionSide)
	if _, exists := r.positions[key]; !exists {
		return errors.NewTradingError(errors.ErrPositionNotFound, "position not found", 0, nil)
	}
	
	delete(r.positions, key)
	return nil
}

// SaveClosedPosition stores a closed position record
func (r *memoryFuturesPositionRepository) SaveClosedPosition(closedPosition *ClosedPosition) error {
	if closedPosition == nil {
		return errors.NewTradingError(errors.ErrInvalidParameter, "closed position cannot be nil", 0, nil)
	}
	
	if closedPosition.Symbol == "" {
		return errors.NewTradingError(errors.ErrInvalidParameter, "symbol cannot be empty", 0, nil)
	}
	
	r.mu.Lock()
	defer r.mu.Unlock()
	
	// Create a copy to avoid external modifications
	closedPositionCopy := *closedPosition
	r.closedPositions = append(r.closedPositions, &closedPositionCopy)
	
	return nil
}

// GetPositionHistory retrieves closed position history within a time range
func (r *memoryFuturesPositionRepository) GetPositionHistory(symbol string, startTime, endTime int64) ([]*ClosedPosition, error) {
	if symbol == "" {
		return nil, errors.NewTradingError(errors.ErrInvalidParameter, "symbol cannot be empty", 0, nil)
	}
	
	if startTime < 0 || endTime < 0 {
		return nil, errors.NewTradingError(errors.ErrInvalidParameter, "time values cannot be negative", 0, nil)
	}
	
	if startTime > endTime {
		return nil, errors.NewTradingError(errors.ErrInvalidParameter, "start time cannot be after end time", 0, nil)
	}
	
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	result := make([]*ClosedPosition, 0)
	for _, closedPos := range r.closedPositions {
		if closedPos.Symbol == symbol && 
		   closedPos.CloseTime >= startTime && 
		   closedPos.CloseTime <= endTime {
			closedPosCopy := *closedPos
			result = append(result, &closedPosCopy)
		}
	}
	
	return result, nil
}
