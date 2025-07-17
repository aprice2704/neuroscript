# NeuroScript Integration Guide (v0.4 — 2025‑07‑16)

This guide shows **exactly** how to embed NeuroScript/FDM into an external Go
application using **only** `import "yourrepo/api"`.  
Under the hood `api` wraps the parser, canonicaliser, loader, and interpreter
so you never import those sub‑packages directly.

> **Golden path:** `Parse → Canonicalise → Sign → Load → Exec*`

---

## 1  Execution models

NeuroScript supports four host‑level workflows, formed by crossing
**Statefulness** (stateless vs. persistent interpreter) with **Security**
(trusted vs. cryptographically verified source):

| | **Stateless (one shot)** | **Stateful (long‑running)** |
| :--- | :--- | :--- |
| **Trusted** | **Mode 1** – quick & dirty | **Mode 3** – daemon/service |
| **Verified** | **Mode 2** – one‑off, signed | **Mode 4** – multi‑tenant |

### API calls per mode

| Step | Mode 1 | Mode 2 | Mode 3 | Mode 4 |
|------|--------|--------|--------|--------|
| **Parse** | ✅ | ✅ | ✅ | ✅ |
| **Canonicalise** | optional | ✅ | optional | ✅ |
| **Sign** | _—_ | ✅ | _—_ | ✅ |
| **Load** | _—_ | ✅ | _—_ | ✅ |
| **Exec** | `ExecInNewInterpreter` | same | `ExecWithInterpreter` | same |

---

## 2  Step‑by‑step workflow

### 2.1 Parse

```go
tree, err := api.Parse(srcBytes, api.ParseSkipComments)
```

`Parse` returns a `*api.Tree` and never touches the network; it is safe for
untrusted input.

### 2.2 Canonicalise

```go
blob, sum, err := api.Canonicalise(tree)
```

Deterministic bytes + **blake2b‑256** hash. The extra `error` return was
added in contract v0.6.

### 2.3 Sign (host responsibility, optional)

Use your own Ed25519 key and package the result into:

```go
signed := &api.SignedAST{Blob: blob, Sum: sum, Sig: sig}
```

### 2.4 Load (verification & vetting)

```go
lu, err := api.Load(ctx, signed, api.LoaderConfig{}, pubKey)
```

On success you receive a `*api.LoadedUnit` with `Tree`, `Hash`, `Mode`,
and the original `RawBytes`.  
**Never** re‑canonicalise after this point; the loader already did.

### 2.5 Execute

* **Stateless**  
 ```go
 result, err := api.ExecInNewInterpreter(ctx, lu, api.ExecConfig{})
 ```

* **Stateful** – reuse an interpreter
 ```go
 interp := api.New()                                          // create once
 cfg   := api.ExecConfig{Interpreter: interp}
 result, err := api.ExecWithInterpreter(ctx, lu, cfg)
 ```

`ExecWithInterpreter` auto‑loads the program into the provided interpreter
and dispatches according to `lu.Mode`.

---

## 3  Interpreter facade (high‑level API)

Create a persistent VM with:

```go
interp := api.New(api.WithStdout(os.Stdout))
```

Key methods:

| Method | Purpose |
|--------|---------|
| `Load(*ast.Program)` | inject a verified program |
| `ExecuteCommands()` | run unnamed `command` block |
| `Run("procName")` | call a procedure |
| `EmitEvent(...)` | push an event into an event‑sink script |

---

## 4  Tool interop (Go ↔ NeuroScript)

A `tool.ToolImplementation` uses primitive Go types; the registry takes care
of wrapping/unwrapping `lang.Value`s. See the bundled example in the template
repo.

---

## 5  Core types & enums

* `api.Tree`, `api.Kind`, `api.Position`, `api.Node`  
* `api.SignedAST`, `api.LoadedUnit`, `api.ExecResult`  
* `api.RunMode{Library, Command, EventSink}` :contentReference[oaicite:6]{index=6}  
* `api.ParseMode{PreserveComments, SkipComments}` :contentReference[oaicite:7]{index=7}

---

## 6  Important “Don’ts”

* **Do not** import `pkg/parser`, `pkg/canon`, `pkg/interpreter`, etc.  
 `api` already re‑exports what you need.  
* **Do not** execute a tree that skipped `api.Load` when security matters.  
* **Do not** re‑canonicalise after verification — keep the original `blob`
 and `sum`.

---

## 7  Metadata

::name: NeuroScript Integration Guide  
::schema: spec  
::serialization: md  
::fileVersion: 4  
::author: Andrew Price  
::created: 2025‑07‑16  
::modified: 2025‑07‑16  
::description: Accurate, up‑to‑date instructions for integrating NeuroScript
  via the public `api` package; aligned with contract v0.6.  
::tags: guide, integration, api, neuroscript, golang  
::howToUpdate: Update call‑flows and type names whenever the API contract
  increments. Bump `fileVersion`.  
::dependsOn: api/parse.go, api/canon.go, api/loader.go, api/exec.go,
  api/interpreter.go, api/reexport.go
