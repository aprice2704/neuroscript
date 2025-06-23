# NeuroScript v0.9.0 Development Tasks

## Vision (Summary)

 - Prep for public release
 - Must be usable for real-world tasks (must have done most of MUST.Overall)

### MUST

- | | 1. Overall
    - [ ] a. refactored into ns, nd, tools, ng, wm, fdm, zgateway, linter
        - [ ] 1. ns -- neuroscript def and libs **free**
        - [ ] 2. nd -- neurodata defs and libs **free**
        - [ ] 3. tools -- ns tools **free**
        - [ ] 4. ng -- ns interpreter **free**
        - [ ] 5. wm -- worker manager
        - [ ] 6. fdm -- zadeh
        - [ ] 7. zgateway -- zadeh gateway
        - [ ] 8. linter -- source code linter for all projects
    - [ ] b. all code brutally linted, well commented, godoc
    - [ ] c. extensive test suites for all areas
    - [ ] d. logging rationalized and made uniform
    - [ ] e. testing rationalized and made uniform
    - [ ] f. docs updated

- | | 2. ns
    - [ ] a. ns syntax locked down, decide compatibility guarantee
    - [ ] b. extensive test suite
    - [ ] c. core refactored into parser, astbuilder, interpreter, tools

- | | 3. nd
    - [ ] a. "offical" first set defined
    - [ ] b. parsers working enough to load into fdm
    - [ ] c. operating subset defined and implemented (checklist, acl, factlist at least)

- | | 4. fdm
    - [ ] a. filesystem fully working (esp go)
    - [ ] b. go update/debug/compile loop working well
    - [ ] c. discussion working
    - [ ] d. frontdoor working