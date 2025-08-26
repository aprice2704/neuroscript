# Extra thoughts on capability gating (promoting funcs to root interp and bus emit)


GPT5:



I’ve walked through your capability source files (`builder.go`, `capability.go`, `constants.go`, `limits.go`, `matcher.go`). Here’s a detailed review and some recommendations — including how they line up with your earlier thought about gating promotions and bus emission.

---

## Overall architecture

* **Data model**: `Capability` with `Resource`, `Verbs`, `Scopes`.
* **Grants and limits**: `GrantSet` bundles `[]Capability`, `Limits`, and live `Counters`.
* **Builder/parsing**: `New`, `NewWithVerbs`, `Parse` and `MustParse` helpers.
* **Limits**: budget (per call, per run), FS, net, per-tool call counts.
* **Matcher**: `CapsSatisfied` + scope-specific matching semantics (wildcards, fs globs, net host\:port, etc.).
* **Constants**: standardized resource/verb strings (`fs`, `net`, `env`, `model`, …; `read`, `write`, `use`, etc.).

This is already a solid, minimal framework. It gives you both expressive (resource/verb/scope triples) and quantitative (limits/counters) levers.

---

## Strengths

1. **Flexible parsing**: string form `"resource:verb1,verb2:scope1,scope2"` keeps config ergonomic.
2. **Good separation**: verbs, scopes, limits, counters are all cleanly decoupled.
3. **Scope semantics**: tailored by resource — `fs` uses glob, `net` uses host\:port with wildcard, `env/model/etc.` use prefix/suffix wildcards.
4. **Error clarity**: clear error types (`ErrBudgetExceeded`, `ErrFSExceeded`, etc.).
5. **Safe defaults**: `NewCounters` allocates maps, so no `nil` panics.

---

## Gaps / Recommendations

### 1. Missing promotion & bus capabilities

Currently the constants cover `fs`, `net`, `env`, `model`, `tool`, `secret`, `budget`.
If promotion and bus emission are going to be first-class sensitive actions, you should add:

```go
const (
    ResModule = "module"   // promotion, rollback, stage, read
    ResBus    = "bus"      // emit, subscribe, request
)
```

And extend verbs to include e.g.:

```go
const (
    VerbPromote   = "promote"
    VerbStage     = "stage"
    VerbRollback  = "rollback"
    VerbEmit      = "emit"
    VerbSubscribe = "subscribe"
)
```

That lets you gate promotion requests as `Capability{Resource:"module", Verbs:{"promote"}, Scopes:{"agent/itg/*"}}` and bus sends as `Capability{Resource:"bus", Verbs:{"emit"}, Scopes:{"fdm.ingest.*"}}`.

### 2. Promotion granularity

You’ll want to treat **version bump level** and **visibility target** as scope attributes. For example:

* `"module:promote:agent/itg/*:level=patch:visibility=team"`
* That can either be encoded as multiple `Scopes` entries (`[]string{"agent/itg/*","level=patch","visibility=team"}`) or as a convention (`module:promote:agent/itg/*,team,patch`).

The matcher code already supports wildcard/prefix/suffix scope matching, so this works out of the box.

### 3. Bus scope semantics

Right now `matcher.go` doesn’t have explicit bus logic — `scopeMatch` handles `net`, `fs`, `env`, `model`, etc.. If you add `ResBus`, you should decide scope semantics:

* Probably treat bus topics like `net`: allow `*`, `prefix*`, or `fdm.ingest.*`.
* That means adding a `case "bus":` branch in `scopeMatch` with topic matching rules.

### 4. Quantitative gating

You already have `ToolMaxCalls`. For bus emission, add:

* `BusMaxEmits int` and `BusMaxBytes int64` to `Limits`
* `BusEmits` and `BusBytes` to `Counters`
* `CountBus(topic string, size int64)` like `CountNet`.

That ties into the rate/payload control we discussed.

### 5. Audit hooks

At the policy layer, add a convention that every **promotion commit** and **bus emit** call results in an audit event. You can reuse your bus for this: emit `audit.module.promote` and `audit.bus.emit` messages.

---

## Dating & versioning

Each file header already carries `// NeuroScript Version: 0.3.0` and `// File version: 1`. I’d recommend bumping **file versions** because we’re proposing schema changes:

* `constants.go`: **File version → 2** (add `ResModule`, `ResBus`, new verbs).
* `limits.go`: **File version → 2** (add bus limits/counters).
* `matcher.go`: **File version → 2** (add `case "bus":` scope match semantics).

The others (`builder.go`, `capability.go`) can stay at version 1 — no structural changes needed.

---

## Suggested next step

1. Extend `constants.go` with `ResModule`, `ResBus` and verbs.
2. Add bus limits and counters in `limits.go`.
3. Add `scopeMatch` case for `bus` in `matcher.go`.
4. Wire your capability enforcement into the promotion and bus codepaths — so `gw.Modules.Commit` and `bus.Emit` both call `CapsSatisfied`.

That way, promotion and bus emission are gated the same way as FS/Net/Model/etc., consistent and easy to audit.

---

Do you want me to draft the actual **code changes (diff-style patches)** for `constants.go`, `limits.go`, and `matcher.go` so you can drop them in directly?
