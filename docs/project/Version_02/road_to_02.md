 :: title: NeuroScript Road to v0.2.0 Checklist
 :: version: 0.2.0-alpha
 :: id: ns-roadmap-v0.2.0-alpha
 :: status: draft
 :: description: Prioritized tasks to implement NeuroScript v0.2.0 based on design discussions Apr 25, 2025.

 # NeuroScript v0.2.0 Development Tasks

 - [-] 1. Implement Core v0.2.0 Syntax Changes (Parser/Lexer Level)
   - [x] Convert all keywords to lowercase.
   - [x] Implement `func <name> [needs...] [optional...] [returns...] means ... endfunc` structure.
   - [x] Implement specific block terminators (`end if`, `end for`, `end while`, `endtry`).
   - [x] Implement triple-backtick string parsing.
   - [x] Implement default `{{placeholder}}` evaluation within ```. *(Interpreter Semantic)*
   - [x] Implement `:: keyword: value` metadata parsing (header & inline). *(Parser/AST Builder done)*
   - [x] Remove old `comment:` block parsing.
   - [x] Implement `no`/`some` keyword parsing in expressions.
   - [x] Implement `must`/`mustBe` statement/expression parsing.
   - [x] Implement basic `try`/`catch`/`finally` structure parsing.
 - [-] 2. Implement Core v0.2.0 Semantics (Interpreter Level)
   - [x] Implement `askAI`/`askHuman`/`askComputer` execution logic.
   - [ ] Implement initial `handle` mechanism for `ask...` functions (e.g., string identifiers).
   - [ ] Implement direct assignment from `ask...` and returning `call tool...` (e.g., `set result = askAI(...)`).
   - [x] Implement multiple return value handling (`returns` clause). *(Grammar, AST, Interpreter, Semantic Check)*
   - [x] Implement `no`/`some` keyword evaluation logic (runtime type zero-value checks).
   - [x] Ensure standard comparison and arithmetic operators function correctly. *(Verified via test fixes)*
 - [ ] 3. Implement Tree Data Type & Core Tools (Phase 1 / Near-Term)
   - [ ] Design internal Go representation for `tree` handle type.
   - [ ] Implement `tool.listDirectoryTree` (or `tool.walkDir`) returning `tree` handle.
   - [ ] Implement basic tree manipulation tools (e.g., `tool.getNodeValue`, `tool.getNodeChildren`).
 - [-] 4. Implement Metadata Handling
     [ ] Move metadata in markdown to eof, not sof
   - [x] Implement storage/access for `::` metadata. *(Parser/AST Builder done)*
   - [ ] Define initial vocabulary for standard metadata keys and inline annotations.
 - [-] 5. Implement Foundational Robustness Features (Phase 1 / Near-Term)
   - [x] Define and implement `must`/`mustBe` failure semantics (halt vs error?). *(Using halt-via-error)*
   - [x] Create essential built-in check functions for `mustBe`. *(Implemented internally)*
   - [ ] Design simple `try/catch` mechanism and semantics. *(Parser done, Semantics needed - Superseded by on_error)*
   - [ ] Consider explicit `halt` command/mechanism? *(New Item)*
 - [x] 6. Tooling & Ecosystem
   - [x] Standardize internal tool naming to use `tool.` prefix consistently. *(Verified via test fixes)*
   - [ ] Update checklists/docs to reflect standardized tool names.
   - [ ] Begin development of `nsfmt` formatting tool.
 - [ ] 7. Example App -- language flashcards *(New)*
   - [ ] Add ability to record sound clips *(New)*
   - [ ] Add ability to upload such sound clips *(New)*