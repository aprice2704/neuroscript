// NeuroScript Version: 0.3.0
// Last Modified: 2025-07-04 02:17:00 EDT
// filename: pkg/tool/gotools/tools_go_diagnostics_test.go

package gotools

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/testutil"
)

// AssertNoError is a test helper (now assumed to exist in testutil)
// If it doesn't exist, replace AssertNoError(t, err) with:
// if err != nil { t.Fatalf("unexpected error: %v", err) }

func TestGoDiagnosticTools(t *testing.T) {
	// Helper function to check map keys
	checkResultMapKeys := func(t *testing.T, resultMap interface{}, toolName string) {
		t.Helper()
		m, ok := resultMap.(map[string]interface{})
		if !ok {
			t.Fatalf("Tool %s did not return a map, got %T", toolName, resultMap)
		}
		expectedKeys := []string{"stdout", "stderr", "exit_code", "success"}
		for _, key := range expectedKeys {
			if _, exists := m[key]; !exists {
				t.Errorf("Tool %s result map missing expected key: %q", toolName, key)
			}
		}
	}

	t.Run("GoVetInvocation", func(t *testing.T) {
		// Setup interpreter with a temporary sandbox
		// *** Fix: Correctly handle (*Interpreter, error) return ***
		interpreter, err := testutil.NewTestInterpreter(t, nil, nil)
		if err != nil {
			t.Fatalf("NewTestInterpreter failed: %v", err)
		}
		sandboxAbsPath := interpreter.SandboxDir()
		t.Logf("Test interpreter created with sandbox: %s", sandboxAbsPath) // Log path (optional)

		// Args for the tool (using default target "./...")
		args := []interface{}{} // No target specified, should default

		// Call the tool function
		resultMap, toolErr := toolGoVet(interpreter, args) // toolErr is type error

		// --- Assertions ---
		// 1. Check for Go-level errors from the tool function itself (toolErr)
		if toolErr != nil {
			t.Fatalf("unexpected error: %v", toolErr)
		}

		// 2. Check if the result is a map with the expected keys
		checkResultMapKeys(t, resultMap, "GoVet")

		// Optional: Log the actual result for debugging if needed
		t.Logf("GoVet result: %+v", resultMap)
	})

	t.Run("StaticcheckInvocation", func(t *testing.T) {
		// Setup interpreter with a temporary sandbox
		// *** Fix: Correctly handle (*Interpreter, error) return ***
		interpreter, err := testutil.NewTestInterpreter(t, nil, nil)
		if err != nil {
			t.Fatalf("NewTestInterpreter failed: %v", err)
		}
		sandboxAbsPath := interpreter.SandboxDir()
		t.Logf("Test interpreter created with sandbox: %s", sandboxAbsPath) // Log path (optional)

		// Args for the tool (using default target "./...")
		args := []interface{}{} // No target specified, should default

		// Call the tool function
		resultMap, toolErr := toolStaticcheck(interpreter, args) // toolErr is type error

		// --- Assertions ---
		// 1. Check for Go-level errors from the tool function itself (toolErr)
		if toolErr != nil {
			t.Fatalf("unexpected error: %v", toolErr)
		}

		// 2. Check if the result is a map with the expected keys
		checkResultMapKeys(t, resultMap, "Staticcheck")

		// Optional: Log the actual result (especially stderr/exit code if staticcheck isn't installed)
		t.Logf("Staticcheck result: %+v", resultMap)
	})
}
