# Checklist: Upgrading Interpreter to AEIOU v3

:: version: 1.1.0
:: updated: 2025-08-31

This checklist outlines the necessary changes within the `pkg/interpreter` to support the AEIOU v3 protocol, incorporating the new `llmconn` package.

---

## 1. Core `ask` Statement Logic (`interpreter_steps_ask.go`)

- |x| **Overhaul `executeAsk`** #(1a01)
  - [x] Create a new `llmconn.LLMConn` using the specified agent model and provider. #(1a02)
  - [x] Create the first `aeiou.Envelope` (V3) with a `USERDATA` section containing the initial prompt. #(1a03)
  - |x| Implement Host-Side Loop #(1a04)
    - [x] Loop based on the `max_turns` property of the agent model. #(1a05)
    - [x] In each turn, call `llmconn.Converse()` with the current envelope. #(1a06)
    - [x] Parse the `ACTIONS` from the AI's response envelope. #(1a07)
    - [x] Execute the `ACTIONS` in a sandboxed interpreter, capturing `emit` (OUTPUT) and `whisper` (SCRATCHPAD). #(1a08)
    - [x] **V3 Control:** Use `LoopController` to process OUTPUT and determine next action. #(1a09)
    - [x] **Progress Guard:** Implement the mandatory progress guard by hashing and comparing the combined OUTPUT and SCRATCHPAD between turns. #(1a10)
    - [x] **Prepare Next Envelope:** If the decision is `CONTINUE`, build the next envelope from captured outputs. #(1a11)
  - [x] **Final Result:** The captured `OUTPUT` from the final turn becomes the return value of the `ask` statement. #(1a12)

---

## 2. AEIOU Turn Execution (`interpreter_steps_ask_aeiou.go`)

- |x| **Update `executeAeiouTurn`** #(2b01)
  - [x] Confirm isolation via cloned interpreter. #(2b02)
  - [x] Ensure `emit` and `whisper` capture mechanisms are robust for populating the next turn's envelope sections. #(2b03)
  - [x] Remove all V2-specific logic. #(2b04)

---

## 3. Host Interaction & State Management

- [x] **Context Propagation:** Ensure session context (`SID`, `TurnIndex`, `TurnNonce`) is passed down through the execution path. #(3c01)

---

## 4. Tooling

- [x] **Implement `tool.aeiou.magic`:** Create the new tool responsible for minting V3 control tokens. #(4d01)

---

## 5. Cleanup and Testing

- |>| **Finalize Migration** #(5e01)
  - [x] **`interpreter_steps_ask_loop.go`:** Deleted. #(5e02)
  - [ ] **V2 Constants:** Search and remove any remaining V2-specific constants from the codebase. #(5e03)
  - [x] **Tests:** Rewrite all `ask` statement tests to validate the new V3 host-loop behavior. #(5e04)