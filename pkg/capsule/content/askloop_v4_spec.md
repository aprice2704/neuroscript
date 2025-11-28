# Ask-Loop v4 — Host Loop Controller Specification
version: 4.0.0
status: active
owners: AJP / Impl Team
updated: 2025-11-27

## 0. Scope
Defines the **host loop controller** executing AEIOU v4 envelopes: turn lifecycle, control marker detection, decision selection, progress guard, and observability. Pairs with “AEIOU v4 — Execution & Envelope Specification”.

---
## 1. Terms & Decisions
* **SID**: Session ID (conversation).
* **Turn**: One execution of `ACTIONS` → exactly one decision.
* **Interpreter**: Fresh NS interpreter per turn.
* **Marker**: `<<<LOOP:DONE>>>` string emitted by the program to signal completion.
* **Decision**: `{CONTINUE, DONE, HALT}`. `HALT` is host-initiated (error/limit).

---
## 2. Lifecycle
* `IDLE` → `EXECUTING` → `SCANNING` → `DECIDED` → (`EMIT_NEXT` | `FINAL`)
* Start turn: set `{SID, turn_index}`; build envelope; start interpreter with quotas & sandbox.

---
## 3. Concurrency & Isolation
* Exactly **one** in-flight turn per SID (per-SID mutex/singleflight).
* Multiple SIDs MAY run concurrently. Tools read `{SID, turn_index}` from context.

---
## 4. Program Execution Rules
* Execute only the single `command … endcommand` block in `ACTIONS`.
* `emit` → OUTPUT (current turn); `whisper` → SCRATCHPAD (current turn).
* To signal completion, the program **MUST** emit the `<<<LOOP:DONE>>>` marker.
* Any text following the marker on the same line is captured as the **Final Result**.
* Host **MUST NOT** scan USERDATA or SCRATCHPAD for control signals.

---
## 5. Marker Intake & Decision
### 5.1 Intake
* Scan current-turn OUTPUT lines in emission order.
* Target: The exact string literal `<<<LOOP:DONE>>>`.

### 5.2 Selection
* **DONE:** If the marker is found.
    * The text immediately following the marker (and potentially subsequent lines, implementation dependent) is captured as the result.
    * The loop terminates successfully.
* **CONTINUE:** If the marker is **not** found.
    * The loop continues to the next turn (feeding OUTPUT back into the next envelope).
* **HALT:** If `MaxTurns` is reached without a DONE signal.

### 5.3 Ambiguity Handling
* If multiple markers appear, the **first** valid marker triggers the decision (fail-fast completion).

---
## 6. Mandatory Progress Guard
* Compute host digest per turn: strip control markers; normalize (`\n`, trim trailing spaces); `sha256("OUT|"+out+"\nSCR|"+scr)`.
* If digest repeats for **N** consecutive turns (`N=3` default): `HALT(reason=ERR_NO_PROGRESS)`.
* Config: `loop.no_progress_n` (default 3). This guard is **MUST**.

---
## 7. Timeouts, Quotas, Sandbox
* Each turn runs with wall/CPU/memory quotas; on breach → terminate interpreter → `HALT(ERR_TIMEOUT|ERR_QUOTA)`.
* Interpreter runs in a sandbox: no ambient network/filesystem, reduced capabilities; all IO via tools; SID-scoped storage only.
* **Pre-Execution Validation:** The Host MUST validate the AST permissions before running the code.

---
## 8. Errors & Lints (Typed)
### Fatal → HALT or rejection
* `ERR_ENV_MARKERS_INVALID`, `ERR_ENV_SECTION_MISSING`, `ERR_ENV_ORDER`
* `ERR_USERDATA_SCHEMA`
* `ERR_TIMEOUT`, `ERR_QUOTA`, `ERR_NO_PROGRESS`
* `ERR_MAX_TURNS_EXCEEDED`

### Lints (decision stands)
* `LINT_DUP_SECTION_IGNORED`
* `LINT_MULTIPLE_MARKERS` (if program emits DONE multiple times)

---
## 9. Observability
* Decision log **MUST** record: `{ts, SID, turn_index, decision, latency_ms, output_bytes, scratch_bytes}`.
* Metrics **SHOULD** track: decisions (`DONE`/`CONTINUE`); HALTs (`progress`, `timeout`, `quota`); turn counts.

---
## 10. Pseudocode (normative core)
```go
func RunTurn(ctx, sid string, turnIdx int, env Envelope) Decision {
  ctx = WithTurnContext(ctx, sid, turnIdx, quotasFromConfig())
  ensureSingleInflightTurn(sid)

  // 1. Validation (Host-Managed Security)
  tree := Parse(env.Actions)
  if err := ValidatePermissions(tree); err != nil {
      return Decision{Kind: HALT, Reason: "ERR_PERMISSIONS"}
  }

  // 2. Execution
  out, scr := ExecuteInFreshInterpreter(ctx, tree)

  // 3. Decision
  if marker, result := ScanForDoneMarker(out); marker {
      return Decision{Kind: DONE, Result: result}
  }

  // 4. Progress Guard
  dg := digestHostObserved(out, scr)
  if repeatedDigest(sid, dg) >= cfg.NoProgressN {
      return Decision{Kind: HALT, Reason: "ERR_NO_PROGRESS"}
  }

  persistTurn(sid, turnIdx, out, scr, CONTINUE)
  return Decision{Kind: CONTINUE}
}
```

# End of spec