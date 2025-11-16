# Capsule System: How It Works
**ID:** `capsule/system-overview`
**Version:** 1
**Description:** High-level overview of the Capsule Registry, Store, and Loader for developers.

---

## 1. Core Concepts

The capsule system provides a layered, read-only key-value store designed to ship documentation, specs, and simple scripts *with* the NeuroScript binary.

It is composed of three main components:

1.  **Capsule:** The data object.
2.  **Registry:** A collection of Capsules.
3.  **Store:** A *layered list* of Registries.

---

## 2. The `Capsule`

A `Capsule` is a simple struct holding content and metadata.

* **`Name`**: The stable logical name (e.g., `capsule/help-topics`). Must match the format `capsule/[a-z0-9_-]+`.
* **`Version`**: An **integer** version number (e.g., `1`, `2`).
* **`ID`**: The fully qualified ID, generated automatically (e.g., `capsule/help-topics@2`).
* **`Description`**: A mandatory one-line description.
* **`Content`**: The string payload (e.g., Markdown text).
* **`MIME`**: The content type (e.g., `text/markdown`).

---

## 3. The `Registry`

A `Registry` is a thread-safe collection of capsules.

* It's a simple map: `name` -> `version` -> `Capsule`.
* You can `Register` new capsules, `Get` a specific version, or `GetLatest` (which finds the highest integer version).

---

## 4. The `Store` (Layering)

The `Store` is the most important concept for application use.

* A `Store` manages a layered set (a slice) of `*Registry`s.
* When you `Add(r *Registry)` a new registry, it is added to the *top* of the list.
* When you `Get(...)` or `GetLatest(...)`, the `Store` searches registries in order, from newest to oldest (top to bottom).
* **This allows for overriding:** If you add a new registry with `capsule/help@5`, it will be returned by `GetLatest("capsule/help")` even if a built-in registry has `capsule/help@10`. The search stops at the first registry layer that contains the *name*.

---

## 5. How It All Loads (The Singletons)

This is how the "split-brain" problem was solved.

1.  **`BuiltInRegistry()`**: A single, private `*Registry` is created. This is where built-in content lives.

2.  **`loader.go`**: At application startup (`init()`), this file walks the embedded `//go:embed all:content` directory. It parses every file, reads its metadata, and `Register`s it as a `Capsule` in the `BuiltInRegistry()`.

3.  **`DefaultStore()`**: A single, **public** `*Store` is created. It is initialized with the `BuiltInRegistry()` as its first (and base) layer.

### Summary of Use:

* **To load built-in content:** Your application code should **always** use `capsule.DefaultStore()`.
    ```go
    // Gets the built-in "capsule/aeiou" capsule
    c, ok := capsule.DefaultStore().GetLatest("capsule/aeiou")
    ```

* **To add new, queryable content at runtime:**
    ```go
    // 1. Create a new registry for your new content
    myRuntimeRegistry := capsule.NewRegistry()
    myRuntimeRegistry.MustRegister(capsule.Capsule{
        Name:        "capsule/user-script-foo",
        Version:     "1",
        Description: "A script added by the user.",
        Content:     `emit "hello"`,
    })

    // 2. Add it as a new layer to the DefaultStore
    capsule.DefaultStore().Add(myRuntimeRegistry)

    // 3. Now it's available and overrides built-ins
    c, ok := capsule.DefaultStore().GetLatest("capsule/user-script-foo")
    ```