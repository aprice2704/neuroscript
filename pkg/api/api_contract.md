# “Integration-readiness” checklist for the **`ns/api` ↔ `lang/*`** pipeline

### 1 AST contract (single source of truth)

| Must-have                                                                 | Notes                                                                                      |
| ------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------ |
| **`type Position struct{ Line, Col int }`**                               | 1-based, immutable. `String()` → `"file.ns:12:7"` (needed by `api.FormatWithRemediation`). |
| **`type Kind uint8`** stable enum                                         | Do **not** reorder once published. Add new kinds at the end only.                          |
| **`type Node interface { Pos() Position; End() Position; Kind() Kind }`** | Every concrete node lives in `lang/ast/*`.                                                 |
| **`type Tree struct { Root Node; Comments []Comment }`**                  | Comments captured as their own nodes **or** in the slice; round-trippable.                 |
| **Unnamed `command` block** maps to **`*ast.CommandBlock`**               | One per file; detector logic relies on `KindCommandBlock`.                                 |
| **`*ast.SecretRef` node**                                                 | Fields: `Path string`, `Enc string`, `Raw []byte` (may be nil pre-prepare).                |

> *Public location:* `lang/ast/ast.go`.
> *The `api` package **aliases** these types; keep packages import-cycle-free.*

---

### 2 Parser guarantees (`lang/parser`)

* `Parse(src []byte, preserveComments bool) (*ast.Tree, error)`

  * `preserveComments=false` may drop comments for speed.
* On success the tree **obeys** invariants:

  * Single `CommandBlock` **xor** ≥0 `FuncDecl` **xor** ≥0 `EventHandler`.
  * All child pointers non-nil, positions increasing.
* Returns `ErrSyntax` on first violation; caller wraps into `RuntimeError`.

### 3 Canonicaliser (`lang/canon`)

| Function                                           | Contract                                                                |
| -------------------------------------------------- | ----------------------------------------------------------------------- |
| `Canonicalise(tree *ast.Tree) ([]byte, [32]byte)`  | Deterministic varint encoding; same input → same bytes on any platform. |
| `Decode(blob []byte) (*ast.Tree, [32]byte, error)` | Shape validation only (no signature check).                             |

*Hash = **blake2b\_256** of canonical bytes. 32 bytes in `[32]byte`.*

### 4 Signature helpers (`lang/sign`)

```go
type SignedAST struct{ Blob []byte; Sum [32]byte; Sig []byte }

Sign(priv ed25519.PrivateKey, blob []byte, sum [32]byte) (*SignedAST, error)
Verify(pub ed25519.PublicKey, s *SignedAST) (*ast.Tree, error)
```

*`Verify` must re-canonicalise → compare `Sum` → verify `Sig`.*

### 5 Interpreter shim (`lang/interp`)

```go
func ExecCommand(ctx context.Context, tree *ast.Tree,
                 cfg interp.Config) (*api.ExecResult, error)
```

\*Assumes tree has been vetted & is `RunModeCommand`.
*`interp.Config` includes `SecretResolver func(ref *ast.SecretRef) (string, error)`.*

### 6 Static-analysis pass hooks (`lang/analysis`)

Expose registry:

```go
type Pass interface{ Name() string; Analyse(*ast.Tree) []api.Diag }
func RegisterPass(p Pass)
```

Built-ins already drafted (shape, typecheck, capability, secret, set-order).

### 7 Error codes (`lang/errors`)

* Ensure the **99901-99909** block exactly matches the catalogue given last (add `ErrorCodeSecretDecryption`).
* `FormatWithRemediation` remains in `api`, but needs `errors.Lookup` underneath.

### 8 Secrets decoding stub (`lang/secret/decoder.go`)

Provide:

```go
func Decode(ref *ast.SecretRef, priv []byte) (string, error) // enc = "none"|"age"|"sealedbox"
```

Return `ErrSecretUnsupported` if `Enc` unknown; interpreter lifts to 99909.

### 9 Package hygiene

* `lang/*` **must not** import `api` (avoid cycles).
* `api/reexport.go` should say:

```go
type Position = ast.Position
type Kind     = ast.Kind
type Node     = ast.Node
type Tree     = ast.Tree
```

*That way external consumers do `import "yourrepo/api"` only.*

### 10 Smoke test to keep green

```
go test ./api -run TestEndToEnd
```

Flow:

1. Read `testdata/template.ns`
2. Parser → Tree
3. Canonicalise → Sign (dummy key)
4. Load → Vet passes
5. Exec → get `"hello world"` output

---

### TL;DR for Gemini

1. **Stabilise AST structs & kinds** — parser, canoniser, interpreter all speak that.
2. **Wire canonicaliser + signer** — deterministic bytes, blake2b, Ed25519.
3. **Expose ExecCommand(tree)** — run only verified, vetted command trees.
4. **Keep comments & Position in AST** — for `api.Format` and diagnostics.

Once those surfaces compile, `api` can lock onto them and external integrators need nothing beyond `import "yourrepo/api"`.
