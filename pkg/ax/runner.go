// NeuroScript Version: 0.8.0
// File version: 6
// Purpose: Adds LoadScript to the RunnerCore to allow executing command blocks.
// filename: pkg/ax/runner.go
// nlines: 34
// risk_rating: LOW

package ax

// RunnerCore defines the basic execution methods.
type RunnerCore interface {
	// LoadScript parses and loads a script's definitions (funcs, events) and
	// top-level command blocks into the runner's interpreter. It does not
	// execute anything. This is idempotent for funcs but will append commands.
	LoadScript(script []byte) error

	// Execute runs the top-level command blocks that have been loaded.
	Execute() (any, error)

	// Run executes a specific, named procedure with the given arguments.
	Run(proc string, args ...any) (any, error)

	EmitEvent(name, source string, payload any)
}

// FnDefsCap allows copying function definitions between runners.
type FnDefsCap interface {
	CopyFunctionsFrom(src RunnerCore) error
}

// ToolCap provides access to the tool registry.
type ToolCap interface {
	Tools() Tools
}

// Runner is the small, composable bundle of capabilities for consumers.
type Runner interface {
	RunnerCore
	IdentityCap
	FnDefsCap
	ToolCap
}

// CloneCap is an optional capability for runners that can be cloned.
type CloneCap interface {
	Clone() Runner
}
