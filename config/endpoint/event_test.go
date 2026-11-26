package endpoint

import (
	"testing"
)

func TestNewEventFromResult(t *testing.T) {
	if event := NewEventFromResult(&Result{Success: true}); event.Type != EventHealthy {
		t.Error("expected event.Type to be EventHealthy")
	}
	if event := NewEventFromResult(&Result{Success: false}); event.Type != EventUnhealthy {
		t.Error("expected event.Type to be EventUnhealthy")
	}
}

func TestNewEventFromResult_Degraded(t *testing.T) {
	// Succeeded, but one of the degraded-conditions failed: degraded, not unhealthy
	degraded := &Result{Success: true, ConditionResults: []*ConditionResult{
		{Condition: "[STATUS] == 200", Success: true},
		{Condition: "[RESPONSE_TIME] < 700", Success: false},
	}}
	if event := NewEventFromResult(degraded); event.Type != EventDegraded {
		t.Errorf("expected event.Type to be EventDegraded, got %s", event.Type)
	}
	// Failed: always unhealthy, even if some conditions pass
	failed := &Result{Success: false, ConditionResults: []*ConditionResult{
		{Condition: "[STATUS] == 200", Success: true},
		{Condition: "[RESPONSE_TIME] < 5000", Success: false},
	}}
	if event := NewEventFromResult(failed); event.Type != EventUnhealthy {
		t.Errorf("expected event.Type to be EventUnhealthy, got %s", event.Type)
	}
	// All conditions pass: healthy
	healthy := &Result{Success: true, ConditionResults: []*ConditionResult{{Condition: "[STATUS] == 200", Success: true}}}
	if event := NewEventFromResult(healthy); event.Type != EventHealthy {
		t.Errorf("expected event.Type to be EventHealthy, got %s", event.Type)
	}
}
