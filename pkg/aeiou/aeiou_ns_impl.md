# How It Works: The NeuroScript AEIOU `ask` Host Loop

**Audience:** Developers & Ops
**Version:** Based on `neuroscript@0.8.0` / `steps_ask_hostloop.go @ v26`

---

## 1. High-Level Overview

This document describes the Go-native implementation of the multi-turn `ask` statement in NeuroScript, which uses the AEIOU v3 envelope structure.

The key change, as you noted, is that the loop control logic is no longer mediated by a NeuroScript tool (like the v2 `tool.aeiou.magic`). Instead, the loop is **entirely managed in Go** by the `runAskHostLoop` function.

This Go function is a "host loop" that orchestrates the entire turn-by-turn conversation. It is responsible for:

1.  Calling the LLM with an AEIOU envelope.
2.  Parsing the LLM's response envelope.
3.  **Sandboxing** and executing the `ACTIONS` block from the response.
4.  **Capturing** all `emit` and `whisper` output from that block.
5.  **Inspecting** the captured `emit`s for a simple control signal (`<<<LOOP:DONE>>>`).
6.  Enforcing safety (max turns, progress guard).
7.  Preparing and sending the *next* envelope to the LLM.

The AI's role is simplified: it just needs to `emit "<<<LOOP:DONE>>>"` when it believes the task is complete. The complex, signed-token-based control from the v3 spec (`tool.aeiou.magic`) is **not** used in this implementation.

---

## 2. The Full `ask` Flow: Step-by-Step

Here is the complete lifecycle, from the `ask` statement to the final result.

### Step 1: Entry Point (`executeAsk` in `steps_ask.go`)

1.  A user's NeuroScript runs an `ask` statement: `ask "my_agent", "prompt" into $result`.
2.  The `executeAsk` function in `steps_ask.go` handles this statement.
3.  It evaluates the agent name (`"my_agent"`) and the prompt (`"prompt"`).
4.  It retrieves the `AgentModel` and its `Provider` (e.g., OpenAI).
5.  It creates an `llmconn.Connector`, which is the Go object responsible for the actual API call.
6.  It builds the **initial** `aeiou.Envelope`. This first envelope is simple:
    * `USERDATA`: Contains the user's prompt (formatted as JSON).
    * `ACTIONS`: Contains an empty `command endcommand` block.
7.  It calls `i.runAskHostLoop()` and passes it the connector and this initial envelope. This is the handoff to the Go-native loop.

### Step 2: The Core Loop (`runAskHostLoop` in `steps_ask_hostloop.go`)

This is the heart of the new system. It starts a `for` loop that runs from `turn = 1` to `maxTurns`.

**Inside a single turn:**

1.  **Set Context:** It creates a unique `sessionID` (once) and a `turnNonce` (per turn), injecting them into a `context.Context`. This context is passed to the LLM and used for executing the `ACTIONS` block, ensuring tools can access session info.

2.  **Call AI:** It calls `conn.Converse(turnCtxForLLM, turnEnvelope)`. This takes the *current* envelope, sends it to the LLM, and gets back an `AIResponse` struct.

3.  **Parse Response:** It parses the `aiResp.TextContent` using `aeiou.Parse()`. This gives it a structured `responseEnvelope` containing the AI's `USERDATA`, `SCRATCHPAD`, `OUTPUT`, and `ACTIONS` sections.

4.  **Execute `ACTIONS` (The Sandbox):** This is a critical step.
    * It **forks** the interpreter: `execInterp := i.fork()`. This creates a new, sandboxed interpreter instance for this turn.
    * It sets the turn-specific context on this fork: `execInterp.SetTurnContext(turnCtxForLLM)`.
    * It creates empty Go containers: `var actionEmits []string` and `var actionWhispers = make(map[string]lang.Value)`.
    * It calls `executeAeiouTurn(execInterp, responseEnvelope, &actionEmits, &actionWhispers)`. (More on this in Section 3).
    * This helper function executes the `ACTIONS` block and populates the `actionEmits` and `actionWhispers` containers.

5.  **Process Output (Control Logic):**
    * The Go loop now has all the `emit`s from the `ACTIONS` block in its `actionEmits` slice.
    * It iterates over this slice and checks each `emit` string.
    * If `trimmedEmit == loopDoneSignal` (`<<<LOOP:DONE>>>`), it sets a `loopDone = true` flag.
    * It filters out the `loopDoneSignal` (and any old v2 tokens) to create `cleanEmits`.
    * The joined `cleanEmits` string becomes the `outputBody`.
    * This `outputBody` is stored as the potential `finalResult`. If the AI emits, then signals DONE, the emitted text is kept as the result.

6.  **Enforce Safety Guards:**
    * **Progress Guard:** It computes an `aeiou.ComputeHostDigest` of the *full* (un-cleaned) `actionEmits` and `actionWhispers`. It feeds this digest to `progressTracker.CheckAndRecord()`. If this digest repeats 3 times (`progressGuardDefaultN`), the loop `break`s.
    * **Termination:** The loop `break`s if:
        * `loopDone == true` (AI signaled completion).
        * `turn == maxTurns` (Turn limit reached).
        * The progress guard was triggered.

7.  **Prepare for Next Turn:**
    * If the loop did *not* `break`, it constructs a *new* `turnEnvelope` for the next iteration.
    * `UserData`: `outputBody` (the clean emits from this turn).
    * `Scratchpad`: `scratchpadBody` (the whispers from this turn).
    * `Output`: `outputBody` (mirrors `UserData` for the AI's context).
    * The `for` loop repeats with `turn + 1`.

### Step 3: Exit and Return

1.  Once the loop `break`s (for any reason), `runAskHostLoop` returns the `finalResult` (which is the last valid `outputBody` it captured).
2.  Back in `executeAsk`, this `finalResult` (a `lang.Value`) is assigned to the `into` variable (e.g., `$result`) in the original NeuroScript.

---

## 3. The `ACTIONS` Sandbox (`executeAeiouTurn` in `steps_ask_aeiou.go`)

This helper function is key to making the Go loop work. It's *only* job is to run the AI's code and capture its I/O.

1.  It's called by `runAskHostLoop` and receives the **forked** interpreter (`execInterp`).
2.  It **overrides the host context** of this forked interpreter:
    * It creates a `turnHostContext` (a shallow copy).
    * It replaces `turnHostContext.EmitFunc` with a new function that appends the emitted string to the `*actionEmits` slice (passed in from the host loop).
    * It replaces `turnHostContext.WhisperFunc` with a new function that adds the whispered data to the `*actionWhispers` map.
3.  It parses the `env.Actions` string.
4.  It executes the parsed program: `i.Execute(program)`.
5.  Because the `EmitFunc` and `WhisperFunc` were overridden, no output "escapes" to the user. It is all captured in the Go slice/map owned by `runAskHostLoop`.
6.  The function then returns, and the forked interpreter (`execInterp`) is discarded.

@   This is a clean sandbox: the `ACTIONS` block runs, its I/O is captured, and it's thrown away. The main interpreter's state is untouched, and the host loop in Go has perfect control over the I/O.

---

## 4. Summary of Loop Control

Loop termination is handled **entirely in Go** by `runAskHostLoop` and is triggered by one of three conditions:

1.  **AI Signal:** The `ACTIONS` block `emit`s the exact string `<<<LOOP:DONE>>>`.
2.  **Max Turns:** The loop hits the `maxTurns` limit defined in the `AgentModel` (and capped at `maxTurnsCap = 25`).
3.  **No Progress:** The `ProgressTracker` sees 3 (`progressGuardDefaultN`) consecutive turns produce the exact same `(output + scratchpad)` digest.

---

## 5. Key Files at a Glance

* `steps_ask.go`: **The Entry Point.** Handles the `ask` statement and kicks off the Go loop.
* `steps_ask_hostloop.go`: **The Controller.** This *is* the new "entirely golang" loop. It manages turns, calls the AI, enforces safety, and decides when to stop.
* `steps_ask_aeiou.go`: **The Sandbox.** A helper that runs the `ACTIONS` block in a fork and *captures* its `emit`/`whisper` output for the Go loop to inspect.
* `steps_ask_test.go`: **The Proof.** This test confirms the new behavior, showing a mock provider that just `emit`s `<<<LOOP:DONE>>>` to terminate the loop, explicitly noting the `tool.aeiou.magic` is obsolete.