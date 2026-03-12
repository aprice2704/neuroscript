# Rules for You (the assistant)

- If the user says "NOCODE" that means do not emit any go or neuroscript code until further notice from the user: "YESCODE". md and derivatives are allowed.

### Code Emission Gate (HARD)
- The assistant must **ALWAYS** provide complete code files, NEVER abbreviated or elided. 
- NEVER provide unchanged files, unless specifically requested

## 1) Non-Negotiables (apply in all modes) 🧱
1) **No guessing.** If a symbol/definition isn’t in context: request file(s) or ask user to run **Piranha**.
2) **No hacks.** No shims/mocks/bypasses to “make it compile”.
6) **Do-not-edit files.** Never edit a file that says “DO NOT EDIT”.
7) **Terminology.** Avoid loaded terms: use `allowlist/blocklist`, `main`, `primary/replica`, etc.
8) **Never ignore returned errors, EVER**. Fix any ignored error instances ON SIGHT to fail fast and return a helpful message (or do something else useful).
9) **NO silent fails** always produce at least a WARN log, possibly ERROR, if something that **should** be true is not.
10) Similarly: we have consistently run into silent failures and deviations; whenever there is a deviation from the "straight and happy path" we must flag it to logs.

## 2) Turn Closure (NO-CRICKETS LAW) 🪨
**Every assistant turn must **end** with a `NEXT:` block. No exceptions.**

### 2.0 Placement rule (anti-scroll) ✅
- The `NEXT:` block must be the **last thing in the response**, sent as **normal chat text** (not inside a code fence).
- If emitting long code/files/logs: put them **before** `NEXT:`.
- **Nothing** may appear after `NEXT:` (no extra commentary, no “also…”, no trailing notes).

A valid `NEXT:` block answers, explicitly:
- **Do you (the user) run tests now?** If yes: exact command(s).
- **What output should be pasted back?** (and how much)
## 3) Response Envelope (what to include)
**CRITICAL:** The Response Envelope (Goal, Facts, Plan, NEXT) must be sent as **normal chat text**. Only the source code or document artifact goes inside the code block.


- **ALWAYS** provide complete code files, NEVER abbreviate or elide. 


### Content emission rule (Visual-first) 📦

- Wrap the entire file or document in **one** Markdown code fence.
- **Backtick Rule:** Use 4 (FOUR) backticks (` ```` `) or tildes `~~~~` to fence code
- `.md` or `.ns` file, you MUST use 4 backticks (` ```` `) or tildes `~~~~` for the outer fence to ensure the UI renders correctly for the transfer to final file
- Do **not** use start/end tags or per-line prefixes.
- Avoid extra prose inside the fence.

## 4) File Headers (only for modified source files)
### 4.1 NeuroScript `.ns` header (MANDATORY on modified `.ns`)
- Must be **native** metadata (NOT commented) at file start, then **one blank line**.
```text
:: product: FDM/NS
:: majorVersion: 1
:: fileVersion: X
:: description: <stable high-level purpose>
:: latestChange: <specific functional change>
:: filename: path/to/file.ns
:: serialization: ns

<first code line...>
```

### 4.2 `.go` header (MANDATORY on modified `.go`)
```go
// :: product: FDM/NS
// :: majorVersion: 1
// :: fileVersion: X
// :: description: <stable high-level purpose>
// :: latestChange: <specific functional change>
// :: filename: path/to/file.go
// :: serialization: go
```
- Use `// ::` primarily for `.go` (and `.sh` only if repo policy or user requests).

### 4.3 `.html` header (MANDATORY on modified `.html`)
To avoid aggressive markdown rendering in your UI, alwys use this format for HTML file headers.
```html
<style>
/* :: product: FDM/NS */
/* :: majorVersion: 1 */
/* :: fileVersion: 1 */
/* :: description: Services view HTML fragment */
/* :: filename: sugeno/static/pages/page-services.html */
/* :: serialization: html */
</style>
```


## 5) Resend Policy (no “just in case”)
Do **not** resend unchanged files unless:
- User explicitly asks (resend/show again), OR
- The file changed since last emission, OR
- Prior emission was corrupted/incomplete (e.g., "gibbled") and you state what was wrong.

## 6) Context Gate (ask early, ask precisely) 🔎
If required context is missing:
- Request **specific files** (definition + call site), OR ask for **Piranha**:
  - `piranha <Query>` (example: `piranha LoadFromUnit ExecWithInterpreter`)

Common requests:
- failing test output + exact command used
- file containing referenced symbol(s)
- `go.mod` if boundaries/deps matter
- relevant `interfaces/` constants and/or generated `nodes/` builders

## 7) Architectural Laws (do not violate) 🚫
### 7.1 Canonical Creation Law
- **Go:** never hand-construct field maps for known node types. Use:
  - `nodes.Build<Type>(nodes.Build<Type>Params{...})`
- **NeuroScript:** use `fdm.nodes.<owner>.<name>.create` tools for instantiation.
- If you touch legacy map construction where a builder exists: refactor to builder.

### 7.2 Handles / naming
- Do not add custom naming fields like `Name`, `Queue_name` for entity naming.
- Use reserved `handle` (unique among non-deleted entities of that type).
- Only compiled-in `handlesvc` may resolve handle -> entity. Other compiled-in tools must not.

### 7.3 NeuroScript import boundary
- Host programs may import ONLY: `github.com/aprice2704/neuroscript/pkg/api`
- Do not import internal neuroscript packages elsewhere.

### 7.4 `interfaces` Law (no magic strings)
- `interfaces` is the single source of truth for NodeType/FieldName/Topic/ValueType/ValueKind.
- Replace string literals with `interfaces` constants when touched.
- If importing `interfaces`, alias as:
  - `ix "github.com/aprice2704/fdm/code/interfaces"`

### 7.5 Immutability (COW)
- Nodes are immutable. Mutations return a new node; propagate new IDs upstream.

### 7.6 Typed payloads + service registry
- Publish events using the exact structs defined in `interfaces/` for that topic.
- No ad-hoc `map[string]any` for system events.
- Don’t instantiate core services directly; get them via `Env` / `ServiceProvider`.

### 7.7 Mutation Signature Law
- `nodes.Mutate<Type>(ctx, graph, entityID, updates map[string]any, opts...)`
- Do not invent `Mutate<Type>Params`.

## 8) DEBUG Protocol (applies when tests fail/hang/flake) 🧪🔥
**Trigger:** any failing test (red), hang, flake, or user says “debug/investigate”.

- **ALWAYS** provide complete code files, NEVER abbreviate or elide. 


### 8.0 Evidence Standards (what counts) 📌
- **Evidence means observations, not outcomes.**
  - stack traces, panic messages, assertion diffs
  - specific `DBG ...` lines with values
  - counts/lengths/types/nil-state, IDs/handles, error chains (`%+v` where useful)
  - timing/progress beacons for hangs
- **PASS/FAIL is not evidence.** It is only a routing signal for what to do next.

### 8.1 Prime Law: every DEBUG turn must create NEW OBSERVATIONS (ENFORCED)
On every DEBUG turn you must do at least one, and it must be a **real observability delta**:
- add new debug output, OR
- add a fast-fail invariant check with context, OR
- add a sub-test that narrows the space, OR
- add progress beacons / timeout guardrails for hangs/flakes.

**Not acceptable:** analysis-only, restating hypotheses, “try running again”, or treating PASS/FAIL alone as progress.

If you cannot edit code due to missing context:
- request the missing file(s)/Piranha, AND
- specify exactly what debug/sub-test you will add next turn (where + what it prints).

### 8.2 Scope discipline (stop twiddling)
- Focus on **exactly one** failing test at a time (name it: `TestXxx`).
- Touch the smallest number of files.
- No refactors/cleanups/renames while red. Add debug instead.

### 8.3 The only allowed debug loop
1) Name failing test + symptom (evidence lines; not “it failed”)
2) Instrument decision points + error sites
3) Run only that test (exact command)
4) State what is now KNOWN (facts)
5) Narrow: function/branch/invariant
6) Add more debug where evidence is missing
7) Add sub-tests if surface area spans > ~2 subsystems
8) Repeat until fixed
9) Only after user confirms pass: optional cleanup

### 8.4 Debug output quality (mandatory)
- Must include discriminating values (IDs, types, lengths, nil state, errors, key flags).
- Must be searchable with stable prefix: `DBG <pkg>.<fn>:` or `DBG <component>:`
- Prefer summaries (counts/hashes/key fields) over blobs.

For hangs/flakes/concurrency also add:
- progress beacons (counter + timestamp)
- timeout guardrails (fail fast with context)
- lock/channel acquire/release or send/receive points if relevant

### 8.5 Forbidden uncertainty words (convert to “KNOW”)

Words like: suspect/think/infer/assume/expect/imagine/probably/maybe/seems/likely

Rule: if you use one, immediately convert it to:
- “We will KNOW by adding debug X at Y printing Z, then running command C.”
- or “We will KNOW by adding sub-test T isolating F with inputs I asserting A.”

### 8.6 Debug removal policy
- Never remove debug output while tests are failing/hanging/flaky.
- Only remove after user explicitly says: “The test passes.”

### 8.7 Instrumentation delta requirement (anti-lip-service)
- Every DEBUG turn must explicitly list what new observability was added:
  - new `DBG ...` lines (where + what they print), OR
  - new assertions/invariants (where + what they assert), OR
  - new sub-tests (name + what they isolate).
- If the list is empty, the turn is invalid.

- **ALWAYS** provide complete code files, NEVER abbreviate or elide. 

## 9) Go Standards (Go 1.25+) 🧰
- Import hygiene: write correct `pkg.Symbol`; `goimports` handles ordering.
- Error handling: do more than ignore errors; fail fast with context on nil/invalid state.
- Typed errors: `errors.Is/As`, wrap with `%w`; no string compares.
- Testing: stdlib `testing` (no testify). Run narrow tests when debugging.
- Delivery: default to emitting **one** changed file per turn unless user requests more.
- LOUD stubs -- if stubbing code you MUST make it emit one or more ERROR logs saying it is a stub


- **ALWAYS** provide complete code files, NEVER abbreviate or elide. 


:: id: capsule/agents_rules
:: version: 2.6
:: description: AGENTS.md ruleset for FDM/NS. Clarified NOCODE exemption for .md/.ns. Mandatory N+1 wrapping for .md/.ns.
:: serialization: md
:: filename: AGENTS.md