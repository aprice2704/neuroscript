:: title: NeuroScript Road to v0.2.0 Checklist
 :: version: 0.2.0-alpha
 :: id: ns-roadmap-v0.2.0-alpha-update1
 :: status: draft
 :: description: Prioritized tasks to implement NeuroScript v0.2.0 based on design discussions Apr 25-29, 2025.

 # NeuroScript v0.2.0 Development Tasks

 - [x] 1. Implement Core v0.2.0 Syntax Changes (Parser/Lexer/AST Level)
   - [x] Convert all keywords to lowercase. *(Verified in grammar)*
   - [x] Implement `func <name> [(][needs...] [optional...] [returns...][)] means ... endfunc` structure (Optional Parens Handled). *(Grammar/AST builder corrected & tested)*
   - [x] Implement specific block terminators (`endif`, `endfor`, `endwhile`, `endon`). *(Verified in grammar)*
   - [x] Implement triple-backtick string parsing. *(Verified in grammar)*
   - [x] Implement default `{{placeholder}}` evaluation within ```. *(Interpreter Semantic - Pending)*
   - [x] Implement `:: keyword: value` metadata parsing (header & inline). *(Parser/AST Builder functional, verified in logs)*
   - [x] Remove old `comment:` block parsing. *(Implicitly done via grammar updates)*
   - [x] Implement `no`/`some` keyword parsing in expressions. *(Verified in grammar)*
   - [x] Implement `must`/`mustBe` statement/expression parsing. *(Verified in grammar/AST builder)*
   - [x] Implement `on_error`/`endon`/`clear_error` parsing. *(Verified in grammar/AST builder)*
   - [x] Add `ask`/`into` statement parsing. *(Verified in grammar/AST builder)*
   - [x] Implement `#` for comments (old `--` may remain). *(Verified in grammar)*
   - [x] Parse map literals `{ "key": value, ... }` *(Verified in grammar/AST builder)*
   - [x] Parse list literals `[ value, ... ]` *(Verified in grammar/AST builder)*
   - [x] Implement element access syntax `list[index]`, `map[key]`. *(Grammar/AST builder corrected & tested)*

 - [x] 2. Implement Core v0.2.0 Semantics (Interpreter Level)
   - [x] Evaluate `func` call / `returns` properly (single & multi-value). *(Interpreter `executeReturn` and `RunProcedure` return checks fixed & tested)*
   - [x] Evaluate specific block terminators correctly. *(Implied complete by passing tests, structure exists)*
   - [x] Evaluate default `{{placeholder}}` evaluation within ``` strings. *(Depends on 1.e)*
   - [x] Evaluate `no`/`some` keywords correctly.
   - [x] Evaluate element access syntax `list[index]`, `map[key]`.
   - [x] Implement `ask`/`into` semantics (Call LLM, store result).

 - [ ] 3. Implement Go Tooling (Core Tools)
   - [ ] Add go module awareness (`go.mod`).
   - [ ] Implement `goimports` functionality.
   - [ ] Implement `go test` execution.
   - [ ] Implement `go build` execution.
   - [ ] Implement `go list` parsing.
   - [ ] Implement symbol finding within Go source.
   - [ ] Implement AST modification tools (e.g., add/remove import, rename symbol, ...).
   - [ ] Tooling for Go node navigation (find parent/children, etc., e.g., `tool.getNodeChildren`).

 - [x] 4. Implement Metadata Handling
   - [x] Move metadata in markdown to eof, not sof *(Pending)*
   - [x] Implement storage/access for `::` metadata. *(Parser/AST Builder done, stored on Program/Procedure/Step)*
   - [x] Define initial vocabulary for standard metadata keys and inline annotations.

 - [x] 5. Implement Foundational Robustness Features (Phase 1 / Near-Term)
   - [x] Define and implement `must`/`mustBe` failure semantics (halt vs error?). *(Using halt-via-error, interpreter logic exists)*
   - [x] Create essential built-in check functions for `mustBe`. *(Interpreter code for basic type checks exists)*
   - [x] Implement `on_error`/`endon`/`clear_error` semantics. *(Interpreter code exists, block context handling added)*
   - [ ] Consider explicit `halt` command/mechanism? *(New Item)*

 - [x] 6. Tooling & Ecosystem
   - [x] Standardize internal tool naming to use `tool.` prefix consistently. *(Parser/AST Builder/Interpreter handle this)*
   - [ ] Update checklists/docs to reflect standardized tool names.
   - [ ] Begin development of `nsfmt` formatting tool.
   - [ ] High priority: reach point where upload, then request-update-compile loop works for golang programs

 - [ ] 7. Example App -- language flashcards *(New)*
   - [ ] Add ability to record sound clips *(New)*
   - [ ] Add ability to upload such sound clips *(New)*

 - [ ] 8. Other
   - signal and image processing tools