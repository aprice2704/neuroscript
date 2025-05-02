 :: title: NeuroScript Road to v0.3.0 Checklist
 :: version: 0.3.0-alpha
 :: id: ns-roadmap-v0.3.0
 :: status: draft
 :: description: Prioritized tasks to implement NeuroScript v0.3.0 based on design discussions Apr 25-29, 2025.

 # NeuroScript v0.3.0 Development Tasks

 - [x] 1. Language features
     [x] `break` and `continue` flow control words

 - [ ] 3. Implement Go Tooling (Core Tools)
   - [ ] Add go module awareness (`go.mod`).
   - [ ] Implement `goimports` functionality.
   - [ ] Implement `go test` execution.
   - [ ] Implement `go build` execution.
   - [ ] Implement `go list` parsing.
   - [ ] Implement symbol finding within Go source.
   - [ ] Implement AST modification tools (e.g., add/remove import, rename symbol, ...).
   - [ ] Tooling for Go node navigation (find parent/children, etc., e.g., `tool.getNodeChildren`).

 - [x] 6. Tooling & Ecosystem
   - [ ] Update checklists/docs to reflect standardized tool names.
   - [ ] Begin development of `nsfmt` formatting tool.
   - [ ] High priority: reach point where upload, then request-update-compile loop works for golang programs

 - [ ] 7. Example App -- language flashcards *(New)*
   - [ ] Add ability to record sound clips *(New)*
   - [ ] Add ability to upload such sound clips *(New)*

 - [ ] 8. Other
   - [ ]...