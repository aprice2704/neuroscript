Great—since imports are clean and `ax` is stdlib-only, here’s a tight, updated **wiring guide** for the dev team. It assumes `ax` lives at `neuroscript/pkg/ax`. The only open item you’ll fill is `AdminCapsuleRegistry` on the API side.

---

# Wiring Guide (NS side)

## 0) Invariants (don’t break these)

* `pkg/ax` imports **stdlib only**. No NS packages.
* All engine specifics (tools, stores, capsules, values) are adapted **in `pkg/api`**.
* Exactly one concrete type satisfies `ax.Runner` in a build (`*axRunner` below).

---

## 1) Implement `ax.Registry` on the public interpreter

**file:** `neuroscript/pkg/api/ax_bridge.go`

```go
package api

import (
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/ax"
)

var _ ax.Registry = (*Interpreter)(nil)

func (i *Interpreter) RegisterBuiltin(name string, fn any) error { return i.internal.RegisterBuiltin(name, fn) }
func (i *Interpreter) RegisterType(name string, factory any) error { return i.internal.RegisterType(name, factory) }

func (i *Interpreter) Use(exts ...ax.Extension) error {
	for _, e := range exts {
		if err := e.Register(i); err != nil {
			return fmt.Errorf("ax extension %q: %w", e.Name(), err)
		}
	}
	return nil
}
```

---

## 2) Adapt the root env to `ax.RunEnv`

**file:** `neuroscript/pkg/api/ax_env_impl.go`

```go
package api

import "github.com/aprice2704/neuroscript/pkg/ax"

// -- Accounts admin --
type axAccountsAdmin struct{ ua interfaces.AccountAdmin }
func (a axAccountsAdmin) Register(name string, cfg map[string]any) error { return a.ua.Register(name, cfg) }

// -- Agent models admin --
type axAgentModelsAdmin struct{ ua interfaces.AgentModelAdmin }
func (a axAgentModelsAdmin) Register(name string, cfg map[string]any) error { return a.ua.Register(name, cfg) }

// -- Capsules admin --
// DEV: Fill this to delegate to your concrete AdminCapsuleRegistry.
// e.g., Install(name, bytes, meta) -> registry.Install(...)
// You control the translation of meta/bytes to your real API.
type axCapsulesAdmin struct{ reg *AdminCapsuleRegistry }
func (a axCapsulesAdmin) Install(name string, content []byte, meta map[string]any) error {
	return a.reg.Install(name, content, meta) // adjust as needed
}

// -- Tools registry --
type axTools struct{ tr tool.ToolRegistry }
func (t axTools) Register(name string, impl any) error { return t.tr.Register(name, impl) }
func (t axTools) Lookup(name string) (any, bool)       { return t.tr.Lookup(name) }

// -- RunEnv bound to a root interpreter --
type axRunEnv struct{ root *Interpreter }

func (e *axRunEnv) AccountsAdmin() ax.AccountsAdmin       { return axAccountsAdmin{ua: e.root.AccountsAdmin()} }
func (e *axRunEnv) AgentModelsAdmin() ax.AgentModelsAdmin { return axAgentModelsAdmin{ua: e.root.AgentModelsAdmin()} }
func (e *axRunEnv) CapsulesAdmin() ax.CapsulesAdmin       { return axCapsulesAdmin{reg: e.root.CapsuleRegistryForAdmin()} }
func (e *axRunEnv) Tools() ax.Tools                       { return axTools{tr: e.root.ToolRegistry()} }
```

> You’ll implement `AdminCapsuleRegistry.Install` appropriately; the adapter above just forwards.

---

## 3) Wrap host identity and expose an `ax.Runner`

**file:** `neuroscript/pkg/api/ax_runner_impl.go`

```go
package api

import (
	"errors"

	"github.com/aprice2704/neuroscript/pkg/ax"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

type hostIdentity struct{ did ax.DID }
func (h hostIdentity) DID() ax.DID { return h.did }

// hostRuntime goes into SetRuntime; it augments your existing runtime with IdentityCap.
type hostRuntime struct {
	Runtime // your existing tool runtime
	id ax.ID
}
func (h *hostRuntime) Identity() ax.ID { return h.id }

// axRunner presents the ax bundle, delegating to the public interpreter.
type axRunner struct {
	env  *axRunEnv
	host *hostRuntime
	itp  *Interpreter
}

var _ ax.Runner = (*axRunner)(nil)

func (r *axRunner) Env() ax.RunEnv { return r.env }

// RunnerCore — adapt any<->engine values here
func (r *axRunner) Execute() (any, error) { return r.itp.Execute() }
func (r *axRunner) Run(proc string, args ...any) (any, error) {
	vs := make([]lang.Value, len(args))
	for i, a := range args {
		v, err := lang.Wrap(a)
		if err != nil { return nil, err }
		vs[i] = v
	}
	return r.itp.Run(proc, vs...)
}
func (r *axRunner) EmitEvent(name, src string, payload any) {
	pv, _ := lang.Wrap(payload)
	r.itp.EmitEvent(name, src, pv)
}

// IdentityCap
func (r *axRunner) Identity() ax.ID { return r.host.id }

// ToolCap
func (r *axRunner) Tools() ax.Tools { return axTools{tr: r.env.root.ToolRegistry()} }

// FnDefsCap
func (r *axRunner) CopyFunctionsFrom(src ax.RunnerCore) error {
	other, ok := src.(*axRunner)
	if !ok || other == nil { return errors.New("CopyFunctionsFrom: incompatible src") }
	return r.itp.CopyFunctionsFrom(other.itp)
}
```

---

## 4) Factory: shared env, config/user runners, fn-copy

**file:** `neuroscript/pkg/api/ax_factory_impl.go`

```go
package api

import (
	"context"
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/ax"
)

type axFactory struct {
	env  *axRunEnv
	root *Interpreter
}

var _ ax.RunnerFactory = (*axFactory)(nil)
var _ ax.EnvCap = (*axFactory)(nil)

func NewAXFactory(ctx context.Context, rootOpts ax.RunnerOpts, baseRt Runtime, id ax.ID) (*axFactory, error) {
	root := New()
	host := &hostRuntime{Runtime: baseRt, id: id}
	if err := root.SetRuntime(host); err != nil { return nil, fmt.Errorf("root SetRuntime: %w", err) }
	if rootOpts.SandboxDir != "" { root.SetSandboxDir(rootOpts.SandboxDir) }
	return &axFactory{env: &axRunEnv{root: root}, root: root}, nil
}

func (f *axFactory) Env() ax.RunEnv { return f.env }

func (f *axFactory) NewRunner(ctx context.Context, mode ax.RunnerMode, opts ax.RunnerOpts) (ax.Runner, error) {
	itp := New()
	// Reuse identity from root; if preferred, stash id in factory and pass that.
	id := f.env.rootIdentity()
	host := &hostRuntime{Runtime: f.root, id: id}
	if err := itp.SetRuntime(host); err != nil { return nil, fmt.Errorf("runner SetRuntime: %w", err) }
	if opts.SandboxDir != "" { itp.SetSandboxDir(opts.SandboxDir) }

	r := &axRunner{env: f.env, host: host, itp: itp}
	if mode == ax.RunnerUser {
		if err := r.CopyFunctionsFrom(f.root); err != nil {
			return nil, fmt.Errorf("copy fn defs: %w", err)
		}
	}
	return r, nil
}
```

> Implement `rootIdentity()` however you prefer:
>
> * Easiest: add `func (i *Interpreter) Identity() ax.ID` that inspects the `hostRuntime` you passed to `SetRuntime`.
> * Or store `id` on the factory when you build it.

Example:

```go
// in api/interpreter.go (or nearby)
func (i *Interpreter) Identity() ax.ID {
	if h, ok := i.hostRuntime.(*hostRuntime); ok { return h.id }
	return nil
}
func (e *axRunEnv) rootIdentity() ax.ID { return e.root.Identity() }
```

---

## 5) Compile-time tripwires

Drop these near the files above:

```go
var _ ax.Registry      = (*Interpreter)(nil)
var _ ax.RunnerFactory = (*axFactory)(nil)
var _ ax.RunEnv        = (*axRunEnv)(nil)
var _ ax.Runner        = (*axRunner)(nil)
var _ ax.IdentityCap   = (*hostRuntime)(nil)
```

---

# Quick FDM usage (sanity test)

```go
fac, _ := nsapi.NewAXFactory(ctx, ax.RunnerOpts{SandboxDir:"/tmp"}, baseRt, myID)
// config
cfg, _ := fac.NewRunner(ctx, ax.RunnerConfig, ax.RunnerOpts{})
env := fac.Env()
_ = env.AccountsAdmin().Register("svcA", map[string]any{"token":"…"})
_ = env.AgentModelsAdmin().Register("gpt-pro", map[string]any{"ctx": 128000})
_ = env.CapsulesAdmin().Install("std", zipBytes, map[string]any{"version":"1.2.3"})
_ = env.Tools().Register("do_thing", myTool)

// user
usr, _ := fac.NewRunner(ctx, ax.RunnerUser, ax.RunnerOpts{})
_, _ = usr.Execute()
```

---

# Testing checklist

* Unit: `NewAXFactory` sets runtime once; second call errors.
* Unit: config runner mutations are visible to user runners via `Env()`.
* Unit: `CopyFunctionsFrom` copies procs, not state.
* Unit: tool registration works and is discoverable via `Tools().Lookup`.
* Unit: identity present on both config/user runners: `usr.Identity().DID() != ""`.

That’s the whole path. With `AdminCapsuleRegistry` defined on your side, the adapters above will compile and give FDM the clean `ax` surface you wanted.
