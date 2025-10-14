// NeuroScript Version: 0.8.0
// File version: 1
// Purpose: Adds comprehensive tests for the interpreter's object handle management system.
// filename: pkg/interpreter/handles_test.go
// nlines: 125
// risk_rating: LOW

package interpreter_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/lang"
)

// A dummy struct to represent a complex, non-serializable object.
type mockResource struct {
	ID   string
	Data string
}

func TestInterpreter_HandleManagement(t *testing.T) {
	t.Run("Successful registration and retrieval", func(t *testing.T) {
		h := NewTestHarness(t)
		interp := h.Interpreter

		resource := &mockResource{ID: "res-123", Data: "live object"}
		handle, err := interp.RegisterHandle(resource, "mock_resource")
		if err != nil {
			t.Fatalf("RegisterHandle() failed unexpectedly: %v", err)
		}

		if !strings.HasPrefix(handle, "mock_resource::") {
			t.Errorf("Expected handle to have prefix 'mock_resource::', got '%s'", handle)
		}

		retrieved, err := interp.GetHandleValue(handle, "mock_resource")
		if err != nil {
			t.Fatalf("GetHandleValue() failed unexpectedly: %v", err)
		}

		retrievedResource, ok := retrieved.(*mockResource)
		if !ok {
			t.Fatalf("Retrieved object has wrong type. Got %T, want *mockResource", retrieved)
		}

		if retrievedResource.ID != "res-123" {
			t.Errorf("Retrieved object data mismatch. Got ID '%s'", retrievedResource.ID)
		}
	})

	t.Run("Get handle with wrong type prefix", func(t *testing.T) {
		h := NewTestHarness(t)
		interp := h.Interpreter

		resource := &mockResource{ID: "res-456", Data: "another object"}
		handle, _ := interp.RegisterHandle(resource, "correct_type")

		_, err := interp.GetHandleValue(handle, "wrong_type")
		if err == nil {
			t.Fatal("Expected an error when getting handle with wrong type, but got nil")
		}

		if !errors.Is(err, lang.ErrHandleWrongType) {
			t.Errorf("Expected error to wrap ErrHandleWrongType, but got: %v", err)
		}
	})

	t.Run("Get non-existent handle", func(t *testing.T) {
		h := NewTestHarness(t)
		interp := h.Interpreter

		_, err := interp.GetHandleValue("non_existent::handle", "any_type")
		if err == nil {
			t.Fatal("Expected an error for a non-existent handle, but got nil")
		}
		if !errors.Is(err, lang.ErrHandleNotFound) {
			t.Errorf("Expected error to wrap ErrHandleNotFound, but got: %v", err)
		}
	})

	t.Run("Remove handle", func(t *testing.T) {
		h := NewTestHarness(t)
		interp := h.Interpreter

		resource := &mockResource{ID: "res-789", Data: "to be removed"}
		handle, _ := interp.RegisterHandle(resource, "temp_resource")

		wasRemoved := interp.RemoveHandle(handle)
		if !wasRemoved {
			t.Error("RemoveHandle() returned false for an existing handle")
		}

		// Verify it's gone
		_, err := interp.GetHandleValue(handle, "temp_resource")
		if !errors.Is(err, lang.ErrHandleNotFound) {
			t.Errorf("Expected ErrHandleNotFound after removal, but got: %v", err)
		}

		// Test removing a non-existent handle
		wasRemoved = interp.RemoveHandle("non_existent::handle")
		if wasRemoved {
			t.Error("RemoveHandle() returned true for a non-existent handle")
		}
	})

	t.Run("Invalid handle format errors", func(t *testing.T) {
		h := NewTestHarness(t)
		interp := h.Interpreter

		testCases := []struct {
			name   string
			handle string
		}{
			{"Empty Handle", ""},
			{"No Separator", "justonepart"},
			{"Empty Prefix", "::12345"},
			{"Empty ID", "prefix::"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				_, err := interp.GetHandleValue(tc.handle, "any")
				if !errors.Is(err, lang.ErrInvalidArgument) {
					t.Errorf("Expected ErrInvalidArgument for handle '%s', but got: %v", tc.handle, err)
				}
			})
		}
	})
}
