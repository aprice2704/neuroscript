// NeuroScript Version: 0.7.0
// File version: 7
// Purpose: A comprehensive test suite for the ns_event toolset, including validation, edge cases, and round-trip tests.
// filename: pkg/tool/ns_event/tools_event_rtrip_test.go
// nlines: 345
// risk_rating: LOW
package ns_event_test

import (
	"reflect"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/json_lite"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

func TestToolEvent_RoundTrip(t *testing.T) {
	// 1. Compose an event
	interp := newEventTestInterpreter(t)
	composeTool, _ := interp.ToolRegistry().GetTool("ns_event.Compose")
	payload := map[string]interface{}{"data": "value", "num": 123.0}
	agentID := "test-agent"
	kind := "test.kind"

	composedEvent, err := composeTool.Func(interp, []interface{}{kind, payload, nil, agentID})
	if err != nil {
		t.Fatalf("Compose failed: %v", err)
	}

	// 2. Extract and verify data using getters
	getIDTool, _ := interp.ToolRegistry().GetTool("ns_event.GetID")
	getKindTool, _ := interp.ToolRegistry().GetTool("ns_event.GetKind")
	getPayloadTool, _ := interp.ToolRegistry().GetTool("ns_event.GetPayload")

	id, _ := getIDTool.Func(interp, []interface{}{composedEvent})
	if id == "" {
		t.Error("GetID returned an empty ID from a composed event")
	}

	gotKind, _ := getKindTool.Func(interp, []interface{}{composedEvent})
	if gotKind != kind {
		t.Errorf("GetKind mismatch: want '%s', got '%s'", kind, gotKind)
	}

	gotPayload, _ := getPayloadTool.Func(interp, []interface{}{composedEvent})
	if !reflect.DeepEqual(gotPayload, payload) {
		t.Errorf("GetPayload mismatch:\nGot:    %#v\nWanted: %#v", gotPayload, payload)
	}
}

func TestToolEvent_Compose_OutputValidation(t *testing.T) {
	// This test acts as a contract, ensuring that any event produced by Compose
	// is always valid according to the shape provided by GetEventShape.
	interp := newEventTestInterpreter(t)
	composeTool, _ := interp.ToolRegistry().GetTool("ns_event.Compose")
	getShapeTool, _ := interp.ToolRegistry().GetTool("ns_event.GetEventShape")

	// 1. Get the canonical shape
	shapeDef, err := getShapeTool.Func(interp, []interface{}{})
	if err != nil {
		t.Fatalf("GetEventShape failed: %v", err)
	}
	shape, err := json_lite.ParseShape(shapeDef.(map[string]interface{}))
	if err != nil {
		t.Fatalf("Failed to parse the canonical shape: %v", err)
	}

	// 2. Compose a standard event
	payload := map[string]interface{}{"data": "value"}
	composedEvent, err := composeTool.Func(interp, []interface{}{"test.kind", payload})
	if err != nil {
		t.Fatalf("Compose failed: %v", err)
	}

	// 3. Validate the composed event against the canonical shape
	// Use strict validation (no extra fields in the envelope).
	err = shape.Validate(composedEvent, &json_lite.ValidateOptions{AllowExtra: false})
	if err != nil {
		t.Errorf("The event created by Compose failed validation against the canonical shape: %v", err)
	}
}

func TestToolEvent_GetPayload(t *testing.T) {
	tests := []eventTestCase{
		{
			name:     "Success: Extract valid payload",
			toolName: "GetPayload",
			args:     []interface{}{newValidFDMEvent()},
			wantResult: map[string]interface{}{
				"target":  "agent:did:zadeh:ping",
				"ping_id": "ping-abc-123",
			},
		},
		{
			name:       "Edge Case: Input is not a map",
			toolName:   "GetPayload",
			args:       []interface{}{"not a map"},
			wantResult: lang.NewMapValue(nil),
		},
		{
			name:       "Edge Case: Missing 'payload' key",
			toolName:   "GetPayload",
			args:       []interface{}{map[string]interface{}{"data": "value"}},
			wantResult: lang.NewMapValue(nil),
		},
		{
			name:       "Edge Case: 'payload' is not a list",
			toolName:   "GetPayload",
			args:       []interface{}{map[string]interface{}{"payload": "not a list"}},
			wantResult: lang.NewMapValue(nil),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testEventToolHelper(t, tt)
		})
	}
}

func TestToolEvent_GetAllPayloads(t *testing.T) {
	tests := []eventTestCase{
		{
			name:     "Success: Extract multiple valid payloads, skipping invalid ones",
			toolName: "GetAllPayloads",
			args:     []interface{}{newCoalescedFDMEvent()},
			wantResult: []interface{}{
				map[string]interface{}{"log_entry": "User logged in"},
				map[string]interface{}{"log_entry": "User accessed file X"},
			},
		},
		{
			name:       "Success: Extract single payload",
			toolName:   "GetAllPayloads",
			args:       []interface{}{newValidFDMEvent()},
			wantResult: []interface{}{map[string]interface{}{"target": "agent:did:zadeh:ping", "ping_id": "ping-abc-123"}},
		},
		{
			name:       "Edge Case: Payload list is empty",
			toolName:   "GetAllPayloads",
			args:       []interface{}{map[string]interface{}{"payload": []interface{}{}}},
			wantResult: []interface{}{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testEventToolHelper(t, tt)
		})
	}
}

func TestToolEvent_Getters(t *testing.T) {
	tests := []eventTestCase{
		{
			name:       "Success: GetID",
			toolName:   "GetID",
			args:       []interface{}{newValidFDMEvent()},
			wantResult: "event-id-123",
		},
		{
			name:       "Success: GetKind",
			toolName:   "GetKind",
			args:       []interface{}{newValidFDMEvent()},
			wantResult: "start.ping",
		},
		{
			name:       "Success: GetTimestamp",
			toolName:   "GetTimestamp",
			args:       []interface{}{newValidFDMEvent()},
			wantResult: int64(1757196281627980800),
		},
		{
			name:       "Fail: GetID from invalid struct",
			toolName:   "GetID",
			args:       []interface{}{map[string]interface{}{"foo": "bar"}},
			wantResult: "",
		},
		{
			name:       "Fail: GetKind from invalid struct",
			toolName:   "GetKind",
			args:       []interface{}{map[string]interface{}{"foo": "bar"}},
			wantResult: "",
		},
		{
			name:       "Fail: GetTimestamp from invalid struct",
			toolName:   "GetTimestamp",
			args:       []interface{}{map[string]interface{}{"foo": "bar"}},
			wantResult: int64(0),
		},
		{
			name:     "Fail: GetID with wrong data type",
			toolName: "GetID",
			args: []interface{}{map[string]interface{}{"payload": []interface{}{
				map[string]interface{}{"ID": 123},
			}}},
			wantResult: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testEventToolHelper(t, tt)
		})
	}
}
