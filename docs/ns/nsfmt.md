# NSFMT

AJP: please consider nsfmt. It should aim to be like gofmt -- opinionated etc. and try for clear, readable layout, with focus on being deterministic (kind of a precursor to our ast canonicalization process) and LLM-friendly since they will be the primary authors and readers; humans mostly readers. We need to be especially careful with metadata and comments, which have caused seemingly endless grief before.

Your thoughts on architecture and layout rules please. Or, do we pass everything through an ast and back?

GPT5: Alright, nsfmt. Goal: deterministic, boring, unambiguous. Humans should never argue with it; LLMs should never be surprised by it.

Here’s a design that keeps you out of refactor hell, preserves every comment/metadata crumb, and yields a gofmt-like one-true-style.

# Architecture

**Three layers, cleanly separated:**

1. **Parse layer (CST + AST):**

   * Keep your existing AST for semantics.
   * Add a lightweight **CST token stream** (all tokens including comments/whitespace).
   * Build an **AST↔CST mapping**: for each AST node, record the token span that produced it (start/end token indices). This is the anchor for comment attachment and stable printing.
   * Reuse the Stage 0/1 token source for parsing; nsfmt itself does **not** need it during printing (formatter decides line breaks).

2. **Attachment layer (trivia pass):**

   * Scan comments (line + block) and attach them as **leading / trailing / inner** to the *nearest* AST node by token index:

     * **Leading**: last comment group ending before node’s first token.
     * **Trailing**: first comment group starting after node’s last token on the same line.
     * **Inner**: comments lexically inside the node’s token span but not owned by a child span.
   * Group adjacent line comments into **comment groups** (no blank line between them) and preserve original relative order within each group.
   * Record **exact text** of comments as-is (don’t normalize; don’t wrap).
   * Treat **file header comment block** (before the first non-comment token) as **preamble metadata** (see “Metadata rules” below).

3. **Printer layer (canonical pretty-printer):**

   * Walk the AST and emit code using **fixed rules** (below).
   * Comments are emitted at well-defined anchor points:

     * Leading group before the node with a single blank line separation policy.
     * Trailing group at end of the node’s last line.
     * Inner comments at structurally meaningful “seams” (before `end`, before `else`, before each list/map element, before each statement in a block).
   * Implement a **soft line-break engine**:

     * Maintain a target **max line length** (e.g., 100).
     * For breakable constructs (argument lists, lists/maps, infix chains, pipelines), choose single-line vs multi-line by a *simple cost model*: if it fits → one line; else → broken with one element per line and consistent indentation. Deterministic; no column alignment games.
   * Output is **idempotent**: reformatting already formatted code yields identical bytes.

Keep **canonicalization** (sorting, renaming, reordering) out of nsfmt; that’s a separate `nscanon` pass later. nsfmt does *layout*, not semantics.

---

# Style rules (opinionated, LLM-friendly)

These are deliberately simple and uniform. No options beyond maybe `-w` (write) and `-l` (list changed files).

## Whitespace & indentation

* Indent with **2 spaces** (no tabs). Reason: LLMs handle spaces more robustly; columns don’t drift.
* One statement per line. No semicolons; rely on the grammar’s line structure.
* Exactly one space:

  * after keywords (`if`, `for`, `proc`, etc.) when followed by an expression
  * around binary operators: `a + b`, `a && b`, `x := y` (or your `set x = y` retains a single space around `=`).
  * after commas/colons in pairs: `["a", "b"]`, `{ "a": 1, "b": 2 }`, but **no** space before commas/colons.
* No extra vertical alignment (no lining up `=` or `:` in columns). Stability > prettiness.

## Blocks & control flow

* Braced/keyword blocks each on their own lines:

  ```
  if cond
    stmt
    stmt
  else
    stmt
  end
  ```
* **No** hanging `else` on the same line as `end`; always:

  ```
  end
  else
  ```

  or, if your syntax supports `elif`, format as:

  ```
  elif cond
  ```

  (single keyword per line; helps diffs and LLM parsing).
* Empty block prints as:

  ```
  if cond
  end
  ```

## Functions/procedures

* Signature on one line if it fits; otherwise break after the opening paren and put one parameter per line with trailing comma:

  ```
  proc foo(a string, b int)  // fits
  ```

  becomes

  ```
  proc foo(
    a string,
    b int,
  )
  ```
* Same rule for callsites.

## Lists / maps / sets

* Short literals fit on one line:

  ```
  [1, 2, 3]
  { "a": 1, "b": 2 }
  ```
* If broken, then:

  * opening token on its own line,
  * **one element per line**,
  * **trailing comma required**,
  * closing token aligned with the start:

  ```
  [
    1,
    2,
    3,
  ]
  {
    "a": 1,
    "b": 2,
  }
  ```
* Key/value: exactly one space after `:`; none before.

## Expressions

* Unary operators tight: `!x`, `-x`.
* Binary operators spaced. For long chains, prefer **operator-at-end** and line break after operator:

  ```
  x =
    a +
    b *
    c
  ```

  If a group with parentheses fits nicely, nsfmt may choose:

  ```
  x = a
    + b*c
  ```

  but the default is one operator per continued line (LLM-friendly).

## Trailing commas

* Multiline argument/element lists: **trailing comma mandatory**.
* Single-line lists: **no** trailing comma.

## Blank lines

* File header comment (preamble) → exactly **one** blank line before the first code token.
* Separate **top-level declarations** by **one** blank line, never more.
* Inside blocks: a single blank line allowed between *logical groups*; nsfmt compresses multiple blank lines to one.
* No blank line before closing `end`, `)`, `]`, `}` unless there’s an inner comment group directly above (which is preserved).

## Comments & metadata (the tricky part)

* Line comments keep their text exactly. No wrapping or capitalization edits.
* **Leading comment groups** stick to the node they precede; nsfmt ensures one blank line before the node (unless it’s the first inside a block).
* **Trailing comments** stay at the end of the line where the construct ends; if the node becomes multiline, trailing → becomes a **post-line** comment on the last line.
* **Inner comments** (between children) are emitted immediately before the next child, with a preceding newline (never inline), preserving order.
* **Preamble metadata block** (very top of file):

  * If the first tokens are only comments, nsfmt treats the first contiguous block as the **file header**.
  * You likely have headers like:

    ```
    ```

// NeuroScript Version: 0.5.0
// File version: 3
// Purpose: ...
// filename: ...
// nlines: ...
// risk\_rating: LOW
\`\`\`

* nsfmt preserves the header block verbatim, then inserts exactly one blank line after it.
* Optional: recognize `//@meta:` style tags and keep them on a single line (no reflow).
* **Directive comments**: reserve a minimal set:

  * `// nsfmt: off` … `// nsfmt: on` — formatter respectfully leaves the enclosed range untouched, aside from normalizing surrounding blank lines.
  * `// nsfmt: keep` — sticky trailing comment; nsfmt won’t lift or reflow it.
  * `// nsfmt: newline-before` / `// nsfmt: newline-after` — rare escape valves.

No other knobs. Opinionated means small surface area.

---

# Determinism & stability guarantees

* **Idempotent:** running nsfmt twice yields byte-identical output.
* **Stable under trivia:** adding/removing horizontal whitespace (not inside strings) anywhere results in the same formatted output.
* **Stable under tokenization:** different original line breaks that parse to the same AST yield the same printed form (subject to comment placement).
* **Versioned printer:** include a visible header in the tool `nsfmt -v` and tag golden tests by version. Never make silent style changes across minor versions; bump a style version.

---

# Error handling (partial formatting)

You don’t want the formatter to explode on half-typed code (especially in LSP). Strategy:

* Try full parse. If it **succeeds** → format from AST (canonical).
* If it **fails** → fall back to a **token-stream reflow**:

  * Minimal whitespace normalization: collapse runs of spaces, fix indentation by counting open blocks/delimiters using a best-effort stack, keep all comments exactly where they occur.
  * Never delete tokens; never join or split identifiers/strings.
  * Mark diagnostics but still produce output so the LSP can display something stable.
* This fallback lives in a tiny, isolated module; simpler than trying to print from a partial AST.

---

# Command-line and LSP integration

* `nsfmt` flags:

  * `-w` write result to file(s).
  * `-l` list files whose formatting differs.
  * `-diff` print unified diff to stdout.
  * `-stdin`/`-stdout` for editor integration.
* **nslsp**:

  * Provide **range formatting** and **on-type formatting** using the same printer with a **range clamp**: reparse the file, reprint the smallest enclosing syntactic unit that fully contains the edit range (statement / literal / arg list). This avoids the “format the whole file on every keystroke” churn while staying deterministic.
  * Keep an in-memory **incremental cache** of parse trees keyed by file version to lower latency.

---

# Implementation sketch (Go)

* `pkg/nfmt/`

  * `printer/` – AST walker + pretty-printer (pure; writes to a `bytes.Buffer`).
  * `attach/` – comment attachment (CST→AST trivia).
  * `fallback/` – token reflow mode for parse errors.
  * `style/` – constants: max line length, indentation size, line-break thresholds.
  * `cmd/nsfmt/` – CLI.
* **Interfaces:**

  ```go
  type Formatter interface {
    FormatFile(src []byte) ([]byte, error)              // full, canonical
    FormatRange(src []byte, start, end int) ([]byte, error) // for LSP
  }
  ```
* **Testing:**

  * Golden-file tests: “ugly.in.ns” → “pretty.out.ns”.
  * Fuzz: random whitespace/comment injection; formatter must stabilize to the same bytes.
  * Idempotence: `fmt(fmt(x)) == fmt(x)`.
  * Comment torture suite: headers, trailing, inner, comments between commas, comments inside empty blocks, comments attached to `end/else`, etc.

---

# AST vs “pass everything through AST”?

Yes: **format from AST**. That’s how you guarantee structure, precedence, and consistent wrapping. But you must **preserve concrete trivia** using the CST mapping so comments don’t get lost or drift. Token-stream-only formatters inevitably become ad-hoc parsers; full-AST formatters that ignore CST inevitably anger users by eating or teleporting comments. The hybrid keeps you safe:

* AST determines structure and line breaks.
* CST decides **where comments bind** and ensures their **verbatim text** is emitted.

Canonicalization (sorting, normalization of map literal keys, import ordering, etc.) remains a **separate tool** (or a `-canon` mode later). Keep nsfmt’s contract small: layout only.

---

# Practical next steps

1. Expose the AST↔token indices you already have (start/end token for each node).
2. Implement the **attach** pass (100–150 LOC if your node spans are reliable).
3. Build a minimal **printer** for a handful of nodes (statements, lists, maps, assignments) and grow outward.
4. Wire `nsfmt` CLI and a **format-on-save** LSP command.
5. Add the **comment torture** golden tests; bake them into CI.

When that’s green, you have a gofmt-grade foundation that won’t get weird on you—and your LLM collaborators will be thrilled by the uniformity.







AJP: Stability > prettiness ---- excellent way to put it. Semantic prettiness rules.
Again, please could you expand on the above and explain more.
At the risk of seeming ignorant: what is CST?
I do think we need a way to ensure a comment block is attached to the right fn -- my usual way to do this in go is with a \\ on its own line; otherwise a comment block is associated with the preceding statement.
I think in the current canon. step we associate comments with the preceding block as doing any kind of "nearest" computation proved maddening.
Please see also the metadata.md and incorporate it.

Got it—your sentinel is a **lone `//` line** (no text). I’ll restate the plan with that baked in, and keep it deterministic, metadata-aware, and boring in the best way.

# What’s a CST (super short)?

CST = **Concrete Syntax Tree**: every token you typed, including whitespace and comments.
AST = **Abstract Syntax Tree**: the semantic skeleton (no trivia).
Formatter rule: **print from AST for structure; consult CST for comment/metadata placement**.

# nsfmt architecture (boring-by-design)

1. **Parse once → AST + CST tokens.**
2. **Attachment pass (CST → AST):** deterministically bind comments/metadata to AST nodes.
3. **Pretty-printer (AST → text):** emit the one true layout; re-emit attached comments verbatim at fixed anchors.
4. **Idempotence:** `nsfmt(nsfmt(x)) == nsfmt(x)` or it’s a bug.

`nsfmt` does **layout only**. Sorting/reordering/etc. stays in `nscanon`.

---

# Comment & metadata rules (deterministic)

## 0) Metadata is first-class

* **File-scope metadata**: must live at the very top; nsfmt preserves it verbatim and enforces **exactly one blank line** after the block.
* **Function-scope metadata**: must live **inside the function block** at the top, in the order written; nsfmt preserves lines verbatim and does **not** reorder metadata keys.
* nsfmt never rewrites metadata text; it only enforces placement/spacing.

## 1) Default association (no directives)

* A contiguous comment group **immediately above** a declaration (proc/const/var/etc.) with a blank line **after** the group becomes that declaration’s **leading** doc comment.
* Comments on the **same physical line** as a construct’s last token are **trailing**.
* Comments between children of a composite (args/elements/statements) are **inner**, anchored to the next child.

## 2) Your sentinel for “attach to next”: **lone `//`**

You want the Go-like nudge, but here with `//` (no text). We’ll make it explicit and machine-readable:

* **Parsing rule:** a line that consists of `//` (optionally followed by spaces) and **nothing else** is treated as **“attach to next declaration”**.
* **Printing rule (canonicalization):** nsfmt **prints** that sentinel as:

  ```
  ```

// ::attach: next

````
This keeps the source self-describing, grep-able, and consistent with the metadata style. It also prevents accidental loss of intent if someone strips “empty comments.”

- Behavior: the `// ::attach: next` (or its `//` sugar) binds the **next declaration** as its doc target **even if** there’s an intervening blank line or attributes.

(If you ever need symmetry, we can also accept `// ::attach: prev` for post-docs. Default remains attach-to-next via the lone `//` sugar.)

## 3) Function docs vs function metadata (clean split)
- **Leading human doc** stays above `proc`:
```ns
// One-sentence summary for humans/LLMs.
proc process(items, threshold)
  ::description: Tool-readable one-liner
  ::param:items: Input list
  ::param:threshold: Optional numeric
  …
end
````

* nsfmt enforces **no blank line** between `proc …` and the first function-scope metadata line (per your standard), and preserves order verbatim.

---

# Layout rules (LLM-friendly; stability > prettiness)

* **Indent:** 2 spaces; no tabs.
* **Line length target:** \~100; break only at syntactic seams.
* **Blocks:**

  ```
  if cond
    stmt
  else
    stmt
  end
  ```

  One keyword per line; no `end else` combos.
* **One statement per line.** No alignment games.
* **Binary ops:** spaces around; for multi-line, break **after** the operator:

  ```
  x =
    a +
    b * c +
    d
  ```
* **Calls/params, lists/maps:** one line if it fits; otherwise one item per line with **trailing comma mandatory**:

  ```
  foo(
    a,
    b,
  )
  {
    "a": 1,
    "b": 2,
  }
  ```
* **Blank lines:**

  * Exactly one between top-level decls.
  * Exactly one after the **file-header metadata block**.
  * Inside blocks: compress runs to at most one (unless an **inner comment group** sits there).
* **Comments:**

  * Never reflow text.
  * Leading groups: one blank line before the node (unless at block start).
  * Trailing comments stay on the construct’s last line; if the construct becomes multi-line, they move to the last line of that construct.
* **Metadata:**

  * File-scope at top; fn-scope inside block start.
  * Keys/values preserved verbatim; no reordering in nsfmt.

---

# Why AST-print + CST attach (not token reflow)?

* AST-print guarantees precedence and consistent wrapping.
* CST attach guarantees every comment/metadata line is placed predictably and preserved verbatim.
* Token-reflow formatters devolve into fragile partial parsers; we avoid that.

**Editor friendliness:** if parse fails mid-edit, nsfmt falls back to a tiny token-reflow that **only** fixes indentation and preserves every comment in place. As soon as the AST builds, the canonical printer takes over.

---

# Determinism contract

* **Idempotent.**
* **Trivia stable:** superficial whitespace changes that don’t alter tokens format to the same bytes.
* **Attachment stable:** with `//` (→ `// ::attach: next`) you never lose intent; without it, the rule-based association is testable and consistent.
* **Versioned style:** bump a style version only on deliberate change.

---

# Implementation sketch you can drop in

* `attach` pass (pseudocode):

  ```
  for each comment group in CST:
    if group.isLoneDoubleSlash(): mark as AttachNext
    if group.startsWith("::attach: next"): mark as AttachNext
    if group.startsWith("::attach: prev"): mark as AttachPrev

    if AttachNext: bind to next decl’s leading slot
    else if AttachPrev: bind to previous decl’s trailing slot
    else:
      apply default: leading/trailing/inner by token spans + blank-line rule
  ```

  Printer: when encountering an attached “AttachNext sugar” group, **emit** `// ::attach: next` as the first line of the group (replacing a bare `//`). Everything else prints verbatim.

* Tests you absolutely want:

  * Lone `//` above a function binds to **next**; formats as `// ::attach: next`.
  * Lone `//` separated by a blank line still binds to **next** (that’s the point).
  * File-header metadata block at top → one blank line after, never two.
  * Fn-scope metadata prints as first lines inside the block, no blank line between `proc` and first meta.
  * Inner/trailing comments preserved with multiline wrapping of calls/lists/blocks.
  * Idempotence and whitespace fuzz tests.

---

That’s the whole trick: explicit **lone `//`** for “attach to next,” canonicalized to a visible `// ::attach: next` so tools (and future you) don’t have to remember folklore. The formatter remains aggressively opinionated, metadata-aware, and predictable—the kind of stability LLMs and CI can lean on.




AJP: maybe we should just drop the // thing

GPT5: That’s probably the right instinct. A special “lone `//`” rule is cute, but it adds magic and ambiguity—the sort of thing that will bite you in six months when you can’t remember why a particular comment bound forward instead of backward. Stability beats cleverness here.

If we drop the `//` convention, you’ve got a cleaner story:

* **Leading comments**: contiguous group directly above a declaration (with a blank line *after* the group) → bound to that declaration.
* **Trailing comments**: same physical line as the construct’s end token → bound as trailing.
* **Inner comments**: between children (args, list items, statements) → bound to the next child.
* **File-scope metadata**: always at top, verbatim, one blank line after.
* **Function-scope metadata**: always first lines inside the block, verbatim, no blank line after the `proc` line.

And if you *really* need an override, you already have the explicit mechanism in your metadata format:

```ns
// ::attach: next
// documentation that explicitly binds to the following decl
```

That’s parsable, grep-able, and survives round-trips without surprises. LLMs and tools can be taught the directive trivially. It also keeps comment handling unified with the rest of your metadata system instead of inventing a one-off marker.

So the recipe becomes:

* No silent sugar (`//`) → less folklore to remember.
* Only **rule-based defaults** + **explicit metadata directives**.
* Formatter enforces spacing rules so that “doc block above decl” is unambiguous 99% of the time.

That way nsfmt is predictable, deterministic, and easier to teach to LLMs—exactly in line with your “stability > prettiness” principle.



