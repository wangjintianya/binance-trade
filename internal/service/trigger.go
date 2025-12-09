package service

import (
	"fmt"
	"sync"
	"time"
)

// TriggerType represents the type of trigger condition
type TriggerType int

const (
	TriggerTypePrice TriggerType = iota
	TriggerTypePriceChangePercent
	TriggerTypeVolume
)

// ComparisonOperator represents comparison operators for trigger conditions
type ComparisonOperator int

const (
	OperatorGreaterThan ComparisonOperator = iota
	OperatorLessThan
	OperatorGreaterEqual
	OperatorLessEqual
	OperatorEqual
	OperatorNotEqual
)

// LogicOperator represents logical operators for composite conditions
type LogicOperator int

const (
	LogicOperatorAND LogicOperator = iota
	LogicOperatorOR
)

// TriggerCondition represents a condition that can be evaluated
type TriggerCondition struct {
	Type          TriggerType
	Operator      ComparisonOperator
	Value         float64
	BasePrice     float64
	TimeWindow    time.Duration
	CompositeType LogicOperator
	SubConditions []*TriggerCondition
}

// TriggerCallback is called when a trigger condition is met
type TriggerCallback func(triggerID string, value float64) error

// Trigger represents a registered trigger
type Trigger struct {
	ID        string
	Condition *TriggerCondition
	Callback  TriggerCallback
	Active    bool
	CreatedAt time.Time
}

// TriggerEngine evaluates trigger conditions and executes callbacks
type TriggerEngine interface {
	// Register a new trigger
	RegisterTrigger(id string, condition *TriggerCondition, callback TriggerCallback) error
	
	// Register a condition (alias for RegisterTrigger with nil callback for compatibility)
	RegisterCondition(id string, condition *TriggerCondition) error
	
	// Unregister a trigger
	UnregisterTrigger(id string) error
	
	// Unregister a condition (alias for UnregisterTrigger for compatibility)
	UnregisterCondition(id string) error
	
	// Evaluate a condition with current value
	EvaluateCondition(condition *TriggerCondition, currentValue float64) (bool, error)
	
	// Check all triggers with current value
	CheckTriggers(triggerType TriggerType, currentValue float64) error
	
	// Get all active triggers
	GetActiveTriggers() []*Trigger
	
	// Get trigger by ID
	GetTrigger(id string) (*Trigger, error)
	
	// Activate/deactivate trigger
	SetTriggerActive(id string, active bool) error
}

// triggerEngine implements TriggerEngine interface
type triggerEngine struct {
	mu       sync.RWMutex
	triggers map[string]*Trigger
}

// NewTriggerEngine creates a new trigger engine instance
func NewTriggerEngine() TriggerEngine {
	return &triggerEngine{
		triggers: make(map[string]*Trigger),
	}
}

// RegisterTrigger registers a new trigger
func (te *triggerEngine) RegisterTrigger(id string, condition *TriggerCondition, callback TriggerCallback) error {
	if id == "" {
		return fmt.Errorf("trigger ID cannot be empty")
	}
	
	if condition == nil {
		return fmt.Errorf("trigger condition cannot be nil")
	}
	
	if callback == nil {
		return fmt.Errorf("trigger callback cannot be nil")
	}
	
	te.mu.Lock()
	defer te.mu.Unlock()
	
	if _, exists := te.triggers[id]; exists {
		return fmt.Errorf("trigger with ID %s already exists", id)
	}
	
	te.triggers[id] = &Trigger{
		ID:        id,
		Condition: condition,
		Callback:  callback,
		Active:    true,
		CreatedAt: time.Now(),
	}
	
	return nil
}

// UnregisterTrigger removes a trigger
func (te *triggerEngine) UnregisterTrigger(id string) error {
	te.mu.Lock()
	defer te.mu.Unlock()
	
	if _, exists := te.triggers[id]; !exists {
		return fmt.Errorf("trigger with ID %s not found", id)
	}
	
	delete(te.triggers, id)
	return nil
}

// EvaluateCondition evaluates a single condition
func (te *triggerEngine) EvaluateCondition(condition *TriggerCondition, currentValue float64) (bool, error) {
	if condition == nil {
		return false, fmt.Errorf("condition cannot be nil")
	}
	
	// Handle composite conditions
	if len(condition.SubConditions) > 0 {
		return te.evaluateCompositeCondition(condition, currentValue)
	}
	
	// Evaluate simple condition
	return te.evaluateSimpleCondition(condition, currentValue), nil
}

// evaluateSimpleCondition evaluates a simple comparison
func (te *triggerEngine) evaluateSimpleCondition(condition *TriggerCondition, currentValue float64) bool {
	switch condition.Operator {
	case OperatorGreaterThan:
		return currentValue > condition.Value
	case OperatorGreaterEqual:
		return currentValue >= condition.Value
	case OperatorLessThan:
		return currentValue < condition.Value
	case OperatorLessEqual:
		return currentValue <= condition.Value
	case OperatorEqual:
		return currentValue == condition.Value
	case OperatorNotEqual:
		return currentValue != condition.Value
	default:
		return false
	}
}

// evaluateCompositeCondition evaluates a composite condition with sub-conditions
func (te *triggerEngine) evaluateCompositeCondition(condition *TriggerCondition, currentValue float64) (bool, error) {
	if len(condition.SubConditions) == 0 {
		return false, fmt.Errorf("composite condition has no sub-conditions")
	}
	
	results := make([]bool, len(condition.SubConditions))
	for i, subCondition := range condition.SubConditions {
		result, err := te.EvaluateCondition(subCondition, currentValue)
		if err != nil {
			return false, err
		}
		results[i] = result
	}
	
	// Apply logic operator
	switch condition.CompositeType {
	case LogicOperatorAND:
		for _, result := range results {
			if !result {
				return false, nil
			}
		}
		return true, nil
		
	case LogicOperatorOR:
		for _, result := range results {
			if result {
				return true, nil
			}
		}
		return false, nil
		
	default:
		return false, fmt.Errorf("unknown logic operator: %d", condition.CompositeType)
	}
}

// RegisterCondition registers a condition without a callback (for compatibility)
func (te *triggerEngine) RegisterCondition(id string, condition *TriggerCondition) error {
	// Register with a no-op callback
	return te.RegisterTrigger(id, condition, func(triggerID string, value float64) error {
		return nil
	})
}

// UnregisterCondition unregisters a condition (alias for UnregisterTrigger)
func (te *triggerEngine) UnregisterCondition(id string) error {
	return te.UnregisterTrigger(id)
}

// CheckTriggers checks all active triggers of a specific type
func (te *triggerEngine) CheckTriggers(triggerType TriggerType, currentValue float64) error {
	te.mu.RLock()
	triggersToCheck := make([]*Trigger, 0)
	for _, trigger := range te.triggers {
		if trigger.Active && trigger.Condition.Type == triggerType {
			triggersToCheck = append(triggersToCheck, trigger)
		}
	}
	te.mu.RUnlock()
	
	// Check each trigger
	for _, trigger := range triggersToCheck {
		met, err := te.EvaluateCondition(trigger.Condition, currentValue)
		if err != nil {
			return fmt.Errorf("failed to evaluate trigger %s: %w", trigger.ID, err)
		}
		
		if met {
			// Execute callback if present
			if trigger.Callback != nil {
				if err := trigger.Callback(trigger.ID, currentValue); err != nil {
					return fmt.Errorf("trigger %s callback failed: %w", trigger.ID, err)
				}
			}
		}
	}
	
	return nil
}

// GetActiveTriggers returns all active triggers
func (te *triggerEngine) GetActiveTriggers() []*Trigger {
	te.mu.RLock()
	defer te.mu.RUnlock()
	
	active := make([]*Trigger, 0)
	for _, trigger := range te.triggers {
		if trigger.Active {
			active = append(active, trigger)
		}
	}
	
	return active
}

// GetTrigger returns a trigger by ID
func (te *triggerEngine) GetTrigger(id string) (*Trigger, error) {
	te.mu.RLock()
	defer te.mu.RUnlock()
	
	trigger, exists := te.triggers[id]
	if !exists {
		return nil, fmt.Errorf("trigger with ID %s not found", id)
	}
	
	return trigger, nil
}

// SetTriggerActive activates or deactivates a trigger
func (te *triggerEngine) SetTriggerActive(id string, active bool) error {
	te.mu.Lock()
	defer te.mu.Unlock()
	
	trigger, exists := te.triggers[id]
	if !exists {
		return fmt.Errorf("trigger with ID %s not found", id)
	}
	
	trigger.Active = active
	return nil
}
