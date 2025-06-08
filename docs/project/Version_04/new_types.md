:: title: Plan for New Type Integration (v0.4)
:: version: 1.0.0
:: status: proposal
:: description: A phased plan to integrate error, event, timedate, and fuzzy types into the NeuroScript core.

---

### Phase 1: Core Value Representation

The first step is to define how these new types exist within the Go runtime of the interpreter.

1.  **Update Type Constants:**
    * Modify `core/type_names.go` to add exported constants for the new type names: `TypeNameError`, `TypeNameEvent`, `TypeNameTimedate`, and `TypeNameFuzzy`.

2.  **Define Value Structs:**
    * In a new file, `core/values.go` (to keep concerns separate from `interpreter.go`), define the Go structs that will represent these values at runtime.
    * `ErrorValue`: Will likely wrap a `map[string]Value` to conform to the standardized error structure.
    * `TimedateValue`: Will wrap Go's `time.Time`.
    * `EventValue`: Will wrap a `map[string]Value` to hold the `name`, `source`, `timestamp`, and `payload` fields.
    * `FuzzyValue`: Will be a struct holding a numeric value and a tolerance/confidence factor.
    * Each of these new structs **must** implement the `core.Value` interface (`Type()`, `String()`, `IsTruthy()`, etc.).

---

### Phase 2: Parser & AST Integration

With the core types defined, we update the parser and AST layers.

1.  **Regenerate Parser:**
    * Run the ANTLR toolchain to regenerate the Go parser and lexer files in `core/generated/` from the `NeuroScript.g4` file we updated. This makes the parser aware of the new keywords.

2.  **Verify `ast.go`:**
    * Review `core/ast.go`. No changes are anticipated here initially, as these complex types won't have a direct literal representation in the script (e.g., you won't write `set t = 2025-06-08T10:00:00Z`). They will be created by tools or runtime events.

---

### Phase 3: Interpreter & Evaluation Logic

@
@@ This phase teaches the interpreter how to understand and operate on the new values.

1.  **Update `typeof` Operator:**
    * In `core/evaluation_logic.go` (or wherever `typeof` is implemented), extend the logic to return the correct type name string for our new `Value` types.

2.  **Update Operators:**
    * In `core/evaluation_comparison.go`, implement comparison logic (`==`, `!=`, `<`, `>`). For instance, `timedate` values should be comparable. `error` and `event` values should likely only support equality checks against `nil`.
    * In `core/evaluation_operators.go`, explicitly disallow non-sensical operations (e.g., arithmetic like `+`, `-`) on the new types by returning a runtime error. This prevents unexpected behavior.

---

### Phase 4: Standard Library Tooling

Users need a way to create and interact with these new types. We will create a small, essential set of tools.

1.  **Propose & Implement Tools:**
    * **Timedate Tool:** A `tool.Time.Now()` to return a new `TimedateValue`.
    * **Error Tool:** An `IsError` helper function to check if a value is an error type, as planned in the roadmap. Also, a `tool.Error.New(code, message)` to construct a standard `ErrorValue`.
    * (Defer `event` and `fuzzy` tools until their core mechanics are further defined).

---

### Phase 5: Comprehensive Testing

Finally, we must validate all changes with thorough testing.

1.  **Add Unit Tests:**
    * For each new `Value` struct, create a `core/values_test.go` to test its methods.
    * Update `core/evaluation_test.go` with test cases for `typeof` and all operator behaviors on the new types.
    * Create test files for the new tools (e.g., `core/tools_time_test.go`) to ensure they function correctly and handle edge cases.

    ### Phase 3½: Fuzzy Logic Semantics  # NEW

> *Why an extra half-phase?*  
> Fuzzy logic is the only type whose behaviour is **not** obvious from Go’s
> built-in semantics.  Getting its operators right early prevents
> contradictory truthiness later.

1.  **Canonical Representation**  
    * `FuzzyValue` **must** store *exactly one* float64 in **[0.0, 1.0]**.  
      A second field called `confidence` sounds attractive but usually models
      *uncertainty of the fuzziness itself*—not needed for core language
      ops.  Keep it single-value and let higher-level logic attach metadata
      if required.

    ```go
    type FuzzyValue struct {
        μ float64 // membership degree 0.0–1.0
    }
    ```

2.  **Truthiness (`IsTruthy`)**  
    * Return `true` if `μ > 0.5`. (0.5 is conventional and aligns with earlier
      roadmap examples.)  
      Document that scripts needing a different threshold should compare
      explicitly (`my_fuzzy > 0.8`).

3.  **Comparison Operators**  
    * `==` and `!=` act on the raw float64 with a **global epsilon**
      (e.g. 1e-6) to avoid tragic rounding errors.  
    * `<`, `>`, `<=`, `>=` likewise act numerically.

4.  **Logical Operators**  # NEW BEHAVIOUR  
    Implement fuzzy versions:

    | Operator | Crisp analogue | Fuzzy definition (`μ` values) |
    |----------|----------------|-------------------------------|
    | `and`    | min            | `min(a.μ, b.μ)` |
    | `or`     | max            | `max(a.μ, b.μ)` |
    | `not`    | negation       | `1 – a.μ` |

    *If either operand is **crisp boolean**, coerce it to `1.0` (true) or
    `0.0` (false) before calculation.*

5.  **Arithmetic Protection**  
    In *Phase 3*’s “evaluation_operators.go” step, explicitly forbid `+`, `-`
    or `*` on fuzzy values with a clear runtime error:  
    *“cannot apply arithmetic operator ‘+’ to fuzzy values – use logical
    operators instead.”*

6.  **Literal Syntax (optional)**  
    A literal isn’t strictly needed—tools can construct fuzzy values—but if
    you want script-level constants, keep it explicit:

    ```neuroscript
    set a = fuzzy(0.8)
    set b = fuzzy(true)   # coerces to 1.0
    ```

    This can be parsed as a built-in function call, so no grammar changes.

7.  **Built-in Helper Functions** (Phase 4 dependency)  
    * `Fuzzy.And(list)` – n-ary min  
    * `Fuzzy.Or(list)`  – n-ary max  
    * `Fuzzy.Distance(a,b)` – `abs(a.μ-b.μ)`  

    These make aggregation in NS scripts easy without custom loops.

8.  **Unit Tests** (Phase 5 extension)  
    * `IsTruthy` boundary cases (`0.5`, `0.51`, `0.49`).  
    * `and/or/not` law sanity (idempotent, commutative, De Morgan
      approximations).

_⇒ With these rules, fuzzy values behave predictably, integrate into the
operator table, and satisfy all examples discussed in the earlier roadmap._

---

### Knock-on Edits to Earlier Phases  # NEW

* **Phase 1** – Add `const TypeNameFuzzy = "fuzzy"` to
  `type_names.go`.
* **Phase 2** – No parser change required if literals are expressed via
  `fuzzy()` built-in.
* **Phase 3** – Extend `typeof` logic so `typeof(fuzzy(0.3))` returns
  `"fuzzy"`.

---

Feel free to fold this section wherever it best fits; all bullet numbers reference your original phase structure.  Once fuzzy logic semantics are locked, the remaining `error`, `timedate`, and `event` work proceeds exactly as in your plan.
