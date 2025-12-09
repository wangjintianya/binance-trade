package service

import (
	"time"
)

// MarketData represents market data for a symbol
type MarketData struct {
	Symbol     string
	Price      float64
	Volume     float64
	Volume24h  float64
	Timestamp  int64
}

// TimeWindow represents a time range
type TimeWindow struct {
	StartTime time.Time
	EndTime   time.Time
}

// IsWithinTimeWindow checks if a timestamp is within a time window
func IsWithinTimeWindow(timestamp int64, window *TimeWindow) bool {
	if window == nil {
		return true
	}
	
	t := time.Unix(timestamp, 0)
	
	// If start time is zero, only check end time
	if window.StartTime.IsZero() {
		if window.EndTime.IsZero() {
			return true
		}
		return t.Before(window.EndTime) || t.Equal(window.EndTime)
	}
	
	// If end time is zero, only check start time
	if window.EndTime.IsZero() {
		return t.After(window.StartTime) || t.Equal(window.StartTime)
	}
	
	// Check both start and end time
	return (t.After(window.StartTime) || t.Equal(window.StartTime)) &&
		(t.Before(window.EndTime) || t.Equal(window.EndTime))
}
