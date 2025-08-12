Here’s a tight integration note you can drop into the repo (e.g., `docs/dev/integration_policy_capabilities.md`). It’s concrete, Go-specific, and split by team.

# Integration Guide — Policy, Capabilities, Effects

## What just landed

* `pkg/policy/capability/*` — capability structs, matcher, limits/counters.
* `pkg/runtime/policy.go` — the gate that enforces trust/allow/deny/caps/limits before any tool call.
* `cmd/ns-lint/*` — lints `::metadata` headers for policy/effects sanity.
* Extended metadata spec (`docs/spec/metadata-extended.md`) describing `::policy*`, `::grant.*`, `::limit.*`, `::effects`, etc.

## Interpreter team — what to wire

1. **Load metadata & build the run policy**

* Parse file header metadata (you may already have a header pass).
* Merge in CLI flags / baked-in defaults:

  * `Context` → `config|normal|test`
  * `Allow`/`Deny` → tool patterns
  * `Grants` → build `[]capability.Capability` from `::grant.*`
  * `Limits` → fill `capability.Limits` from `::limit.*`
* Construct:

  ```go
  pol := runtime.ExecPolicy{
      Context: runtime.ContextNormal, // or config/test
      Allow:   allowPatterns,         // []string
      Deny:    denyPatterns,          // []string
      Grants:  capability.NewGrantSet(grants, limits),
  }
  ```

2. **Gate every tool call**
   Right before dispatching a tool implementation:

```go
meta := runtime.ToolMeta{
    Name:          toolName,
    RequiresTrust: registryEntry.RequiresTrust,
    RequiredCaps:  registryEntry.RequiredCaps, // []capability.Capability
    Effects:       registryEntry.Effects,      // optional
}
if err := pol.CanCall(meta); err != nil {
    return nil, fmt.Errorf("policy: %w", err)
}
```

If the call proceeds and you **know bytes/spend**, account them:

```go
_ = pol.Grants.CountNet(nBytes)     // on net ops
_ = pol.Grants.CountFS(nBytes)      // on fs ops
_ = pol.Grants.CheckPerCallBudget("CAD", cents)
_ = pol.Grants.ChargeBudget("CAD", cents) // accumulate
```

If any returns an error, abort the call and surface the sentinel error (`ErrNetExceeded`, etc.).

3. **Persist AgentModels only in config**

* Your AgentModel registry lives in the interpreter state.
* Only `Context=config` with appropriate grants may call `Register/Update/Delete`.
* `ask` should **not** have its own net/secret caps; it **consumes** the envelope of the selected model (host, secret key name, budget currency). Enforce that the current `ExecPolicy.Grants` satisfy that envelope before performing the request.

4. **Effects for determinism/caching (optional but recommended)**

* If a function/script declares `::pure: true`, reject any tool with effects that imply impurity (`readsNet`, `readsFS`, `readsClock`, `readsRand`).
* In `Context=test`, require explicit `grant.clock.read` / `grant.rand.read` (or seed) to allow non-determinism.

5. **Provenance**

* Log: tool name, callsite, matched caps, limit deltas, spend deltas, model name/host, duration.
* Attach a compact provenance tag to outputs if your pipeline supports it.

6. **Tests you must add**

* Trust: `RequiresTrust` tool in `normal` → blocked with `ErrTrust`.
* Allow/Deny precedence: deny wins.
* Caps: missing → `ErrCapability`; wildcard host/path → allowed.
* Limits: per-call > run; bytes/calls overflows on net/fs; per-tool limit.
* Metadata: config script with admin grants but `::policyContext != config` → linter should flag (unit test `cmd/ns-lint`).

## Tool team — what to expose

1. **Annotate every tool in the registry**
   Add or extend your registry struct:

```go
type ToolEntry struct {
    Name          string
    RequiresTrust bool
    RequiredCaps  []capability.Capability
    Effects       []string // "idempotent","readsNet","readsFS","readsClock","readsRand"
}
```

Populate it **once** at init. Examples:

* Config-only:

  ```go
  {
    Name: "tool.agentmodel.Register",
    RequiresTrust: true,
    RequiredCaps: []capability.Capability{
        {Resource:"model", Verbs:[]string{"admin"}, Scopes:[]string{"*"}},
    },
    Effects: []string{"idempotent"}, // admin update is conceptually idempotent
  }
  {
    Name: "tool.os.Getenv",
    RequiresTrust: true,
    RequiredCaps: []capability.Capability{
        {Resource:"env", Verbs:[]string{"read"}, Scopes:[]string{"OPENAI_API_KEY"}},
    },
    Effects: []string{"idempotent"},
  }
  ```
* Read-only helpers (safe in normal):

  ```go
  {
    Name: "tool.agentmodel.List",
    RequiresTrust: false,
    RequiredCaps: nil,
    Effects: []string{"idempotent"},
  }
  ```
* `ask` tooling (invocation path):

  * The **tool** that performs network I/O should declare `Effects:["readsNet"]`.
  * It should **not** require `env`/`net` caps directly if you enforce via the AgentModel envelope + policy. If you want an extra belt: set `RequiredCaps` to `net:read:<host>` and `budget` limits — still satisfied by the `ExecPolicy`.

2. **Instrument resource usage**

* Network: estimate or count response/request bytes → `CountNet`.
* Filesystem: count bytes read/written → `CountFS`.
* Spend: if you have token → cost mapping, convert to cents (use CAD here if that’s your standard) and call `CheckPerCallBudget` then `ChargeBudget`.

3. **Return correct sentinel errors**
   Use the exported errors from `pkg/policy/capability` *verbatim*, or wrap with `%w`:

```go
if err := pol.Grants.CountNet(n); err != nil {
    return nil, fmt.Errorf("net accounting failed: %w", err)
}
```

Consumers can `errors.Is(err, capability.ErrNetExceeded)` per AI\_RULES.

## Shared conventions (both teams)

* **No name-based security**: Namespaces are for clarity; enforcement is via `RequiresTrust` + policy gate + caps.
* **Default-deny**: ship the interpreter with **no** grants and an empty allowlist. Host/CLI or config scripts must opt in.
* **Metadata is declarative intent**: script headers (“`::policy*` / `::grant.*` / `::limit.*` / `::effects`”) inform the policy builder; **they do not auto-grant** anything. Runtime policy (CLI/baked-in) remains the authority.
* **AI\_RULES headers**: keep them on any new files you add or touch.

## CLI examples for local runs

* **Trusted bootstrap** (register models; read env):

  ```
  ns run --context=config \
         --allow=tool.agentmodel.Register,tool.os.Getenv \
         --grant=env.read:OPENAI_API_KEY \
         --grant=model.admin:* \
         scripts/init_models.ns
  ```

* **Normal run** (no admin tools; budget and host limited):

  ```
  ns run --context=normal \
         --allow=tool.agentmodel.List \
         --grant=model.use:mini \
         --grant=net.read:*.openai.com:443 \
         --limit=budget.CAD.max:6000 \
         scripts/summarize.ns
  ```

## Quick checklist

* [ ] Registry entries updated with `RequiresTrust`, `RequiredCaps`, `Effects`.
* [ ] Policy built from metadata + CLI; deny > allow; grants + limits wired.
* [ ] `ExecPolicy.CanCall` invoked before every tool dispatch.
* [ ] Net/fs/spend accounted with counters; errors surfaced.
* [ ] AgentModel envelope enforced for `ask`.
* [ ] `ns-lint` added to CI (fail on ERROR).
* [ ] Unit tests cover trust/allow/deny/caps/limits; `ns-lint` fixtures cover metadata edge cases.

If you want, I can also draft a tiny `AgentModel envelope` struct and the 20-line check you drop into the `ask` path to validate host/secret/budget against `ExecPolicy`.
