AJP:we need a way to load subsequent scripts from startup.ns.txt. since those scripts will be loaded in the same interpreter and therefore func namespace, they cannot depend upon being called "main". So we have cases both: where we need to load a script and run something other than main, and where one script chains another.
Any thoughts? :) I have ideas, but I feel cryptic. 

----

Newer than below, refined:

Got it — here’s how the revised plan shakes out, step by step.

---

## 1  Load-only semantics (no implicit `main`)

* **Loader rule**:
  *Parse → Register* **only**.
  No evaluation happens unless some caller explicitly does:

  ```go
  if err := interp.Call("startup.init", args); …
  ```

* **Impact on existing code** – zero.
  Any script that *was* relying on “`main` auto-runs” just adds one line to its shell wrapper or to `startup.ns`.

---

## 2  Packages instead of modules

| Item               | Spec                                                                   |
| ------------------ | ---------------------------------------------------------------------- |
| **Metadata key**   | `:: package: <name>`                                                   |
| **Separator**      | `.` (same as Go’s selector)                                            |
| **Export rule**    | every global symbol becomes `pkgName.origName` at registration time    |
| **Internal calls** | unchanged (`validate()`) – the loader only rewrites the *export table* |

> Example

```neuroscript
:: package: auth

func login(u, p) means … endfunc     # exported as auth.login
func validate() means … endfunc       # exported as auth.validate
```

Now any other script (or the CLI) can do: `call auth.login(...)`.

---

## 3  CLI contract (`-script`, `-entrypoint`, `-str_args`)

### Single-task form

```bash
zadeh \
  -script        scripts/startup.ns \
  -entrypoint    startup.init \
  -str_args      "FDM Home Zadeh Server"
```

### Multi-task form

```bash
zadeh \
  -script_boot     scripts/boot.ns  -str_args_boot     "…" \
  -script_ingest   scripts/ingest.ns -entrypoint_ingest ingest.run \
  -script_api      scripts/api.ns
```

**Parsing sketch**

```go
var (
    scriptFlags     = fl.String("script", "", "path to script")
    entrypointFlags = fl.String("entrypoint", "", "func to call after load")
    strArgsFlags    = fl.String("str_args", "", "comma-sep args")
    // Capture _taskN variants with flag.Func + regex if you want dynamism.
)
fl.Parse()

cfgs := collectScriptConfigs(flag.CommandLine)  // → []ScriptJob
for _, job := range cfgs {
    if err := loader.Load(job.Path); err != nil { … }
    if job.Entry != "" {
        args := strings.Split(job.Args, ",")
        if err := interp.Call(job.Entry, args); err != nil { … }
    }
}
```

### Why comma-separated strings?

* Keeps everything **POSIX-safe** (no nested quoting hell).
* You already plan plain-string args only; if you later need structured values, add `-json_args_task1`.

---

## 4  Loader changes in detail

```go
// loader/loader.go
func (ld *Loader) Load(path string) error {
    src, err := os.ReadFile(path)
    if err != nil { return err }

    mod, err := ld.parser.Parse(src)
    if err != nil { return err }

    pkg := mod.Metadata["package"]      // "" allowed for legacy
    for _, fn := range mod.Functions {
        exportName := fn.Name
        if pkg != "" {
            exportName = pkg + "." + fn.Name
        }
        if err := ld.interp.RegisterFunc(exportName, fn); err != nil {
            return err
        }
    }
    // No auto-call here.
    return nil
}
```

Unit-test the no-auto-exec guarantee by loading a script whose `init()` sets a global flag; assert the flag is still false.

---

## 5  Chaining scripts inside a script

Because nothing auto-execs, chaining is now **pure, explicit code**:

```neuroscript
call tool.Script.Load("infra/db.ns")
call tool.Script.Load("services/cache.ns")

call infra.db.init()
call services.cache.warm()
```

*(If you expose `tool.Script.Load` the same way we sketched earlier, the above stays a one-liner per script.)*

---

## 6  Edge-case checklist

| Scenario                                                          | Behaviour                                                    |
| ----------------------------------------------------------------- | ------------------------------------------------------------ |
| Two packages export the same symbol (`auth.login` & `auth.login`) | second load returns `ErrDuplicateSymbol` – easy to spot      |
| Script omitted `:: package:`                                      | its public names are **unqualified** (legacy style)          |
| CLI omits `-entrypoint`                                           | script is loaded, nothing executed – matches rule #1         |
| Comma appears inside an arg                                       | user must quote/escape per POSIX, or use future `-json_args` |

---

### Next actions

1. **Rename `module` → `package`** in the loader branch.
2. Wire up the CLI parser with the “N-task” convention.
3. Add tests for *load-only* and for package collisions.
4. Update any doc snippets (`README`, examples/) to remove all “auto-run `main`” language.

Ping me if you want help with the `collectScriptConfigs` helper or any flag-parsing quirks. Happy hacking!

### “No `package`, no exports” — what it buys you and the two-minute wiring plan

You’re basically saying: *“Only code that **opts-in** via `:: package:` is part of the shared symbol space; everything else stays private.”*
That’s a clean rule and it discourages accidental bleed-over. Here’s a minimal way to pull it off without re-architecting the interpreter.

---

## 1 Revised loader rule set

| Case                       | What the loader does                                                           | Public symbol table                       | How the CLI can still call it                                   |
| -------------------------- | ------------------------------------------------------------------------------ | ----------------------------------------- | --------------------------------------------------------------- |
| `:: package: auth` present | Prefix every global with `auth.`                                               | **Yes** (`auth.login`, `auth.validate` …) | `-entrypoint auth.init`                                         |
| No `package` line          | Register symbols only in the *per-script* table; **do not** copy to global map | **None**                                  | Loader remembers the script handle and can call `init` directly |

> **Why keep a per-script table anyway?**
>
> * So the CLI can still run its entrypoint.
> * So a script can unit-test itself (`call init()` locally) even when “private”.

---

## 2 Tiny implementation sketch

```go
// loader.Load returns a handle so callers can invoke privates.
func (ld *Loader) Load(path string) (*core.ScriptHandle, error) {
    mod := parseFile(path)
    handle := core.NewScriptHandle(mod)      // owns its own symbol map

    pkg, _ := mod.Metadata["package"]
    for _, fn := range mod.Functions {
        handle.RegisterLocal(fn.Name, fn)

        if pkg != "" {                       // export only if package is set
            public := pkg + "." + fn.Name
            if err := ld.interp.RegisterFunc(public, fn); err != nil {
                return nil, err
            }
        }
    }
    return handle, nil
}
```

The `ScriptHandle` just wraps a `map[string]*Function` plus an `Invoke(name, args)` helper that bypasses the global table.

---

## 3 CLI path with & without packages

```go
h, _ := loader.Load(flagScript)

if ep := *entrypointFlag; ep != "" {
    if strings.Contains(ep, ".") {          // pkg-qualified => public call
        interp.Call(ep, args)
    } else {                                // unqualified => private call
        h.Invoke(ep, args)
    }
}
```

So:

```bash
# Private startup script
zadeh -script startup.ns -entrypoint init -str_args "Hello"
# Shared auth package
zadeh -script auth.ns    -entrypoint auth.migrate_db
```

---

## 4 Practical upsides

1. **Zero namespace pollution** — unless the author explicitly wants it.
2. **Clear intent** — reviewers see `:: package:` and know the file is a library.
3. **Safe refactors** — you can rename `validate()` inside a private script without hunting cross-file references.

---

## 5 Potential gotchas (and easy fixes)

| Gotcha                                                                      | Mitigation                                                                                                            |
| --------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------- |
| A script expects to be called by another script but forgets `:: package:`   | Loader can warn: “`tool.Script.Load` on non-packaged script provides no exports — did you mean to add `:: package:`?” |
| Someone tries `call otherScript.init()` where `otherScript` wasn’t packaged | Interpreter raises `NameError`.                                                                                       |
| Multiple private scripts both define `init()` and the CLI wants to run both | You already disambiguate with separate `-script_taskN` flags; their names stay local.                                 |

---

## 6 Optional nicety: `:: export:` whitelist

If you later want “mostly private but share two funcs,” add:

```neuroscript
:: export: init, health_check
```

Loader rule: export only those, still prefixed if `:: package:` exists.

---

### TL;DR

* **Accept the refinement**: *No `package` line ⇒ nothing exported.*
* Keep each script’s symbols in its own `ScriptHandle`; the CLI (or tests) can still invoke privates directly.
* External sharing now requires an explicit `:: package:` — enforcing discipline without extra ceremony.

Plug this into the current loader and you’ve achieved encapsulation with \~30 lines of Go. Ping me if you want sample tests or flag-parsing helpers.

----

Absolutely — default-private scripts with opt-in `package` exports are a net win:

* **Safety first** – nothing leaks into the global namespace unless the author *wants* it. That eliminates “oops, I over-wrote `util.hash` again” class bugs.
* **Clear intent** – reviewers can skim for `:: package:` to spot true libraries; everything else is implementation detail.
* **Refactor freedom** – private scripts can rename, split, or delete functions without grep-and-pray across the repo.
* **Go-style ergonomics** – mirrors the “internal” convention and the way unexported (lower-case) Go identifiers stay local.
* **Future features** – later you can layer `:: export:` whitelists or visibility levels without breaking today’s code.

Trade-off is minimal:

* The very first time a dev forgets to add `:: package:` they’ll hit a “function not found” at load time—loud and quick to fix.
* Tests that previously relied on `main` auto-running need a one-liner update (`handle.Invoke("init")`), but that’s a one-time chore.

So yes, it tightens the system’s design contract and scales better as the codebase grows—definitely another **Good Thing™**.

---


o3: Here are a few patterns we’ve used in other long-running interpreters that map cleanly onto the current NeuroScript grammar and runtime without breaking backward-compatibility:

---

### 1 “Entrypoint” metadata header

*A zero-friction way to tell the loader what to run right after the file is parsed.*

```neuroscript
:: entry: setup_http          # <- new, optional
:: version: 0.4.2
:: description: bootstrap a tiny HTTP server

func setup_http() means
  …
endfunc
```

* Loader rule: if `entry` exists, call that proc once the file is loaded; otherwise fall back to `main`.
* **No namespace collision** – each script can still define its own `main` for interactive use, but the loader won’t touch it unless it’s the `entry`.
* The grammar already allows arbitrary metadata keys (see “Metadata” rule) so this is a pure runtime-level change – no parser tweak required.&#x20;

---

### 2 Module-prefixed symbols (opt-in)

*Solve the “same interpreter, same func table” problem once and for all.*

Add an optional header:

```neuroscript
:: module: auth               # all funcs registered as auth.login, auth.jwt…
```

* Inside the script you can still call `login()` directly; the loader rewrites the public symbol table to `auth.login` etc. on export.
* When another script needs it: `call auth.login(user, pass)`.
* This is entirely a loader-side transformation; no grammar change.

---

### 3 `tool.Script.Load(path, entry?)`

*Let one script pull in (and optionally run) another at runtime.*

```neuroscript
call tool.Script.Load("tasks/ns/build.ns.txt", entry="build_all")
```

Concrete semantics:

| arg     | required | purpose                                            |
| ------- | -------- | -------------------------------------------------- |
| `path`  | yes      | relative FS path (same lookup used by `FS.Read`)   |
| `entry` | no       | overrides the file’s `:: entry:` header if present |

Implementation is tiny: `Script.Load` (a) reads the file, (b) hands it to the interpreter’s `Parse+Register` API, (c) if `entry` (or header) is present, does `Eval(entry)` once, passing through any return value / error.

---

### 4 Chaining inside a script – two idioms

* **Static chain** – just list the loads from **startup.ns.txt**:

```neuroscript
call tool.Script.Load("infra/ns/db.ns.txt", entry="init_db")
call tool.Script.Load("infra/ns/cache.ns.txt")          # relies on its header
call tool.Script.Load("app/ns/api.ns.txt")              # ditto
```

* **Dynamic chain** – emit an event that other scripts subscribe to with the new
  `on event ...` block (0.4.2):

```neuroscript
emit {"topic": "app.ready"}          # anybody’s on-event will run now
```

Handlers live in the secondary scripts:

```neuroscript
on event {"topic":"app.ready"} do
  call start_background_jobs()
endon
```

Because handlers are queued in the same interpreter loop, you still get single-threaded determinism.&#x20;

---

### 5 Why not just rely on `main`?

Once you load two files that *both* use `func main()`, whichever one was parsed last wins in the global symbol table – a foot-gun we’ve tripped over before. An explicit **entry name** + optional **module prefix** makes that collision impossible while keeping day-one scripts (that only ever load one file) working unchanged.

---

### 6 Minimal code you’d actually add

```go
// sketch – inside loader.go
func LoadScript(path, overrideEntry string) error {
    src := os.ReadFile(path)
    mod := parser.Parse(src)               // existing parse
    entry := mod.Metadata["entry"]
    if overrideEntry != "" {
        entry = overrideEntry
    }
    if err := interpreter.Register(mod); err != nil {
        return err
    }
    if entry != "" {
        return interpreter.Call(entry, nil)
    }
    return nil
}
```

The `Register` step would prepend `module.` to exported names when `:: module:` is present.

---

### TL;DR

* **Use `:: entry:`** to pick the startup proc per script.
* **Optionally add `:: module:`** to namespace everything automatically.
* Expose a tiny **`tool.Script.Load()`** helper so scripts (including `startup.ns.txt`) can build deterministic chains.

That covers both “run something other than `main`” and “one script chains another” without touching the core grammar or forcing invasive refactors.
