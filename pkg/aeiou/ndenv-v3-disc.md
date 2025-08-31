Short answer: no. Only two sections are truly required for correctness; the others are operational conveniences.

* **Must-have (required):**

  * `USERDATA` — this is where the host puts the actual instructions; without it the agent has nothing to follow (and the spec says to ignore text outside the envelope). &#x20;
  * `ACTIONS` — the agent’s response must live here, and it must contain a single `command` block. &#x20;

* **Nice-to-have (optional by turn):**

  * `OUTPUT` — a public log from the **previous** turn; it is fed back from the `emit` stream. On the first turn it’s naturally empty, so it’s optional in an initial envelope. &#x20;
  * `SCRATCHPAD` — private notes from the **previous** turn; only populated via `whisper`. Also optional on the first turn.&#x20;

Practical rule of thumb for v2:

* **Minimum valid envelope:** `START … USERDATA … ACTIONS … END`. &#x20;
* **Multi-turn runs:** host will populate `OUTPUT` from your last `emit`, and `SCRATCHPAD` from your last `whisper`; include them when present, but they are not required to appear every turn. &#x20;

If you want to lock this down, add one normative line to AEIOU v2: “Exactly one `USERDATA` and one `ACTIONS` **MUST** appear; `OUTPUT` and `SCRATCHPAD` **MAY** appear zero or one times (when present from a prior turn). Duplicates are ignored after the first.” That keeps determinism high without burdening first-turn envelopes.


Short version: the “magic tool” should be the only way an agent can mint a live control token. That tool runs inside the host’s trust boundary, stitches in a per-turn nonce plus other metadata, then MAC-signs the whole thing. Anything that merely *looks* like a control token but wasn’t minted by the tool is inert text.

Here’s a concrete v3 design you can drop in.

# Goals

* Prevent spoofing (user echoing control text).
* Prevent tampering (payload cannot be altered without detection).
* Prevent replay (token from an earlier turn can’t be reused).
* Keep it ergonomic (agent never has to escape JSON or touch secrets).

# Tool interface (agent side)

* **Name:** `tool.aeiou.magic`
* **Signature:**
  `magic(kind: string, payload: any, opts?: object) -> string | error`
* **Kinds:** `"LOOP"` (and any future control channels you define).
* **Behavior:** returns a single-line UTF-8 string that is the **only** accepted control token format. Agent just `emit token`.

Example (agent):

```
let ctrl = tool.aeiou.magic("LOOP", {
  action: "continue",
  request: { min_tokens: 2048, reason: "need_more_context" },
  telemetry: { turn_index: ctx.turn_index }
})
emit ctrl
```

# Token on-wire format

Single line, no quotes, no whitespace padding:

```
<<<NSMAG:V3:{KIND}:{B64URL(PAYLOAD)}.{B64URL(TAG)}>>>
```

* `NSMAG` is a short, unambiguous prefix for v3.
* `{KIND}` mirrors the tool’s kind (e.g., `LOOP`).
* `PAYLOAD` is canonical JSON (see below).
* `TAG` is an authenticity tag (HMAC or signature).

## Canonical payload (JSON)

Use a deterministic encoding (JCS/RFC-8785 or simple lexicographic key order). Fields:

```json
{
  "v": 3,
  "kind": "LOOP",
  "jti": "f3c9e1f2-2b3c-4c71-8d5e-3b2b0f53f2a1",       // unique per token
  "session_id": "S…",                                  // host-scoped
  "turn_index": 12,                                    // current turn
  "turn_nonce": "7QqkK7L5j3tH8sX5xX9l6A",              // host-minted, per turn
  "issued_at": 1725012345,                             // unix seconds
  "ttl": 120,                                          // seconds (optional)
  "kid": "hmac-main-2025-08",                          // key id
  "payload": { "action":"continue", "request":{…}, … } // agent-supplied
}
```

## Authenticity tag (`TAG`)

Two easy, robust options:

* **HS256 (HMAC-SHA-256)** with a host-held secret, keyed by `kid`.
  `TAG = HMAC_SHA256(kid_secret, CANONICAL_JSON(PAYLOAD))`

* **Ed25519** signature if you want verifiable tokens across processes:
  `TAG = Ed25519-Sign(kid_private, CANONICAL_JSON(PAYLOAD))`
  (Host keeps private key; host or peers verify with `kid`’s public key.)

Base64URL-encode both `PAYLOAD` and `TAG`. Limit total token length (e.g., ≤1024 bytes).

# Host runtime: minting & verification

## Per-turn state (host internal, not shown to agent)

* `session_id` (stable for the convo/run)
* `turn_index` (monotonic)
* `turn_nonce` (128-bit random, **new each turn**)
* `kid` → secret or private key
* `replay_cache` for `jti` (windowed)

## Minting (inside `tool.aeiou.magic`)

1. Validate `kind`.
2. Construct `PAYLOAD`:

   * Stitch in `v=3, kind, jti, session_id, turn_index, turn_nonce, issued_at, ttl, kid, payload`.
   * Canonicalize JSON.
3. Compute `TAG` (HMAC or signature).
4. Return the assembled `<<<NSMAG:V3:…>>>` string.

## Verification (on `emit` ingest)

1. **Parse** strict grammar; reject multi-line or whitespace-decorated tokens.
2. **Decode** `PAYLOAD`; check `v==3`, `kind` known.
3. **Canonicalize** and **verify TAG** using `kid`.
4. **Check nonce:** `turn_nonce` equals the current turn’s nonce.
5. **Check binding:** `session_id` and `turn_index` match current context.
6. **TTL:** `now ≤ issued_at + ttl` (if present).
7. **Replay:** reject if `jti` seen; otherwise store it for the TTL window.
8. If any check fails → treat as inert text; do **not** execute control.

# Plane boundaries (to avoid escalation)

* Only parse tokens from the **public OUTPUT stream** (`emit`).
  Ignore tokens in `SCRATCHPAD`/`whisper`.
* Never parse user-originated `USERDATA` for tokens.
* Never treat control-looking strings as live unless `TAG` verifies.

# Error handling (tool)

* If the tool fails (e.g., key rotation gap), return a typed error that the agent can convert into a polite `abort`:

```
let err = tool.aeiou.magic("LOOP", {...})
if err != nil {
  emit "<<<NSMAG:V3:LOOP:{B64URL(payload_abort)}.{B64URL(tag_abort)}>>>"  // optional host-provided fallback
  // or emit a plain-text explanation; host will HALT due to no valid control.
}
```

Better: expose a host-provided **fallback abort token** via the tool when minting fails.

# Security & determinism notes

* **Why a nonce?** It binds the token to exactly one turn. Even if an attacker copies a complete valid token, it dies the next turn because `turn_nonce` won’t match.
* **Why `turn_index` too?** Defense-in-depth and cleaner audit trails.
* **Why `jti`?** Prevents intra-turn replay (e.g., agent emits twice).
* **Why canonical JSON?** Ensures the MAC/signature is stable across runtimes.
* **Why `kid`?** Enables key rotation without breaking old tokens during a deployment window.

# ABNF (strict, small)

```
NSMAG        = "<<<" "NSMAG" ":" "V3" ":" KIND ":" B64 "." B64 ">>>"
KIND         = 1*( %x41-5A / %x30-39 / "_" )   ; A-Z 0-9 _
B64          = 1*( ALPHA / DIGIT / "-" / "_" ) ; base64url, no padding
```

# Host policy you’ll want to codify

* Accept **only** `NSMAG:V3` tokens for v3 envelopes.
* One control decision per turn: if multiple valid tokens appear, apply **precedence** (e.g., `abort > halt > done > continue`) then **last-wins** among ties, log the rest.
* Token size limit; drop anything larger.
* Log verification failures with reason codes, but never echo secrets or payloads.

# Minimal conformance tests (add to your suite)

1. Valid token passes; wrong nonce fails.
2. Valid everything, expired TTL → inert.
3. Altered payload byte → MAC fails.
4. Right nonce, wrong session\_id → inert.
5. Duplicate `jti` in same turn → first wins, second inert.
6. Token in `SCRATCHPAD` → inert.
7. Two tokens (`continue`, then `abort`) → `abort` applied (precedence + last-wins).

# Migration / ergonomics

* Agents never see or touch the nonce; the tool handles it.
* Agents don’t escape JSON; the tool returns a final single-line token.
* Your examples can simplify to:

```
emit tool.aeiou.magic("LOOP", {"action":"done"})
```

If you want, I’ll draft the Go structs, JSON canonicalizer, HMAC/Ed25519 helpers, and the verifier (plus 10 table-driven tests) so you can wire this straight into Zadeh and your interpreter loop.




Yes—good catch. In your architecture the “agent” doesn’t execute anything. The **NS program** the agent authored is executed **by the NS interpreter inside the host**. When I said “agent does X,” read it as “the NS program (on the agent’s behalf) does X while running inside the host.”

Here’s the same flow with precise actors:

1. **Host** starts a turn and creates a fresh nonce (plus turn\_index, session\_id).
2. **NS interpreter (inside the host)** runs the agent’s submitted `ACTIONS` code.
3. That code calls **`tool.aeiou.magic`** → this tool runs **in the host**, adds the nonce/turn context, signs, and returns a one-line control token.
4. The NS code then `emit`s that token → it lands in **OUTPUT**.
5. The **host loop controller** reads **OUTPUT**, verifies the token, and applies the control (continue/done/abort).
6. New turn → new nonce. Tokens in `SCRATCHPAD` are ignored; user-authored `USERDATA` is never parsed as control.

Tiny concrete restatement (who runs what):

* **Runs code:** NS interpreter (host).
* **Mints token:** magic tool (host).
* **Decides loop:** host loop controller.
* **Stores:** OUTPUT gets `emit`ted text (including the token); SCRATCHPAD gets `whisper`; host keeps nonce/turn/replay cache.

To kill the ambiguity in v3, add this one-liner up front:

> **Execution model:** “Agent” denotes the logical author. All `ACTIONS` execute **inside the host’s NS interpreter**. Calls like `tool.aeiou.magic(...)` are host-owned; only tokens minted by that tool and `emit`ted into **OUTPUT** can drive control. `SCRATCHPAD` is never parsed for control.

If you want, I’ll draft the v3 diff that replaces “agent does …” with “NS program (in host) …” and drop that glossary note into §1.






AJP: But hang on: "That code calls tool.aeiou.magic → this tool runs in the host, adds the nonce/turn context, signs, and returns a one-line control token. The NS code then emits that token → it lands in OUTPUT. The host loop controller reads OUTPUT, verifies the token, and applies the control (continue/done/abort)." is there really any need to pipe the magic through OUTPUT? Can't the tool (called something else perhaps) mark the turn verified internally? Or did I miss a step?



You didn’t miss a step. You **can** skip piping through `OUTPUT` and flip the turn state internally—but you lose some very useful properties. Here’s the trade in plain terms.

### Two ways to signal control

1. **Token via OUTPUT (current v3 plan)**

* NS code calls `tool.aeiou.magic("LOOP", {...})` → gets a one-line token.
* NS code `emit`s that token → it shows up in **OUTPUT**.
* Host reads OUTPUT, verifies token, then applies the decision.

**Why keep this path?**

* **Single source of truth:** Everything that changes control flow is visible in the transcript. Auditable by a human, greppable by a script.
* **Reproducible replay:** You can re-run a session from the envelopes/OUTPUT alone. No hidden side effects.
* **Loose coupling:** The thing that mints (tool) and the thing that decides (loop controller) can be separate processes or even machines. The token is the contract.
* **Safety by scoping:** Only a very constrained text artifact can drive the loop; nothing “mystical” happens off-channel.

Cost: one extra line of `emit`. That’s it.

2. **Direct decide (side-effectful tool)**

* NS code calls `tool.loop.decide("continue", {...})`.
* Tool marks the host’s turn state immediately; no token, no emit needed.

**Why this is tempting?**

* Fewer moving parts, no marshalling to text.
* Lower chance the model “forgets to emit.”

**But the downsides are sharp:**

* **Hidden control path:** Now you need a *second* log (tool call audit) to explain why the host advanced. Output transcript alone is no longer authoritative.
* **Weaker replay:** Replaying envelopes won’t reproduce the run unless you also simulate tool side effects.
* **Tight coupling:** The interpreter must talk to the specific loop controller instance; harder to distribute.
* **Foot-gun:** A stray tool call can advance the loop before the agent’s reasoning is “committed” to text.

### Recommendation (clean + safe)

Keep **OUTPUT-token as the commit path**. If you want ergonomics, add a wrapper:

* `tool.loop.request("continue", payload)` → internally calls `tool.aeiou.magic(...)`, **returns the token**, and maybe caches its `jti` for faster verify.
* NS code still **must** `emit` the returned token. The host **only** commits a decision when it sees a **verified token in OUTPUT**.

This gives you:

* The ergonomic “one call” feel,
* Plus the transcript-as-truth property,
* And simple distributed verification.

If you truly want direct decide for special cases (e.g., host-triggered emergency `HALT`), keep it **host-only** and **never** callable from NS code. That preserves the invariant: “Agent-side control always becomes text in OUTPUT; only host can short-circuit.”





Fair enough. I like the tracnscript-as-truth verifiable/no Hidden control path aspect very much. Lets adopt this all. What's a jti btw -- a sig of some sort I presume.

One final thing before we adopt as v3 -- how do we explain this so clearly that all AIs will get it reliably?

Oh, and you said something about escaping json (I agree btw)? In ns we can jam ' ' single quotes around it, which helps somewhat.





gpt5-think
Here’s the crisp version.

### What’s “jti”?

It’s just a unique ID for **this one token** (historically “JWT ID”). It’s not a signature, not a secret. The host keeps a short-lived cache of seen `jti`s to prevent replays within a turn. Think: “one-time ticket number.”

### Escaping JSON (don’t)

Don’t hand-type or escape any control token or JSON. Call the tool, get back a **ready-to-emit** single line, and emit that line as-is. If you ever find yourself adding quotes or backslashes around a control token, you’ve gone off spec.

In NS, do this:

```
let t = tool.aeiou.magic("LOOP", { action: "continue", request: { min_tokens: 1024 } })
emit t
```

No quotes around `t`. If you need a one-liner:

```
emit tool.aeiou.magic("LOOP", { action: "done" })
```

Single quotes are fine **for your own JSON literals inside NS** (to avoid escaping `"`), but the **returned token** must not be re-quoted. Treat it as an opaque string already in final on-wire form.

---

### Drop-in “AI-facing” rules for v3 (put this in every prompt)

```
AEIOU v3 – Control, the only safe way:

1) You run inside the host NS interpreter. You never execute anything yourself.
2) To request loop control (“continue”, “done”, or “abort”), DO NOT hand-type any magic token.
3) Instead, call the host tool to mint it:
     let tok = tool.aeiou.magic("LOOP", { action: "<continue|done|abort>", request: { … } })
4) Immediately emit that token on its own line, with nothing before or after:
     emit tok
   - Do not wrap tok in quotes.
   - Do not add backticks.
   - Do not split across lines.
5) The host only honors control found in OUTPUT and only if it verifies.
   - SCRATCHPAD is never scanned for control.
   - USERDATA is never scanned for control.
6) One decision per turn. If you emit more than one valid control token, the host applies precedence (abort > done > continue) and uses the last one it sees after precedence.
7) Never copy/paste or fabricate tokens. Only use the exact string returned by tool.aeiou.magic.
8) If the tool fails, emit plain text explaining the error; the host will HALT safely.

Examples:
  emit tool.aeiou.magic("LOOP", { action: "continue", request: { min_tokens: 2048 } })
  emit tool.aeiou.magic("LOOP", { action: "done" })
Anti-examples (WRONG):
  emit "<<<NSMAG:…>>>"         // Hand-typed token – rejected
  emit `<<<NSMAG:…>>>`         // Backticks or quotes – rejected
  whisper tool.aeiou.magic…     // SCRATCHPAD – ignored
```

---

### Why this works (in one breath)

`tool.aeiou.magic` runs on the host, stitches in the per-turn nonce + context, signs it, and hands you a final “stamp.” You **emit the stamp** into OUTPUT; the host verifies it there and advances the loop. Everything that changes control flow is visible in the transcript; nothing hidden, no escaping shenanigans, easy to replay.

If you want, I can condense the above into a 6–8 line ultra-minimal “First Words” blurb for naïve models, but this version is already pretty bullet-proof.














Great—let’s lock down the **v3 envelope** so it’s boring, deterministic, and impossible to “interpret creatively.”

# AEIOU v3 — Envelope (tight spec)

**Purpose.** One block, between `START` and `END`, contains everything the model should read/write. Text outside is inert. This preserves the “transcript-as-truth” property you liked. (v2 already set this pattern. )

## Markers (short & strict)

* `<<<NSENV:V3:START>>>`
* `<<<NSENV:V3:USERDATA>>>`
* `<<<NSENV:V3:SCRATCHPAD>>>`  *(optional, prev turn)*
* `<<<NSENV:V3:OUTPUT>>>`      *(optional, prev turn)*
* `<<<NSENV:V3:ACTIONS>>>`
* `<<<NSENV:V3:END>>>`

**All markers are single-line, ASCII, no trailing spaces.** Content for a section runs **until the next v3 marker** or `END`. No per-section closing tags needed.

## Canonical order & cardinality (normative)

* Exactly **one** `USERDATA`, exactly **one** `ACTIONS`.
* `SCRATCHPAD` and `OUTPUT` **may** appear 0–1 times each (when present from the prior turn).
* **Order is fixed:** `USERDATA`, `SCRATCHPAD?`, `OUTPUT?`, `ACTIONS`.
* If duplicates appear, the host **uses the first** and **ignores the rest** (log a warning).
  This tightens what v2 only implied about sections and their roles.&#x20;

## Section semantics (unchanged intent, clearer rules)

* **USERDATA**: Host payload (usually JSON). The *instructions* live in values; don’t hunt outside the envelope.&#x20;
* **SCRATCHPAD**: Private notes from prior `whisper`. **Never parsed for control.**&#x20;
* **OUTPUT**: Public log from prior `emit`. **Inert context** until you parse it. **Only control tokens minted by the magic tool and emitted here can drive the loop.**&#x20;
* **ACTIONS**: The only actionable part. Must contain **one** valid `ns` `command` block.&#x20;

## Encoding & whitespace

* UTF-8, `\n` newlines.
* Host strips BOMs and trailing whitespace **on markers only**; section content is byte-preserved.

## Size guards (host)

* Envelope ≤ 1 MiB by default; each section ≤ 512 KiB; any single `emit` line ≤ 8 KiB. Oversize → reject turn with `HALT(reason:"envelope-too-large")` (host signal pattern follows v2. )

## Control path (recap)

* Agent code (running **inside** the host NS interpreter) calls `tool.aeiou.magic("LOOP", {...})`, gets a **ready-to-emit** one-liner, and **emits** it.
* Host only honors control **found in OUTPUT** and only if the token verifies. (This keeps the audit trail and replay guarantees v2 established. )

## ABNF (markers & layout)

```
ENVELOPE   = START LF USERDATA [LF SCRATCHPAD] [LF OUTPUT] LF ACTIONS LF END
START      = "<<<NSENV:V3:START>>>"
USERDATA   = "<<<NSENV:V3:USERDATA>>>" LF *OCTET
SCRATCHPAD = "<<<NSENV:V3:SCRATCHPAD>>>" LF *OCTET
OUTPUT     = "<<<NSENV:V3:OUTPUT>>>" LF *OCTET
ACTIONS    = "<<<NSENV:V3:ACTIONS>>>" LF *OCTET
END        = "<<<NSENV:V3:END>>>"
OCTET      = %x00-FF ; content runs until next marker line or END
LF         = %x0A
```

## Minimal valid example (first turn)

```text
<<<NSENV:V3:START>>>
<<<NSENV:V3:USERDATA>>>
{"subject":"onboard-001","brief":"Ack and request docs","fields":{"goal":"bootstrap"}}
<<<NSENV:V3:ACTIONS>>>
command
  emit "ACK AEIOUv3 | subject: onboard-001 | status: bootstrapping"
  whisper self, tool.docs.getcapsule("ns/latest/short")
  emit tool.aeiou.magic("LOOP", {'control':'continue','notes':'Plan next turn'})
endcommand
<<<NSENV:V3:END>>>
```

Notes:

* Using single quotes **inside NS** is fine for your JSON literals (avoids escaping), but **never** add quotes around the **returned token**; it’s already final-form. v2 trained this habit; v3 makes it explicit.&#x20;

## Multi-turn example (with prior streams)

```text
<<<NSENV:V3:START>>>
<<<NSENV:V3:USERDATA>>>
{"subject":"apply-plan","brief":"Apply approved plan"}
<<<NSENV:V3:SCRATCHPAD>>>
# your private crumbs from last turn
<<<NSENV:V3:OUTPUT>>>
PLAN: [{"op":"set","path":"/q/x","value":"taken"}]
<<<NSENV:V3:ACTIONS>>>
command
  let ok = tool.memory.CAS("/q/x","expected_ver","taken")
  if !ok {
    emit tool.aeiou.magic("LOOP", {'control':'abort','reason':'cas-failed'})
  } else {
    emit tool.aeiou.magic("LOOP", {'control':'done','notes':'applied'})
  }
endcommand
<<<NSENV:V3:END>>>
```

This mirrors the v2 two-turn example, but with v3 markers and the magic tool front-and-center.&#x20;

---

## Ultra-compact “AI-facing” blurb (drop into every prompt)

```
Read only between START/END.
Sections appear once in this order: USERDATA, [SCRATCHPAD], [OUTPUT], ACTIONS.
Only ACTIONS is executable. Write exactly one ns `command` block there.
SCRATCHPAD/OUTPUT are last turn’s notes/logs. They’re inert unless you parse them.
To control the loop, DO NOT hand-type tokens. Emit ONLY what the magic tool returns:
  emit tool.aeiou.magic("LOOP", {'control':'continue'|'done'|'abort', ...})
Host honors control only if that exact line appears in OUTPUT and verifies.
```

If you want, I can draft a tiny conformity checker: it reads a blob, validates marker order/cardinality, enforces size limits, and extracts `USERDATA`/`ACTIONS` spans—handy for unit tests and rejecting malformed envelopes early.











You’ve got it: **USERDATA is the host→program boundary.** It’s the input slot the host controls, and the only place “the boss speaks.” Everything else is either the program’s code (`ACTIONS`) or the program’s outputs (`OUTPUT`/`SCRATCHPAD`). Keeping that lane clean buys you safety, replayability, and sane debugging.

Here’s the idea in concrete, no-magic terms.

### What the boundary is

Think OS pipes:

* **USERDATA ≈ stdin** (host writes, program reads; program can’t overwrite it)
* **ACTIONS ≈ the program itself** (what gets executed)
* **OUTPUT ≈ stdout** (program emits public stuff)
* **SCRATCHPAD ≈ a private log** (program emits private stuff)

The host **only** writes USERDATA this turn. The program **never** writes there.

### How v3 enforces it in one turn

1. **Host constructs the envelope** with:

   * `USERDATA` = fresh JSON instructions for this turn.
   * `SCRATCHPAD`/`OUTPUT` = readbacks from last turn (if any).
   * `ACTIONS` = blank slot the model must fill with a single `command … endcommand`.
2. **Model fills `ACTIONS`**. That’s the only section it can change pre-execution.
3. **Host runs `ACTIONS`** in the NS interpreter:

   * `emit` writes to **OUTPUT**,
   * `whisper` writes to **SCRATCHPAD**,
   * `tool.aeiou.magic(...)` returns the ready-to-emit control token.
4. **Host accepts control only** if the **verified** magic token appears in **OUTPUT** (and it must be the last line you emit).

At no point can the program “edit” USERDATA. It can only **read** it (e.g., `userdata.fields.plan_id` inside its logic).

### What goes wrong without this boundary

* **Instruction forgery:** If instructions lived anywhere the program writes, it could “rewrite the boss,” claiming the host told it to do X. With USERDATA read-only, that’s impossible.
* **Ambiguous provenance:** Interleaving host text and program text makes audits murky. USERDATA tells you exactly “what was asked,” separate from “what was done/said.”
* **Replay pain:** Reproducing a run requires a stable input blob. Hashing USERDATA gives you that; intermingled text doesn’t.
* **Hidden control vectors:** You want one place to **read** instructions (USERDATA) and one place to **look** for control (OUTPUT). Splitting them eliminates accidental triggers.

### Rules to bake into v3 (tight, testable)

* **Host-only writer:** Only the host may populate `USERDATA`. The program must treat it as immutable input.
* **JSON, not prose:** `USERDATA` **MUST** be a JSON object. The host can validate schema, and tools can safely plumb IDs (“capsules”, URIs) without scraping text.
* **Never parsed for control:** Host **MUST NOT** scan `USERDATA` for magic tokens. Control is **only** in `OUTPUT`.
* **Fixed order & singletons:** Exactly one `USERDATA`, exactly one `ACTIONS`, optional single `SCRATCHPAD`/`OUTPUT`, in that order. Duplicates ignored after the first.
* **Token-last commit:** The control token (from `tool.aeiou.magic`) **MUST** be the last line emitted this turn.

### Tiny example showing the boundary at work

```text
<<<NSENV:V3:START>>>
<<<NSENV:V3:USERDATA>>>
{"subject":"apply-plan","fields":{"plan_id":"P-42"}}
<<<NSENV:V3:ACTIONS>>>
command
  let id = 'P-42'  // read from USERDATA.fields.plan_id in real code
  let ok = tool.plan.apply(id)
  if !ok {
    emit tool.aeiou.magic("LOOP", {'control':'abort','reason':'apply-failed'})
  } else {
    emit tool.aeiou.magic("LOOP", {'control':'done'})
  }
endcommand
<<<NSENV:V3:END>>>
```

* Only the host could have written `{"plan_id":"P-42"}` into `USERDATA`.
* The program can’t alter it; it can only act on it and then commit via the **last** emitted token.

### On JSON & escaping

* Keep `USERDATA` as **JSON** so UIs and validators can reason over it.
* Inside NS code, use **single quotes** for your own JSON literals to avoid backslash soup.
* Never quote or re-escape the **returned control token**; emit it raw on one line:

  ```
  emit tool.aeiou.magic("LOOP", {'control':'continue', 'notes':'need more context'})
  ```

If you want to make the boundary even more obvious to naïve models, add one sentence near the top of the spec: *“Only the host writes USERDATA; you never do. Only you write ACTIONS/OUTPUT/SCRATCHPAD; the host never does.”* That single line prevents a shocking amount of confusion.






Exactly—*with one nit*: it’s not a hidden side-channel. It’s the one **in-envelope, host-authored lane** the program reads each turn. Everything remains in the transcript. The model “listens” only there for instructions; everything it says or does is elsewhere.

Here’s a tight drop-in for v3 to make that crystal clear:

```
Execution model
- “Agent” = author; their ACTIONS run inside the host’s NS interpreter.
- Only the host writes USERDATA; the program never does. Treat USERDATA as immutable input.
- Only the program writes ACTIONS/OUTPUT/SCRATCHPAD; the host never does.

Envelope sections (order & cardinality)
- Sections appear once in this order: USERDATA, [SCRATCHPAD], [OUTPUT], ACTIONS.
- Exactly one USERDATA (JSON object) and exactly one ACTIONS are required.
- SCRATCHPAD/OUTPUT may appear 0–1 times each (from the prior turn).
- If duplicates occur, parse the first and ignore the rest.

Control path (transcript-as-truth)
- To change loop state, the program MUST emit exactly one verified control token:
    emit tool.aeiou.magic("LOOP", { 'control':'continue'|'done'|'abort', ... })
- Host honors control **only** when that exact line appears in OUTPUT and verifies.
- The control token MUST be the last line the program emits this turn (“commit-last”).
- USERDATA and SCRATCHPAD are never scanned for control.
```

Two practical reminders:

* **jti** is just a unique ID per token (anti-replay), not a signature.
* In NS, use single quotes for your own JSON literals to avoid escaping—but **emit the returned token raw**, no quotes or wrapping.

Lock these invariants in the spec and your “side-channel” becomes an audited, deterministic input lane rather than a spooky action at a distance. Next step is just updating your examples to make the token the last emitted line and dropping ACK entirely.






Yep—if `ask` is synchronous and blocking, keep it boring:

### Recommendation

Run **one turn at a time per conversation (SID)**, with a **fresh NS interpreter per turn**. No multiplexing, no background threads, no reentrancy. It’s the least clever design—and that’s why it works.

### Why this is the easy, correct path

* **Zero interleaving hazards.** A turn can’t step on another; OUTPUT/SCRATCHPAD are unambiguous.
* **No hidden state bleed.** Ephemeral interpreter per turn means whatever the model “left on the floor” dies with the process.
* **Deterministic audits & replay.** The transcript + USERDATA reproduce the run exactly; you don’t have to simulate scheduler timing.
* **Security stays intact.** Keep the v3 token flow (nonce + jti + token-last commit). Synchronous execution doesn’t reduce spoof/replay risk; the magic token still earns its keep.

### Minimal control loop (mental model)

1. Host builds envelope: `USERDATA` (JSON) + last `SCRATCHPAD`/`OUTPUT` (if any) + blank `ACTIONS`.
2. Model fills `ACTIONS`; host executes it synchronously in a fresh interpreter (with `{SID, turn_index, nonce}` in context).
3. Code `emit`s logs and **finally** `emit`s the magic token.
4. Host verifies token (SID + turn + nonce + jti), applies control (`abort > done > continue`), increments turn, and discards the interpreter.

### Keep these invariants even in the “simple” world

* **One in-flight turn per SID.** Enforce with a per-SID mutex.
* **Token-last commit.** Lint if anything follows the token.
* **Tight timeouts/quotas.** Pass `context.Context` with deadlines into the interpreter and tools; hard-stop on overrun.
* **Namespace every tool by SID.** Even if you’re single-threaded today, this future-proofs you for later concurrency.
* **Still use `jti`.** It’s cheap and prevents intra-turn double-commit.

### When to revisit multiplexing

Only when you *must* overlap I/O (e.g., long tool calls, streaming, many SIDs). If that day comes, the above invariants (SID-scoped context, token-last, nonce/jti) let you scale without redesigning the protocol.

Net: don’t engineer a rat king. Synchronous, per-SID, per-turn interpreter + magic-token commit is the straight road.



# AEIOU v3 — Protocol Properties (Targets for the Spec)

This is **not design**; it’s the checklist of qualities the protocol **must** achieve. Use it to judge designs.

## 1) Core Objectives

* **Determinism:** For a given envelope + model output, the host’s decision and side effects are uniquely determined.
* **Transcript-as-truth:** Every host-visible cause of control-flow change is present in the transcript; no hidden side effects.
* **Simplicity:** Minimal moving parts; easy for naïve LLMs to follow without clever prompt judo.

## 2) Execution Model

* **Host-run only:** All `ACTIONS` execute inside the host’s interpreter; the “agent” is just the author.
* **Single-turn semantics:** Exactly one turn is active per session (SID). Within a turn, there is exactly one commit point.
* **Idempotent replay:** Re-executing the same turn with the same envelope yields the same decision and outputs.

## 3) Envelope Properties

* **Isolation:** Content outside `START…END` is inert.
* **Sections:** Exactly one `USERDATA` (host→program, JSON) and one `ACTIONS` (program code). Optional `SCRATCHPAD`, optional `OUTPUT` (both from prior turn). Fixed order: `USERDATA`, `[SCRATCHPAD]`, `[OUTPUT]`, `ACTIONS`.
* **Immutability:** Program cannot alter `USERDATA` within a turn.
* **Byte preservation:** Section contents are treated as opaque bytes by the host (except where explicitly parsed).
* **Size & encoding:** UTF-8, `\n` newlines, bounded sizes (host-enforced); oversize envelopes are rejected predictably.

## 4) Control-Plane Properties

* **Magic-only control:** Loop control is accepted **only** via a host-minted, verifiable token (the “magic token”).
* **OUTPUT-only intake:** Host reads control **only** from the public `OUTPUT` stream; never from `USERDATA` or `SCRATCHPAD`.
* **Commit-last rule:** The intended control token is the **last line** emitted this turn; that line defines the decision.
* **Single decision:** Host applies at most one decision per turn. If multiple valid tokens exist, apply precedence (`abort > done > continue`) then **last-wins** among equals; lints may fire but behavior is defined.
* **Anti-replay & binding:** Tokens are bound to `{session_id, turn_index, turn_nonce}` and carry a unique `jti`; replays or cross-session misuse are inert.
* **Authenticity:** Tokens carry a verifiable tag (MAC or signature) under a `kid`; altered tokens fail verification.

## 5) Security Properties

* **Non-forgeability:** User text or echoed strings that look like control are inert unless minted by the host tool.
* **Non-confusion:** Host never scans `USERDATA`/`SCRATCHPAD` for control tokens.
* **Namespace discipline:** Tools and storage operations are scoped to the current SID; cross-session access requires explicit, audited APIs.
* **Key rotation:** Control tokens include `kid` to allow rotation without ambiguity.
* **Fail-safe:** On verification error, schema violation, timeout, or resource ceiling breach, host resolves to a safe state (e.g., HALT/end turn) with audit logs.

## 6) Freshness & Replay

* **Per-turn nonce:** Each turn has a fresh random nonce; a valid token is valid only for that turn.
* **Deterministic canon:** Where signatures/MACs are used, the signed payload is canonically encoded (stable field order & whitespace rules) so verification is unambiguous.
* **Reproducible logs:** A transcript plus host config is sufficient to replay decisions; you don’t need hidden side-channel state.

## 7) Observability & Audit

* **Attribution:** Every decision records `{SID, turn_index, decision, reason, token_meta}`.
* **Explainability:** Control actions are explainable from the transcript; no separate “secret log” is required to justify state transitions.
* **Telemetry (optional):** Tokens may carry bounded, non-semantic telemetry (e.g., `est_tokens`, `notes`) that does not affect verification.

## 8) Error Handling & Robustness

* **Typed failures:** Tool errors (including token minting) surface as typed failures the program can handle; absence of a valid token leads to a predictable host outcome.
* **Malformed input:** Duplicated sections, invalid markers, or malformed tokens are handled deterministically (first-section-wins, token ignored, turn rejected, etc.), with clear diagnostics.
* **Time bounds:** Turn-level deadlines and quotas terminate runaway execution predictably.

## 9) Multi-Session Concurrency

* **SID isolation:** Multiple sessions may proceed concurrently; each maintains independent `{turn_index, nonce, OUTPUT, SCRATCHPAD, quotas}`.
* **No interleaving within SID:** A session never has more than one in-flight turn.
* **Token scoping:** A token from one SID is non-actionable in any other SID by construction.

## 10) Ergonomics for Models

* **Minimal rules to remember:** “Fill `ACTIONS` with one command; emit exactly one magic token as the last line; read instructions only from `USERDATA`.”
* **No escaping footguns:** Models never hand-construct tokens; the host tool returns a ready-to-emit, single-line string.
* **Schema clarity:** `USERDATA` is always a JSON object with a small, stable skeleton (`subject`, optional `brief`, `fields` object).

## 11) Extensibility & Versioning

* **Versioned markers:** Envelope and token carry explicit major version identifiers; mismatches are detectable.
* **Forward compatibility:** Unknown optional fields are ignored; unknown control kinds are rejected with a typed error.
* **Pluggable auth:** MAC or signature schemes can change under `kid` without altering higher-level semantics.

## 12) Conformance & Testability

* **ABNF-level spec:** Markers and token formats have small, strict grammars (single-line, ASCII for markers; base64url segments for tokens).
* **Golden cases:** A canonical conformance suite (valid/invalid envelopes and tokens) with expected host decisions.
* **Lints, not mysteries:** Non-fatal deviations (extra sections, multiple tokens) yield deterministic outcomes plus lints; fatal deviations yield typed rejections.

## 13) Performance & Resource Governance

* **Predictable costs:** Envelope sizes, token lengths, and per-turn compute/memory are capped and enforced.
* **No N^2 parsing:** Host parsing is single-pass over markers; section bodies are not re-tokenized except where explicitly required.

## 14) Non-Goals (keep the scope sharp)

* **No hidden direct-decide tools** callable from `ACTIONS`. Only token-in-OUTPUT commits control.
* **No implicit control through prose.** Natural language in OUTPUT never changes loop state without a valid token.
* **No intra-turn streaming semantics** beyond the final commit token (streaming logs are fine; commit is discrete).

---

**How to use this:** take any proposed design (markers, magic tool API, verifier, interpreter loop) and check it line-by-line against these properties. Anything that can’t be proven to satisfy them needs redesign or enforcement in code.


Here’s a blunt scorecard of the **v3 design** against the protocol properties we set.

|  # | Property cluster            | Status | Key design hooks in v3                                                               | Gaps / next actions                                                         |
| -: | --------------------------- | :----: | ------------------------------------------------------------------------------------ | --------------------------------------------------------------------------- |
|  1 | Determinism                 |    ✅   | Magic-only control; **commit-last**; fixed section order; one `command`              | Ship ABNF for markers; codify “first-section-wins” rule                     |
|  2 | Transcript-as-truth         |    ✅   | Control only via token **in OUTPUT**; no direct-decide tools; ACK removed            | Add lint: anything after control token → warning                            |
|  3 | Simplicity                  |   🟡   | Model’s rule-of-three: fill `ACTIONS`, read `USERDATA`, emit one token               | Provide 6–8-line “First Words” blurb in spec; trim optional knobs           |
|  4 | Execution model             |    ✅   | Host-run NS; one turn in flight per SID; fresh interpreter per turn                  | Make “no reentrancy per SID” normative text                                 |
|  5 | Envelope properties         |   🟡   | `USERDATA`+`ACTIONS` required; optional `SCRATCHPAD/OUTPUT`; fixed order; UTF-8/`\n` | Specify exact size caps; publish ABNF; define duplicate-section handling    |
|  6 | Control-plane rules         |    ✅   | `tool.aeiou.magic` mints token; OUTPUT-only intake; precedence + last-wins           | Define exact precedence text (`abort > done > continue`) in spec body       |
|  7 | Security                    |    ✅   | HMAC/Ed25519 tag with `kid`; per-turn nonce; ignore tokens outside OUTPUT            | Note “SCRATCHPAD never parsed” as a MUST; add token length limit            |
|  8 | Freshness & replay          |   🟡   | `session_id` + `turn_index` + `turn_nonce` + `jti`; canonical JSON intent            | Choose canonicalization (JCS/RFC-8785) and ship a canonicalizer             |
|  9 | Observability & audit       |   🟡   | Decision logs implied (`SID, turn, decision`)                                        | Make fields normative; add reason codes; specify redaction rules            |
| 10 | Error handling & robustness |   🟡   | “Typed failure → HALT safe” pattern; token verify failure → inert                    | Define error taxonomy; timeouts/ceilings as MUST with defaults              |
| 11 | Multi-session concurrency   |    ✅   | SID-bound tokens; per-SID quotas/state; one turn/ SID                                | Require SID scoping for **all** tools and storage namespaces                |
| 12 | Ergonomics for models       |    ✅   | No hand-crafted tokens; single-line return; single decision                          | Include 2 minimal examples (continue/done) in spec front-matter             |
| 13 | Extensibility & versioning  |    ✅   | `NSENV:V3` markers; token `V3`; `kid` for key rotation                               | Add “unknown fields ignored” rule; reject unknown control kinds             |
| 14 | Conformance & testability   |   🟡   | Clear behaviors defined                                                              | Publish golden corpus: valid/invalid envelopes + tokens; table-driven tests |
| 15 | Performance & governance    |   🟡   | Interpreter per turn; quotas mentioned                                               | Set concrete defaults (e.g., envelope ≤1 MiB; line ≤8 KiB; CPU/mem/time)    |
| 16 | Non-goals (guardrails)      |    ✅   | No direct decide; no prose control; discrete commit                                  | Add explicit “no streaming commit semantics” line                           |

### TL;DR

You’re **green** on the big rocks (determinism, transcript-as-truth, security, execution model). To reach “spec-complete,” finish: ABNF + size caps, canonical JSON choice, normative error/precedence text, and a tiny conformance suite. Once those are in, the design will be boring in exactly the right way.



feedback from gemini2.5-deepthink:

This review assesses the AEIOU v3 and Ask-Loop v3 specifications, the implementation checklist, and the preceding design discussion regarding their suitability for a reliable, secure host-AI agentic protocol.

The AEIOU v3 design is exceptionally robust and well-considered. It demonstrates a deep understanding of the challenges in agentic systems—such as ambiguity, spoofing, replay attacks, and hidden control paths—and addresses them with deterministic and security-conscious mechanisms. The specifications are detailed, normative, and implementation-ready.

### Strengths of the AEIOU v3 Protocol

The protocol excels in several key areas required for reliable and secure interaction:

#### 1. Security and Integrity
The security model is the strongest aspect of this design.
*   **Magic-Only Control (Anti-Spoofing):** Control is only possible via host-minted, cryptographically verified tokens (`NSMAG:V3`) generated by `tool.aeiou.magic`. This ensures that user input or agent-fabricated strings resembling tokens are inert.
*   **Robust Binding and Anti-Replay:** Tokens are tightly bound to a specific interaction point using `session_id`, `turn_index`, and a fresh `turn_nonce`. The `jti` (unique token ID) prevents intra-turn replay. This combination effectively neutralizes several classes of attacks.
*   **Integrity and Authenticity:** The use of canonical JSON (JCS/RFC-8785) and strong authentication (HS256 or Ed25519) ensures the token payload cannot be tampered with and verification is stable.

#### 2. Determinism and Reliability
The protocol enforces strict rules that eliminate ambiguity.
*   **Strict Execution Model:** Running exactly one turn at a time per Session ID (SID), synchronously, and using a fresh interpreter for each turn eliminates state bleed, race conditions, and interleaving hazards.
*   **Commit-Last Semantics:** The rule that the last valid token defines the turn's decision, combined with clear precedence (`abort > done > continue`), provides a deterministic outcome even if the agent erroneously emits multiple tokens.
*   **Strict Envelope Structure:** The fixed order of sections, required elements (`USERDATA`, `ACTIONS`), and the "first-section-wins" rule ensure the envelope is parsed consistently.

#### 3. Isolation and Clear Boundaries
The protocol defines clear roles and access controls.
*   **Host/Agent Separation:** `USERDATA` is strictly host-to-agent and read-only for the agent, preventing instruction forgery.
*   **Control Intake Scoping:** The host *only* scans the current turn's `OUTPUT` for control. `USERDATA` and `SCRATCHPAD` are never scanned, eliminating common injection vectors aimed at confusing the host parser.

#### 4. Auditability (Transcript-as-Truth)
By requiring the agent to `emit` the control token into the `OUTPUT` stream (rather than using a side-channel "direct decide" tool), every control decision is visible and replayable from the transcript alone. This is invaluable for auditing, debugging, and compliance.

### Recommendations: Additions and Upgrades

While the design is solid, the following suggestions would further enhance its completeness, security, and operational robustness.

#### 1. Upgrade: Mandatory Progress Guards (Anti-Looping)
Ask-Loop v3 §9 lists progress guards as optional. This should be upgraded to mandatory.

*   **The Risk:** Agents can get stuck in loops, consuming resources without advancing the session goal.
*   **Recommendation:** Upgrade §9 from MAY to **MUST**. The host must enforce progress. To make this robust, the mechanism should ideally be based on host-observed outputs rather than agent-provided telemetry. If the `ACTIONS` code, `OUTPUT`, and `SCRATCHPAD` content are identical (or semantically equivalent) for N consecutive turns (e.g., N=3), the host should `HALT(reason=ERR_NO_PROGRESS)`.

#### 2. Add: Interpreter Sandboxing and Tool Security
The protocol's security guarantees rely implicitly on the security of the underlying execution environment (the NS Interpreter) and the tools exposed to it.

*   **The Risk:** A vulnerability in an exposed tool or a sandbox escape could allow unauthorized side effects or cross-SID data access, bypassing the protocol layer.
*   **Recommendation:** Update AEIOU v3 §11 (Security Requirements) to mandate that the NS interpreter **MUST** be strictly sandboxed (e.g., minimal capabilities, network isolation). In the Checklist, add tasks for "Sandbox implementation and hardening" (M4) and explicitly include "Sandbox escape assessment" and "Tool vulnerability assessment" in the penetration test scope (M9).

#### 3. Add: Formalized Fallback Abort Token
The discussion history mentions providing a fallback token if `tool.aeiou.magic` fails, but this is missing from the final specs.

*   **The Risk:** If the magic tool fails (e.g., key store unavailable during rotation), the agent cannot signal intent, leading to a potentially disruptive `HALT(ERR_TOKEN_MISSING)` or timeout.
*   **Recommendation:** Formalize this mechanism. The interpreter context (M4) should include a pre-signed, generic "fallback abort" token. If the magic tool encounters an internal error, it should return this fallback, allowing the turn to terminate gracefully and visibly in the transcript.

#### 4. Upgrade: Prioritize Ed25519 for Architectural Flexibility
The specs default to HS256 (HMAC), which requires a shared secret. Ed25519 is listed as optional (M2) and a nice-to-have (M11).

*   **The Risk:** HS256 complicates distributed architectures and auditing, as any verifying service needs the host's secrets.
*   **Recommendation:** If the interpreter runs in a different trust domain than the loop controller, or if external auditing is required, prioritize Ed25519. Public key cryptography allows verification without sharing the private signing key, improving the security posture.

#### 5. Add: Rigorous JCS Conformance Testing
AEIOU v3 §6.3 mandates JCS (RFC-8785). Discrepancies in implementation between the minter and verifier will cause verification failures.

*   **Recommendation:** The Conformance Suite (M6) must include rigorous tests for JCS edge cases (e.g., Unicode normalization, floating-point representation) to ensure the canonicalizer implementation is robust and consistent across all relevant platforms.

#### 6. Enhance Observability and Monitoring
The specs mandate good decision logs (M7), but specific details are crucial for security monitoring.

*   **Recommendation:**
    *   **Explicit Verification Failures:** Expand the structured decision log schema (M7) to include a `verification_failure_reason` (e.g., `BAD_TAG`, `TTL_EXPIRED`, `SCOPE_MISMATCH`). Spikes in specific errors are critical security signals.
    *   **Monitor Lints:** Ensure metrics (M7) specifically track the rates of `LINT_POST_TOKEN_TEXT` and `LINT_MULTI_TOKENS`. High rates indicate models are struggling with the protocol ergonomics.

#### 7. Refine Implementation Details
Several implementation details require explicit constraints.

*   **Bounded Replay Cache:** In Ask-Loop v3 §6.2, explicitly state that the per-SID `jti` cache MUST be bounded by both TTL and a maximum number of entries (e.g., LRU eviction) to prevent excessive memory usage.
*   **Key Rotation Grace Periods:** The key management strategy (M2) must explicitly define the rotation grace period. When a new key is deployed for minting, the old key must remain available for verification for at least the maximum `ttl` of previously issued tokens.



Here’s the straight read.

### Verdict

You’ve got a production-grade protocol. With the v3.1 edits (Ed25519 default, mandatory progress guard, sandboxing, fallback signer, bounded replay cache, JCS tests), the design is tight, auditable, and hostile to spoofing. I’d green-light an implementation behind a feature flag and canary it.

### What’s strongest

* **Security posture:** magic-only control, OUTPUT-only intake, per-turn nonce + `jti`, Ed25519 default, no hidden decide paths.
* **Determinism & audit:** commit-last, fixed section order, transcript-as-truth, typed errors/lints, golden conformance plan.
* **Ops readiness:** explicit metrics, decision logs with failure reasons, rotation/grace rules, fallback signer.

### Sharp edges to watch (and how to sand them)

1. **Canonical JSON (JCS) drift.** Biggest correctness risk.
   *Mitigate:* use a single reference canonicalizer; forbid floats (already in spec); add golden JCS edge cases (Unicode, escapes, large ints).

2. **Progress-guard tuning.** Risk of false HALTs on long tasks that generate similar logs.
   *Mitigate:* N=3 is a sane default; ignore control lines and trailing whitespace (already); optionally ignore known benign prefixes (`TRACE:`, timestamps). Monitor `no_progress_total` and adjust N per workload.

3. **Token-last ergonomics.** Models sometimes add a newline or comment after the token.
   *Mitigate:* define “last **non-empty** line” as the commit point and lint on any non-blank trailing text. Keep the single-decision rule.

4. **Key rotation & fallback.** Operational footguns if misconfigured.
   *Mitigate:* enforce “mint with K\_new; verify with {K\_new,K\_old} for ≥ ttl+grace”; healthcheck signer; alert on fallback-signer use.

5. **Sandbox & tool surface.** Real risk lives here, not in the protocol.
   *Mitigate:* keep the tool allowlist tiny; SID-scope all storage; block ambient net/fs; pen-test tool calls; log cross-SID attempts.

### Go/No-Go bar (canary success criteria)

* `verify_fail_total / decisions_total` < **0.5%**
* `LINT_MULTI_TOKENS / decisions_total` < **2%** (trending down)
* `LINT_POST_TOKEN_TEXT / decisions_total` < **5%** (after docs)
* `no_progress_total / decisions_total` < **1%**, with manual spot-checks
* 0 incidents of cross-SID access or unsigned control acceptance

### Remaining redlines before broad rollout (tight list)

* Ship the **ABNF** and **reference canonicalizer** with tests.
* Lock “**last non-empty line** is the commit line” into the spec & linter.
* Finalize **default TTL**, `jti` TTL/cap, and nonce size (128-bit min).
* Document **Ed25519 key storage** & rotation procedure (who holds private key, where public keys are published).
* Complete the **golden corpus** (env/tokens/JCS edge cases) and gate CI on it.

### Likely failure modes (and expected behavior)

* **Hand-typed/quoted token:** inert → `HALT(ERR_TOKEN_MISSING)`; lint may note quoting.
* **Valid token with wrong SID/turn/nonce:** inert → `HALT(ERR_TOKEN_SCOPE)`.
* **Signature mismatch / bad base64:** inert → `HALT(ERR_TOKEN_VERIFY)`.
* **Repeated `jti`:** first effective; later inert; replay counter increments.
* **Three identical digests in a row:** `HALT(ERR_NO_PROGRESS)` by policy.

### Bottom line

You’ve converged on the boring, correct shape: one immutable input lane (USERDATA), one executable lane (ACTIONS), one public output lane (OUTPUT), and one cryptographic stamp that commits the turn—visible in the transcript. The remaining work is implementational hygiene and guard-rail tuning, not re-architecture.
