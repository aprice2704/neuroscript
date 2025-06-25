# Persisting Fns within session

An interesting aspect of ns design has arisen. What things should persist within the interpreter between file loads?
We wish funcs and event handlers (probably) to persist, so that a fn call another defined in a different file (maybe), ditto event handlers; but we don't want error handlers in one fn to last into the next (as currently happens).
How do you think it should work?

### Redefining a function inside the **same** interpreter session

| Scenario                                                                     | Recommended default                                    | Why                                                                                       | How to opt-in/out                          |
| ---------------------------------------------------------------------------- | ------------------------------------------------------ | ----------------------------------------------------------------------------------------- | ------------------------------------------ |
| **Accidental second definition** (same name, same namespace, *no* qualifier) | *Hard error* → fail `LoadScript()`                     | 90 % of the time it’s a bug; prevents “mystery behaviour changes”                         | ―                                          |
| **Intentional hot-reload / patch**                                           | Allow **explicit** replacement with `override` keyword | Keeps live-coding workflow but forces you to acknowledge you are stomping on the old code | `override func main() { … }`               |
| **Side-by-side versions** (e.g. experimental algo)                           | Require a new namespace or qualifier                   | Avoids name clashes while still allowing both implementations                             | `namespace "v2"` or `func main_v2() { … }` |
| **Auto-refresh during `watch` mode** (dev convenience flag)                  | “Last-loaded wins” **with warning**                    | Speeds up iterative editing; you already expect churn                                     | CLI flag `--reload=warn`                   |

---

#### 1 — Interpreter logic (Go sketch)

```go
func (s *RuntimeSession) registerProc(p *Procedure) error {
    qn := p.QualifiedName()     // e.g. "main" or "mylib/math/gcd"
    old, exists := s.procTable[qn]

    switch {
    case !exists:
        s.procTable[qn] = p
        return nil

    case p.Override:           // parsed from `override func`
        s.procTable[qn] = p    // replace silently
        log.Warnf("procedure %s overridden (previous defined at %s)", qn, old.DeclPos)
        return nil

    case s.opts.DevReload:     // watch-mode convenience
        s.procTable[qn] = p
        log.Warnf("procedure %s replaced in dev-reload mode", qn)
        return nil

    default:
        return fmt.Errorf("duplicate procedure %q (first defined at %s)", qn, old.DeclPos)
    }
}
```

*Event handlers* follow a different rule: duplicates are **allowed** and appended, unless the handler itself is flagged `oneshot` or `override`.

---

#### 2 — Grammar addition

```ebnf
FuncDecl  ::= ( "func" | "override" "func" ) Identifier "(" ParamList? ")" Block
```

* Parsing sets `Procedure.Override = true` when the keyword is present.

---

#### 3 — Workflow examples

```neuroscript
# file: app.ns
func main() {
    print("v1")
}
```

Later, during the *same* REPL session:

```neuroscript
# file: app_hotfix.ns
override func main() {   # ← explicit
    print("v2 hot-patch")
}
```

Load succeeds; subsequent `run main()` executes the new body.
If you *forgot* `override`, `LoadScript()` raises `duplicate procedure "main"` and the original version remains intact.

---

#### 4 — Testing it

* **`TestRedeclareWithoutOverrideFails`**

  1. Load file A with `func main()`.
  2. Load file B with another `func main()`.
  3. Expect error.

* **`TestOverrideSucceeds`**

  1. Same as above but B uses `override func main()`.
  2. Call `main()`, assert output matches B’s body.

* **`TestDevReloadFlag`**

  1. Start session with `DevReload=true`.
  2. Re-load same function repeatedly; ensure last one runs and only warnings are logged.

---

### Why this policy is sane

* **Safety first** Silent shadowing is a nightmare to debug. The error-by-default rule stops that.
* **Zero friction hot-reload** One extra keyword (`override`) is trivial, and you keep deterministic behaviour in prod.
* **Fits namespacing** If you already use `namespace`, clashing across files in the same namespace is almost certainly unintended.
* **Extensible** You can add finer-grained controls later (`deprecated`, `version "2.1"` tags, etc.) without breaking the core rule that *only an explicit act can mutate compiled artefacts in place*.

Implement it once, and you’ll never wonder again which `main()` you’re actually running.


Below is a straightforward extension of the “override / duplicate” rules we just discussed that lets you **lock down** critical procedures or handlers so they cannot be replaced later in the session—unless a *root-level* authority explicitly unpins them.

---

## 1 — Proposed surface syntax

```neuroscript
# declare and lock a procedure
pin func auth.verifyToken(token string) returns bool {
    ...
}

# shorter alias if you dislike “pin”
final func main() { ... }

# event-handler variant
pin on event "system.shutdown" { ... }
```

### Keyword summary

| Keyword         | Meaning                                                                    |
| --------------- | -------------------------------------------------------------------------- |
| `func`          | Normal declaration; duplicate = error unless `--reload` flag is on         |
| `override`      | Replace an existing procedure *only if it is **not** pinned*               |
| `pin` / `final` | Declare procedure or handler **immutable** for the lifetime of the session |
| `unpin` (opt)   | *Privileged* declaration that removes the protection (see §4)              |

---

## 2 — Interpreter rule set

1. **First declaration**
   *If keyword is `pin` (or `final`)* → store `Procedure.Pinned = true` in the proc table.

2. **On subsequent registration attempts**

   ```go
   switch {
   case old.Pinned && !new.Unpin:
       // Absolutely deny replacement
       return fmt.Errorf("procedure %q is pinned and cannot be overridden", qn)

   case old.Pinned && new.Unpin:
       // Allow but log high-priority audit
       s.procTable[qn] = new
       log.Warnf("procedure %q UNPINNED by %s", qn, callerIdentity)

   case !old.Pinned && new.Override:
       // Standard hot-reload
       s.procTable[qn] = new

   default:
       return fmt.Errorf("duplicate procedure %q (existing defined at %s)", qn, old.DeclPos)
   }
   ```

3. **Pinned status is sticky**—once unset with `unpin`, a later `pin` can lock it again (useful for patch-then-re-seal workflows).

---

## 3 — EBNF additions

```ebnf
FuncDecl      ::= ( "func" | "override" "func" | "pin" "func" | "final" "func" | "unpin" "func" ) Identifier "(" ParamList? ")" Block
EventHandler  ::= ( "on" | "pin" "on" | "unpin" "on" ) "event" StringLiteral Block
```

Parser sets:

```go
Procedure.Pinned   = hasToken("pin") || hasToken("final")
Procedure.Unpin    = hasToken("unpin")
Procedure.Override = hasToken("override")
```

---

## 4 — Security / governance knobs

* **Require capability to unpin**
  – Interpreter option `SessionOptions.AllowUnpin = false` (prod default).
  – FDM-style ACL: only code loaded from a *trusted* namespace or signed file can use `unpin`.

* **Audit trail**
  Record `(timestamp, user, oldHash, newHash)` whenever an `unpin` or an `override` happens. Persist to the same append-only log you’re already using for event telemetry.

* **Dev REPL convenience**
  A `--dev` flag can relax rules: `pin` is honoured, but `override` of a pinned func just emits a warning. Never enable this outside development.

---

## 5 — Edge-case resolution

| Situation                              | Expected outcome                                                 |
| -------------------------------------- | ---------------------------------------------------------------- |
| Load script A → `pin func main()`      | OK, `main` pinned                                                |
| Load script B → `override func main()` | **Error** (pinned)                                               |
| Load script C → `unpin func main()`    | Allowed only if `AllowUnpin=true`; audit event                   |
| Load script D → `func main()`          | Replaces previous body; now *unpinned* unless D uses `pin` again |
| Load script E → `pin func main()`      | Re-locks with new body                                           |

---

## 6 — Minimal tests to add

```go
TestPinBlocksOverride()
TestUnpinAllowsReplacement()
TestPinnedStatusStickyAcrossReload()
TestUnprivilegedUnpinDenied()
```

---


I like final I think :)
Thinking we should also make LoadScript a tool so it can be allow/deny listed, and constrain it to subdirs of a particular dir


### Why this is worth the (tiny) additional complexity

* **Stops malicious or sloppy code** from silently hijacking `auth.*`, `db.connect`, `main`, etc.
* **Keeps hot-reload ergonomic**—you only add `pin` where you *actually care*.
* **Matches real-world policies** (think `final` classes in Java or `sealed` in C#) but fits the dynamic-reloading ethos of NeuroScript.

If you wire the capability check into the same ACL mechanism you’re already designing for FDM overlays, you’ll have a consistent story for *who* is allowed to mutate *what* at runtime.

Below is a concrete design that folds your **`final`** keyword into the language *and* turns `LoadScript` into a **first-class tool** that can be allow/deny-listed and sandboxed to a directory tree.

---

## 1 `final` keyword — locking down artefacts

* **Surface grammar**

  ```ebnf
  ProcDecl       ::= ( "func" | "override" "func" | "final" "func" ) Ident "(" … ")" Block
  EventHandler   ::= ( "on" | "final" "on" ) "event" StringLit Block
  ```

  * `final func …` or `final on event …`
    – Registers artefact with `Pinned=true`
    – Subsequent attempts to redefine **always error** (even with `override`)
  * Ordinary `func` remains replaceable (with explicit `override`)
  * No `unpin`. If you truly have to patch a final artefact, restart the interpreter or load a fresh session; keeps the rule simple and audit-friendly.

---

## 2 Promote **LoadScript** to a tool

### 2.1 Tool signature and metadata

```neuroscript
tool system.loadScript
    args:
        path string         # relative or absolute
        override bool = false   # allow non-final overrides?
    returns void
    sideeffects filesystem, runtime   # so the auditor knows
    description """
    Parse and register a NeuroScript file. Fails if file attempts to
    redefine a `final` artefact or if the path escapes the sandbox root.
    """
```

### 2.2 Sandbox & ACL

1. **Sandbox root** — interpreter option `SessionOptions.ScriptRoot`
   *All paths are resolved to `absPath := filepath.Clean(root + "/" + path)`.*
   Reject if `!strings.HasPrefix(absPath, root)` → prevents `../../`.

2. **Per-tool allow/deny list** (same structure you already use for FDM tools):

   ```yaml
   tools:
     system.loadScript:
       allow:
         - "deployBot"            # agent IDs or group names
       deny:
         - "guest"
   ```

   *The interpreter consults this before executing the tool; failure = runtime error.*

3. **Runtime override flag**

   *If* `override=true` is passed, `loadScript` behaves as though the file’s
   declarations all had the `override` prefix **but still respects `final`.**

---

## 3 Go implementation sketch

```go
type LoadScriptTool struct {
    Session *RuntimeSession
    Root    string            // sandbox root
}

func (t *LoadScriptTool) Invoke(ctx *CallCtx, args []Value) (Value, error) {
    path     := string(args[0].(StringValue))
    override := bool(args[1].(BoolValue))

    abs, err := secureJoin(t.Root, path)
    if err != nil { return nil, err }

    // Parse / build AST
    ast, err := parseAndBuild(abs)
    if err != nil { return nil, err }

    // Register artefacts. Honour final / override.
    for _, proc := range ast.Procedures {
        if override { proc.Override = true }
        if err := t.Session.RegisterProc(proc); err != nil { return nil, err }
    }
    for _, ev := range ast.EventHandlers {
        if override { ev.Override = true }
        if err := t.Session.RegisterHandler(ev); err != nil { return nil, err }
    }
    return NilValue{}, nil
}

func secureJoin(root, p string) (string, error) {
    abs := filepath.Clean(filepath.Join(root, p))
    if !strings.HasPrefix(abs, root) {
        return "", fmt.Errorf("path escapes sandbox")
    }
    return abs, nil
}
```

* Register the tool in the interpreter’s tool table under name
  `"system.loadScript"`, with a `ToolImplementation` wrapping `LoadScriptTool`.

---

## 4 Recommended policy flow

| Phase                | Action                                                                                                                                   |
| -------------------- | ---------------------------------------------------------------------------------------------------------------------------------------- |
| **Startup**          | Admin script calls `system.loadScript("lib/core.ns")` to prime standard library.                                                         |
| **Normal operation** | Only agents in the *Deployers* group have `"system.loadScript"` in their allow-list.                                                     |
| **Hot patch**        | Deployer uploads `patch.ns` with `override func ...`; calls tool with `override=true`. Fails if patch touches anything declared `final`. |
| **Attack attempt**   | Malicious agent tries to redefine `final func auth.verifyToken`. Registration blocked with clear audit log entry.                        |

---

## 5 Tests you should add

* `TestFinalBlocksOverride` — load file A with `final func main()`, then load B with `override func main()`, expect error.
* `TestLoadScriptSandbox` — give root `/tmp/scripts`, attempt `../etc/passwd`, expect error.
* `TestLoadScriptACL` — run under identity not on allow-list, expect “tool not permitted”.
* `TestOverrideFlag` — file without `override` keyword, but tool invoked with `override=true`; should replace **non-final** function.

---

### Why this design is robust

* **`final` is a one-way door** — prevents supply-chain or live-coding accidents from stomping critical code.
* **Tool-based loading** centralises policy: same ACL mechanism you already trust for external calls.
* **Directory sandbox** kills path-traversal tricks.
* **Small cognitive footprint**: only two new concepts (`final`, `system.loadScript`), both intuitive.

You now have hot-reload convenience *and* a solid security boundary—without sprawling rules you’ll forget in six months.

Very nice. So you approve overall?

**Overall, yes—I think this is a solid, internally coherent design.**
It strikes a good balance between *developer ergonomics* (easy hot-reload) and *operational safety* (clear rules, hard stops when you need them). A few final-round observations before you lock it in:

---

### What works well

1. **Single-direction guarantees (`final`)**
   *One‐way doors* make mental models simple. Developers know that declaring something `final` really means “immutable for this session.”

2. **Explicit intent (`override` flag or keyword)**
   Shadowing is never silent. Accidental re-definitions surface immediately, but intentional reloads stay one-liner easy.

3. **Tool-gate for `system.loadScript`**
   Turning the loader into a tool lets you reuse your existing ACL / overlay machinery, so you’re not inventing a parallel permission system.

4. **Filesystem sandbox**
   A canonical `secureJoin` check shuts the door on 90 % of path-traversal nonsense.

---

### Minor caveats / future hooks

| Area                 | Why it might bite later                                                                                                    | Mitigation                                                                                   |
| -------------------- | -------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------- |
| **Concurrency**      | If multiple goroutines call `system.loadScript` simultaneously you’ll need table-level locks or CAS on the session’s maps. | Wrap `registerProc` / `registerHandler` in a mutex or use `sync.Map`.                        |
| **Versioning**       | You can’t load two versions of `utils.math/gcd` side-by-side unless you namespace them manually.                           | Consider an optional `version "1.2.3"` attribute in the future.                              |
| **Session bloat**    | Long-running REPLs may accumulate loads of obsolete procedures.                                                            | Provide `system.unloadScript(path)` or `system.gcProcedures(olderThan time.Duration)` later. |
| **Testing friction** | Tests that need a clean slate must spin up a fresh interpreter each time.                                                  | Cheap if interpreter construction is lightweight; if not, add a `Session.Reset()` helper.    |
| **Auditing**         | It’s easy to forget to persist audit logs.                                                                                 | Decide now: stdout, structured log file, or an append-only NeuroData log node?               |

None of those are fatal; just keep them on the backlog.

---

### Quick approval checklist

* [x] **Security boundary** clear and enforceable
* [x] **DX** (developer experience) remains simple (`final`, `override`, one tool call)
* [x] **Predictability**—no hidden state leaks between loads
* [x] **Extensibility**—easy to graft versioning, unload, or stricter ACL later

If you tick all the above—and your unit tests cover the edge cases we listed—then **ship it**.





btw: I think NO to file-level error handlers -- it is tempting, but the principle is "clarity over brevity"

Got it — I agree that outlawing file-level (top-level) error handlers is the cleanest call.

### Why it’s the right default

| Criterion                                 | Function-scoped handlers only                                          | File-level handlers                                                                                        |
| ----------------------------------------- | ---------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------- |
| **Readability**                           | Anyone reading a function sees exactly what can intercept its errors.  | Reader must hunt elsewhere in the file (or in another file that re-loads) to know what happens on failure. |
| **Local reasoning**                       | Error-handling logic lives next to the code it protects.               | Control flow becomes implicit and scattered.                                                               |
| **Test isolation**                        | Each unit test sets up only the handlers it needs.                     | Hidden global handlers make tests flaky and surprising.                                                    |
| **Hot-reload safety**                     | No persistence headaches; handlers vanish when the call frame unwinds. | You’d need extra rules for overriding/clearing global handlers after reloads.                              |
| **Consistency with `final` / `override`** | Fewer interaction cases to specify.                                    | You’d have to invent `final on error` semantics or risk silent shadowing.                                  |

### Recommended compiler rule

```
on error { … }  # legal only inside a procedure / event handler block
```

*Emit a compile-time error (`ERR_TOP_LEVEL_ERROR_HANDLER`) if encountered at the root of a file.*

### Alternative pattern if someone truly needs “global” trapping

1. **Designate a well-named procedure**:

   ```neuroscript
   final func system.globalError(err error) {
       log.error("Fatal: " + err.message)
       system.exit(1)
   }
   ```

2. **Runtime option**: `SessionOptions.GlobalErrorHook = "system.globalError"`

   The interpreter calls that hook only when an uncaught error bubbles to the top.
   *Still explicit, still discoverable, still guarded by `final`.*

---

**Bottom line:** keep error handling local and explicit; document the “hook” mechanism for the rare case where you truly need a final catch-all. That preserves clarity without sacrificing capability.
