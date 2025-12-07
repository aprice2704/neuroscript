// :: product: FDM/NS
// :: majorVersion: 1
// :: fileVersion: 3
// :: description: Tests the routing and state preservation logic of LoadOrExecute.
// :: latestChange: Updated mixed script test to accept parser syntax errors as valid rejection.
// :: filename: pkg/api/exec_routing_test.go
// :: serialization: go

package api_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/api"
)

func TestLoadOrExecute_RoutingAndPersistence(t *testing.T) {
	// 1. Setup: Create a persistent interpreter.
	var stdout bytes.Buffer
	hc := newTestHostContext(nil)
	// Override stdout to capture output
	hc.Stdout = &stdout

	interp := api.New(api.WithHostContext(hc))

	// 2. Path A: Definitions Only (Library Load)
	// We load a function 'get_message'.
	libSource := `
func get_message() means
    return "persistence_is_working"
endfunc
`
	libTree, err := api.Parse([]byte(libSource), api.ParseSkipComments)
	if err != nil {
		t.Fatalf("Failed to parse library script: %v", err)
	}

	// EXECUTE PATH A
	// This should route to AppendScript, preserving the function.
	val, err := api.LoadOrExecute(context.Background(), interp, libTree)
	if err != nil {
		t.Fatalf("LoadOrExecute (Defs path) failed: %v", err)
	}
	if val != nil {
		t.Errorf("Expected nil result from Defs path, got %v", val)
	}

	// Verify state: The function should exist.
	// We use the public introspection API to check.
	if _, ok := interp.KnownProcedures()["get_message"]; !ok {
		t.Fatal("Function 'get_message' was not loaded into the interpreter.")
	}

	// 3. Path B: Commands Only (Execution)
	// We run a command that DEPENDS on the previously loaded function.
	// If LoadOrExecute wipes state (the bug), this will fail.
	cmdSource := `
command
    emit get_message()
endcommand
`
	cmdTree, err := api.Parse([]byte(cmdSource), api.ParseSkipComments)
	if err != nil {
		t.Fatalf("Failed to parse command script: %v", err)
	}

	// EXECUTE PATH B
	// This should route to AppendScript + Execute, PRESERVING the function.
	_, err = api.LoadOrExecute(context.Background(), interp, cmdTree)
	if err != nil {
		t.Fatalf("LoadOrExecute (Cmds path) failed: %v", err)
	}

	// Verify Result: Output should contain the string from the function.
	output := stdout.String()
	if !strings.Contains(output, "persistence_is_working") {
		t.Errorf("State lost? Expected output 'persistence_is_working', got: %q", output)
	}
}

func TestLoadOrExecute_MixedScriptFailure(t *testing.T) {
	interp := api.New(api.WithHostContext(newTestHostContext(nil)))

	// A mixed script that defines a function and has a command block.
	// The NeuroScript grammar currently rejects this at the parse level.
	src := `
func f() means
    return
endfunc

command
    emit "bad"
endcommand
`
	tree, err := api.Parse([]byte(src), api.ParseSkipComments)

	// Case 1: Parser rejects it (Current behavior)
	if err != nil {
		t.Logf("Success: Mixed script rejected by parser: %v", err)
		return
	}

	// Case 2: Parser allows it (Future proofing), Router must reject it.
	_, err = api.LoadOrExecute(context.Background(), interp, tree)
	if err == nil {
		t.Fatal("Expected error for mixed script (from Router), got nil")
	}
	if !strings.Contains(err.Error(), "mixed script detected") {
		t.Errorf("Expected 'mixed script' error, got: %v", err)
	}
}
