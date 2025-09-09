// NeuroScript Version: 0.7.0
// File version: 2
// Purpose: Tests for pre-defined shapes and convenience functions.
// filename: pkg/api/shape/predefined_test.go
// nlines: 75
// risk_rating: LOW

package shape_test

import (
	"reflect"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/api/shape"
)

func TestValidateNSEvent(t *testing.T) {
	validEvent, err := shape.ComposeNSEvent("test.kind", map[string]interface{}{"data": "value"}, nil)
	if err != nil {
		t.Fatalf("ComposeNSEvent failed during test setup: %v", err)
	}

	invalidEvent := map[string]interface{}{
		"payload": []interface{}{
			map[string]interface{}{"ID": "123"}, // Missing Kind and TS
		},
	}

	t.Run("valid event passes", func(t *testing.T) {
		if err := shape.ValidateNSEvent(validEvent, nil); err != nil {
			t.Errorf("expected valid event to pass validation, but got: %v", err)
		}
	})

	t.Run("invalid event fails", func(t *testing.T) {
		if err := shape.ValidateNSEvent(invalidEvent, nil); err == nil {
			t.Error("expected invalid event to fail validation, but it passed")
		}
	})
}

func TestComposeNSEvent(t *testing.T) {
	t.Run("composes with minimal arguments", func(t *testing.T) {
		payload := map[string]interface{}{"data": "value"}
		event, err := shape.ComposeNSEvent("test.kind", payload, nil)
		if err != nil {
			t.Fatalf("ComposeNSEvent failed unexpectedly: %v", err)
		}
		if err := shape.ValidateNSEvent(event, nil); err != nil {
			t.Fatalf("Composed event failed validation: %v", err)
		}

		// Check payload integrity
		path, err := shape.ParsePath("payload[0].Payload")
		if err != nil {
			t.Fatalf("Failed to parse path: %v", err)
		}
		p, _ := shape.Select(event, path, nil)
		if !reflect.DeepEqual(p, payload) {
			t.Errorf("payload was not set correctly")
		}
	})

	t.Run("composes with all options", func(t *testing.T) {
		opts := &shape.NSEventComposeOptions{
			ID:      "my-custom-id",
			AgentID: "my-agent",
		}
		event, err := shape.ComposeNSEvent("test.kind", nil, opts)
		if err != nil {
			t.Fatalf("ComposeNSEvent failed unexpectedly: %v", err)
		}

		idPath, _ := shape.ParsePath("payload[0].ID")
		id, _ := shape.Select(event, idPath, nil)
		if id != "my-custom-id" {
			t.Errorf("expected ID to be 'my-custom-id', got '%s'", id)
		}

		agentIDPath, _ := shape.ParsePath("payload[0].AgentID")
		agentID, _ := shape.Select(event, agentIDPath, nil)
		if agentID != "my-agent" {
			t.Errorf("expected AgentID to be 'my-agent', got '%s'", agentID)
		}
	})
}
