# JSON keys standard for NS

Nice snapshot. You’ve got three clashing styles living side-by-side:

* **snake\_case** (session\_id, turn\_index, base\_url)
* **lowerCamelCase** (defaultValue, returnType, promptTokenCount)
* **oddballs / dotted** (nslsp.externalToolMetadata) and ultra-short (“v”)

That mix will bite you—hard—when you start round-tripping, diffing, or asking AIs to reason over payloads.

## The target standard (clear, boring, durable)

**Rule 0 (scope):** Apply these rules to **all internal NS/FDM wire formats** (anything produced/consumed by your code or docs). For **third-party provider structs**, mirror the provider exactly and convert at your boundary.

**1) Casing:**
Use **snake\_case** for all keys. No exceptions for acronyms.
Examples: `id`, `base_url`, `full_name`, `tool_choice`, `response_format`.

**2) Names:**

* Prefer **short, literal nouns** over cleverness: `jti` is domain-standard; keep it.
* Keep `version` as a real word **except** in the AEIOU magic envelope where compactness is intentional—keep `v`.
* Avoid pluralization in containers unless necessary (prefer `children` over `child_ids` only if the value is naturally the array; if you carry IDs, say `child_ids`).

**3) Units & types:**

* Timestamps: `issued_at` = Unix seconds (int). If you ever need RFC3339 strings, add a **new field** (`issued_at_rfc3339`) rather than type-flipping.
* Durations/TTL: suffix with units, `ttl_s` (seconds), `ttl_ms` (milliseconds). You currently have `ttl` (ambiguous) → make it `ttl_s`.
* Currencies: ISO-4217 uppercase in values, key stays snake case: `budget_currency: "CAD"`.
* Token counts and sizes: suffix the unit, e.g., `context_ktok` is cute but unclear; prefer `context_tokens_k` **or** just `context_tokens` (plain count) and drop the kilo; the value can be big.

**4) Optional fields:**
Use `omitempty` only when **absence** means “use default.” If `false` vs “unset” matters for booleans, use a pointer `*bool` or an explicit tri-state (`mode: "auto"|"on"|"off"`). Today you have several booleans with `omitempty` where `false` is meaningful (`auto_loop_enabled`, `tool_loop_permitted`)—consider pointers or enum modes.

**5) Namespacing:**
Never use dots in JSON keys. `nslsp.externalToolMetadata` should be a nested object:

```json
{ "nslsp": { "external_tool_metadata": [...] } }
```

**6) Collections & maps:**

* Arrays: plural name, e.g., `tool_calls`, `stop_sequences`.
* Maps: name the map role, e.g., `settings`, `attributes`, `fields`.

**7) Schema/versioning:**

* Public envelopes (AEIOU, user\_data, tool I/O) MUST carry a stable version indicator. For AEIOU, keep `v`. For others, use `schema_version` (int).
* Do **additive** changes only within a version; breaking changes require bump.

**8) Canonical JSON:**
You already test canonicalization—good. Specify: UTF-8, no NaN/Inf, integers as decimal, `\u` escaped control chars, **stable key ordering** (lexicographic), and no insignificant whitespace. Keep that as a doc’d invariant.

---

## Concrete cleanups by file (minimal churn, maximal payoff)

### pkg/types/agentmodel.go

* `account_name`, `base_url`, `budget_currency` → good.
* **Rename** `context_ktok` → `context_tokens` (plain count).
* `price_table`, `generation`, `tools`, `safety` → good.
* Generation knobs:

  * `max_output_tokens`, `stop_sequences`, `presence_penalty`, `frequency_penalty`, `repetition_penalty`, `top_p`, `top_k`, `seed`, `log_probs`, `response_format` → already snake case; keep.
* Looping:

  * `tool_loop_permitted`, `auto_loop_enabled`, `tool_choice` → keep, but consider enum for loop policy later.
* Prices:

  * `input_per_mtok`, `output_per_mtok` → rename to `input_price_per_mtok`, `output_price_per_mtok` **or** better, drop “mtok”: `input_price_per_token`, `output_price_per_token`. (Using per-million invites silent unit bugs.)

### pkg/tool/tools\_types.go

* `defaultValue` → `default_value`
* `returnType` → `return_type`
* `returnHelp` → `return_help`
* `errorConditions` → `error_conditions`
* `groupname` → `group_name`
* `fullname` → `full_name`
* `requiredCaps` → `required_caps`
* `signatureChecksum` → `signature_checksum`

### pkg/utils/tree\_types.go

* `rootId` → `root_id`
* Consider making the node’s parent references explicit in JSON only if needed; you already use `json:"-"` which is fine.

### pkg/interfaces/llm.go

* You’re consistent: `tool_calls`, `tool_results`, `full_name`, `arguments`. Good.

### pkg/provider/google/google.go (and any other provider bindings)

* **Do not change**; mirror provider JSON exactly (e.g., `promptTokenCount`). Convert to internal snake\_case DTOs at the boundary.

### pkg/aeiou/\*

* Envelope: keep the compactness where it serves the protocol.

  * Keep `v`, `jti`, `kid`.
  * `session_id`, `turn_index`, `turn_nonce`, `issued_at` are fine.
  * **Change** `ttl` → `ttl_s` (seconds).
* `payload`, `telemetry` are good.
* Userdata: `subject`, `brief`, `fields` fine.

### pkg/nslsp/config.go

* Replace `nslsp.externalToolMetadata` with nested object as noted above.

---

## Small but important policy decisions

* **IDs vs names:** `id` is an opaque identifier. `name` is human/stable (but renameable). Keep both when both are useful.
* **Booleans as modes:** Replace families of booleans with a single enum string where feasible (`tool_loop_mode: "forbid"|"allow"|"auto"`).
* **Numbers with scale:** Make unit explicit (price per token, seconds vs ms). Avoid “k”, “m” semantics in field names.

---

## Suggested migration plan (low-risk, mechanical)

1. **Freeze** current public schemas: write short `.md` files per payload (AgentModel, ToolSpec, AEIOU Control, UserData, Tree). Include version, fields, types, units.

2. **Introduce vNext tags** side-by-side:

   * Add new struct tags for renamed fields, keep old tags with `omitempty` and mark **deprecated** in comments.
   * On encode: populate **both** for one release (or via a custom marshal wrapper).
   * On decode: accept both, but prefer new names.

3. **Boundary converters**:

   * For providers, add `toInternal()`/`fromInternal()` mappers that transform camelCase to snake\_case once, at the edge.

4. **Linter check**:

   * Add a trivial CI check that fails if a tag doesn’t match `^[a-z0-9_]+$` and flags dots or capitals. Also flag `ttl$` without a unit suffix and mixed-case like `returnType`. (I can draft the checker if you want it turnkey.)

5. **Deprecation window**:

   * One minor release shipping both names; then remove the deprecated tags.

---

## Quick rename map (the obvious outliers)

* defaultValue → default\_value
* returnType → return\_type
* returnHelp → return\_help
* errorConditions → error\_conditions
* groupname → group\_name
* fullname → full\_name
* requiredCaps → required\_caps
* signatureChecksum → signature\_checksum
* rootId → root\_id
* context\_ktok → context\_tokens
* input\_per\_mtok → input\_price\_per\_token
* output\_per\_mtok → output\_price\_per\_token
* ttl → ttl\_s
* nslsp.externalToolMetadata → nslsp: { external\_tool\_metadata: \[...] }

---

## Final sanity notes

* Keep AEIOU’s compact envelope (`v`, `jti`, etc.)—it’s a special case and documented. Everywhere else, be boring and predictable.
* Document units in code comments next to tags. Future-you will thank present-you.
* Don’t let provider quirks leak inside. Convert at the edge and forget about them.

If you want, I can spit out a tiny Go AST pass or a `staticcheck`-style linter rule that flags non-snake JSON tags and unitless duration fields so this stays fixed.
