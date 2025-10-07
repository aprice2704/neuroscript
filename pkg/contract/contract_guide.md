Here are two developer-complete guides you can hand to the teams. They give a clear backbone (identity + extension seam), spell out the minimal interfaces, and lay out migration steps that won’t detonate your tests. They’re conceptual on purpose, but concrete enough to implement directly.

---

# Guide A — Neuroscript “Provider” Side

*(define small contracts, expose them via `api`, and offer a clean extension seam to add compiler/interpreter features)*

## 0) Objectives

* Publish a **tiny, stable contract layer** (identity, etc.) with no heavy deps or `any`.
* Re-export those contracts via `api` so downstreams (FDM/Zadeh/Lotfi/tools) import from one stable surface.
* Provide one **explicit** seam to add language features (builtins/types/ops) without `init()` magic or blank imports.

## 1) Packages & Namespacing (the “where”)

* `neuroscript/pkg/contracts` → **pure** interfaces, no heavy imports. Think of this as the “wire protocol” for capabilities.
* `neuroscript/pkg/api/contracts` → **re-exports** of the above. Consumers import here.
* `neuroscript/pkg/api/compilext` → extension seam: `Registrar` + `Extension`.
* `neuroscript/pkg/api` → public interpreter that **implements** `compilext.Registrar` and provides `Use(...)`.

**Why:** Package paths are your namespace. Callers learn: “contracts live at `api/contracts`; extensions enter via `api/compilext`.”

## 2) Contract Layer (identity first, others later)

Keep these interfaces tiny and orthogonal.

```go
// neuroscript/pkg/contracts/identity.go
// NS-CONTRACT v0.1
package contracts

type DID string

type Identity interface {
    DID() DID
}

type Signer interface {
    DID() DID
    PublicKey() []byte
    Sign(data []byte) ([]byte, error)
}

// Optional capability for runtimes/interpreters
type RuntimeIdentity interface {
    Identity() Identity
    // Optional: Signer() Signer
}
```

Re-export them so consumers don’t touch internals:

```go
// neuroscript/pkg/api/contracts/contracts.go
package contracts

import base "github.com/aprice2704/neuroscript/pkg/contracts"

type (
    DID             = base.DID
    Identity        = base.Identity
    Signer          = base.Signer
    RuntimeIdentity = base.RuntimeIdentity
)
```

**Design rules**

* No crypto imports here; use `[]byte` for keys/sigs.
* One job per interface: Identity = “who”, Signer = “can sign”, RuntimeIdentity = “where to fetch current identity”.
* Version in comments (`NS-CONTRACT vX.Y`). Only revise with additive changes where possible.

## 3) Extension Seam (clean add-to-compiler hook)

Public seam lives under `api/compilext`.

```go
// neuroscript/pkg/api/compilext/compilext.go
package compilext

type Registrar interface {
    RegisterBuiltin(name string, fn any) error
    RegisterType(name string, factory any) error
    // Add new categories sparingly: RegisterOp, RegisterMacro, etc.
}

type Extension interface {
    Name() string
    Register(Registrar) error
}
```

Interpreter implements `Registrar` and exposes `Use(...)`:

```go
// neuroscript/pkg/api/interpreter_ext.go
package api

import (
    "fmt"
    "github.com/aprice2704/neuroscript/pkg/api/compilext"
)

func (i *Interpreter) RegisterBuiltin(name string, fn any) error {
    // reflect signature, adapt to internal calling convention, validate, then register
    // reject duplicates; return error
    return i.internal.RegisterBuiltin(name, fn)
}

func (i *Interpreter) RegisterType(name string, factory any) error {
    return i.internal.RegisterType(name, factory)
}

func (i *Interpreter) Use(exts ...compilext.Extension) error {
    for _, e := range exts {
        if err := e.Register(i); err != nil {
            return fmt.Errorf("compilext %q: %w", e.Name(), err)
        }
    }
    return nil
}
```

**Rules**

* **No `init()` side effects** and no blank imports. All additions happen via `Use(...)`.
* Validate inputs at the boundary; fail fast with clear errors.
* Deterministic order = call order of `Use(...)`.

## 4) Internal wiring guidance (inside NS)

* Keep the internal registries private; only expose them through `Registrar`.
* Make registration **idempotent or strictly duplicate-rejected**; don’t silently overwrite.
* Thread safety: guard registries with a mutex if registration can happen post-construction.

## 5) Migration (low blast radius)

1. Add `pkg/contracts` and `api/contracts` (no callers change yet).
2. Add `api/compilext` and `Use(...)` + `Register*` methods to the public interpreter.
3. Move any existing ad-hoc registrations into small `Extension`s and call `Use(...)` in app wiring (dev binaries / FDM / Zadeh). Keep legacy registration behind a build tag for a week if you need to straddle.

## 6) Testing checklist

* Unit: registering duplicate names yields error; invalid function signatures yield error.
* Unit: `Use(A,B)` calls A then B (spy extension to assert order).
* Unit: zero extensions leaves base language unchanged.
* Fakes: `fakeID` and `fakeSigner` one-liners for tests.
* Compile-time: `var _ compilext.Registrar = (*Interpreter)(nil)`.

## 7) Pitfalls to avoid

* Don’t make `Identity` carry keys or profiles—keep that in `Signer` or higher layers.
* Don’t leak internal interpreter types through `Registrar` signatures.
* Don’t rely on global singletons or blank-import “registration.”

---

# Guide B — FDM “Consumer” Side

*(consume NS contracts from `api`, replace `any` seams, and wire extensions explicitly)*

## 0) Objectives

* Stop passing identity as `any`; use the **typed** contract from NS API.
* Provide one boring method to fetch identity at runtime.
* If you need to talk to NS’s concrete interpreter struct, do it behind a **tiny local interface** and a one-file wrapper. No system-wide refactor.

## 1) Imports & Namespacing

* **Only** import NS contracts from `neuroscript/pkg/api/contracts`.
* If you add language features, import the seam from `neuroscript/pkg/api/compilext`.

```go
import (
    nsct "github.com/aprice2704/neuroscript/pkg/api/contracts"
    xt   "github.com/aprice2704/neuroscript/pkg/api/compilext"
)
```

## 2) Replace `any` identity seams

Change interfaces that currently use `any` for identity.

```go
// fdm/code/interfaces/identity.go
package interfaces

import nsct "github.com/aprice2704/neuroscript/pkg/api/contracts"

type Identity = nsct.Identity
type Signer   = nsct.Signer
type DID      = nsct.DID
```

```go
// fdm/code/interfaces/identity_service.go
package interfaces

type IdentityService interface {
    GetIdentity() Identity
    SetIdentity(Identity)
    SaiDID() string
    SetSaiDID(string)
    KMS() KMS
    Signer() Signer
    // ...existing admin/capsule methods unchanged
}
```

**Why:** compile-time safety; no runtime `type` assertions; no crypto pulled into the wrong places.

## 3) Make runtime expose identity (single source of truth)

Add one additive method. Leave existing methods in place for now.

```go
// fdm/code/interfaces/runtime.go
package interfaces

import nsct "github.com/aprice2704/neuroscript/pkg/api/contracts"

type Runtime interface {
    // existing: Actor()..., AppendScript(...), Execute()..., SetEmitFunc(...)
    Identity() nsct.Identity
    // optionally: Signer() nsct.Signer
}
```

**Guard:** in hot paths that execute code, assert identity is present early and loudly.

```go
func requireIdentity(rt Runtime) nsct.Identity {
    id := rt.Identity()
    if id == nil || string(id.DID()) == "" {
        panic("runtime has no identity") // or return error; your call
    }
    return id
}
```

## 4) Talking to NS’s interpreter (without coupling)

Define **your** tiny local interface for what FDM actually calls, and wrap NS’s concrete struct once.

```go
// fdm/code/interfaces/ns_interpreter_core.go
package interfaces

import api "github.com/aprice2704/neuroscript/pkg/api"

type NSInterpreterCore interface {
    Execute() (api.Value, error)
    Stop()
    // Add LoadUnit(...), AppendScript(...), etc., only if FDM truly needs them
}
```

```go
// fdm/code/adapters/ns_interpreter_adapter.go
package adapters

import (
    api  "github.com/aprice2704/neuroscript/pkg/api"
    nsit "github.com/aprice2704/neuroscript/pkg/interpreter"
)

type NSInterpreterAdapter struct{ Inner *nsit.Interpreter }

func (a *NSInterpreterAdapter) Execute() (api.Value, error) { return a.Inner.Execute() }
func (a *NSInterpreterAdapter) Stop()                        { a.Inner.Stop() }

// Compile-time assertion:
var _ interfaces.NSInterpreterCore = (*NSInterpreterAdapter)(nil)
```

**Why:** you code to a **small** interface, not a concrete external struct. The adapter is one file and isolates reality’s oddities.

## 5) Adding language features explicitly (no spooky init)

Where you wire the interpreter, call `Use(...)` with extensions.

```go
interp, err := api.NewInterpreter(/* existing opts */)
if err != nil { /* handle */ }

// Example extension
type DatesExt struct{}
func (DatesExt) Name() string { return "dates" }
func (DatesExt) Register(r xt.Registrar) error {
    return r.RegisterBuiltin("now", func() int64 { return time.Now().Unix() })
}

// Wire them explicitly
if err := interp.Use(DatesExt{} /*, MoreExt{} */); err != nil { /* handle */ }
```

**Policy:** extensions are pinned and ordered in code. Tests don’t change unless they want to opt in.

## 6) Migration plan (keep tests stable)

1. Change only `IdentityService` (`any` → `Identity`) and add `Runtime.Identity()`. Fix the handful of sites that cast `any`.
2. Remove (or rename) any **empty** `Interpreter` interfaces that shadow NS names.
3. Introduce `NSInterpreterCore` (tiny) and wrap NS’s concrete struct once.
4. Migrate any existing side-effect registrations to explicit `Use(...)` in wiring code. Keep legacy paths behind a short-lived build tag if you need overlap.

## 7) Testing checklist

* Compile-time:
  `var _ interfaces.Runtime = (*MyRuntime)(nil)`
  `var _ interfaces.IdentityService = (*MyIDSvc)(nil)`
  `var _ interfaces.NSInterpreterCore = (*NSInterpreterAdapter)(nil)`
* Unit: a fake identity implementing `Identity` with a fixed DID; ensure plumbing uses it.
* Unit: hot paths error when identity is missing (don’t fail mid-flight).
* Integration: `Use(...)` registers extensions in order; duplicates are rejected.

## 8) Operational guardrails

* Log extension registration at startup with names and counts.
* Expose a `GET /healthz/runtime-identity` debug handler that returns the current DID (no keys), to catch missing identity in environments early.
* Keep any admin registry access **out** of the hot execution path; prefer `Runtime.Identity()` in tools.

## 9) Things **not** to do

* No `any` for identity. Ever.
* No blank imports to auto-register compiler features.
* No “fat” interfaces that force Signer into components that only need DID.

---

## Final notes for both teams

* **Small, composable interfaces** are the lever. Start with Identity, then add micro-contracts (Clock, Entropy, SandboxFS) as *separate* files when truly needed.
* **Explicit wiring beats magic.** `Use(...)` is the only door for compiler additions.
* **Namespacing via package paths** is your friend: `api/contracts` (types) and `api/compilext` (extensions) are the stable surfaces the rest of the world should depend on.

If you want, I can turn this into three PRs: (1) NS contracts + re-exports, (2) NS `compilext` + Interpreter `Use(...)`, (3) FDM identity/Runtime + adapter + explicit `Use(...)` wiring.
