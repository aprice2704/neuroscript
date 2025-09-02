# AEIOU v3 Host Integration Guide

This guide provides a high-level overview and pseudocode for integrating the `aeiou` package into a host interpreter's main loop. It assumes the host is responsible for managing state on a per-session (SID) basis.

---

## 1. Per-Session State Management

For each active session (identified by a unique `SessionID`), your host must maintain the following state across turns:

- **`TurnIndex`**: An integer that starts at 0 or 1 and increments each turn.
- **`LoopController`**: An instance of `*aeiou.LoopController`. Created once per session.
- **`ReplayCache`**: An instance of `*aeiou.ReplayCache`. Created once per session to detect replayed tokens.
- **`ProgressTracker`**: An instance of `*aeiou.ProgressTracker`. Created once per session to detect loops.

---

## 2. The Turn Lifecycle (Pseudocode)

The core logic for handling a single turn is as follows. This function would be called by your main `ask` handler.

```go
// Host-level function to process one turn for a given session.
func (h *Host) processTurn(sid string, turnIndex int, outputFromInterpreter, scratchpadFromInterpreter string) (decision *aeiou.Decision, nextTurnIndex int, shouldHalt bool) {

    // 1. Get the session's stateful components.
    session := h.getSession(sid) // Your session management logic.
    lc := session.loopController
    replayCache := session.replayCache
    progressTracker := session.progressTracker

    // 2. Create the context for this specific turn.
    // A new nonce MUST be generated for every turn.
    hostCtx := aeiou.HostContext{
        SessionID: sid,
        TurnIndex: turnIndex,
        TurnNonce: generateNewNonce(), // e.g., 128-bit random string
        KeyID:     h.config.CurrentSigningKeyID,
        TTL:       h.config.TokenTTL,
    }

    // 3. Process the interpreter's output to find a decision.
    decision, err := lc.ProcessOutput(outputFromInterpreter, hostCtx, replayCache)
    if err != nil {
        // Log fatal error and HALT
        log.Errorf("SID %s, Turn %d: critical error in ProcessOutput: %v", sid, turnIndex, err)
        return nil, 0, true
    }

    // 4. If no valid token was found, the decision is HALT.
    if decision == nil {
        log.Warnf("SID %s, Turn %d: HALT - No valid control token found.", sid, turnIndex)
        return nil, 0, true
    }

    // 5. Check for progress guard violation.
    // This MUST be done *after* a valid token is found.
    digest := aeiou.ComputeHostDigest(outputFromInterpreter, scratchpadFromInterpreter)
    if progressTracker.CheckAndRecord(digest) {
        log.Warnf("SID %s, Turn %d: HALT - No progress detected for %d turns.", sid, turnIndex, progressTracker.maxRepeats)
        return nil, 0, true
    }

    // 6. The turn is successful. Return the decision and prepare for the next state.
    if decision.Action == aeiou.ActionContinue {
        return decision, turnIndex + 1, false
    }

    // For DONE or ABORT, the loop terminates.
    return decision, turnIndex, true
}
```

---

## 3. Key `aeiou` Package Components Reference

- **`aeiou.Parse(envelopeReader)`**: Use this first to parse the raw envelope string into the `Envelope` struct.
- **`aeiou.NewLoopController(verifier)`**: Creates the controller. The `verifier` needs a `KeyProvider` that knows your public keys.
- **`lc.ProcessOutput(...)`**: The main function you call each turn to get a `Decision`.
- **`aeiou.ComputeHostDigest(output, scratchpad)`**: Creates the string digest for the progress guard.
- **`aeiou.NewProgressTracker(maxRepeats)`**: Creates the tracker. Use `0` for the default repeat count (3).
- **`tracker.CheckAndRecord(digest)`**: Call this once per turn. It returns `true` if the session should halt.
