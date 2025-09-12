# Capsule: Privilege System (Least Privilege, Audited Effects)

**Version:** 0.1  
**Audience:** Dev AI gating side effects

---

## Primer
Privileges gate side effects. Agents and interpreters run with **least privilege**. Privs are explicit, enumerable, and audited. When a capability is missing, fail closed and log the denial.

## Assumption Brake
- **H1:** Missing priv declaration at startup.
- **H2:** A helper bypasses the guard (mixed concerns).
- **Discriminating check:** Run the same path with `PRIV=*` vs minimal; compare outcomes and logs.

## Do First / Don’t Ever
- Do: Declare required privs at module/tool init.
- Do: Guard each effect site; return a typed error on deny.
- Do: Emit `fdm.priv.denied` with `trace_id` (no secrets).
- **Never** “temporarily” widen privs to pass tests unless marked **TEMP** + TODO + issue link—and keep debug until double‑green.
- Don’t: Read env or network unless declared.

## Skeleton / Steps
1. **Identify effects**: fs write? bus publish? network?
2. **Map to priv flags**: e.g., `PRIV_FS_READ`, `PRIV_BUS_PUB(topic)`, `PRIV_NET_OUT(domain?)`.
3. **Request at startup** in one place; fail on partial grants.
4. **Guard each effect** with a single helper (`RequirePriv(ctx, PRIV_X)`).
5. **Emit & log** on deny: structured log + `fdm.priv.denied`.
6. **Tests**: one pass path with minimal priv; one deny path.

## Human‑in‑loop Commands
- `PRIVS="" go test ./pkg/... -run PrivDenied -v`  
  **Success:** typed denial error + `fdm.priv.denied` observed with `trace_id`.
- `PRIVS="PRIV_FS_READ,PRIV_BUS_PUB(fdm.meta.event)" go test ./pkg/... -run PrivPass -v`  
  **Success:** pass with minimal required privs only.

## Checks / Tests
- Running without the priv fails with the expected code.
- Granting just the minimal priv passes.
- Audit log includes actor, priv, and `trace_id`.

## Gotchas
- “Convenience” helpers that do IO + network—split them, guard each effect.
- Tests that inherit wide privs from the harness—pin per test.
- Cached handles (open files) can bypass later checks—re‑validate at use.

## Evidence (fill these)
- Priv declarations: `<repo>/pkg/<mod>/priv.go`
- Enforcement helper: `<repo>/pkg/<mod>/privcheck.go`
- Event schema: `fdm.priv.denied` in `docs/capsules/bus-typed-topics.md`
