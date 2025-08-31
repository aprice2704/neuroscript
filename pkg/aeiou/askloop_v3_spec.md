# Ask-Loop v3 — Host Loop Controller Specification (rev)
version: 3.1.0
status: draft (incorporates review)
owners: AJP / Impl Team
updated: 2025-08-31

## 0. Scope
Defines the **host loop controller** executing AEIOU v3 envelopes: turn lifecycle, token verification, decision selection, progress guard, and observability. Pairs with “AEIOU v3 — Execution & Envelope Specification”.

---
## 1. Terms & Decisions
- **SID**: Session ID (conversation).
- **Turn**: One execution of `ACTIONS` → exactly one decision.
- **Interpreter**: Fresh NS interpreter per turn.
- **Token**: `NSMAG:V3` line returned by `tool.aeiou.magic` and emitted by the program.
- **Decision**: `{CONTINUE, DONE, ABORT, HALT}`. `HALT` is host-only.

---
## 2. Lifecycle
- `IDLE` → `EXECUTING` → `VERIFYING` → `DECIDED` → (`EMIT_NEXT` | `FINAL`)
- Start turn: mint `turn_nonce`, set `{SID, turn_index}`; build envelope; start interpreter with quotas & sandbox.

---
## 3. Concurrency & Isolation
- Exactly **one** in-flight turn per SID (per-SID mutex/singleflight).
- Multiple SIDs MAY run concurrently. Tools read `{SID,turn_index}` from context; cross-SID access forbidden unless explicitly allowed and audited.

---
## 4. Program Execution Rules
- Execute only the single `command … endcommand` block in `ACTIONS`.
- `emit` → OUTPUT (current turn); `whisper` → SCRATCHPAD (current turn).
- Intended control token **MUST** be the **last emitted line** (commit-last).
- Host **MUST NOT** scan USERDATA or SCRATCHPAD for control.

---
## 5. Token Intake, Verification, Selection
### 5.1 Intake
- Scan only current-turn OUTPUT lines in emission order. Candidate = `NSMAG:V3…` shape, ≤ 1024 B.

### 5.2 Verification
- Parse; base64url-decode; canonicalize payload (JCS); verify TAG (Ed25519 default; HS256 optional).
- Scope checks: `{session_id == SID, turn_index == current, turn_nonce == current}`.
- TTL (if present): `now ≤ issued_at + ttl`.
- Replay: per-SID `jti` cache (bounded LRU + TTL). Reuse → reject.

### 5.3 Selection
- If ≥1 valid: precedence `abort > done > continue`; among equals, **last-wins**.
- If any text follows the chosen token: log `LINT_POST_TOKEN_TEXT`.
- If 0 valid: `HALT` with typed reason (`ERR_TOKEN_MISSING|VERIFY|SCOPE|TTL|REPLAY`).

---
## 6. Mandatory Progress Guard
- Compute host digest per turn: strip control tokens; normalize (`\n`, trim trailing spaces); `sha256("OUT|"+out+"\nSCR|"+scr)`.
- If digest repeats for **N** consecutive turns (`N=3` default): `HALT(reason=ERR_NO_PROGRESS)`.
- Config: `loop.no_progress_n` (default 3). This guard is **MUST**. Additional app-specific guards MAY be added.

---
## 7. Timeouts, Quotas, Sandbox
- Each turn runs with wall/CPU/memory quotas; on breach → terminate interpreter → `HALT(ERR_TIMEOUT|ERR_QUOTA)`.
- Interpreter runs in a sandbox: no ambient network/filesystem, reduced capabilities; all IO via tools; SID-scoped storage only.

---
## 8. Errors & Lints (Typed)
### Fatal → HALT or rejection
- `ERR_ENV_MARKERS_INVALID`, `ERR_ENV_SECTION_MISSING`, `ERR_ENV_ORDER`, `ERR_ENV_SECTION_DUP`, `ERR_ENV_SIZE`
- `ERR_USERDATA_SCHEMA`
- `ERR_TOKEN_PARSE`, `ERR_TOKEN_VERIFY`, `ERR_TOKEN_SCOPE`, `ERR_TOKEN_TTL`, `ERR_TOKEN_REPLAY`, `ERR_TOKEN_MISSING`
- `ERR_MAGIC_TOOL_INTERNAL` (minting failure after fallback attempt)
- `ERR_TIMEOUT`, `ERR_QUOTA`, `ERR_NO_PROGRESS`

### Lints (decision stands)
- `LINT_POST_TOKEN_TEXT`, `LINT_MULTI_TOKENS`, `LINT_DUP_SECTION_IGNORED`

---
## 9. Observability
- Decision log **MUST** record: `{ts,SID,turn_index,decision,reason?,kid?,jti?,latency_ms,output_bytes,scratch_bytes,verification_failure_reason?}`.
- Metrics **SHOULD** track: decisions; verify_failures by reason; TTL expiries; `jti` dupes; lints (`POST_TOKEN_TEXT`, `MULTI_TOKENS`); progress-guard HALTs.

---
## 10. Key Management & Rotation
- `kid` selects key. Rotation retains old keys ≥ max token `ttl` + grace (e.g., 60s) for verification.
- Fallback signer kept warm in memory; used to mint an **abort** token if primary signer unavailable.

---
## 11. Pseudocode (normative core)
```go
func RunTurn(ctx, sid string, turnIdx int, env Envelope) Decision {
  nonce := NewTurnNonce()
  ctx = WithTurnContext(ctx, sid, turnIdx, nonce, quotasFromConfig())
  ensureSingleInflightTurn(sid)

  out, scr := ExecuteInFreshInterpreter(ctx, env.Actions)

  // Token verification
  valids := verifyCandidates(out.lines, sid, turnIdx, nonce, now())
  dec := selectDecision(valids) // precedence + last-wins; none -> HALT(ERR_TOKEN_MISSING)

  // Progress guard (strip tokens, normalize, digest)
  dg := digestHostObserved(out.bodyWithoutTokens(), scr.bodyNormalized())
  if repeatedDigest(sid, dg) >= cfg.NoProgressN { dec = Decision{Kind: HALT, Reason: "ERR_NO_PROGRESS"} }

  persistTurn(sid, turnIdx, out, scr, dec)
  if dec.Kind == CONTINUE { scheduleNext(sid, turnIdx+1) }
  return dec
}
```

# End of spec
