// NeuroScript Version: 0.8.0
// File version: 2
// Purpose: Tests ExecuteSandboxedAST. Removes duplicate mockLogger.
// filename: pkg/api/exec_2_test.go
// nlines: 88

package api_test

import (
	"context"
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/api"
	"github.com/aprice2704/neuroscript/pkg/lang"
	// "github.com/aprice2704/neuroscript/pkg/logging" // No longer needed
)

// FIX: Removed duplicate mockLogger struct.
// We assume 'mockLogger' is defined in another '*_test.go' file
// in this package (like interpreter_test.go or exec_identity_test.go).

// FIX: Removed newTestInterpreter helper.
// We will create the interpreter inline in each test,
// using the existing mockLogger.

func TestExecuteSandboxedAST_Success(t *testing.T) {
	// --- ARRANGE ---
	script := `
	command
		set x = "this variable is local"
		emit "hello emit"
		whisper "my_handle", "hello whisper"
		emit "second emit"
	endcommand
	`
	tree, err := api.Parse([]byte(script), api.ParseSkipComments)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Create interpreter using the existing mockLogger
	hc, err := api.NewHostContextBuilder().
		WithLogger(&mockLogger{}).
		WithStdout(os.Stdout).
		WithStdin(os.Stdin).
		WithStderr(os.Stderr).
		Build()
	if err != nil {
		t.Fatalf("Failed to build host context: %v", err)
	}
	interp := api.New(api.WithHostContext(hc))

	// --- ACT ---
	emits, whispers, execErr := api.ExecuteSandboxedAST(interp, tree, context.Background())

	// --- ASSERT ---
	if execErr != nil {
		t.Fatalf("ExecuteSandboxedAST failed unexpectedly: %v", execErr)
	}

	// Check emits
	expectedEmits := []string{"hello emit", "second emit"}
	if !reflect.DeepEqual(emits, expectedEmits) {
		t.Errorf("Emits mismatch:\n  Got: %v\n  Want: %v", emits, expectedEmits)
	}

	// Check whispers
	expectedWhispers := map[string]api.Value{
		"my_handle": lang.StringValue{Value: "hello whisper"},
	}
	if !reflect.DeepEqual(whispers, expectedWhispers) {
		t.Errorf("Whispers mismatch:\n  Got: %v\n  Want: %v", whispers, expectedWhispers)
	}

	t.Log("SUCCESS: ExecuteSandboxedAST correctly captured emits and whispers.")
}

func TestExecuteSandboxedAST_ExecError(t *testing.T) {
	// --- ARRANGE ---
	script := `
	command
		emit "this will run"
		fail "a test error"
		emit "this will not run"
	endcommand
	`
	tree, err := api.Parse([]byte(script), api.ParseSkipComments)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Create interpreter using the existing mockLogger
	hc, err := api.NewHostContextBuilder().
		WithLogger(&mockLogger{}).
		WithStdout(os.Stdout).
		WithStdin(os.Stdin).
		WithStderr(os.Stderr).
		Build()
	if err != nil {
		t.Fatalf("Failed to build host context: %v", err)
	}
	interp := api.New(api.WithHostContext(hc))

	// --- ACT ---
	emits, whispers, execErr := api.ExecuteSandboxedAST(interp, tree, context.Background())

	// --- ASSERT ---
	if execErr == nil {
		t.Fatal("ExecuteSandboxedAST succeeded, but was expected to fail.")
	}

	// Check that the error is the correct type and message
	var rtErr *api.RuntimeError
	if !errors.As(execErr, &rtErr) {
		t.Fatalf("Expected a *lang.RuntimeError, but got %T", execErr)
	}
	if !strings.Contains(rtErr.Error(), "a test error") {
		t.Errorf("Error message mismatch:\n  Got: %v\n  Want: %v", rtErr.Error(), "a test error")
	}

	// Check that partial emits were still captured
	expectedEmits := []string{"this will run"}
	if !reflect.DeepEqual(emits, expectedEmits) {
		t.Errorf("Emits mismatch:expected partial emits before fail:\n  Got: %v\n  Want: %v", emits, expectedEmits)
	}

	// Check that whispers (which didn't happen) are empty
	if len(whispers) != 0 {
		t.Errorf("Expected empty whispers map, but got %v", whispers)
	}

	t.Logf("SUCCESS: ExecuteSandboxedAST correctly captured partial emits and returned error: %v", execErr)
}
