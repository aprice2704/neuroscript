# AEIOU Envelope — v2 (capsule/aeiou/2)

Status: **Stable**
Spec ID: **capsule/aeiou/2**
Exec order: **U → I → E → A** (OUTPUT is inert)

---

## 0. Core Principles

1.  **Block-Based Parsing**: The entire envelope is a single, self-contained block, making it easy to find and parse robustly. All parsing happens on the content *between* the `START` and `END` markers.
2.  **Unified Control Syntax**: All protocol control lines (delimiters, headers, diagnostics) share one consistent format: `<<<[MAGIC]:[VERSION]:[TYPE]:[OPTIONAL_JSON]>>>`.
3.  **Abstraction**: Agents SHOULD use the `tool.aeiou.magic(type, payload)` helper to generate these control strings, avoiding errors from hard-coded constants.

---

## 1. Envelope Structure (Normative)

An AEIOU envelope MUST be a single block of text delimited by `START` and `END` markers.

**Example:**
```
<<<NSENVELOPE_MAGIC_9E3B6F2D:V2:START>>>
# --- Envelope Content Goes Here ---
<<<NSENVELOPE_MAGIC_9E3B6F2D:V2:END>>>
```

---

## 2. Sections and Semantics

The content within the envelope is divided by section markers.

- **HEADER**: `<<<NSENVELOPE_MAGIC_9E3B6F2D:V2:HEADER:{"proto":"NSENVELOPE",...}>>>`
  - An optional, structured header containing advisory metadata like budgets and capabilities.
- **A — ACTIONS**: `<<<NSENVELOPE_MAGIC_9E3B6F2D:V2:ACTIONS>>>`
  - The **only** place for side-effects. A single `command` block.
- **E — EVENTS**: `<<<NSENVELOPE_MAGIC_9E3B6F2D:V2:EVENTS>>>`
  - Reactive `on event` handlers.
- **I — IMPLEMENTATIONS**: `<<<NSENVELOPE_MAGIC_9E3B6F2D:V2:IMPLEMENTATIONS>>>`
  - Pure, reusable `func` definitions. No side-effects on load.
- **O — OUTPUT**: `<<<NSENVELOPE_MAGIC_9E3B6F2D:V2:OUTPUT>>>`
  - Natural language context. Never executed.
- **U — USERDATA**: `<<<NSENVELOPE_MAGIC_9E3B6F2D:V2:USERDATA>>>`
  - Host-to-AI data payload. Never written by the AI.

**Execution order (normative):** U → I → E → A.

---

## 3. Diagnostics

Diagnostics MUST be emitted from the `ACTIONS` block using the unified format.

**Example:**
```neuroscript
command
  # Agent emits a structured diagnostic message
  let diag = tool.aeiou.magic("DIAGNOSTIC", {"type":"denied", "tool":"tool.memory.write", "reason":"capability_missing"})
  emit diag
endcommand
```

---

## 4. Hard Rules

1.  **Parsing**: The interpreter first finds the `START`/`END` block, then parses its content. All other text is ignored.
2.  **Side-effects**: Only **A** and **E** may cause side-effects, and only via tools.
3.  **Purity**: **I** is pure. **O** is inert. **U** is host-only.
4.  **Idempotence**: A/E tool calls MUST be idempotent or guarded (e.g., CAS).

---

## 5. Conformance

- A conforming interpreter MUST extract the `START`/`END` block before parsing.
- A conforming agent MUST use `tool.aeiou.magic()` to generate all control strings.