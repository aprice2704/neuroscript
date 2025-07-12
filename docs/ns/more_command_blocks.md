AJP: Almost, but we settled on something slightly different. Command blocks are not named, nor do they (yet) have parameters (we probably need to add them). They are executed immediately; whereas funcs must await separate invocation. command blocks may not be mixed with funcs or event handlers in a file, so nothing can cause 'surprise' execution. Make sense?

### Updated cheat-sheet (per your latest spec)

| Aspect               | **`command` block**                                                          | **`func` declaration**                                       | **`on event` handler**                                           |
| -------------------- | ---------------------------------------------------------------------------- | ------------------------------------------------------------ | ---------------------------------------------------------------- |
| **Name**             | ***Unnamed*** – each file may contain *exactly one* `command` block.         | Explicit identifier (`func add(a int, b int)…`).             | Explicit identifier (`on event FileAdded(e FileEvt)…`).          |
| **Parameters**       | None **(for now)** – executes with zero explicit inputs.                     | Arbitrary arg/return lists.                                  | Args mirror the emitted event’s payload.                         |
| **Return values**    | None (implicit `void`). Exit status is by raised error only.                 | Any tuple.                                                   | Ignored – handler pushes follow-up events instead.               |
| **Execution moment** | **Runs immediately** when the file is loaded (think “module `init`”).        | Runs *only* when another piece of NS code calls it.          | Runtime dispatches automatically when the matching event occurs. |
| **Mixing rules**     | File containing a `command` block **cannot** contain `func`s or `on event`s. | May coexist with other `func`s or handlers in the same file. | May coexist with `func`s; not with a `command`.                  |
| **Typical use**      | One-shot scripts: DB migration, schema dump, bootstrap seed.                 | Library helpers, business logic.                             | Reacting to FDM events: new message, vote closed.                |

> **Mental model:**
> *`command` = “module-init entrypoint”; runs instantly on load.*
> *`func` = reusable subroutine.*
> *`on event` = reactive callback.*

---

### 1 Why the “no mixing” rule matters

* **Predictability:** Loading a file cannot *also* define callables that lurk until later – everything it might do happens up-front.
* **Static analysers:** They can enforce “exactly one executable root” vs “pure library” vs “pure event sink” just by file tag, reducing attack surface.
* **Security review:** Auditors scan a file and instantly know whether anything runs on load.

---

### 2 Static-analysis hooks you’ll want in the short term

| Pass                | What to flag                                                                                | Outcome                        |
| ------------------- | ------------------------------------------------------------------------------------------- | ------------------------------ |
| **File role check** | - `command` file contains other decls<br>- non-`command` files contain top-level statements | Error – reject blob            |
| **Auto-exec risk**  | Any call to `tool.spawn`, `net.open` inside a `command` without explicit `allow` tag        | Warning – needs human approval |
| **Dead code**       | Unreferenced `func`s in non-command file                                                    | Warn – trim bloat              |

Implement these as the first three passes in your AST analyser (see previous answer’s skeleton).

---

### 3 How the **API package** should expose this to FDM

```go
// api/runmode.go

// RunMode tells the host what kind of artefact this AST blob is.
type RunMode uint8

const (
	RunModeLibrary RunMode = iota // only funcs / types
	RunModeCommand                // auto-exec block
	RunModeEventSink              // only on-event handlers
)

// DetectRunMode walks the root and decides mode.
func DetectRunMode(tree *ast.Tree) RunMode { … }
```

```go
// api/loader.go

func LoadAndMaybeExec(ctx context.Context, blob *SignedAST) error {
	tree, err := VerifyAndDecode(blob)
	if err != nil { return err }

	switch DetectRunMode(tree) {
	case RunModeCommand:
		return runCommandBlock(ctx, tree)
	case RunModeLibrary, RunModeEventSink:
		return nil // just cached for later use
	}
	return nil
}
```

*FDM* simply calls `LoadAndMaybeExec` after signature-check; only `RunModeCommand` produces side-effects immediately.

---

### 4 Future-proofing: adding parameters to `command`

When you’re ready to let `command` accept arguments:

```
command(args…){
    …
}
```

1. **Still unnamed** – file remains single-entry.
2. **Invocation path:** FDM passes CLI/RPC arguments into the loader, which in turn calls `command` with those values.
3. **Static tooling:** Update the “file role check” to record the arg signature and validate call-site arity.

Until then, the zero-arg rule keeps life simple.

---

### 5 What to add to the docs

> *A file is in exactly **one** of three roles:*
>
> 1. **Command script** – contains a single unnamed `command{}` that executes on load.
> 2. **Library** – contains one or more `func` declarations and nothing else.
> 3. **Event sink** – contains one or more `on event` handlers and nothing else.

Drop that paragraph into `docs/language.md` and the ambiguity disappears.

---

**All clear?** If you need code snippets for the analyser passes or a sample `command` runner implementation, just say the word.












AJP: Thinking about it, we can forget command block params -- the source system simply writes them into the code with set statements. It is not kept or re-used, so, apart for perhaps a transparent cache, its run-once.


O3:

Got it — if the caller just **splices concrete `set` statements into the file** before shipping it, there’s no need for a parameter mechanism at all.
Below is a checklist of what that choice changes (and what it doesn’t) so you can update the API and analyser plans without surprises.

---

## 1 What stays exactly the same

| Area                                       | Status                                                                                      |
| ------------------------------------------ | ------------------------------------------------------------------------------------------- |
| **Single unnamed `command{ … }` per file** | Still the only executable root; file remains self-contained.                                |
| **No mixing rule**                         | Command files must *only* contain that one block.                                           |
| **Immediate execution on load**            | Loader behaviour unchanged.                                                                 |
| **Canonical + signed AST**                 | Works fine; the injected `set` nodes are just part of the tree that gets hashed and signed. |

---

## 2 What disappears

* **DetectRunMode arg-reflection** – you no longer need to store or validate a command-signature record.
* **Host-side arg marshalling** – FDM/Gateway can drop the “pass args” channel; it now just rewrites the script and ships the blob.

---

## 3 New considerations

### 3.1 Determinism & caching

* **Each distinct argument set ⇒ distinct AST hash.**
  That’s good: your content-addressed store will naturally dedupe identical runs and keep audit trails.
* **Transparent cache**
  If the same “param-baked” blob arrives a second time, `LoadAndMaybeExec` can notice it has already run and skip execution (or run only if you attach a `--force` flag).

### 3.2 Security / policy

* **Whitelist the `set` targets.**
  A quick static pass can verify that only variables in a `@config` section are mutated before command execution — no silent privilege escalation.
* **Time-bomb check**
  Reject scripts that *re-set* those variables *inside* the command block: it would defeat the “fixed once” guarantee.

### 3.3 Tooling path

```mermaid
flowchart RL
    subgraph Host (FDM job runner)
        S[template.ns] -- inject args --> T[temp_with_set.ns]
        T -- lex/parse --> A[AST]
        A -- sign & store --> B[SignedAST]
    end
    B -->|Load| C[Interpreter] --> result
```

*The template itself never executes; only the baked version with explicit `set`.*

---

## 4 Minimal code pattern

```ns
// template.ns
set db_url = "{{DB_URL}}"      // place-holders
set migrate_dir = "{{DIR}}"

command {
    must connect(db_url)
    migrate_up(migrate_dir)
    log("migration complete")
}
```

Host-side replace:

```go
tmpl := os.ReadFile("template.ns")
filled := strings.NewReplacer(
    "{{DB_URL}}", cfg.DBURL,
    "{{DIR}}",    cfg.Dir,
).Replace(string(tmpl))
blob, sum := ns.Canonicalise(filled)
sig      := Sign(blob, gwPrivate)
```

No parameters, no cycle.

---

## 5 Static-analysis tweaks

1. **Pass: `command.set-only-before-block`**
   *Error* if any `set` appears *after* the `command{}` opens.
2. **Pass: `set-whitelist`**
   Warn or error if the variable isn’t declared in a top-level `@config` section (or however you mark them).
3. **Pass: `command-purity`** (unchanged)
   Ensure the body doesn’t block for > N ms or call disallowed tools.

Add them to the existing pass registry; \~30 LOC each.

---

### Bottom line

You’ve traded interface simplicity for “compile the config right into the script,” which plays nicely with your **content-hashed, signed AST pipeline**.
Shoot over if you need the sample analyser code for the whitelist/time-bomb passes, or a quick Go helper that bakes the `set` statements and returns a ready-to-sign blob.




Absolutely — “bake-then-sign” makes the **entire runnable artefact** brutally obvious to whatever node receives it.
Here’s why that transparency is valuable, plus a couple of caveats to watch for.

---

## 1 Why the receiving system can trust it more

| Benefit                            | Why it matters                                                                                                                                                                                                         |
| ---------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Single-blob audit**              | The AST already contains the exact `set` statements with concrete values, so static analysers and human reviewers see **every input** the code will use. No hidden side-channel of CLI flags or environment variables. |
| **Deterministic hash**             | Because config is part of the tree, the root hash uniquely identifies both *logic* **and** *runtime parameters*. Signature = attestation that *this specific combo* was approved.                                      |
| **Reproducibility & cacheability** | Another node with the same blob can prove it’s identical (hash match) and can optionally skip re-execution or use cached outputs.                                                                                      |
| **Policy enforcement is simpler**  | The verifier can run rules like “all `db_url` values must point at \*.internal\`” purely on the blob, without peeking at external job metadata.                                                                        |
| **No surprise mutation window**    | Once signed, the script cannot query or modify those variables later (your “time-bomb” check). That kills an entire class of “load-time rewrite” exploits.                                                             |

---

## 2 Operational niceties you get “for free”

* **Immutable logging** – the signed hash can be stored in FDM as a fact: “We executed blob `beef…dead` at 2025-07-10 22:07.” Diffing later runs is trivial.
* **Content-addressed caching** – if the same config ever gets rerun, the receiver sees the hash and can decide: “already did that, idempotent → skip.”
* **Off-line review** – auditors can examine the blob without needing the templating engine or original config files.

---

## 3 Two things to be careful about

| Risk                                                                                                                                           | Mitigation                                                                                                                                                                                                   |
| ---------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| **Secrets disclosure** – the injected `set db_password = "SuperSecret"` string now lives—signed and immutable—in your artefact store and logs. | • Inject only references (e.g. `set db_password = secret("db/main")`) and have the interpreter resolve those via a secrets vault.<br>• Or encrypt the value with the receiver’s public key before embedding. |
| **Template tampering before sign** – whoever performs the substitution could slip malicious code into the template.                            | • Run the substitution in a *trusted, air-gapped* preparer service.<br>• Sign immediately after generation so later transit can’t alter it.                                                                  |

---

## 4 What your API / tooling should reflect

1. **A “prepare” endpoint**

   ```go
   type Preparer interface {
       Bake(templateID string, args map[string]string) (*SignedAST, error)
   }
   ```

   *Runs on the trusted side, returns a ready-to-ship blob.*

2. **Loader verifies & optionally short-circuits**

   ```go
   if cache.Has(blob.Sum) {
       return cache.Result(blob.Sum) // avoid re-exec
   }
   exec.Run(blob)
   ```

3. **Static pass: `NoSetAfterLoad`**
   Flags any assignment to those baked variables inside the command body; zero runtime mutability.

---

### Bottom line

By merging “config” and “code” into **one immutable, signed artefact**, you give downstream systems everything they need to *see*, *validate*, and *reproduce* exactly what will run—no hidden knobs. Provided you treat secrets carefully and run the baking step in a trusted compartment, it’s a very solid move.

ajp: We would need a standardized way for command blocks to get secrets from the environment I suppose or some kind of one-way encoding perhaps (as you suggested with the rsa sigs?) But making a pre-defined standard would be a good enabling step i think :)

o3:


### A *portable* secret-reference standard for NeuroScript command blocks

*(so we never hard-code secrets, yet still ship a single signed blob)*

---

## 1 Syntax: **`secret("scope/key" [, opts…])`**

```ns
set db_user     = "app"
set db_password = secret("prod/db/main", enc="age")
```

* **Always a compile-time literal** (not a call):
  the canonicaliser records it as a *SecretRef* node, not as a string value.
* **`scope/key`**: flat namespace you map to your vault; keep it boring (`prod/db/main`).
* **`opts`** (optional)

  | key    | allowed values                                 | meaning                                         |
  | ------ | ---------------------------------------------- | ----------------------------------------------- |
  | `enc`  | `"none"` (default) \| `"age"` \| `"sealedbox"` | How the preparer packages the payload           |
  | `vers` | `"latest"` \| explicit id                      | Pin to a version so the blob stays reproducible |

---

## 2 Roles & flow

```mermaid
sequenceDiagram
    Participant Dev
    Participant Preparer    as Trusted&nbsp;Preparer
    Participant Vault       as Secrets&nbsp;Vault
    Participant Gateway     as FDM&nbsp;Gateway
    Participant Runner      as Interpreter/Runner

    Dev->>Preparer: template.ns + cfg.json
    Preparer->>Vault: get("prod/db/main")
    Vault-->>Preparer: "S3cr3t!"
    Preparer->>Preparer: package → SecretBlob
    Preparer->>Gateway: SignedAST (secret ref → encrypted payload)
    Gateway-->>Runner: deliver blob
    Runner->>Runner: decrypt & inject\n(set db_password)
    Runner->>Runner: execute command{}
```

---

## 3 Packaging formats

| `enc` opt   | What the preparer stores in the AST                                        | Runner’s job                 | When to use                                         |
| ----------- | -------------------------------------------------------------------------- | ---------------------------- | --------------------------------------------------- |
| `none`      | Base64 plaintext (still inside signed blob)                                | Just decode                  | Non-sensitive config (feature flags)                |
| `age`       | [age](https://age-encryption.org) encrypted blob using Runner’s public key | Decrypt with its private key | Secrets at *host* security level, simple key distro |
| `sealedbox` | libsodium sealed box with Runner’s Curve25519 pk                           | Decrypt                      | You already use libsodium elsewhere                 |

*All flavours hash the **ciphertext**, so the signed root stays stable.*

---

## 4 Interpreter contract

1. During **load**, walk the tree:

   * If node = `SecretRef{enc:"none"}`, replace with literal.
   * If `enc:"age"| "sealedbox"`, call `secret.decode(blob)`.
2. If decoding fails → raise `ErrorCodeSecretDecryption` (new code 99909?).
3. Keep no plaintext after command completes unless caller flagged `cache_secret=true`.

---

## 5 Static-analysis passes to add now

| Pass                         | Reject when…                                                                         |       |           |
| ---------------------------- | ------------------------------------------------------------------------------------ | ----- | --------- |
| **`secret-outside-command`** | SecretRef found in *library* file. (We don’t want long-lived funcs to hide secrets.) |       |           |
| **`secret-without-enc`**     | `enc="none"` but key path matches regex \`(?i)(password                              | token | secret)\` |
| **`secret-unknown-opts`**    | Any opt key ≠ enc/vers                                                               |       |           |

30-40 LOC each; no crypto needed.

---

## 6 Go helper snippets

```go
// api/secret/package.go
type SecretRef struct {
    Path string            // "prod/db/main"
    Enc  string            // "none" / "age" / "sealedbox"
    Raw  []byte            // ciphertext or plain
}

func Decode(ref SecretRef, privKey []byte) (string, error) {
    switch ref.Enc {
    case "none":
        return string(ref.Raw), nil
    case "age":
        return age.Decrypt(ref.Raw, privKey)
    case "sealedbox":
        return box.OpenAnonymous(nil, ref.Raw, &privKey32, &key32)
    default:
        return "", fmt.Errorf("unknown enc %q", ref.Enc)
    }
}
```

`Canonicaliser` sees `secret("…")` → emits `*ast.SecretRef` ∴ still deterministic.

---

## 7 Operational policy stub

```yaml
# fdm/policy/secret.yml
allowed_encodings:
  - age
  - sealedbox
max_secret_size: 8KiB
forbid_scopes:
  - "dev/*"
```

Loader consults this before accepting a blob.

---

## 8 Migration path

| Stage         | What to implement                                            |
| ------------- | ------------------------------------------------------------ |
| **0 (today)** | Accept `enc="none"` only; secrets stay plaintext but signed. |
| **1**         | Add `age` packaging in Preparer + decode in Runner.          |
| **2**         | Wire vault pull + policy checks, add analyser passes.        |
| **3**         | Flip policy to *forbid* `enc="none"` for sensitive keys.     |

---

### Bottom line

The **`secret("scope/key", enc=…)` literal** gives you

* signed, deterministic blobs,
* zero runtime CLI/env leakage,
* and a clear hook for stronger crypto later.

Start with plaintext to validate the flow, then bolt on `age` without touching any command-block source.
