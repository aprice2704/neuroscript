:: title: AEIOU v3 – Implementation Checklist
:: version: 0.2.0
:: owner: Impl Team (AJP)
:: component: Zadeh Host / NS Interpreter / Tools
:: dependsOn: aeiou_spec_v3_3.1.md, askloop_v3_3.1.md, ns runtime, crypto/keymgmt
:: howToUse: Work top-down. Token-last commit is mandatory. Ed25519 default. Progress guard is MUST.

- |x| M0: Project setup #(m0-setup)
  - [x] Update specs to v3.1 (this diff) #(spec-update)
  - [x] Feature flag `aeiou.v3.enabled` (default off) and `aeiou.v3.progress_guard` (default on) (cancelled, v3 universal) #(fflags)
  - [x] CODEOWNERS & security review owners set (cancelled n/a) #(governance)

- |>| M1: Envelope v3 (markers, parser, caps) #(m1-envelope)
  - [x] Markers & ABNF implemented; fixed order; **first-section-wins** #(parser-impl)
  - [x] Size caps: env ≤ 1 MiB; section ≤ 512 KiB; line ≤ 8 KiB #(caps)
  - [ ] USERDATA JSON validation helper; schema errors typed #(userdata-validate)
  - [ ] Lints: out-of-order, duplicates(after first), post-token text #(env-lints)
  - [x] Tests: valid/invalid markers; duplicates; ordering; size; UTF-8/BOM #(env-tests)

- |>| M2: Control plane — magic tool & token #(m2-control)
  - [x] **Ed25519** default signing; HS256 optional (single-process) #(ed25519-default)
  - [x] Token wire: `<<<NSMAG:V3:{KIND}:{B64URL(PAYLOAD)}.{B64URL(TAG)}>>>` #(token-wire)
  - [x] Payload schema; **JCS** canonicalizer (RFC-8785); disallow floats in payload fields #(canonical-json)
  - [x] Implement `tool.aeiou.magic(kind string, payload any, opts? map) -> string` #(magic-api)
  - [x] Token constraints: length ≤1024, single-line, no quotes/backticks #(token-constraints)
  - [x] Verifier: signature, scope `{SID,turn_index,turn_nonce}`, TTL, replay (`jti`) #(verifier)
  - [x] Replay cache: per-SID **bounded LRU** + TTL (defaults: 5m, cap 4096) #(replay-bounds)
  - [ ] Key mgmt: `kid` rotation with grace ≥ max `ttl` + 60s; hot-reload keys #(key-rotation)
  - [ ] **Fallback signer**: preloaded in memory; if primary fails, mint **abort** token; else return typed error #(fallback-signer)
  - [x] Tests: altered payload fails; wrong SID/nonce/turn; expired TTL; duplicate `jti`; oversize; quoted/backticked #(ctrl-tests)

- |x| M3: Loop controller — commit-last + progress guard #(m3-loop)
  - [x] OUTPUT-only scanning; precedence `abort>done>continue`; last-wins #(selection)
  - [ ] Hard lint if text after chosen token (decision stands) #(post-token-lint)
  - [x] **Progress guard (MUST):** normalize(out/scr), strip tokens, sha256 digest; `HALT(ERR_NO_PROGRESS)` after N=3 repeats #(progress-guard)
  - [ ] Config knobs: `loop.no_progress_n` (default 3) #(progress-config)
  - [x] Tests: multi-token, no token, progress-loop detection #(loop-tests)

- | | M4: Interpreter & sandbox #(m4-interpreter)
  - [ ] Fresh interpreter per turn; per-SID mutex (ask is blocking) #(interp-fresh)
  - [ ] **Sandbox**: no ambient net/fs; capability drop; env scrub; all IO via tools; SID-scoped storage #(sandbox)
  - [ ] Tool context: every tool reads `{SID,turn}`; reject mismatches #(tool-ctx)
  - [ ] Quotas: wall/CPU/mem; cancellation; typed overrun errors #(quotas)
  - [ ] Tests: sandbox escape attempts; cross-SID access blocked; quota timeouts #(sbx-tests)

- | | M5: USERDATA discipline #(m5-userdata)
  - [ ] JSON object enforced; `subject` string, `fields` object #(schema)
  - [ ] Helpers: `userdata.get(path)`; IDs/URIs instead of large blobs #(helpers)

- |>| M6: Conformance suite (goldens) #(m6-conformance)
  - [ ] Envelopes: minimal; with SCRATCH/OUTPUT; duplicates; wrong order; oversize; non-UTF-8 #(golden-env)
  - [ ] Tokens: valid Ed25519; bad tag; wrong scope; TTL expired; `jti` dup; multi-line; quoted; oversize #(golden-tok)
  - [x] **JCS edge tests**: Unicode normalization; escapes; integer extremes; key ordering; ensure no floats #(golden-jcs)
  - [ ] Expected outcomes (JSON); CLI runner; CI gate #(conf-ci)

- | | M7: Observability & ops #(m7-obs)
  - [ ] Decision log: `{ts,SID,turn,decision,reason,kid?,jti?,latency_ms,output_bytes,scratch_bytes,verification_failure_reason?}` #(obs-log)
  - [ ] Metrics: decisions; verify_fail_total{reason}; ttl_expired_total; jti_dup_total; **lints**; no_progress_total #(obs-metrics)
  - [ ] Admin: `zadeh tokens verify <line>`; `zadeh sessions dump <SID>`; `zadeh conformance run` #(obs-admin)

- | | M8: Docs & AI-facing materials #(m8-docs)
  - [ ] Update specs with progress guard, sandbox, Ed25519 default, fallback signer text #(spec-sync)
  - [ ] “First Words” blurb (token-last pattern) #(first-words)
  - [ ] Examples: continue/done/abort; no token; progress-guard HALT #(examples)

- | | M9: Security review & pen tests #(m9-sec)
  - [ ] Threat model updated: spoofing, replay, cross-SID injection, token-length abuse, JCS pitfalls, sandbox escape, tool vulns #(threat-model)
  - [ ] Negative tests: tokens in SCRATCHPAD/USERDATA inert; echoed look-alikes inert; base64 tamper; oversize tokens #(neg-tests)
  - [ ] **Sandbox escape assessment**; **Tool vulnerability assessment**; key-rotation grace verified #(pentest)

- | | M10: Release gating & rollout #(m10-release)
  - [ ] Feature-flagged rollout; monitor verify_fail and lint rates; rollback thresholds #(rollout)
  - [ ] Freeze any “direct-decide” paths; CI checks enforce OUTPUT-only commit #(freeze-direct)
  - [ ] Tag `aeiou-v3.1.0`; release notes #(release)

- | | M11: Nice-to-haves (post-v3.1) #(m11-future)
  - [ ] External auditor verifier binary (public key only) #(auditor)
  - [ ] Visual transcript viewer highlighting control tokens and lints #(viewer)

# End of checklist