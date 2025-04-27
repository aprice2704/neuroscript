:: title: NeuroScript Road to v0.2.0 Checklist
:: version: 0.2.0-alpha
:: id: ns-roadmap-v0.2.0-alpha
:: status: draft
:: description: Prioritized tasks to implement NeuroScript v0.2.0 based on design discussions Apr 25, 2025.

# NeuroScript v0.2.0 Development Tasks

- [ ] 1. Implement Core v0.2.0 Syntax Changes (Parser/Lexer Level)
  - [ ] Convert all keywords to lowercase.
  - [ ] Implement `func <name> [needs...] [optional...] [returns...] means ... endfunc` structure.
  - [ ] Implement specific block terminators (`end if`, `end for`, `end while`).
  - [ ] Implement triple-backtick string parsing.
  - [ ] Implement default `{{placeholder}}` evaluation within triple-backtick strings.
  - [ ] Implement `:: keyword: value` metadata parsing (header & inline).
  - [ ] Remove old `comment:` block parsing.
  - [ ] Implement `no`/`some` keyword parsing in expressions.
  - [ ] Implement `must`/`mustBe` statement/expression parsing.
  - [ ] Implement basic `try`/`catch`/`finally` structure parsing.
- [ ] 2. Implement Core v0.2.0 Semantics (Interpreter Level)
  - [ ] Implement `askAI`/`askHuman`/`askComputer` execution logic.
  - [ ] Implement initial `handle` mechanism for `ask...` functions (e.g., string identifiers).
  - [ ] Implement direct assignment from `ask...` and returning `call tool...` (e.g., `set result = askAI(...)`).
  - [ ] Implement multiple return value handling (`returns` clause).
  - [ ] Implement `no`/`some` keyword evaluation logic (runtime type zero-value checks).
  - [ ] Ensure standard comparison and arithmetic operators function correctly[cite: 4].
- [ ] 3. Implement Tree Data Type & Core Tools (Phase 1 / Near-Term) [cite: 1, 3, 5]
  - [ ] Design internal Go representation for `tree` handle type.
  - [ ] Implement `tool.listDirectoryTree` (or `tool.walkDir`) returning `tree` handle.
  - [ ] Implement basic tree manipulation tools (e.g., `tool.getNodeValue`, `tool.getNodeChildren`).
- [ ] 4. Implement Metadata Handling
    [ ] Move metadata in markdown to eof, not sof
  - [ ] Implement storage/access for `::` metadata.
  - [ ] Define initial vocabulary for standard metadata keys and inline annotations.
- [ ] 5. Implement Foundational Robustness Features (Phase 1 / Near-Term) [cite: 1, 5]
  - [ ] Define and implement `must`/`mustBe` failure semantics (halt vs error?).
  - [ ] Create essential built-in check functions for `mustBe`.
  - [ ] Design simple `try/catch` mechanism and semantics (implementation might extend beyond 0.2.0).
- [ ] 6. Tooling & Ecosystem
  - [ ] Standardize internal tool naming to use `tool.` prefix consistently[cite: 1, 2, 3, 4].
  - [ ] Update checklists/docs to reflect standardized tool names.
  - [ ] Begin development of `nsfmt` formatting tool[cite: 5].