### Embedding NeuroScript-style Metadata + SDI in Go
*(compact, compiler-safe, godoc-friendly)*

---

#### 1 Syntax snapshot

`go  
// Package foo provides …  
//  
// :: product: FDM/NS
// :: majorVersion: 1
// :: fileVersion: 1
// :: description: Foo-bar bridge to NeuroScript tools  
// :: sdiSpec:   bridge            // ← ID in docs/spec/bridge.md  
// :: contract:   value_wrapping  
// :: schema: sdi-go
// :: serialization: go
//  
// sdi:design single chokepoint unwrap-coerce-wrap  
package foo  
  
// sdi:impl bridge  
func (r *Registry) CallFromInterpreter(/* … */) { … }  
`

---

#### 2 Directive grammar

| Form | Scope | Example |
|------|-------|---------|
| `// :: key: value` | **File-level** metadata | `// :: tags: bridge, llm` |
| `// sdi:design <text>` | **Immediately before** pkg/function | `// sdi:design pure functional tokenizer` |
| `// sdi:impl <specID>` | On every exported type/func | `// sdi:impl bridge` |

*(all are plain `//` comments ⇒ ignored by compiler, shown by godoc)*

---

#### 3 Placement guide

\* **Top-of-file doc comment** → all `::` keys plus optional overall `sdi:design`.
\* **Above exported APIs** → `sdi:design` + mandatory `sdi:impl <specID>`.
\* **Private helpers** → tag only if they realise a distinct spec.

---

#### 4 Contracts

`go  
// :: contract: value_wrapping  
// :: contract_ref: docs/spec/value_wrapping.md  
`

CI can verify the markdown file exists; IDE or LLM can jump to it.

---

#### 5 Tooling sketch (≤50 LOC)

\* **grep / parser**: `grep -nE '^// (::|sdi:)'` for quick scans.
\* **vet rule**: every `sdi:impl X` must match some `:: sdiSpec: X`.
\* **fail CI** if a `contract:` path is missing.

---

#### 6 Minimal working example

`go  
// Package memorystore persists fractal detail memories.  
//  
// :: product: FDM/NS
// :: majorVersion: 1
// :: fileVersion: 3
// :: langVersion: neuroscript@0.4.1 
// :: description: Core snapshot store with time-travel telemetry
// :: author: Andrew Price  
// :: copyright: Copyright (c) 2025 Andrew Price
// :: license: MIT  
// :: sdiSpec:   memoryStore  
// :: contract:   valueWrapping  
// :: schema: sdi-go
// :: serialization: go
//
`