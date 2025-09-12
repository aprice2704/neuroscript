# Capsule: Debug Output (Keep Signal, Kill Noise)

**Version:** 0.1  
**Audience:** Dev AI deciding how/what to log

---

## Primer
Debug output is a *contract with future you and other agents*. Remove noise, not signal. Logs are structured, bounded, and **trace‑linked**. They must make failures diagnosable without guessing.

## Do First / Don’t Ever
- Do: Use structured logging (key/value); include `trace_id`, `component`, `event`.
- Do: Log *before* and *after* any side effect (intent → result).
- Do: Redact secrets; bound size (emit counts/hashes for large data).
- **Never** remove or demote debug logs until targeted tests are **green twice**. Keep the diff small; keep the logs.
- Don’t: Replace logs with comments; comments don’t run at 2 a.m.

## Skeleton / Steps
1. **Choose levels**: `debug` (flow), `info` (state change), `warn/error` (invariant breach).
2. **Shape entries** with `msg`, `trace_id`, `actor`, `topic/symbol`, optional `delta`.
3. **Bound** arrays/strings; include counts & hashes for truncation.
4. **Redact**: mark `redacted:true`; log reasons, not contents.
5. **Test**: snapshot representative log lines for key paths.

## Human‑in‑loop Commands
- Run targeted tests with logs on: `LOG_LEVEL=debug go test ./<pkg> -run <TestName> -v`  
  **Success:** log lines show intent→effect with `trace_id` present.
- Grep a trace: `grep "trace_id=<ID>" <logfile> | head -n 50`  
  **Success:** coherent sequence without secret leakage.

## Checks / Tests
- End‑to‑end trace continuity in E2E.
- No secrets present in captured logs (scanner passes).
- Under load, log volume under cap; no hot‑path regressions.

## Gotchas
- Streaming handlers: log open/close boundaries, not each chunk.
- Retries: log attempt numbers; avoid double‑counting.
- Overconfidence trap: a single “ah‑ha” log isn’t proof—pair it with what *would* differ if you were wrong.

## Evidence (fill these)
- Logging helper: `<repo>/pkg/<mod>/log.go`
- Redaction policy: `docs/security/redaction.md`
- Log parsing tests: `<repo>/pkg/<mod>/log_test.go`
