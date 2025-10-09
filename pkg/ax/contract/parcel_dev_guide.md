 # Runner Parcel + Clone Discipline — Dev Implementation Guide (v1.1)
 Date: 2025-10-08
 Audience: NS/AX engine devs, Zadeh host integrators
 Goal: Eliminate context-loss during cloning, unify API/engine semantics, and stop “globals” from leaking into locals by standardizing on a **Runner Parcel** carried by both AX and the internal interpreter. Keep changes surgical.

 ---
 ## 0) TL;DR — What to implement
 - Create three tiny contracts in a neutral package (`ax/contract`): `RunnerParcel`, `ParcelProvider`, `SharedCatalogs`.
 - Make **both** the AX Runner and the internal Interpreter carry a `RunnerParcel` pointer (copy-by-ref across clones).
 - In **every** clone path (API-level and engine-internal): copy the parcel pointer, reset exec state, reuse SharedCatalogs by reference.
 - Remove any code that copies “globals” into local variables; expose them only via `parcel.Globals()` (read-only).
 - Route identity/AEIOU/policy/logger lookups only via the parcel; route funcs/events/accounts/models/tools/capsules via SharedCatalogs.

 ---
 ## 1) Concepts — the three-bucket model
 1) **Shared Catalogs (by reference):** function defs, event handlers, accounts, agent models, tools, capsules. These are long-lived registries behind a single façade; clones reuse the same pointer.
 2) **Runner Parcel (by reference):** your execution “who/where/rights” bundle: AEIOU envelope, Identity, Policy, Logger, and a **read-only** Globals view. The parcel is the single source of truth for runtime context.
 3) **Exec State (fresh per clone):** variable stack, PC, temporaries. Always new per clone; never shared.

 ---
 ## 2) Contracts (place in `ax/contract`; keep it light to avoid import cycles)
 type RunnerParcel interface {
     AEIOU() AEIOUContext
     Identity() Identity
     Logger() Logger
     Policy() Policy
     Globals() map[string]any       // MUST be read-only to callers
     Fork(mut func(*ParcelMut)) RunnerParcel  // API clones may fork; engine clones do not
 }
 
 type ParcelMut struct {
     AEIOU  *AEIOUContext
     ID     *Identity
     Logger *Logger
     Policy *Policy
     // No direct Globals mutation; set at construction. If needed later, add controlled replacement.
 }
 
 type ParcelProvider interface {
     GetParcel() RunnerParcel
     SetParcel(RunnerParcel)
 }
 
 type SharedCatalogs interface {
     Funcs() FuncCatalog          // lookup-only
     Events() EventCatalog        // lookup-only
     Accounts() AccountStore
     AgentModels() AgentModelStore
     Tools() ToolRegistry
     Capsules() CapsuleRegistry
 }

 ---
 ## 3) Concrete parcel (simple, predictable)
 - Implement `RunnerParcel` as a `parcel` struct: holds pointers to AEIOU, Identity, Logger, Policy, and a private map for Globals.
 - `Globals()` must return a **defensive read-only** view (copy on read or a proxy that forbids writes).
 - `Fork()` shallow-copies fields, applies `ParcelMut`, does **not** mutate the parent.
 
 // Sketch (keep real impl ~50 lines; no heavy deps)
 type parcel struct {
     aeiou   *AEIOUContext
     id      *Identity
     log     *Logger
     pol     *Policy
     globals map[string]any // private backing store
 }
 func (p *parcel) AEIOU() AEIOUContext     { return *p.aeiou }
 func (p *parcel) Identity() Identity      { return *p.id }
 func (p *parcel) Logger() Logger          { return *p.log }
 func (p *parcel) Policy() Policy          { return *p.pol }
 func (p *parcel) Globals() map[string]any { return readOnlyCopy(p.globals) }
 func (p *parcel) Fork(mut func(*ParcelMut)) RunnerParcel {
     cp := *p
     m := ParcelMut{AEIOU: cp.aeiou, ID: cp.id, Logger: cp.log, Policy: cp.pol}
     mut(&m)
     cp.aeiou, cp.id, cp.log, cp.pol = m.AEIOU, m.ID, m.Logger, m.Policy
     return &cp
 }

 ---
 ## 4) Shared Catalogs façade
 - Implement `sharedCatalogs` wrapping existing registries/stores; just getters.
 - Pass **one** `SharedCatalogs` pointer into runners/interpreters; stop threading N pointers.
 
 type sharedCatalogs struct {
     funcs  FuncCatalog
     events EventCatalog
     accts  AccountStore
     models AgentModelStore
     tools  ToolRegistry
     caps   CapsuleRegistry
 }
 // Implement SharedCatalogs with trivial getters.

 ---
 ## 5) AX layer changes (factory/runner)
 **Factory**
 - Accept `WithParcel(p RunnerParcel)` or build one from existing AEIOU/Identity/Logger/Policy if omitted: `ParcelFromAEIOU(...)`.
 - Construct or inject `SharedCatalogs`; avoid scattering stores/registries.
 
 **Runner**
 - Add fields: `parcel RunnerParcel`, `catalogs SharedCatalogs`, and exec state.
 - Implement `ParcelProvider` + a `Catalogs() SharedCatalogs` getter.
 - In `Runner.Clone(ctx, opts)`:
   * Reuse `catalogs` by reference.
   * Set the **same parcel pointer** (`child.SetParcel(parent.GetParcel())`) by default; allow `opts.ForkParcel` to call `Fork(...)` if the caller wants to tweak AEIOU/policy/caps.
   * Always create a fresh exec state (locals/stack).
   * Do **not** materialize parcel globals into locals.

 ---
 ## 6) Engine-internal interpreter changes (this fixes the “double clone” bug)
 **Interpreter struct**
 - Add a private field: `parcel RunnerParcel`.
 - Add private accessors: `getParcel() RunnerParcel`, `setParcel(RunnerParcel)`.
 - Ensure tool/identity/AEIOU/policy/logger lookups go via `getParcel()`.
 
 **Internal clone sites**
 - Identify all engine-level clone paths (e.g., pre-proc exec, pre-event handler).
 - In each path, after shallow-clone internals:
   * `child.setParcel(parent.getParcel())`   // pointer copy
   * `child.runtime = child`                 // as you already do
   * `child.resetExecState()`                // new locals/stack
 - Remove any code that “rebuilds” identity/AEIOU/policy/logger. The parcel is the single source of truth.

 ---
 ## 7) “Globals” policy (no more copying into locals)
 - Remove code that copies “global” variables from root into clone local maps.
 - Expose “globals” only via `parcel.Globals()` (read-only map).
 - If script-level access is needed, provide a read-only accessor builtin (e.g., `getglobal(name)`), backed by `parcel.Globals()`.
 - If temporary compatibility is needed, add a read-through in variable resolution (on miss, consult `parcel.Globals()`), but **do not** write into locals and **do not** allow mutation.

 ---
 ## 8) Sensitive lookups — replace directly with parcel/catalog calls
 - Identity  → `interp.getParcel().Identity()`
 - AEIOU     → `interp.getParcel().AEIOU()`
 - Policy    → `interp.getParcel().Policy()`
 - Logger    → `interp.getParcel().Logger()`
 - Tools     → `interp.Catalogs().Tools()`
 - Accounts  → `interp.Catalogs().Accounts()`
 - Models    → `interp.Catalogs().AgentModels()`
 - Capsules  → `interp.Catalogs().Capsules()`
 - Functions → `interp.Catalogs().Funcs()`
 - Events    → `interp.Catalogs().Events()`
 - Globals   → `interp.getParcel().Globals()` (read-only)
 **Action:** grep existing code paths and perform a mechanical replacement. This is where most tangles disappear.

 ---
 ## 9) Tests (fast to write, catch regressions)
 **T1: Internal clone preserves parcel**
 - Parent parcel: Identity=A, AEIOU=X, Globals={k:v}
 - Call a proc (triggers engine clone)
 - Inside tool: sees Identity=A, AEIOU=X via parcel; writes to Globals fail (or are no-ops).
 
 **T2: API clone fork**
 - API clone with `Fork` changing AEIOU→Y (or dropping a capability)
 - Child sees AEIOU=Y; parent remains X; catalogs shared; locals are fresh.
 
 **T3: No locals bleed**
 - Parent sets local `x=1`; callee sets `x=2` after internal clone
 - Parent’s `x` unchanged; parcel globals unchanged.
 
 **T4: Catalog propagation**
 - Child registers a new account via `Catalogs().Accounts()`
 - Parent can read it (shared by reference).

 ---
 ## 10) Migration plan (order of operations)
 1) Land `ax/contract` with interfaces and simple `parcel` + `sharedCatalogs` concretes.
 2) Update AX Factory/Runner to accept/build a parcel and expose `Catalogs()`.
 3) Add `parcel` field + accessors to internal Interpreter; switch identity/AEIOU/policy/logger reads to parcel.
 4) Patch **all** internal clone sites to copy parcel pointer and reset exec state.
 5) Remove globals-to-locals copying; add read-only access path via parcel.
 6) Run tests T1–T4; patch any missed direct-access sites they reveal.
 7) Keep temporary back-compat getters (deprecated) that forward to parcel/catalog; schedule removal.

 ---
 ## 11) Common pitfalls (and their antidotes here)
 - **Parcel re-synthesis:** If you see code building identity/AEIOU inside the engine, delete it. Parcel is the single source.
 - **Forgotten clone site:** Centralize in a helper `cloneForExec()` and route all engine clones through it.
 - **Globals mutation through locals:** Remove globals materialization; assert `Globals()` is read-only; add tests.
 - **Registry pointer soup:** Replace N pointers with one `SharedCatalogs` façade.

 ---
 ## 12) Done criteria (ship checklist)
 - [ ] All clone paths (API + engine) copy parcel by ref and reset exec state.
 - [ ] No code path reads identity/AEIOU/policy/logger except via parcel.
 - [ ] No code path copies globals into locals; `Globals()` is read-only.
 - [ ] SharedCatalogs is in use wherever registries are referenced.
 - [ ] Tests T1–T4 pass and are stable.
 - [ ] Deprecated shims present but unused by core; removal ticket filed.

 ---
 ## 13) Quick integration notes for your current files
 - **interpreter.go:** add `parcel RunnerParcel` field and private accessors; replace identity/AEIOU/policy/logger reads with `getParcel()` calls.
 - **interpreter_clone.go:** in the clone function right after `child := parent.shallowCloneInternals()`, add `child.setParcel(parent.getParcel())`; then reset exec state. Do this for every clone entry (proc and event).
 - **AX Runner:** ensure `Clone()` reuses catalogs, copies parcel pointer (or forks), always fresh exec state; never copy globals to locals.

 ---
 ## 14) Epilogue — why this won’t churn 20M files
 - You’re adding one private field and two accessors to the engine interpreter, a one-liner in each clone site, and a handful of mechanical renames to read via `parcel`/`catalogs`.
 - The parcel wraps your **existing** AEIOU envelope, so you’re not reinventing that wheel—just making it the one wheel everybody uses.
 - After this, “double clone loses identity” becomes a class of bugs you simply can’t represent.
