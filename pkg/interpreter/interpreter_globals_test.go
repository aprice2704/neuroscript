// :: product: FDM/NS
// :: majorVersion: 1
// :: fileVersion: 10
// :: description: Added TestGlobalConstantRedefinition to verify tool.ns.def_global_const respects WithAllowRedefinition.
// :: latestChange: Added test cases for global constant redefinition.
// :: filename: pkg/interpreter/interpreter_globals_test.go
// :: serialization: go

package interpreter

import (
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/interfaces" // Need this for Load
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/policy" // Added for manual policy construction
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

// TestGlobalConstantRedefinition verifies that tool.ns.def_global_const respects the AllowRedefinition flag.
func TestGlobalConstantRedefinition(t *testing.T) {
	t.Run("Allows redefining global constant with WithAllowRedefinition(true)", func(t *testing.T) {
		// 1. Create a config interpreter with redefinition enabled
		hc := newTestHostContext(t, nil)
		// Manually build policy as def_global_const requires ContextConfig
		cfgPolicy := policy.NewBuilder(policy.ContextConfig).Allow("*").Build()
		interp := NewInterpreter(
			WithHostContext(hc),
			WithExecPolicy(cfgPolicy),
			WithAllowRedefinition(true), // <--- The feature under test
		)

		// 2. Define the constant initially
		cmdScript1 := `
			command
				must tool.ns.def_global_const("MY_CONST", 100)
			endcommand
		`
		tree1, _ := interp.Parser().Parse(cmdScript1)
		prog1, _, _ := interp.ASTBuilder().Build(tree1)
		interp.state.commands = prog1.Commands
		if _, err := interp.ExecuteCommands(); err != nil {
			t.Fatalf("Initial definition failed: %v", err)
		}

		// Verify initial value
		val1, exists := interp.GetVariable("MY_CONST")
		if !exists {
			t.Fatal("MY_CONST not found after initial definition")
		}
		if n, _ := lang.ToFloat64(val1); n != 100 {
			t.Errorf("Expected MY_CONST to be 100, got %v", val1)
		}

		// 3. Redefine the constant
		cmdScript2 := `
			command
				must tool.ns.def_global_const("MY_CONST", 200)
			endcommand
		`
		tree2, _ := interp.Parser().Parse(cmdScript2)
		prog2, _, _ := interp.ASTBuilder().Build(tree2)
		interp.state.commands = prog2.Commands
		if _, err := interp.ExecuteCommands(); err != nil {
			t.Fatalf("Redefinition failed despite AllowRedefinition=true: %v", err)
		}

		// Verify updated value
		val2, exists := interp.GetVariable("MY_CONST")
		if !exists {
			t.Fatal("MY_CONST not found after redefinition")
		}
		if n, _ := lang.ToFloat64(val2); n != 200 {
			t.Errorf("Expected MY_CONST to be 200 after redefinition, got %v", val2)
		}
	})

	t.Run("Fails to redefine global constant by default (AllowRedefinition=false)", func(t *testing.T) {
		// 1. Create a config interpreter (defaults to AllowRedefinition=false)
		hc := newTestHostContext(t, nil)
		cfgPolicy := policy.NewBuilder(policy.ContextConfig).Allow("*").Build()
		interp := NewInterpreter(
			WithHostContext(hc),
			WithExecPolicy(cfgPolicy),
		)

		// 2. Define the constant initially
		cmdScript1 := `
			command
				must tool.ns.def_global_const("MY_CONST", 100)
			endcommand
		`
		tree1, _ := interp.Parser().Parse(cmdScript1)
		prog1, _, _ := interp.ASTBuilder().Build(tree1)
		interp.state.commands = prog1.Commands
		if _, err := interp.ExecuteCommands(); err != nil {
			t.Fatalf("Initial definition failed: %v", err)
		}

		// 3. Attempt to redefine
		cmdScript2 := `
			command
				must tool.ns.def_global_const("MY_CONST", 200)
			endcommand
		`
		tree2, _ := interp.Parser().Parse(cmdScript2)
		prog2, _, _ := interp.ASTBuilder().Build(tree2)
		interp.state.commands = prog2.Commands
		_, err := interp.ExecuteCommands()

		if err == nil {
			t.Fatal("Expected redefinition to fail by default, but it succeeded")
		}
		// We expect the specific error message from interpreter_tools.go
		if !strings.Contains(err.Error(), "already defined as a global constant") {
			t.Errorf("Expected 'already defined' error, got: %v", err)
		}
	})
}
