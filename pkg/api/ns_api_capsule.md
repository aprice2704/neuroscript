# API Guide Addendum: Capsule Management

**Audience:** Host application developers (e.g., FDM team) who need to load custom capsules or build a capsule service.

This document explains how to use the `pkg/api` functions for parsing, validating, and managing collections of capsules. This API allows a host application to create its own sets of capsules and layer them with the built-in default capsules.

---

## 1. Core Concepts: Registry vs. Store

The API provides two types for managing capsules:

* **`api.CapsuleRegistry`**: A single, thread-safe collection of capsules, indexed by name and version. This is the base unit of storage.
* **`api.CapsuleStore`**: A read-only, layered collection of one or more `CapsuleRegistry` instances. When you query a `Store`, it searches each `Registry` in order until it finds a match.

This allows you to create a `CapsuleStore` that layers your own custom capsules on top of the default, built-in ones.

---

## 2. Host-Side Loading: Parsing & Validating a Capsule

The `api` package provides a single function to handle all parsing and validation for a raw capsule file.

### `api.ParseCapsule(content []byte) (*api.Capsule, error)`

This is the primary function for loading a capsule. You read your capsule file from disk (or any source) into a `[]byte`, and this function does the rest.

It automatically performs all required validation:
1.  Parses the metadata and separates it from the content body.
2.  Checks for the required `::id`, `::version`, and `::description` fields.
3.  Validates that `::id` starts with `capsule/` and contains only `[a-z0-9_-]`.
4.  Validates that `::version` is a whole integer.
5.  Calculates the `SHA256` and `Size` from the content.
6.  Generates the final `ID` (e.g., `capsule/my-fdm-tool@1`).

If any check fails, it returns a descriptive error.

**Workflow:**

1.  **Read your capsule file:**
    ```go
    rawContent, err := os.ReadFile("path/to/my/capsule.md")
    if err != nil {
        // handle error
    }
    ```

2.  **Parse and validate it:**
    ```go
    myCap, err := api.ParseCapsule(rawContent)
    if err != nil {
        // handle validation error
        // e.g., "invalid capsule '::version' 'v1.0': must be an integer"
    }
    ```

3.  **Add the valid capsule to your custom registry:**
    ```go
    myCustomRegistry := api.NewAdminCapsuleRegistry()
    
    // Use MustRegister for setup code where an error is fatal
    myCustomRegistry.MustRegister(*myCap)

    // Or use Register for runtime loading
    err = myCustomRegistry.Register(*myCap)
    ```

---

## 3. Creating and Using a Layered Store

Once you have your custom registry populated, you can create a `CapsuleStore` that layers it with the built-in capsules.

### `api.DefaultCapsuleRegistry() *api.CapsuleRegistry`

This function returns the read-only, singleton registry containing all the standard capsules embedded in the NeuroScript binary (like `capsule/bootstrap_agentic`).

### `api.NewCapsuleStore(initial ...*api.CapsuleRegistry) *api.CapsuleStore`

This creates a new store. Registries are searched in the order they are provided. To make your custom capsules override the default ones, place your registry first.

```go
// Get the default built-in capsules
defaultReg := api.DefaultCapsuleRegistry()

// Create a store where myCustomRegistry is searched first.
myStore := api.NewCapsuleStore(myCustomRegistry, defaultReg)
```

---

## 4. Querying a Store or Registry

Both `api.CapsuleStore` and `api.CapsuleRegistry` provide the same read methods:

* **`Get(name string, version string) (api.Capsule, bool)`**:
    Retrieves a capsule by its exact name and version.
    ```go
    c, ok := myStore.Get("capsule/my-fdm-tool", "1")
    ```

* **`GetLatest(name string) (api.Capsule, bool)`**:
    Retrieves the highest *integer* version of a capsule by its name. When used on a `Store`, it finds the latest version *only in the first registry* that contains that name.
    ```go
    c, ok := myStore.GetLatest("capsule/bootstrap_agentic")
    ```

* **`List() []api.Capsule`**:
    Returns a list of all unique capsules from all registries, sorted by `Priority` then `ID`.
    ```go
    allCapsules := myStore.List()
    ```

---

## 5. Integrating the Store

You have two primary ways to make your `CapsuleStore` available to the interpreter:

1.  **Option 1: (Hook) As a `CapsuleProvider` Backend**
    This is the recommended approach for a full service like FDM.
    * You implement the `api.CapsuleProvider` interface.
    * Your implementation's methods (`GetLatest`, `Read`, etc.) query *your* `api.CapsuleStore`.
    * You inject your provider into the interpreter:
        `interp := api.New(api.WithCapsuleProvider(myProvider))`
    * Now, all `tool.capsule.*` calls from within NeuroScript are routed directly to your Go service.

2.  **Option 2: (Inject) As the `CapsuleRegistry`**
    This option replaces the interpreter's default capsule source with your store.
    * You inject your `api.CapsuleStore` (which also satisfies the `api.CapsuleRegistry` interface) at creation:
        `interp := api.New(api.WithCapsuleRegistry(myStore))`
    * Now, the *built-in* `tool.capsule.*` tools will query your store instead of the default registry.