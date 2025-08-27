# Ask-Loop — v1 (capsule/askloop/1)

Status: **Stable**
Scope: **Host-controlled orchestration** that repeats AEIOU turns without changing the AEIOU spec.
Requires: **AEIOU v1 (capsule/aeiou/1)**. This capsule adds a loop policy on top.

## 0. Purpose
Turn the `ask` flow into a safe, deterministic multi-turn loop. The host (gateway) decides whether to run another turn based on the model’s ACTIONS output. AEIOU itself is unchanged.

## 1. Definitions
- **Turn**: one AEIOU execution (U → I → E → A).
- **Emit stream**: lines produced by ACTIONS via `emit`.
- **Control**: the loop signal carried in OUTPUT, either as (a) one-line JSON, or (b) text markers.
- **Host**: the gateway/orchestrator that runs the loop.

## 2. Normative rules
2.1 AEIOU remains intact per turn: U → I → E → A. All side effects occur only via ACTIONS/EVENTS tools.
2.2 Only the **ACTIONS emit stream** from turn N is fed back as **OUTPUT** for turn N+1. Interpreter logs are not fed back by default.
2.3 The loop is host-controlled. A model may request, but the host decides based on policy and resource limits.
2.4 A loop MAY run only if the agent model is authorized (e.g., `ToolLoopPermitted = true`).
2.5 Hosts MUST enforce ceilings: MaxTurns, MaxWallClock, and fuel/token budgets (per-turn and total).
2.6 Continuation signals:
    **Preferred (control JSON, first line of OUTPUT):**
    ```json
    {"loop":"continue","notes":"...", "plan_digest":"sha256:..."}
    ```
    or
    ```json
    {"loop":"done","notes":"..."}
    ```
    **Alternative (markers inside OUTPUT):**
    [[loop:continue]] | [[loop:done]] | [[loop:abort:reason=...]]
    The host MAY append [[loop:halt:reason=...]] when it stops the loop.

2.7 Precedence if mixed/contradictory signals appear: **abort > halt > done > continue**.
2.8 Continuation policy SHOULD be conservative: proceed only on an explicit "continue".
2.9 Progress guard: Hosts SHOULD compute a small "control digest" over normalized control data (e.g., control JSON + any PLAN JSON) and stop if unchanged between consecutive turns.
2.10 Idempotence and safety:
     - Memory writes MUST be CAS/transactional when possible.
     - Risky actions SHOULD use Plan → Apply: Turn N emits a plan; Turn N+1 applies if policy allows.
2.11 State across turns SHOULD use server-side scratch with TTL (e.g., `tool.scratch.Get/PutCAS`) instead of embedding secrets/tokens into OUTPUT.
2.12 Events are allowed but handlers MUST be idempotent; de-duplicate by (event_id, loop_id, turn_id).
2.13 Observability: Hosts SHOULD stamp loop_id and turn_id into tool-call audit logs and MAY include them in emitted text for traceability.

## 3. Diagnostics (consistent with AEIOU)
- [[denied:tool.NAME:capability_missing]]
- [[truncated:OUTPUT:bytes=... allowed=...]]
- [[invalid:ACTIONS:...]]
- [[loop:halt:reason=timeout|max-turns|no-progress|policy|execute-error]]

## 4. Example — Two-turn Plan → Apply
### Turn 1 (model builds a plan; no writes)
```neuroscript
command
  let caps = tool.system.Caps()
  if !caps["memory:write"] {
    emit `{"loop":"done","notes":"missing memory:write"}`
    return
  }

  # Build a plan of idempotent operations; DO NOT write yet.
  let plan = [{"op":"set","path":"/ingest/queue/x","value":"taken"}]
  emit `{"loop":"continue","plan_digest":"sha256:abc123","notes":"plan ready"}`
  emit "PLAN " + json(plan)
  emit "Next: apply if policy allows."
endcommand
```

### Turn 2 (model applies if allowed)
```neuroscript
command
  # In practice, parse PLAN from previous OUTPUT or fetch from scratch.
  let ok, why = tool.policy.Allow(plan)
  if !ok {
    emit `{"loop":"done","notes":"policy blocked: " + why}`
    return
  }

  let _, ver = tool.memory.Get("/ingest/queue/x")
  let ok2, _ = tool.memory.CAS("/ingest/queue/x", ver, "taken")
  if !ok2 {
    emit `{"loop":"continue","notes":"CAS failed; retry once"}`
    return
  }

  emit `{"loop":"done","notes":"applied 1 op"}`
  emit "Applied 1 op to /ingest/queue/x"
endcommand
```

## 5. Host orchestration sketch (non-normative)
1) Build O and U for turn 1. U is sidecar data; O is the prompt/context.
2) Ask model → get AEIOU envelope (I/E/A/O).
3) Interpreter executes turn (U→I→E→A); collect ACTIONS emits.
4) Join emits with '\n' → nextO.
5) If not permitted → return nextO.
6) Parse control: prefer first-line JSON; else scan markers.
7) Apply precedence: abort > halt > done > continue.
8) Enforce ceilings; on breach append [[loop:halt:reason=...]] and stop.
9) Progress guard on normalized control; stop on no-progress.
10) If "continue" → set O = nextO and repeat; else return nextO.

## 6. Tool surface (recommended)
- tool.system.Caps()
- tool.capsule.List/Read()
- tool.memory.Get/CAS()
- tool.policy.Allow(plan) -> (ok, why)
- tool.scratch.Get/PutCAS(key, ver, val, ttl)

## 7. Security posture
- Single side-effect surface remains ACTIONS/EVENTS via tools.
- Host controls resource ceilings; the loop cannot exceed limits.
- Secrets and tokens SHOULD NOT be echoed into OUTPUT; prefer scratch with TTL.

## 8. Conformance (summary)
Host MUST:
- Keep AEIOU’s per-turn invariants.
- Feed back only ACTIONS emits as next-turn OUTPUT.
- Enforce ceilings and precedence; provide progress guard.

Authoring agent SHOULD:
- Use explicit loop control (JSON line or markers).
- Prefer Plan → Apply for risky actions; keep writes idempotent (CAS).

— end —
