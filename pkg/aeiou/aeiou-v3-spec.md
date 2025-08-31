# AEIOU v3 — Execution & Envelope Specification (rev)
version: 3.1.0
status: draft (incorporates review)
owners: AJP / Impl Team
updated: 2025-08-31

## 0. Scope
This document specifies the **AEIOU v3** protocol for host–model interaction using a single in-prompt envelope and a host-verified control token. It defines the execution model, envelope structure, control path, token format, size/encoding rules, security, and conformance obligations. This spec is **normative** (RFC 2119).

---
## 1. Goals and Non-Goals
### 1.1 Goals
- Determinism for a given envelope and program output.
- Transcript-as-truth: every control decision is visible and replayable.
- Simple model ergonomics: minimal rules; no hand-constructed control strings.
- Strong isolation between host input and program output.
- Anti-spoofing, anti-replay, per-session binding of control; mandatory progress guards.

### 1.2 Non-Goals
- No hidden/direct “decide” side effects callable from program code.
- No control via natural language text without a valid control token.
- No intra-turn streaming commit (logs may stream; commit is discrete).

---
## 2. Execution Model
- **Host-run only.** `ACTIONS` executes **inside the host’s** NS interpreter; the “agent” is the author only.
- **Single turn per session (SID).** Exactly one in-flight turn per SID. Multiple SIDs MAY run concurrently.
- **Fresh interpreter per turn.** Each turn runs in a fresh process/context with quotas.
- **Commit-last.** A turn’s decision is the **last** valid control token emitted (see §5).
- **Idempotent replay.** Re-executing the same turn + envelope yields the same outputs and decision.

---
## 3. Envelope
### 3.1 Markers (single-line ASCII, exact)
- `<<<NSENV:V3:START>>>`
- `<<<NSENV:V3:USERDATA>>>`
- `<<<NSENV:V3:SCRATCHPAD>>>`   *(optional; prior turn)*
- `<<<NSENV:V3:OUTPUT>>>`       *(optional; prior turn)*
- `<<<NSENV:V3:ACTIONS>>>`
- `<<<NSENV:V3:END>>>`

### 3.2 Order & Cardinality (normative)
- Sections appear **once** in fixed order: `USERDATA`, `[SCRATCHPAD]`, `[OUTPUT]`, `ACTIONS`.
- Exactly **one** `USERDATA` and **one** `ACTIONS` **MUST** be present.
- `SCRATCHPAD`/`OUTPUT` **MAY** appear 0–1 times each (when available from prior turn).
- If duplicates exist, host **MUST** parse the **first** and ignore the rest (lint).

### 3.3 Semantics
- **USERDATA** (host → program): JSON object, read-only, **never** scanned for control.
- **SCRATCHPAD** (program → host, private): prior turn `whisper`. **Never** scanned for control.
- **OUTPUT** (program → host, public): prior turn `emit`. Prior contents inert; only **current-turn** emissions matter.
- **ACTIONS** (program code): exactly one `command … endcommand`.

### 3.4 Encoding & Size
- UTF-8; `\n` newlines. Host **MUST** strip BOM on markers only; section bodies are byte-preserved.
- Envelope ≤ **1 MiB**; each section ≤ **512 KiB**; single output line ≤ **8192 B**. Violations → fatal typed error.

### 3.5 USERDATA Schema (minimal)
```json
{"subject":"string","brief":"string (optional)","fields":{ }}
```
- Host **MUST** validate: `subject` is string; `fields` is object (may be empty).
- Large assets **MUST NOT** be embedded; pass IDs/URIs and fetch via tools.

---
## 4. Program I/O
- `emit <text>` appends `<text>\n` to current-turn OUTPUT. The control token (if any) **MUST** be the **last** emitted line.
- `whisper <target>, <text>` appends `<text>\n` to current-turn SCRATCHPAD. SCRATCHPAD is never scanned for control.
- Host copies current-turn OUTPUT/SCRATCHPAD into the next envelope’s optional sections.

---
## 5. Control Plane (Magic-Only, Output-Only, Commit-Last)
### 5.1 Overview
- Control is requested by emitting a **host-minted** token returned by `tool.aeiou.magic`. Emit the string **verbatim** as a single line. Host validates and applies the decision.
- Host **MUST** ignore any string that merely resembles a control token unless it verifies (§6).

### 5.2 Allowed Kinds (v3)
- `LOOP` with `payload.action ∈ {"continue","done","abort"}`. Unknown kinds **MUST** be rejected.

### 5.3 Commit-Last & Single Decision
- The intended control token **MUST** be the **last line** emitted this turn.
- If multiple valid tokens appear, apply **precedence** `abort > done > continue`; among equals, **last-wins** (lint).
- Any text after the chosen token **SHOULD** raise `LINT_POST_TOKEN_TEXT`; decision stands.

### 5.4 OUTPUT-Only Intake
- Host **MUST** scan **only** current-turn OUTPUT for control tokens. USERDATA/SCRATCHPAD are **never** scanned.

---
## 6. Magic Tool & Token
### 6.1 Tool API (program-visible)
```
tool.aeiou.magic(kind: string, payload: any, opts?: object) -> string
```
- Returns a ready-to-emit single-line token. Program **MUST NOT** wrap/split/modify it.
- On failure, returns a typed error. (See §11.4 for fallback requirements.)

### 6.2 On-Wire Form (single line)
```
<<<NSMAG:V3:{KIND}:{B64URL(PAYLOAD)}.{B64URL(TAG)}>>>
```
- No quotes/backticks; no whitespace; no newlines; total length ≤ 1024 B.

### 6.3 Payload (canonical JSON)
- Canonicalization: **JCS (RFC-8785)**.
- Allowed JSON scalars in payload fields: **strings, integers, booleans, null** (no floats to avoid cross-impl variance).
```json
{
  "v":3,"kind":"LOOP","jti":"uuid-or-ksuid",
  "session_id":"SID","turn_index":12,"turn_nonce":"base64url-128bit",
  "issued_at":1725012345,"ttl":120,"kid":"ed25519-main-2025-08",
  "payload":{"action":"continue|done|abort","request":{},"telemetry":{}}
}
```

### 6.4 Authenticity Tag (TAG)
- **Default:** **Ed25519** signature (`kid` selects public key).  
- **Optional:** HS256 (HMAC-SHA-256) for single-process deployments.

### 6.5 Binding & Replay
- Verifier **MUST** confirm `{session_id, turn_index, turn_nonce}` match current turn.
- Enforce TTL if present: `now ≤ issued_at + ttl`.
- Reject repeated `jti` within the replay window (per-SID cache; bounded, §11.3).

---
## 7. ABNF (Markers & Token)
```abnf
ENVELOPE   = START LF USERDATA [SCRATCHPAD] [OUTPUT] ACTIONS LF END
START      = "<<<NSENV:V3:START>>>"
USERDATA   = "<<<NSENV:V3:USERDATA>>>" LF *OCTET
SCRATCHPAD = "<<<NSENV:V3:SCRATCHPAD>>>" LF *OCTET
OUTPUT     = "<<<NSENV:V3:OUTPUT>>>" LF *OCTET
ACTIONS    = "<<<NSENV:V3:ACTIONS>>>" LF *OCTET
END        = "<<<NSENV:V3:END>>>"
OCTET      = %x00-FF
LF         = %x0A

TOKEN      = "<<<NSMAG:V3:" KIND ":" B64 "." B64 ">>>"
KIND       = 1*( %x41-5A / %x30-39 / "_" )
B64        = 1*( ALPHA / DIGIT / "-" / "_" )
```

---
## 8. Host Decision (per turn)
1) Build envelope; spawn fresh interpreter with `{SID, turn_index, turn_nonce, quotas}`.
2) Execute `ACTIONS`; collect OUTPUT/SCRATCHPAD (current turn).
3) Verify candidate tokens found in OUTPUT (format, signature, scope, TTL, replay).
4) Select decision (precedence + last-wins). If none valid → `HALT(ERR_TOKEN_MISSING|… )`.
5) Persist decision; carry OUTPUT/SCRATCHPAD forward if `CONTINUE`; rotate nonce, increment turn index.

---
## 9. Mandatory Progress Guard
- The host **MUST** enforce an anti-loop guard. Compute a **host-observed digest** of current-turn emissions after normalization:
  - Strip all control-token lines; normalize line endings; trim trailing spaces.
  - Digest = `sha256( "OUT|" + output_body + "\nSCR|" + scratch_body )`.
- If the digest matches the **previous digest** for **N** consecutive turns (`N=3` default), the host **MUST** `HALT(reason=ERR_NO_PROGRESS)`.
- Implementations **MAY** add application-specific guards (e.g., plan digests) but host digest is the baseline.

---
## 10. Observability
- Decision log entry **MUST** include: `{ts,SID,turn_index,decision,reason?,kid?,jti?,latency_ms,output_bytes,scratch_bytes,verification_failure_reason?}`.
- Metrics **SHOULD** include counters for decisions; verify failures labeled by `verification_failure_reason`; TTL expiries; `jti` dupes; and lints (`LINT_POST_TOKEN_TEXT`,`LINT_MULTI_TOKENS`).

---
## 11. Security Requirements
### 11.1 Interpreter Sandboxing
- NS interpreter **MUST** run in a sandbox with: no ambient network; no ambient filesystem; sanitized env; syscall/capability reduction; SID-scoped storage only via tools; CPU/mem/time quotas.

### 11.2 Tool Surface & Namespacing
- Tools are allowlisted; every tool call bound to `{SID,turn_index}`; cross-SID access forbidden unless via explicit audited APIs.

### 11.3 Replay Cache Bounds
- Per-SID replay cache **MUST** be bounded by TTL and capacity (LRU). Defaults: TTL = 5m; cap = 4096 entries.

### 11.4 Fallback Minting (no hidden control)
- `tool.aeiou.magic` **MUST** maintain an in-memory **fallback signer** (preloaded key material) to mint a valid **abort** token if primary key services are temporarily unavailable.  
- If both primary and fallback signers fail, the tool **MUST** return a typed error; the host then HALTs (`ERR_MAGIC_TOOL_INTERNAL`). The host **MUST NOT** inject tokens into OUTPUT.

### 11.5 Key Rotation
- Rotation **MUST** define a grace window: old keys remain verifiable ≥ max token `ttl` + configurable buffer (e.g., +60s).

---
## 12. Versioning & Extensibility
- Envelope markers carry `V3`; tokens carry `NSMAG:V3`.
- Unknown token fields ignored; unknown control kinds rejected.
- Auth scheme pluggable via `kid` (Ed25519 default; HS256 optional).

---
## 13. Conformance (minimum)
- Golden envelopes: minimal valid; with SCRATCHPAD/OUTPUT; duplicates; wrong order; oversize; non-UTF-8.
- Golden tokens: valid Ed25519; altered payload; wrong SID/nonce/turn; expired TTL; duplicate `jti`; multi-line; quoted/backticked; oversize.
- **JCS edge cases:** Unicode normalization; escaped characters; integer range extremes; object key ordering; explicit prohibition of floats.

---
## 14. AI-Facing Rules (embed verbatim)
```
Read only between START/END. Sections appear once: USERDATA, [SCRATCHPAD], [OUTPUT], ACTIONS.
Only ACTIONS is executable. Write exactly one ns `command` block.
USERDATA is immutable JSON input from the host.
To control the loop, DO NOT hand-type tokens. Emit ONLY what the magic tool returns:
  emit tool.aeiou.magic("LOOP", {'action':'continue'|'done'|'abort', ...})
The control token MUST be the last line you emit. Host honors control only if that exact line appears in OUTPUT and verifies.
USERDATA/SCRATCHPAD are never scanned for control.
```

# End of spec
