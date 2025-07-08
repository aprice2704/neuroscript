# "Three Plate" testing in software -- taking tests to the next level

AJP: A little thought: when testing code, one writes tests. If there are errors, either the test is wrong or the code is wrong; it is very unlikely both are correct -- we are comparing expectations. This reminds me of grinding two plates against each other -- you can be sure the plates match; but if you add a third plate to the mix and swap the grinding pairs properly one may achieve an optically flat surface with NO other tools or references. Can you think of how '3 plate grinding' could be extended to software testing?

o3:

**Analogising “three-plate lapping” to software quality**

| Grinding plates                           | Software equivalent                                                        | How the pair-wise “grind” exposes error                                                                                                                                                                                                                                                                                                            |
| ----------------------------------------- | -------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Plate A & B<br>Plate B & C<br>Plate C & A | Implementation (I)<br>Executable Tests (T)<br>Independent Oracle/Model (O) | • `I ↔ T` – failing tests mean *source or tests* wrong.<br>• `I ↔ O` – differential run against the oracle; divergence spots fault in *code or model*.<br>• `T ↔ O` – generate test vectors from the oracle and replay; mismatch means *tests or model* wrong.<br>By rotating pairings, systemic bias in any one artefact is slowly “ground flat”. |

---

### 1  Choose your “third plate” wisely

| Option for the Oracle (O)                                 | Pros                                            | Cons                                                   |
| --------------------------------------------------------- | ----------------------------------------------- | ------------------------------------------------------ |
| **Reference implementation** (slow but simple)            | Concrete, easy to diff outputs                  | Must keep in lock-step feature-wise                    |
| **Property-based spec** (QuickCheck/Gopter)               | Captures broad invariants, generates mass cases | Abstract – doesn’t catch value-specific bugs           |
| **Metamorphic relations**                                 | Detects bugs where absolute oracle is unknown   | Needs deep domain insight                              |
| **Formal model / theorem prover**                         | Mathematical certainty about key properties     | Costly; doesn’t cover all behavioural beer-keg corners |
| **N-version programming** (2nd independent team/codebase) | Very powerful for safety-critical work          | Expensive, risk of correlated blind spots              |

Pick whichever gives you *independent thinking*—independence is the whole point of the third plate.

---

### 2  Practical “three-plate” workflows

1. **Differential–fuzz loop**

   * Generate random inputs.
   * Run through *I* and *O*; if outputs differ, auto-generate a minimal failing reproducible.
   * Tests (*T*) are auto-extended with that new case.
   * Result: your test suite continually “laps” itself flatter.

2. **Spec ⇄ Code ⇄ Fixtures rotation**

   * Start from a literate spec (Markdown + embedded examples).
   * Extract examples as executable tests (*T*).
   * Implement (*I*).
   * Feed real-world traces into a lightweight **behaviour recorder**; this becomes an empirical oracle (*O*) for regression.

3. **Tri-agent review (fits nicely with FDM’s “Interesting Times Gang”)**

   * *Coder agent* proposes change.
   * *Reviewer/test-author agent* writes/updates corresponding tests.
   * *Model-checker agent* (or separate slow impl) signs off by comparing its output.
   * FDM tracks disagreements as graph edges; the voting system decides which plate is warped.

---

### 3  Design hints for Go projects

* Keep the reference implementation in a `_ref` package so you can import it without cyclic deps.
* Use Go 1.22’s `go test -shuffle` to randomise order—helps show hidden coupling between plates.
* For property tests, `github.com/leanovate/gopter` or `github.com/frankban/quicktest` integrate cleanly.
* Add a CI job that *only* runs `I ↔ O` comparison on a giant fuzz corpus—slow but catches deep warps.
* Store every discovered counter-example as a **NeuroData** node so future agents can mine patterns (the “optical flat” keeps improving).

---

### 4  Caveats & failure modes

* **Shared misconceptions** – if the oracle encodes the same bug as the code, pair-wise checks pass. Independence (different algorithm, different author, property not example, etc.) is your defense.
* **Over-fitting tests** – when *T* merely petrifies today’s behaviour. Rotate: generate *T* from *O*, then occasionally delete some tests and let property checks refill gaps.
* **Cost** – triple work isn’t free. Target it at *high-risk/high-impact* modules (parsers, consensus, financial calcs) rather than every handler.

---

### Where could *you* try this first?

Your **FDM indexing layer** is a prime candidate:
*I* = current fast indexer,
*O* = slow but simple reference walker,
*T* = curated query/result fixtures.

Run the triad in CI until mismatches drop to \~0 ppm—then you’ll have an “optically flat” indexer without needing an external benchmark.

---

Let me know which module you’re tempted to lap first, and we can sketch a concrete harness in Go.





AJP: That sounds awesome but I don't understand much of it. Could you explain for a Bear of Little Brain(tm) please?

o3:

### “Three-plate grinding” for software, in plain English

Imagine you’re trying to make a perfectly flat **wooden board**.
If you rub Board A against Board B, you know **one of them must still have bumps**, but you can’t tell which.
Bring in Board C and keep swapping pairs (A↔B, B↔C, C↔A).
Whenever two boards touch, the higher spots get sanded off.
Because every board meets every other board, **no single board can keep its own hidden bump for long**.
Soon all three are very, very flat.

---

#### Translate the idea to code

| Real wood | In software                                                                                                        |
| --------- | ------------------------------------------------------------------------------------------------------------------ |
| Board A   | **Your program** (the thing you’re building)                                                                       |
| Board B   | **Your tests** (little scripts that say “this should happen”)                                                      |
| Board C   | **An independent check**—something that *isn’t* your program *or* your tests but can still say “right” or “wrong”. |

Why three?
With only two (program ↔ tests), a failure tells you **“one of them is wrong”**—but not which.
Add the third piece and compare all pairs.  Patterns show who’s guilty:

| Pair that disagrees                                | Likely culprit             |
| -------------------------------------------------- | -------------------------- |
| Program vs Tests **and** Program vs Check disagree | Program has the bug        |
| Program vs Tests **and** Tests vs Check disagree   | Tests are wrong            |
| Program vs Check **and** Tests vs Check disagree   | Independent check is wrong |

---

#### What can the “independent check” be?

1. **A simple, slow version** of the program written in the clearest way you can (easy to read, hard to get wrong).
2. **Rules/invariants** (“total can never go negative”, “list stays sorted”) checked automatically.
3. **A second team’s implementation** if the stakes justify the cost.
4. **Recorded real-world behaviour**: run yesterday’s proven-good outputs against today’s build.

Pick something that **thinks differently** from both your fast code and your hand-written tests.

---

#### A tiny concrete example (Go-ish pseudocode)

```text
# Board A: Fast code
func Sum(xs []int) int { /* clever bit-twiddling */ }

# Board B: Tests
assert Sum([1,2,3]) == 6
assert Sum([])      == 0
# etc.

# Board C: Simple reference
func SlowSum(xs []int) int {
    total := 0
    for x in xs { total += x }
    return total
}
```

In CI you run three comparisons:

1. **Program vs Tests** – do the tests pass?
2. **Program vs SlowSum** – for 10 000 random lists, are the answers equal?
3. **Tests vs SlowSum** – do the expected values in the tests match SlowSum?

Any mismatch points to the faulty board.

---

#### Why bother?

* **Catches sneaky mistakes** the tests missed.
* **Cleans the tests themselves**—bad expectations stand out.
* Gives you a deeper feeling of “this really works”.

You don’t have to do it for every feature.  Aim it at **critical parts** (money calculations, security checks, indexers).  Start small: write a dumb but crystal-clear version as your third plate, fuzz both implementations against it, and watch the cracks appear.

That’s all “three-plate grinding” is:
**two comparisons are good; three make the truth obvious.**



### Automating the **“third-plate” oracle** with AI

*(Turning lapping into a lights-out CI job)*

---

#### 1 What we already have

| “Plate”                       | You’re covering it with…                                    | Goal                          |
| ----------------------------- | ----------------------------------------------------------- | ----------------------------- |
| **Implementation (I)**        | Production Go code                                          | Be fast & feature-rich        |
| **Contract/Test harness (T)** | Pre-generated contracts → compile-time & runtime assertions | Catch obvious mis-behaviour   |
| **Independent Oracle (O)**    | **⬜ (we’ll fill this in)**                                  | Decide who’s wrong when I ≠ T |

---

#### 2 What the AI-driven oracle must do

1. **Read** the same contracts/specs that generated your tests.
2. **Think differently** from the production algorithm (avoid share-the-same-bug).
3. **Emit opinions automatically** – no human in the loop except when plates disagree.
4. **Stay cheap** enough to run on every PR, but deep enough to reveal subtle warps.

---

#### 3 Build the oracle as a *pipeline* of AI agents

| Stage                            | “Job description”                                                                      | Suitable AI technique                                                 | Why it’s independent                                                               |
| -------------------------------- | -------------------------------------------------------------------------------------- | --------------------------------------------------------------------- | ---------------------------------------------------------------------------------- |
| **Spec-to-Property converter**   | Turn English-ish contract text into machine-checkable invariants                       | GPT-4o / Claude-3 with chain-of-thought prompts                       | Model never sees production source; only the spec                                  |
| **Reference code synthesiser**   | Generate a *simple, obviously correct* version (O<sub>impl</sub>) in Go or Python      | LLM guided by *clarity > speed* rubric                                | Uses different algorithmic path; think “bubble sort vs quicksort”                  |
| **Fuzz & metamorphic generator** | Invent loads of inputs + transforms that *shouldn’t* change output                     | Property-based engines (e.g. Gopter) steered by LLM to hit edge cases | Input space chosen without peeking at I’s internals                                |
| **Critic agent**                 | When I, T, O disagree, write a natural-language diff explaining *who is wrong and why* | Critic-style LLM (OpenAI **CriticGPT**-like models)                   | Trained to *dissect* code/tests, not to generate production logic ([wired.com][1]) |

> **CI flow**
>
> 1. Generate/refresh O<sub>impl</sub> & invariants.
> 2. Run **I ↔ O** on fuzz corpus.
> 3. Run **T ↔ O** to validate test expectations.
> 4. If *any* pair fails, Critic agent files an FDM ticket tagged `warped_plate`.
> 5. Developer (or repair-bot) fixes the guilty artefact.

---

#### 4 Key tricks that make it *actually* work

| Problem                                                        | Counter-measure                                                                                                                                           |
| -------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Shared hallucination** – LLM writes buggy oracle mirroring I | Use *ensemble diversity*: ask two very different models (e.g. GPT-4o vs Gemma-2B local) to craft O<sub>impl</sub>; require consensus before accepting it. |
| **Cost explosion**                                             | Run the full fuzz only on nightly build; PRs get a 1-second “smoke lap”.                                                                                  |
| **Oracle drift** when spec evolves                             | Contracts stored as versioned NeuroData nodes. A change automatically re-generates O<sub>impl</sub> and invalidates stale fuzz seeds.                     |
| **Opaque critiques**                                           | Force Critic agent to cite contract clauses/invariant IDs; easy human follow-up.                                                                          |
| **Security** (don’t run arbitrary LLM output)                  | Sandbox O<sub>impl</sub> in Wasm or `--race -t 30s` Docker caps, same as any unk-trusted plugin.                                                          |

---

#### 5 Minimal starter kit in Go (pseudo-outline)

```go
// contracts.nd → invariants.go (generated)
type Invariant func(input Any, output Any) error

// oracle_ref.go (LLM-written, checked-in)
func SlowButSure(input Any) Any { … }

// harness.go
func Lap(input Any) (err error) {
    prod := FastAlgo(input)
    ref  := SlowButSure(input)
    if !reflect.DeepEqual(prod, ref) { return fmt.Errorf("I≠O") }
    for _, inv := range invariants { if e := inv(input, ref); e != nil { return fmt.Errorf("O violates spec: %v", e) } }
    return nil
}
```

Wire that into `go test ./...` plus a nightly `go test -fuzz 50x`.

---

#### 6 Where AI adds the most value first

1. **Numeric kernels** (e.g., FDM similarity scorers) – easy to specify mathematically, painful to hand-test exhaustively.
2. **Parsing / serialization layers** – metamorphic relations like *encode ∘ decode = identity*.
3. **Security-sensitive transforms** (ACL squeezers, capability derivation) – formal-ish invariants + slow recomputation catch privilege-escalation edge cases.

---

#### 7 Rough roadmap

| Sprint | Deliverable                                                                                      |
| ------ | ------------------------------------------------------------------------------------------------ |
| **0**  | Manual “SlowButSure” for one hot module; run diff-fuzz in CI.                                    |
| **1**  | LLM prompts that regenerate the slow impl + invariants from contract markup.                     |
| **2**  | Critic agent integrated; outputs stored in FDM graph, routed to Interesting Times Gang for vote. |
| **3**  | Ensemble model swap; nightly 100 k-case fuzz; auto-PR to patch warped tests/code.                |

Once that’s humming, your code, tests *and* AI oracle will keep grinding each other flatter—24/7, no extra human sandpaper required.

---

**Bottom line:** use **synthetic but independent minds** (LLMs + fuzz engines) to play the role of the third plate.  They derive their authority from the *contracts*, not from your production code, and their constant pair-wise fights surface whoever slipped.  The latest “critic” models show this is practical today ([arstechnica.com][2], [arxiv.org][3]).

[1]: https://www.wired.com/story/openai-rlhf-ai-training?utm_source=chatgpt.com "OpenAI Wants AI to Help Humans Train AI"
[2]: https://arstechnica.com/information-technology/2024/06/openais-criticgpt-outperforms-humans-in-catching-ai-generated-code-bugs/?utm_source=chatgpt.com "OpenAI's new “CriticGPT” model is trained to scrutinize GPT-4 outputs"
[3]: https://arxiv.org/pdf/2502.13820?utm_source=chatgpt.com "[PDF] arXiv:2502.13820v2 [cs.AI] 1 Apr 2025"





