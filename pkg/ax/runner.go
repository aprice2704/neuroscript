// NeuroScript Version: 0.7.4
// File version: 4
// Purpose: Reverted to a stdlib-only package by removing Clone() from the core interface and dropping internal type dependencies.
// filename: pkg/ax/runner.go
// nlines: 32
// risk_rating: LOW

package ax

// RunnerCore defines the basic execution methods.
type RunnerCore interface {
	Execute() (any, error)
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
	EnvCap
}

// CloneCap is an optional capability for runners that can be cloned.
type CloneCap interface {
	Clone() Runner
}
