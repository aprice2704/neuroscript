### 1 · Reality-check on the proposed split

| Piece                          | Your intent (file) | Risk if **separate** repos                                                    | Synergy if **kept together**                                                                    |
| ------------------------------ | ------------------ | ----------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------- |
| **ns** (core grammar + stdlib) | **free**           | Hard-pledge to semantic stability; every downstream break is public.          | Independent semver; outsiders can embed or fork.                                                |
| **nd** (NeuroData defs)        | **free**           | Data-model changes cascade into FDM; version drift likely.                    | Pairing nd+ns in one repo keeps schema & language in lock-step.                                 |
| **tools** (built-ins)          | **free**           | Tools call FDM internals; split repo means importing non-free code for tests. | Put **minimal** tool shims in open repo, advanced ones in FDM repo.                             |
| **ng** (interpreter)           | **free**           | Must import nd, tools; becomes de-facto “aggregate” repo anyway.              | If ng lives with ns/nd, go-mod replace lines vanish.                                            |
| **wm** (worker manager)        | closed             | Only interacts with FDM; safe to isolate.                                     | None.                                                                                           |
| **fdm** (“Zadeh” engine)       | closed             | N /A                                                                          | N /A                                                                                            |
| **zgateway**                   | closed             | Needs ng’s Go structs for request decoding; vendoring pain.                   | Keep proto files in a public “api” repo so closed gateway can vendor them without leaking code. |
| **linter**                     | free               | OK standalone                                                                 | Could reuse parser from ns; colocate to avoid duplicate vendoring.                              |

**Take-away:** the hard dependency chain is **ns → nd → tools → ng**.
If you scatter those four across multiple repos you’ll fight `go.mod` churn and circular CI triggers. Merge them unless you have a strong branding reason not to.

---

### 2 · Recommended repo topology

#### Option A – **Two-tier open-core (simplest)**

```
github.com/zadeh-ai/open/
    ns/        (grammar, parser, stdlib)
    nd/        (schemas, codec helpers)
    tools/     (safe built-ins)
    ng/        (interpreter)
    linter/    (import ns parser)
    api/       (proto + OpenAPI for gateway)
github.com/zadeh-ai/enterprise/
    fdm/         (closed)
    wm/          (closed)
    zgateway/    (closed, imports api)
```

*Pros*: one public module path to vendor; private repos can `require github.com/zadeh-ai/open/... v0.9.0`.
*Cons*: still six public modules (ns, nd, tools, ng, linter, api); semver juggling.

#### Option B – **Single public monorepo + directory licence split**

```
/core/       MIT or Apache-2.0
    ns/
    nd/
    tools/
    ng/
    api/
/pro/        BUSL-1.1 or proprietary
    fdm/
    wm/
    zgateway/
/build/
/docs/
```

Go 1.21 `toolchain` + `go.work` let you keep two modules (`core`, `pro`) in one tree.
*Pros*: zero cross-repo PRs, atomic feature branches, one CI pipeline.
*Cons*: need licence headers and CODEOWNERS to stop accidental leaks.

**I’d pick Option B** until you truly need separate security perimeters.

---

### 3 · Licence & revenue model

| Layer                              | Licence                                                                                          | Rationale                                                                                     |
| ---------------------------------- | ------------------------------------------------------------------------------------------------ | --------------------------------------------------------------------------------------------- |
| **ns, nd, tools, ng, linter, api** | **Apache-2.0** (permissive)                                                                      | Lowers barrier for university & vendor adoption; encourages third-party gateways.             |
| **fdm, wm, zgateway**              | **Business Source (BUSL-1.1)** <br> (source-available, production use forbidden without licence) | Gives you SaaS & on-prem revenue, avoids GPL fears, still lets startups read code for audits. |
| **Docs & Specs**                   | CC-BY-SA                                                                                         | Maximises mind-share; no code implications.                                                   |

If you later open-source FDM, flip BUSL to Apache after a two-year delay (MySQL model).

---

### 4 · CI / release scaffolding

1. **Matrix build**: `core` (public) and `pro` (internal) run in the same GitHub Actions workflow; internal jobs use self-hosted runners with the private licence secret.
2. **Version tagging**: tag root `v0.9.0`, Go modules get pseudo-versions automatically; only publish semver tags from the `core/` sub-module when grammar changes.
3. **Swagger / proto sync job**: every merge to `main` regenerates `api/gen/*` and opens a PR in `zgateway` if diff detected.

---

### 5 · Free vs paid surface test

| Feature                         | Free tier?            | Metric to up-sell                      |
| ------------------------------- | --------------------- | -------------------------------------- |
| Parse & run NS locally          | ✅ forever             | –                                      |
| Linter & language server        | ✅                     | –                                      |
| FDM single-user + CLI           | ✅ (community licence) | Charge on multi-tenant or >5 GB index. |
| Zadeh Gateway clustering & auth | ❌                     | Require licence token.                 |
| Worker Manager autoscaling      | ❌                     | Paid.                                  |
| Cloud SaaS                      | –                     | Metered by script-fuel or index-GB.    |

This keeps hobbyists and OSS projects happy while reserving high-value ops features for paying teams.

---

### 6 · Immediate next steps toward 0 .9 launch

1. **Pick Option B repo skeleton** and push a stub commit; migrate one package (ns) now.
2. **Add SPDX licence headers** to every file; CI fails on missing header.
3. **Publish `api` proto** at `github.com/zadeh-ai/open/api` so early adopters can codegen clients.
4. **Draft CONTRIBUTING.md** clarifying that ns syntax changes need an NIP (NeuroScript Improvement Proposal).
5. **Write “public roadmap” issue** mirroring the TODO list from the uploaded doc so external devs can watch progress.

Completing those gives you a clean, public-ready face while keeping the monetisable engine private.

---

**Bottom line:**

* Merge the **ns + nd + tools + ng** stack; don’t scatter it.
* Put everything in one GitHub organisation; split licence boundaries by directory, not repo, until scale requires more.
* Use Apache-2.0 for language/runtime, BUSL for revenue pieces.
* Publish the `api` proto early so third-party “Zadeh Gateways” stay viable.

That layout maximises adoption, minimises dev friction, and leaves clear levers for commercial growth.
