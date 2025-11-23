# NeuroScript Handle Interface Specification (v0.1)

## 1. Purpose

A **handle** is an opaque reference from NeuroScript (NS) code to a Go-side
object stored in a host-managed registry. Handles enable NS programs to work
with complex host objects (files, ASTs, graph views, etc.) without exposing
internal structure, without serializing large blobs, and without requiring
string-based marshalling.

This document defines the:
- public Go interface for creating and resolving handles,
- NS-level semantics,
- runtime guarantees,
- integration guidance for tools and host systems (e.g., FDM).


## 2. Core Concepts

A handle is:
- a small opaque NS value (internally implemented as a formatted string),
- backed by a Go pointer stored in a per-runtime registry,
- non-persistable and non-canonicalizable,
- local to one interpreter instance,
- type-tagged for safe downcasting in tools.

NS code can:
- store handles in variables,
- pass them to functions and tools,
- receive them from tools,
- treat them as opaque tokens with no introspection.

Host-side Go code can:
- create handles from objects,
- look up the underlying payload,
- check type assertions,
- implement handle-bound methods via tools.


## 3. API Surface (Go Side)

### 3.1 Handle Value

The interpreter exposes a distinct internal value kind for handles.
Externally, it is represented as `api.HandleValue`.

```
type HandleValue interface {
    api.Value                   // satisfies standard NS value interface
    HandleID() string           // returns canonical handle id string
}
```


### 3.2 Handle Registry

Each interpreter/runtime maintains a private registry mapping:
    handleID → payload(any)

A new public interface is exposed:

```
type HandleRegistry interface {
    NewHandle(payload any) (HandleValue, error)
    GetHandle(id string) (any, error)
    DeleteHandle(id string) error
}
```


### 3.3 Runtime Extensions

The `api.Runtime` gains:

```
HandleRegistry() HandleRegistry
```

This allows tools and host infrastructure to create and resolve handles from
within the running interpreter.


### 3.4 LangWrap / Unwrap Semantics

`LangWrap` MUST:
- detect if the input is a `HandleValue`,
- pass it through unchanged.

`Unwrap` MUST:
- when given a `HandleValue`, return the `HandleValue` object itself,
  NOT the underlying payload,
- never implicitly unwrap into Go objects (safety: require explicit tool code).


## 4. Handle Lifecycle

Handles have the following lifecycle:

1. **Create**  
   A Go tool or host function calls:
   `hv, _ := rt.HandleRegistry().NewHandle(payload)`

2. **Use**  
   - NS code stores/forwards the handle.
   - Tools receive handles as `Value` → cast to `HandleValue`.
   - Tools then call `payload, _ := rt.HandleRegistry().GetHandle(h.HandleID())`

3. **Delete (optional)**  
   - Tools may call `DeleteHandle` if the object should be freed early.

4. **Automatic cleanup**  
   - When an interpreter instance ends, all handles in its registry are freed.


## 5. Formatting and Safety Rules

### 5.1 Canonical String Form

Internally handles use strings such as:
    "handle:XXYYZZ"

Rule:
- NS code cannot fabricate valid handles, because registry lookups will fail.


### 5.2 Non-Persistability

Handles:
- MUST NOT appear in canonicalization,
- MUST NOT be stored in graph nodes,
- MUST be rejected when a tool attempts to write them into persistent storage.

The interpreter MUST report an error on canonicalization attempts containing
handle values.


### 5.3 Locality

Handles are valid only within the interpreter instance that created them.
Cross-runtime use is undefined and should return `ErrHandleNotFound`.


## 6. Tool Conventions (Go)

Tools that accept handles should:

```
switch hv := args[0].(type) {
case api.HandleValue:
    objAny, err := rt.HandleRegistry().GetHandle(hv.HandleID())
    if err != nil { return nil, err }
    obj := objAny.(*MyType)   // strong downcast
    // operate on obj
default:
    return nil, fmt.Errorf("expected handle")
}
```


## 7. NS Language Semantics

### 7.1 Opaque

In NS:

- A handle cannot be printed (shows as `<handle>`).
- A handle cannot be iterated, indexed, or concatenated.
- A handle can be:
    - assigned to variables,
    - passed to functions and tools,
    - returned from functions.


### 7.2 Type Checks

The NS operator `typeof x` returns `"handle"` for handle values.


### 7.3 Equality

Handles compare by ID:

```
let a = TOOL.MakeHandle(...);
let b = TOOL.MakeHandle(...);
a == b       # false

let c = a;
a == c       # true
```


## 8. Recommended Standard Handle Types

These are host-side types intended for use in FDM:

### 8.1 FileMetaHandle

Payload: Go struct with:
- path (abs & rel)
- size
- modtime
- is_dir
- mime (optional)

NS will access via tools:

- `TOOL.FS.Path(h)`
- `TOOL.FS.IsDir(h)`
- `TOOL.FS.Size(h)`


### 8.2 FileContentHandle

Methods exposed by tools:
- `ReadAll(h)`
- `ReadLines(h)`
- `ReadRange(h, off, len)`
- `Hash(h)`


### 8.3 AstHandle (Go AST or NS AST)

Payload: pointer to a Go AST root.

Tools:
- `AST.Root(h)`
- `AST.Children(h)`
- `AST.FindSymbol(h, name)`
- `AST.ToText(h)`


### 8.4 GraphQueryHandle (optional)

Payload: precomputed node set or iterator.


## 9. Minimal Implementation Requirements

A compliant interpreter must:
- support `api.HandleValue`,
- include a per-runtime handle registry,
- expose `HandleRegistry()` on runtime,
- preserve handles through NS execution,
- reject handles during program canonicalization and persistence,
- guarantee lookup isolation across runtimes.


## 10. Example Usage

### 10.1 Tool Creating a Handle

```
func (m *NSMediator) ToolFsMakeMeta(rt api.Runtime, args []any) (any, error) {
    path := args[0].(string)
    stat, _ := os.Stat(path)
    payload := &FileMeta{Path: path, Size: stat.Size()}
    hv, _ := rt.HandleRegistry().NewHandle(payload)
    return hv, nil
}
```


### 10.2 NS Code Using a Handle

```
let h = TOOL.FS.MakeMeta("src/main.go")
let size = TOOL.FS.Size(h)
print("size is", size)
```


## 11. Future Extensions

- Automatic method dispatch via `h.method()` syntax.
- Borrow semantics for streaming large files.
- Structured AST diff handles.
- Cross-interpreter ephemeral handle export (carefully).

---------------------------

Right now you’ve basically got “handles as cosplay strings”:

`"fsmeta:1234"` – looks like a string, quacks like a string, *is* a string, but you’re treating it as a handle.

You want:

* **true opaqueness** (NS code can’t fake or poke inside them),
* **type detection/checking** (fsmeta vs ast vs whatever),
* ideally without wrecking everything.

Here’s how I’d tighten it up.

---

## 1. Make handles a distinct NS value kind

Core move: stop treating handles as strings *at the NS level*.

Internally you can still store the ID as `"fsmeta:1234"` if you like, but the interpreter should expose a separate value variant, e.g.:

```go
type handleValue struct {
    id   string      // internal unique id, maybe "H~fsmeta~abc123"
    kind string      // "fsmeta", "ast", "graphquery", ...
}
```

NS side:

* `typeof(h)` → `"handle"` (or `"handle:fsmeta"` if you like)
* `h`:

  * **cannot** be concatenated (`h + "x"` is an error),
  * **cannot** be indexed (`h[0]` is an error),
  * **cannot** be iterated (`for x in h` is an error).

Printing it can show something debuggy like:

```text
<handle fsmeta#42>
```

…but that’s *rendered*, not the underlying ID.

This is the single biggest change that makes handles truly opaque:
they stop *being* strings as far as NS is concerned.

You still keep your current implementation (string prefix + seq#) inside `handleValue.id`, but NS never sees that.

---

## 2. Keep type detection via a “kind” tag

You want type detection/checking. Best way: every handle entry carries a **kind** string.

Internally:

```go
type HandleKind string

type handleEntry struct {
    Kind    HandleKind
    Payload any
}
```

Registry stores `map[string]handleEntry`.

Then you can provide helpers in Go:

```go
func GetHandle[T any](rt Runtime, hv api.HandleValue, expect HandleKind) (T, error) {
    entryAny, err := rt.HandleRegistry().GetHandle(hv.HandleID())
    if err != nil { return zero, err }

    entry := entryAny.(handleEntry)
    if entry.Kind != expect {
        return zero, fmt.Errorf("handle kind mismatch: got %s, want %s", entry.Kind, expect)
    }

    payload, ok := entry.Payload.(T)
    if !ok {
        return zero, fmt.Errorf("handle payload type mismatch")
    }
    return payload, nil
}
```

NS-level helpers (implemented in runtime, not by sniffing strings):

* `typeof(h)` → `"handle"`
* `handle_kind(h)` → `"fsmeta"` / `"ast"` / etc.

Now tools can:

```go
func ToolFsSize(rt api.Runtime, args []any) (any, error) {
    hv, ok := args[0].(api.HandleValue)
    if !ok { return nil, fmt.Errorf("expected handle") }

    meta, err := GetHandle[*FileMeta](rt, hv, "fsmeta")
    if err != nil { return nil, err }
    return meta.Size, nil
}
```

Type checking preserved, but implemented via registry metadata, not string prefixes.

---

## 3. Stop exposing sequential IDs; make them unguessable

Right now you’ve got “prefix + seq#”. That’s:

* predictable,
* easy to forge,
* annoying if you *ever* treat handles as capabilities.

Better:

* keep whatever internal counter you like,
* but expose **opaque, random-ish IDs** to the NS value.

For example:

```go
id := fmt.Sprintf("H~%s~%s", kind, ulid.Make().String())
```

or a random 128-bit hex. The actual string only matters to your registry.

Key properties:

* NS can’t guess a valid ID with any reliability.
* If someone tries to fabricate `"H~fsmeta~foo"`, `GetHandle` will fail.
* You can still embed the kind in the ID if you like, but **don’t use that for type-checking**; use the registry’s `Kind` field.

If you still like the human-visible sequential bit for debugging, keep that in the payload (`FileMeta{DebugID: n}`) or log it, not in the NS-visible ID.

---

## 4. Make handles non-serialisable and non-persistable

For opaqueness you also need: **handles never leak into stored data**.

Rules:

* Canonicalisation / serialisation:

  * if a value tree contains a handle, fail with a clear error:

    * `"cannot serialize: value contains non-persistable handle"`.
* Tool implementors:

  * never write handles into graph nodes, ingest plans, etc.
  * if they try, the canonicaliser barks.

This forces the pattern:

* handles are purely *ephemeral runtime objects*,
* all durable information must be pushed into proper nodes/values,
* NS can’t accidentally sneak a handle into long-term data.

---

## 5. Maintain backwards compatibility (if you care)

If you already have scripts that treat handles as strings, you have a choice:

1. **Hard break** (cleanest, but painful):

   * flip them to real handle values,
   * update code that concatenates/prints them.

2. **Soft transition**:

   * for a while, keep accepting both:

     * if a tool gets a `string` that starts with your old handle prefix, treat it like a handle *internally* but warn.
   * add a runtime flag / version to disable “string-handle mode” later.

I’d lean toward a relatively quick hard break in NS 0.9 or whatever, because letting handles masquerade as strings forever will haunt you.

---

## 6. Summary of concrete improvements

To move from “string-with-prefix” to “true opaque handle, but type-checkable”:

1. **Add a dedicated `handleValue` type** to the NS value system.

   * Holds `id string` and `kind string`.
   * `typeof` → `"handle"`.
   * Disallow string/array operations on it.
   * `print` shows only a debug placeholder.

2. **Introduce a proper registry entry type with `Kind`**.

   * Registry maps `id → {Kind, Payload}`.
   * Tools downcast via helper that checks `Kind`.

3. **Change ID scheme to unguessable**.

   * Prefix+ULID or random hex.
   * Stop using monotonically increasing ints in the NS-visible part.

4. **Make handles non-serialisable**.

   * Canonicalisation fails if a handle is present.
   * Tools discouraged from storing handles in persistent data.

5. **Expose a tiny runtime API**:

   * `HandleRegistry().NewHandle(payload, kind)` → `HandleValue`.
   * `HandleRegistry().GetHandle(id)` → entry.
   * Optional `DeleteHandle`.

With that, your handles become:

* *actually* opaque,
* still type-checked,
* safe to use for FS/AST/graph objects,
* and you’re no longer depending on “string prefix cosplay” for semantics.
