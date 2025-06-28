### What exactly *is* an “AI agent” inside an FDM?

Below is a clearer, non-overlapping decomposition you can drop straight into design docs or metadata schemas.

| Layer                            | Core Idea                                          | Typical FDM Artefacts                                                                                                                                                                                                                | Notes                                                                                               |
| -------------------------------- | -------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | --------------------------------------------------------------------------------------------------- |
| **1. Cognitive Core (“Mind”)**   | The *model* that does the thinking.                | • `model_id` (OpenAI gpt-4o-mini, local LLAMA-8b, …)<br>• `vendor_account_id` (if SaaS)<br>• `runtime` (spec: CPU/GPU reqs, container hash)                                                                                          | One agent = one primary model *series*; you can hot-swap newer checkpoints under the same agent ID. |
| **2. Disposition / Personality** | How the agent *sounds* and *behaves* emotionally.  | • `tone_profile.md` (cheerful? stoic?)<br>• `temperature`, `top_p`, etc.<br>• Approved style guides / taboo lists                                                                                                                    | Lives alongside the model; small but high-impact knobs.                                             |
| **3. Authority Envelope**        | What the agent *may* do.                           | • `credentials.json` (API keys, repo tokens)<br>• `capability_manifest.nd`<br>   – allowed tools<br>   – filesystem scope<br>   – budget / token limits                                                                              | Enforced by Zadeh’s security layer & Pottery-Barn accountability.                                   |
| **4. Instruction Stack**         | *Why* the agent exists and *how* it should think.  | **(a) Mission charter** – high-level goal (“Defend FDM <Interesting Times Gang> discussions”).<br>**(b) Role card(s)** – concrete duties (“Label off-topic posts”).<br>**(c) Operation SOPs** – fine-grained recipes in NeuroScript. | Stored in `agents/<id>/instructions/`, versioned like code.                                         |
| **5. Private Memory**            | Long-lived data the agent *owns*.                  | • `embedding_index.bin` (vectors of its favourite docs)<br>• Personal notebooks / scratchpads<br>• Reputation & feedback ledger                                                                                                      | Indexed but access-controlled; other agents must request via A2A.                                   |
| **6. Shared Memory Hooks**       | What the agent contributes back to the collective. | • Fractal-Detail overlay links<br>• Event logs (“tagged 42 messages”, “proposed plan v3”)                                                                                                                                            | First-class FDM nodes → searchable by everyone.                                                     |
| **7. Telemetry & Provenance**    | Evidence of what it actually did.                  | • Execution traces, cost reports<br>• Error + retry logs                                                                                                                                                                             | Feeds the reputation engine; allows audits or rollbacks.                                            |

---

#### How the layers interact in practice 🚦

```
┌──────────────────────────────────────────────────────────────────────┐
│      AGENT  “Curious-Curator”                                       │
├─────────────────────────────────────────────┬────────────────────────┤
│ 1  Cognitive Core      (gpt-4o-mini)       │ 6  Shared Memory Hooks │
│ 2  Disposition          (curious, polite)  │      ⇡ overlay links   │
│ 3  Authority  ▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒ │                        │
│       FS=/content/wiki   Tools=[Embed,Tag] │ 5  Private Memory      │
│       TokenBudget=2k/hr                     │      vector index      │
│ 4  Instruction Stack  (mission → SOPs)     │                        │
│                                            │ 7  Telemetry stream    │
└─────────────────────────────────────────────┴────────────────────────┘
```

**Startup sequence**

1. **Zadeh** loads layers 1-4 from versioned storage, validates layer 3 (credentials, quota).
2. Agent spins up with its *private memory* mounted read-write.
3. Tasks arrive via CatHerder or A2A calls.
4. Outputs flow to both private (drafts, embeddings) and shared memories (final tags, overlay edges).
5. Telemetry is appended asynchronously; reputation scores update.

---

#### Minimal metadata stub (`agent.yaml`)

```yaml
agent_id: curious-curator
model:
  series: gpt-4o-mini
  vendor: openai
personality:
  tone_profile: curious
  temperature: 0.6
authority:
  token_budget: 2000   # per hour
  tools: [Embed, Tag]
  fs_scope: /content/wiki
instructions:
  mission: docs/charters/content_curation.md
  roles:
    - docs/roles/wiki_tagger.md
    - docs/roles/off_topic_sentry.md
memory:
  private_path: agents/curious-curator/mem
```

---

**Key design cues**

* *All* layers are plain files inside the FDM, versioned & reviewable.
* Layers 3-4 define the sandbox; breach = fast-fail, logged to reputation.
* Personal memory (layer 5) stays searchable but only via authorised queries — no silent data hoarding.
* Swap-in a new LLM?  Update layer 1, keep layers 2-7 intact → zero-downtime personality continuity.
