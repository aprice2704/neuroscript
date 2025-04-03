// pkg/core/test_helpers_test.go
package core

// "log" // Import if logger needs configuration, currently using nil
// "io"  // Import if logger needs configuration
// Import testing only if needed within helpers (e.g., t.Helper())

// --- Shared Test Helper Functions ---

// newTestInterpreterEval creates an interpreter instance for evaluation tests.
// It initializes variables and the last call result.
func newTestInterpreterEval(vars map[string]interface{}, lastResult interface{}) *Interpreter {
	// Note: Using NewInterpreter(nil) to disable logging during tests by default.
	// Change to log.New(os.Stderr, ...) if debug logging is needed for tests.
	interp := NewInterpreter(nil)
	if vars != nil {
		interp.variables = make(map[string]interface{}, len(vars))
		for k, v := range vars {
			interp.variables[k] = v
		}
	} else {
		interp.variables = make(map[string]interface{})
	}
	interp.lastCallResult = lastResult // Use the specific field name
	return interp
}

// newDummyInterpreter creates a basic interpreter, often used for tool tests
// where initial variable state or last result isn't the focus.
func newDummyInterpreter() *Interpreter {
	// Using NewInterpreter(nil) to disable logging during tests by default.
	interp := NewInterpreter(nil)
	// No initial vars or last result set
	return interp
}

// makeArgs simplifies creating []interface{} slices for tool arguments in tests.
func makeArgs(vals ...interface{}) []interface{} {
	// Using a direct variadic function is cleaner than returning vals directly sometimes
	// though just `return vals` works too. This ensures it's always []interface{}.
	args := make([]interface{}, len(vals))
	for i, v := range vals {
		args[i] = v
	}
	return args
}

// createTestStep helper remains local to interpreter_test.go as it uses Step struct internals extensively.
// func createTestStep(...)

// createIfStep helper remains local to interpreter_test.go.
// func createIfStep(...)

// runExecuteStepsTest helper remains local to interpreter_test.go.
// func runExecuteStepsTest(...)
