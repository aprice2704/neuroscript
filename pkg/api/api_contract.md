<!--
 NS/FDM API CONTRACT — v0.6 (2025‑07‑16)
 This file is *normative* and MUST be kept in‑sync with
 both the public `api` package and the Integration Guide.
 Any signature drift requires a simultaneous version bump.
-->

# 📜 “Integration‑readiness” contract for the **`api` ↔ `lang/*`** pipeline

The goal of this document is to guarantee that a consumer needs *only*  
`import "yourrepo/api"` to load, verify, and execute a NeuroScript unit.  
Everything below is **stable** once released; breaking changes demand a
major‑version bump of the overall module.

---

## 0 Scope & philosophy

* Public surface first: types and helpers that integrators touch.  
* No import cycles: `lang/*` never imports `api`; `api` sugar‑wraps `lang`.  
* Determinism: equal source trees → identical canonical bytes → identical hash.  
* Safety‑first: only verified + vetted trees reach the interpreter.

---

## 1 Package map

| Package            | Purpose                                       | Notes |
|--------------------|-----------------------------------------------|-------|
| **`api`**          | Single public façade for outsiders            | Re‑exports key types |
| **`pkg/types`**    | Canonical AST & small enums                   | No interpreter logic |
| **`pkg/canon`**    | Canonicalisation + signing helpers            | Pure functions |
| **`pkg/loader`**   | Verification, vetting, and caching            | No code‑gen allowed |
| **`pkg/interp`**   | Runtime interpreter                           | Internal; optional for integrators |

---

## 2 Stable AST contract (`pkg/types`)

```go
package types

type Position struct{ Line, Col int }          // 1‑based

type Kind uint8                                // append‑only enum

type Node interface {
   Pos() Position
   End() Position
   Kind() Kind
}

type Tree struct {
   Root     Node        // *always* non‑nil
   Comments []Comment   // may be empty; retained verbatim
}

// Selected nodes that tooling relies on ------------------------------
type CommandBlock struct{ /* ... */ }        // unnamed `command` → this node

type SecretRef struct {
   Path string // FDM URI to encrypted payload
   Enc  string // "age", "pgp", "none"
   Raw  []byte // nil until prepare‑stage injects
}
```

*`api` re‑exports `Position`, `Kind`, `Tree`, and `Node` so external
packages can stay import‑cycle‑free.*

---

## 3 Canonicalisation & signing (`pkg/canon`)

```go
// Canonicalise serialises a validated *types.Tree into deterministic bytes.
// Returns the canonical blob, its BLAKE2b‑256 digest, and a possible error
// (e.g. unsupported node, hash mismatch in subtree, etc.).
func Canonicalise(t *types.Tree) (
   blob []byte,
   hash [32]byte,
   err  error,
)
```

*Breaking change from v0.5:* an explicit `error` result was added.

Signing helpers (`Sign`, `Verify`, dummy Ed25519 test key) remain unchanged.

---

## 4 Loader & vetting (`pkg/loader`)

```go
// LoadVerifiedTree performs:
//  1. Decode transport wrapper  (base64 / JSON / TLV)
//  2. Verify signature & digest
//  3. Run registered analysis passes
//  4. Return a *types.Tree ready for execution
func LoadVerifiedTree(ctx context.Context, payload []byte) (*types.Tree, error)
```

**MANDATE:** *Never re‑canonicalise a tree after signature verification.*  
Keep and pass along the original `blob` + `hash` produced by step 3.

### Analysis‑pass registry

```go
package analysis

type Pass func(*types.Tree) error

func Register(name string, p Pass)   // may panic on duplicate
func RunAll(t *types.Tree) error     // called by loader
```

Error‑code block **99901‑99909** is reserved for loader/analysis fatal errors.

---

## 5 Execution entry points (`api/exec.go`)

```go
// Quick one‑shot: parse, verify, exec in a fresh interpreter.
func ExecInNewInterpreter(
   ctx  context.Context,
   src  string,             // raw NeuroScript source
   opts ...Option,          // e.g. WithStdout(io.Writer)
) (result Value, err error)

// Advanced: reuse an interpreter instance across many trees.
func ExecWithInterpreter(
   ctx   context.Context,
   interp *Interpreter,
   tree  *types.Tree,
   opts  ...Option,
) (result Value, err error)
```

The legacy `ExecCommand(ctx, tree, cfg)` shim is **deprecated** and slated
for removal in v0.7. Keep only if a large integrator still compiles against it.

---

## 6 Secret decoding stub (`pkg/secret`)

```go
// Decode decrypts t.SecretRef nodes in‑place.
// The integrator supplies the key‑fetch callback.
func Decode(t *types.Tree, keyring func(path string) ([]byte, error)) error
```

---

## 7 Dependency rules

1. `lang/*` → **must not** import `api` (prevents cycles).  
2. `pkg/interp` → **must not** import `pkg/canon` or `pkg/loader`.  
3. Only `cmd/*` and `api/*` may depend on every internal package.

A static analyser (`tools/depcheck`) enforces the graph during CI.

---

## 8 Smoke‑test (compiler gate)

```go
func TestPublicSurfaceCompiles(t *testing.T) {
   src := `command { print("hello") }`
   if _, err := api.ExecInNewInterpreter(context.Background(), src); err != nil {
       t.Fatalf("failed round‑trip: %v", err)
   }
}
```

Add this test to every module that vendors or forks the API.

---

## 9 Revision history

| Version | Date         | Notes |
|---------|--------------|-------|
| v0.4    | 2025‑05‑11   | First public draft |
| v0.5    | 2025‑06‑02   | Renamed packages; added loader rules |
| **v0.6**| 2025‑07‑16   | Added `error` to `Canonicalise`; replaced `ExecCommand` with two helpers; moved AST home to `pkg/types`; documented analysis pass registry |

---

**End of file**
