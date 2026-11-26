package endpoint

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestResult_Degraded(t *testing.T) {
	scenarios := []struct {
		name     string
		result   *Result
		expected bool
	}{
		{"healthy", &Result{Success: true, ConditionResults: []*ConditionResult{{Condition: "[STATUS] == 200", Success: true}}}, false},
		{"failed", &Result{Success: false, ConditionResults: []*ConditionResult{{Condition: "[STATUS] == 200", Success: false}}}, false},
		{"degraded", &Result{Success: true, ConditionResults: []*ConditionResult{{Condition: "[STATUS] == 200", Success: true}, {Condition: "[RESPONSE_TIME] < 700", Success: false}}}, true},
		{"nil", nil, false},
	}
	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			if scenario.result.Degraded() != scenario.expected {
				t.Errorf("expected Degraded() to be %v", scenario.expected)
			}
		})
	}
}

func TestResult_MarshalJSONIncludesDegraded(t *testing.T) {
	result := &Result{Success: true, ConditionResults: []*ConditionResult{{Condition: "[RESPONSE_TIME] < 700", Success: false}}}
	output, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(output), `"degraded":true`) || !strings.Contains(string(output), `"success":true`) {
		t.Errorf("unexpected output: %s", output)
	}
	output, _ = json.Marshal(&Result{Success: true})
	if strings.Contains(string(output), "degraded") {
		t.Errorf("healthy result should not have the degraded field: %s", output)
	}
}

func TestEndpoint_EvaluateHealth_DegradedConditions(t *testing.T) {
	// No call is made for an ICMP-less, URL-less endpoint, so we rely on [CONNECTED]-free conditions only
	ep := &Endpoint{
		Name:               "degraded",
		URL:                "https://example.org",
		Conditions:         []Condition{"1 == 1"},
		DegradedConditions: []Condition{"1 == 2"},
	}
	result := &Result{Success: true}
	for _, condition := range ep.Conditions {
		if !condition.evaluate(result, false, false, nil) {
			result.Success = false
		}
	}
	if result.Success {
		for _, condition := range ep.DegradedConditions {
			condition.evaluate(result, false, false, nil)
		}
	}
	if !result.Success {
		t.Error("a failing degraded condition must not fail the result")
	}
	if !result.Degraded() {
		t.Error("expected the result to be degraded")
	}
}
