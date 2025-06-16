# Value–Wrapping Contract for NeuroScript / FDM
“One wrapper to rule them all”

## 0️⃣ TL;DR
Inside the interpreter every datum is a core.Value. The only exceptions are the implementations of built-in functions, which – just like external tools – consume and return raw Go primitives.
A single adapter (evaluateUserOrBuiltInFunction) unwraps []core.Value → []any, calls the built-in, and re-wraps the result.

---

## 1️⃣ Layer Map & Allowed Types

| Layer | Accepts | Returns | Notes |
|-------|---------|---------|-------|
| Interpreter Core (AST exec, env, stack) | core.Value | core.Value | Pure wrapper world; keeps equality, GC, and future types simple. |
| Built-in Adapter (evaluateUserOrBuiltInFunction) | []core.Value | core.Value | Unwrap → call built-in → wrap back. |
| Built-in Implementation (builtin_*.go) | primitives | primitives | Behaves exactly like a tool impl; zero wrapper noise. |
| Tool Adapter (one per tool) | []core.Value | core.Value | Same pattern as built-in adapter. |
| Validation Layer (tools_validation.go) | primitives | primitives / error | Business rules; never import core. |
| Tool Implementation (tools_*.go) | primitives | primitives / error | Third-party authors write idiomatic Go. |
| Tests | Integration: wrappers · Unit: primitives | mirrors runtime | See § 4 for patterns. |

Visual flow:

Interpreter (wrappers) │ ▼ Adapter ──► Built-in ▸ primitives │ └──► Validator / Tool ▸ primitives ◄──────────────────────────── wraps result 

---

### 1️⃣ bis — Layer Details (text)

* Interpreter Core
* Accepts/Returns: core.Value only
* Rationale: single tagged-union future-proofs Money, Duration, etc.

* Built-in Adapter
* Accepts: []core.Value from the stack
* Action: core.UnwrapSlice, call built-in, core.Wrap result
* Lives in evaluation_main.go.

* Built-in Implementation
* Accepts/Returns: raw primitives (float64, string, …)
* Imports math, time, etc. freely; no wrapper boiler-plate.

* Tool Adapter / Validation / Tool Impl / Tests – unchanged from v1.0.

---

## 2️⃣ Hard Rules

1. No wrapper leaves the interpreter except through an adapter.
2. No primitive enters the interpreter except through an adapter.
3. Validators must never import core/value.go.
4. Any new ValueKind must implement Wrap/Unwrap helpers.
5. Unit tests that hit validators/tools use primitives; integration tests assert on core.Value.
6. Built-in implementations must not accept or return core.Value; the adapter handles conversion.

---

## 3️⃣ Reference Helpers

go // core/value.go func Wrap(x any) (core.Value, error) // primitives → wrapper func Unwrap(v core.Value) (any, error) // wrapper → primitives func UnwrapSlice(vs []core.Value) ([]any, error) 

go // auto-generated adapter skeleton func CallSin(args []core.Value) (core.Value, error) { raw, err := core.UnwrapSlice(args) // []any if err != nil { return nil, err } out := builtinSin(raw) // primitives return core.Wrap(out) // back to wrappers } 

---

## 4️⃣ Testing Patterns

go // integration (through interpreter) res, err := interp.Eval(`sin(0.5)`) // res is core.Value want, _ := core.Wrap(0.4794255386) assert.InDelta(t, want.Float(), res.Float(), 1e-9) // validator unit test (primitive) err := validateList([]any{"x", 1}) require.NoError(t, err) 

---

## 5️⃣ FAQ

| Question | Answer |
|----------|--------|
| Why do built-ins live on the primitive side? | Consistency with tools, reuse of math/stdlib without wrapper noise, and a single conversion choke-point in the adapter. |
| Can validators return wrappers for efficiency? | No. They return primitives; wrapping is the adapter’s job. |
| Can tools inspect core.Value metadata? | Provide a helper inside the adapter, not inside the tool. |
| What if I need streaming outputs? | Stream primitives (e.g. chan any); adapter wraps each item. |

---

### Commit-message template when touching this contract

core/value: maintain wrapper ↔ primitive boundary * No wrappers in validator/tool or built-in impl packages * Added Wrap/Unwrap helpers for <NewKind> * Updated adapters to enforce contract 

> Merge without this template = code-review block 🔒









