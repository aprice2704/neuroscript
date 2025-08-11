AJP: Good. We are about to add some tools that do config things like set the agentmodels available to the ask statement (some discussio attached), read env vars (to set api keys), set sandbox, etc. so all things that should only be run in a trusted config script (have added cli args for this to ng, will add to fdm shortly).
So do we:
put all these tools in their own group "tool.needstrust.getenv" or "tool.sys.AddAgentModel" 
add a flag to the tools in their typedefs
or something else
in order to allow list them for execution in a special trusted config script?
we could "bake in" an allow list into the exe (?)
what do you think?

G5:

Short version: don’t rely on namespacing alone. Give “dangerous” tools an explicit **trust requirement** in their type definition *and* enforce an **execution policy** (allowlist/denylist) at interpreter-run level. Defense in depth.

Here’s the concrete setup I recommend, wired to how we’ve been shaping `ask` and AgentModels. &#x20;

# Recommendation

1. **Add trust metadata to tools (source of truth)**

* In each tool typedef, include minimal, machine-checkable flags:

  * `requires_trust: true|false`
  * `capabilities: ["config", "fs", "net", ...]`  *(tiny enum; “config” is the important one here)*
* Examples:

  * `tool.agentmodel.Register/Update/Delete` → `requires_trust: true`, `capabilities: ["config"]`
  * `tool.agentmodel.List/Select` → `requires_trust: false`
  * `tool.os.Getenv` (reads secrets) → `requires_trust: true`, `capabilities: ["config"]`

2. **Enforce a per-run execution policy (host-controlled)**

* When you spin up an interpreter, provide a **Policy**:

  * `Allowlist`: exact tool names (optionally with pattern support), e.g. `["tool.agentmodel.Register", "tool.os.Getenv"]`
  * `Denylist`: exact names/patterns (wins over allowlist)
  * `Context`: `"config"` or `"normal"`
* The checker is:
  `call_allowed = (!tool.requires_trust || policy.context=="config") && in_allowlist && !in_denylist`
* Run your **startup script** with `context="config"` and a small allowlist; run all normal scripts with `context="normal"` and a strict default allowlist (or default-deny).

3. **Keep namespaces for readability, not security**

* Use `tool.agentmodel.*` for human clarity, but don’t assume `tool.agentmodel.*` == safe. The trust check must be flag+policy, not string-prefix.

4. **Bake-in only the *minimum* allowlist needed to bootstrap**

* Ship the binary with a tiny compiled allowlist that only permits:

  * `tool.agentmodel.Register/Update/Delete`
  * `tool.os.Getenv` (or your secure secret fetch)
* But still allow the host to **override/extend** via CLI/env:

  * `--config-allow=tool.agentmodel.Register,tool.os.Getenv`
  * Default build should **not** allow any config tools unless the host opts in.

5. **Audit trail, always**

* Log every invocation of a `requires_trust` tool with: script id, callsite, args hash (not raw), outcome.
* Emit an event (e.g., `sys.config.tool_called`) that FDM can capture for provenance.

6. **Make `ask` explicitly pick an AgentModel (already agreed)**

* Grammar: `ask <agentmodel_expr>, <prompt_expr> [with <options>] [into <lvalue>]`
* The AgentModel referenced must exist in the interpreter’s **internal AgentModel registry**, populated *only* by those config tools during the startup script. (No implicit defaults.)&#x20;

# Minimal shapes in code (Go-ish)

**Tool metadata:**

```go
type ToolMeta struct {
    Name          string
    RequiresTrust bool
    Capabilities  []string // e.g., {"config"}
}
```

**Policy attached to interpreter instance:**

```go
type ExecPolicy struct {
    Context   string   // "config" | "normal"
    Allowlist []string // exact names or globs
    Denylist  []string
}
```

**Gate on every tool call:**

```go
func (rt *Runtime) canCall(t ToolMeta) bool {
    if t.RequiresTrust && rt.policy.Context != "config" { return false }
    if !inAllow(rt.policy.Allowlist, t.Name) { return false }
    if inDeny(rt.policy.Denylist, t.Name) { return false }
    return true
}
```

# Practical wiring

* **Startup path (trusted)**
  Host starts interpreter with `Context=config`, `Allowlist=[tool.agentmodel.Register, tool.agentmodel.Update, tool.agentmodel.Delete, tool.os.Getenv]`.
  Run `init.ns`:

  ```neuroscript
  must tool.agentmodel.Register("mini", {
    "provider":"openai", "model":"gpt-4o-mini",
    "api_key": tool.os.Getenv("OPENAI_API_KEY")
  })
  ```

  AgentModels persist in the interpreter’s private state.

* **Normal execution (untrusted)**
  Reuse the *same* interpreter instance (state carries forward), but now `Context=normal`, `Allowlist` excludes config tools.
  Scripts can do:

  ```neuroscript
  ask "mini", "Summarize {{txt}}" with {"json": true} into out
  ```

  If someone tries `tool.agentmodel.Register(...)` here → blocked by policy & metadata.

* **Read-only helpers always available**
  `tool.agentmodel.List/Select` are non-trust tools; keep them in the default allowlist. Selection can enforce your criteria without granting mutation.

# Why this combo?

* **Flag-only** is too implicit; **names-only** is brittle.
  **Flag + policy + (optional) baked-in bootstrap allowlist** gives you layered control with minimal complexity.
* It matches the rest of our design: explicit, scriptable, and host-governed; no “magic defaults.”&#x20;

If you like this, I’ll sketch the tiny `ExecPolicy` gate, add `RequiresTrust` to the tool registry, and mark:

* `tool.agentmodel.Register/Update/Delete`: `requires_trust=true`
* `tool.os.Getenv`: `requires_trust=true`
* `tool.agentmodel.List/Select`: `requires_trust=false`

We can have this policy enforced before any tool invocation hits its implementation.

AJP:  Looks good. I'll work with the tool and interpreter teams to implement this. Any other ideas? You mentioned Capabilities more-or-less in  passing -- care to expand on it?

Great—glad we’re aligned. Yes, I only waved at “Capabilities”; here’s a concrete, minimal-but-powerful way to use them without overcomplicating NS.

# The point of capabilities

Flags like `requires_trust` tell you **whether** a tool is sensitive. Capabilities tell you **what kind** of authority it needs and **how much**. That buys us:

* Least privilege: tools get only the slice of power they actually need.
* Auditable, composable policy: easy to log, diff, and reason about.
* Future-proofing: we can add new capability kinds without changing NS syntax.

# Capability model (simple, explicit)

Think **resource : verb : scope (+ limits)**.

* **Resource**: `net`, `fs`, `secrets`, `env`, `time`, `rand`, `model`, `sandbox`, `proc`, `clock`, `budget`.
* **Verb**: `read`, `write`, `admin`, `exec`, etc. (small enum)
* **Scope**: pattern or selector for the resource.
* **Limits**: quantitative guards (bytes, rate, timeouts, counts, spend).

Examples (expressed as Go structs; adapt to your registry format):

```go
type Capability struct {
    Resource string   // "net"|"fs"|"secrets"|"env"|"model"|...
    Verbs    []string // e.g. {"read"}, {"write"}, {"admin"}
    Scope    any      // typed per resource (see below)
    Limits   Limits
}
type Limits struct {
    MaxBytes     int64
    MaxCalls     int
    TimeoutMS    int
    MaxSpendCents int // or tokens; see "budget" below
}
```

### Canonical scopes per resource

Keep these tight and boring:

* `net`: `{ Hosts: ["*.openai.com"], Ports: [443], TLSOnly: true }`
* `fs`: `{ Paths: ["/etc/ns/*.json", "/var/app/config/"], Mode: "ro" }`
* `secrets`: `{ Keys: ["OPENAI_API_KEY","ANTHROPIC_API_KEY"] }`
* `env`: `{ Keys: ["OPENAI_API_KEY"] }` (yes, split from `secrets` if you also support vaults)
* `model`: `{ Names: ["mini","sonnet"], Actions: ["register","select"] }`
* `sandbox`: `{ Profiles: ["py-sandbox-1"], Actions: ["define","select"] }`
* `time`: `{ FixedNow: "2025-08-11T00:00:00Z"? }` (for determinism tests)
* `rand`: `{ DeterministicSeed: 1234? }`
* `budget`: `{ Currency: "USD", Max: 1500, PerCallMax: 50 }`

# Where capabilities live

1. **Declared on tools** (static)

* Each tool carries the *capabilities it requires* (minimum set).
* Example:

  * `tool.os.Getenv`: `requires_trust=true`, caps: `[env:read:{Keys:["OPENAI_API_KEY"]}]`
  * `tool.agentmodel.Register`: `requires_trust=true`, caps: `[model:admin:{Names:["*"]}]`
  * `ask`: usually `requires_trust=false` but **consumes budget** and **net** indirectly *via its selected AgentModel*. That’s handled by the model’s capability envelope (below), not by `ask` itself.

2. **Attached to the interpreter policy** (grants)

* The host provides an **ExecPolicy** with an allowlist and a **capability grant set**.
* At call time: `tool.requires_trust` + allowlist + **capabilities grant must satisfy tool’s needs** + denylist + limits.

```go
type ExecPolicy struct {
    Context    string     // "config" | "normal" | "test"
    Allowlist  []string   // tools allowed
    Denylist   []string
    Grants     []Capability
}
```

3. **AgentModel as a capability envelope**

* When you `Register` an AgentModel, you store:

  * Endpoint + model id
  * A **capability envelope** required to use it (net host, secrets key, budget limits).
* When `ask "mini", ...` runs, the runtime checks that the current policy **grants** the envelope of `"mini"`.

This keeps `ask` pure(ish) and pushes risk to the config stage where it belongs.

# Enforcement logic (tight and testable)

* **Tool gate** (as you already proposed):

  * trust context + allowlist/denylist
  * then: `capSatisfies(tool.RequiredCaps, policy.Grants)`
* **Capability matching** is structural:

  * Resource matches exactly.
  * Verbs required ⊆ granted.
  * Scope required ⊆ granted (wildcards/patterns OK).
  * Limits: track counters and decrement; deny when exhausted.

Pseudocode:

```go
func (rt *Runtime) canCall(tool ToolMeta) error {
    if tool.RequiresTrust && rt.policy.Context != "config" { return ErrTrust }
    if !inAllow(rt.policy.Allowlist, tool.Name) || inDeny(rt.policy.Denylist, tool.Name) { return ErrPolicy }
    if !capsSatisfied(tool.RequiredCaps, rt.policy.Grants) { return ErrCapability }
    return nil
}
```

# Budget as a first-class capability

Avoid burying spend in tool options.

* Add a **budget** resource that `ask` consumes via the model’s envelope:

  * `budget: {Currency:"CAD", Max: 500, PerCallMax: 20}`
  * Track spend per run and per script; refuse when exceeded.
* Emit **spend events** and include them in provenance.

# Purity & effects

Augment tools with a tiny “effects” descriptor. This is not security; it’s for reasoning, caching, and tests.

```go
type Effects struct {
    Pure        bool // true => deterministic, no IO
    Idempotent  bool
    ReadsClock  bool
    ReadsRand   bool
}
```

* Linter can warn if a “pure only” script includes impure calls.
* Test harness can force `time`/`rand` through deterministic caps.

# Script-level declarations (optional but nice)

Allow a *header* in trusted config scripts that **requests** capabilities; host policy still decides. Keeps intent visible.

```
#@requires:
#  - {resource:"env", verbs:["read"], scope:{keys:["OPENAI_API_KEY"]}}
#  - {resource:"model", verbs:["admin"], scope:{names:["*"]}}
```

(Represent as an NS comment block that the loader parses into a map.)

# Object-capability taste (without going full ocap)

If you want extra rigor later, you can **mint tokens** representing specific capabilities and require them as an argument:

```
let secret = tool.capability.Mint({resource:"secrets", verbs:["read"], scope:{keys:["OPENAI_API_KEY"]}})
must tool.os.Getenv("OPENAI_API_KEY", with=secret)
```

This makes authority **explicit in the call graph**. You don’t need this day one, but the design above doesn’t preclude it.

# Defaults that won’t bite us

* Default-deny for `net`, `fs`, `secrets`, `env`, `proc`, `model:admin`, `sandbox:admin`.
* Default-allow for truly pure tools (math, string ops, local transforms).
* Config context gets only what the host grants *per run* (CLI/env/embedded policy).
* Normal context inherits **model use** but not **model admin**.

# Observability

Log per call:

* tool name, capability IDs used, decremented limits, outcome, spend deltas, script id/line.
* Tag outputs with a provenance summary (“depends on net:\*.openai.com, secrets\:OPENAI\_API\_KEY, budget: \$0.23”).

# How this maps to your near-term tools

* `tool.agentmodel.Register/Update/Delete`
  `requires_trust=true`, caps: `[ {resource:"model", verbs:["admin"], scope:{names:["*"]}} ]`
* `tool.os.Getenv`
  `requires_trust=true`, caps: `[ {resource:"env", verbs:["read"], scope:{keys:["OPENAI_API_KEY"]}} ]`
* `tool.sandbox.SetProfile`
  `requires_trust=true`, caps: `[ {resource:"sandbox", verbs:["admin"], scope:{profiles:["py-sandbox-1"]}} ]`
* `ask`
  `requires_trust=false`, no direct caps; **consumes** the envelope of the selected model,
  which itself requires `net`, `secrets`, `budget` caps to be present in the **policy**.

If you’re good with this, I can draft:

* A tiny `capmatch` package (structs + matcher + counters).
* Registry annotations for the current tool set.
* A policy JSON schema your CLI can ingest (`--policy @policy.json`).
* A linter rule: “config scripts must declare requested capabilities.”

That keeps us principled, debuggable, and hard to foot-gun, without turning NS into a security dissertation.




AJP: We can use metadata for script (or fn) require things -- see attached
We should just use ns, not json, for policy
I esp like the effects thing
Please continue opining; you do specs, teams do impl :)




Love it—using NS (and your in-situ metadata lines) for policy keeps everything uniform and trivially parseable. Building on your metadata spec, here’s a concrete, minimal extension that gets you **capabilities**, **effects**, and **policy**—all enforceable by the interpreter without introducing JSON or sidecar formats.&#x20;

# What to add (spec-level)

## 1) New metadata vocabulary (file + block scopes)

Add these keys to the *standard vocabulary* in the metadata spec (bump spec version). Names follow your `::key: value` rules and avoid new grammar.

### Policy / Allowlisting

* `policyContext` = `config | normal | test`
* `policyAllow` = comma-sep tool names/patterns (e.g., `tool.agentmodel.Register, tool.os.Getenv`)
* `policyDeny` = comma-sep names/patterns (deny wins)

### Capability Grants (object-capability, but flat)

Prefix with `grant.<resource>.<verb>`; value is **scope list**.

* `grant.env.read` = env keys
  e.g., `OPENAI_API_KEY, ANTHROPIC_API_KEY`
* `grant.secrets.read` = secret names/paths
* `grant.net.read` = hosts (wildcards OK), `host[:port]`
  e.g., `*.openai.com:443, api.anthropic.com:443`
* `grant.fs.read` = paths/globs
* `grant.fs.write` = paths/globs
* `grant.model.admin` = model names (`*` allowed)
* `grant.model.use` = model names permitted to *ask*
* `grant.sandbox.admin` = profile IDs
* `grant.proc.exec` = executable IDs (if you have any)
* `grant.clock.read` = `true` (opt-in for non-deterministic time)
* `grant.rand.read` = `true` or `seed:<int>` (for determinism)

### Capability Limits

* `limit.budget.<CURRENCY>.max` = integer amount per run
  e.g., `limit.budget.CAD.max: 500`
* `limit.budget.<CURRENCY>.perCall` = integer per call
* `limit.net.maxBytes` / `limit.net.maxCalls`
* `limit.fs.maxBytes` / `limit.fs.maxCalls`
* `limit.tool.<toolName>.maxCalls` (optional granularity)

### Function/Block requirements (developer intent → enforced)

* `requiresCapability` = list of **required** caps (resource\:verb\:scope)
  e.g., `env:read:OPENAI_API_KEY; model:use:mini`
* `requiresContext` = `config | normal | test`
* `requiresTool` (you already have)
* `pure` (you already have)
* `effects` = comma-sep from: `idempotent, readsClock, readsRand, readsNet, readsFS`

> Rationale: stays within your one-line `::key: value` format; easy to parse, lint, and diff; no new syntax.&#x20;

## 2) Enforcement order (deterministic)

At *tool call time*:

1. **Trust**: if tool `requires_trust` and `policyContext != config` ⇒ reject.
2. **Allow/Deny**: name/pattern in `policyAllow` and **not** in `policyDeny`.
3. **Capabilities**: tool’s `RequiredCaps` ⊆ union of file-level `grant.*` plus runtime policy (CLI/baked-in).

   * Structural match: `resource` equal, `verb` required ⊆ granted, `scope` required ⊆ granted (wildcards OK).
4. **Limits**: decrement counters; reject on exhaustion.
5. **Function requirements** (if inside a func): `requiresCapability` and `requiresContext` must be satisfied (these are *additional* constraints, not grants).

Merging/precedence:

* **Baked-in** (binary) ≤ **CLI** policy ≤ **File header** `::policy*` ≤ **Function block** `::requires*` (most restrictive wins).
* **Deny** always wins over allow.

## 3) Effects (for reasoning, caching, and tests)

Store on the tool registry + optionally on the function:

* `pure: true|false`
* `effects: idempotent, readsClock, readsRand, readsNet, readsFS`

Interpreter/linter use:

* Reject “pure only” scripts if an impure call is present.
* Test harness can force deterministic `clock`/`rand` unless explicitly granted via metadata.

# How this looks in NS files

## A trusted config script header

```neuroscript
::schema: neuroscript
::serialization: ns
::description: Init agent models + sandbox for this workspace
::policyContext: config
::policyAllow: tool.agentmodel.Register, tool.agentmodel.Delete, tool.os.Getenv, tool.sandbox.SetProfile
::policyDeny: tool.fs.WriteRecursive

::grant.env.read: OPENAI_API_KEY, ANTHROPIC_API_KEY
::grant.model.admin: *
::grant.net.read: *.openai.com:443, api.anthropic.com:443
::grant.sandbox.admin: py-sandbox-1
::limit.budget.CAD.max: 500
::limit.budget.CAD.perCall: 50

# setup section...
must tool.agentmodel.Register("mini", {
  "provider": "openai",
  "model": "gpt-4o-mini",
  "api_key": tool.os.Getenv("OPENAI_API_KEY")
})

must tool.sandbox.SetProfile("py-sandbox-1")
```

## A normal (untrusted) script header

```neuroscript
::schema: neuroscript
::serialization: ns
::description: Use model 'mini' to summarize docs
::policyContext: normal
::policyAllow: tool.agentmodel.List  # read-only helpers ok
::grant.model.use: mini
::grant.net.read: *.openai.com:443
::limit.budget.CAD.max: 60
```

## Function-level requirements & effects

```neuroscript
func Summarize(needs txt returns out) means
  ::description: Summarize with the 'mini' model as JSON
  ::requiresContext: normal
  ::requiresCapability: model:use:mini
  ::effects: idempotent, readsNet
  ::pure: false
  ask "mini", "Summarize: {{txt}}" with {"json": true} into out
endfunc
```

If someone swaps `"mini"` for `"sonnet"` here, the runtime will reject unless `grant.model.use: sonnet` is present *and* the tool+net+budget envelopes are satisfied.

# Interpreter wiring (kept tight)

* **Tool meta** (unchanged from your plan): `RequiresTrust bool`, `RequiredCaps []Cap`.
* **ExecPolicy** (runtime): `Context string`, `Allow, Deny []Pattern`, `Grants []Cap`, `Limits map[...]Counters`.
* **Cap**: `{Resource, Verbs[], ScopeSet}`, scope matcher supports wildcard & prefix for `net`, glob for `fs`, exact list for `env/secrets/model`.

Gate:

```go
if t.RequiresTrust && policy.Context != "config" { deny }
if !allowed(t.Name, policy.Allow, policy.Deny) { deny }
if !capsSatisfied(t.RequiredCaps, policy.Grants) { deny }
if !limitsOk(t, policy.Limits) { deny }
// ok → call
```

# Provenance & audit

Emit on every `requires_trust` call and every networked `ask`:

* tool, callsite, capability IDs consumed, limit deltas, spend (currency/tokens), model name, host, duration.
* Attach a compact provenance tag to outputs:
  `prov: net[*.openai.com], model[mini], budget[CAD +0.23]`

# CI / Lint rules to add (cheap wins)

* Fail if `policyContext=config` is missing from a file that invokes any `requires_trust` tool.
* Warn if `effects` contradict `pure`.
* Fail if a function’s `requiresCapability` isn’t covered by file-level `grant.*` (and/or CLI policy) **in config runs**.
* Option: “deterministic test mode” that rejects `readsClock|readsRand` unless `grant.clock.read` or `grant.rand.read` present (or `seed:` provided).

# Baked-in minimal bootstrap (unchanged principle)

Ship the binary with **no** grants and **no** allows by default. Require the host to opt-in via CLI *or* a trusted config script:

* CLI: `--policy-context=config --allow=tool.agentmodel.Register,tool.os.Getenv --grant=env.read:OPENAI_API_KEY --grant=model.admin:*`
* Scripts then persist AgentModels; later runs flip to `policyContext=normal`.

---

If you want, I’ll draft:

1. A delta patch for `metadata.md` adding the `policy*`, `grant.*`, `limit.*`, `requiresCapability`, and `effects` keys (with examples).
2. A tiny `capmatch` package + unit tests (scope matching, limit counters, deny>allow precedence).
3. Linter checks described above.

This keeps us explicit, testable, and hard to foot-gun—without inventing a second config language or violating your metadata rules.&#x20;
