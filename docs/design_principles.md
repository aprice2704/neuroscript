# NeuroScript Project Design Principles  
*Criteria against which every design choice must be weighed*  
  
---  
  
## Root Principles  
  
### 1. Obvious  
- **Everything in NS must be obvious or self-documenting.**  
  
*Why?* High-Agency Contributors (HACs) should behave correctly with minimal onboarding.  
  
### 2. Bounded  
- **Every construct must declare positive resource bounds.**  
  
*Why?* Real systems live under finite budgets; clear limits enable accountability.  
  
### 3. Defined  
*“NS must be obvious **and** picky.”*  
- **Behaviour must be explicitly specified and testable.**  
  
*Why?* Predictability is the bedrock of reliability and security.  
  
### 4. Predictable  
- **Same inputs → same outputs → same side-effects.**  
  
*Why?* If you cannot foresee consequences you cannot behave well.  
  
### 5. Accountable  
*“The bedrock of security is knowing **who** did **what**, **when**.”*  
- **Every change is attributable to the immediate HAC and traceable up the causal chain.**  
  
### 6. Safe-by-Default  
- **The default posture denies risky or irreversible actions; opt-ins must be explicit and auditable.**  
  
*Why?* A cautious baseline limits blast-radius and makes accidental harm unlikely.  
  
## Derived Principles  
  
### D1. Simple  
*From Obvious + Defined + Bounded + Predictable*  
- **As simple as possible ... but never simpler.**  
  
### D2. Orthogonal  
*From Predictable + Simple*  
- **Each feature does one thing, with no hidden side-effects.**  
- **Combinations compose without new rules.**  
  
### D3. Versioned  
- **Every artefact and interface advertises an explicit version.**  
  
*Why?* Enables evolution without surprise breakage, guarantees reproducibility, and underpins roll-back safety.  

### D4. Auditable  
*From Accountable + Bounded + Defined*  
- **All mutations leave a tamper-evident trail (hash + timestamp + agent ID).**  
- **Inspecting costs ≤ performing.**  
  
### D5. Deterministic-First  
*From Predictable*  
- **Deterministic unless an explicit nondeterministic tool is invoked—and that tool must tag its output.**  
  
### D6. Capability-Scoped  
*From Defined + Bounded + Safe-by-Default*  
- **Least privilege by default; escalation is explicit and audited.**  
- **Imports never silently widen capabilities.**  
  
### D7. Fuel-Metered Everywhere  
*From Accountable + Bounded*  
- **Every phase—lint, parse, static check, VM step—burns fuel at a public tariff; overruns abort before neighbours starve.**  
  
### D8. Progressive Disclosure  
*From Bounded*  
- **Minimal core syntax; advanced features gated by clear opt-ins** (`version`, `ttl`, `command endcommand` blocks).  
  
### D9. Reversible  
*From Bounded + Versioned*  
- **Every state change has a symmetric undo (or documented rationale).**  
- **Versioned overlays + alias tables make rollback cheap.**  
  
### D10. Minimum Surprise  
*From Obvious + Predictable*  
- **Look-alike constructs act alike; hidden magic (implicit imports, silent promotions) is forbidden.**  
  
### D11. Isolated  
*From Safe-by-Default + Bounded + Predictable*  
- **Each execution runs in a sandbox with enforced fuel and capability limits; faults stay contained.**  
  
### D12. Evolvable  
*From Versioned + Reversible*  
- **Schema and API migrations occur via additive versions and alias promotion—never in-place mutation.**  
  
---  
  
:: name: NeuroScript Design Principles  
:: schema: spec  
:: serialization: md  
:: fileVersion: 3  
:: author: Andrew Price & ChatGPT-o3
:: created: 2025-06-29  
:: modified: 2025-06-29  
:: description: Canonical design-decision rubric for the NeuroScript language and runtime.  
:: tags: designPrinciple, neuroscript, fdm  
:: dependsOn: none  
:: howToUpdate: Propose edits via command block; bump fileVersion on acceptance.  
:: glossaryNote: *actor* means any entity that can initiate actions—human, AI model, or autonomous computer process.
