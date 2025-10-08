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
	EnvCap // expose the shared environment
	NewRunner(ctx context.Context, mode RunnerMode, opts RunnerOpts) (Runner, error)
}
