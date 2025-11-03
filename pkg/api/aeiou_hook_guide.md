# NeuroScript AEIOU Service Hook: Architecture Guide

**Version:** 1.0
**Date:** 2025-11-02
**Audience:** Developers and Operators integrating the NeuroScript interpreter.

---

## 1. Executive Summary

The `ask` statement in NeuroScript is the entry point for complex, multi-turn AI conversations. This architecture (AEIOU v2) **externalizes the `ask` loop** from the interpreter into a host-provided "Orchestrator Service" (like FDM's `AeiouService`).

This change moves the responsibility for managing the LLM connection, parsing responses, validating code, and tracking loop state from the NeuroScript interpreter to the host application (FDM).

This provides three main benefits:
1.  **Security:** The host (FDM) validates the AI-generated code *before* it ever reaches the interpreter for execution.
2.  **Control:** The host can implement custom loop logic, such as advanced progress tracking or error handling, without modifying the interpreter.
3.  **Encapsulation:** The interpreter is no longer responsible for managing stateful LLM connections, simplifying its design.

This document explains the "how" of this interaction: the **Hook** (how `ask` calls the service) and the **Callback** (how the service executes code).

---

## 2. The Hook: How `ask` Calls Your Service

The connection is established by the host application (FDM) *injecting* its service into the interpreter's `HostContext` at creation time.

### The Contract: `AeiouOrchestrator`

The host service must implement the `AeiouOrchestrator` interface. The *internal* definition (used by the interpreter) is:

```go
// Defined in: pkg/interfaces/aeiou.go

const AeiouServiceKey = "AeiouService"

type AeiouOrchestrator interface {
    RunAskLoop(
        callingInterp any, // This will be an *api.Interpreter
        agentModelName string,
        initialPrompt string,
    ) (lang.Value, error)
}
```

### The Injection: `WithServiceRegistry`

The host application (e.g., FDM's `zadeh/main.go`) creates its service and places it in a map. This map is then injected into the interpreter using the public API.

```go
// In FDM's setup code (example):

import "[github.com/aprice2704/neuroscript/pkg/api](https://github.com/aprice2704/neuroscript/pkg/api)"
import "[github.com/aprice2704/neuroscript/pkg/interfaces](https://github.com/aprice2704/neuroscript/pkg/interfaces)"

// 1. Create the service
myAeiouService := aeiou.NewService(...) 

// 2. Create the registry
serviceRegistry := map[string]any{
    interfaces.AeiouServiceKey: myAeiouService,
}

// 3. Build the HostContext
hc, _ := api.NewHostContextBuilder().
    WithLogger(...).
    WithStdout(...).
    WithServiceRegistry(serviceRegistry). // <-- Injection happens here
    Build()

// 4. Create the interpreter
interp := api.New(api.WithHostContext(hc))
```

### The Trigger: `executeAsk`

When a NeuroScript script executes `ask "my_agent", "my_prompt" into result`, the interpreter's internal `executeAsk` function is called. This function now performs the following logic:

1.  It checks `i.hostContext.ServiceRegistry`.
2.  It looks for a key matching `interfaces.AeiouServiceKey`.
3.  **If found**, it type-asserts the value to `interfaces.AeiouOrchestrator`.
4.  **On success**, it *immediately* calls `service.RunAskLoop(i.PublicAPI, "my_agent", "my_prompt")`. The service now has full control. The value returned by the service is placed into the `result` variable.
5.  **If not found (or wrong type)**, it logs a warning and gracefully **falls back** to the legacy internal `ask` loop (`executeLegacyAsk`). This ensures backward compatibility for older scripts or simple setups.

---

## 3. The Callback: How Your Service Executes Code

When the hook is called, the Orchestrator Service (FDM) is responsible for the *entire* multi-turn loop. This loop involves calling the LLM and then securely executing its `ACTIONS` response.

The service **must not** execute the `ACTIONS` string directly. It must use the NeuroScript API to parse, validate, and run the code in a sandbox.

### The Service's Responsibility (The Loop)

The `RunAskLoop` implementation in your service (e.g., FDM's `hostloop.go`) will look something like this:

```go
// In FDM's AeiouService (example):

func (s *aeiouService) RunAskLoop(
    callingInterp any,
    agentModelName string,
    initialPrompt string,
) (lang.Value, error) {

    // 'callingInterp' is the agent's *api.Interpreter
    agentInterp := callingInterp.(*api.Interpreter) 

    // 1. Setup connection using the SDK
    conn, _ := api.NewConnector(...) 
    envelope := ... // Build initial envelope

    for {
        // 2. Call LLM
        resp, _ := conn.Converse(ctx, envelope)
        actionsString := resp.Actions // Get the code from the LLM

        // 3. PARSE
        tree, err := api.Parse([]byte(actionsString), api.ParseSkipComments)
        if err != nil {
            // Handle parse error (e.g., send error to LLM)
            continue 
        }

        // 4. VALIDATE
        // CheckScriptTools uses the agent's interpreter to check permissions
        err = api.CheckScriptTools(tree, agentInterp)
        if err != nil {
            // Handle permission error (e.g., send error to LLM)
            continue
        }

        // 5. EXECUTE (The Callback)
        // This is the key: call back into the API
        emits, whispers, execErr := api.ExecuteSandboxedAST(
            agentInterp,
            tree,
            context.Background(),
        )
        
        // 6. Process results
        if execErr != nil {
            // Handle execution error
            continue
        }

        if IsLoopDone(emits) {
            return GetFinalResult(emits), nil
        }

        // 7. Loop
        envelope = BuildNextEnvelope(emits, whispers)
    }
}
```

### The Sandbox: `ExecuteSandboxedAST`

The most critical function is `api.ExecuteSandboxedAST`. When the service calls this, the function:

1.  Takes the agent's `*api.Interpreter` (passed into the hook).
2.  Calls the internal `interpreter.ForkSandboxed()` method to create a **new, sandboxed child interpreter**.
3.  This fork inherits the root interpreter's functions and tools but has its own isolated memory.
4.  It **redirects the fork's I/O** (its `EmitFunc` and `WhisperFunc`) to capture output into slices.
5.  It executes the `command` blocks from the AST using this sandboxed, I/O-redirected fork.
6.  It returns the captured `emits`, `whispers`, and any `execErr` to the service.

This ensures that the AI-generated code never runs with host-level privileges and cannot modify the agent interpreter's main state.

---

## 4. Summary: The Full Flow

1.  **Setup:** FDM Host injects `AeiouService` into the `HostContext` via `WithServiceRegistry`.
2.  **Script:** `ask "agent", "prompt" into result` is executed.
3.  **Hook:** `executeAsk` finds `AeiouService` and calls `RunAskLoop()`, passing the `*api.Interpreter`.
4.  **Service Loop:**
    a. `AeiouService` calls the LLM.
    b. `AeiouService` receives an `ACTIONS` string.
    c. `AeiouService` calls `api.Parse()` and `api.CheckScriptTools()`.
5.  **Callback:**
    a. `AeiouService` calls `api.ExecuteSandboxedAST()`.
    b. NeuroScript creates a sandboxed fork, runs the code, and captures I/O.
    c. NeuroScript returns `emits` and `whispers` to the service.
6.  **Service Loop (End):**
    a. `AS` processes `emits`, checks for `<<<LOOP:DONE>>>`.
    b. `AeiouService` either loops (back to 4a) or returns the `finalResult`.
7.  **Script (End):** The `finalResult` is assigned to the `result` variable.