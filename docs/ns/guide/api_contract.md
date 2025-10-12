
# 📜 “Integration-readiness” contract for the `api` ↔ `lang/*` pipeline

The goal of this document is to guarantee that a consumer needs *only* `import "yourrepo/api"` to load, verify, and execute a NeuroScript unit.
Everything below is **stable** once released; breaking changes demand a
major-version bump of the overall module.

---

## 0 Scope & philosophy

* Public surface first: types and helpers that integrators touch.
* No import cycles: `lang/*` never imports `api`; `api` sugar-wraps `lang`.
* Determinism: equal source trees → identical canonical bytes → identical hash.
* Safety-first: only verified + vetted trees reach the interpreter.

---

## 1 Package map

| Package            | Purpose                                       | Notes |
|--------------------|-----------------------------------------------|-------|
| **`api`** | Single public façade for outsiders            | Re-exports key types |
| **`pkg/types`** | Canonical AST & small enums                   | No interpreter logic |
| **`pkg/canon`** | Canonicalisation + signing helpers            | Pure functions |
| **`pkg/loader`** | Verification, vetting, and caching            | No code-gen allowed |
| **`pkg/interp`** | Runtime interpreter                           | Internal; optional for integrators |
| **`pkg/testutil`** | Shared helpers for test files only            | Not for use in production code |

---

## 2 Stable AST contract (`pkg/types`)

The core AST structures remain unchanged. `api` re-exports `Position`, `Kind`, `Tree`, and `Node`.

---

## 3 Canonicalisation & signing (`pkg/canon`)

```go
// Canonicalise serialises a validated *types.Tree into deterministic bytes.
// Returns the canonical blob, its BLAKE2b-256 digest, and a possible error.
func Canonicalise(t *types.Tree) (
   blob []byte,
   hash [32]byte,
   err  error,
)
```

---

## 4 Loader & vetting (`pkg/loader`)

```go
// Load performs signature verification and analysis on a signed AST.
// On success, it returns a LoadedUnit containing the verified tree.
func Load(ctx context.Context, s *SignedAST, cfg LoaderConfig, pubKey ed25519.PublicKey) (*LoadedUnit, error)
```

**MANDATE:** *Never re-canonicalise a tree after signature verification.*

---

## 5 Execution entry points (`api/exec.go`, `api/interpreter.go`)

The primary way to execute code is by creating an interpreter instance and passing it the code to run.

```go
// Create a new interpreter, providing mandatory options like the sandbox directory.
interp := api.New(api.WithSandboxDir("/path/to/safe/dir"))

// Load a verified unit of code into the interpreter.
err := api.LoadFromUnit(interp, loadedUnit)

// Run a specific procedure from the loaded code.
result, err := api.RunProcedure(ctx, interp, "myProc", arg1, arg2)
```

### Key Configuration Options

| Option Signature                            | Purpose |
|---------------------------------------------|---------|
| `api.WithSandboxDir(path string)`           | **Mandatory for file I/O.** Sets the secure root directory for file operations. |
| `api.WithLogger(logger api.Logger)`         | Provides a custom logger. |
| `api.WithStdout(w io.Writer)`               | Sets the standard output stream. |

---

## 6 Critical Error Handling (`api/reexport.go`)

The API provides a mechanism to override the default `panic` behavior for critical internal errors.

```go
// RegisterCriticalErrorHandler allows the host application to set a custom handler.
func RegisterCriticalErrorHandler(h func(*lang.RuntimeError))
```

---

## 7 Dependency rules

1. `lang/*` → **must not** import `api`.
2. `pkg/interp` → **must not** import `pkg/canon` or `pkg/loader`.
3. `pkg/testutil` → **must not** be imported by non-test (`_test.go`) files.
4. Only `cmd/*` and `api/*` may depend on every internal package.

---

## 8 Revision history

| Version | Date       | Notes |
|---------|------------|-------|
| v0.6    | 2025-07-16 | Added `error` to `Canonicalise`; added FDM team API functions. |
| **v0.7**| 2025-07-23 | **Removed global `api.Init`**. Sandbox is now configured per-interpreter via `api.New(api.WithSandboxDir(...))`. Added `testutil` package. |

---

**End of file**