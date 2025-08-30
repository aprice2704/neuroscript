o3:

Great—this is exactly the right time to turn `ask` from “just enough to demo” into a first-class bridge between ns and models. Based on your docs and ns philosophy (explicit, statement-driven, minimal surprises), here’s a concrete design + implementation plan that stays idiomatic to ns while giving you real power. &#x20;

# Goals (in order)

1. **Keep simple uses simple.** One line, one answer, no config ceremony.
2. **Make advanced uses explicit, not magical.** Options are just maps, messages are just lists of maps.
3. **Exploit ns’s event model and string interpolation.** Streaming = events; prompts = `…` with `{{placeholders}}`.&#x20;

# Proposed syntax (back-compat)

Current (keep working):

```neuroscript
ask "What is the capital of France?" into answer
```

New (all optional parts are orthogonal):

````neuroscript
# 1) Options map
ask "Summarize {{text}}" with {
  "model": "gpt-4o-mini",
  "temperature": 0.2,
  "max_tokens": 800
} into summary

# 2) Message list (system/user/assistant)
ask [
  {"role": "system", "content": "You are concise."},
  {"role": "user",   "content": ```Summarize {{text}} in 5 bullets.```}
] with {"model": "gpt-4o-mini"} into bullets

# 3) Structured output (JSON mode) with schema
ask "Extract name and email from:\n{{blob}}" with {
  "model": "gpt-4o",
  "json": true,
  "schema": {
    "type": "object",
    "required": ["name", "email"],
    "properties": {
      "name": {"type": "string"},
      "email": {"type": "string", "format": "email"}
    }
  }
} into person

# 4) Streaming tokens → events (for live UI / logs / FDM capture)
ask "Write a short haiku about Kamloops" with {
  "model": "gpt-4o-mini",
  "stream": true,
  "stream_event": "ai.token"
} into poem
````

### Grammar sketch

```
ask_stmt
  : 'ask' expr
    ('with' expr)?                // map of options
    ('into' lvalue (',' lvalue)*)?// allow 1..N (see extraction below)
    ('stream' expr)?              // shorthand, but prefer options.stream_event
  ;
```

* `expr` is either a string (single shot) **or** a list of `{role, content}` maps (chat).
* `with` takes a plain ns map—no new syntax needed.
* `into` allows **one** lvalue for text/JSON, **or multiple** when `extract` is provided (below).
* For basic users, none of this exists; they still write `ask "…" into x`. That honors the “no surprises” rule.&#x20;

# Options (host-validated, not grammar-enforced)

All optional; defaults should make sense.

```json
{
  "model": "string",
  "system": "string",            // prepended if using single-string prompt
  "temperature": 0.2,
  "top_p": 1.0,
  "max_tokens": 1024,
  "timeout_ms": 30000,
  "seed": 0,                     // if provider supports it (determinism)
  "json": false,                 // request structured JSON
  "schema": { ... },             // JSON Schema draft 2020-12
  "messages": [ ... ],           // override: explicit message list
  "tools": [ { "name": "tool.fs.ReadFile", "args": {"path":"string"} } ],
  "tool_choice": "auto|none|required",
  "stream": false,
  "stream_event": "ai.token",    // event name to emit per delta
  "attachments": [ {"type":"image","data": <bytes> }, ... ],
  "cache": true,                 // host-side response cache
  "budget": {"tokens": 100000, "cad": 2.00}, // soft limits
  "extract": {                   // structured extraction → multiple lvalues
    "name": "$.name",
    "email": "$.email"
  },
  "metadata": { "run_id": "…" }  // echoed to events & logs
}
```

* **Messages**: if you pass a list, that wins. Otherwise host composes `{system?, user: prompt}`.
* **JSON/Schema**: host enforces; ns just sees a `map` on success.
* **Tools**: *optional*. If enabled, the host may call whitelisted `tool.<group>.<name>` with explicit args; each call shows up as an event; the interpreter executes and returns results to the model (function-calling style) with a strict budget + step cap.
* **Streaming**: deltas emitted via `emit <event>` immediately (synchronous as per ns events), then final assembled result assigned into your `into` lvalue.&#x20;

# Return shapes

* Default: **string** (final text).
* With `"json": true`: **map** (validated) or error.
* With `"extract"` and multiple `into` targets: host applies JSONPath and assigns in order:

  ```neuroscript
  ask "Give {name,email}" with {"json":true, "extract":{"a":"$.name","b":"$.email"}} into name, email
  ```

# Error semantics (fits existing model)

* Any provider/model/tool failure → raises `error`. Catch with `on error do … endon`. Use `clear_error` to resume, consistent with other failures.&#x20;

# Examples you’ll actually use in FDM

**1) Deterministic metadata extractor (cached)**

````neuroscript
set doc = tool.fs.ReadFile("customer_email.txt")

ask ```Extract a JSON object:
- name
- email
- company
From:
{{doc}}``` with {
  "model": "gpt-4o",
  "json": true,
  "schema": {"type":"object","properties":{"name":{"type":"string"},"email":{"type":"string"},"company":{"type":"string"}}, "required":["name","email"]},
  "temperature": 0.0,
  "cache": true
} into customer

must typeof(customer) == "map"
emit {"event":"parsed.customer", "data": customer}
````

**2) Live UI stream while writing**

```neuroscript
on event "ai.token" as tok do
  call tool.log.info(tok)  # or push to websocket
endon

ask "Draft a 150-word release note about {{feature}}" with {
  "model": "gpt-4o-mini",
  "stream": true,
  "stream_event": "ai.token"
} into release_note
```

**3) Tool-using model (strictly whitelisted)**

```neuroscript
ask "Count lines of src/ns/*.go and return JSON {total}" with {
  "model": "gpt-4o",
  "json": true,
  "tools": [
    {"name":"tool.fs.Glob", "args":{"pattern":"string"}},
    {"name":"tool.fs.ReadFile", "args":{"path":"string"}}
  ],
  "tool_choice": "auto",
  "schema": {"type":"object","required":["total"],"properties":{"total":{"type":"number"}}}
} into stats
```

# Interpreter & runtime (Go) — implementation plan

### AST

Add an `AskStmt` node:

```go
type AskStmt struct {
    PromptExpr   Expr   // string OR list-of-maps OR nil if using Options.Messages only
    OptionsExpr  Expr   // map or nil
    IntoTargets  []LValue // 0..N, usually 0 or 1; N>1 requires Options.extract
    StreamName   Expr   // deprecated shorthand; prefer options.stream_event
    SrcRange     Range
}
```

### Runtime contracts

```go
type AIMessage struct { Role, Content string }
type AIToolSpec struct { Name string; Args map[string]string } // arg types are hints
type AIExtract map[string]string // name -> JSONPath

type AIRequest struct {
    Model       string
    Messages    []AIMessage     // if empty and Prompt non-empty, host composes from System+Prompt
    Prompt      string          // optional (single-shot)
    System      string          // optional
    JSON        bool
    Schema      any
    Temperature float64
    TopP        float64
    MaxTokens   int
    Timeout     time.Duration
    Seed        *int64
    Tools       []AIToolSpec
    ToolChoice  string
    Stream      bool
    StreamEvent string
    Attachments []Attachment
    Cache       bool
    Budget      Budget
    Extract     AIExtract
    Metadata    map[string]any
}

type AIStreamChunk struct {
    DeltaText  string
    ToolCall   *ToolInvocation // if the model requests a tool
    Done       bool
}

type AIResponse struct {
    Text     string
    JSON     map[string]any
    TokensIn int
    TokensOut int
    CostCAD  float64
    Raw      any
}

type AIProvider interface {
    Chat(ctx context.Context, req AIRequest) (iter <-chan AIStreamChunk, done func() (AIResponse, error), err error)
}
```

### Execution flow

1. **Evaluate** `PromptExpr` / `OptionsExpr` to Go values (string | \[]map | map).
2. **Normalize** to `AIRequest` (merge defaults; if messages absent, compose from `system` + `prompt`).
3. If `Stream`:

   * Start goroutine reading `iter`.
   * For each chunk with `DeltaText`, `emit StreamEvent` with payload `{delta, metadata}` (sync with ns event semantics).
   * For a `ToolCall`, **validate name** ∈ whitelist, marshal args, call the named ns tool synchronously, send result back to provider (depends on provider API; store partial state in `done` closure or side-channel).
4. On completion, `done()` → `AIResponse`.
5. **Assign**:

   * If `json: true`, assign `map`.
   * Else assign `Text`.
   * If `extract` present and multiple `into`, apply JSONPath to the JSON result and assign in the provided order; raise `error` on missing path.
6. **Emit** a summary event like `"ai.ask.done"` with `{tokens_in, tokens_out, cost_cad, model, metadata}` for FDM auditing. (Nice fit for your “track every decision” posture.)&#x20;

### Validation & safety

* Enforce `tool_choice`/`tools` strictly; hard cap tool invocations per ask (e.g., 8).
* Apply JSON Schema (if provided) with `gojsonschema`. On failure: raise `error`.
* Respect `timeout` and `budget` early-abort.
* Default `temperature` low (0.2). If `seed` is supported, allow determinism for testability.
* Provider abstraction lets you back models with OpenAI, local, or your own gateway. (Neatly matches `tool` philosophy.)&#x20;

# Tests (outline)

**Interpreter unit tests**

* `ask_basic_text`: returns string.
* `ask_json_schema_ok/bad`: validates JSON or errors.
* `ask_extract_multi_into`: assigns two vars from JSON via JSONPath.
* `ask_stream_emits_tokens`: collects `ai.token` events and concatenates to final result.
* `ask_toolcall_whitelist`: model requests forbidden tool → error; allowed tool → executed; budget cap enforced.
* `ask_on_error`: proves normal ns error handling works (`on error` + `clear_error`).&#x20;

# Why this fits ns

* It’s still a **single explicit statement** (`ask ...`) that composes with `set`, `must`, `on error`, and events—no hidden control flow.
* Config is **data**, not syntax: options are a map; messages are a list; schemas are plain maps.
* **Streaming = events**, leveraging your existing synchronous event model.
* Back-compat preserved; new power when you need it.&#x20;

---

If you want, I can draft:

1. the `AskStmt` parser rule + AST node,
2. the `AIProvider` interface and a no-op fake for tests, and
3. a first provider implementation against your current model gateway.

Say the word and I’ll sketch the Go diffs next.

AJP: Overall: very good, esp adherence to design principles. Some points:
1. I'm not a big fan of jsonschema or jsonpath -- I find them overly complex and a bit hard to read. Can you think of simpler alternatives?
2. Please explain how streaming would work.
3. We need to specify accounts/creds somehow and/or associate uploaded files (fileapi) -- ideas?
4. We need to make sure we have provision for setting budget, perhaps getting estimates and certainly reporting tokens used

o3: Great notes. Here’s how I’d tune the design to match your “simple, explicit, no surprises” ethos.

# 1) Simpler than JSON Schema / JSONPath

### a) **Shape-lite** (tiny type spec, no keywords soup)

Use a tiny, readable type map instead of JSON Schema. Supported primitives are obvious and few.

```neuroscript
ask "Extract contact from:\n{{blob}}" with {
  "json": true,
  "shape": {
    "name":   "string",
    "email":  "email",    # builtin validator
    "age?":   "int",      # '?' = optional
    "tags":   "list<string>",
    "meta":   "map<any>"
  }
} into contact
```

* **Exactly** those keys are allowed unless `"allow_extra": true`.
* Primitives: `string,int,float,bool,email,url,isoDatetime,any`.
* Containers: `list<T>`, `map<T>`, `object{...}` (alias of nested shapes).
* Errors are blunt: first mismatch wins; report a short path, e.g. `meta.flags[2] expected bool, got string`.

Implementation: a \~200 LOC validator; no external schema lib; dead simple to read and maintain.

### b) **Select-lite** (dot/bracket paths only)

For extraction, no JSONPath—just dot + numeric index. Also allow an **array of path segments** to dodge escaping issues.

```neuroscript
ask "Return JSON with {name, contact:{email}}" with {
  "json": true,
  "extract": {
    "name":   "name",
    "email":  "contact.email",
    "first":  "items[0]"
    # or:  "email": ["contact", "email"], "first": ["items", 0]
  }
} into name, email, first
```

* No wildcards, filters, or predicates. If the path isn’t there → error (or `"missing_ok": true` to assign `nil`).

If you want **zero new options**, you can also do this post-hoc with a tiny stdlib function:

```neuroscript
set email = ai.select(contact, "contact.email")  # same path rules
```

# 2) Streaming — exact mechanics

**Goal:** live tokens without breaking ns’s synchronous event model.

* Adapter returns an iterator of chunks `{deltaText?, toolCall?, done?}`.

* Runtime spawns a goroutine to read chunks and **accumulate** final text in a buffer.

* For each `deltaText`:

  * If `options.stream_event` is set, emit:

    ```neuroscript
    emit {"event": options.stream_event, "data": {
      "delta": "...",
      "seq":   n,              # monotonically increasing
      "accum_len": len_so_far, # bytes or runes; we’ll pick runes
      "model": req.Model,
      "run_id": options.metadata.run_id
    }}
    ```
  * Event handlers run **synchronously** (as ns defines). If a handler is heavy, it will slow the stream—by design. For UI, do minimal work and buffer app-side.

* Tool-calls (if allowed) temporarily **pause** token deltas:

  * Emit `ai.tool_call` with `{name,args,seq}`.
  * Execute the whitelisted tool; emit `ai.tool_result`.
  * Resume streaming.

* On completion:

  * Assemble `final = concat(deltas)`.
  * If `json==true`, parse once at the end, then run `shape` validation and/or `extract`.
  * Assign to `into` target(s).
  * Emit `ai.ask.done` with usage (see §4).

Edge behaviors you’ll care about:

* **Errors mid-stream** → emit `ai.ask.error` (with partial text length); by default the `into` target is **not assigned**. Add `"partial_ok": true` if you want the partial assigned.
* **Back-pressure**: if you want to avoid over-emitting, support:

  * `"coalesce_ms": 25` to batch small deltas,
  * `"emit_on": "token|word|line|paragraph"` simple heuristics (e.g., emit only when we see whitespace/newline/period).

# 3) Accounts / creds / files (fileapi)

### a) Account selection (explicit but low ceremony)

In config, define named accounts:

```yaml
ai_accounts:
  default:
    provider: openai
    base_url: https://api.openai.com/v1
    api_key: ${OPENAI_API_KEY}
    price_table: # optional; see §4
      input_per_1k: 0.00   # CAD
      output_per_1k: 0.00
  local:
    provider: openrouter-like
    base_url: http://localhost:11434
    api_key: ""
```

In code:

```neuroscript
ask "…" with {"account": "default", "model": "gpt-4o-mini"} into x
```

* Default account if unspecified.
* **Model aliases** live in config (decouples scripts from vendor renames). If alias isn’t found, treat literally.

### b) Associating files (no magic, one format)

Treat attachments as **explicit file refs** resolving through your fileapi. The adapter decides whether to stream bytes inline or upload to the model provider first.

```neuroscript
ask "Describe this picture" with {
  "model": "…",
  "attachments": [
    {"file_id": "f:9b2c…", "purpose": "input", "mime": "image/png"},
    {"file_id": "f:ab93…", "purpose": "knowledge", "mime": "application/pdf"}
  ],
  "file_policy": {
    "upload": "auto|inline|forbid",    # default: auto
    "max_upload_bytes": 20_000_000,
    "scope": "ephemeral|account|run"   # provider-specific
  }
} into desc
```

* The runtime calls `fileapi.Get(file_id)` to obtain a signed URL or bytes.
* Provider adapters implement “upload if needed; else reference by URL if supported”.
* All file movements are logged. We’ll also emit `ai.file.attach` events for audit.

# 4) Budgets, estimates, and usage reporting

### a) Budgets (hard stops, soft caps)

```neuroscript
ask "Summarize {{doc}}" with {
  "budget": { "tokens": 8000, "cad": 1.50 },   # both optional
  "reserve_output_tokens": 512,                # truncation guidance
  "abort_if_over_budget": true                 # based on estimate
} into s
```

* **Preflight estimate** (best-effort):

  * If the adapter has a tokenizer, it estimates input tokens.
  * If not, use a local fallback (approx BPE) with a ±10–15% warning flag.
  * If estimated *input + reserve\_output* would exceed `budget.tokens` or `budget.cad` (using price table), abort early with `error` before calling the model (unless `abort_if_over_budget:false`).

* **Live guardrails**:

  * We always pass `max_output_tokens` to the provider (derived from `reserve_output_tokens` and budget).
  * We can apply **prompt truncation** with a simple strategy:

    ```json
    "truncation": {"type":"front|middle|none", "max_input_tokens": 6000}
    ```

    If we truncate, emit `ai.ask.truncated` with counts.

### b) Usage & cost (CAD)

On completion (or error), emit and return:

```json
{
  "usage": {
    "input_tokens": 1234,
    "output_tokens": 456,
    "cache_read_tokens": 0,
    "cache_write_tokens": 0,
    "total_tokens": 1690,
    "latency_ms": 842,
    "model": "…",
    "account": "default",
    "estimated": false,
    "cost": {"cad": 0.23, "calc": "price_table:v1"}
  }
}
```

* Cost is computed from a **per-account price table** in CAD:

  ```yaml
  price_table:
    input_per_1k: 0.80   # CAD
    output_per_1k: 3.20
    rounding: ceil_1k    # or exact
  ```
* If the provider returns authoritative usage, we trust it; else we fallback to our estimate and mark `"estimated": true`.
* These metrics get attached to `ai.ask.done` and are accessible via `last_error` context on failure.

---

## Minimal diffs to your earlier plan

* Replace `"schema"` with `"shape"` and `"allow_extra"`; keep `"json": true`.
* Replace JSONPath with **Select-lite** paths (string or `[]any`).
* Add `"account"`, `"attachments"`, and `"file_policy"`.
* Add `"budget"`, `"reserve_output_tokens"`, `"truncation"`, `"abort_if_over_budget"`, and robust `"usage"` reporting.
* Streaming clarified with `stream_event`, `coalesce_ms`, and `emit_on`.

---

## Quick end-to-end examples

**A. Simple + budget + usage**

```neuroscript
ask "One-paragraph summary of:\n{{doc}}" with {
  "model": "mini",
  "budget": {"cad": 0.10},
  "reserve_output_tokens": 256
} into s
```

**B. JSON with shape + extraction**

```neuroscript
ask "Return {name, contact:{email,phone?}, tags:list}" with {
  "json": true,
  "shape": {"name":"string","contact":"object{email:string,phone?:string}","tags":"list<string>"},
  "extract": {"name":"name", "email":"contact.email"}
} into name, email
```

**C. Streaming to UI with coalescing**

```neuroscript
on event "ai.delta" as e do
  call tool.ui.append(e.delta)  # do as little as possible here
endon

ask "Write a 150-word release note about {{feature}}" with {
  "stream": true,
  "stream_event": "ai.delta",
  "coalesce_ms": 25,
  "emit_on": "word"
} into final_text
```

**D. Account + files**

```neuroscript
ask "Extract key facts from the attached PDF and reply JSON." with {
  "account": "default",
  "model": "gpt-x",
  "json": true,
  "shape": {"title":"string","bullets":"list<string>"},
  "attachments": [{"file_id":"f:7b1…","purpose":"knowledge","mime":"application/pdf"}],
  "file_policy": {"upload":"auto","scope":"ephemeral"}
} into facts
```

---

If you’re happy with this direction, I’ll:

1. wire `shape` + `select` validators,
2. add `account`/`attachments` plumbing (adapter + fileapi shim),
3. implement streaming with `coalesce_ms` and `emit_on`,
4. add preflight estimation + price-table CAD costs,
5. deliver a focused test suite (shape happy/sad, extract, stream, budget-abort, file attach).




ajp: Are shape-lite and path-lite things you invented? In any event, I like them -- please do some more worked examples for me




o3:
Short answer: yeah—“shape-lite” and “path-lite” are my working names here (not standards). They’re deliberately tiny: just enough to be readable, testable, and hard to misuse. Think “typed-JSON with training wheels” + “dot/index paths, nothing fancy.”

Below are worked examples you can drop straight into ns.

# Shape-lite: quick cheat-sheet

* Primitives: `string`, `int`, `float`, `bool`, `email`, `url`, `isoDatetime`, `any`
* Containers: `list<T>`, `map<T>`, `object{...}` (inline object)
* Optional field: `field?`
* Global flags on `ask … with { … }`:

  * `"json": true`, `"shape": {…}`, `"allow_extra": false` (default)

You can also reuse shapes via a variable (plain ns map).

# Path-lite: quick cheat-sheet

* String paths with `.` and `[index]`: `a.b[0].c`
* Or array form to avoid escaping: `["odd key", 0, "inner"]`
* Missing → error by default; set `"missing_ok": true` (per-extract or global) if you want `nil`

---

# 1) Minimal “typed JSON” extract

```neuroscript
set blob = tool.fs.ReadFile("customer.txt")

ask """Return JSON with:
- name
- email
- company
""" with {
  "json": true,
  "shape": {
    "name":   "string",
    "email":  "email",
    "company":"string"
  }
} into customer

# customer is a map with those 3 fields, else error
```

# 2) Optional fields and extra keys

```neuroscript
ask "Return {name, email, phone?} plus anything else you like" with {
  "json": true,
  "shape": {"name":"string","email":"email","phone?":"string"},
  "allow_extra": true
} into c

# c may contain extra keys; the three typed ones are validated
```

# 3) Nested objects (two styles)

**a) Inline with `object{…}`**

```neuroscript
ask "Return {name, contact:{email, phone?}}" with {
  "json": true,
  "shape": {
    "name": "string",
    "contact": "object{email:email, phone?:string}"
  }
} into c
```

**b) As a nested shape map (equivalent, usually easier to read)**

```neuroscript
set ContactShape = {"email":"email","phone?":"string"}

ask "Return {name, contact:{email, phone?}}" with {
  "json": true,
  "shape": {"name":"string", "contact": ContactShape}
} into c
```

# 4) Lists of objects

```neuroscript
ask "Return {title, items:[{name, qty}]}" with {
  "json": true,
  "shape": {
    "title":"string",
    "items":"list<object{name:string, qty:int}>"
  }
} into order
```

# 5) Maps (dictionary) of uniform values

```neuroscript
ask "Return a map of feature flags as booleans" with {
  "json": true,
  "shape": {"flags":"map<bool>"}
} into cfg
```

# 6) Extraction with path-lite → multiple variables

```neuroscript
ask "Return {user:{name,email}, items:[{sku,price}]}" with {
  "json": true,
  "shape": {
    "user":"object{name:string, email:email}",
    "items":"list<object{sku:string, price:float}>"
  },
  "extract": {
    "name":  "user.name",
    "email": "user.email",
    "first": "items[0].sku"
  }
} into name, email, first
```

# 7) Path array form for “weird” keys

```neuroscript
# Suppose the model returns {"weird key!": {"inner value": 42}}
ask "Return that structure" with {
  "json": true,
  "shape": {"weird key!":"object{inner value:int}"},
  "extract": {"answer": ["weird key!", "inner value"]}
} into answer
```

# 8) Graceful missing fields during extract

```neuroscript
ask "Return {user:{name}, notes?}" with {
  "json": true,
  "shape": {"user":{"name":"string"}, "notes?":"string"},
  "extract": {"notes":"notes"},
  "missing_ok": true
} into notes
# If 'notes' absent → notes = nil (no error)
```

# 9) Validating *outside* `ask` (std-lib helpers)

You’ll sometimes validate model JSON you already have.

```neuroscript
set shape = {"name":"string","email":"email"}
set ok = ai.validate(candidate, shape, /*allow_extra=*/false)
must ok == true

set v = ai.select(candidate, "contact.email")  # or ["contact","email"]
```

* `ai.validate(value, shape, allow_extra=false) -> bool` (or raises with path on failure—choose one; I’d raise)
* `ai.select(value, path, missing_ok=false) -> any|nil`

# 10) Tight error reporting (what it looks like)

If validation fails, you get a precise path:

```
error ai.shape: contact.phone expected string, got number (path: contact.phone)
```

If extraction fails (and `missing_ok` is false):

```
error ai.extract: path items[3].sku not found (list length 2)
```

# 11) Mixed list shapes (when you must accept variance)

```neuroscript
# Keep it pragmatic: if the provider sometimes mixes types, accept 'any'
ask "Return {events:[...]} as JSON" with {
  "json": true,
  "shape": {"events":"list<any>"}
} into evtlog
```

(Deliberately no unions—stays simple.)

# 12) Schema reuse and evolution

```neuroscript
set V1Person = {"name":"string","email":"email"}
set V2Person = {"name":"string","email":"email","company?":"string"}

ask "Return person v2 if available, else v1" with {
  "json": true,
  "shape": V2Person,
  "allow_extra": false
} into p

# If you need to accept either, validate twice in a pinch:
on error do
  clear_error
  must ai.validate(p, V1Person)  # fall back acceptance
endon
```

# 13) List constraints (lightweight, optional)

If you want a touch more control without sliding into JSON-Schema territory, we can add **non-type** hints as keys prefixed with `_` inside `object{…}` or nested maps. These are **advisory**—the validator enforces them if present, but we keep the set tiny.

Supported (if you want them): `_minItems`, `_maxItems`, `_min`, `_max`, `_pattern` (for strings; literal Go regex).

```neuroscript
set Product = {"sku":"string","price":"float"}
set Cart = {
  "_minItems": 1,
  "items":"list<object{sku:string, price:float}>"
}

ask "Return {items:[{sku,price}]}" with {"json":true,"shape":Cart} into cart
```

If you’d rather **not** add these, skip them—the core stays as shown above.

# 14) Round-trip example (shape + extract + post-use)

```neuroscript
set Doc = tool.fs.ReadFile("meeting_notes.md")

ask """Extract JSON:
- title
- decisions: list of {text, owner?, due? isoDatetime}
- summary: string (<= 120 words)
""" with {
  "json": true,
  "shape": {
    "title": "string",
    "summary":"string",
    "decisions":"list<object{text:string, owner?:string, due?:isoDatetime}>"
  },
  "extract": {
    "first_decision": "decisions[0].text",
    "owner":          "decisions[0].owner"
  },
  "missing_ok": true
} into first_decision, owner

emit {"event":"parsed.decisions", "data": first_decision}
```

---

If you want, I’ll codify this as a tiny EBNF and wire up:

* `ai.validate` + `ai.select` helpers,
* the `shape` parser (≤300 LOC),
* and a deterministic test set (happy/sad paths, nested, lists, extract, missing\_ok, allow\_extra).

Anything you’d tweak before I lock the mini-spec?




AJP: We don't have objects -- just maps, so remove those for now plz.
_min and friends hints look ok as they are self-documenting rather than needing odd syntax.
We could perhaps use shapes with `must` or `must be`?
Later we can add keywords for various interrogations and extractions, but for now, apart for expanding the ask syntax lets put those in tools.



Got it. Here’s a cleaned-up, **maps-only** take that keeps v1 of `ask` lean (just `with` and `into`), and pushes interrogation/extraction into tiny **tools**.

# Shape-lite (v0.2: maps + lists only)

**What a “shape” is:** a plain ns **map** whose keys are field names and whose values are type strings *or* nested shapes.
No “object{…}”, no keywords soup, no constraints.

* **Primitives:** `string`, `int`, `float`, `bool`, `email`, `url`, `isoDatetime`, `any`
* **Optional field:** `field?`
* **List of T:** use `[]` on the **field name** → `tags[]: "string"` or `items[]: {sku:"string", qty:"int"}`
* **Nested maps:** just nest another shape map as the value
* **Extras:** default **disallowed**; pass `allow_extra=true` to the tool when you *want* to ignore unknown keys

**Note on weird literal keys:** because `?` and `[]` have meaning in shape keys, you can’t validate a field that literally ends with `?` or `[]` via shapes in v0.2. If that ever matters, handle it with `allow_extra=true` or validate that sub-map manually.

### Examples (shape definitions)

```neuroscript
# Flat
set Person = {"name":"string","email":"email","company?":"string"}

# Nested
set ContactCard = {
  "name":"string",
  "contact":{"email":"email","phone?":"string"}
}

# List of primitives
set Tags = {"tags[]":"string"}

# List of maps
set Cart = {"items[]":{"sku":"string","qty":"int","price":"float"}}
```

# Path-lite (v0.2)

Trivial, readable selection paths.

* **String form:** `a.b[0].c`  (segments: identifiers + `[index]`)
* **Array form:** `["odd key", 0, "inner"]` (use when keys have spaces/punctuation)
* Missing path → **error** by default; per-call `missing_ok=true` to return `nil`.

# v1 `ask` (no schema/extract inside it)

Keep `ask` focused:

```neuroscript
ask "Return JSON for {name,email}" with {
  "model": "mini",
  "json": true,         # request JSON, return as map
  "stream": true,       # if desired
  "stream_event": "ai.delta"
} into raw
```

Then immediately post-process with tools:

```neuroscript
must tool.ai.Validate(raw, Person, /*allow_extra=*/false)

set email = tool.ai.Select(raw, "email")                 # string path
set firstSku = tool.ai.Select(raw, ["items", 0, "sku"])  # array path
```

# Tools — candidate signatures

### 1) Validation

```neuroscript
# Returns bool. On false, sets a descriptive last_error; pair with `must`.
tool.ai.Validate(value:any, shape:map, allow_extra?:bool=false) -> bool
```

* Fails fast with a short path: `contact.phone expected string, got number`
* Applies `allow_extra` at **every** nested map level

### 2) Selection (one path)

```neuroscript
# path: string "a.b[0].c" OR array ["a", 0, "c"]
tool.ai.Select(value:any, path:any, missing_ok?:bool=false) -> any
```

### 3) Selection (many at once)

```neuroscript
# extracts: {"name":"user.name", "email":["user","email"]}
tool.ai.SelectMany(value:any, extracts:map, missing_ok?:bool=false) -> map
```

### 4) Convenience (type checks you’ll actually use)

```neuroscript
tool.ai.IsShape(value:any, shape:map, allow_extra?:bool=false) -> bool   # alias of Validate
tool.ai.Keys(value:any) -> list<string>                                   # map keys (guard rails)
tool.ai.Len(value:any) -> int                                             # list length or map size
```

*(All pure, no side effects. `tool.ai.*` is a normal tool namespace like `tool.fs.*`.)*

# Worked flows (drop-in)

## A) Ask → validate → extract → assign

```neuroscript
set Person = {"name":"string","email":"email","company?":"string"}

ask "Return person as JSON {name,email,company?}" with {"json":true} into raw

must tool.ai.Validate(raw, Person)

set name  = tool.ai.Select(raw, "name")
set email = tool.ai.Select(raw, "email")
```

## B) Ask streaming, then validate

```neuroscript
on event "ai.delta" as e do
  call tool.log.info(e.delta)   # keep handlers tiny; they run synchronously
endon

ask "Summarize this:\n{{doc}}" with {
  "model":"mini",
  "json": true,
  "stream": true,
  "stream_event": "ai.delta"
} into summaryMap

must tool.ai.Validate(summaryMap, {"summary":"string","bullets[]":"string"})
```

## C) Lists of maps + bulk extraction

```neuroscript
set Cart = {"items[]":{"sku":"string","qty":"int","price":"float"}}

ask "Return a shopping cart JSON with items[{sku,qty,price}]" with {"json":true} into cart

must tool.ai.Validate(cart, Cart)

set first = tool.ai.SelectMany(cart, {
  "sku":  "items[0].sku",
  "qty":  "items[0].qty",
  "price":"items[0].price"
})
```

## D) Optional field + allow extra in a nested blob

```neuroscript
set Card = {"name":"string","contact":{"email":"email","phone?":"string"}}

ask "Return a contact card JSON" with {"json":true} into cc

# Allow the model to include extras like 'notes', 'tags', etc.
must tool.ai.Validate(cc, Card, /*allow_extra=*/true)

set phone = tool.ai.Select(cc, "contact.phone", /*missing_ok=*/true)
```

# Error shapes you’ll see

* **Validation**:

  * `ai.shape: contact.phone expected string, got number (path: contact.phone)`
  * `ai.shape: items expected list, got map (path: items)`
  * `ai.shape: missing required field email (path: email)`

* **Select**:

  * `ai.select: path items[2] not found (list length 2)`
  * with `missing_ok=true` → returns `nil`, no error

# Go host stubs (so you can wire it fast)

```go
// Path-lite
type Path = []any // elements are string (map key) or int (list index)

func ParsePath(s string) (Path, error) // "a.b[0].c" -> []any{"a", "b", 0, "c"}

// Validate
func ValidateShape(value any, shape map[string]any, allowExtra bool) error

// Select
func Select(value any, path Path, missingOK bool) (any, error)

// SelectMany
func SelectMany(value any, extracts map[string]any, missingOK bool) (map[string]any, error)
```

**Shape rules encoded in `ValidateShape`:**

* For each `k` in `shape`:

  * If `k` ends with `?`, treat as optional; trim the `?` for lookup.
  * If `k` ends with `[]`, expect a **list**;

    * If value is a **string** type name → enforce primitive element type.
    * If value is a **map** → enforce each element against that nested shape.
  * Else expect a **map** (if nested map) or a primitive.
* `allowExtra`: when false, any extra keys at any nested map level raise.

# Minimal tests you’ll want

* Happy path: flat, nested, list-of-primitive, list-of-map
* Optional present/absent
* Extra keys rejected/accepted (allow\_extra toggle)
* Wrong primitive type
* Wrong container (list vs map)
* Select: good string path, good array path, missing element, out-of-range index, weird keys via array form

---

If this looks right, I’ll:

* PR `tool.ai.Validate`, `tool.ai.Select`, `tool.ai.SelectMany` with the above signatures,
* wire path parsing + shape checker,
* add \~25 tests (happy/sad + edge cases),
* and leave `ask` v1 strictly to `with`/`into` (plus streaming/budget you already approved).

