package errors

import "fmt"

// ErrorType represents the type of error
type ErrorType int

const (
	ErrNetwork ErrorType = iota
	ErrAuthentication
	ErrRateLimit
	ErrInsufficientBalance
	ErrInvalidParameter
	ErrOrderNotFound
	ErrRiskLimitExceeded
	ErrInvalidTriggerCondition
	ErrConditionalOrderNotFound
	ErrStopOrderNotFound
	ErrOrderAlreadyTriggered
	ErrTimeWindowExpired
	// Futures-specific errors
	ErrInsufficientMargin
	ErrInvalidLeverage
	ErrPositionModeConflict
	ErrMarginModeConflict
	ErrLiquidationRisk
	ErrMaxPositionExceeded
	ErrReduceOnlyViolation
	ErrPositionNotFound
)

// TradingError represents a trading system error
type TradingError struct {
	Type    ErrorType
	Message string
	Code    int
	Cause   error
}

// Error implements the error interface
func (e *TradingError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

// Unwrap returns the underlying error
func (e *TradingError) Unwrap() error {
	return e.Cause
}

// NewTradingError creates a new TradingError
func NewTradingError(errType ErrorType, message string, code int, cause error) *TradingError {
	return &TradingError{
		Type:    errType,
		Message: message,
		Code:    code,
		Cause:   cause,
	}
}
