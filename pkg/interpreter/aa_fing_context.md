# NeuroScript Interpreter: Context, Identity & Sandboxing
aka -- if you fiddle with context, may God help you

Revision: 2025-Oct-15
Status: Implemented & Verified

---

## 1. Overview

This document outlines the architectural principles governing how the NeuroScript interpreter manages state, identity, and security boundaries. Correctly propagating these contexts is critical for sandboxing script execution, especially within multi-turn agentic loops (ask statements) and nested procedure calls.

The system is built on three distinct types of "context," each with a specific lifecycle and propagation mechanism. Understanding the interplay between these three is key to understanding the interpreter's behavior.

## 2. The Three Contexts of Execution

Every operation within the interpreter occurs within the combined influence of these three contexts.

### A. The Host Context (*HostContext)
- Lifecycle: Long-Lived & Immutable. Created once when the root interpreter is instantiated and shared by reference with all descendants.
- Purpose: Represents the interpreter's connection to the outside world. It holds host-provided resources and long-term identity.
- Contains:
  - I/O handles (Stdout, Stderr, Stdin).
  - A structured logger (Logger).
  - Host-defined callbacks (EmitFunc, WhisperFunc).
  - The long-term Actor identity (Actor), representing the user or system agent on whose behalf the script is running.
- Propagation: The pointer to the HostContext is copied verbatim to all forked interpreters. It is a shared, read-only resource.

### B. The Turn Context (context.Context)
- Lifecycle: Request-Scoped & Ephemeral. Created at the beginning of a specific operation (like an ask loop) and destroyed when the operation completes.
- Purpose: Carries request-scoped data that is relevant only for the duration of a single, continuous operation. It follows the standard Go context pattern.
- Contains:
  - AEIOU loop data (AeiouSessionIDKey, AeiouTurnIndexKey).
  - Cancellation signals and deadlines (standard Go context features).
- Propagation: This is the most critical and complex propagation. The turnCtx is passed down a chain of interpreters. Each new context must be created from its parent (context.WithValue(parentCtx, ...)), not from context.Background(), to preserve the chain.

### C. The Interpreter State (*interpreterState)
- Lifecycle: Sandbox-Scoped. A new interpreterState is created for every fork, isolating it from its parent.
- Purpose: Represents the "memory" of a specific execution sandbox. It prevents a called procedure from affecting the variables of its caller.
- Contains:
  - The local variable scope (variables).
  - A reference to loaded procedure definitions (knownProcedures).
  - The sandbox directory path.
- Propagation: This state is intentionally not propagated. Each fork gets a new, clean interpreterState, ensuring sandbox integrity. The only exception is that global variables from the root interpreter are copied into a new clone's state at creation.

## 3. The fork() Process: The Heart of Sandboxing

The fork() method is the primary mechanism for creating sandboxes. When a procedure is called or an ask loop executes an AI's ACTIONS block, it does so within a forked interpreter. The fork is a lightweight clone with specific rules for what is shared, what is isolated, and what gets a new "view."

* SHARED (by pointer):
    - HostContext: All forks share one connection to the host.
    - Root-level stores (modelStore, accountStore, etc.): Definitions are shared globally.
    - knownProcedures: All forks can see all loaded function definitions.

* ISOLATED (newly created):
    - id: Each fork has a unique ID for debugging.
    - interpreterState: Each fork gets its own memory for local variables.

* NEW VIEW (special case):
    - ToolRegistry: This is the most critical component. The clone does not share the parent's ToolRegistry object. Instead, it creates a new view of the registry (NewViewForInterpreter(clone)). This new view shares the underlying map of tool definitions but is explicitly bound to the clone's own runtime. This was the key fix that solved the "stale context" bug, ensuring that when a tool is called, it receives the runtime of its immediate caller (the clone), not the runtime of the parent that originally owned the registry.

## 4. The Chain of Custody: Tracing Context to a Tool

The successful propagation of context to a tool function relies on every step in the chain performing its role correctly.

1.  Initiation (ask statement): The runAskHostLoop function begins the process. It creates the initial turnCtx by adding session and turn IDs to the interpreter's existing context (context.WithValue(i.GetTurnContext(), ...)). This context is set on the interpreter managing the loop.

2.  Sandboxing (executeAeiouTurn): The ask loop calls i.fork() to create a sandbox interpreter for the AI's ACTIONS block.

3.  Inheritance (fork()): The fork() method ensures the new clone inherits the turnCtx from its parent by calling clone.setTurnContext(i.GetTurnContext()). It also creates the new ToolRegistry view bound to the clone.

4.  Tool Call (CallFromInterpreter): When the script calls a tool (e.g., tool.test.probeContext()), the CallFromInterpreter bridge method is invoked. It receives the live sandbox interpreter as its interp argument.

5.  Execution (impl.Func): CallFromInterpreter correctly calls the tool's Go function, passing the live sandbox interpreter (interp) as the tool.Runtime.

6.  Access: The tool's Go function can now safely access all contexts:
    - It can access the HostContext and its Actor via the Runtime.
    - It can access the turnCtx by type-asserting the Runtime to an interpreter.TurnContextProvider and calling GetTurnContext().

This unbroken chain ensures that no matter how many layers of sandboxing are applied, the tool always has a direct and accurate connection to its immediate execution environment and the request-scoped data it contains.