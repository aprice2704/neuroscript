# NeuroScript Integration Guide (v0.4 — 2025‑07‑22)

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
| **Exec** | `ExecInNewInterpreter` | `ExecWithInterpreter` | `ExecWithInterpreter` | `ExecWithInterpreter` |

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

Use your own Ed25519 key to sign the **hash (`sum`)** of the canonical blob, not the blob itself.

```go
// Sign the hash, not the full blob.
sig := ed25519.Sign(privKey, sum[:])

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

* **Stateless (Trusted Source)**
  ```go
  // For quick, untrusted scripts, parse and execute directly.
  result, err := api.ExecInNewInterpreter(ctx, string(srcBytes))
  ```

* **Stateful (Verified Source)**
  ```go
  // For verified scripts, use the tree from the loaded unit.
  interp := api.New()
  result, err := api.ExecWithInterpreter(ctx, interp, lu.Tree)
  ```

`ExecWithInterpreter` loads the program's definitions into the interpreter and
runs any top-level `command` blocks.

---

## 3. Registering Custom Tools

You can extend the NeuroScript interpreter with custom functionality by registering your own tools. A tool is a standard Go function that conforms to the `api.ToolFunc` signature and is packaged in an `api.ToolImplementation` struct.

**Example:**

```go
package main

import (
	"context"
	"fmt"
	"os"

	"[github.com/aprice2704/neuroscript/pkg/api](https://github.com/aprice2704/neuroscript/pkg/api)"
)

// 1. Define your tool's function. It must match the api.ToolFunc signature.
func GreeterFunc(rt api.Runtime, args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("greeter tool expects 1 argument, got %d", len(args))
	}
	name, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("greeter tool expects a string argument, got %T", args[0])
	}
	// Use the runtime to interact with the interpreter, for example, to print output.
	rt.Println(fmt.Sprintf("Hello, %s!", name))
	return true, nil // Return a primitive Go type.
}

func main() {
	// 2. Define the tool's implementation using re-exported API types.
	greeterImpl := api.ToolImplementation{
		FullName: "host.greeter",
		Spec: api.ToolSpec{
			Name: "greeter",
			Group: "host",
			Args: []api.ArgSpec{{Name: "name", Type: "string", Required: true}},
			ReturnType: "bool",
		},
		Func: GreeterFunc,
	}

	// 3. Create an interpreter option for your tool using the API helper.
	toolOpt := api.WithTool(greeterImpl)

	// 4. Create a new interpreter with your tool registered.
	// Note: api.WithStdout is also re-exported from the api package.
	interp := api.New(toolOpt, api.WithStdout(os.Stdout))

	// Now you can run NeuroScript code that uses your tool.
	src := `command { emit host.greeter("World") }`
	_, err := api.ExecWithInterpreter(context.Background(), interp, &api.Tree{/* ...parsed tree... */})
	if err != nil {
		fmt.Println("Execution failed:", err)
	}
}
```

---

## 4  Interpreter facade (high‑level API)

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

## 5  Tool interop (Go ↔ NeuroScript)

A `tool.ToolImplementation` uses primitive Go types; the registry takes care
of wrapping/unwrapping `lang.Value`s. See the bundled example in the template
repo.

---

## 6  Core types & enums

* `api.Tree`, `api.Kind`, `api.Position`, `api.Node`
* `api.SignedAST`, `api.LoadedUnit`, `api.Value`
* `api.RunMode{Library, Command, EventSink}`
* `api.ParseMode{PreserveComments, SkipComments}`

---

## 7  Important “Don’ts”

* **Do not** import `pkg/parser`, `pkg/canon`, `pkg.com/interpreter`, etc.
  `api` already re‑exports what you need.
* **Do not** execute a tree that skipped `api.Load` when security matters.
* **Do not** re‑canonicalise after verification — keep the original `blob`
  and `sum`.

---

## 8  Metadata

::name: NeuroScript Integration Guide
::schema: spec
::serialization: md
::fileVersion: 8
::author: Andrew Price
::created: 2025‑07‑16
::modified: 2025‑07‑22
::description: Accurate, up‑to‑date instructions for integrating NeuroScript
  via the public `api` package; aligned with contract v0.6.
::tags: guide, integration, api, neuroscript, golang
::howToUpdate: Update call‑flows and type names whenever the API contract
  increments. Bump `fileVersion`.
::dependsOn: api/parse.go, api/canon.go, api/loader.go, api/exec.go,
  api/interpreter.go, api/reexport.go