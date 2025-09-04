# API Team Directive: Implementing Stateful Interpreter Sessions
**Date:** 2025-Sep-01
**To:** API Team
**From:** Interpreter Team
**Subject:** Critical fix for state persistence bug between `RunProcedure` calls.

---

## 1. The Problem: State Loss Between API Calls

We have confirmed a critical bug where state changes from one `api.RunProcedure` call do not persist to the next. For example, a script that calls `tool.account.Register` in a `setup()` procedure cannot be used by a subsequent `main()` procedure because the registered account has vanished.

This prevents the use of setup scripts and any multi-step operations that rely on durable state, rendering the current API unusable for its primary purpose.

---

## 2. Root Cause: Incorrect Session Management

The root cause of this issue is in the `api` package's execution harness. The current implementation incorrectly treats each `RunProcedure` call as a temporary, isolated event. It likely creates a new, sandboxed `interpreter.Clone()` for every call and discards it afterward, along with all its state changes.

This behavior is incorrect. The `api.Interpreter` object that users create and pass to your functions **must be treated as a persistent session object**.

---

## 3. The Solution: Persistent Interpreter Sessions

The fix is to modify the `api.RunProcedure` function (and any related execution functions) to operate **directly on the provided interpreter instance**. Do not create a temporary clone for the execution.

The interpreter is responsible for managing its own internal execution stack and variable scope. The API layer's only responsibility is to invoke the procedure on the persistent session object it was given.

### Conceptual Implementation

Your current logic is likely similar to this:

**INCORRECT (Current Logic):**
```go
// file: pkg/api/execution.go

func RunProcedure(ctx context.Context, interp *Interpreter, procName string, args ...lang.Value) (lang.Value, error) {
    // BUG: Creates a new, sandboxed clone that gets discarded, losing all state.
    runClone := interp.Clone() 
    return runClone.RunProcedure(procName, args...)
}
```

Please modify it to follow this pattern:

**CORRECT (New Logic):**
```go
// file: pkg/api/execution.go

func RunProcedure(ctx context.Context, interp *Interpreter, procName string, args ...lang.Value) (lang.Value, error) {
    // CORRECT: Executes directly on the provided interpreter instance,
    // treating it as a persistent session.
    return interp.RunProcedure(procName, args...)
}
```

---

## 4. Coordination with Interpreter Team

To fix this bug, we previously implemented a workaround in the `interpreter` package that forced clones to share state. **Once you have fixed the `api` package, this workaround must be reverted.**

Please ask the `interpreter` team to replace the contents of `pkg/interpreter/interpreter_clone.go` with the following code. This restores the intuitive sandboxing behavior of `Clone()` and is critical for the long-term health of the system.

**Code for `pkg/interpreter/interpreter_clone.go` (to be applied AFTER API fix):**
```go
// NeuroScript Version: 0.7.0
// File version: 5
// Purpose: Restores intuitive cloning behavior, creating a sandboxed state for isolated execution contexts.
// filename: pkg/interpreter/interpreter_clone.go
// nlines: 70
// risk_rating: HIGH

package interpreter

// Clone creates a new interpreter instance that is a child of the current one.
// It creates a new, isolated execution state but shares the parent's core,
// persistent stores (accounts, models) by reference.
func (i *Interpreter) Clone() *Interpreter {
	clone := &Interpreter{
		// Share core components by reference.
		logger:              i.logger,
		fileAPI:             i.fileAPI,
		tools:               i.tools,
		eventManager:        i.eventManager,
		aiWorker:            i.aiWorker,
		stdout:              i.stdout,
		stdin:               i.stdin,
		stderr:              i.stderr,
		maxLoopIterations:   i.maxLoopIterations,
		bufferManager:       i.bufferManager,
		objectCache:         i.objectCache,
		llmclient:           i.llmclient,
		skipStdTools:        i.skipStdTools,
		modelStore:          i.modelStore,
		accountStore:        i.accountStore,
		ExecPolicy:          i.ExecPolicy,
		root:                i, // Link back to the parent
		aiTranscript:        i.aiTranscript,
		transientPrivateKey: i.transientPrivateKey,
	}

	// Create a new, isolated execution state.
	clone.state = newInterpreterState()
	clone.state.providers = i.state.providers

	// Find the ultimate root to copy global variables from.
	root := i
	for root.root != nil {
		root = root.root
	}

	// Copy global variables from the root interpreter.
	root.state.variablesMu.RLock()
	for name, value := range root.state.variables {
		if _, isGlobal := root.state.globalVarNames[name]; isGlobal {
			clone.state.variables[name] = value
			clone.state.globalVarNames[name] = true
		}
	}
	root.state.variablesMu.RUnlock()

	// Set up evaluation context for the clone.
	clone.evaluate = &evaluation{i: clone}
	clone.customWhisperFunc = clone.defaultWhisperFunc

	return clone
}
```

Thank you for your prompt attention to this critical issue.

-------------------

Okay, I have read and understood the directive from the interpreter team in `stateful_inter.md` and the core rules in `AGENTS.md`.

Here is my opinion and plan:

### Opinion

The directive is clear and addresses a critical bug. The root cause is correctly identified: the API layer is likely creating ephemeral clones for execution, which discards state changes immediately. This negates the core purpose of having a persistent interpreter object for multi-step operations.

The proposed solution is the correct one: all execution functions in the `api` package must operate directly on the persistent, internal interpreter session provided to them, not on a temporary clone. This will ensure that state modifications (like account registration, variable setting, etc.) are durable across calls.

### Plan

1.  I will modify the `api.Interpreter.Run` method in `pkg/api/interpreter.go` to ensure it directly calls the `Run` method on its internal, persistent `interpreter` instance. This is the central fix required.
2.  I will verify that `api.RunProcedure` in `pkg/api/exec.go` correctly uses this persistent `Run` method, ensuring the fix propagates.
3.  I will not touch any other files, as the directive is specific to the execution flow of the interpreter instance, which is handled by these two files. This adheres to the "minimal, targeted edits" rule from `AGENTS.md`.

I have fully considered the request and am ready to proceed. I will await your explicit instruction to generate the updated files.

-------------

AJP: This does leave the host program (calling api) the options to discard interpreters for code isolation and thus safety though?


gem2.5:
Show thinking
Yes, absolutely. That is the correct and intended way to manage isolation.

This change ensures that a single Interpreter instance acts as a persistent session, so state is maintained between calls to that specific instance.

However, the host application (your Go program) retains full control over the lifecycle of these instances. You can, and should, continue to manage safety and isolation like this:

For Persistent State: Create one api.New() interpreter and reuse that same instance for a series of related operations that need to share state.

For Isolation: Create a fresh api.New() interpreter for any task that needs to run in a clean, sandboxed environment, and simply let it be garbage collected when you're done.

The two patterns are not mutually exclusive; this fix simply makes the "persistent state" pattern work as expected.