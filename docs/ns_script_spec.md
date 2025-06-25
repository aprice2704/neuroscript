# NeuroScript — Language Specification (v0.4.4)

> **Status:** DRAFT – reflects grammar file `NeuroScript.g4 v0.4.2`
> **Last-updated:** 2025-06-24

---

## 0 · Compatibility guide

| Version | Breaking? | What changed |
|---------|-----------|--------------|
| 0.4.4 | no | **Documentation only:** Added section on line continuation. |
| 0.4.3 | no | **Documentation only:** Added sections on scope, assignment, built-ins, and data access. |
| 0.4.2 | **no** | **Event handlers**, `as` support in **both** `on event` **and** `on error`, `clear_event`, `rep>=` filter, `len()` built-in, new native types `bytes` and `error`. |
| 0.4.1 | no | `must` enhancements (mandatory assignment, map-key/type assertions). |
| 0.4.0 | yes | First public cut, metadata header required. |

---

## 1 · Design goals

* **Executable documentation** – readable by humans & AIs.
* **First-class skill objects** – every `func` is storable, searchable, callable.
* **Strong defensive runtime** – `must …` and `on error` promote “fail-fast”.
* **Extensible** – new blocks/keywords only ever add to the grammar (no breaking rewrites).

---

## 2 · File layout

1. **Metadata header** (`:: key: value`) – mandatory, must come first.
2. **Zero ⁺ procedures** (`func … endfunc`).
3. Blank lines and comments (`#`, `--`, `//`) may appear anywhere.

---

## 3 · Procedure syntax (recap)

```neuroscript
func Name (needs a, b  optional cfg  returns out) means
  :: description: …
  # body
endfunc
```

---

## 4 · Variable Scope, Assignment & Execution

### 4.1 Scope

Variables have **function-level scope**. A variable declared with `set` is visible anywhere within its containing `func` block after the point of declaration.

### 4.2 Assignment (`set`)

Use the `set` keyword to declare a variable and assign it a value.

```neuroscript
set name = "Zadeh"
set parts = tool.Split("a,b,c", ",")
```

### 4.3 Execution (`call`)

Use the `call` keyword to execute a tool function for its side-effects when you do not need its return value.

```neuroscript
call tool.Print("Process complete.")
```

### 4.4 Line Continuation

A single statement can be split across multiple physical lines by placing a backslash (`\`) at the very end of a line. This is most common when defining large map or list literals.

```neuroscript
call tool.FDM.CreateNode("text/v1", { \
  "body": "some long text...", \
  "author": "system" \
})
```

---

## 5 · Control blocks

### 5.1 Event handler

```neuroscript
on event user.registered  named "welcome-mail"  as ev do
  call tool.Mail.Send(ev.email, "Welcome!", "Thanks for joining")
endon
```

### 5.2 Error handler (+ `as err`)

```neuroscript
on error as err do
  emit "fatal: " + err.message
  clear_error
endon
```

### 5.3 Clear event subscription

```neuroscript
clear_event "welcome-mail"
```

---

## 6 · Statement / expression additions

| Feature | Added | Notes |
|---------|-------|-------|
| `clear_event`        | 0.4.2 | See §5.3 |
| `must …`             | 0.4.1 | Boolean assert, mandatory assign, key/type checks |
| `mustbe f(x)`        | 0.4.1 | Convenience for custom validators |
| ACL filter `rep>=N`  | 0.4.2 | Runtime filter, not script syntax |

---

## 7 · Built-in Functions

* `len(expr)`
* `typeof(expr)`

---

## 8 · Native types (complete list)

| Type       | Literals / creators |
|------------|--------------------|
| **string** | `"text"` |
| **int / float** | `42`, `3.14` |
| **bool** | `true`, `false` |
| **nil** | `nil` |
| **list** | `[1, 2]` |
| **map** | `{"k": v}` |
| **timedate** | `tool.Time.Now()` |
| **bytes** | `b"48656c6c6f"` |
| **event** | delivered to `on event` blocks |
| **error** | bound by `on error as err` |
| **fuzzy** | `tool.Fuzzy.Make(0.7)` |

### 8.1 Accessing Collection Data

* **Maps:** `my_map["key"]`
* **Lists:** `my_list[0]`

---

## 9 · Error handling model

* **Operational errors** → tools return `error` objects; use `must` to abort fast.
* **Runtime panics** (failed `must`, index OOB) trigger nearest `on error`.
* `clear_error` swallows a panic within an `on error` block.
* `fail "msg"` raises a panic directly.

---

## 10 · Event-loop semantics

1. Execute next statement.
2. Dispatch one ready event/error handler.
3. Repeat.

---

## 11 · Reserved keywords (v0.4.2)

```
acos and as asin ask atan break bytes call clear clear_error
continue cos do each else emit endfor endfunc endif endon endwhile
error eval event fail false for func fuzzy if in into last len ln log
means must mustbe named needs nil no not on optional or rep return
returns set sin some tan timedate tool true typeof while
```

---

## 12 · Formal grammar reference

See `ns_script_spec_formal.md` – generated automatically from `NeuroScript.g4`.

---

## 13 · Change history

* **2025-06-24** – v0.4.4 — **Docs-only update:** Added section on line continuation (`\`).
* **2025-06-24** – v0.4.3 — **Docs-only update:** Added sections on scope, assignment vs. execution, built-in functions, and collection data access.
* **2025-06-23** – v0.4.2 — event handlers, `on error as`, `clear_event`, `bytes`, `len()`, ACL `rep>=`.
* **2025-06-17** – v0.4.1 — `must` improvements.
* **2025-06-16** – v0.4.0 — first public cut.