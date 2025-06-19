Below is a clean Markdown version: the expanded **Actors Matrix** table first, followed by one concise rationale-and-practice paragraph for each SDI phase.

---

### SDI Actors & Responsibilities Matrix

| Stage         | Guiding question                     | Actors (H = Human / A = AI agent / C = Code-system)                                                                                                                                                                | Expected artifacts                                                            | Failure mode it guards against                      |
| ------------- | ------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ----------------------------------------------------------------------------- | --------------------------------------------------- |
| **Specify**   | *What problem are we solving & why?* | **Customer:** Exec sponsor (H), Product owner (H), End-user reps (H)<br>**Vendor:** Engagement PM (H), Business analyst (H/A), Workshop chatbot (A)<br>**Shared:** Legal counsel (H), Requirements validator (C/A) | Problem statement, KPIs, constraints, contract outline                        | Building the wrong thing, unowned scope creep       |
| **Design**    | *How could we solve it?*             | **Customer:** Architecture board (H), UX lead (H)<br>**Vendor:** Solution architect (H), System designer (H/A), Cost-estimator bot (A)<br>**Partners:** Security advisor (H), Third-party API owners (H)           | Architecture diagram, interface & test contracts, prototype, estimate dossier | Over-engineering, hidden assumptions, budget shock  |
| **Implement** | *Make it real & verifiable.*         | **Vendor:** Dev/ML engineers (H/A), SRE & pipelines (C), QA harness (C/A)<br>**Customer:** UAT users (H), Ops hand-over team (H)<br>**Shared:** CI/CD botnet (C), Telemetry stack (C)                              | Running system, tests, deployment scripts, training docs                      | Ad-hoc hacks, fragile release, unusable deliverable |

---

### Phase Spotlights (one-paragraph each)

**Specify** – Lock the *why* and *what* before touching the *how*. Run a facilitated workshop (human + AI scribe) to capture pain points, success metrics, and hard constraints; publish a one-page spec that the executive sponsor must sign. Freeze vocabulary with a mini-glossary to prevent later semantic drift. A solid Specify phase keeps domain experts from “helpfully” sketching solutions that derail design later.

**Design** – Convert the approved spec into a viable architecture while trade-offs are still cheap. Architects (human) and design-bots (AI) iterate on models, spike the riskiest integration, and generate machine-readable interface contracts that future tests will enforce. The phase ends with a **design freeze**: any scope change after this triggers a formal change order, protecting timeline and budget.

**Implement** – Build exactly what was frozen in Design, nothing hidden up sleeves. Pair programmers (human+AI) crank code that CI/CD pipelines continuously test against the contractual specs; chaos and load testing run from day one to harden ops. A telemetry dashboard proves KPI attainment, and a clean hand-over package lets customer ops take the wheel without calling the vendor at 3 a.m.

### SDI — Expanded Actors Matrix (v0.2)

| Stage         | Guiding Question                     | **Actors (H = Human, A = AI/Agent, C = Code/System)**                                                                                                                                                                                                                                                                       | Expected Artifacts                                                                | Failure Mode it Guards Against                      |
| ------------- | ------------------------------------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------- | --------------------------------------------------- |
| **Specify**   | *What problem are we solving & why?* | **Customer-side**<br>• Executive sponsor (H)<br>• Product owner / SME (H)<br>• End-user reps (H)<br><br>**Vendor-side**<br>• Engagement PM (H)<br>• Business analyst (H/A)<br>• Facilitation chatbot for workshops (A)<br><br>**Shared/Neutral**<br>• Legal/governance counsel (H)<br>• Requirements parser/validator (C/A) | Problem statement, success KPIs, constraints, contract outline                    | Building the wrong thing, unowned scope creep       |
| **Design**    | *How could we solve it?*             | **Customer-side**<br>• Architecture review board (H)<br>• UX lead / prototyping users (H)<br><br>**Vendor-side**<br>• Solution architect (H)<br>• System designer / modeler (H/A)<br>• Cost estimator bot (A)<br><br>**Integration Partners**<br>• Security/compliance advisor (H)<br>• Third-party API owners (H)          | System architecture, interface & test contracts, prototype/demo, estimate dossier | Over-engineering, hidden assumptions, budget shock  |
| **Implement** | *Make it real & verifiable.*         | **Vendor-side**<br>• Dev/ML engineer swarm (H/A)<br>• SRE / infra-as-code pipelines (C)<br>• QA automation harness (C/A)<br><br>**Customer-side**<br>• UAT users (H)<br>• Ops hand-over team (H)<br><br>**Shared/Neutral**<br>• CI/CD botnet (C)<br>• Observability/telemetry stack (C)                                     | Running system, tests, deployment scripts, training material                      | Ad-hoc hacks, fragile release, unusable deliverable |

---

#### Notes & Rationale

1. **Explicit Customer vs Vendor Roles**
   • Keeps decision rights clear. If a customer insists on diving into Design or even Implementation, you can point to this matrix and ask which hat they’re wearing.

2. **H-A-C Triplet**
   • **H**umans for judgement, negotiation, and domain nuance.
   • **A**gents/LLMs for drafting, summarising, and synthetic tests—cheap iteration.
   • **C**ode/Systems for repeatability and orchestration.

3. **PMBOK Alignment without the Bloat**

   | PMBOK Knowledge Area                 | Where it lives in SDI                             |
   | ------------------------------------ | ------------------------------------------------- |
   | Scope & Stakeholder                  | **Specify** (problem statement, sponsor sign-off) |
   | Risk, Cost, Procurement              | **Design** (trade-off log, estimate dossier)      |
   | Quality, Integration, Communications | **Implement** (CI/CD, telemetry, UAT hand-over)   |

4. **Churn Counter-Measure**
   • Define *“RACI-within-SDI”* early: e.g. Customer SME = *Consulted* at Design, *Informed* during day-to-day Implementation.
   • Lock *Design-freeze* milestones tied to payment schedules—protects both sides.

---

**Next up?**

* We can draft a lightweight RACI template or produce sample onboarding slides that visualise these roles. Let me know which deliverable moves the needle for you.

### SDI (Specify → Design → Implement) — Working Summary v0.1

| Stage         | Guiding Question                                           | Typical Actors                                  | Expected Artifacts                                                             | Common Failure Mode it Guards Against          |
| ------------- | ---------------------------------------------------------- | ----------------------------------------------- | ------------------------------------------------------------------------------ | ---------------------------------------------- |
| **Specify**   | *What problem are we solving, and **why** does it matter?* | Domain experts, product owners, policy thinkers | One-page problem statement, success metrics, constraints, “definition of done” | Building the wrong thing, goal drift           |
| **Design**    | *Given the spec, **how** could we solve it?*               | Architects, senior engineers, researchers       | High-level architecture, trade-off log, test strategy, interface contracts     | Over-engineering, hidden assumptions           |
| **Implement** | *Make it real and verifiable.*                             | Coders, fabricators, AIs, QA                    | Running system, tests, build scripts, deployment notes                         | Ad-hoc hacks, brittleness, unmaintainable code |

---

#### 1. Why SDI?

1. **Up-front clarity without heavyweight process.**
   Big-frameworks (SSADM, SAFe, “Agile-in-name-only”) add ceremony; hack-and-hope adds chaos. SDI keeps the parts that matter (clear specs, deliberate design) and jettisons the church bells.

2. **Deliberate skill hand-offs.**
   • Specification needs domain context and stakeholder language.
   • Design rewards systems thinking.
   • Implementation thrives on focused execution (human or AI).
   By naming the hand-offs, SDI reduces cross-discipline fog.

3. **Fails fast on paper, not in production.**
   Forcing an explicit spec and design surfaces impossibilities early—before sunk cost sets in.

---

#### 2. Core Principles

| Principle                    | Practical Meaning                                                                                                      |
| ---------------------------- | ---------------------------------------------------------------------------------------------------------------------- |
| **Lean, not lax**            | One-pager beats a 50-page doc, but *some* doc beats none.                                                              |
| **Iterative passes allowed** | You can loop back, but must be explicit about it (“Spec v2 after user tests”).                                         |
| **Contracts > Hope**         | Formalise interface & behaviour contracts in *Design* so *Implement* can be automated or delegated, including to LLMs. |
| **Traceability**             | Every implementation commit links back to a design line-item and spec requirement—your future debugger will thank you. |

---

#### 3. Where It Shines

* **AI-assisted coding** — Clear spec-&-design prompts dramatically improve LLM output quality.
* **Mixed hardware/software projects** — Natural boundary between electrical, mechanical, and firmware tasks.
* **Policy drafting** — Spec = goals & guard-rails; Design = policy levers & scenarios; Implement = legislation wording & roll-out.

---

#### 4. How It Plays with Existing Methods

| If you already use… | SDI acts as…                                                                               |
| ------------------- | ------------------------------------------------------------------------------------------ |
| **Scrum/Agile**     | A pre-sprint grooming discipline (Spec & Design) and a definition of “ready”.              |
| **Waterfall**       | A lighter-weight, cyclic micro-waterfall inside larger phases.                             |
| **DevOps**          | The “Spec → Design” artifacts feed CI/CD pipelines as machine-readable requirements/tests. |

---

#### 5. Payoff Metrics

* Fewer spec-change defects after sprint start.
* Shorter onboarding time (new devs read spec/design first).
* Higher automated-test coverage (tests drafted in Design stage).
* Reduced “re-write from scratch” events.

---

#### 6. What SDI Is *Not*

* A rigid stage-gate: You can—and should—bounce back when new info arrives.
* A replacement for user feedback loops: Prototype early, feed results into a *Spec vN*.
* A silver bullet: It still needs disciplined humans (and AIs) to enforce the hand-offs.

---

#### 7. Next Steps for This Doc

1. **Examples** — walk through a real feature or hardware part using SDI tags.
2. **Metadata format** — show how SDI headings embed in Go/NeuroScript comments.
3. **Tooling hooks** — outline how CI can reject PRs missing SDI links.

*(Let me know which of these to flesh out first, or what gaps you see.)*
