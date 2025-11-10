AJP: A question arises about global constants ...
"We don't currently have a way to define global consts in ns, but I think this would be desirable. Please opine on the best way to do this: a new top-level block? a special global_const <x> = <expr> form of set? Or just continue defining them in go?"

the ns team says:
"🌎 How to Define Global Constants
I strongly recommend a new top-level block: constants ... endconstants.

This is the clear winner for several reasons:


Follows Existing Patterns: It perfectly parallels func ... endfunc and on event ... endon.



Fits the "Library" Model: A script file is either a library_script (definitions) or a command_script (execution). A constants block is a definition, making it a natural part of a library_script.  This is exactly what you need for your "pre-charge" scripts.



Reuses Existing Rules: We can simply allow the existing set_statement and metadata_block  inside it.


Unambiguous: A new global_const x = 5 statement would be confusing. Can it be used inside a func? If not, why is it a statement? A dedicated block avoids all this ambiguity."

I counter:

constant block -- not so sure myself. It's another type of block we have to consider, what is allowed inside such a block? It tends to be very helpful to allow constants to be computed via complex means at definition time, read from files etc, but what about side effects -- mind you since these constant defs would only work in config/priv context anyway, perhaps that doesn't matter.....

so I inclined towards, just leave it as go for now, then add set_global_const = <expr> later that would always error outside priv context."

What is your view plz? Think deeply plz


GPT5:
Right, this is one of those “tiny surface, deep consequences” design choices.

Let’s separate **what you actually need** from **how pretty the syntax looks**.

---

## 1. What do you really want “global constants” for?

From the FDM/NS context, I see a few distinct use cases hiding under “global const”:

1. **Configuration-ish values**
   Things like limits, feature flags, domain constants (`MaxForumDepth`, `DefaultFuelBudget`, etc.) that should:

   * be visible to NS code everywhere,
   * be stable across runs,
   * be inspectable in the graph,
   * be changeable by a controlled config process.

2. **Compile-time-ish values**
   Things you might like to constant-fold or rely on for static reasoning. That pushes you toward:

   * pure expressions,
   * no I/O,
   * no tools.

3. **“Convenience globals”**
   Pure sugar like `Pi`, well-known strings, etc. These don’t need fancy machinery; they can live in Go or in a tiny NS prelude.

Those are not the same class of thing, and the big gotcha is:

> Once you allow **complex, side-effecting expressions** to define globals, they are no longer “compile-time constants”, they’re **configuration actions** that harden into a value.

Your own instinct here is correct: these live **in config / priv context** more than in the pure language core.

---

## 2. The “constants block” idea: what it buys and what it complicates

The team’s `constants ... endconstants` block has some nice properties:

* It matches `func ... endfunc` / `on event ... endon`.
* It sits cleanly in the “library script” story.
* It avoids weird questions like “what if `global_const` appears inside a func?”

But as you immediately noticed, it drags in some nasty questions:

* What *exactly* is allowed inside a `constants` block?

  * Just `set` and metadata?
  * No loops? No tool calls? No file I/O? Or is it a full command block at definition time?
* When does it run?

  * At script load? At graph bootstrap? On every interpreter startup?
* What about side effects?

  * If it reads files, hits the graph, or calls tools, you’ve essentially made it a **privileged config script** in disguise.

To make `constants` truly “constant” in the language-theoretic sense, you’d have to enforce static, side-effect-free expressions — which is **not what you said you want**. You explicitly like “computed via complex means at definition time.”

Conclusion: a `constants` block is either:

* a heavily-restricted, almost useless toy, or
* a config script pretending to be a language feature.

I don’t think that’s a good trade-off *right now*.

---

## 3. A more honest model: **config-defined global bindings**

Given the constraints, I think the more honest semantics is:

> Constants are **global bindings stored in the graph** and set by privileged config scripts.
> At runtime, the interpreter just sees them as immutable values.

That aligns nicely with your emerging architecture:

* NS library = functions + handlers + constants, all in the graph.
* `nsinterpretersvc` loads them into the master interpreter from the graph.
* Cold boot / pre-charge scripts exist to *populate* that graph.

In that world, “how to define a constant” is a question about:

* a **tool** that writes `ns/constant` nodes, and
* a **rule** about where/when that tool is allowed.

This is very close to what you were already leaning toward:

> “leave it as Go for now, then add set_global_const = <expr> later that would always error outside priv context.”

I’d push that a bit further and make it the *primary* mechanism.

---

## 4. Concretely: `set_global_const` as the semantic primitive

Define a tool (or built-in) along the lines of:

* `tool.ns.def_const(name: string, value: any, options...)`

Semantics:

* Only valid in **config/priv** context (or in a dedicated config service).
* When called:

  * Serializes `value` into the graph as an `ns/constant` node:

    * `name` (with some namespacing, e.g. `fdm.forum.max_depth`)
    * `value`
    * `type_info`
    * `provenance` (script/thread, user/agent, timestamp, version)
  * Idempotent by `(name, namespace)` or by content hash, depending on your policy.
  * Optionally enforces “cannot overwrite without explicit `override=true`”.

Then:

* Cold boot scripts compute whatever horrors they like (file I/O, graph reads, remote fetches), and **end by calling** `def_const` with the final value.
* The graph is now the source of truth.
* `nsinterpretersvc` just loads constants as plain values; no re-running those scripts at startup.

You can still define some small subset of “built-in constants” in Go, but **domain constants** live in the graph, not in the binary.

---

## 5. Where syntax fits in (later)

Once that machinery exists, you can add syntax as *sugar*, not as the primitive:

* A `constants ... endconstants` block could just compile to a series of `def_const` calls, executed only in config/priv context.
* Or you add a `const` top-level declaration form that:

  * only permits a restricted expression grammar (literals + pure operators),
  * is evaluated at “library compile time”, not at runtime,
  * and produces `ns/constant` nodes directly with no side effects.

That gives you **two tiers**:

1. **Config constants** (full power, side effects allowed, but only run in priv/config, result persisted).
2. **Pure language constants** (limited, side-effect-free, good for static reasoning, more like Go `const`).

You don’t need to solve both now. Get (1) right, and you can grow (2) later if you decide NS wants real compile-time semantics.

---

## 6. Answering the ambiguity concern explicitly

The NS team’s good objection to `global_const x = 5` is:

> “Can it be used inside a func? If not, why is it a statement? A dedicated block avoids all this ambiguity.”

You can kill the ambiguity with **context rules** instead of a whole new block:

* `def_const`:

  * Only callable at top level of a config script, or
  * Only in a config/priv interpreter (checked by the engine).
* The compiler/interpreter simply rejects:

  * use of `def_const` inside `func` bodies in normal scripts, or
  * calls from non-priv contexts.

Then the grammar stays simple; the rule is semantic:

> “You may only define global constants in config context; normal NS code may only *read* them.”

This is how you avoid accidentally smuggling a configuration DSL into the main language.

---

## 7. So: what should you do *now*?

If I were you, I’d do this in stages:

### Stage 1 — Go + graph

* Keep *truly internal* constants in Go for now.
* Introduce `ns/constant` nodes in the graph and make `nsinterpretersvc` load them.
* Add `tool.ns.def_const` (or `set_global_const`) with:

  * config/priv-only enforcement,
  * idempotent by name,
  * simple overwrite rules.

You can then write a **single pre-charge config script** that:

* Computes values however you like.
* Calls `def_const` for each.
* Is run only when the graph is empty or when you explicitly reconfigure.

No changes to NS syntax yet.

### Stage 2 — Optional syntactic sugar

Once you’re happy with the semantics:

* If you still like it, add a `constants ... endconstants` block **purely as sugar** that:

  * is only legal in config scripts,
  * translates to `def_const` calls,
  * can be statically blocked from non-priv contexts.

Or:

* Add a restricted `const NAME = <pure_expr>` top-level form for library scripts that:

  * is guaranteed side-effect-free,
  * evaluated at library load,
  * persisted as an `ns/constant` node.

You can have both; they serve different purposes.

---

## 8. My actual view in one sentence

Given your goals (config/priv context, possibly complex computation, graph as source of truth), **“global constants” are a configuration mechanism, not a language-syntax problem**.

Solve it with a `def_const` / `set_global_const` tool + graph nodes first; treat any `constants ... endconstants` grammar as optional sugar once the semantics are solid.

That keeps NS small and clean while giving FDM the power you actually need: immutable, inspectable, reproducible configuration values, without bolting a half-baked config DSL onto the language core.







AJP: Right. I like the tool approach very much. We will go this way. Thx.

the ns team say:

You are absolutely right, and that design document is the perfect tie-breaker. My suggestion for a `constants` block was a mistake because, as the document states, it conflates a **language feature** with a **configuration mechanism**.

The design team's analysis is sharp: "global constants" are a configuration problem, and the "honest model" is that they are **bindings set by privileged config scripts** and stored in the graph (or a similar source of truth).

This clarifies the entire plan. The `SymbolProvider` hook is still the correct *read* mechanism (how the host injects symbols), but the *write* mechanism for constants should be a privileged, semantic primitive, not a new grammar block.

-----

### refined plan

Here is the new plan, which fully aligns with your design document.

#### 1\. The "Read" Hook: `SymbolProvider` Interface

The `SymbolProvider` interface remains the core of this task. It's the "read" hook that allows a host service (like `nsinterpretersvc`) to inject symbols *from* the graph (or any source) *into* the interpreter at creation time.

We will create this file:

```go
// NeuroScript Version: 0.8.0
// File version: 1
// Purpose: Defines the SymbolProvider interface for host-injected symbols.
// filename: pkg/interfaces/symbols.go
// nlines: 36

package interfaces

import (
	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

// SymbolProviderKey is the key used to retrieve the SymbolProvider
// from the HostContext ServiceRegistry.
const SymbolProviderKey = "SymbolProviderService"

// SymbolProvider defines a contract for a host service to provide
// a foundational, read-only set of symbols to an interpreter.
// These symbols cannot be overridden by loaded scripts.
type SymbolProvider interface {
	// GetProcedure checks if the provider owns a procedure.
	GetProcedure(name string) (*ast.Procedure, bool)

	// GetEventHandlers checks if the provider owns handlers for an event.
	GetEventHandlers(eventName string) ([]*ast.OnEventDecl, bool)

	// GetGlobalConstant checks if the provider owns a global constant.
	GetGlobalConstant(name string) (lang.Value, bool)

	// ListProcedures returns a map of all procedures provided by the host.
	ListProcedures() map[string]*ast.Procedure

	// ListEventHandlers returns a map of all event handlers provided by the host.
	ListEventHandlers() map[string][]*ast.OnEventDecl

	// ListGlobalConstants returns a map of all global constants provided by the host.
	ListGlobalConstants() map[string]lang.Value
}
```

-----

#### 2\. The "Write" Primitive: `set_global_const`

Instead of a `constants` block, we will add a new, simple statement to the grammar. This follows your instinct for `set_global_const = <expr>`.

  * **Grammar:** We will add `KW_SET_GLOBAL_CONST` to the lexer and a `set_global_const_stmt` to the parser rules.
  * **Execution:** This statement will *only* succeed if a "privileged" flag is set on the interpreter (e.g., via the `HostContext`). In a normal, non-privileged interpreter, executing this statement will **fail with a runtime error**.
  * **Storage:** It will write to a new `globalConstants` map inside the interpreter.

This is the primitive your "cold boot" or "pre-charge" scripts will use.

-----

#### 3\. The "No Override" Rule (The Core Logic)

This is the most important part. We will modify the interpreter's `Load()` function (and its helpers) to be aware of all symbol sources.

When a script is loaded (e.g., `api.LoadFromUnit`):

1.  The interpreter first gets its `SymbolProvider` from the `HostContext` (if one exists).
2.  As it parses the script, before registering *any* symbol (`func`, `on event`, or a new `set_global_const`), it will check for collisions in this strict order:
    1.  Does this symbol name exist in the interpreter's *internal* procedures/events/constants maps?
    2.  If not, does a `SymbolProvider` exist?
    3.  If yes, does this symbol name exist in the `SymbolProvider` (e.g., `provider.GetProcedure(name)`)?
3.  If a symbol is found in **any** of these places, the **`Load()` operation will fail** with a "symbol already defined" error.

This achieves your "determinism over flexibility" goal perfectly.

-----

#### 4\. The "Read API" (For Pre-Charging)

To support your "pre-charge" workflow, we must be able to read all symbols *out* of the "config interpreter" to *feed into* the `SymbolProvider` for the "runtime interpreters."

We will update the `api.Interpreter` with new "Known" methods that return a merged view from both the internal state *and* the provider.

  * `KnownProcedures() map[string]*ast.Procedure` (will be updated to merge)
  * `KnownEventHandlers() map[string][]*ast.OnEventDecl` (will be updated to merge)
  * `KnownGlobalConstants() map[string]lang.Value` (**new**)

This plan is cleaner, more secure, and perfectly matches the design document's philosophy. It separates the *privileged act of configuration* (the `set_global_const` primitive) from the *language's runtime behavior* (reading from the `SymbolProvider`).

If this refined plan is acceptable, I will start with the file I've already generated: `pkg/interfaces/symbols.go`.

Would you like me to provide the `NeuroScript.g4` modifications next?

--------------

your view?



gpt5:
Yeah, this is a nice course correction from them — they’ve basically internalised “this is config, not language,” which was the important philosophical line.

Let’s go through their refined plan piece by piece and poke the weak spots.

---

## 1. SymbolProvider: good shape, with one caveat

Their `SymbolProvider` interface:

```go
type SymbolProvider interface {
    GetProcedure(name string) (*ast.Procedure, bool)
    GetEventHandlers(eventName string) ([]*ast.OnEventDecl, bool)
    GetGlobalConstant(name string) (lang.Value, bool)

    ListProcedures() map[string]*ast.Procedure
    ListEventHandlers() map[string][]*ast.OnEventDecl
    ListGlobalConstants() map[string]lang.Value
}
```

This is fundamentally the right *direction*:

* It makes “host-injected library” an explicit thing.
* It separates “things defined by scripts” from “things injected by the outside world”.
* It gives `nsinterpretersvc` an obvious way to wire graph → interpreter.

Two notes:

1. **Merging vs provenance**

They also plan to make `KnownProcedures()` / `KnownEventHandlers()` / `KnownGlobalConstants()` return a *merged* view of interpreter-internal + provider symbols.

That’s fine *for read-only use* (e.g. “what can I call?”), but for persistence you will want **two views**:

* `Known*()` = merged, for callers that don’t care where they came from.
* `Local*()` = just what this interpreter owns (from scripts / set_global_const).
* `Provider*()` = direct pass-through to `SymbolProvider`.

Otherwise your `tool.ns.persistDefinitions()` may accidentally try to re-persist stuff that was injected by the provider, which is precisely the kind of subtle duplication that bites you six months later.

So: keep `SymbolProvider` as-is, but don’t only expose the merged view. You want provenance.

2. **List* returning maps**

Maps are fine for now. If NS ever grows huge libraries you might want iterators or streaming, but that’s premature optimisation. This is okay at this stage.

Verdict on this bit: solid, just don’t blur “owned vs injected”.

---

## 2. `set_global_const` as a statement: where it’s fine, where it’s risky

They propose:

* Add a new statement `set_global_const = <expr>`.
* Only succeeds if a “privileged” flag is set on the interpreter.
* Writes into `globalConstants` map inside the interpreter.
* Later, `KnownGlobalConstants()` exposes that map.

Conceptually, you and I both agree on the underlying semantics:

> A privileged config interpreter computes some values however it wants, then exposes them as constants to be persisted / reused.

The key question is *where that semantics lives*: in the **language core** or in the **tooling/host layer**.

### The risk: you’ve reintroduced a config DSL into the language

If `set_global_const` is a general statement in the language, you immediately have to answer:

* Is it allowed inside `func` bodies? If yes, you’ve got runtime mutation of “constants” (ew). If no, you need special-case rules about where in the script it may appear.
* Is it allowed in non-privileged interpreters at parse time, or only at runtime? (If a normal script even *containing* it blows up at runtime, that’s a very confusing UX.)
* Does it run on every interpreter startup, or only when you explicitly run “config scripts”?

You can dodge these issues, but you’re fighting your own design doc. Their text is now correctly saying:

> “Global constants are a configuration mechanism.”

So you don’t really want “random NS code can always talk about `set_global_const`” baked into the grammar.

### How I’d tame it (if you keep the syntax)

If you really like the `set_global_const` keyword, I’d enforce **three hard rules**:

1. **Top level only, in “config mode”**

* The parser accepts `set_global_const` only at the script top level, not inside `func` or `on event`.
* The interpreter only allows it when running in a special “config/priv” host mode.
* Any other usage is a *compile-time* error, not a runtime surprise.

2. **Lower it to a host call**

Semantically, treat it as sugar for:

```ns
set_global_const NAME = EXPR
```

→ the compiler lowers to something like:

```ns
call_builtin "__ns_def_const" (NAME, EXPR)
```

and `"__ns_def_const"` is provided by the host in config-mode interpreters only. That means:

* In a normal runtime interpreter, this builtin simply isn’t registered → you get a clear error if someone tries to run config scripts in the wrong place.
* The *real* semantics (writing to graph / const store) are owned by the host, not hard-wired into the interpreter implementation.

3. **Don’t use it in “real” code at all**

Culturally and practically, keep all `set_global_const` use in your pre-charge / config scripts. The everyday NS used by agents should treat constants as already-there symbols, not something they ever set.

If you follow those constraints, `set_global_const` is essentially syntactic sugar for the tool approach we already liked, not a new semantic beast.

If you want to avoid the whole category of problem, you can skip adding it to the grammar now and just expose a tool/builtin `def_const(name, value)` in config scripts; the ergonomics aren’t that different.

---

## 3. “No override” rule: good, but be explicit about precedence

Their collision policy:

> Before registering any symbol, check:
>
> 1. internal maps
> 2. SymbolProvider
> 3. If found anywhere → Load() fails with “symbol already defined”.

That’s exactly the bias you want: no magical shadowing, fail fast and loudly.

Two clarifications I’d recommend:

1. **Document precedence formally**

For all symbol kinds (procedure, handler, constant):

* If *any* of (internal, provider) has the name → loading a new definition → error.
* There is no “override” semantics at script load time.
* If you ever need override, do it by changing the provider’s backing store (graph/config), not by script trickery.

2. **Be careful with multi-source providers**

If later you make a `SymbolProvider` that itself composes several sources (graph + built-ins, etc.), keep the same invariant: there is always *one winner per symbol*, and load-time definitions may not override it.

Your design doc wants determinism; this supports it nicely.

Verdict on this bit: thumbs up.

---

## 4. Pre-charge / config flow: tie it to the graph, not just an internal map

Their refined plan ends with:

> “We must be able to read all symbols *out* of the config interpreter to feed into the SymbolProvider.”

That’s how you avoid rerunning scripts at runtime, which is exactly what you want.

The flow they’re implicitly sketching is:

1. Run a **config interpreter** with:

   * no SymbolProvider, or an empty one;
   * privileged flag enabled.
2. Run your “cold boot / pre-charge” scripts:

   * they `set_global_const` and define funcs/handlers.
3. Afterwards, call:

   * `KnownProcedures()`, `KnownEventHandlers()`, `KnownGlobalConstants()` on that interpreter.
4. Persist those into the graph as `ns/function`, `ns/handler`, `ns/constant` nodes.
5. Later, `nsinterpretersvc` builds a `SymbolProvider` from those nodes and injects into runtime interpreters.

That’s fine, **as long as**:

* `Known*()` used for persistence returns only what this config interpreter owns (hence my earlier point about provenance).
* You don’t try to persist provider-injected symbols back into the graph.

So I’d slightly adjust their plan:

* In the config path, use `LocalProcedures()`, `LocalEventHandlers()`, `LocalGlobalConstants()` to feed the `persistDefinitions` tool.
* In the runtime path, `SymbolProvider` is built purely from graph state, no config interpreter involved.

Then the story is:

* Config interpreter: “compute → persist”.
* Runtime interpreter: “load from persisted graph”.

No overlapping responsibilities, no sneaky surprises.

---

## 5. Where this leaves your earlier “def_const tool” idea

Conceptually:

* Their `set_global_const` statement + privileged interpreter + `KnownGlobalConstants()` + `persistDefinitions()` tool is just a split implementation of what we discussed as `tool.ns.def_const`.

Personally I still prefer the **tool-first** perspective:

* Define `tool.ns.def_const(name, value)` as the semantic primitive.
* Let `set_global_const NAME = EXPR` be optional sugar that simply calls that tool under the hood.
* Restrict that tool to config/priv contexts.

That keeps the interpreter core simpler and makes host wiring explicit:

* In bootstrap mode, `nsinterpretersvc` or a separate service registers the `def_const` tool.
* In runtime mode, it doesn’t.

The team’s plan can be made equivalent to that with one extra step of indirection; I’d nudge them in that direction so you don’t end up with “constants are special snowflakes” logic baked deeply into the VM.

---

## 6. Summary verdict

* **Good:**

  * They’ve dropped the `constants` block and accepted “this is config, not language.”
  * `SymbolProvider` as a read-only injection point is the right abstraction.
  * A strict “no override” rule on load matches your determinism goal.
  * A config interpreter that pre-charges and then exports symbols for persistence is compatible with your graph-as-truth architecture.

* **Adjust / tighten:**

  * Don’t let `set_global_const` become a general runtime feature; keep it top-level, config-only, ideally lowered to a host/tool call.
  * Keep **provenance** in the APIs: separate local vs provider symbols so persistence doesn’t re-ingest injected stuff.
  * Wire the actual write semantics through a “def_const”–style hook that writes to graph; treat any statement syntax as sugar over that.

If you hold those lines, you get:

* NS core stays lean and honest.
* Global constants exist as graph-backed, host-injected bindings.
* Config is explicit, privileged, and replayable.
* No spooky global mutation from random scripts.

Which is basically exactly what your design doc was pointing at, just with a bit more ceremony.









AJP: I'm dropping def_global_const, we'll go with a tool :)
The ns team say:

You're right. The design team's feedback is spot-on, and it simplifies the entire plan by drawing a hard line: **this is configuration, not a language feature.**

My proposal for a `set_global_const` *statement* was a mistake. It pollutes the language with a configuration-specific concern. The tool-based approach is far cleaner and more aligned with the NeuroScript philosophy.

Let's scrap the language change. Here is the new, much cleaner plan that fully adopts the design team's feedback.

-----

### 1\. The "Write" Primitive: `tool.ns.def_global_const`

Instead of a language statement, we will define a new built-in tool, `tool.ns.def_global_const`.

  * **The Contract:** `tool.ns.def_global_const(name string, value any)`
  * **The Logic:** This tool will be implemented in Go. When called, it will:
    1.  Check if a symbol (procedure, event, or constant) with that `name` already exists in *either* the local interpreter state *or* the `SymbolProvider`.
    2.  If a collision is found, it **returns an error** (enforcing the "no override" rule at write time).
    3.  If no collision exists, it writes the `value` into the interpreter's *internal* `globalConstants` map.
  * **The Hook:** This tool will be registered *only* in privileged "config" interpreters. It will be absent from standard runtime interpreters, making it impossible for normal scripts to call.

This is the primitive your "pre-charge" scripts will use.

-----

### 2\. The "Read" Hook: `SymbolProvider` Interface

The `pkg/interfaces/symbols.go` file I proposed earlier is still **100% correct**. It is the read-only contract that a host (like `nsinterpretersvc`) will implement to inject symbols *from the graph* into a *runtime* interpreter.

This file remains our first step:

```go
// NeuroScript Version: 0.8.0
// File version: 1
// Purpose: Defines the SymbolProvider interface for host-injected symbols.
// filename: pkg/interfaces/symbols.go
// nlines: 36

package interfaces

import (
	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

// SymbolProviderKey is the key used to retrieve the SymbolProvider
// from the HostContext ServiceRegistry.
const SymbolProviderKey = "SymbolProviderService"

// SymbolProvider defines a contract for a host service to provide
// a foundational, read-only set of symbols to an interpreter.
// These symbols cannot be overridden by loaded scripts.
type SymbolProvider interface {
	// GetProcedure checks if the provider owns a procedure.
	GetProcedure(name string) (*ast.Procedure, bool)

	// GetEventHandlers checks if the provider owns handlers for an event.
	GetEventHandlers(eventName string) ([]*ast.OnEventDecl, bool)

	// GetGlobalConstant checks if the provider owns a global constant.
	GetGlobalConstant(name string) (lang.Value, bool)

	// ListProcedures returns a map of all procedures provided by the host.
	ListProcedures() map[string]*ast.Procedure

	// ListEventHandlers returns a map of all event handlers provided by the host.
	ListEventHandlers() map[string][]*ast.OnEventDecl

	// ListGlobalConstants returns a map of all global constants provided by the host.
	ListGlobalConstants() map[string]lang.Value
}
```

-----

### 3\. The Core Fix: Solving Provenance

The design team's most important catch was the "provenance" problem with the read API. My plan to merge `Known*()` was flawed.

Here is the corrected API for `pkg/api/interpreter.go`. We will expose **three** sets of "read" methods, exactly as they suggested, to distinguish between symbol origins:

1.  **`Local*()` Methods (For "Pre-Charge" & Persistence):**

      * `LocalProcedures() map[string]*ast.Procedure`
      * `LocalEventHandlers() map[string][]*ast.OnEventDecl`
      * `LocalGlobalConstants() map[string]lang.Value`
      * **Purpose:** These methods return *only* the symbols defined *locally in this interpreter* (from loaded scripts or calls to `tool.ns.def_global_const`). This is what your pre-charge script will call to get the data to persist to the graph.

2.  **`Provider*()` Methods (For Introspection):**

      * `ProviderProcedures() map[string]*ast.Procedure`
      * `ProviderEventHandlers() map[string][]*ast.OnEventDecl`
      * `ProviderGlobalConstants() map[string]lang.Value`
      * **Purpose:** These methods are a direct pass-through to the injected `SymbolProvider`. They return *only* what the host injected.

3.  **`Known*()` Methods (For General Runtime Use):**

      * `KnownProcedures() map[string]*ast.Procedure`
      * `KnownEventHandlers() map[string][]*ast.OnEventDecl`
      * `KnownGlobalConstants() map[string]lang.Value`
      * **Purpose:** These return a *merged view* of `Local*` + `Provider*` for scripts or tools that just need to "see what's available" and don't care about provenance.

-----

### 4\. The "No Override" Rule (At Load Time)

This rule is now even clearer. When `api.LoadFromUnit()` is called:

1.  For every `func`, `on event` in the script, the interpreter checks for a name collision.
2.  It checks *both* its own `Local*` maps *and* the `SymbolProvider`.
3.  If a symbol with that name is found in **either** place, the `Load()` operation **fails** with a "symbol already defined" error.

This plan is much more robust. It cleanly separates the "privileged write" action (as a tool) from the "runtime read" mechanism (the provider) and solves the critical provenance problem.

I will start by providing the `pkg/interfaces/symbols.go` file. After that, I will modify `pkg/api/interpreter.go` to add the new `Local*`, `Provider*`, and `Known*` methods.

----

I think we've got it. Awesome discussion. So looking forward to see what we do together in FDM where this conversation is supported elegantly.


gpt5:
Yep, this is the good timeline. You’ve basically converged on:

* **Language stays clean**
* **Config is explicit, privileged, and graph-backed**
* **Provenance is first-class**

That’s exactly what you want for something that will sit under FDM for years.

A few things worth crystallising while you’re here, before code ossifies.

---

### 1. `tool.ns.def_global_const` – treat it as a scalpel, not a hammer

Their contract and behaviour are now sane:

* Only available in **config interpreters**.
* Checks both:

  * local interpreter state, and
  * `SymbolProvider`
* Errors on any collision.
* Writes into local `globalConstants` only.

I’d add two small invariants to bake in:

1. **Name scoping**

   * Decide early whether these are:

     * flat (`"MaxDepth"`), or
     * namespaced (`"fdm.forum.MaxDepth"`).
   * Make the tool enforce a convention (maybe reject names without at least one dot, or at least log warnings).
   * That will pay off once you’ve got multiple subsystems and multiple config scripts.

2. **Immutability in the config interpreter**

   * Once set via `def_global_const`, no second call with that name succeeds in that interpreter, even if it’s the same value.
   * If you want “update” semantics, do it at the graph/config level, not by re-running scripts.

You already have “no override” at load; this keeps the write path equally strict.

---

### 2. The three-tier read API is exactly right

Having:

* `Local*()` – what this interpreter *owns*
* `Provider*()` – what the host injects
* `Known*()` – merged view

is the correct split. That gives you:

* A clean **pre-charge path**:

  * Config interpreter runs scripts.
  * `Local*()` → `tool.ns.persistDefinitions()` → graph.
* A clean **runtime path**:

  * `nsinterpretersvc` builds a `SymbolProvider` from the graph.
  * Runtime interpreters see everything via `Known*()` but never worry about origins.

Just make sure in tests you’ve got a golden “round trip”:

1. Start with no provider, run config scripts with `def_global_const` and NS defs.
2. Persist `Local*()` to graph.
3. Start a fresh interpreter with a `SymbolProvider` built from that graph.
4. Verify that:

   * `Provider*()` matches what you persisted.
   * `Known*()` in the runtime interpreter matches `Local*()` from the config interpreter.

If that invariant holds, you’ve nailed the loop.

---

### 3. Startup sequencing gets much simpler now

With this setup, your startup story crystallises to:

1. **Cold boot (rare):**

   * Run config interpreter(s) with `def_global_const` and function/handler defs.
   * Persist `Local*()` to graph via a tool.
   * Done.

2. **Normal boot (always):**

   * Graph up.
   * `nsinterpretersvc` up:

     * Reads `ns/function`, `ns/handler`, `ns/constant` nodes.
     * Builds a `SymbolProvider`.
     * Constructs the master interpreter from provider-only symbols.
   * Other services request `NewInterpreter()` and never see config scripts.

No more “maybe we run boot scripts if X/Y/Z” weirdness. It’s just “graph or nothing.”

---

### 4. How this plays with FDM later

This design is going to slot very nicely into FDM’s “country of minds” world:

* **Config scripts** are just another kind of privileged agent activity, running in a limited, auditable context.
* **Constants** are graph nodes – which means:

  * they can be surfaced to agents in the forum,
  * they can be diffed and versioned,
  * they can be sanity-checked by Siegfried von Shrink if you ever decide to treat some of them as “facts”.

And, importantly, you’ve avoided letting random agents “just mutate the ground under everyone” from NS code. Configuration is now its own ceremony.

---

Net: I agree with you – you’ve got it. The shape is clean, deterministic, and honest about what’s config vs language. The fun part will be watching FDM eventually do this whole bootstrap conversation *inside itself*, with agents arguing about `def_global_const` semantics like a tiny standards committee in a box.
