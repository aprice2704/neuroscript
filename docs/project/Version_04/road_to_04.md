:: title: NeuroScript Road to v0.4.0 Checklist
:: version: 001
:: id: ns-roadmap-v0.4.0
:: status: draft
:: description: Prioritized tasks to implement NeuroScript v0.4.0
:: updated: 2025-06-03

# NeuroScript v0.4.0 Development Tasks

## Vision (Summary)

Achieve an AI-driven development loop where a Supervising AI (SAI) interacts with NeuroScript (ng) and a File API to:
- Apply a prompt to each file in a file tree 
(We now have a basic human/ai chat facility in ng, which is great -- its been a long fight.

I want to get the the point where we can deliver a tree of files to a set of AI workers, have them apply instructions to the files (code initially) then have the files updated on the local drive.

For this, we need the ability to queue jobs to a work queue, then have the wm dispatch them to AIs, potentially allow the AIs to use local tools, and pass back the processed files write out to disk.

I am thinking this should mostly be a ns that ng runs.)


- Coordinate tasks, potentially using worker AIs.
- Track progress via shared checklists (Checklist tools).

## Checklist

- [ ] 1. Language Features
    - [ ] simplfy returned arg checking

- [ ] 2. WM system
    - [ ] Basic work queue with multiple workers
    - [ ] Result accumulation & Lessons Learned
    - [-] Next version of: Apply prompt to each file in a tree (file filter, sys prompt etc.) -- Map/Reduce analog
    - [ ] Investigate potential for parallel step execution? (Future/low priority)
    - [ ] Configuration: Mechanism for SAI to configure ng sandbox/tools?

- [ ] 3. Core Interpreter / Runtime
    - [ ] Error Handling: Review on_error behavior, especially interaction with return.
    - [ ] Rationalize dratted loggers
    - [ ] Handle Management: Add explicit ReleaseHandle tool? Review potential memory leaks from unreleased handles.
    - [ ] Performance: Basic profiling pass (if major slowdowns observed).

    - [ ] Consider FS.DiffFiles tool (or is NSPatch sufficient?).
    - [ ] FileAPI Review: Ensure consistency/no overlap between FS.* tools and direct FileAPI.
    - [ ] ng -> FileAPI Sync: Design and implement mechanism/tools.
    - [ ] Build Artifacts: Review GoBuild output handling; add tools if needed (e.g., FS.Copy, retrieve artifacts).

- [ ] 6. Tooling & Ecosystem
    - [ ] Documentation: Update tool docs, checklists to reflect new tools/status.
    - [ ] Formatting: Begin development of nsfmt formatting tool for .ns files.
    - [ ] Workflow Test: Create end-to-end test script simulating SAI interaction (upload, read, modify, build, test).
    - [ ] Design for stateful worker interaction and task lifecycle.
    - [ ] Tooling for SAI to assign/monitor tasks on workers.
    - [ ] Agent mode permitted, allow & deny lists

- [ ] 7. Example App -- language flashcards (New)
    - [ ] Define core features (add card, review, save/load).
    - [ ] Design data structure (simple list/map, maybe JSON file).
    - [ ] Implement basic TUI or script interaction logic.

- [ ] 9. NS LSP
    - [ ] Flag unknown tools