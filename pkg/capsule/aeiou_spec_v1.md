# AEIOU Envelope — v1 (capsule/aeiou/1)

Status: **Stable**
Spec ID: **capsule/aeiou/1**
Exec order: **U → I → E → A** (OUTPUT is inert)

This capsule defines a lean, UPSERTS-free protocol for AI↔NS collaboration.
Two *live* surfaces (**ACTIONS**, **IMPLEMENTATIONS**) plus optional **EVENTS**;
**OUTPUT** is prose only; **USERDATA** is host→AI payload only.

---

## 0. Optional envelope header (line 0)
A single JSON line preceding the first section header:

{"proto":"NSENVELOPE","v":1,"budgets":{"A":8192,"E":6144,"I":12288,"O":4096},"caps":["memory:write","module:promote"]}

If omitted, defaults apply. This header is *advisory*; authoritative checks are done by tools at runtime.

---

## 1. Sections and semantics

- **A — ACTIONS**: The **only** place for side-effects. Must be a single `command … endcommand` block. May call tools; all mutations are gated by capabilities, budgets, and policy.
- **E — EVENTS**: Reactive handlers `on event … endon`. Same side-effect rules as **A**. Handlers must be idempotent.
- **I — IMPLEMENTATIONS**: Pure functions `func … endfunc`. Loaded early; **no side-effects on load**. Callables for A/E.
- **O — OUTPUT**: Natural-language context / prompt / notes. Never parsed or executed.
- **U — USERDATA**: Host→AI data payload (text, JSON, etc.). The AI never writes **U**.

**Authoring order (recommended):** I, E, A, O, U.
**Execution order (normative):** U → I → E → A.

---

## 2. Hard rules (normative)

1. **Single side-effect surface:** Only **A** and **E** may cause effects, and only via tools.
2. **I is pure:** No IO, no tool calls at load time. Deterministic, referentially transparent.
3. **O is inert:** Never parsed, never executed.
4. **U is host-only:** The model does not write U.
5. **Determinism & replays:** A/E tool calls MUST be idempotent or guarded (CAS/transactions).
6. **Budget & caps enforced by interpreter:** Exceeding a section budget yields a diagnostic and the section is truncated; capability denials are structured errors.

---

## 3. Diagnostics (machine-parsable stubs)

On validation or runtime issues, emit these markers to **OUTPUT** (and logs):

- [[invalid:SECTION:reason]] e.g., [[invalid:ACTIONS:missing_command]]
- [[truncated:SECTION:bytes=N allowed=M]]
- [[denied:tool.NAME:capability_missing]]
- [[exceeded:SECTION:byte_budget allowed=M got=N]]

---

## 4. Budgets (default suggested)

- I: 12 KiB, E: 6 KiB, A: 8 KiB, O: 4 KiB
Hosts may override via the line-0 header. The interpreter is the source of truth.

---

## 5. Patterns

### 5.1 Probe-based discovery
Ask the runtime what you can do—no declarative wish list needed.

```neuroscript
command
  let caps = tool.system.Caps()
  if !caps["module:promote"] {
    emit "[info] no module:promote; staying ephemeral"
  }
endcommand
```

### 5.2 CAS-based memory mutation (replay-safe)

```neuroscript
command
  let cur, ver = tool.memory.Get("/projects/ns/status")
  let ok, _ = tool.memory.CAS("/projects/ns/status", ver, "ready")
  if !ok { emit "[retryable] memory CAS failed" }
endcommand
```

### 5.3 Plan → Apply (batch + policy)

```neuroscript
command
  let plan = [{"op":"set","path":"/a","value":"1"},{"op":"set","path":"/b","value":"2"}]
  emit "PLAN " + json(plan)

  if tool.policy.Allow(plan) {
    for step in plan {
      let _, ver = tool.memory.Get(step.path)
      tool.memory.CAS(step.path, ver, step.value)
    }
  } else {
    emit "[blocked] policy"
  }
endcommand
```

### 5.4 Ask-Loop (host-controlled)
When an agent model's `ToolLoopPermitted` flag is true, the host MAY feed the ACTIONS emit stream of turn N as
the OUTPUT of turn N+1. The model SHOULD signal intent explicitly by emitting:
  `[[loop:continue]]` | `[[loop:done]]` | `[[loop:abort:reason=...]]`
Hosts MUST enforce resource ceilings (turns/fuel/time) and SHOULD stop on no-progress.

---

## 6. Minimal envelope skeleton

<<<NSENVELOPE_MAGIC::IMPLEMENTATIONS::v1>>>
```neuroscript
func helper(x string) string
  return x + "-ok"
endfunc
```

<<<NSENVELOPE_MAGIC::ACTIONS::v1>>>
```neuroscript
command
  let caps = tool.system.Caps()
  emit "caps=" + json(caps)
  emit helper("ready")
endcommand
```

<<<NSENVELOPE_MAGIC::OUTPUT::v1>>>
```md
This section is inert. Use it for human-readable notes.
```

<<<NSENVELOPE_MAGIC::USERDATA::v1>>>
```json
{"example":true}
```

---

## 7. Security posture

- All tool calls are capability-gated and fuel-budgeted.
- No direct filesystem/network access from A/E; everything goes through tools.
- Promotion of code (making I persistent) is not automatic; it must go through an explicit tool pipeline and policy gates.

---

## 8. Success criteria (conformance)

A conforming interpreter MUST:
- Enforce U→I→E→A execution order.
- Reject/diagnose non-conforming sections and continue safely where possible.
- Gate all side-effects through tools with capability checks and budgets.
- Treat O as inert and U as host-only.

A conforming authoring agent SHOULD:
- Keep I small and composable; push effects into A/E; use probe/CAS/plan patterns.
- Emit diagnostics verbatim to help the next turn self-correct.

— end —