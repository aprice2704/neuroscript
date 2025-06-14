 # Value–Wrapping Contract for NeuroScript / FDM ( “One wrapper to rule them all” )

 ## 0️⃣ TL;DR
 Inside the interpreter every datum is a core.Value. Outside—validator & tool code—everything is plain Go primitives.
 The only code that unwraps ↔ wraps lives in a thin Adapter/Bridge layer auto-generated (or hand-written for now).

 ---

 ## 1️⃣ Layer Map & Allowed Types

 | Layer | Accepts | Returns | Notes |
 |-------|-------------|-------------|-------|
 | Interpreter Core (AST exec, env, stack) | core.Value wrappers only | core.Value | Tagged-union; future-proof for Money, Duration, etc. |
 | Adapter / Bridge (one per tool) | []core.Value | core.Value | Sole place that unwraps args → calls validation/tool → wraps result. |
 | Validation (tools_validation.go) | Raw primitives (string, int64, []any, …) | same / error | Pure business rules; never handles wrappers. |
 | Tool Implementation (tools_*.go) | Raw primitives | Raw primitives / error | Third-party authors can write idiomatic Go. |
 | Tests | • Integration path: wrappers via interpreter<br>• Unit path: primitives directly | Mirrors real runtime | See § 4 for examples. |

 Interpreter (wrappers) ──► Adapter (unwrap) ──► Validator & Tool (primitives)
 ◄────────────────────── wraps result ◄──────────────

 ---

 ## 1️⃣ Layer Accepts & Returns (non-table)

 Interpreter Core (AST exec, env, stack)
 • Accepts: core.Value wrappers only
 • Returns: core.Value
 • Why: Single tagged-union keeps equality, GC, and future extensions (Money, Duration…) simple.

 Adapter / Bridge (one stub per tool)
 • Accepts: slice of wrappers []core.Value from the interpreter
 • Returns: a single wrapped result core.Value back to the interpreter
 • Why: Sole choke-point that unwraps args → calls validator/tool → wraps output.

 Validation Layer (tools_validation.go)
 • Accepts: raw Go primitives (string, int64, []any, …)
 • Returns: raw primitives (possibly coerced) or error
 • Why: Keeps business rules free of wrapper boilerplate.

 Tool Implementation (tools_*.go)
 • Accepts: raw primitives (exact types that make sense to tool author)
 • Returns: raw primitives or error
 • Why: Enables idiomatic Go; third-party authors need not import core.

 Tests
 • Integration tests: call the interpreter → supply / assert on wrappers.
 • Unit tests: call validators or tools directly → use primitives.
 • Why: Mirrors real runtime boundaries without extra wrapping noise.

 ## 2️⃣ Hard Rules (enforced via review & lint)
 1. No wrapper leaves the interpreter except through Adapter.
 2. No primitive enters the interpreter except through Adapter.
 3. Validators must never import core/value.go.
 4. Any new ValueKind must implement Wrap / Unwrap helpers.
 5. Unit tests targeting tools/validators use primitives only; integration tests that run scripts assert on core.Value.

 ---

 ## 3️⃣ Reference Helpers

 go  // core/value.go  func Wrap(x any) (core.Value, error) // primitives -> wrapper  func Unwrap(v core.Value) (any, error) // wrapper -> primitives   // convenience  func UnwrapSlice(vs []core.Value) ([]any, error) 

 go  // auto-generated adapter skeleton  func CallListTool(args []core.Value) (core.Value, error) {  raw, err := core.UnwrapSlice(args) // []any  if err != nil { return nil, err }   if err := validateList(raw); err != nil {  return nil, err  }  out := listToolImpl(raw) // primitives  return core.Wrap(out)  } 

 ---

 ## 4️⃣ Testing Patterns

 go  // integration (through interpreter)  res, err := interp.Eval(`list(["a","b"])`) // res is core.Value  want, _ := core.Wrap([]any{"a","b"})  assert.Equal(t, want, res)   // validator unit test (primitive)  err := validateList([]any{"x", 1})  require.NoError(t, err) 

 ---

 ## 5️⃣ FAQ

 | Question | Answer |
 |----------|--------|
 | Can validators return wrappers for efficiency? | No. They return primitives; wrapping is Adapter’s job. |
 | Can tools access core.Value to inspect metadata? | Write a helper inside the adapter, not in the tool. |
 | What if I need streaming outputs? | Stream primitives (e.g. chan any); Adapter converts each item. |

 ---

 ### Commit message template when touching this contract

 text  core/value: maintain wrapper ↔ primitive boundary   * No wrappers in validator/tool packages  * Added Wrap/Unwrap helpers for <NewKind>  * Updated <adapter> to enforce contract 

 > Merge without this template = code review block 🔒