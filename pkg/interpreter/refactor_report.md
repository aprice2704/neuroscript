# Interpreter Refactoring Summary
**Revision:** 2025-Oct-11

## 1. Architectural Goal

The primary goal was to address systemic instability in the interpreter caused by a complex `clone()` method, unclear state management, and blurred lines between configuration and runtime. The refactoring aimed to establish clear, predictable boundaries for state, dependencies, and execution scope.

## 2. Core Changes Implemented

### a. Introduction of `HostContext`
- A new `HostContext` struct was created to hold all immutable, host-provided dependencies (Logger, I/O streams, Emit/Whisper callbacks).
- The `Interpreter` now holds a single, immutable pointer to this context.
- This completely resolves the issue of "losing" injected functions in nested scopes, as they are no longer part of the interpreter's own state.

### b. Refined Forking Model
- The problematic `clone()` method was replaced with a private `fork()` method.
- `fork()` creates a new interpreter with an isolated variable scope but directly inherits the `HostContext` pointer and the `root` interpreter reference. This is more efficient and eliminates state-copying bugs.

### c. Decoupled Evaluation and Policy
- **`eval` Package:** All expression evaluation logic was moved out of the interpreter into a new `pkg/eval` package. The evaluator operates on a `Runtime` interface, which the interpreter now implements, cleanly separating the "how" of evaluation from the "what" of the interpreter's state.
- **`policygate` Package:** All security and capability checks were centralized into a new `pkg/policygate` package. This simplifies logic within the interpreter and ensures all security-sensitive operations are vetted through a single, consistent point.

### d. Enforced Read-Only Globals
- The `SetVariable` method was modified to programmatically prevent forked (non-root) interpreters from modifying global variables. This enforces a critical architectural principle and prevents side effects from subroutines.

### e. Slimmed Public API
- The public API of the `interpreter` package was significantly reduced.
- Redundant methods (`Run`), convenience wrappers (`LoadAndRun`), and state-mutating methods (`SetStdout`, `SetEmitFunc`, etc.) were removed.
- The API now focuses on the core responsibilities: configuration (at startup), loading scripts, and running procedures.

## 3. Outcome

The interpreter's architecture is now fundamentally more robust and predictable. The clear separation between **Host Capabilities (`HostContext`)**, **Persistent State (`rootInterpreter`)**, and **Transient Scope (`fork`)** provides a solid foundation for the FDM's use cases. This surgical refactoring has addressed the identified root causes of instability without requiring a complete rewrite.