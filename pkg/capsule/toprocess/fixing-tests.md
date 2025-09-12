# Capsule: Fixing Tests Efficiently (Unit & Integration)

**Version:** 0.1  
**Audience:** Dev AI about to modify code  
**Meta-rule:** You’re smart but not omniscient. Write down two hypotheses, prove one with a discriminating test, and keep debug trails until tests are green **twice**.

---

## Primer
Broken tests are signal. The aim is *fast local isolation* and the *smallest fix that preserves invariants*. Do not widen assertions or change public behavior unless the spec says so.

## Assumption Brake
- **H1:** The test reflects the intended spec; production drifted.  
- **H2:** The test is correct, but a nondeterministic precondition (time/seed/fs/net) makes it flaky.  
- **Discriminating check:** Run the target test 10× with fixed seed/env. If it flips pass/fail, treat as flake first.

## Do First / Don’t Ever
- Do: Run **the single failing test** with `-run` and subtest filters.
- Do: Read the test name, table cases, and contract comment before code.
- Do: Add *temporary* debug logs on the tested path.
- **Never** remove or reduce debug logs until the targeted tests are **green twice** (back‑to‑back runs).
- Don’t: Silence asserts or broaden tolerances just to make green.
- Don’t: Change public behavior without updating the spec/test doc.

## Skeleton / Steps
1. **Narrow scope**: `go test ./<pkg> -run <TestName> -v` (use `/<SubtestName>$` for subtests).
2. **Stabilize**: fix seeds/env used in the test; disable parallelism if relevant.
3. **Extract the invariant**: copy the test’s intent into a one‑line rule.
4. **Instrument**: add debug logs *only on the path under test*.
5. **Fix minimally**: prefer tightening pre/postconditions to match the test.
6. **Deflake check**: `-count=10` on the single test.
7. **Sweep**: run the package, then repo: `go test ./...`.

## Human‑in‑loop Commands
- `go test ./<pkg> -run <TestName> -v -count=10`  
  **Success:** 10/10 passing, no retries.
- `go test ./...`  
  **Success:** no new failures elsewhere.

## Checks / Tests
- The target test passes 10× in a row.
- No unrelated tests fail after the change.
- The diff is minimal (fewest touched files/lines).

## Gotchas
- Table‑driven tests hide the failing case—print the case index and name.
- Snapshots/goldens: regenerate only after re‑reading the spec; commit both updated golden and reason.
- “Found a bug” is not “found **the** bug.” If first fix fails, **revert** and run your discriminating check.

## Evidence (fill these)
- Example failing test path: `<repo>/pkg/<mod>/<file>_test.go:Test<...>`
- Spec / invariants doc: `docs/capsules/<module>-invariants.md`
- Flake harness (if any): `<path>`
