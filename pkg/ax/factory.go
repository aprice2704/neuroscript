// NeuroScript Version: 0.8.0
// File version: 2
// Purpose: Removes the deprecated EnvCap from the factory interface.
// filename: pkg/ax/factory.go
// nlines: 16
// risk_rating: LOW

package ax

import "context"

// RunnerMode differentiates normal "work" runners from privileged "config" ones.
type RunnerMode int

const (
	RunnerUser RunnerMode = iota
	RunnerConfig
)

// RunnerOpts: add the small set you truly need (sandbox dir, IO, policy, etc.)
type RunnerOpts struct {
	SandboxDir string
	// Stdout/Stderr, ExecPolicy, Emit handler, Providers... as needed
}

// RunnerFactory creates runners bound to a RunEnv.
// A single factory instance can mint both user and config runners.
type RunnerFactory interface {
	NewRunner(ctx context.Context, mode RunnerMode, opts RunnerOpts) (Runner, error)
}
