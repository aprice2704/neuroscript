# Ask-Loop — v2 (capsule/askloop/2)

Status: **Stable**
Scope: **Host-controlled orchestration** of AEIOU turns.
Requires: **AEIOU v2 (capsule/aeiou/2)**

---

## 0. Purpose

To define a safe, deterministic, multi-turn loop for agentic workflows, built on top of the AEIOU v2 protocol. The host gateway controls the loop based on structured signals from the agent.

---

## 1. Core Principles

1.  **Host-Controlled**: The host is the ultimate authority on loop continuation. It enforces all policies and resource limits.
2.  **Unified Control Syntax**: Loop control signals use the same unified `<<<...>>>` format as the rest of the AEIOU v2 protocol, generated via `tool.aeiou.magic()`.
3.  **Stateless Feedback**: Only the `emit` stream from an `ACTIONS` block is fed back as the `OUTPUT` for the next turn.

---

## 2. Loop Control (Normative)

The agent signals its intent to continue, stop, or abort by emitting a `LOOP` control string from its `ACTIONS` block.

**How to Signal:**
```neuroscript
command
  # Use the magic tool to create a loop control message with a JSON payload.
  let signal = tool.aeiou.magic("LOOP", {
    "control": "continue",
    "notes": "Plan is ready, proceeding to apply.",
    "plan_digest": "sha256:abc123..."
  })
  emit signal
endcommand
```

**Payload Schema:**
- `control` (string, required): One of `continue`, `done`, `abort`.
- `notes` (string, optional): Human-readable rationale.
- `reason` (string, optional): Machine-readable reason, especially for `abort`.
- Any other fields are for logging and diagnostics.

**Host Response:** The host parses these `LOOP` signals from the `emit` stream to decide the next action.

---

## 3. Host Responsibilities

1.  **Authorization**: A loop MAY only run if the agent model has the `ToolLoopPermitted` flag set.
2.  **Resource Ceilings**: The host MUST enforce `MaxTurns`, `MaxWallClock`, and fuel/token budgets. If a limit is breached, the host emits a `HALT` signal and terminates the loop.
    - **Example Host Signal:** `<<<NSENVELOPE_MAGIC_9E3B6F2D:V2:HALT:{"reason":"max-turns"}>>>`
3.  **Precedence**: If multiple or conflicting signals appear, the host MUST resolve them in the order: `abort` > `halt` > `done` > `continue`.
4.  **Progress Guard**: The host SHOULD track a digest of the control payload and halt the loop if no progress is being made between turns.

---

## 4. Example: Two-Turn Plan → Apply

### Turn 1: Model proposes a plan
```neuroscript
# ACTIONS block
command
  let plan = [{"op":"set", "path":"/q/x", "value":"taken"}]
  emit tool.aeiou.magic("LOOP", {"control":"continue", "notes":"Plan ready."})
  emit "PLAN: " + json(plan)
endcommand
```

### Turn 2: Model applies the plan
```neuroscript
# ACTIONS block of the next turn
command
  # (Code to parse plan from previous OUTPUT and get policy approval)
  let ok = tool.memory.CAS("/q/x", "expected_ver", "taken")
  if !ok {
    emit tool.aeiou.magic("LOOP", {"control":"abort", "reason":"cas-failed"})
  } else {
    emit tool.aeiou.magic("LOOP", {"control":"done", "notes":"Successfully applied plan."})
  }
endcommand
```