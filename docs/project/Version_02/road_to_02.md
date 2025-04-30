:: title: NeuroScript Road to v0.2.0 Checklist
 :: version: 0.2.0-alpha
 :: id: ns-roadmap-v0.2.0-alpha-update1
 :: status: draft
 :: description: Prioritized tasks to implement NeuroScript v0.2.0 based on design discussions Apr 25-29, 2025.

 # NeuroScript v0.2.0 Development Tasks

 - [x] 1. Implement Core v0.2.0 Syntax Changes (Parser/Lexer/AST Level)
   - [x] Convert all keywords to lowercase. [cite: 1]
   - [x] Implement `func <name> [(][needs...] [optional...] [returns...][)] means ... endfunc` structure (Optional Parens Handled). [cite: 1]
   - [x] Implement specific block terminators (`endif`, `endfor`, `endwhile`, `endon`). [cite: 1]
   - [x] Implement triple-backtick string parsing. [cite: 1]
   - [ ] Implement default `{{placeholder}}` evaluation within ```. *(Interpreter Semantic - Pending)*
   - [x] Implement `:: keyword: value` metadata parsing (header & inline). *(Parser/AST Builder functional)* [cite: 1, 4]
   - [x] Remove old `comment:` block parsing. *(Implicitly done)*
   - [x] Implement `no`/`some` keyword parsing in expressions. [cite: 1]
   - [x] Implement `must`/`mustBe` statement/expression parsing. [cite: 1]
   - [x] Implement `on_error`/`endon`/`clear_error` syntax parsing. *(Replaced try/catch)* [cite: 1]
 - [-] 2. Implement Core v0.2.0 Semantics (Interpreter Level)
   - [ ] Implement `askAI`/`askHuman`/`askComputer` execution logic. *(Not reviewed in this session)*
   - [ ] Implement initial `handle` mechanism for `ask...` functions (e.g., string identifiers). *(Not reviewed)*
   - [ ] Implement direct assignment from `ask...` and returning `call tool...` (e.g., `set result = askAI(...)`). *(See function return status below)*
   - [x] Implement multiple return value handling (`returns` clause). *(Interpreter code exists)* [cite: 2, 6]
   - [x] Implement `no`/`some` keyword evaluation logic (runtime type zero-value checks). *(Interpreter code exists)* [cite: 3]
   - [x] Ensure standard comparison and arithmetic operators function correctly. *(Verified via test fixes)*
 - [ ] 3. Implement Tree Data Type & Core Tools (Phase 1 / Near-Term)
   - [ ] Design internal Go representation for `tree` handle type.
   - [ ] Implement `tool.listDirectoryTree` (or `tool.walkDir`) returning `tree` handle.
   - [ ] Implement basic tree manipulation tools (e.g., `tool.getNodeValue`, `tool.getNodeChildren`).
 - [x] 4. Implement Metadata Handling
   - [ ] Move metadata in markdown to eof, not sof *(Pending)*
   - [x] Implement storage/access for `::` metadata. *(Parser/AST Builder done, stored on Program/Procedure/Step)* [cite: 4]
   - [ ] Define initial vocabulary for standard metadata keys and inline annotations.
 - [x] 5. Implement Foundational Robustness Features (Phase 1 / Near-Term)
   - [x] Define and implement `must`/`mustBe` failure semantics (halt vs error?). *(Using halt-via-error)* [cite: 2]
   - [x] Create essential built-in check functions for `mustBe`. *(Interpreter code exists)* [cite: 3]
   - [x] Implement `on_error`/`endon`/`clear_error` semantics. *(Interpreter code exists)* [cite: 2]
   - [ ] Consider explicit `halt` command/mechanism? *(New Item)*
 - [x] 6. Tooling & Ecosystem
   - [x] Standardize internal tool naming to use `tool.` prefix consistently. *(Verified via test fixes)* [cite: 1]
   - [ ] Update checklists/docs to reflect standardized tool names.
   - [ ] Begin development of `nsfmt` formatting tool.
 - [ ] 7. Example App -- language flashcards *(New)*
   - [ ] Add ability to record sound clips *(New)*
   - [ ] Add ability to upload such sound clips *(New)*