# AEIOU envelope format discussion

AJP: I want to add a new package to ns, which will handle the AEIOU format for AIs to send data back to ns from ask. The discussion about it is attached -- it is the later part.

First, please devise a crisp, short, clear, precise standard/instructions for LLMs that explain it. md, inline code box, @@@ prepended to each line plz as usual.


Gemini2.5: Got it. Based on the detailed discussion, here is a crisp and precise standard for LLMs on how to use the AEIOU envelope format.

````text
@@@# AEIOU Envelope Standard for LLM Responses (v1)
@@@
@@@**Purpose:** To structure your responses in a way that NeuroScript can safely and reliably parse and execute.
@@@
@@@**Magic Constant:** `NSENVELOPE_MAGIC_9E3B6F2D`
@@@
@@@---
@@@## Core Rule
@@@
@@@- Every turn, you MUST wrap your entire response in a single AEIOU envelope.
@@@- The envelope consists of up to five sections, each with a unique magic header.
@@@- Sections MUST appear in this exact order: ACTIONS, EVENTS, IMPLEMENTATIONS, OUTPUT, UPSERTS.
@@@- Any section can be omitted if you have no content for it.
@@@
@@@---
@@@## Sections
@@@
@@@### 1. ACTIONS
@@@`<<<NSENVELOPE_MAGIC_9E3B6F2D::ACTIONS::v1>>>`
@@@- **Contains:** One or more NeuroScript `command` blocks.
@@@- **Use:** For immediate, one-off tasks.
@@@- **Strictly Prohibited:** `func`, `return`, `on event`.
@@@
@@@### 2. EVENTS
@@@`<<<NSENVELOPE_MAGIC_9E3B6F2D::EVENTS::v1>>>`
@@@- **Contains:** NeuroScript `on event` handlers only.
@@@- **Use:** To install reactive logic.
@@@- **Strictly Prohibited:** `command`, `func`.
@@@
@@@### 3. IMPLEMENTATIONS
@@@`<<<NSENVELOPE_MAGIC_9E3B6F2D::IMPLEMENTATIONS::v1>>>`
@@@- **Contains:** NeuroScript `func` definitions only.
@@@- **Use:** To define reusable, procedural knowledge.
@@@- **Strictly Prohibited:** `command`, `on event`.
@@@
@@@### 4. OUTPUT
@@@`<<<NSENVELOPE_MAGIC_9E3B6F2D::OUTPUT::v1>>>`
@@@- **Contains:** Markdown or plain text.
@@@- **Use:** For your notes, rationale, and summaries for human review. This is your "chain of thought."
@@@- **Note:** This section is never executed.
@@@
@@@### 5. UPSERTS
@@@`<<<NSENVELOPE_MAGIC_9E3B6F2D::UPSERTS::v1>>>`
@@@- **Contains:** A single JSON object.
@@@- **Use:** To update your memory and request knowledge assets.
@@@- **Schema:**
@@@  ```json
@@@  {
@@@    "memory": [
@@@      {"op":"set", "path":"a.b.c", "value": "new_value"},
@@@      {"op":"append", "path":"list.x", "value": "item"},
@@@      {"op":"delete", "path":"a.b"}
@@@    ],
@@@    "requests": {
@@@      "capsules": ["caps/ns/1"],
@@@      "toolcards": ["fdm.read"],
@@@      "overlays": ["gao@v1:pkg/types/*"]
@@@    }
@@@  }
@@@  ```
@@@---
@@@## Example
@@@
@@@<<<NSENVELOPE_MAGIC_9E3B6F2D::IMPLEMENTATIONS::v1>>>
@@@```neuroscript
@@@func greet(needs name returns msg) means
@@@  set msg = "Hello, " + name
@@@  return msg
@@@endfunc
@@@```
@@@
@@@<<<NSENVELOPE_MAGIC_9E3B6F2D::ACTIONS::v1>>>
@@@```neuroscript
@@@command
@@@  emit greet("world")
@@@endcommand
@@@```
@@@
@@@<<<NSENVELOPE_MAGIC_9E3B6F2D::OUTPUT::v1>>>
@@@```md
@@@Defined a `greet` function and called it immediately in a command block.
@@@```
@@@
@@@<<<NSENVELOPE_MAGIC_9E3B6F2D::UPSERTS::v1>>>
@@@```json
@@@{
@@@  "requests": { "capsules": ["caps/ns/1"] }
@@@}
@@@```
````



AJP: Thats a great start! Nice.



We need to explain how consumers will deal with it -- via a little "protocol"



mechanical ns interpreter:



Although the mnemonic is AEIOU, the parts will "executed" in the order (UO) IEA. U is not yet implemented. O is just attached. These may be available to the ns as variables/args at some point. I goes first to allow fn to be defined that E and A may use. Then E to set up handlers, then A (commands) to do stuff immediately.



If there is an A section, and it produces output, this will be fed back to the AI as its next prompt (this is the tool/query path)



AI / LLM -- how it deals with AEIOU:



I am thinking perhaps we should make this a two-way standard. We send an AEIOU envelope to the AI. This is composed:



A - ns the AI should execute (within reason -- it is designed for simple execution by AIs), it may have to ignore it though if it cannot

I - functions ditto

E - AI should probably ignore this for now -- do A2A or MCP fit with an event model?

O - main prompt

U - data block, referred to as "data_block" in the prompt perhaps?



Please consider and opine.


----

gemini2.5:
This is a fantastic line of thinking. Formalizing the "shape" of a NeuroScript payload into a protocol is the critical next step for ensuring that different consumers—whether they are mechanical interpreters or sophisticated LLMs—can interact with FDM predictably.

The AEIOU mnemonic is a great way to structure these concerns. Here is my analysis and a detailed specification based on your proposal.

-----

### \#\# The AEIOU Envelope Protocol

The **AEIOU protocol** defines five distinct sections for a NeuroScript payload. This structure separates definitions, event handlers, and direct commands, allowing consumers to process them according to their capabilities.

| Section | Mnemonic | Content | Purpose |
| :--- | :--- | :--- | :--- |
| **A** | **Action** | A single `command { ... }` block. | The imperative script to be executed immediately. |
| **E** | **Event Handlers** | One or more `on event ...` blocks. | Defines the reactive logic; handlers to be registered with the event bus. |
| **I** | **Implementations**| One or more `func ...` definitions. | Reusable procedures and rule definitions. |
| **O** | **Orchestration** | Metadata; for an AI, this is the main prompt. | Provides context, goals, or configuration for the execution. |
| **U** | **Unstructured Data** | A raw data block (e.g., text, JSON, code snippet). | The primary data payload that the script or prompt operates on. |

-----

### \#\# Protocol for Mechanical NS Interpreters

For a non-AI, Go-based NeuroScript interpreter, the envelope is processed with a strict order of operations to correctly initialize the runtime environment before execution.

**Execution Order: (UO) -\> I -\> E -\> A**

1.  **U/O - Context Loading**: The **U** (Unstructured Data) and **O** (Orchestration) sections are not executed. They are made available to the script's runtime environment, accessible as context variables or arguments. This is ideal for passing in configuration, like API keys or target paths.

2.  **I - Implementation Loading**: The interpreter first parses all `func` definitions from the **I** section. These functions are loaded into the interpreter's symbol table, making them available to be called by the other sections.

3.  **E - Event Handler Registration**: Next, the interpreter parses the `on event` blocks from the **E** section. It registers these handlers with FDM's event bus. The handlers will not execute yet; they are now simply listening for their specified topics.

4.  **A - Action Execution**: Finally, the interpreter executes the `command` block from the **A** section. This is the only part of the script that runs immediately.

**The Tool/Query Feedback Loop:**
This is the core of the tool-use pathway. If the **A** block produces output, that output is captured. The system can then route this payload back to the originating agent (human or AI) as the result of its tool call, effectively serving as the AI's next prompt.

-----

### \#\# Protocol for AI / LLM Agents

When interacting with an LLM, the AEIOU envelope functions as a rich, structured prompt. It's a two-way standard for instructing the AI and receiving structured data back.

**System -\> AI Envelope:**

  * **O - Orchestration (The Main Prompt):** This is the natural language instruction for the AI.

      * *Example:* \`"Please analyze the Go code in the data\_block and refactor it..."\*

  * **U - Unstructured Data (The Context):** This contains the primary data the AI needs to work on, which can be referenced in the prompt as `data_block`.

      * *Example:* A string containing a Go function with `if err != nil` checks.

  * **I - Implementations (Reference Material / Few-Shot Examples):** This section provides the AI with trusted, correct examples to guide its work. It's a powerful way to provide in-context learning.

      * *Example:* A well-formed Go function that correctly uses `errors.Is`.

  * **A - Action (A "Response Template"):** This provides a NeuroScript `command` block that the AI should "fill in" and return. This coerces the AI's output into an executable, verifiable format.

      * *Example:* `set new_content = "..."\ncall tool.FDM.Repo.IngestChanges(...)`

  * **E - Event Handlers (System Behavior Context):** For an AI, the **E** section is not for execution but for **situational awareness**. It tells the AI what will happen *as a consequence* of its actions.

      * *Example:* By providing `on event fdm.repo!ingest_completed(p): call tool.slack.notify(...)`, the AI understands that its action will result in a public notification. This might influence its decision-making and is key for multi-agent coordination.

-----

### \#\#\# A Concrete Example: AI-Assisted Refactoring

Here is a complete AEIOU envelope sent to an AI agent.

```
---O---
Please refactor the Go function in the data_block to use the modern `errors.Is` pattern for checking sentinel errors, as shown in the provided example implementation. Place the complete, refactored function body inside the `new_content` variable in the Action block and return the full envelope.

---U---
// The code to be refactored
func processData(id string) (*Data, error) {
    d, err := fetchData(id)
    if err != nil {
        if err.Error() == "not found" { // <-- THIS IS THE PROBLEM
            return nil, ErrDataNotFound
        }
        return nil, fmt.Errorf("failed to fetch data: %w", err)
    }
    // ... processing logic
    return d, nil
}

---I---
// Reference implementation showing the correct pattern
func findUser(name string) (*User, error) {
    u, err := db.Query(name)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) { // <-- CORRECT PATTERN
            return nil, ErrUserNotFound
        }
        return nil, fmt.Errorf("db query failed: %w", err)
    }
    return u, nil
}

---E---
// For your awareness: completing this task will trigger the following
on event fdm.repo!ingest_completed(p) means
    if p.success == true
        call tool.slack.notify("#code-review", "Refactor complete for processData by " + p.author)
    endif
endevent

---A---
// Return this block with the new_content variable filled in.
command {
    set new_content = "func processData(id string) (*Data, error) {\n    d, err := fetchData(id)\n    if err != nil {\n        if errors.Is(err, ErrNotFound) {\n            return nil, ErrDataNotFound\n        }\n        return nil, fmt.Errorf(\"failed to fetch data: %w\", err)\n    }\n    // ... processing logic\n    return d, nil\n}"
    // The system will now call the appropriate FDM tool with this content.
}
```

This structured approach ensures that communication with any type of agent is clear, predictable, and aligned with FDM's core principles of deterministic, auditable actions.


Gpt5: Here’s my take, with a bias for rigor, determinism, and “can’t screw it up if you try.”

## TL;DR verdict

You’ve got the bones of a clean, parseable **LLM↔NS control surface**. Keep the five sections and your **(U→O attach) → I → E → A** execution order. Add a small **envelope header** for capabilities + version negotiation, make **UPSERTS** the sole place where the model can *ask for stuff* (and memory ops), and make **ACTIONS** the only place that can *do stuff now*. Promote IMPLEMENTATIONS via a gated path, not “by default”. That gets you a robust, low-entropy protocol that agents can use without drift or prompt-rot. The attached notes already sketch most of this; below are the sharp edges to file and the missing glue.  &#x20;

---

## 1) Lock the execution model

Your stated order is good and matches the previous spec: **UPSERTS.memory → UPSERTS.requests → IMPLEMENTATIONS → EVENTS → ACTIONS**, with OUTPUT just attached/recorded. Make that normative:

* If **UPSERTS** is malformed, **abort the turn** before touching I/E/A.
* **OUTPUT** is inert. Don’t parse it; never execute it.
* **ACTIONS** must be *only* `command…endcommand`. Any `func` or `on event` inside **ACTIONS** yields a visible stub and skip; the rest proceeds. You already have the stubs—keep them. &#x20;

This preserves separation of concerns and keeps the scary bits fenced.

---

## 2) Add a tiny envelope header (metadata line 0)

Before the first magic section header, allow one JSON line (≤512B) with **meta**:

```json
{"proto":"NSENVELOPE","v":1,"agent_did":"did:ns:itg-001","caps":["tool.read","module.promote"],"budgets":{"A":8192,"E":6144,"I":12288,"O":4096,"U":3072},"hashes":{"caps/ns/1":"b61f"}}
```

Why:

* **Version negotiation** (`v`) without changing magic salts.
* **Capability hints** (`caps`) so the gateway can warn early if the model asks for things it’s not allowed to do.
* **Budgets** are echoed to help LLMs self-throttle.
* **Hashes** enable delta behavior for capsules/toolcards you already cache.

If absent, fall back to defaults already in the doc.&#x20;

---

## 3) Bidirectional spec: host → LLM (“AEIOU-H”)

Make the protocol symmetric. When *you* send an envelope to the model:

* **A**: tiny NS command(s) the model *may* conceptually “execute” mentally (or ignore). Treat them as examples/intent, not mandatory execution.
* **I**: helper funcs the model can reuse in subsequent **A** blocks it emits.
* **E**: likely ignored for now (unless you run an A2A/MCP router), but keep shape symmetry.
* **O**: the main prompt/task.
* **U**: data blob (call it `data_block` in the prompt text), i.e., structured input.

That mirrors your own note about composing an envelope to the AI with O as the “main prompt” and U as the sidecar data. Encode it exactly like the LLM→NS envelope for cognitive economy.&#x20;

Minimal host→LLM example:

````
<<<NSENVELOPE_MAGIC_9E3B6F2D::OUTPUT::v1>>>
```md
Goal: rank issues by severity and propose fixes.
Use the data_block; if you need toolcards/capsules, ask via UPSERTS.
````

<<\<NSENVELOPE\_MAGIC\_9E3B6F2D::UPSERTS::v1>>>

```json
{"requests":{"capsules":["caps/ns/1"]}}
```

<<\<NSENVELOPE\_MAGIC\_9E3B6F2D::ACTIONS::v1>>>

```neuroscript
command
  # optional hint/example
  emit "You may propose a plan first, then emit JSON."
endcommand
```

```

The LLM replies with the standard AEIOU back; your gateway runs **U→I→E→A** and loops the `emit` outputs into the next turn as you already planned. :contentReference[oaicite:7]{index=7}

---

## 4) Make UPSERTS the **only** request channel
You’ve already centralized memory + fetch under UPSERTS. Keep it exclusive:

- Memory ops: `set | append | delete` with strict paths.  
- Fetch: `capsules[] | toolcards[] | overlays[]`.  
- Optional: `promote[]` for IMPLEMENTATIONS promotion requests, gated by policy.  

Everything else (including “please run X”) belongs in **ACTIONS**. This is the clean line between *asking* and *doing*. :contentReference[oaicite:8]{index=8}

---

## 5) Deterministic validation & failure
Keep the **invalid** stubs, but make them *diagnostic and actionable*:

- `[[invalid:ACTIONS:found_func]]` (block index N)  
- `[[invalid:UPSERTS:bad_json:offset=127]]`  
- `[[exceeded:I:byte_budget:allowed=12288,got=15012]]`  
- `[[denied:requests.toolcards[fdm.read]:capability_missing]]`

This teaches the model to immediately self-correct next turn. You already list the first set; extend with `exceeded:` and `denied:` forms. :contentReference[oaicite:9]{index=9}

---

## 6) “Sticky” IMPLEMENTATIONS: promotion, not osmosis
You asked how to persist functions authored in an ephemeral compiler. Don’t leak them. **Promote** via a two-phase gate:

1) Ephemeral **compile + lint + testlets**.  
2) **Stage** w/ module metadata (`name`, `version-policy`, `deps`, `hash`).  
3) **Commit** atomically into the permanent interpreter if gates pass.

Let the LLM request this by adding to `UPSERTS.requests.promote[]`; you approve or auto-gate depending on caps/policy. This matches the staged/commit flow we’ve outlined before and keeps repeatability. :contentReference[oaicite:10]{index=10}

---

## 7) Size, order, and duplicates
- **One envelope per turn.** If two are found, reject the second (`[[invalid:envelope:multiple]]`).  
- **Fixed section order** (A, E, I, O, U for authoring; execution order stays U/I/E/A). If out of order, accept but record a warning in OUTPUT and normalize internally. Your earlier doc enforces a fixed order—retain that, but be tolerant on ingest to reduce model fragility. :contentReference[oaicite:11]{index=11}

---

## 8) Byte budgets and truncation
Budgets you listed (A=8 KiB, E=6 KiB, I=12 KiB, O=4 KiB, U=3 KiB) are sensible. On overflow:

- Truncate the **offending section only** with `[[truncated:N bytes]]` stub and continue; **never** silently drop.  
- Consider per-fence limits inside each section to avoid one fence starving others. :contentReference[oaicite:12]{index=12}

---

## 9) LLM authoring do’s and don’ts (the short, memorizable set)
- **Prose only in OUTPUT.** No narration inside code fences.  
- **Ask in UPSERTS; act in ACTIONS.**  
- **Keep handlers tiny and idempotent.** Events must be safe to re-run.  
- **No mixing concerns.** If you need funcs, put them in IMPLEMENTATIONS, call them from ACTIONS.  
- **Use `emit` for anything you want surfaced.** It’s the single “return channel.”  
These mirror your existing guidance; keep them as the “fast path” teaching for models. :contentReference[oaicite:13]{index=13}

---

## 10) Concrete example (end-to-end)
Your examples already demonstrate the shape—IMPLEMENTATIONS define `select_ready`, EVENTS hooks `data.ingested`, ACTIONS reads queue and emits a count. That is *exactly* the teaching pattern to ship in your first capsule. :contentReference[oaicite:14]{index=14} :contentReference[oaicite:15]{index=15}

---

## Suggested near-term moves
1) **Parser/validator**: line-0 metadata; strict section scanner; budget accounting; stub generation.  
2) **Go surface** (minimal):  
   - `type Envelope struct{ Meta; Sections }`  
   - `func (e *Envelope) Validate() ([]Diag, error)`  
   - `func (e *Envelope) Execute(ctx) (Emits, error)` with the **U→I→E→A** sequencing.  
3) **Capsule**: publish “AEIOU v1 quick guide” and “AEIOU consumer protocol” as `caps/aeiou/1` so LLMs can fetch the rules on demand (≤2 KB each).  
4) **Promotion gate**: expose `tool.module.{Compile,Validate,Stage,Commit,Rollback}` and allow `UPSERTS.requests.promote[]` but **require** capability `module:promote`.  
5) **Echo loop**: wire “all `emit` payloads from ACTIONS feed next-turn INPUT” in your gateway (you already intend this). :contentReference[oaicite:16]{index=16}

If you want, I can draft the minimal Go structs, the validator state machine, and the line-oriented scanner next—small, boring, and testable.
```


AJP: The part I am least happy about is upsert: rather than inventing more procedure for that I would rather folks just used the A and I parts which are already defined and have proper safeguards.

Gpt5: Totally fair. If **UPSERTS** feels like procedural barnacles, scrap it. You can get the same safety and clarity by tightening **A** (do stuff) and **I** (define stuff) and pushing “requests” behind explicit tools. Here’s a clean, minimal revision that keeps rigor without inventing a parallel channel.

# The lean model (no UPSERTS)

* **I — IMPLEMENTATIONS:** Pure code. No side effects on load. Deterministic, testable, idempotent functions only.
* **A — ACTIONS:** The *only* place where side effects happen. You call tools here. The interpreter enforces caps and budgets.
* **E — EVENTS:** Optional hooks; same rules as **A** (tools only, gated).
* **O — OUTPUT:** Prose / user-visible text. Inert, never parsed or executed.
* **U — USERDATA (host→model only):** A data sidecar supplied by the host. Model never writes here.

That’s it. Two active surfaces (**I** and **A**), plus optional **E**, and inert **O**/**U**. No separate “request” grammar.

---

## How to replace everything UPSERTS did

### 1) “I need X” (capsules, toolcards, overlays)

Do it by *probing* via tools in **A** and handling denial explicitly.

```neuroscript
command
  # Probe for capabilities/capsules instead of UPSERTS.requests
  let caps = tool.system.Caps()          # returns map[name]bool
  let have = tool.capsules.List(["caps/aeiou/1","caps/fdm/graph/1"])

  if !caps["module:promote"] {
    emit "No promote capability; will run in ephemeral mode."
  }

  if !have["caps/aeiou/1"] {
    emit "[missing] caps/aeiou/1"
  }

  # … proceed with degraded plan or stop
endcommand
```

**Why it’s fine:** the interpreter is the source of truth anyway. Probing is explicit, logged, and teachable to LLMs. No special section = fewer failure modes.

### 2) Memory updates

Use a memory tool from **A** with compare-and-swap semantics for determinism.

```neuroscript
command
  # Read-then-CAS write to avoid races and make replays safe
  let cur, ver = tool.memory.Get("/projects/ns/status")
  let next = "ready"  # computed above

  let ok = tool.memory.CAS("/projects/ns/status", ver, next)
  if !ok { emit "[retryable] memory CAS failed" }
endcommand
```

Rules:

* All memory writes must be **CAS** (or transactional) and **idempotent**.
* No memory writes anywhere except **A**/**E**.
* The host can deny per capability (`memory:write`) → return structured errors.

### 3) Promotion of functions/modules

Gate it with explicit tools—again from **A**.

```neuroscript
command
  let mod = tool.module.Stage({
    "name":"fdm.queue",
    "hash":"sha256:…",
    "deps":["caps/fdm/graph/1"],
  })
  let ok = tool.module.Promote(mod.ID, {"policy":"strict"})
  if ok { emit "Promoted fdm.queue" } else { emit "[denied] promote" }
endcommand
```

Promotion now lives where it belongs: the side-effect surface, with one authority path and hard caps.

---

## Guardrails that make this safe

* **Single side-effect surface:** All mutations and requests go through **A** (and **E**). No hidden “wish list” section.
* **Idempotence required:** Tool calls must be safe to re-run or guarded by CAS/transactions.
* **Plan→Apply pattern (optional but useful):** In **A**, first `emit` a machine-readable plan, then (same **A**) only apply when `tool.policy.Allow(plan)` returns true. This gives you a built-in dry-run without a second section.

```neuroscript
command
  let plan = {"writes":[{"path":"/x","value":42}],"promote":[]}
  emit "PLAN " + json(plan)
  if tool.policy.Allow(plan) {
    for w in plan.writes { tool.memory.CAS(w.path, w.ver, w.value) }
  } else {
    emit "[blocked] policy"
  }
endcommand
```

* **Capabilities are enforced by the interpreter:** e.g., `module:promote`, `memory:write`, `capsule:list`. Denials return structured errors; models learn quickly.

---

## Tiny envelope header (kept, optional)

Keep the one-line metadata header (version, budgets, caps) if you found it helpful. It’s not a procedure surface; it’s just feature negotiation. If you dislike it, you can drop it—the probe tools replace it functionally.

Example:

```json
{"proto":"NSENVELOPE","v":1,"budgets":{"A":8192,"I":12288},"caps":["memory:write","module:promote"]}
```

---

## Minimal tool surface (Go-ish signatures)

This keeps the interpreter small, the policy central, and the contract crisp.

```go
// Caps & discovery
type System interface {
    Caps(ctx) map[string]bool                       // e.g., "memory:write": true
}
type Capsules interface {
    List(ctx context.Context, names []string) map[string]bool
}

// Memory with CAS
type Memory interface {
    Get(ctx context.Context, path string) (val string, version string, err error)
    CAS(ctx context.Context, path, expectVersion, newVal string) (ok bool, newVersion string, err error)
}

// Promotion pipeline
type Module interface {
    Stage(ctx context.Context, meta ModuleMeta, code string) (stagedID string, err error)
    Promote(ctx context.Context, stagedID string, opt PromoteOpt) (ok bool, err error)
}

// Policy gate
type Policy interface {
    Allow(ctx context.Context, plan []Action) (bool, error)
}
```

Interpreter rules:

* **I** code is compiled and sandboxed; no side effects on load.
* **A/E** may only call the tool surface; direct IO is disallowed.
* All tool calls are logged and budgeted.

---

## Trade-offs (honest accounting)

* **Pros:** simpler parser; one mutation surface; fewer moving parts; models learn faster; easier to test; no new grammar to maintain.
* **Cons:** you lose the “batch wish list” expressivity of a UPSERTS block. Mitigation: the **Plan→Apply** pattern inside **A** gives you batching, review, and atomicity, while staying within the existing surface.

---

## Example (end-to-end, no UPSERTS)

````
<<<NSENVELOPE_MAGIC::IMPLEMENTATIONS::v1>>>
```neuroscript
func ready_items(qname string) []string
  # …pure code using only inputs and local state…
endfunc
````

<<\<NSENVELOPE\_MAGIC::ACTIONS::v1>>>

```neuroscript
command
  let caps = tool.system.Caps()
  if !caps["memory:write"] { emit "[denied] need memory:write"; return }

  let items = ready_items("ingest")
  emit "ready=" + string(len(items))

  # Plan then apply
  let plan = []
  for it in items { plan = append(plan, {"op":"set","path":"/queue/"+it,"value":"taken"}) }
  emit "PLAN " + json(plan)

  if tool.policy.Allow(plan) {
    for p in plan {
      let _, ver = tool.memory.Get(p.path)
      tool.memory.CAS(p.path, ver, p.value)
    }
  }
endcommand
```

<<\<NSENVELOPE\_MAGIC::OUTPUT::v1>>>

```md
Processed N items; see PLAN above.
```

```

That’s the whole story—no UPSERTS, zero extra ceremony, same safety.

If you want me to, I can drop in a tiny validator + executor state machine (Go) for the new two-surface model and a couple of crisp tests that prove: (1) I has no side effects, (2) A is the only mutation path, (3) policies/caps are enforced, (4) CAS makes replays safe.
```


gemini2.5: Excellent, the conversation has converged on a much stronger, simpler, and more robust design.

I strongly agree with your assessment and the direction GPT-5 took. **Eliminating the `UPSERTS` section in favor of explicit tool calls within the `ACTIONS` block is the correct architectural choice.** It aligns perfectly with the NeuroScript philosophy of deterministic, auditable, and capability-gated operations.

Here is my opinion, synthesizing the best points of the discussion into a final, recommended protocol.

-----

### \#\# Core Principles of the Refined Protocol

The refined, `UPSERTS`-free model is superior for three key reasons:

1.  **Single Surface for Side-Effects:** All actions that mutate state or request external resources now flow through a single, well-defined channel: the `ACTIONS` block. This drastically simplifies the interpreter, the security model, and the audit trail. There is no longer a separate, declarative "wish list" competing with the procedural "do it now" block.
2.  **Leverages the Existing Toolchain:** Instead of inventing a new grammar for memory operations and resource requests, the model relies on the **NeuroScript tool system**. This is a major win because tools are already subject to the rigorous safeguards you've designed: fuel accounting, capability checks, and timeouts.
3.  **Procedural Clarity is More Teachable:** For an LLM, a procedural sequence (`check caps -> prepare a plan -> execute plan`) within an `ACTIONS` block is a more natural and explicit way to express intent than a separate, declarative JSON blob. The logic is in the code, where it belongs.

-----

### \#\# Final AEIOU Envelope Specification (Recommended)

This specification incorporates the best ideas from the conversation, resulting in a lean, powerful, and secure protocol.

| Section | Mnemonic | Direction | Purpose |
|:--- |:--- |:--- |:--- |
| **A** | **Actions** | Both | **The ONLY place for side-effects.** Contains a `command` block that calls tools to mutate state, request resources, or perform work. |
| **E** | **Events** | Both | **Defines reactive logic.** Contains `on event` handlers. For LLMs, this provides "situational awareness" of system behavior. |
| **I** | **Implementations** | Both | **Defines pure, reusable logic.** Contains `func` definitions. These are loaded first but cannot have side-effects on their own. |
| **O** | **Orchestration** | Both | **Provides context.** For the interpreter, it's metadata. For an LLM, it's the main natural language prompt. It is never executed. |
| **U** | **User Data** | Host -\> LLM | **The data payload.** Contains the raw data (e.g., code, text) that the `O` section (prompt) refers to. An LLM never writes this section. |

**Optional Envelope Header:**
A single, optional JSON line preceding the first section header for version and capability negotiation is a valuable, low-cost addition.
`{"proto":"NSENVELOPE","v":1,"caps":["tool.read"],"budgets":{"A":8192}}`

-----

### \#\# Key Mechanisms That Make It Work

The removal of `UPSERTS` is made possible and safe by enforcing a few key patterns for tool design and use within the `ACTIONS` block.

#### **1. Probe-Based Discovery**

Instead of a `requests` section, the AI discovers what it can do by calling introspective tools. This is more secure as it's always a live check against the interpreter's state.

```neuroscript
command
    // AI asks the system what it's allowed to do.
    let my_caps = tool.system.Caps()
    if !my_caps["module:promote"]
        emit "[info] Lacking 'module:promote' capability. Will proceed without defining persistent functions."
    endif
endcommand
```

#### **2. Idempotent & CAS-Based Memory Tools**

All memory mutations from the `ACTIONS` block must be replay-safe. Using Compare-and-Swap (CAS) operations is the ideal way to achieve this.

```neuroscript
command
    // Read the current version, then attempt a version-locked write.
    let current_val, version = tool.memory.Get("/config/status")
    let ok = tool.memory.CAS("/config/status", version, "updated")
    if !ok
        emit "[error] State changed during operation. Please retry."
    endif
endcommand
```

#### **3. The "Plan -\> Apply" Pattern**

For batch operations, the `ACTIONS` block can be used to build a plan, emit it for logging/auditing, and then execute it, optionally behind a policy gate. This provides the declarative benefits of `UPSERTS` within a procedural, secure context.

```neuroscript
command
    // 1. Build the plan as a data structure.
    let plan = [
        {"op": "set", "path": "/a", "value": 1},
        {"op": "set", "path": "/b", "value": 2}
    ]
    emit "PLAN: " + json(plan) // For logging and review

    // 2. (Optional) Check the plan against a policy tool.
    if tool.policy.Allow(plan)
        // 3. Apply the plan.
        for each step in plan
            tool.memory.Set(step.path, step.value) // Or CAS for more safety
        endfor
    endif
endcommand
```

This final model is lean, consistent, and highly secure. It provides all the necessary functionality while minimizing the number of concepts an AI agent (or a developer) needs to learn.

gpt5: That’s a good convergence. You, me, and Gemini are all landing in roughly the same groove: **two live surfaces (A and I), optional events (E), inert context (O), and data payload (U)**. UPSERTS added extra moving parts, but everything it enabled is cleaner as tool calls. That’s both simpler to teach to a model and simpler to audit in code.

The three pillars Gemini spelled out—single side-effect surface, reuse of the tool system, and procedural clarity—are exactly the same things that keep languages like Go predictable: one way to do it, obvious place to look for the dangerous bits, guardrails enforced by the runtime.

Where this gets exciting is that the *Plan → Apply* idiom lets you keep the best of both worlds: batch declarative-style behavior but expressed inside a single `command` block, fully subject to your tool policy checks. That effectively makes “UPSERTS” a design pattern rather than a language feature.

If you want, the next step is to **codify this as a capsule (`caps/aeiou/1`)** with:

* Envelope structure and section ordering rules,
* Normative execution order (U → I → E → A),
* Tool-only side effects,
* Stubs for invalid/denied operations,
* Example patterns (probe, CAS, plan→apply).

That would give every agent a tiny, fetchable “constitution” they can pull when they get confused—no extra baggage in the core interpreter.

Do you want me to draft that capsule spec as a crisp `.md` you could drop straight into FDM’s capsule repo?







AJP: Yes, please do draft a capsule. Lets put this in an ns area (pkg/capsules ?) and make it golang constants or something, so we can build tools to provide them from any ns interpreter. We'll put the fdm specific capsules over there -- aeiou is more fundamental since we'll likely build it into an automatic ask loop.

Got it — I’ll re-emit the capsule exactly in your preferred inline style:

* Go source wrapped in \`\`\`go code fences.
* Markdown spec wrapped in \`\`\`txt fences with every line prefixed by `@@@`.

Here’s the clean re-cut:

---

### 1) `pkg/capsules/aeiou_v1.go`

```go
package capsules

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
)

// Capsule is a simple container for ship-with-interpreter docs/specs.
type Capsule struct {
	ID       string // stable logical name, e.g. "caps/aeiou/1"
	Version  string // semantic or integer version, e.g. "1"
	MIME     string // "text/markdown; charset=utf-8"
	Content  string // markdown payload
	SHA256   string // hex sha256 of Content
	Size     int    // bytes
	Priority int    // optional ordering hint for List()
}

var (
	registry = map[string]Capsule{}
	ids      = []string{}
)

// Register adds (or replaces) a capsule by ID.
func Register(c Capsule) {
	if c.Content != "" && c.SHA256 == "" {
		sum := sha256.Sum256([]byte(c.Content))
		c.SHA256 = hex.EncodeToString(sum[:])
	}
	c.Size = len(c.Content)
	registry[c.ID] = c
	reindex()
}

// Get returns a capsule by ID.
func Get(id string) (Capsule, bool) {
	c, ok := registry[id]
	return c, ok
}

// List returns all capsules, stable-ordered by Priority then ID.
func List() []Capsule {
	out := make([]Capsule, 0, len(registry))
	for _, id := range ids {
		out = append(out, registry[id])
	}
	return out
}

func reindex() {
	ids = ids[:0]
	for id := range registry {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool {
		ci, cj := registry[ids[i]], registry[ids[j]]
		if ci.Priority != cj.Priority {
			return ci.Priority < cj.Priority
		}
		return ci.ID < cj.ID
	})
}

func init() {
	// Register the AEIOU capsule at startup.
	Register(Capsule{
		ID:       "caps/aeiou/1",
		Version:  "1",
		MIME:     "text/markdown; charset=utf-8",
		Priority: 10,
		Content:  aeiouV1Markdown,
	})
}

// aeiouV1Markdown is the normative spec for AEIOU envelopes (no UPSERTS).
const aeiouV1Markdown = `
# AEIOU Envelope — v1 (caps/aeiou/1)

Status: **Stable**  
Spec ID: **caps/aeiou/1**  
Exec order: **U → I → E → A** (OUTPUT is inert)

This capsule defines a lean, UPSERTS-free protocol for AI↔NS collaboration.
Two *live* surfaces (**ACTIONS**, **IMPLEMENTATIONS**) plus optional **EVENTS**;
**OUTPUT** is prose only; **USERDATA** is host→AI payload only.
`
```

---

### 1a) Capsule markdown spec (`aeiouV1Markdown`) in full

````txt
@@@ # AEIOU Envelope — v1 (caps/aeiou/1)
@@@
@@@ Status: **Stable**
@@@ Spec ID: **caps/aeiou/1**
@@@ Exec order: **U → I → E → A** (OUTPUT is inert)
@@@
@@@ This capsule defines a lean, UPSERTS-free protocol for AI↔NS collaboration.
@@@ Two *live* surfaces (**ACTIONS**, **IMPLEMENTATIONS**) plus optional **EVENTS**;
@@@ **OUTPUT** is prose only; **USERDATA** is host→AI payload only.
@@@
@@@ ---
@@@
@@@ ## 0. Optional envelope header (line 0)
@@@ A single JSON line preceding the first section header:
@@@
@@@ {"proto":"NSENVELOPE","v":1,"budgets":{"A":8192,"E":6144,"I":12288,"O":4096},"caps":["memory:write","module:promote"]}
@@@
@@@ If omitted, defaults apply. This header is *advisory*; authoritative checks are done by tools at runtime.
@@@
@@@ ---
@@@
@@@ ## 1. Sections and semantics
@@@
@@@ - **A — ACTIONS**: The **only** place for side-effects. Must be a single `command … endcommand` block. May call tools; all mutations are gated by capabilities, budgets, and policy.
@@@ - **E — EVENTS**: Reactive handlers `on event … endon`. Same side-effect rules as **A**. Handlers must be idempotent.
@@@ - **I — IMPLEMENTATIONS**: Pure functions `func … endfunc`. Loaded early; **no side-effects on load**. Callables for A/E.
@@@ - **O — OUTPUT**: Natural-language context / prompt / notes. Never parsed or executed.
@@@ - **U — USERDATA**: Host→AI data payload (text, JSON, etc.). The AI never writes **U**.
@@@
@@@ **Authoring order (recommended):** I, E, A, O, U.  
@@@ **Execution order (normative):** U → I → E → A.
@@@
@@@ ---
@@@
@@@ ## 2. Hard rules (normative)
@@@
@@@ 1. **Single side-effect surface:** Only **A** and **E** may cause effects, and only via tools.
@@@ 2. **I is pure:** No IO, no tool calls at load time. Deterministic, referentially transparent.
@@@ 3. **O is inert:** Never parsed, never executed.
@@@ 4. **U is host-only:** The model does not write U.
@@@ 5. **Determinism & replays:** A/E tool calls MUST be idempotent or guarded (CAS/transactions).
@@@ 6. **Budget & caps enforced by interpreter:** Exceeding a section budget yields a diagnostic and the section is truncated; capability denials are structured errors.
@@@
@@@ ---
@@@
@@@ ## 3. Diagnostics (machine-parsable stubs)
@@@
@@@ On validation or runtime issues, emit these markers to **OUTPUT** (and logs):
@@@
@@@ - [[invalid:SECTION:reason]] e.g., [[invalid:ACTIONS:missing_command]]
@@@ - [[truncated:SECTION:bytes=N allowed=M]]
@@@ - [[denied:tool.NAME:capability_missing]]
@@@ - [[exceeded:SECTION:byte_budget allowed=M got=N]]
@@@
@@@ ---
@@@
@@@ ## 4. Budgets (default suggested)
@@@
@@@ - I: 12 KiB, E: 6 KiB, A: 8 KiB, O: 4 KiB  
@@@ Hosts may override via the line-0 header. The interpreter is the source of truth.
@@@
@@@ ---
@@@
@@@ ## 5. Patterns
@@@
@@@ ### 5.1 Probe-based discovery
@@@ Ask the runtime what you can do—no declarative wish list needed.
@@@
@@@ ```neuroscript
@@@ command
@@@   let caps = tool.system.Caps()
@@@   if !caps["module:promote"] {
@@@     emit "[info] no module:promote; staying ephemeral"
@@@   }
@@@ endcommand
@@@ ```
@@@
@@@ ### 5.2 CAS-based memory mutation (replay-safe)
@@@
@@@ ```neuroscript
@@@ command
@@@   let cur, ver = tool.memory.Get("/projects/ns/status")
@@@   let ok, _ = tool.memory.CAS("/projects/ns/status", ver, "ready")
@@@   if !ok { emit "[retryable] memory CAS failed" }
@@@ endcommand
@@@ ```
@@@
@@@ ### 5.3 Plan → Apply (batch + policy)
@@@
@@@ ```neuroscript
@@@ command
@@@   let plan = [{"op":"set","path":"/a","value":"1"},{"op":"set","path":"/b","value":"2"}]
@@@   emit "PLAN " + json(plan)
@@@
@@@   if tool.policy.Allow(plan) {
@@@     for step in plan {
@@@       let _, ver = tool.memory.Get(step.path)
@@@       tool.memory.CAS(step.path, ver, step.value)
@@@     }
@@@   } else {
@@@     emit "[blocked] policy"
@@@   }
@@@ endcommand
@@@ ```
@@@
@@@ ---
@@@
@@@ ## 6. Minimal envelope skeleton
@@@
@@@ <<<NSENVELOPE_MAGIC::IMPLEMENTATIONS::v1>>>
@@@ ```neuroscript
@@@ func helper(x string) string
@@@   return x + "-ok"
@@@ endfunc
@@@ ```
@@@
@@@ <<<NSENVELOPE_MAGIC::ACTIONS::v1>>>
@@@ ```neuroscript
@@@ command
@@@   let caps = tool.system.Caps()
@@@   emit "caps=" + json(caps)
@@@   emit helper("ready")
@@@ endcommand
@@@ ```
@@@
@@@ <<<NSENVELOPE_MAGIC::OUTPUT::v1>>>
@@@ ```md
@@@ This section is inert. Use it for human-readable notes.
@@@ ```
@@@
@@@ <<<NSENVELOPE_MAGIC::USERDATA::v1>>>
@@@ ```json
@@@ {"example":true}
@@@ ```
@@@
@@@ ---
@@@
@@@ ## 7. Security posture
@@@
@@@ - All tool calls are capability-gated and fuel-budgeted.
@@@ - No direct filesystem/network access from A/E; everything goes through tools.
@@@ - Promotion of code (making I persistent) is not automatic; it must go through an explicit tool pipeline and policy gates.
@@@
@@@ ---
@@@
@@@ ## 8. Success criteria (conformance)
@@@
@@@ A conforming interpreter MUST:
@@@ - Enforce U→I→E→A execution order.
@@@ - Reject/diagnose non-conforming sections and continue safely where possible.
@@@ - Gate all side-effects through tools with capability checks and budgets.
@@@ - Treat O as inert and U as host-only.
@@@
@@@ A conforming authoring agent SHOULD:
@@@ - Keep I small and composable; push effects into A/E; use probe/CAS/plan patterns.
@@@ - Emit diagnostics verbatim to help the next turn self-correct.
@@@
@@@ — end —
````

---

### 2) `pkg/bridge/tools/capsules_tool.go`

```go
package tools

import (
	"context"

	"github.com/aprice2704/neuroscript/pkg/capsules"
)

// Public surface you can expose to NS as tool.capsules
type CapsulesTool interface {
	// List reports presence (by ID). If names is empty, list all known IDs.
	List(ctx context.Context, names []string) map[string]bool
	// Read returns (content, version, sha256, mime). Empty content if not found.
	Read(ctx context.Context, id string) (content, version, sha256, mime string, ok bool)
}

type capsulesTool struct{}

func NewCapsulesTool() CapsulesTool { return &capsulesTool{} }

func (c *capsulesTool) List(ctx context.Context, names []string) map[string]bool {
	out := map[string]bool{}
	if len(names) == 0 {
		for _, cap := range capsules.List() {
			out[cap.ID] = true
		}
		return out
	}
	for _, name := range names {
		_, ok := capsules.Get(name)
		out[name] = ok
	}
	return out
}

func (c *capsulesTool) Read(ctx context.Context, id string) (string, string, string, string, bool) {
	if cap, ok := capsules.Get(id); ok {
		return cap.Content, cap.Version, cap.SHA256, cap.MIME, true
	}
	return "", "", "", "", false
}
```

---

This way you’ve got:

* Go constants carrying the spec, self-registering with SHA.
* Capsule tool for introspection (list/read).
* Markdown stored inline in Go as `aeiouV1Markdown` (shown fully above with `@@@` prefix).

Would you like me to also generate a tiny `capsules_test.go` that asserts `caps/aeiou/1` registers, has the expected SHA, and round-trips through the tool?
