package service

import (
	"fmt"
	"testing"
)

func TestNewTriggerEngine(t *testing.T) {
	engine := NewTriggerEngine()
	if engine == nil {
		t.Fatal("NewTriggerEngine returned nil")
	}
}

func TestRegisterTrigger(t *testing.T) {
	engine := NewTriggerEngine()
	
	condition := &TriggerCondition{
		Type:     TriggerTypePrice,
		Operator: OperatorGreaterThan,
		Value:    100.0,
	}
	
	callback := func(triggerID string, value float64) error {
		return nil
	}
	
	// Test successful registration
	err := engine.RegisterTrigger("test1", condition, callback)
	if err != nil {
		t.Errorf("RegisterTrigger failed: %v", err)
	}
	
	// Test duplicate registration
	err = engine.RegisterTrigger("test1", condition, callback)
	if err == nil {
		t.Error("Expected error for duplicate trigger ID")
	}
	
	// Test empty ID
	err = engine.RegisterTrigger("", condition, callback)
	if err == nil {
		t.Error("Expected error for empty trigger ID")
	}
	
	// Test nil condition
	err = engine.RegisterTrigger("test2", nil, callback)
	if err == nil {
		t.Error("Expected error for nil condition")
	}
	
	// Test nil callback
	err = engine.RegisterTrigger("test3", condition, nil)
	if err == nil {
		t.Error("Expected error for nil callback")
	}
}

func TestRegisterCondition(t *testing.T) {
	engine := NewTriggerEngine()
	
	condition := &TriggerCondition{
		Type:     TriggerTypePrice,
		Operator: OperatorGreaterThan,
		Value:    100.0,
	}
	
	// Test successful registration
	err := engine.RegisterCondition("test1", condition)
	if err != nil {
		t.Errorf("RegisterCondition failed: %v", err)
	}
	
	// Verify it was registered
	trigger, err := engine.GetTrigger("test1")
	if err != nil {
		t.Errorf("GetTrigger failed: %v", err)
	}
	if trigger == nil {
		t.Error("Expected trigger to be registered")
	}
}

func TestUnregisterTrigger(t *testing.T) {
	engine := NewTriggerEngine()
	
	condition := &TriggerCondition{
		Type:     TriggerTypePrice,
		Operator: OperatorGreaterThan,
		Value:    100.0,
	}
	
	callback := func(triggerID string, value float64) error {
		return nil
	}
	
	// Register a trigger
	err := engine.RegisterTrigger("test1", condition, callback)
	if err != nil {
		t.Fatalf("RegisterTrigger failed: %v", err)
	}
	
	// Unregister it
	err = engine.UnregisterTrigger("test1")
	if err != nil {
		t.Errorf("UnregisterTrigger failed: %v", err)
	}
	
	// Try to unregister non-existent trigger
	err = engine.UnregisterTrigger("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent trigger")
	}
}

func TestEvaluateSimpleCondition(t *testing.T) {
	engine := NewTriggerEngine()
	
	tests := []struct {
		name      string
		operator  ComparisonOperator
		threshold float64
		value     float64
		expected  bool
	}{
		{"GreaterThan true", OperatorGreaterThan, 100.0, 101.0, true},
		{"GreaterThan false", OperatorGreaterThan, 100.0, 99.0, false},
		{"GreaterEqual true", OperatorGreaterEqual, 100.0, 100.0, true},
		{"GreaterEqual false", OperatorGreaterEqual, 100.0, 99.0, false},
		{"LessThan true", OperatorLessThan, 100.0, 99.0, true},
		{"LessThan false", OperatorLessThan, 100.0, 101.0, false},
		{"LessEqual true", OperatorLessEqual, 100.0, 100.0, true},
		{"LessEqual false", OperatorLessEqual, 100.0, 101.0, false},
		{"Equal true", OperatorEqual, 100.0, 100.0, true},
		{"Equal false", OperatorEqual, 100.0, 99.0, false},
		{"NotEqual true", OperatorNotEqual, 100.0, 99.0, true},
		{"NotEqual false", OperatorNotEqual, 100.0, 100.0, false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			condition := &TriggerCondition{
				Type:     TriggerTypePrice,
				Operator: tt.operator,
				Value:    tt.threshold,
			}
			
			result, err := engine.EvaluateCondition(condition, tt.value)
			if err != nil {
				t.Errorf("EvaluateCondition failed: %v", err)
			}
			
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestEvaluateCompositeCondition(t *testing.T) {
	engine := NewTriggerEngine()
	
	// Test AND condition
	andCondition := &TriggerCondition{
		Type:          TriggerTypePrice,
		CompositeType: LogicOperatorAND,
		SubConditions: []*TriggerCondition{
			{
				Type:     TriggerTypePrice,
				Operator: OperatorGreaterThan,
				Value:    100.0,
			},
			{
				Type:     TriggerTypePrice,
				Operator: OperatorLessThan,
				Value:    200.0,
			},
		},
	}
	
	// Value in range (100, 200)
	result, err := engine.EvaluateCondition(andCondition, 150.0)
	if err != nil {
		t.Errorf("EvaluateCondition failed: %v", err)
	}
	if !result {
		t.Error("Expected AND condition to be true for value 150")
	}
	
	// Value out of range
	result, err = engine.EvaluateCondition(andCondition, 250.0)
	if err != nil {
		t.Errorf("EvaluateCondition failed: %v", err)
	}
	if result {
		t.Error("Expected AND condition to be false for value 250")
	}
	
	// Test OR condition
	orCondition := &TriggerCondition{
		Type:          TriggerTypePrice,
		CompositeType: LogicOperatorOR,
		SubConditions: []*TriggerCondition{
			{
				Type:     TriggerTypePrice,
				Operator: OperatorLessThan,
				Value:    100.0,
			},
			{
				Type:     TriggerTypePrice,
				Operator: OperatorGreaterThan,
				Value:    200.0,
			},
		},
	}
	
	// Value satisfies first condition
	result, err = engine.EvaluateCondition(orCondition, 50.0)
	if err != nil {
		t.Errorf("EvaluateCondition failed: %v", err)
	}
	if !result {
		t.Error("Expected OR condition to be true for value 50")
	}
	
	// Value satisfies neither condition
	result, err = engine.EvaluateCondition(orCondition, 150.0)
	if err != nil {
		t.Errorf("EvaluateCondition failed: %v", err)
	}
	if result {
		t.Error("Expected OR condition to be false for value 150")
	}
}

func TestCheckTriggers(t *testing.T) {
	engine := NewTriggerEngine()
	
	callCount := 0
	var lastValue float64
	
	callback := func(triggerID string, value float64) error {
		callCount++
		lastValue = value
		return nil
	}
	
	// Register a trigger
	condition := &TriggerCondition{
		Type:     TriggerTypePrice,
		Operator: OperatorGreaterThan,
		Value:    100.0,
	}
	
	err := engine.RegisterTrigger("test1", condition, callback)
	if err != nil {
		t.Fatalf("RegisterTrigger failed: %v", err)
	}
	
	// Check with value that doesn't meet condition
	err = engine.CheckTriggers(TriggerTypePrice, 99.0)
	if err != nil {
		t.Errorf("CheckTriggers failed: %v", err)
	}
	if callCount != 0 {
		t.Error("Callback should not have been called")
	}
	
	// Check with value that meets condition
	err = engine.CheckTriggers(TriggerTypePrice, 101.0)
	if err != nil {
		t.Errorf("CheckTriggers failed: %v", err)
	}
	if callCount != 1 {
		t.Errorf("Expected callback to be called once, got %d", callCount)
	}
	if lastValue != 101.0 {
		t.Errorf("Expected callback value 101.0, got %f", lastValue)
	}
}

func TestGetActiveTriggers(t *testing.T) {
	engine := NewTriggerEngine()
	
	callback := func(triggerID string, value float64) error {
		return nil
	}
	
	condition := &TriggerCondition{
		Type:     TriggerTypePrice,
		Operator: OperatorGreaterThan,
		Value:    100.0,
	}
	
	// Register multiple triggers
	engine.RegisterTrigger("test1", condition, callback)
	engine.RegisterTrigger("test2", condition, callback)
	engine.RegisterTrigger("test3", condition, callback)
	
	// Deactivate one
	engine.SetTriggerActive("test2", false)
	
	active := engine.GetActiveTriggers()
	if len(active) != 2 {
		t.Errorf("Expected 2 active triggers, got %d", len(active))
	}
}

func TestGetTrigger(t *testing.T) {
	engine := NewTriggerEngine()
	
	callback := func(triggerID string, value float64) error {
		return nil
	}
	
	condition := &TriggerCondition{
		Type:     TriggerTypePrice,
		Operator: OperatorGreaterThan,
		Value:    100.0,
	}
	
	err := engine.RegisterTrigger("test1", condition, callback)
	if err != nil {
		t.Fatalf("RegisterTrigger failed: %v", err)
	}
	
	// Get existing trigger
	trigger, err := engine.GetTrigger("test1")
	if err != nil {
		t.Errorf("GetTrigger failed: %v", err)
	}
	if trigger.ID != "test1" {
		t.Errorf("Expected trigger ID 'test1', got '%s'", trigger.ID)
	}
	
	// Get non-existent trigger
	_, err = engine.GetTrigger("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent trigger")
	}
}

func TestSetTriggerActive(t *testing.T) {
	engine := NewTriggerEngine()
	
	callback := func(triggerID string, value float64) error {
		return nil
	}
	
	condition := &TriggerCondition{
		Type:     TriggerTypePrice,
		Operator: OperatorGreaterThan,
		Value:    100.0,
	}
	
	err := engine.RegisterTrigger("test1", condition, callback)
	if err != nil {
		t.Fatalf("RegisterTrigger failed: %v", err)
	}
	
	// Deactivate trigger
	err = engine.SetTriggerActive("test1", false)
	if err != nil {
		t.Errorf("SetTriggerActive failed: %v", err)
	}
	
	trigger, _ := engine.GetTrigger("test1")
	if trigger.Active {
		t.Error("Expected trigger to be inactive")
	}
	
	// Activate trigger
	err = engine.SetTriggerActive("test1", true)
	if err != nil {
		t.Errorf("SetTriggerActive failed: %v", err)
	}
	
	trigger, _ = engine.GetTrigger("test1")
	if !trigger.Active {
		t.Error("Expected trigger to be active")
	}
	
	// Try to set non-existent trigger
	err = engine.SetTriggerActive("nonexistent", true)
	if err == nil {
		t.Error("Expected error for non-existent trigger")
	}
}

func TestCallbackError(t *testing.T) {
	engine := NewTriggerEngine()
	
	callback := func(triggerID string, value float64) error {
		return fmt.Errorf("callback error")
	}
	
	condition := &TriggerCondition{
		Type:     TriggerTypePrice,
		Operator: OperatorGreaterThan,
		Value:    100.0,
	}
	
	err := engine.RegisterTrigger("test1", condition, callback)
	if err != nil {
		t.Fatalf("RegisterTrigger failed: %v", err)
	}
	
	// Check with value that meets condition
	err = engine.CheckTriggers(TriggerTypePrice, 101.0)
	if err == nil {
		t.Error("Expected error from callback")
	}
}
