# Fact List

## A new simple data type

### “FactList” — a confidence-weighted checklist for NeuroData

#### 1. What problem does it solve?

| Current pain                                                                                          | FactList benefit                                                                                                                    |
| ----------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------- |
| **Checklists are binary.** Either a box is ticked or it isn’t.                                        | Adds graded certainty (`0 … 1`) so downstream agents can weigh evidence instead of treating everything as true/false.               |
| **Free-form metadata gets buried.** Facts written in comments or JSON blobs aren’t machine-mergeable. | Uniform container with first-class ops: merge, filter, rank, aggregate.                                                             |
| **Worth-of-work heuristics are ad hoc.**                                                              | A FactList attached to a work product acts as a reproducible quality filter (e.g. “≥90 % cumulative confidence across core-facts”). |

---

#### 2. Core data model

```go
type Fact struct {
    Text       string    // single declarative sentence
    Confidence float64   // 0.0–1.0
    Source     SourceRef // optional: URL, file, tool run
    Stamp      int64     // unix ns for decay / provenance
}

type FactList []Fact
```

**Invariant:** `Text` must be *sentence-like* (starts with capital, ends with `.`) to encourage concise assertions.

---

#### 3. Built-in operations (MVP)

| API                                                           | Purpose                                                   |
| ------------------------------------------------------------- | --------------------------------------------------------- |
| `Add(f Fact)`                                                 | Append or up-rev identical `Text` with higher confidence. |
| `Merge(other FactList, mode string)`                          | `"max"`, `"avg"`, `"bayes"` fusion of confidences.        |
| `Filter(fn func(Fact) bool)`                                  | Predicate filter; cheap for pipelines.                    |
| `Score(require []string) (covered float64, missing []string)` | Supply must-have fact IDs; returns coverage metric.       |
| `TopN(n int)`                                                 | Grab highest confidence facts for summaries.              |

Extend `typeof` in NeuroScript to return `"factlist"`.

---

#### 4. Integration points

1. **New `core.Value` kind**
   *Registration:*

   ```go
   core.RegisterKind(core.KindFactList, wrapFactList, unwrapFactList)
   ```

2. **Interpreter syntax sugar**

```neuroscript
factlist baseline := [
  "All functions have unit tests." @0.9,
  "Code compiles under Go 1.23."   @1.0,
  "Benchmarks meet SLA."          @0.6
]

baseline.add("Security scan passes." @0.8)
```

`@"confidence"` postfix keeps authoring friction low.

3. **Check-like evaluation**

```neuroscript
set covered, missing = baseline.score(require core_facts)
if covered < 0.9
  fail "Insufficient confidence (" + covered + ")"
endif
```

4. **Neuromorphic overlay hook**
   Reward signal could boost edges between facts that co-occur in successful builds → better suggestion of missing facts.

---

#### 5. Persistence & indexing

* Use **columnar** key-value: `(FactHash → confidence, source, stamp)`; lists store only hashes to minimise duplication.
* Secondary index on `(Text trigram)` for quick dedup/null matching.
* Snapshot every N ops; actor model from previous discussion slots straight in.

---

#### 6. Failure modes & guards

| Risk                                          | Mitigation                                                                                                        |
| --------------------------------------------- | ----------------------------------------------------------------------------------------------------------------- |
| Confidence inflation (“Lake Wobegon” effect). | *Calibration pass* in CI: if manual review disproves a fact, halve confidence of similar facts (Levenshtein ≤ 5). |
| Free-text vagueness.                          | CI linter: must match regex `^[A-Z][^.]+[.]$`; flag weasel words (`may`, `possibly`).                             |
| Fact explosion.                               | Soft cap per artifact (e.g. 100); auto-merge duplicates by semantic hash (`text + stem`).                         |

---

#### 7. Worth-of-work metric exemplar

```neuroscript
func Worth(needs facts FactList, weights map[string]float) returns score means
  set covered, _ = facts.score(require weights.keys())
  set score = 0
  for each f in facts
    if f.text in weights
      set score += weights[f.text] * f.confidence
    endif
  endfor
  return score * covered              # emphasise completeness
endfunc
```

Attach this to CI; gate merges on `score >= 0.85`.

---

### Verdict

*Low lift*—a single new `ValueKind`, a handful of ops.
*High leverage*—turns subjective “looks good” into a quantifiable, mergeable artifact.
Prototype inside `neurodata/facts` with 3–4 critical ops; wire it into your existing checklist flows to validate UX before deeper overlay ties.
