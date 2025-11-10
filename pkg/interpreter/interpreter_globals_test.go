// NeuroScript Version: 0.8.0
// File version: 9
// Purpose: Fixes parser error by changing '::' comments to '//'.
// filename: pkg/interpreter/interpreter_globals_test.go
// nlines: 161

package interpreter

import (
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/interfaces" // Need this for Load
	"github.com/aprice2704/neuroscript/pkg/lang"
)

// --- The Test ---

func TestGlobalSymbolWorkflow(t *testing.T) {
	var provider *mockSymbolProvider // This type is defined in interpreter_test_helpers.go

	// --- Step 1: Config Context (Write Path) ---
	t.Run("Step 1: Config Context - Load and Extract Symbols", func(t *testing.T) {
		// FIX: Change comment syntax from '::' to '//' to fix parser error.
		const declScript = `
            // a config proc
            func config_func() means
                emit "from config_func"
            endfunc

            // a config event
            on event "config_event" as config_handler do
                emit "from config_event"
            endon
        `
		const cmdScript = `
            command
                must tool.ns.def_global_const("CONFIG_CONST_INT", 123)
                must tool.ns.def_global_const("CONFIG_CONST_STR", "hello_provider")
            endcommand
        `
		configInterp := newConfigInterpreter(t)

		// Load all declarations in one call.
		mustLoadString(t, configInterp, declScript)

		// Manually parse, build, and inject the commands.
		tree, pErr := configInterp.Parser().Parse(cmdScript) //
		if pErr != nil {
			t.Fatalf("Failed to parse command script: %v", pErr)
		}
		program, _, bErr := configInterp.ASTBuilder().Build(tree) //
		if bErr != nil {
			t.Fatalf("Failed to build command AST: %v", bErr)
		}
		configInterp.state.commands = program.Commands

		// Run commands to define constants
		_, err := configInterp.ExecuteCommands() //
		if err != nil {
			t.Fatalf("Failed to execute config commands: %v", err)
		}

		// Use the "Provenance API" to extract *local* symbols
		localProcs := configInterp.KnownProcedures()       //
		localHandlers := configInterp.KnownEventHandlers() //
		localConsts := configInterp.KnownGlobalConstants() //

		// Verify extraction
		if len(localProcs) != 1 || localProcs["config_func"] == nil {
			t.Errorf("Expected 1 local proc, got %d", len(localProcs))
		}
		if len(localHandlers) != 1 || localHandlers["config_event"] == nil {
			t.Errorf("Expected 1 local handler, got %d", len(localHandlers))
		}
		if len(localConsts) != 2 || localConsts["CONFIG_CONST_INT"] == nil {
			t.Errorf("Expected 2 local consts, got %d", len(localConsts))
		}
		if val, _ := lang.ToString(localConsts["CONFIG_CONST_STR"]); val != "hello_provider" {
			t.Errorf("Constant value mismatch, got %v", localConsts["CONFIG_CONST_STR"])
		}

		// Create the provider for the runtime
		provider = newMockSymbolProvider(localProcs, localHandlers, localConsts)
	})

	// --- Step 2: Runtime Context (Read from Provider) ---
	t.Run("Step 2: Runtime Context - Read from Provider", func(t *testing.T) {
		if provider == nil {
			t.Skip("Skipping Step 2, provider not created in Step 1")
		}

		const runtimeScript = `
            func main(returns string) means
                return "consts are " + CONFIG_CONST_INT + " and " + CONFIG_CONST_STR
            endfunc

            func call_provider_func() means
                call config_func()
            endfunc
        `
		runtimeInterp := newRuntimeInterpreter(t, provider)
		mustLoadString(t, runtimeInterp, runtimeScript)

		// Verify that the runtime can execute and see provider symbols
		val, err := runtimeInterp.RunProcedure("main") //
		if err != nil {
			t.Fatalf("Runtime failed to execute 'main': %v", err)
		}
		if s, _ := lang.ToString(val); s != "consts are 123 and hello_provider" {
			t.Errorf("Runtime execution failed, wrong result: %s", s)
		}

		// Verify calling a provider procedure
		_, err = runtimeInterp.RunProcedure("call_provider_func") //
		if err != nil {
			t.Fatalf("Runtime failed to call provider proc 'config_func': %v", err)
		}
	})

	// --- Step 3: "No Override" Rule (Write Time) ---
	t.Run("Step 3: 'No Override' Rule - Config Time Collision", func(t *testing.T) {
		const declScript = `
            func foo() means
                return
            endfunc
        `
		const cmdScript = `
            command
                must tool.ns.def_global_const("foo", 123)
            endcommand
        `
		configInterp := newConfigInterpreter(t)
		mustLoadString(t, configInterp, declScript)

		// Manually parse, build, and inject the command
		tree, pErr := configInterp.Parser().Parse(cmdScript) //
		if pErr != nil {
			t.Fatalf("Failed to parse command script: %v", pErr)
		}
		program, _, bErr := configInterp.ASTBuilder().Build(tree) //
		if bErr != nil {
			t.Fatalf("Failed to build command AST: %v", bErr)
		}
		configInterp.state.commands = program.Commands

		// Run commands and expect the collision error
		_, err := configInterp.ExecuteCommands() //
		if err == nil {
			t.Fatal("Expected error when defining const 'foo' that collides with proc 'foo', but got nil")
		}
		if !strings.Contains(err.Error(), "already defined as a procedure") { //
			t.Errorf("Expected collision error, but got: %v", err)
		}
	})

	// --- Step 4: "No Override" Rule (Load Time) ---
	t.Run("Step 4: 'No Override' Rule - Runtime Load Collision", func(t *testing.T) {
		if provider == nil {
			t.Skip("Skipping Step 4, provider not created in Step 1")
		}

		const collisionScript = `
            func config_func() means
                emit "this should fail to load"
            endfunc
        `
		runtimeInterp := newRuntimeInterpreter(t, provider)

		// We can't use mustLoadString because we expect an error
		tree, pErr := runtimeInterp.Parser().Parse(collisionScript) //
		if pErr != nil {
			t.Fatalf("Parse failed: %v", pErr)
		}
		program, _, bErr := runtimeInterp.ASTBuilder().Build(tree) //
		if bErr != nil {
			t.Fatalf("AST build failed: %v", bErr)
		}

		err := runtimeInterp.Load(&interfaces.Tree{Root: program}) //
		if err == nil {
			t.Fatal("Expected error when loading script that collides with provider, but got nil")
		}
		if !strings.Contains(err.Error(), "provided by the host and cannot be overridden") { //
			t.Errorf("Expected override error, but got: %v", err)
		}
	})
}
