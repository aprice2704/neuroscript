### NeuroScript — Your AI-native, event-savvy scripting layer

*Hand-optimised for humans, Go code, and large language models.*

---

#### 1. **Why NeuroScript?**

| Challenge                        | How NeuroScript helps                                                                                                                                     |
| -------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Hidden prompt spaghetti          | Step-by-step procedures that read like code yet double as chain-of-thought.                                                                               |
| Uncaught runtime surprises       | Hardened **`must` / `on_error`** model keeps failures loud and contained .                                                                                |
| Glue code scattered across tools | One language to script AI calls, file I/O, JSON, DBs, HTTP… anything you register as a `tool.*`.                                                          |
| Evolving requirements            | Metadata lines (`:: key: value`) make every file self-describing and versioned for painless diffing and search .                                          |
| Reactive automation              | First-class **events** and top-level **`on event ... endevent`** handlers let scripts wake up on “file\_uploaded”, “model\_done”, etc., without polling . |

---

#### 2. **Flagship capabilities**

##### • Event-driven hooks

Define global handlers once and forget cron jobs:

```neuroscript
on event UserSignedUp(payload) means
  set welcome = "Hi " + payload["name"] + "!"
  call tool.Email.Send(payload["email"], welcome)
endevent
```

The runtime queues incoming `event` values (name, timestamp, payload) and dispatches them transactionally.

##### • Industrial-grade error handling

* `must expr` turns any falsy check or `error` map into a panic caught by the nearest `on_error` block.
* `fail "msg"` aborts intentionally.
* `clear_error` lets you recover and continue .

##### • Multi-return & destructuring

Procedures can return many values without tuple boiler-plate:

```neuroscript
set sum, prod = Math.AddMul(needs 3, 5)
```

Callers get an ordered list, tests stay simple .

##### • Rich, explicit types

Beyond the obvious (`string`, `number`, `list`, `map`) you get:

* `error` – standard structured error maps.
* `timedate` – wraps `time.Time`.
* `event` – runtime signals.
* `fuzzy` – 0-1 truth values with min/max logic .

##### • Built-in AI verb

Send prompts directly from code:

```neuroscript
ask "Summarise: " + doc into summary
```

`summary` is plain text; the last LLM reply is always in `last` for quick chaining .

---

#### 3. **Developer ergonomics you’ll feel day one**

* **Single-file onboarding** – paste a `.ns` file, run it; no build step.
* **Go-first interop** – register a Go function once; adapters unwrap primitives for you.
* **Readable diffs** – metadata up top, business logic below.
* **Searchable knowledge base** – every procedure is a discoverable “skill” for humans *and* agents.
* **Zero-ceremony concurrency** – event handlers run in isolation; state stays per-event.

---

#### 4. **Under the hood**

```
Interpreter (core.Value wrappers)
         │
         ▼
   Adapter layer
   ├── Built-ins   (sin, typeof…)
   ├── Tools       (your Go functions)
   └── Event bus   (enqueue / dispatch)
```

A strict wrapper ↔ primitive boundary keeps core deterministic while letting tools live in idiomatic Go.

---

#### 5. **Ready to try?**

1. Install the Go package `github.com/aprice2704/neuroscript`.
2. Register a tool:

```go
ns.Register("FS.Read", func(path string) (string, error) { … })
```

3. Drop a `.ns` file in your project and call `Interp.Run("file.ns")`.

That’s it—you’ve added an event-aware, AI-native automation layer without sacrificing type safety or your debugging sanity.

> *Write procedures like comments, run them like code.*
