// NeuroScript Version: 0.8.0
// File version: 1
// Purpose: Adds advanced tests for symbol precedence and case-sensitivity.
// filename: pkg/interpreter/interpreter_globals_adv_test.go
// nlines: 106

package interpreter

import (
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/lang"
)

// --- New Test: Symbol Precedence ---

func TestSymbolPrecedence(t *testing.T) {
	// 1. Create a config interpreter to define a provider constant
	const declScript = `
        func dummy() means
			return
        endfunc
    `
	const cmdScript = `
        command
            must tool.ns.def_global_const("MY_VAL", 100)
        endcommand
    `
	configInterp := newConfigInterpreter(t)
	mustLoadString(t, configInterp, declScript)

	// Manually parse, build, and inject the commands
	tree, pErr := configInterp.Parser().Parse(cmdScript)
	if pErr != nil {
		t.Fatalf("Failed to parse command script: %v", pErr)
	}
	program, _, bErr := configInterp.ASTBuilder().Build(tree)
	if bErr != nil {
		t.Fatalf("Failed to build command AST: %v", bErr)
	}
	configInterp.state.commands = program.Commands
	if _, err := configInterp.ExecuteCommands(); err != nil {
		t.Fatalf("Config commands failed: %v", err)
	}

	// 2. Create the provider
	provider := newMockSymbolProvider(
		configInterp.KnownProcedures(),
		configInterp.KnownEventHandlers(),
		configInterp.KnownGlobalConstants(),
	)

	// 3. Create a runtime interpreter
	const runtimeScript = `
        func check_precedence(returns number) means
            set MY_VAL = 10 // This should create a *local* variable
            return MY_VAL
        endfunc
    `
	runtimeInterp := newRuntimeInterpreter(t, provider)
	mustLoadString(t, runtimeInterp, runtimeScript)

	// 4. Run the test
	// First, check that the provider const is visible
	val, exists := runtimeInterp.GetVariable("MY_VAL")
	if !exists {
		t.Fatal("Provider constant 'MY_VAL' not found")
	}
	// FIX: Use lang.ToFloat64 instead of lang.ToNumber
	if n, _ := lang.ToFloat64(val); n != 100 {
		t.Fatalf("Provider constant 'MY_VAL' has wrong value, expected 100, got %v", val)
	}

	// Now, run the function that creates a local shadow
	result, err := runtimeInterp.RunProcedure("check_precedence")
	if err != nil {
		t.Fatalf("check_precedence failed: %v", err)
	}
	// FIX: Use lang.ToFloat64 instead of lang.ToNumber
	if n, _ := lang.ToFloat64(result); n != 10 {
		t.Fatalf("Function should have returned local '10', but got %v", result)
	}

	// 5. Check the variable *after* the run
	// This proves the local 'set' did not overwrite the provider's constant
	val, exists = runtimeInterp.GetVariable("MY_VAL")
	if !exists {
		t.Fatal("Provider constant 'MY_VAL' disappeared after run")
	}
	// FIX: Use lang.ToFloat64 instead of lang.ToNumber
	if n, _ := lang.ToFloat64(val); n != 100 {
		t.Fatalf("Provider constant 'MY_VAL' was overwritten, expected 100, got %v", val)
	}
}

// --- New Test: Case Sensitivity ---

func TestSymbolCaseSensitivity(t *testing.T) {
	// 1. Create a provider with a lowercase procedure
	const declScript = `
        func config_func() means
            return "lowercase"
        endfunc
    `
	configInterp := newConfigInterpreter(t)
	mustLoadString(t, configInterp, declScript)
	provider := newMockSymbolProvider(
		configInterp.KnownProcedures(),
		configInterp.KnownEventHandlers(),
		configInterp.KnownGlobalConstants(),
	)

	// 2. Create a runtime interpreter
	const runtimeScript = `
        func main() means
            // Try to call the uppercase version
            call CONFIG_FUNC()
        endfunc
    `
	runtimeInterp := newRuntimeInterpreter(t, provider)
	mustLoadString(t, runtimeInterp, runtimeScript)

	// 3. Run and expect a "not found" error
	_, err := runtimeInterp.RunProcedure("main")
	if err == nil {
		t.Fatal("Expected procedure call to fail due to case-sensitivity, but it succeeded.")
	}

	// Check for the *correct* error
	if !strings.Contains(err.Error(), "procedure 'CONFIG_FUNC' not found") {
		t.Errorf("Expected 'procedure not found' error, but got: %v", err)
	}
}
