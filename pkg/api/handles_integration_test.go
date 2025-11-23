// NeuroScript Version: 1
// File version: 3
// Purpose: Integration tests for the opaque object handle system via the public API. Fixed script syntax: removed 'returns' clause from function signatures to satisfy parser expectation of 'means'.
// filename: pkg/api/handles_integration_test.go
// nlines: 215

package api_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/api"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

// testPayload represents a host-side object we want to manage via handles.
type testPayload struct {
	ID    int
	Value string
}

func TestHandles_Integration(t *testing.T) {
	// Setup: Define tools that act as Producer and Consumer of handles.

	// Producer: Creates a testPayload and returns a handle to it.
	// Signature: tool.test.create_resource(id: int, val: string) -> handle
	createTool := api.ToolImplementation{
		Spec: api.ToolSpec{
			Name:  "create_resource",
			Group: "test",
			Args: []api.ArgSpec{
				{Name: "id", Type: "int"},
				{Name: "val", Type: "string"},
			},
		},
		Func: func(rt api.Runtime, args []any) (any, error) {
			id := int(args[0].(int64)) // api unmarshals numbers as int64 or float64 usually, let's assume int64 for safe casting from "int" arg type
			val := args[1].(string)

			payload := &testPayload{ID: id, Value: val}

			// Use the new HandleRegistry API
			hv, err := rt.HandleRegistry().NewHandle(payload, "test_resource")
			if err != nil {
				return nil, fmt.Errorf("failed to create handle: %w", err)
			}

			// DEBUG
			fmt.Printf("[DEBUG Producer] Created handle: %s (Kind: %s)\n", hv.HandleID(), hv.HandleKind())

			return hv, nil
		},
	}

	// Consumer: Takes a handle, verifies kind, and returns the internal value.
	// Signature: tool.test.read_resource(h: handle) -> string
	readTool := api.ToolImplementation{
		Spec: api.ToolSpec{
			Name:  "read_resource",
			Group: "test",
			Args: []api.ArgSpec{
				{Name: "h", Type: "any"}, // Accepting 'any' to check type manually for the test
			},
		},
		Func: func(rt api.Runtime, args []any) (any, error) {
			hv, ok := args[0].(api.HandleValue)
			if !ok {
				return nil, fmt.Errorf("arg 0 is not a handle, got %T", args[0])
			}

			// 1. Check Kind
			if hv.HandleKind() != "test_resource" {
				return nil, fmt.Errorf("expected kind 'test_resource', got '%s'", hv.HandleKind())
			}

			// 2. Resolve
			obj, err := rt.HandleRegistry().GetHandle(hv.HandleID())
			if err != nil {
				return nil, fmt.Errorf("failed to resolve handle: %w", err)
			}

			// 3. Cast
			payload, ok := obj.(*testPayload)
			if !ok {
				return nil, fmt.Errorf("payload type mismatch: expected *testPayload, got %T", obj)
			}

			// DEBUG
			fmt.Printf("[DEBUG Consumer] Resolved handle %s to payload ID=%d\n", hv.HandleID(), payload.ID)

			return payload.Value, nil
		},
	}

	// Setup Interpreter Helper
	setupInterpreter := func(t *testing.T) *api.Interpreter {
		hc, err := api.NewHostContextBuilder().
			WithLogger(api.NewNoOpLogger()).
			WithStdout(io.Discard).
			WithStderr(os.Stderr). // Keep stderr for debug
			WithStdin(os.Stdin).
			Build()
		if err != nil {
			t.Fatalf("Failed to build host context: %v", err)
		}

		// Allow our test tools
		policy := api.NewPolicyBuilder(api.ContextNormal).
			Allow("tool.test.create_resource").
			Allow("tool.test.read_resource").
			Build()

		interp := api.New(
			api.WithHostContext(hc),
			api.WithExecPolicy(policy),
		)

		// Register tools
		interp.ToolRegistry().RegisterTool(createTool)
		interp.ToolRegistry().RegisterTool(readTool)

		return interp
	}

	// --- Test Cases ---

	t.Run("End-to-End: Create, Pass, and Read Handle", func(t *testing.T) {
		interp := setupInterpreter(t)
		ctx := context.Background()

		// Script: Removed 'returns result' from signature.
		script := `
		func test_e2e() means
			set h = tool.test.create_resource(101, "Integration Success")
			set result = tool.test.read_resource(h)
			return result
		endfunc
		`

		tree, err := api.Parse([]byte(script), api.ParseSkipComments)
		if err != nil {
			t.Fatalf("Parse failed: %v", err)
		}

		// Loading the definition
		_, err = api.ExecWithInterpreter(ctx, interp, tree)
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}

		// Running the procedure
		result, err := api.RunProcedure(ctx, interp, "test_e2e")
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		// Unwrap and verify
		strRes, err := api.Unwrap(result)
		if err != nil {
			t.Fatalf("Unwrap failed: %v", err)
		}

		if strRes != "Integration Success" {
			t.Errorf("Expected 'Integration Success', got '%v'", strRes)
		}
	})

	t.Run("Safety: Handle Kind Mismatch", func(t *testing.T) {
		interp := setupInterpreter(t)
		ctx := context.Background()

		// Manually create a handle of the WRONG kind and inject it via a different mechanism
		badProducer := api.ToolImplementation{
			Spec: api.ToolSpec{Name: "create_wrong_kind", Group: "test"},
			Func: func(rt api.Runtime, args []any) (any, error) {
				return rt.HandleRegistry().NewHandle(&testPayload{ID: 999}, "wrong_kind")
			},
		}
		interp.ToolRegistry().RegisterTool(badProducer)
		interp.GetExecPolicy().Allow = append(interp.GetExecPolicy().Allow, "tool.test.create_wrong_kind")

		// Script: Using 'call' for bare tool execution (statement-only language).
		script := `
		func test_safety() means
			set h = tool.test.create_wrong_kind()
			call tool.test.read_resource(h)
		endfunc
		`
		tree, err := api.Parse([]byte(script), api.ParseSkipComments)
		if err != nil {
			t.Fatalf("Parse failed: %v", err)
		}

		_, err = api.ExecWithInterpreter(ctx, interp, tree)
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}

		_, err = api.RunProcedure(ctx, interp, "test_safety")
		if err == nil {
			t.Fatal("Expected error due to kind mismatch, got success")
		}

		if !contains(err.Error(), "expected kind 'test_resource', got 'wrong_kind'") {
			t.Errorf("Error message did not match expected. Got: %v", err)
		}
	})

	t.Run("Lifecycle: Handle Deletion", func(t *testing.T) {
		interp := setupInterpreter(t)
		ctx := context.Background()

		// 1. Create Handle via script (Removed 'returns h' from signature)
		scriptCreate := `
		func make_handle() means
			set h = tool.test.create_resource(202, "To Be Deleted")
			return h
		endfunc
		`
		treeCreate, err := api.Parse([]byte(scriptCreate), api.ParseSkipComments)
		if err != nil {
			t.Fatalf("Parse failed: %v", err)
		}
		_, err = api.ExecWithInterpreter(ctx, interp, treeCreate)
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}

		resVal, err := api.RunProcedure(ctx, interp, "make_handle")
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		hv, ok := resVal.(api.HandleValue)
		if !ok {
			t.Fatalf("Expected api.HandleValue from return, got %T", resVal)
		}

		id := hv.HandleID()
		fmt.Printf("[DEBUG Lifecycle] Handle ID to delete: %s\n", id)

		// 2. Verify it exists
		if _, err := interp.HandleRegistry().GetHandle(id); err != nil {
			t.Fatalf("Handle should exist but GetHandle failed: %v", err)
		}

		// 3. Delete it (Host side operation)
		if err := interp.HandleRegistry().DeleteHandle(id); err != nil {
			t.Fatalf("DeleteHandle failed: %v", err)
		}

		// 4. Verify it is gone
		_, err = interp.HandleRegistry().GetHandle(id)
		if err == nil {
			t.Fatal("GetHandle succeeded after deletion, expected error")
		}
		if !errors.Is(err, lang.ErrHandleNotFound) {
			t.Errorf("Expected ErrHandleNotFound, got: %v", err)
		}

		// 5. Verify script access fails
		_, err = readTool.Func(interp, []any{hv})
		if err == nil {
			t.Fatal("Tool succeeded with deleted handle, expected failure")
		}
		if !contains(err.Error(), "failed to resolve handle") {
			t.Errorf("Expected resolution error, got: %v", err)
		}
	})
}

// simple helper for string check
func contains(s, substr string) bool {
	return len(s) >= len(substr) && search(s, substr)
}

func search(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
