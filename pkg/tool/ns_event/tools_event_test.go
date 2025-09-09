// NeuroScript Version: 0.7.0
// File version: 7
// Purpose: A comprehensive test suite for the ns_event toolset, including validation, edge cases, and round-trip tests.
// filename: pkg/tool/ns_event/tools_event_test.go
// nlines: 325
// risk_rating: LOW
package ns_event_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/policy"
	toolnsevent "github.com/aprice2704/neuroscript/pkg/tool/ns_event"
	"github.com/aprice2704/neuroscript/pkg/types"
)

type eventTestCase struct {
	name          string
	toolName      types.ToolName
	args          []interface{}
	checkFunc     func(t *testing.T, result interface{}, err error)
	wantResult    interface{}
	wantToolErrIs error
}

func newEventTestInterpreter(t *testing.T) *interpreter.Interpreter {
	t.Helper()
	interp := interpreter.NewInterpreter(interpreter.WithExecPolicy(policy.AllowAll()))
	for _, toolImpl := range toolnsevent.EventToolsToRegister {
		if _, err := interp.ToolRegistry().RegisterTool(toolImpl); err != nil {
			t.Fatalf("Failed to register tool '%s': %v", toolImpl.Spec.Name, err)
		}
	}
	return interp
}

func testEventToolHelper(t *testing.T, tc eventTestCase) {
	t.Helper()
	interp := newEventTestInterpreter(t)
	fullname := types.MakeFullName(toolnsevent.Group, string(tc.toolName))
	toolImpl, found := interp.ToolRegistry().GetTool(fullname)
	if !found {
		t.Fatalf("Tool %q not found in registry", tc.toolName)
	}

	result, err := toolImpl.Func(interp, tc.args)

	if tc.checkFunc != nil {
		tc.checkFunc(t, result, err)
		return
	}

	if tc.wantToolErrIs != nil {
		if !errors.Is(err, tc.wantToolErrIs) {
			t.Errorf("Expected error wrapping [%v], but got: %v", tc.wantToolErrIs, err)
		}
	} else if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if err == nil {
		if !reflect.DeepEqual(result, tc.wantResult) {
			t.Errorf("Result mismatch.\nGot:    %#v\nWanted: %#v", result, tc.wantResult)
		}
	}
}

// --- Test Data Fixtures ---

func newValidFDMEvent() map[string]interface{} {
	return map[string]interface{}{
		"payload": []interface{}{
			map[string]interface{}{
				"ID":      "event-id-123",
				"Kind":    "start.ping",
				"AgentID": "did:zadeh:ping",
				"TS":      int64(1757196281627980800),
				"Payload": map[string]interface{}{
					"target":  "agent:did:zadeh:ping",
					"ping_id": "ping-abc-123",
				},
			},
		},
	}
}

func newCoalescedFDMEvent() map[string]interface{} {
	return map[string]interface{}{
		"payload": []interface{}{
			map[string]interface{}{
				"ID":      "event-id-123",
				"Payload": map[string]interface{}{"log_entry": "User logged in"},
			},
			map[string]interface{}{
				"ID":      "event-id-456",
				"Payload": map[string]interface{}{"log_entry": "User accessed file X"},
			},
			map[string]interface{}{
				"ID": "event-id-789", // No Payload key
			},
			map[string]interface{}{
				"ID":      "event-id-abc",
				"Payload": "not-a-map", // Payload is wrong type
			},
		},
	}
}

// --- Test Functions ---

func TestToolEvent_Compose(t *testing.T) {
	t.Run("Success: Compose with required args", func(t *testing.T) {
		tc := eventTestCase{
			toolName: "Compose",
			args:     []interface{}{"system.log", map[string]interface{}{"level": "info"}},
			checkFunc: func(t *testing.T, result interface{}, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				evt, ok := result.(map[string]interface{})
				if !ok {
					t.Fatalf("expected result to be a map, got %T", result)
				}
				p, ok := evt["payload"].([]interface{})
				if !ok || len(p) != 1 {
					t.Fatal("expected a single envelope in payload")
				}
				envelope, ok := p[0].(map[string]interface{})
				if !ok {
					t.Fatal("envelope is not a map")
				}
				if envelope["Kind"] != "system.log" {
					t.Errorf("expected kind 'system.log', got %v", envelope["Kind"])
				}
				if _, ok := envelope["ID"].(string); !ok || envelope["ID"] == "" {
					t.Error("expected a non-empty string ID to be generated")
				}
				if _, ok := envelope["TS"].(int64); !ok || envelope["TS"] == int64(0) {
					t.Error("expected a non-zero timestamp to be generated")
				}
				if !reflect.DeepEqual(envelope["Payload"], map[string]interface{}{"level": "info"}) {
					t.Errorf("payload mismatch")
				}
			},
		}
		testEventToolHelper(t, tc)
	})

	t.Run("Success: Compose with all args", func(t *testing.T) {
		tc := eventTestCase{
			toolName: "Compose",
			args:     []interface{}{"system.log", map[string]interface{}{"msg": "test"}, "my-id-123", "my-agent"},
			checkFunc: func(t *testing.T, result interface{}, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				envelope := result.(map[string]interface{})["payload"].([]interface{})[0].(map[string]interface{})
				if envelope["ID"] != "my-id-123" {
					t.Errorf("ID mismatch: expected my-id-123, got %v", envelope["ID"])
				}
				if envelope["AgentID"] != "my-agent" {
					t.Errorf("AgentID mismatch: expected my-agent, got %v", envelope["AgentID"])
				}
			},
		}
		testEventToolHelper(t, tc)
	})
}

func TestToolEvent_Compose_Failures(t *testing.T) {
	tests := []eventTestCase{
		{
			name:          "Fail: kind is not a string",
			toolName:      "Compose",
			args:          []interface{}{123, map[string]interface{}{}},
			wantToolErrIs: lang.ErrInvalidArgument,
		},
		{
			name:          "Fail: payload is not a map",
			toolName:      "Compose",
			args:          []interface{}{"kind.string", "not-a-map"},
			wantToolErrIs: lang.ErrInvalidArgument,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testEventToolHelper(t, tt)
		})
	}
}
