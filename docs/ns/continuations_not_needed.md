# Removing (most of) the need for line continuations

AJP: Attached is the ns grammar g4 file. I'd like to make continuation marks (\) unneeded in most places. Please devise a plan that doesn't lead into a tar pit of refactoring and bugs :P


GPT5: Alright Andrew — here’s a plan that buys you “newline freedom in most places” without dragging the whole grammar through a swamp. We’ll do it in *two small, reversible stages*, and stop as soon as it feels good enough. No heroic parser surgery, no brittle precedence games.

---

# Stage 0 (low risk, 1–2 touch points): “Newlines don’t matter inside () \[] {}”

**Goal:** Make `\` unnecessary anywhere you’re already *syntactically nested* (arg lists, grouped expressions, list/map literals).
**Tactic:** Keep the parser as-is. Teach the **lexer** to hide `NEWLINE` whenever we’re nested.

### What to change

1. **Track nesting in the lexer**

   * Increment on `LPAREN/LBRACK/LBRACE`, decrement on `RPAREN/RBRACK/RBRACE`.
   * You already have `NEWLINE` and a `LINE_ESCAPE_GLOBAL`. Leave escapes working; we’ll just need them far less.

2. **Route `NEWLINE` to HIDDEN channel when nested > 0**

   * Concretely: in the `NEWLINE` rule action (or a token-source wrapper), if `nestingDepth > 0`, emit it to `channel(HIDDEN)` (i.e., treat as whitespace). Otherwise, keep emitting `NEWLINE` normally (your grammar still uses it to end statements and to start blocks).

3. **No parser edits**

   * Your existing rules like:

     ```
     KW_IF expression NEWLINE non_empty_statement_list …
     statement NEWLINE …
     list_literal: LBRACK expression_list_opt RBRACK;
     ```

     continue to work. Inside `(...)`/`[...]`/`{...}` you can write:

     ```ns
     set x = fn(
       a,
       b,
       c
     )
     set y = [
       1,
       2,
       3
     ]
     set z = {
       "a": 1,
       "b": 2
     }
     ```

     with no `\`.

### Why it’s safe

* Parser remains untouched (zero risk of ripple bugs).
* Changes are strictly lexical; outside of nesting, `NEWLINE` still acts exactly like before, so your block structure and “one statement per line” discipline remain intact.

---

# Stage 1 (still small, nice quality-of-life): “Dangling operators imply continuation”

**Goal:** Allow *operator-led wrapping* without `\` even when you’re not inside parens/brackets/braces, e.g.

```ns
set x = a +
        b +
        c
```

**Tactic:** Still **lexer-only**. Treat a `NEWLINE` as whitespace (HIDDEN) when the previous *emitted* token syntactically demands a continuation.

### What to change

1. **Keep `nestingDepth` from Stage 0.**

2. **Track `continuationWanted`**

   * After emitting a token from the set:

     * binary/infix ops: `PLUS, MINUS, STAR, SLASH, PERCENT, EQ, NEQ, GT, LT, GTE, LTE, AMPERSAND, PIPE, CARET`
     * assignment and separators where a RHS is expected: `ASSIGN, COLON`
     * member/dot and call openers: `DOT, LPAREN, LBRACK, LBRACE` (though these are already covered by nesting)
   * set `continuationWanted = true`.
   * Clear it after you emit any non-whitespace token that could start an expression/atom.

3. **NEWLINE handling**

   * If `nestingDepth > 0` **OR** `continuationWanted == true` ⇒ send `NEWLINE` to `channel(HIDDEN)`.
   * Otherwise, keep `NEWLINE` as a real token (no behavior change for statement endings and block headers).

### Scope & limitations (by design)

* This covers the 99%: infix wraps, chained dots that happen to be line-wrapped, and all inside-delimiters.
* You don’t have to add optional `NEWLINE*` all over expression rules. Parser still unchanged.

---

# Optional Stage 2 (only if you really want it later): “Virtual semicolons” like Go

**Goal:** Let `NEWLINE` become *pure whitespace* and insert a synthetic end-of-statement token (`STMTEOL`) only when the line *could* end a statement and we’re not nested. This would remove almost all explicit references to `NEWLINE` from the parser.

**Tactic (brief):**

* Add a small token-source wrapper (“semicolon injector”) around the generated lexer. On every `NEWLINE`:

  * If `nestingDepth == 0` **and** the previous token is an “end-of-statement” token (`IDENTIFIER`, literals, `)`, `]`, `}`, `KW_LAST`, `KW_TRUE/FALSE/NIL`, `return` with or without expr, `fail`, etc.), **emit** `STMTEOL` and drop the `NEWLINE` to HIDDEN.
  * Otherwise drop the `NEWLINE` to HIDDEN with no `STMTEOL`.
* Parser changes are mechanical: replace `NEWLINE` with `STMTEOL` in `statement` lists and block headers, and usually accept `STMTEOL+` (one or more) where you previously had `NEWLINE`/blank lines.

**Why this is optional**

* It’s a clean architecture, but it touches many spots. Stages 0–1 deliver 90% of the ergonomics without refactoring grammar rules.

---

## Concrete implementation notes (Go target, ANTLR)

* **Minimal intrusion path:** don’t edit the generated lexer’s internals; instead, wrap it:

  ```go
  type nsTokenSource struct {
      antlr.TokenSource
      prev antlr.Token
      nest int
      cont bool
      // small ring buffer if you need to peek/emit STMTEOL etc.
  }

  func (s *nsTokenSource) NextToken() antlr.Token {
      for {
          t := s.TokenSource.NextToken()
          typ := t.GetTokenType()

          switch typ {
          case NeuroScriptLexerLPAREN, NeuroScriptLexerLBRACK, NeuroScriptLexerLBRACE:
              s.nest++
              s.cont = true
          case NeuroScriptLexerRPAREN, NeuroScriptLexerRBRACK, NeuroScriptLexerRBRACE:
              if s.nest > 0 { s.nest-- }
              s.cont = false
          case NeuroScriptLexerPLUS, NeuroScriptLexerMINUS, /* … all infix ops … */ NeuroScriptLexerASSIGN, NeuroScriptLexerCOLON, NeuroScriptLexerDOT:
              s.cont = true
          case NeuroScriptLexerNEWLINE:
              if s.nest > 0 || s.cont {
                  // treat as whitespace: loop to next token
                  continue
              }
              s.cont = false
          default:
              // normal token starts an atom ⇒ no pending continuation
              if typ != antlr.TokenEOF && t.GetChannel() == antlr.TokenDefaultChannel {
                  s.cont = false
              }
          }

          s.prev = t
          return t
      }
  }
  ```

  * Plug this into your `CommonTokenStream`:

    ```go
    base := parser.NewNeuroScriptLexer(input)
    src  := &nsTokenSource{TokenSource: base}
    stream := antlr.NewCommonTokenStream(src, 0)
    p := parser.NewNeuroScript(stream)
    ```

* **Leave** your `LINE_ESCAPE_GLOBAL`/`CONTINUED_LINE` in place; they become a harmless escape hatch.

* **Tests to add (before/after snapshots)**

  * Multiline call args (no `\`), lists, maps.
  * Infix wraps without parens.
  * Chained `DOT` across lines.
  * Ensure block headers still require the visual line break (e.g., `if … <newline>` opens a block exactly as today).
  * Ensure comments at end of a continued line don’t accidentally terminate statements:

    ```ns
    set x = a +  // comment
            b
    ```
  * Ensure triple-backticks still handle multiline strings; single-line quotes still error on raw newlines.

---

## What this *doesn’t* do (on purpose)

* It doesn’t introduce indentation sensitivity or “Python rules.” Your language stays line-oriented where it already is, and permissive where devs want to wrap things.
* It doesn’t modify precedence or expression grammar. Zero risk there.

---

## Rollout order

1. Implement Stage 0 (nesting-based NL hiding). Ship.
2. Add Stage 1 (dangling-op continuation). Ship.
3. Stop unless there’s a clear, painful need for Stage 2.

This gives you the ergonomic win (“\ is mostly unnecessary”) with the least surface area. If anything misbehaves, you can disable Stage 1 with a single feature flag while keeping Stage 0’s universally safe behavior.

If you want, I can draft the tiny Go wrapper (`nsTokenSource`) and a handful of golden tests against your current `NeuroScript.g4` so you can drop them into `pkg/parser/` and see green bars.

Love it. Let’s implement Stage 0 + Stage 1 exactly as scoped, and wire it so both the CLI parser and **nslsp** use the same behavior.

I’ll give you:

* a tiny **token-source wrapper** (`nsTokenSource`) that hides NEWLINE when nested or when a line “obviously continues”
* a clean **integration shim** so all parser entry points (CLI + tests + nslsp) get the wrapper automatically
* a **feature flag** to toggle Stage 1 without touching code
* **tests** that pin the behavior
* notes for **nslsp** (semantic tokens, on-type formatting, folding, diagnostics)

No grammar edits required.

---

# 1) Token-source wrapper (Go, ANTLR4)

Create `pkg/parser/lexwrap/ns_tokensource.go`:

```go
// Package lexwrap provides a thin TokenSource that implements Stage 0/1 newline
// elision for NeuroScript without changing the grammar or the generated lexer.
//
// Stage 0: NEWLINE is hidden inside (), [], {}.
// Stage 1: NEWLINE is hidden if the previous *emitted* token requires continuation
//          (e.g., infix operator, assignment, dot).
//
// Toggle Stage 1 with NS_CONTINUATION_DANGLING_OP=0/1 (default: 1).
package lexwrap

import (
	"os"
	"strings"

	"github.com/antlr4-go/antlr/v4"

	// Adjust import path to your generated lexer/parser package:
	"github.com/aprice2704/neuroscript/pkg/parser/gen" // e.g., module alias "gen"
)

type NSTokenSource struct {
	antlr.TokenSource
	prev antlr.Token
	nest int
	cont bool
	// cfg
	enableStage1 bool
}

// NewNSTokenSource wraps a base ANTLR lexer TokenSource.
func NewNSTokenSource(base antlr.TokenSource) *NSTokenSource {
	enable := true
	if v := strings.TrimSpace(os.Getenv("NS_CONTINUATION_DANGLING_OP")); v != "" {
		if v == "0" || strings.EqualFold(v, "false") || strings.EqualFold(v, "off") {
			enable = false
		}
	}
	return &NSTokenSource{
		TokenSource:  base,
		enableStage1: enable,
	}
}

func (s *NSTokenSource) NextToken() antlr.Token {
	for {
		t := s.TokenSource.NextToken()
		typ := t.GetTokenType()

		switch typ {
		// --- Stage 0: manage nesting and hide NEWLINE when nested > 0 ---
		case gen.NeuroScriptLexerLPAREN, gen.NeuroScriptLexerLBRACK, gen.NeuroScriptLexerLBRACE:
			s.nest++
			// opening delimiters imply continuation of the current syntactic construct
			s.cont = true

		case gen.NeuroScriptLexerRPAREN, gen.NeuroScriptLexerRBRACK, gen.NeuroScriptLexerRBRACE:
			if s.nest > 0 {
				s.nest--
			}
			// a closing delimiter is a valid "line ender"; clear continuation
			s.cont = false

		case gen.NeuroScriptLexerNEWLINE:
			// Inside delimiters: NEWLINE is whitespace.
			if s.nest > 0 {
				continue
			}
			// Stage 1: dangling-op continuation (when enabled).
			if s.enableStage1 && s.cont {
				continue
			}
			// Otherwise, surface NEWLINE (grammar still uses it at top-level).
			s.prev = t
			return t

		default:
			// Stage 1: mark whether this token requests a continuation.
			if s.enableStage1 && isDanglingContinuationToken(typ) {
				s.cont = true
			} else if typ != antlr.TokenEOF && t.GetChannel() == antlr.TokenDefaultChannel {
				// Any real token that can start/continue an expression clears pending continuation.
				s.cont = false
			}
		}

		s.prev = t
		return t
	}
}

// isDanglingContinuationToken returns true for tokens that, if they end a line,
// clearly imply the next line is a continuation (so the intervening NEWLINE
// should be dropped).
func isDanglingContinuationToken(tok int) bool {
	switch tok {
	// arithmetic / logical / bitwise infix operators (add the full set you generate)
	case gen.NeuroScriptLexerPLUS,
		gen.NeuroScriptLexerMINUS,
		gen.NeuroScriptLexerSTAR,
		gen.NeuroScriptLexerSLASH,
		gen.NeuroScriptLexerPERCENT,
		gen.NeuroScriptLexerAMP,      // &
		gen.NeuroScriptLexerPIPE,     // |
		gen.NeuroScriptLexerCARET,    // ^
		gen.NeuroScriptLexerAND,      // && (if distinct)
		gen.NeuroScriptLexerOR,       // || (if distinct)
		gen.NeuroScriptLexerEQ,       // ==
		gen.NeuroScriptLexerNEQ,      // !=
		gen.NeuroScriptLexerLT,
		gen.NeuroScriptLexerGT,
		gen.NeuroScriptLexerLTE,
		gen.NeuroScriptLexerGTE,

	// assignment / separators that expect a RHS or continuation
		gen.NeuroScriptLexerASSIGN,   // =
		gen.NeuroScriptLexerCOLON,    // :
		gen.NeuroScriptLexerCOMMA,    // allow trailing-comma line breaks

	// member / call / index openers (redundant with Stage 0, but harmless)
		gen.NeuroScriptLexerDOT,
		gen.NeuroScriptLexerLPAREN,
		gen.NeuroScriptLexerLBRACK,
		gen.NeuroScriptLexerLBRACE:
		return true
	default:
		return false
	}
}
```

**Notes**

* Replace the import path `github.com/aprice2704/neuroscript/pkg/parser/gen` with whatever you actually use for your generated ANTLR artifacts, and update the `NeuroScriptLexer...` constants to match your lexer.
* Stage 1 can be toggled at runtime: `NS_CONTINUATION_DANGLING_OP=0` to disable (useful for bisecting any edge-case).

---

# 2) Parser construction shim

Create `pkg/parser/factory/factory.go` so callers don’t forget to apply the wrapper:

```go
package factory

import (
	"github.com/antlr4-go/antlr/v4"
	"github.com/aprice2704/neuroscript/pkg/parser/gen"
	"github.com/aprice2704/neuroscript/pkg/parser/lexwrap"
)

type SourceConfig struct {
	// Future: knobs like enabling Stage 1 explicitly, etc.
}

type Bundle struct {
	Lexer  *gen.NeuroScriptLexer
	Tokens *antlr.CommonTokenStream
	Parser *gen.NeuroScript
}

func NewParserFromInputStream(in antlr.CharStream, _ *SourceConfig) *Bundle {
	lex := gen.NewNeuroScriptLexer(in)
	src := lexwrap.NewNSTokenSource(lex) // <-- the magic lives here
	tok := antlr.NewCommonTokenStream(src, antlr.TokenDefaultChannel)
	p := gen.NewNeuroScript(tok)
	return &Bundle{Lexer: lex, Tokens: tok, Parser: p}
}

// Convenience for string input.
func NewParserFromString(s string, cfg *SourceConfig) *Bundle {
	return NewParserFromInputStream(antlr.NewInputStream(s), cfg)
}
```

Now swap your existing parser construction sites (CLI, tests, **nslsp**) to call `factory.NewParserFrom...`.

---

# 3) Tests (pin Stage 0 and Stage 1)

Create `pkg/parser/lexwrap/ns_tokensource_test.go`:

```go
package lexwrap_test

import (
	"os"
	"testing"

	"github.com/antlr4-go/antlr/v4"
	"github.com/aprice2704/neuroscript/pkg/parser/factory"
	"github.com/aprice2704/neuroscript/pkg/parser/gen"
)

func parseOK(t *testing.T, src string) {
	t.Helper()
	b := factory.NewParserFromInputStream(antlr.NewInputStream(src), nil)
	b.Parser.BuildParseTrees = true
	// Replace "Program" with your actual start rule
	tree := b.Parser.Program()
	if b.Parser.GetNumberOfSyntaxErrors() != 0 || tree == nil {
		t.Fatalf("unexpected syntax errors parsing:\n%s", src)
	}
}

func TestStage0_NewlinesInsideDelimiters(t *testing.T) {
	src := `
set x = fn(
  a,
  b,
  c
)

set y = [
  1,
  2,
  3
]

set z = {
  "a": 1,
  "b": 2
}
`
	parseOK(t, src)
}

func TestStage1_DanglingOperatorContinuation_On(t *testing.T) {
	t.Setenv("NS_CONTINUATION_DANGLING_OP", "1")
	src := `
set x = 1 +
        2 +
        3

set p = a
        .b
        .c

set q = a &&
        b ||
        c
`
	parseOK(t, src)
}

func TestStage1_DanglingOperatorContinuation_Off(t *testing.T) {
	t.Setenv("NS_CONTINUATION_DANGLING_OP", "0")
	// This should fail *unless* you parenthesize, which is fine; we only assert that
	// enabling Stage 1 makes the same text valid. Here we just check we can toggle.
	src := `set x = 1 +
        2`
	b := factory.NewParserFromInputStream(antlr.NewInputStream(src), nil)
	_ = b.Parser.Program()
	if b.Parser.GetNumberOfSyntaxErrors() == 0 {
		t.Fatalf("expected syntax error with Stage 1 disabled")
	}
}

func TestCommentsOnContinuedLine(t *testing.T) {
	os.Setenv("NS_CONTINUATION_DANGLING_OP", "1")
	defer os.Unsetenv("NS_CONTINUATION_DANGLING_OP")
	src := `
set x = 1 +  // trailing comment
        2
`
	parseOK(t, src)
}

// Optional: ensure block headers still require a visible newline at top-level.
func TestIfHeaderStillNeedsNewline(t *testing.T) {
	src := `
if 1 < 2
  set x = 3
end
`
	parseOK(t, src)
}
```

Add more table cases as you extend your operator list.

---

# 4) CLI wiring

Wherever the CLI constructs the parser (e.g., in `cmd/zadeh` or `pkg/runner`), replace ad-hoc `lexer→tokens→parser` with `factory.NewParserFrom…`. That ensures Stage 0/1 behavior is uniform in daemon mode, trusted config scripts, etc.

---

# 5) **nslsp** provisions

Short version: **use the same factory** for *all* parses in the LSP server. That keeps tokens/AST consistent across:

* initial full parse
* incremental reparse (changed ranges)
* features that reparse snippets (hover on demand, code actions)
* on-type validation (after `Enter`, `]`, `)`)

Minimal diffs:

1. **Parser construction**

* Wherever nslsp builds a parser (usually a `ParserFor(doc)` helper), swap to `factory.NewParserFromInputStream(...)`.
* If you cache parse results per document, cache the `Bundle` or at least ensure fresh token streams are created via the wrapper.

2. **Semantic tokens**

* If your semantic tokens provider is AST-driven, no changes.
* If it’s lexer-token-stream-driven, you’re already reading *default-channel* tokens; newline suppression won’t alter positions of other tokens, since character offsets come from the CharStream. Keep comments on the HIDDEN channel as you do today.

3. **Diagnostics**

* Good news: many “unexpected NEWLINE” false-positives will disappear.
* Keep your error listeners the same. The wrapper doesn’t mutate indices, so line/column math still lines up.

4. **On-type formatting / indentation**

* You didn’t add Pythonic indentation rules, so your on-type indent logic is unchanged.
* If you have a “continuation line indent” heuristic, you can optionally make it smarter by peeking the previous *visible* token using the same `isDanglingContinuationToken` predicate. That’s cosmetic.

5. **Folding ranges**

* If your fold provider uses `{}`/`[]`/`()` or syntactic blocks: unaffected.
* If it inspects NEWLINE tokens explicitly: switch to computing folds from AST node spans instead of raw NEWLINE tokens (more robust anyway).

6. **Code actions / quick fixes**

* Any action that inserts line breaks after an operator or inside delimiters becomes safer (no need to append `\`).
* If you auto-wrap long lines, allow breaking **after** a dangling operator or **after commas**; both are recognized as continuations by Stage 1.

7. **Config surface**

* Plumb `NS_CONTINUATION_DANGLING_OP` through the LSP server’s environment (or expose an LSP setting that flips a process env and re-parses). Not strictly necessary; default is “on”.

---

# 6) Edge cases & guardrails

* **String literals**: raw newlines inside single-line strings should still error. You’re not touching string rules; safe.
* **Triple-quoted / multiline strings**: unchanged.
* **Comments at end of a continued line**: covered in tests—NEWLINE after comment is dropped if `cont==true`.
* **Comma-terminated lines**: Stage 1 includes `COMMA` ⇒ multiline lists/maps without braces/inside braces both feel natural.
* **EOF after a dangling operator**: still a parse error (good).
* **Recovery**: ANTLR error recovery continues to work; you’ve only hidden some NEWLINE tokens.

---

# 7) Developer ergonomics

* Default behavior: Stage 0 + Stage 1 on.
* To bisect any weird report: `NS_CONTINUATION_DANGLING_OP=0` and retry; if the bug disappears, your case is a Stage 1 corner we can handle by adding/removing a token from `isDanglingContinuationToken`.

---

## Summary

* We shipped Stage 0 + Stage 1 via a **token-source wrapper**—no grammar changes, no refactor tar pit.
* **nslsp** just needs to construct parsers through the same factory; semantic tokens/formatting/diagnostics remain stable or improve.
* Tests pin the new multiline ergonomics and can be expanded as you encounter real code.

If you want, I can tailor the operator list to your exact lexer constants (paste the relevant section of `NeuroScript.g4` ops, and I’ll map them precisely). Next good step is to wire the factory into `nslsp` and your CLI, run the tests, and try a few deliberately ugly multi-line expressions to feel the difference.
