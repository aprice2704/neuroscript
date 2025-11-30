# AEIOU v4 — Execution & Envelope Specification
version: 4.0.0
status: active
owners: AJP / Impl Team
updated: 2025-11-27

## 0. Scope
This document specifies the **AEIOU v4** protocol for host–model interaction using a single in-prompt envelope and a simplified, host-managed control signal. It defines the execution model, envelope structure, text-marker control, and security obligations. This spec is **normative** (RFC 2119).

---
## 1. Goals and Non-Goals
### 1.1 Goals
* **Simplicity:** Minimal cognitive load for the Model; use standard text markers (`<<<LOOP:DONE>>>`).
* **Host-Managed Security:** Security is enforced by the Host via pre-execution AST validation, not by the Model via cryptographic tokens.
* **Determinism:** For a given envelope and program output.
* **Transcript-as-truth:** Every control decision is visible and replayable.
* **Strong isolation:** Between host input (`USERDATA`) and program output (`ACTIONS`).

### 1.2 Non-Goals
* No client-side signing or cryptographic proofing (removed in v4).
* No intra-turn streaming commit (logs may stream; commit is discrete).

---
## 2. Execution Model
* **Host-run only:** `ACTIONS` executes **inside the host’s** NS interpreter.
* **Single turn per session (SID):** Exactly one in-flight turn per SID.
* **Fresh interpreter per turn:** Each turn runs in a fresh process/context with quotas.
* **Pre-Execution Validation:** The Host **MUST** parse and validate the `ACTIONS` AST (e.g., checking allowed tools) *before* execution begins.
* **Idempotent replay:** Re-executing the same turn + envelope yields the same outputs.

---
## 3. Envelope
### 3.1 Markers (single-line ASCII, exact)
* `<<<NSENV:V4:START>>>`
* `<<<NSENV:V4:USERDATA>>>`
* `<<<NSENV:V4:SCRATCHPAD>>>`   *(optional; prior turn)*
* `<<<NSENV:V4:OUTPUT>>>`       *(optional; prior turn)*
* `<<<NSENV:V4:ACTIONS>>>`
* `<<<NSENV:V4:END>>>`

### 3.2 Order & Cardinality (normative)
* Sections appear **once** in fixed order: `USERDATA`, `[SCRATCHPAD]`, `[OUTPUT]`, `ACTIONS`.
* Exactly **one** `USERDATA` and **one** `ACTIONS` **MUST** be present.
* `SCRATCHPAD`/`OUTPUT` **MAY** appear 0–1 times each (when available from prior turn).
* If duplicates exist, host **MUST** parse the **first** and ignore the rest (lint).

### 3.3 Semantics
* **USERDATA** (host → program): JSON object, read-only.
* **SCRATCHPAD** (program → host, private): prior turn `whisper`.
* **OUTPUT** (program → host, public): prior turn `emit`.
* **ACTIONS** (program code): exactly one `command … endcommand`.

### 3.4 Encoding & Size
* UTF-8; `\n` newlines.
* Envelope ≤ **1 MiB**; each section ≤ **512 KiB**. Violations → fatal typed error.

### 3.5 USERDATA Schema (minimal)
```json
{"subject":"string","brief":"string (optional)","fields":{ }}
```

---
## 4. Program I/O
* `emit <text>` appends `<text>\n` to current-turn OUTPUT.
* `whisper <target>, <text>` appends `<text>\n` to current-turn SCRATCHPAD.
* Host copies current-turn OUTPUT/SCRATCHPAD into the next envelope’s optional sections.

---
## 5. Control Plane (Text Marker)
### 5.1 Overview
Control is requested by emitting a specific text marker. The Host scans the `emit` stream for this marker to determine if the loop should terminate.

### 5.2 Control Marker (V4)
* **Marker:** `<<<LOOP:DONE>>>`
* **Semantics:** Signals that the Agent has completed its task.
* **Payload:** Any text immediately following the marker on the same line (or subsequent lines if the host supports it) is considered the **Final Result**.

### 5.3 Commit & Decision
* If `<<<LOOP:DONE>>>` is found in the emission stream:
    * The loop transitions to **DONE**.
    * The **Final Result** is extracted from the emission text.
* If the marker is **not** found:
    * The loop transitions to **CONTINUE** (unless `MaxTurns` is reached).
    * The output is fed back into the next turn.

---
## 6. (Removed)
*V4 removes the "Magic Tool" and cryptographic token requirements.*

---
## 7. ABNF (Markers)
```abnf
ENVELOPE   = START LF USERDATA [SCRATCHPAD] [OUTPUT] ACTIONS LF END
START      = "<<<NSENV:V4:START>>>"
USERDATA   = "<<<NSENV:V4:USERDATA>>>" LF *OCTET
SCRATCHPAD = "<<<NSENV:V4:SCRATCHPAD>>>" LF *OCTET
OUTPUT     = "<<<NSENV:V4:OUTPUT>>>" LF *OCTET
ACTIONS    = "<<<NSENV:V4:ACTIONS>>>" LF *OCTET
END        = "<<<NSENV:V4:END>>>"

CONTROL_MARKER = "<<<LOOP:DONE>>>" [SP *OCTET]
```

---
## 8. Host Decision (per turn)
1) Build envelope; spawn fresh interpreter with `{SID, turn_index}`.
2) **Validate AST:** Parse `ACTIONS` and check permissions (`api.CheckScriptTools`). If invalid → `HALT`.
3) Execute `ACTIONS`; collect OUTPUT/SCRATCHPAD.
4) **Scan for Marker:** Check OUTPUT for `<<<LOOP:DONE>>>`.
5) Select decision: `DONE` if marker found; otherwise `CONTINUE`.

---
## 9. Mandatory Progress Guard
* The host **MUST** enforce an anti-loop guard. Compute a **host-observed digest** of current-turn emissions after normalization.
* Digest = `sha256( "OUT|" + output_body + "\nSCR|" + scratch_body )`.
* If the digest matches the **previous digest** for **N** consecutive turns (`N=3` default), the host **MUST** `HALT(reason=ERR_NO_PROGRESS)`.

---
## 10. Observability
* Decision log entry **MUST** include: `{ts, SID, turn_index, decision, latency_ms, output_bytes, final_result?}`.
* Metrics **SHOULD** include counters for: `decisions`, `validations_failed`, `progress_halts`.

---
## 11. Security Requirements (Host-Managed)
### 11.1 Pre-Execution Validation
* The Host **MUST NOT** execute code that violates the Agent's permissions.
* Before execution, the Host MUST walk the AST to verify that all called tools are permitted by the `AgentModel` configuration.

### 11.2 Interpreter Sandboxing
* NS interpreter **MUST** run in a sandbox with: no ambient network; no ambient filesystem; SID-scoped storage only via tools; CPU/mem/time quotas.

---
## 12. Versioning
* Envelope markers carry `V4`.
* Backwards compatibility with V3 (Tokens) is **NOT** required.

---
## 13. AI-Facing Rules (embed verbatim)
```
Read only between START/END. Sections appear once: USERDATA, [SCRATCHPAD], [OUTPUT], ACTIONS.
Only ACTIONS is executable. Write exactly one ns `command` block.
USERDATA is immutable JSON input from the host.
To control the loop:
  - To Continue: Simply emit text/tools.
  - To Finish: emit "<<<LOOP:DONE>>> <Optional Final Result>"
The control marker MUST be in the output for the loop to stop.
```

# End of spec

::schema: spec
::serialization: md
::id: capsule/aeiou
::version: 4
::fileVersion: 1
::author: NeuroScript Docs Team
::modified: 2025-11-27
::description: Defines the AEIOU Envelope Protocol v4, a host-controlled protocol for AI agents.