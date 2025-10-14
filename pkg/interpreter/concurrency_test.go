// NeuroScript Version: 0.8.0
// File version: 2
// Purpose: Corrected test to set global variables directly from Go to accurately test concurrency.
// filename: pkg/interpreter/concurrency_test.go
// nlines: 75
// risk_rating: HIGH

package interpreter_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/lang"
)

func TestInterpreter_Concurrency(t *testing.T) {
	t.Run("Concurrent writes to different global variables", func(t *testing.T) {
		h := NewTestHarness(t)
		interp := h.Interpreter

		var wg sync.WaitGroup
		numGoroutines := 50
		wg.Add(numGoroutines)

		// Each goroutine will write to a *different* global variable directly
		// using the interpreter's Go API. This correctly tests the thread safety
		// of the underlying variable map without being blocked by the sandbox.
		for i := 0; i < numGoroutines; i++ {
			go func(n int) {
				defer wg.Done()
				varName := fmt.Sprintf("global_%d", n)
				val := float64(n * 10)

				// Use SetInitialVariable as it correctly registers the variable as global.
				// This simulates multiple external events or setup steps happening concurrently.
				err := interp.SetInitialVariable(varName, val)
				if err != nil {
					t.Errorf("Concurrent execution failed for goroutine %d: %v", n, err)
				}
			}(i)
		}

		wg.Wait()

		// Verify that all globals were set correctly.
		for i := 0; i < numGoroutines; i++ {
			varName := fmt.Sprintf("global_%d", i)
			expectedVal := lang.NumberValue{Value: float64(i * 10)}

			val, exists := interp.GetVariable(varName)
			if !exists {
				t.Errorf("Global variable '%s' was not set.", varName)
				continue
			}
			if val != expectedVal {
				t.Errorf("Mismatch for global '%s'. Got: %#v, Want: %#v", varName, val, expectedVal)
			}
		}
	})
}
