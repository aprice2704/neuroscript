# FDM / NeuroScript — Agent-Facing Guide to Script Loading & Packages

*(Updated: now clarifies multi-file packages and the “Pottery Barn” rule.)*

---

## 1 Design Principles

| Principle | What it means |
|-----------|---------------|
| **Opt-in visibility** | A script exports *nothing* unless it declares `:: package:`. |
| **No implicit `main`** | Loading ≠ running.  A function runs **only** when explicitly invoked. |
| **Single global symbol space** | All exports are `<package>.<func>`, preventing accidental shadowing. |
| **Multi-file packages** | Several files may share the same `:: package:`; loader merges them.  Duplicate *function* names → fast-fail error. |
| **Pottery Barn rule** | *“You load it **or** run it, you own it.”*  Any agent/script that loads or invokes code is responsible for handling its failures & conflicts. |
| **Uniform CLI contract** | Every FDM/NS binary understands the same `-script`, `-entrypoint`, `-str_args` pattern. |

---

## 2 Script Anatomy

### 2.1 Package header *(exports enabled)*

```neuroscript
:: package: auth

func login(u, p) means … endfunc   # exported as auth.login
func validate()   means … endfunc  # exported as auth.validate
```

*Multiple files* may also use `:: package: auth`.  
If **two functions export the same name**, the loader stops with  
`ErrDuplicateSymbol` — satisfy the Pottery Barn rule by avoiding collisions.


### 2.2 Private script *(no package header)*

```neuroscript
# helpers/cleanup.ns         (nothing exported)

func init() means … endfunc
```

The file’s functions stay in its **local** symbol table; other scripts
cannot call them via the global interpreter.

---

## 3 Loading & Invoking at Runtime

| Action                           | Example inside NeuroScript                 |
|----------------------------------|--------------------------------------------|
| Load another file                | `call tool.Script.Load("infra/db.ns")`     |
| Call an exported function        | `call auth.login(user, pass)`              |
| Call a *private* function        | `call tool.Script.Invoke("cleanup.init")`  |

*`tool.Script.Load`* returns a **handle**; use `Invoke` on that handle to
reach private symbols without polluting the global map.

---

## 4 Standard CLI Flags (all executables)

### 4.1 Single-task invocation

```bash
zadeh \
  -script        scripts/startup.ns   \
  -entrypoint    startup.init         \
  -str_args      "FDM Home Zadeh Server"
```

| Flag          | Meaning                                                    |
|-------------- |------------------------------------------------------------|
| `-script`     | Path to the NeuroScript file to **load**                   |
| `-entrypoint` | Function to **run** after loading (public `pkg.fn` **or** private `fn`) |
| `-str_args`   | Comma-separated plain strings passed as arguments          |

### 4.2 Multi-task invocation

```bash
zadeh \
  -script_boot   scripts/boot.ns   -str_args_boot   ""            \
  -script_jobs   scripts/jobs.ns   -entrypoint_jobs jobs.run       \
  -script_api    scripts/api.ns
```

Pattern: `-script_<tag>`, `-entrypoint_<tag>`, `-str_args_<tag>`.

---

## 5 Examples in Practice

### 5.1 Chaining without globals

```neuroscript
# startup.ns
call tool.Script.Load("infra/db.ns")          # load-only
call infra.db.migrate()                       # explicit call

call tool.Script.Load("services/cache.ns")
call services.cache.warm()
```

### 5.2 Private utility with explicit invoke

```bash
zadeh -script helpers/cleanup.ns -entrypoint init
```

Even though `cleanup.ns` exports nothing, the CLI still runs `init`
via the script’s local handle.

---

## 6 Best-Practice Checklist

* **Declare `:: package:`** the moment you expect cross-script calls.
* Keep scripts private unless sharing is **required**.
* Use meaningful package prefixes (`auth.db`, `auth.jwt`, `utils.strings`).
* Resolve duplicate-symbol errors promptly — *Pottery Barn rule*.
* Document expected entrypoints in header comments.

---

## 7 FAQ

**Q: What if two files in the same package declare `login()`?**  
A: Loader fails fast.  The loader/agent that triggered the load “owns” the
   fix — rename or refactor per the Pottery Barn rule.

**Q: Can I split a package into functional slices?**  
A: Yes. Prefer sub-packages (`auth.db`, `auth.jwt`) over numeric suffixes
   for readability.

**Q: Commas inside `-str_args`?**  
A: Escape per POSIX, or switch to future `-json_args` once supported.

---

**Bottom line:**  
Scripts are private by default; packages export intentionally; loaders
enforce fast-fail on collisions; and whoever loads or runs code “owns”
any mess created—FDM’s **Pottery Barn rule** in action.


:: language: markdown
:: lang_version: n/a
:: file_version: 1
:: author: OpenAI Assistant
:: created: 2025-06-25
:: modified: 2025-06-25
:: description: Agent-facing guide to script loading, packages, CLI flags, and the Pottery Barn rule in FDM / NeuroScript.
:: tags: guide, fdm, neuroscript, packages, cli, pottery-barn-rule, loading
:: type: documentation
:: subtype: guide
:: dependsOn: docs/metadata.md
