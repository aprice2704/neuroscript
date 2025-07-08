## NeuroScript - Developer Primer (v0.4)

> **Tag-line:** *Procedural scaffolding for humans, AIs and plain Go code.*

---

### 1. Why bother?

* **Readable steps, no hidden “prompt soup”.**
  Code is the chain-of-thought.

* **Single source of truth.**
  A `.ns` file *is* documentation, executable logic, and metadata for search/RAG.

* **Interop with Go.**
  Tools map 1-to-1 onto Go functions; the interpreter is a thin wrapper layer.

---

### 2. File skeleton

```neuroscript
:: lang_version: neuroscript@0.4.0
:: file_version: 1

func DoThing(needs x, y optional opts returns out1, out2) means
  :: description: Adds then multiplies.         # metadata
  set sum = x + y                               # statement
  set prod = sum * (opts["factor"] or 1)
  return sum, prod
endfunc
```

* **Metadata** (`:: key: value`) **always first** – parsed, indexed, searchable.
* **`func … endfunc`** – one procedure = one reusable “skill”.
* **Signature clauses:** `needs`, `optional`, `returns`; parentheses optional.

---

### 3. Types at a glance

| Kind         | Literal / creator                                    | Notes                                                 |
| ------------ | ---------------------------------------------------- | ----------------------------------------------------- |
| `string`     | `"hi"` or ` `raw` `                                  | Triple-backtick raw strings allow `{{placeholders}}`. |
| `number`     | `1`, `3.14`                                          | `int64` or `float64` under the hood.                  |
| `bool`       | `true`, `false`                                      |                                                       |
| `list`       | `[1, 2, 3]`                                          | Any element types.                                    |
| `map`        | `{"k": 1}`                                           | Keys must be string literals.                         |
| `nil`        | `nil`                                                | Absence of value.                                     |
| **New v0.4** |                                                      |                                                       |
| `error`      | Returned by tools: `{"code":"ENOENT","message":"…"}` |                                                       |
| `timedate`   | `tool.Time.Now()`                                    | Wraps `time.Time`.                                    |
| `event`      | System events, emitted by runtime.                   |                                                       |
| `fuzzy`      | Tool-created real ∈ \[0,1]                           | Fuzzy logic operators `and`/`or`/`not`.               |

`typeof(expr)` → string constants (`"string"`, `"list"`, …) for quick checks.

---

### 4. Multiple return values

* **Always positional** (`return a, b`) → caller receives list `[a, b]`.
* Recommended pattern: immediately destructure:

```neuroscript
set sum, product = MyMath(needs x, y)
```

---

### 5. Error handling – fail fast, fail loud

| Tool returns         | You write                                                    | Result           |
| -------------------- | ------------------------------------------------------------ | ---------------- |
| **Normal value**     | `set data = must tool.FS.Read(path)`                         | `data` assigned. |
| **`error` map**      | `must` triggers runtime error → jumps to nearest `on_error`. |                  |
| **Unexpected panic** | Propagates straight to `on_error`.                           |                  |

**Idioms**

```neuroscript
on_error means                # try/catch
  emit "fatal: " + system.error_message
  return "Failed"
endon

set cfg  = must tool.JSON.Parse(text)         # parse or die
set port = must cfg["port"] as int            # key + type assertion
```

* `must expr` – boolean assertion.
* `set v = must expr` – mandatory success check (error-map aware).
* `fail "msg"` – deliberate abort.
* `clear_error` – swallow inside an `on_error` block.

---

### 6. Control flow & statements

* `if / else / endif`
* `while / endwhile`
* `for each item in list / endfor`
* `break`, `continue`
* `return`, `emit`, `call` (side-effect invoke)

---

### 7. Tools & built-ins

* **Tools**: namespaced Go functions (`tool.FS.Read`, `tool.JSON.Parse`).
  *Adapters unwrap/rewrap values; your Go code sees primitives.*

* **Built-ins**: maths & helpers (`ln`, `sin`, `typeof`).
  Identical calling convention to tools.

Tool names **must** be complete: tool.<group>.<action> where group is "FS", "io" etc. and action is the action the tool performs, e.g. "read", "write". Hence: tool.fs.read
Tool names are **case-insensitive**

---

### 8. AI integration (`ask`)

```neuroscript
set prompt = eval("Summarise {{topic}} in 3 bullets.")
ask prompt into summary
```

* Routes to configured LLM.
* Response stored in var **and** `last`.

---

### 9. Convenience features

* **`last`** – value from most recent successful call.
* **Placeholders** – `{{var}}` auto-resolved in raw strings, or via `eval()` in normal strings.
* **Line continuation** – trailing `\` to wrap long expressions.

---

### 10. Mental model

```
Interpreter (core.Value wrappers)
        │
        ▼
   Adapter layer
        │
   ┌────┴─────────────┐
   │ Built-in fns     │  (primitives)
   │ Tool validators  │
   │ Tool impls       │
   └──────────────────┘
```

One choke-point = one place to debug.

---

#### Next steps

* Browse the illustrative examples in `prompts.go`.
* Read `ns_script_spec.md` for deeper semantics.
* Implement your own tool in Go: write a plain function, register, done.

> “Write procedures like you’d write comments—then run them.”
