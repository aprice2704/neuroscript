### Moving **FDM** into the same repo that already hosts **NeuroScript (ns)**

*(i.e. one canonical Git repository for everything)*

---

## 1 · Immediate benefits

| Upside                     | Why it really helps                                                                                        |
| -------------------------- | ---------------------------------------------------------------------------------------------------------- |
| **Atomic commits**         | Grammar tweak in `ns/` + index-logic fix in `fdm/` land together → zero chase-down of cross-repo breakage. |
| **Single CI pipeline**     | One `go test ./...` and one release tag; no dance of version bumps and replace directives.                 |
| **Simpler IDE experience** | Contributors open a single workspace; “go to definition” jumps across runtime and engine.                  |
| **Easier refactors**       | Rename a core type and run `gofumpt` globally; no need for multi-repo PR chains.                           |
| **Shared vendoring**       | One `go.sum`; third-party deps deduplicated, smaller mod cache.                                            |

---

## 2 · New headaches you must police

| Risk                              | Why it bites                                                                                                         | Mitigation                                                                                                                                                                        |
| --------------------------------- | -------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Licence leakage**               | Closed-source FDM files now live beside Apache-2.0 code; a careless `git push` from a public fork could expose them. | *Enforce directory-level ACLs*: keep `fdm/` on a **private** sub-module or protected branch; add CI check that public workflows fail if they detect `internal_only = true` files. |
| **Public PR friction**            | OSS contributors can’t run tests that reach into `fdm/`; they’ll see red CI.                                         | Provide a **`make oss`** target that builds only open dirs; pull-request CI uses that target.                                                                                     |
| **Repo weight**                   | FDM binaries, fixtures, and embeddings may balloon clone size.                                                       | Use `git-lfs` or keep heavy artefacts in a separate storage bucket; gate large files with `*.bin filter=lfs` rules.                                                               |
| **Security audit boundary blurs** | Having closed and open code intertwined makes automated SBOM and licence scans harder.                               | Tag each directory with an SPDX `LICENSE` file; run `go-licenses` separately for `open/` and `pro/` sub-modules.                                                                  |

---

## 3 · Practical structure that keeps things safe

```
repo-root/
├── open/                 # Apache-2.0
│   ├── ns/               # grammar, parser, stdlib
│   ├── nd/               # NeuroData schemas
│   ├── tools/            # safe built-ins
│   ├── ng/               # interpreter
│   ├── linter/
│   └── api/              # proto + OpenAPI
├── pro/                  # BUSL-1.1 or closed
│   ├── fdm/              # “engine” proper
│   ├── wm/               # worker manager
│   └── gateway/          # Zadeh Gateway
└── go.work               # stitches open and pro modules
```

* **Two Go modules** (`open`, `pro`).
  `open/go.mod` is public; `pro/go.mod` *requires* `../open` via a `replace`.
* **Two independent CI jobs**
  `oss.yaml` runs unit tests on `open/` only, triggered by PRs from forks.
  `internal.yaml` runs full matrix but lives in a private workflow, triggered on `main` merges.
* **GitHub branch protection**
  `main` is private by default; a public mirror (e.g. `mirror/main`) receives only the `open/` subtree via a scheduled script or `git subtree push --prefix open`.

This mirrors how **HashiCorp** managed Terraform (AGPL core + enterprise dirs) before relicensing.

---

## 4 · Decision checklist

| Question                                                                                                    | If “yes”, monorepo is probably worth it |
| ----------------------------------------------------------------------------------------------------------- | --------------------------------------- |
| Do you make frequent cross-cutting changes that span `ns` and `fdm`?                                        | ✅                                       |
| Will the same core team develop both open and closed layers for the next 12–18 months?                      | ✅                                       |
| Can you enforce directory-based access controls or a publish script to strip private code before mirroring? | ✅                                       |
| Are external contributors mostly *language* or *spec* focused rather than engine internals?                 | ✅                                       |

If any of those flip to “no”, keep FDM in its own private repo and depend on tagged releases of the open repo.

---

### TL;DR

* **Putting FDM inside the NS repo is perfectly workable** and buys dev velocity, as long as you invest in tooling that prevents accidental licence bleed.
* Use a **two-module, two-CI-job, directory-licence** layout; mirror only the `open/` subtree publicly.
* Revisit the arrangement when you have truly independent teams or external engine contributors—until then, monorepo friction is lower than multi-repo drift.
