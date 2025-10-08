# NeuroScript `ax` API: A Practical Guide for Integrators

**Audience:** Developers & Ops building with NeuroScript (e.g., FDM/Zadeh Team)
**Version:** 2.3 (2025-10-07)
**Purpose:** This guide explains the "how" and "why" of the `ax` (Agent eXecution) interface, focusing on the practical lifecycle of a host application. It covers the boot process, shared function libraries, identity injection, and cloning.

---

## 1. The `ax` Philosophy: Stability and Security

The `ax` package provides a stable, clean "ports and adapters" layer for NeuroScript. Its purpose is to give host applications a simple, unchanging set of Go interfaces to code against, decoupling them from the complexities and future changes of the core engine.

**Key Principles:**
- **Stability:** The `ax` interfaces will rarely change.
- **Decoupling:** `ax` is a "leaf" package with **zero dependencies** on NeuroScript internals. All the "wiring" happens in the `pkg/api` layer.
- **Security:** The API is designed around a clear separation of privileged setup and unprivileged execution.

---

## 2. The Core Components

The entire `ax` model revolves around three main concepts: the **Factory**, the **Environment**, and the **Runner**.

* **`RunnerFactory`**: The single entry point. Its job is to hold the shared state and "mint" new `Runners`.
* **`RunEnv`**: The shared, persistent environment holding all resources like accounts, agent models, and tools.
* **`Runner`**: An actual NeuroScript interpreter, which comes in two modes: privileged **`Config`** runners for setup and unprivileged **`User`** runners for sandboxed execution.

---

## 3. The Application Lifecycle

The `ax` API is designed to model a standard application lifecycle: a one-time, privileged boot phase followed by many unprivileged runtime requests.

### Phase 1: The Boot Process (Configuration)

At application startup, you run a trusted boot script using the `api.AXBootLoad` helper function. This function automatically:
1.  Creates a short-lived, privileged **`Config` Runner**.
2.  Parses your boot script, loading all `func...endfunc` definitions into the factory's shared library.
3.  Executes the script's `command...endcommand` block to populate the `RunEnv` with accounts, models, etc.

### Phase 2: Per-Request Execution

For every incoming user request, you create a fresh, sandboxed **`User` Runner** using `factory.NewRunner()`. The factory automatically copies the shared function library into this new runner. To execute a script, you use the `api.AXRunScript` helper, which handles parsing, loading, and running a specific procedure.

---

## 4. Injecting Identity and Host Context

To make host-level information (like a user's ID) available to tools, you create a custom Go struct that implements the `api.Runtime` and `ax.IdentityCap` interfaces. Your tools should not be coupled to this specific struct; instead, they should depend on the **`ax.IdentityCap` capability interface**. This keeps your tools decoupled and reusable.

---

## 5. Advanced Usage: Cloning a Runner

While each `User` runner is sandboxed, you may occasionally need to create a "snapshot" of a runner's state to perform a speculative execution without affecting the original. This is the purpose of cloning.

Cloning is an **optional capability**. Not all runners can be cloned. To safely clone a runner, you must check if it implements the `ax.CloneCap` interface.

A clone inherits the **entire state** of its parent at the moment of cloning, including all variables. However, it is a separate instance; changes made in the clone will **not** affect the original runner. This is useful for "what-if" scenarios or running a sub-task that might fail without corrupting the state of the main task.

**Example: Safely Cloning a Runner**
```go
// Assume 'userRunner' is an existing ax.Runner

if cloneable, ok := userRunner.(ax.CloneCap); ok {
    // It's safe to clone.
    speculativeRunner := cloneable.Clone()

    // You can now run operations on speculativeRunner without
    // affecting userRunner.
    _, err := speculativeRunner.Run("somePotentiallyFailingTask")
    if err != nil {
        // The original userRunner is unaffected by the failure.
    }
}
```

---

## 6. Full Code Example: The Modern `ax` Pattern

This example demonstrates the full, decoupled lifecycle using the public helper functions.

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/aprice2704/neuroscript/pkg/api"
	"github.com/aprice2704/neuroscript/pkg/ax"
)

// --- Custom Runtime that implements ax.IdentityCap ---
type AppRuntime struct {
	id ax.ID
}
func (r *AppRuntime) Identity() ax.ID { return r.id }
// ... other api.Runtime methods would go here ...

// --- Identity-Aware Tool ---
func WhoAmITool(rt api.Runtime, _ []any) (any, error) {
	// The tool asserts the CAPABILITY, not the concrete type.
	if idCap, ok := rt.(ax.IdentityCap); ok && idCap.Identity() != nil {
		return "called by: " + string(idCap.Identity().DID()), nil
	}
	return nil, errors.New("runtime is missing identity capability")
}

func main() {
	ctx := context.Background()

	// --- PHASE 1: BOOTSTRAP ---
	bootID := &api.mockID{did: "did:app:system"}
	baseRt := &AppRuntime{id: bootID}
	factory, _ := api.NewAXFactory(ctx, ax.RunnerOpts{}, baseRt, bootID)

    // example only, not valid syntax
	bootScript := `
        # This function is now in the shared library.
        func get_system_message(returns string) means
            return "System configured successfully."
        endfunc

        # This command configures the shared environment.
        command
            must tool.account.register("service-acct", {"kind":"test", "api_key":"key-123"})
        endcommand
    `
	// Use the public helper for the entire boot process.
	if err := api.AXBootLoad(ctx, factory, []byte(bootScript)); err != nil {
		panic(err)
	}
	fmt.Println("--- Boot complete. ---")

	// --- PHASE 2: PER-REQUEST EXECUTION ---
	userRunner, _ := factory.NewRunner(ctx, ax.RunnerUser, ax.RunnerOpts{})

	whoamiImpl := api.ToolImplementation{Spec: api.ToolSpec{Name: "whoami", Group: "host"}, Func: WhoAmITool}
	userRunner.Tools().Register("tool.host.whoami", whoamiImpl)

	userScript := `
        func main(returns string) means
            # This function was inherited from the boot script.
            return get_system_message()
        endfunc
    `
	// Use the public helper to run the script.
	result, _ := api.AXRunScript(ctx, userRunner, []byte(userScript), "main")
	fmt.Printf("Result from user script: %v\n", result)

	// Read from the shared environment via the pure ax.RunEnv interface.
	_, ok := factory.Env().AccountsReader().Get("service-acct")
	fmt.Printf("User runner found service account: %v\n", ok)
}
```