# Capsule: Making NS Tools (Authoring, Packaging, Contracts)

**Version:** 0.1  
**Audience:** Dev AI writing or modifying NS tools

---

## Primer
NS tools are small, predictable effectors. The **contract beats cleverness**. Tools must be testable without network and discoverable via their manifest.

## Assumption Brake
- **H1:** Input contract mismatch (schema/flags).
- **H2:** Nondeterministic dependency (clock/temp path).
- **Discriminating check:** Same inputs twice → byte‑identical outputs; if not, identify the nondeterminism.

## Do First / Don’t Ever
- Do: Define inputs/outputs precisely; fail with typed errors.
- Do: Provide man‑page style description + examples.
- **Never** reach outside declared inputs, privs, or the bus.
- **Never** remove verbose tool logs until goldens are green twice.

## Skeleton / Steps
1. **Manifest**: `tool.yaml` with name, version, inputs, outputs, required privs.
2. **Entry**: single `ns` command mapping inputs → outputs deterministically.
3. **Validation**: reject bad inputs with one error code + message.
4. **Idempotency**: same inputs → same outputs; log any nondeterminism.
5. **Docs**: README with examples and exit codes.
6. **Tests**: golden I/O cases + one failure case per validation rule.

## Human‑in‑loop Commands
- `ns ./tools/<NAME> -in testdata/in.json -out /tmp/out.json`  
  **Success:** exit code 0; output hash equals `testdata/golden.json.sha256`.
- `diff -u /tmp/out.json testdata/golden.json`  
  **Success:** empty diff.

## Checks / Tests
- `--help` prints name, version, inputs/outputs.
- Golden tests pass; failure test yields documented error.
- No network/env access unless declared by privs.

## Gotchas
- Hidden globals (temp dirs, clock). Add flags or deterministic mode.
- Over‑broad errors that leak internals—keep them structured and minimal.
- Exit code drift—pin in tests.

## Evidence (fill these)
- Example tool: `<repo>/tools/<NAME>/`
- Template manifest: `<repo>/tools/_template/tool.yaml`
- Test harness: `<repo>/tools/<NAME>/tool_test.go`
