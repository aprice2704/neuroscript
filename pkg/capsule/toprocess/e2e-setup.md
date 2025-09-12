# Capsule: Complex E2E Setup (Bus + Interpreter vs Config Interpreter)

**Version:** 0.1  
**Audience:** Dev AI orchestrating an end‑to‑end test

---

## Primer
Complex E2E requires the **FDM bus** and one (or both) of the **interpreters**. The runtime `ns` interpreter executes scripts; the **config interpreter** evaluates configuration flows with different privileges and init. Treat them as two distinct modes.

## Assumption Brake
- **H1:** Using the wrong interpreter path (runtime vs config) causes the failure.
- **H2:** Race on bus subscription/producer ordering.
- **Discriminating check:** Start the minimal set of components and assert the first expected topic with a stable `trace_id`.

## Do First / Don’t Ever
- Do: Decide up front which path you need: **runtime ns** or **config ns**—they are not swappable.
- Do: Start **bus first**, then interpreters, then producers/consumers.
- Do: Keep runs hermetic (temp dirs, fixed ports/sockets).
- **Never** remove debug logs before two green E2E runs.
- Don’t: Share stateful dirs between runs.

## Skeleton / Steps
1. **Topology**: create `testdata/e2e/topo.yaml` listing: bus, ns, config‑ns, stubs.
2. **Start bus**: ephemeral instance; capture AF_UNIX socket path or port.
3. **Start interpreters**:  
   - `ns` runtime with script path and minimal privileges.  
   - `config-ns` *only* if the scenario exercises config evaluation.
4. **Seed fixtures**: publish `testdata/e2e/fixtures/*.json` to the bus.
5. **Run scenario**: an orchestrator script emits steps via topics.
6. **Assert via topics**: subscribe for expected outputs and timeouts.
7. **Teardown**: kill processes; assert no orphan subscribers.

## Human‑in‑loop Commands
- `./scripts/e2e_start.sh && ./scripts/e2e_run.sh scenario.yaml`  
  **Success:** expected `fdm.*` topics observed within 30s; same `trace_id` through the flow.
- `./scripts/e2e_teardown.sh`  
  **Success:** no processes left; bus shows zero active subs.

## Checks / Tests
- Required topics observed: list them explicitly for the scenario.
- Trace continuity verified end‑to‑end.
- Run time under fixed timeout; no leaked goroutines.

## Gotchas
- Config interpreter cannot execute runtime tools—pick the correct one.
- Event ordering under load: rely on monotonic timestamps + idempotent consumers.
- Don’t hard‑code paths; use temp dirs and pass them via env/flags.

## Evidence (fill these)
- Bus typed topics schema: `docs/capsules/bus-typed-topics.md`
- Interpreter entrypoints: `cmd/ns/main.go`, `cmd/nsconfig/main.go`
- Example orchestrator: `<repo>/scripts/e2e_run.sh` or `cmd/e2e/main.go`
