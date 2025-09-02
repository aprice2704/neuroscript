# Graph as Data Type in ns

AJP: in fdm we need to do code analysis. we have built a bunch of overlays, but we just got out of a dep pickle using a package dep graph. I am wondering if we should add a 'graph' (node-edge) data type to ns (of which tree is a subtype natch) and some generic analysis tools? (isdag, hasloops, findrootiest??)


gpt5: Short answer: yes—add a first-class directed graph type, but keep it lean, immutable-by-default, and instrumented for provenance. Most code-analysis artifacts you care about (imports, calls, CFG/DFG, type refs, ownership, overlay relations) are naturally graphs. Treat “tree” as a constrained view over the same type, don’t fork the model.

Here’s a concrete, minimal plan that won’t paint you into a corner.

# 1) Core value type

* **Type**: `graph` (directed, possibly multi-edge, labeled).
* **Identity model**:

  * `NodeID`: stable, 64-bit (or string) IDs; separate **payload** map for attributes.
  * `EdgeID`: optional; edges = `(src, dst, label?, weight?, attrs?)`.
* **Representation** (canonical, persistent; adjacency lists):

  * `nodes: map<NodeID, map<string, any>>`
  * `out: map<NodeID, list<Edge>>`
  * `in:  map<NodeID, list<Edge>>`  (maintained for O(1) indegree/ancestors)
* **Immutability**: structural-sharing updates: `g2 = graph.with_edge(g, e)`; bulk builders for speed.
* **Labels**: strings from a tiny controlled vocabulary (`"imports"`, `"calls"`, `"cfg"`, `"type"`, `"owns"`, `"includes"`, …). Avoid ad-hoc snowflakes.

Trees become graphs with constraint: `label in {X}` and `indegree(n)≤1` and `acyclic`. You can expose a `graph.view_tree(g, root, label?)` that checks constraints or fails.

# 2) Interop & provenance

* **Provenance** on graph and on (node|edge) attrs: `{source_overlay, source_file, span, commit, extractor, version}`.
* **Formats**: import/export DOT, GraphML, and a compact JSON Lines form:

  * `{"n": id, "a": {...}}` for nodes; `{"e": [src, dst, "label", {"w":1.0,...}]}` for edges.
* **Indexing hooks**: optional secondary indices (by label, by attr key/value).

# 3) API surface (MVP)

Deterministic, total functions; avoid fancy names—make semantics crisp. In ns terms, think pure functions returning new graphs/values.

* **Construction & mutation (persistent)**
  `graph.new()`
  `graph.add_node(g, id, attrs={}) -> g'`
  `graph.upd_node(g, id, patch) -> g'`
  `graph.add_edge(g, src, dst, label="", attrs={}) -> g'`
  `graph.del_node(g, id) -> g'`
  `graph.del_edge(g, src, dst, label?)-> g'`
* **Introspection**
  `graph.nodes(g) -> list<NodeID>`
  `graph.edges(g, label?)-> list<Edge>`
  `graph.out_neighbors(g, id, label?)` / `graph.in_neighbors(...)`
  `graph.indegree/graph.outdegree(g, id, label?)`
* **Queries**
  `graph.filter_edges(g, label|pred) -> g'`
  `graph.filter_nodes(g, pred) -> g'`
  `graph.project(g, keep_node_attrs=[...], keep_edge_attrs=[...]) -> g'`
  `graph.contract(g, groups: map<NodeID->GroupID>) -> g'`  (for condensation)
* **Algorithms (deterministic)**

  * `graph.has_cycle(g) -> bool`
  * `graph.toposort(g) -> list<NodeID>` (fails if cyclic; or returns `(order, is_dag)`)
  * `graph.scc(g) -> list<set<NodeID>>` (Tarjan/Kosaraju)
  * `graph.condensation(g) -> g'` (DAG of SCCs)
  * `graph.roots(g, label?) -> list<NodeID>` (indegree 0 under label filter)
  * `graph.sinks(g, label?) -> list<NodeID>`
  * `graph.ancestors(g, id, label?) -> set<NodeID>` / `graph.descendants(...)`
  * `graph.shortest_path_dag(g, src, dst, weight_attr?) -> path` (for import layers)
  * `graph.longest_path_dag(g, src?, weight_attr?)` (useful on DAGs only)
  * `graph.articulation_points/graph.bridges` (on undirected view when needed)
  * `graph.dominators(cfg, entry)` and `graph.post_dominators(cfg, exit)` (for CFG work)
  * `graph.k_core(g, k)` (helps find tightly coupled clusters)
* **Views**
  `graph.as_undirected(g, label?) -> g'`
  `graph.view_tree(g, root, label?) -> tree` (validates constraints)

Notes on your names:

* `isdag(g)` ≡ `!graph.has_cycle(g)` and all edges treated as directed.
* `hasloops(g)` is ambiguous; prefer `has_cycle`.
* `findrootiest(g)`—define it explicitly. Options:

  1. “roots” = indegree 0 under a label.
  2. “rootiest” = argmin indegree, tie-break by max reachable set size. Provide both `roots()` and `root_score()`; don’t slip in folklore.

# 4) Code-analysis use cases this unlocks

* **Package import DAG**: detect cycles, layer packages by topo rank, compute “root packages” (no inbound imports), surface “bridges” (packages whose removal splits modules).
* **Call graph**: SCCs to find mutually recursive clusters; k-core to reveal “hot” subsystems; betweenness (later) to find chokepoints.
* **CFG**: dominators/post-dominators, natural loop detection from back-edges in a DFS tree; compute loop nests.
* **Type/reference graph**: cluster by SCC, detect “god types,” measure instability (afferent/efferent coupling).
* **Overlay dependency graph**: keep your overlay pipeline honest; topologically schedule overlay builds and invalidations.

# 5) Performance & scale

* Favor adjacency lists with integer-dense IDs; maintain `label→edge_index` for O(1) filtering by label.
* Algorithms: Tarjan SCC is linear `O(V+E)`. Dominators via Lengauer–Tarjan is near-linear; implement that only once you need CFG power.
* **Incremental updates**: start batch-only (recompute). For imports/callgraphs, changes are often localized; later you can add incremental SCC/toposort if it’s hot.
* **Caps & safety**: put a hard cap (configurable) on node/edge counts per graph build; spill to disk/LMDB if needed.

# 6) Determinism & testing

* All traversals must be **order-stable**: when choices tie, sort by NodeID to avoid heisenbugs.
* Provide “witnesses” for decisions: e.g., `toposort` returns both order and a reason if no DAG (the cycle it found).
* Golden tests on small graphs: build helpers like `graph.parse_dot(text)` to seed fixtures.

# 7) NS surface sketch (illustrative)

```
let g = graph.new()
g = graph.add_node(g, "pkg:fdm/core", {"kind":"package"})
g = graph.add_node(g, "pkg:fdm/overlay", {"kind":"package"})
g = graph.add_edge(g, "pkg:fdm/overlay", "pkg:fdm/core", "imports")

if graph.has_cycle(graph.filter_edges(g, "imports")) {
  emit.warn("import cycle detected")
}

let order, is_dag = graph.toposort(graph.filter_edges(g, "imports"))
if is_dag {
  ns.cmd("build.schedule", {"packages": order})
} else {
  let sccs = graph.scc(g)
  ns.emit("dep.sccs", {"clusters": sccs})
}
```

# 8) “Tree as subtype” without type contortions

Expose `tree` as a thin checked view over `graph`:

* `tree.new(root)` builds a graph with constraints.
* All tree ops (`parent`, `children`, `lca`) delegate to graph ops with checks.
* Internally, it’s one engine; externally, tree users get crisp guarantees.

# 9) What not to do (yet)

* Don’t ship centrality metrics (betweenness, PageRank) in v1; they’re expensive and not core to code analysis.
* Don’t add weighted shortest-paths beyond DAG weights; you don’t need Dijkstra for imports.
* Don’t allow arbitrary user-defined edge semantics without labels; you’ll drown in ambiguity.

---

## Recommended v1 backlog (in order)

1. Core type + add/get/delete + label filtering + indegree/outdegree
2. `has_cycle`, `toposort`, `roots/sinks`
3. `scc` + `condensation`
4. `ancestors/descendants` + `as_undirected` + `articulation_points`
5. CFG add-on: `dominators`, `post_dominators`, loop detection
6. Import/export (DOT, compact JSONL) + provenance

This gives you immediate wins for package/call/overlay graphs and a clean runway to CFG/DFG when you need it. If you want, I’ll draft the ns function signatures and the Go backing structs/interfaces to slot into your overlay engine next.





AJP: that looks good. names -- I know nothing about graph theory, those were guesses, forget them and use standard ones.
Interop -- we have some for trees, but the most important use-case would be working on graphs derived from fdm overlays.
Obvs graphs will be worked on via handles -- read only is fine by me.
Please review, ponder deeply and update.
Also, explaining some of the complex tools in more detail would be nice :) -- ir BoLB (bear of little brain)


gpt5: Love it. Let’s lock this down so FDM can treat “graph” as a first-class analytic value while staying simple, immutable, and honest about provenance. I’ll use **standard names** and keep everything **handle-based** (read-only views). Trees are just checked views over graphs.

# What you add to NS/FDM

## 1) Opaque graph handles (read-only)

* **Type**: `graph` (directed, labeled). Returned as an opaque **handle** that points into an internal Graph Store.
* **Creation (derive-only in v1)**
  Graphs are **derived** from overlays (imports, calls, CFG, type refs, “contains”, etc.). No user mutation API in v1.
* **Determinism**: NodeIDs internally normalized to dense uint64; external names kept as attributes.
* **Provenance** stored with every graph: `{overlays, relations, filters, commit, extractor, versions, build_time}`.

### NS surface (MVP)

```
graph.from_overlay(overlay_handle, relation: string, opts={}) -> graph
graph.from_overlays(list<overlay_handle>, relation: string, opts={}) -> graph
graph.filter_edges(g: graph, label: string | pred) -> graph
graph.induced_subgraph(g: graph, nodes: list<string|id>) -> graph
graph.as_undirected(g: graph, label?: string) -> graph

graph.labels(g) -> set<string>
graph.order(g) -> int          # |V|
graph.size(g) -> int           # |E|
graph.sources(g, label?) -> list<NodeID>   # indegree==0
graph.sinks(g, label?)   -> list<NodeID>   # outdegree==0

graph.is_cyclic(g) -> bool
graph.is_dag(g)    -> bool
graph.toposort(g)  -> {order: list<NodeID>|null, cycle: list<Edge>|null}
graph.scc(g)       -> {components: list<set<NodeID>>, comp_of: map<NodeID->int>}
graph.condensation(g) -> {dag: graph, comp_of: map<NodeID->int>}

# CFG-centric:
graph.dominator_tree(cfg: graph, entry: NodeID) -> {idom: map<NodeID->NodeID>, tree: graph}
graph.post_dominator_tree(cfg: graph, exit: NodeID) -> {pidom: map<NodeID->NodeID>, tree: graph}
graph.natural_loops(cfg: graph, entry: NodeID) -> list<{header: NodeID, members: set<NodeID>}>
graph.loop_nest_tree(cfg: graph, entry: NodeID) -> tree

# Utilities:
graph.ancestors(g, v, label?)   -> set<NodeID>
graph.descendants(g, v, label?) -> set<NodeID>
graph.articulation_points(g_ud: graph) -> set<NodeID>
graph.bridges(g_ud: graph) -> set<Edge>

# Export for debugging:
graph.to_dot(g, opts={}) -> string
graph.stats(g) -> map
```

> Trees: `tree.view(g, root)` checks (indegree≤1 & acyclic) under an optional label filter and returns a `tree` handle. Your existing tree tools remain; they just delegate to `graph`.

## 2) Interop: derive graphs from overlays (the main use-case)

You already emit structured relations in overlays. We formalize a tiny relation vocabulary:

* `imports`, `calls`, `cfg`, `type_ref`, `contains`, `owns`, `defines`, `uses`, `overrides`
* Overlay plumbing: `overlay_depends_on`, `produces`, `consumes`

**Derivation examples**

```
# Package import DAG
let g = graph.from_overlay(gao, "imports")

# Call graph (per-package or whole repo)
let cg = graph.from_overlay(gao, "calls", {"scope": "pkg:fdm/core"})

# A function's CFG from GAO
let cfg = graph.from_overlay(gao, "cfg", {"func": "fdm/core/plan.Build"})
```

**Why handles only?**

* You don’t want in-place mutation semantics (too many foot-guns).
* Every transform returns a fresh handle with lineage; you can cache or free them.

## 3) Failure modes come with witnesses

* `graph.toposort(g)` returns `{order, cycle}`. If cyclic, `order=null` and `cycle` is a **concrete cycle** (list of edges) you can report.
* `graph.is_dag(g)` is just `!graph.is_cyclic(g)`, but keep both for ergonomics.
* `graph.stats(g)` surfaces: `|V|, |E|, label_histogram, max_outdegree, max_indegree, scc_count`.

---

# How to use this in FDM work (practical flows)

### A) Scheduling overlay builds safely

```
let depg = graph.from_overlay(fdm_overlays, "overlay_depends_on")
let topo = graph.toposort(depg)
if topo.order != null {
  ns.cmd("overlays.build.schedule", {"order": topo.order})
} else {
  emit("overlay.dep.cycle", {"witness": topo.cycle})
}
```

### B) Detect and explain import cycles

```
let ig = graph.from_overlay(gao, "imports")
if graph.is_cyclic(ig) {
  let w = graph.toposort(ig).cycle
  emit("imports.cycle", {"cycle": w})  # presents the actual package loop
}
```

### C) Loop structure for an optimizer or analyzer

```
let cfg = graph.from_overlay(gao, "cfg", {"func": F})
let dom = graph.dominator_tree(cfg, entry=ENTRY(F))
let loops = graph.natural_loops(cfg, entry=ENTRY(F))
emit("cfg.loops", {"function": F, "loops": loops})
```

---

# BoLB explanations of the “scary” tools

I’ll keep these crisp, with “why” and “what you get back”.

## Strongly Connected Components — `graph.scc`

**What:** Groups where every node can reach every other (in directed sense).
**Why you care:** A component with ≥2 nodes (or a self-loop) **is a cycle**. Collapsing SCCs cleans tangled graphs.
**Output:** `components: [ {n1,n2,...}, ... ]` and `comp_of` map for quick lookups.

## Condensation DAG — `graph.condensation`

**What:** Collapse each SCC to a single node; the result is always a **DAG**.
**Why:** Lets you schedule or layer “tangled clumps” as single steps.
**Output:** `{dag, comp_of}` so you can label condensed nodes with original membership.

## Topological Sort — `graph.toposort`

**What:** An ordering of nodes that respects edge direction; only exists for DAGs.
**Why:** Build orders, migration steps, init sequences.
**Behavior:** If a cycle exists, you get a **cycle witness** instead of an order.

## Dominators / Post-Dominators — `graph.dominator_tree`, `graph.post_dominator_tree`

**What (dominators):** In a **CFG**, node A **dominates** node B if **all** paths from function entry to B go through A. “Immediate dominator” (idom) is the closest such A.
**Why:**

* Find **single-entry regions** (basic blocks that gate everything after them).
* Place invariants or instrumentation once at the dominator instead of everywhere.
  **Post-dominators:** Same idea from the **exit** backward—great for reasoning about guaranteed cleanup and exception paths.
  **Output:** Maps: `idom[v] = u`. Also a tree handle you can walk.

## Natural Loops & Loop Nests — `graph.natural_loops`, `graph.loop_nest_tree`

**What:** A **back edge** `u→h` where `h` **dominates** `u` defines a loop with **header** `h`. The loop is `h` plus everything that can reach `u` without passing through `h`.
**Why:** Precisely identify loops, their headers, and nesting—critical for any optimizer or “hot path” analysis.
**Output:** Each loop: `{header, members}`. A nest tree orders them by containment.

## Articulation Points & Bridges — `graph.articulation_points`, `graph.bridges`

**What:** In the **undirected view**, vertices/edges whose removal disconnects the graph.
**Why:** Flag fragile chokepoints in import or call graphs (e.g., a “god package”).
**Output:** Sets of nodes/edges.

## Ancestors / Descendants — `graph.ancestors`, `graph.descendants`

**What:** All nodes that can reach (or be reached from) a node under a label filter.
**Why:** “What must be built before X?” / “What will X pull in if we touch it?”

---

# Naming & semantics (standards, not vibes)

* `is_cyclic`, `is_dag`, `toposort`, `scc`, `condensation`, `dominators`, `post_dominators`, `natural_loops`, `articulation_points`, `bridges`, `sources`, `sinks`, `ancestors`, `descendants`.
* No jokey names in API; if you want “rootiest” UX, implement it as:
  `sources(g)` then, if empty, `argmin indegree` with `argmax reachability` as a heuristic score—but keep API explicit.

---

# Implementation notes (Go, inside FDM)

* **Graph Store**

  ```go
  type GraphID string

  type GraphStore interface {
      FromOverlay(ctx context.Context, ov OverlayHandle, rel string, opts Map) (GraphID, error)
      FromOverlays(ctx context.Context, ovs []OverlayHandle, rel string, opts Map) (GraphID, error)
      FilterEdges(ctx context.Context, g GraphID, pred EdgePred) (GraphID, error)
      InducedSubgraph(ctx context.Context, g GraphID, nodes []NodeID) (GraphID, error)
      AsUndirected(ctx context.Context, g GraphID, label *string) (GraphID, error)

      Labels(ctx context.Context, g GraphID) ([]string, error)
      Order(ctx context.Context, g GraphID) (int, error)
      Size(ctx context.Context, g GraphID) (int, error)

      Toposort(ctx context.Context, g GraphID) (order []NodeID, cycle []EdgeID, err error)
      SCC(ctx context.Context, g GraphID) (components [][]NodeID, compOf map[NodeID]int, err error)
      Condensation(ctx context.Context, g GraphID) (dag GraphID, compOf map[NodeID]int, err error)
      DominatorTree(ctx context.Context, cfg GraphID, entry NodeID) (idom map[NodeID]NodeID, tree GraphID, err error)
      // ... plus Ancestors/Descendants/Articulation/Bridges
  }
  ```

  Backing store: compact CSR/adjacency with per-label edge indexes; stable NodeID mapping; LZ4-compressed blobs cached by GraphID hash. All traversals use **ID-sorted** queues to stay deterministic.

* **Provenance** is attached once at build and carried through derivations (`parent_graph`, `op`, `params`).

* **Witnesses**: keep a tiny bounded buffer during DFS/Kahn to emit one specific cycle; it’s gold for debugging.

* **Safety**: refuse builds over caps (e.g., >2e6 edges) unless `opts.allow_large=true`. Emit stats before bailing.

---

# A few “recipes” you’ll actually use

* **Layered imports view** (useful visualization / “how deep are we?”):

  ```
  let ig = graph.from_overlay(gao, "imports")
  let dag = graph.condensation(ig).dag
  let layers = graph.topological_layers(dag)   # minor helper: Kahn levels
  emit("imports.layers", {"layers": layers})
  ```

* **Find chokepoints in calls**:

  ```
  let cg  = graph.from_overlay(gao, "calls", {"scope": "pkg:fdm/core"})
  let ud  = graph.as_undirected(cg, "calls")
  let aps = graph.articulation_points(ud)
  emit("calls.chokepoints", {"functions": aps})
  ```

* **Explain a CFG loop**:

  ```
  let cfg = graph.from_overlay(gao, "cfg", {"func": F})
  let loops = graph.natural_loops(cfg, ENTRY(F))
  for L in loops { emit("cfg.loop", {"header": L.header, "size": len(L.members)}) }
  ```

---

# TL;DR for BoLB

* **SCC** finds the tangled clumps.
* **Condensation** squashes clumps into a tidy DAG so you can schedule.
* **Toposort** gives the build/run order, or a specific cycle if that’s impossible.
* **Dominators** tell you where control must pass; **natural loops** tell you where it keeps circling.
* **Articulation/Bridges** flag “break this and the graph splits” hotspots.

If you want, I can turn this into concrete NS stubs (`graph.*` function docs + error codes) and a tiny test corpus (a 7-node import graph with one cycle, one bridge, and a CFG with two nested loops) so you can drop it into your overlay tests and see green bars.


gpt5: nice—here’s a drop-in starter pack: **NS API docs, NS stubs, and a tiny test corpus** you can wire straight into your overlay tests. It’s handle-based, read-only, deterministic, and comes with clear error codes plus “witness” payloads for failures.

@@@

# file: docs/graph.api.md

# version: 0.1.0

# status: draft

# purpose: Public NS-facing API for read-only graph analysis

## Model

* `graph` — opaque handle referencing an internal, immutable, directed, labeled graph.
* `NodeID` — stable, store-local identifier (string in NS; internally dense uint64).
* `Edge` — `{src: NodeID, dst: NodeID, label: string, attrs?: map}`.
* Graphs are **derived** from overlays or other graphs; no in-place mutation.

## Naming (standard graph theory)

`is_cyclic`, `is_dag`, `toposort`, `scc`, `condensation`, `dominators`, `post_dominators`,
`natural_loops`, `articulation_points`, `bridges`, `sources`, `sinks`, `ancestors`, `descendants`,
`labels`, `order`, `size`, `as_undirected`, `filter_edges`, `induced_subgraph`, `stats`.

## Provenance

Every handle carries:

```
{ parents?: [graph], source_overlays?: [overlay_handle], relation?: string,
  params?: map, commit?: string, extractor?: string, versions?: map, built_at?: timestamp }
```

---

## Functions

### Build / Derive

* `graph.from_overlay(ov: overlay, relation: string, opts?: map) -> graph`
* `graph.from_overlays(ovs: list<overlay>, relation: string, opts?: map) -> graph`

`opts` common keys:

* `scope` (string): narrow by package/module/etc.
* `func` (string): for CFG/DFG derivations.
* `allow_large` (bool): override size caps for big graphs.

Errors: `E.GRAPH.NO_RELATION`, `E.GRAPH.BUILD_FAILED`, `E.GRAPH.LIMIT_EXCEEDED`.

### Basic info

* `graph.labels(g) -> set<string>`
* `graph.order(g) -> int`      # |V|
* `graph.size(g) -> int`       # |E|
* `graph.stats(g) -> map`      # {V, E, label\_hist, max\_in, max\_out, scc\_count?}

Errors: `E.GRAPH.BAD_HANDLE`.

### Structure queries

* `graph.sources(g, label?: string) -> list<NodeID>`   # indegree==0 (optionally under label)
* `graph.sinks(g, label?: string) -> list<NodeID>`
* `graph.ancestors(g, v: NodeID, label?: string) -> set<NodeID>`
* `graph.descendants(g, v: NodeID, label?: string) -> set<NodeID>`

Errors: `E.GRAPH.BAD_HANDLE`, `E.GRAPH.NO_SUCH_NODE`, `E.GRAPH.NO_SUCH_LABEL`.

### Views / Filters

* `graph.filter_edges(g, label_or_pred) -> graph`      # returns new handle

  * `label_or_pred` may be a string label or a predicate closure: `(e: Edge) -> bool`.
* `graph.induced_subgraph(g, nodes: list<NodeID>) -> graph`
* `graph.as_undirected(g, label?: string) -> graph`

Errors: `E.GRAPH.BAD_HANDLE`, `E.GRAPH.NO_SUCH_LABEL`.

### DAG / Cycles

* `graph.is_cyclic(g) -> bool`
* `graph.is_dag(g) -> bool`
* `graph.toposort(g) -> {order?: list<NodeID>, cycle?: list<Edge>}`

  * If cyclic: `order` omitted, `cycle` returns a concrete cycle (edge list).

Errors: `E.GRAPH.BAD_HANDLE`.

### SCC & Condensation

* `graph.scc(g) -> {components: list<set<NodeID>>, comp_of: map<NodeID,int>}`
* `graph.condensation(g) -> {dag: graph, comp_of: map<NodeID,int>}`

Errors: `E.GRAPH.BAD_HANDLE`.

### CFG-specific

* `graph.dominator_tree(cfg: graph, entry: NodeID) -> {idom: map<NodeID,NodeID>, tree: graph}`
* `graph.post_dominator_tree(cfg: graph, exit: NodeID) -> {pidom: map<NodeID,NodeID>, tree: graph}`
* `graph.natural_loops(cfg: graph, entry: NodeID) -> list<{header: NodeID, members: set<NodeID>}>`
* `graph.loop_nest_tree(cfg: graph, entry: NodeID) -> tree`   # returns checked tree view

Errors: `E.GRAPH.NOT_CFG`, `E.GRAPH.NO_ENTRY`, `E.GRAPH.NO_EXIT`.

### Debug export

* `graph.to_dot(g, opts?: map) -> string`  # DOT text; deterministic order

Errors: `E.GRAPH.BAD_HANDLE`.

---

## Error Codes (NS constants)

* `E.GRAPH.BAD_HANDLE`        — unknown/expired handle
* `E.GRAPH.NO_RELATION`       — overlay(s) don’t expose the relation
* `E.GRAPH.NO_SUCH_LABEL`     — label filter not present in graph
* `E.GRAPH.NO_SUCH_NODE`      — node not in graph
* `E.GRAPH.BUILD_FAILED`      — extractor/derivation failed (see `.reason`)
* `E.GRAPH.LIMIT_EXCEEDED`    — graph exceeded configured caps
* `E.GRAPH.NOT_CFG`           — operation requires CFG semantics
* `E.GRAPH.NO_ENTRY`          — missing entry node for CFG op
* `E.GRAPH.NO_EXIT`           — missing exit node for CFG op

On failure, functions **emit a witness** when practical:

* `toposort`: `cycle: list<Edge>`
* `from_overlay`: `{missing: relation, overlays: [...]}` in `.reason`
* CFG ops: `{expected: "entry", have: [...]}`
  @@@

@@@

# file: stdlib/graph.ns

# version: 0.1.0

# purpose: NS stubs that bind to the FDM Graph Store (opaque, read-only)

ns.module "graph"

# --- Intrinsics (provided by runtime / bridge) ---

extern tool "tool.fdm.graph.FromOverlay"      # (ov, relation, opts) -> graph\_handle
extern tool "tool.fdm.graph.FromOverlays"     # (ovs, relation, opts) -> graph\_handle
extern tool "tool.fdm.graph.Labels"           # (g) -> set<string>
extern tool "tool.fdm.graph.Order"            # (g) -> int
extern tool "tool.fdm.graph.Size"             # (g) -> int
extern tool "tool.fdm.graph.Stats"            # (g) -> map
extern tool "tool.fdm.graph.FilterEdges"      # (g, pred\_spec) -> g'
extern tool "tool.fdm.graph.Induced"          # (g, nodes) -> g'
extern tool "tool.fdm.graph.AsUndirected"     # (g, label?) -> g'
extern tool "tool.fdm.graph.Sources"          # (g, label?) -> \[NodeID]
extern tool "tool.fdm.graph.Sinks"            # (g, label?) -> \[NodeID]
extern tool "tool.fdm.graph.Ancestors"        # (g, v, label?) -> set<NodeID>
extern tool "tool.fdm.graph.Descendants"      # (g, v, label?) -> set<NodeID>
extern tool "tool.fdm.graph.IsCyclic"         # (g) -> bool
extern tool "tool.fdm.graph.Topo"             # (g) -> {order?:\[NodeID], cycle?:\[Edge]}
extern tool "tool.fdm.graph.SCC"              # (g) -> {components:\[\[NodeID]], comp\_of\:map}
extern tool "tool.fdm.graph.Condensation"     # (g) -> {dag: graph, comp\_of: map}
extern tool "tool.fdm.graph.DominatorTree"    # (g, entry) -> {idom: map, tree: graph}
extern tool "tool.fdm.graph.PostDominatorTree" # (g, exit) -> {pidom: map, tree: graph}
extern tool "tool.fdm.graph.NaturalLoops"     # (g, entry) -> \[ {header, members} ]
extern tool "tool.fdm.graph.LoopNestTree"     # (g, entry) -> tree\_handle
extern tool "tool.fdm.graph.ToDOT"            # (g, opts?) -> string

# --- Public NS Functions (light wrappers; argument checks; deterministic defaults) ---

fn from\_overlay(ov, relation: string, opts?: map) -> graph {
if relation == "" { error E.GRAPH.NO\_RELATION, {"reason":"empty relation"} }
return call tool.fdm.graph.FromOverlay with (ov, relation, opts ?? {})
}

fn from\_overlays(ovs: list, relation: string, opts?: map) -> graph {
if len(ovs) == 0 { error E.GRAPH.BUILD\_FAILED, {"reason":"empty overlays"} }
if relation == "" { error E.GRAPH.NO\_RELATION, {"reason":"empty relation"} }
return call tool.fdm.graph.FromOverlays with (ovs, relation, opts ?? {})
}

fn labels(g: graph) -> set<string> {
return call tool.fdm.graph.Labels with (g)
}

fn order(g: graph) -> int { return call tool.fdm.graph.Order with (g) }
fn size(g: graph)  -> int { return call tool.fdm.graph.Size with (g) }
fn stats(g: graph) -> map { return call tool.fdm.graph.Stats with (g) }

fn filter\_edges(g: graph, label\_or\_pred) -> graph {

# Predicates are serialized to a small DSL by the bridge; strings are label filters.

let spec = if is\_string(label\_or\_pred) { {"label": label\_or\_pred} } else { {"pred": label\_or\_pred} }
return call tool.fdm.graph.FilterEdges with (g, spec)
}

fn induced\_subgraph(g: graph, nodes: list) -> graph {
return call tool.fdm.graph.Induced with (g, nodes)
}

fn as\_undirected(g: graph, label?: string) -> graph {
return call tool.fdm.graph.AsUndirected with (g, label)
}

fn sources(g: graph, label?: string) -> list { return call tool.fdm.graph.Sources with (g, label) }
fn sinks(g: graph, label?: string)   -> list { return call tool.fdm.graph.Sinks with (g, label) }

fn ancestors(g: graph, v, label?: string) -> set { return call tool.fdm.graph.Ancestors with (g, v, label) }
fn descendants(g: graph, v, label?: string) -> set { return call tool.fdm.graph.Descendants with (g, v, label) }

fn is\_cyclic(g: graph) -> bool { return call tool.fdm.graph.IsCyclic with (g) }
fn is\_dag(g: graph) -> bool { return !is\_cyclic(g) }

fn toposort(g: graph) -> map {

# Returns {order?: list<NodeID>, cycle?: list<Edge>}

return call tool.fdm.graph.Topo with (g)
}

fn scc(g: graph) -> map {
return call tool.fdm.graph.SCC with (g)
}

fn condensation(g: graph) -> map {
return call tool.fdm.graph.Condensation with (g)
}

# --- CFG-specific ---

fn dominator\_tree(cfg: graph, entry) -> map {
return call tool.fdm.graph.DominatorTree with (cfg, entry)
}

fn post\_dominator\_tree(cfg: graph, exit) -> map {
return call tool.fdm.graph.PostDominatorTree with (cfg, exit)
}

fn natural\_loops(cfg: graph, entry) -> list {
return call tool.fdm.graph.NaturalLoops with (cfg, entry)
}

fn loop\_nest\_tree(cfg: graph, entry) -> tree {
return call tool.fdm.graph.LoopNestTree with (cfg, entry)
}

fn to\_dot(g: graph, opts?: map) -> string {
return call tool.fdm.graph.ToDOT with (g, opts ?? {"sort":"id"})
}

# --- Test-only helpers (guarded by runtime) ---

extern tool "tool.fdm.graph.test.LoadJSONL"   # (path) -> graph
extern tool "tool.fdm.graph.test.Kind"        # (g) -> string  # e.g., "cfg", "imports"

fn test\_load\_jsonl(path: string) -> graph {
return call tool.fdm.graph.test.LoadJSONL with (path)
}

fn test\_kind(g: graph) -> string {
return call tool.fdm.graph.test.Kind with (g)
}
@@@

@@@

# file: tests/graph\_fixtures.jsonl

# version: 0.1.0

# purpose: Minimal fixtures: (1) 7-node import graph w/ one cycle + one bridge; (2) CFG with two nested loops.

# ---------- (1) IMPORT GRAPH ----------

# Nodes: A,B,C form a cycle; D is a bridge from C to the rest; F-G chain hangs off D; E depends on D.

# Undirected articulation point: D (removing D disconnects {A,B,C} from {E,F,G})

{"kind":"graph","name":"imports\_demo","relation":"imports","nodes":\[
{"id":"A","attrs":{"kind":"pkg"}},
{"id":"B","attrs":{"kind":"pkg"}},
{"id":"C","attrs":{"kind":"pkg"}},
{"id":"D","attrs":{"kind":"pkg"}},
{"id":"E","attrs":{"kind":"pkg"}},
{"id":"F","attrs":{"kind":"pkg"}},
{"id":"G","attrs":{"kind":"pkg"}}
]}
{"kind":"edge","label":"imports","src":"A","dst":"B"}
{"kind":"edge","label":"imports","src":"B","dst":"C"}
{"kind":"edge","label":"imports","src":"C","dst":"A"}   # cycle A->B->C->A
{"kind":"edge","label":"imports","src":"C","dst":"D"}   # D is bridge to the rest
{"kind":"edge","label":"imports","src":"D","dst":"E"}
{"kind":"edge","label":"imports","src":"D","dst":"F"}
{"kind":"edge","label":"imports","src":"F","dst":"G"}

# ---------- (2) CFG WITH TWO NESTED LOOPS ----------

# Function foo: ENTRY->B->C->D->B (outer loop), C->E->C (inner loop), D->XIT exit path

{"kind":"graph","name":"cfg\_foo","relation":"cfg","func":"foo","nodes":\[
{"id":"ENTRY","attrs":{"kind":"bb"}},
{"id":"B","attrs":{"kind":"bb","role":"header\_outer"}},
{"id":"C","attrs":{"kind":"bb","role":"header\_inner"}},
{"id":"D","attrs":{"kind":"bb"}},
{"id":"E","attrs":{"kind":"bb"}},
{"id":"XIT","attrs":{"kind":"bb","role":"exit"}}
]}
{"kind":"edge","label":"cfg","src":"ENTRY","dst":"B"}   # entry
{"kind":"edge","label":"cfg","src":"B","dst":"C"}       # outer loop body
{"kind":"edge","label":"cfg","src":"C","dst":"D"}       # path continues
{"kind":"edge","label":"cfg","src":"D","dst":"B"}       # back-edge => outer loop (header B)
{"kind":"edge","label":"cfg","src":"C","dst":"E"}       # inner loop branch
{"kind":"edge","label":"cfg","src":"E","dst":"C"}       # back-edge => inner loop (header C)
{"kind":"edge","label":"cfg","src":"D","dst":"XIT"}     # normal exit
@@@

@@@

# file: tests/graph\_api\_spec.ns

# version: 0.1.0

# purpose: Sanity tests using fixtures; intended to run in overlay test harness

proc run() {

# --- Imports graph ---

let ig = graph.test\_load\_jsonl("tests/graph\_fixtures.jsonl#imports\_demo")

assert(graph.size(ig) == 7)                        # 7 edges total
assert(graph.order(ig) == 7)                       # 7 nodes
assert(graph.is\_cyclic(ig) == true)

let topo = graph.toposort(ig)
assert(is\_null(topo.order))
assert(len(topo.cycle) >= 3)                       # witness present

let ud = graph.as\_undirected(ig, "imports")
let aps = algo\_sort(graph.articulation\_points(ud)) # deterministically sort for snapshot
assert(aps == \["D"])                               # D is the chokepoint

# --- Condensation: cycle A,B,C collapses to one SCC node; dag becomes acyclic

let s = graph.scc(ig)
assert(len(s.components) == 5)                     # {ABC}, D, E, F, G
let cd = graph.condensation(ig).dag
assert(graph.is\_dag(cd) == true)

# --- CFG graph ---

let cfg = graph.test\_load\_jsonl("tests/graph\_fixtures.jsonl#cfg\_foo")
assert(graph.is\_cyclic(cfg) == true)               # loops present (CFG’s are generally cyclic)

let idom = graph.dominator\_tree(cfg, "ENTRY").idom

# Headers dominate their bodies:

assert(idom\["B"] == "ENTRY")
assert(idom\["C"] == "B")
assert(idom\["D"] == "C")

# Exit post-dominates nothing but itself; check post-dominators form:

let pidom = graph.post\_dominator\_tree(cfg, "XIT").pidom
assert(pidom\["XIT"] == "XIT")

let loops = graph.natural\_loops(cfg, "ENTRY")

# Expect two loops: headers B and C

let headers = map(L in loops -> L.header) |> set()
assert(headers == set("B","C"))

let dot = graph.to\_dot(cfg, {"sort":"id"})
assert(starts\_with(dot, "digraph"))

emit("tests.graph.ok", {"imports": graph.stats(ig), "cfg": graph.stats(cfg)})
}

# tiny deterministic sort helper for stable assertions

fn algo\_sort(xs: list) -> list {
return sort(xs, fn (a,b) { return a < b })
}
@@@

---

### How to use

1. Drop these files into your repo (`docs/`, `stdlib/`, `tests/`).
2. Bind the `extern tool ...` names to your Go GraphStore methods. Each should be pure, returning new handles.
3. Implement a tiny test loader for `tests/graph_fixtures.jsonl`:

   * Parse the JSONL stream.
   * `{"kind":"graph", name, relation, ...}` starts a new in-memory graph build.
   * `{"kind":"edge", ...}` adds an edge to the current build.
   * Return a handle whose provenance marks `relation` and (for CFG) `func`.

### Design nudges (BoLB-friendly)

* Always return **witnesses** on failure—you’ll save hours of grep-rage.
* Keep traversals **ID-sorted** at tie points; that’s your determinism anchor.
* Don’t over-fit: this set gets you import cycles, build order, chokepoints, SCC cleanup, and proper loop detection—the essentials for code analysis.

If you want, I can also sketch the Go `GraphStore` interface + a concrete in-memory CSR/adjacency implementation scaffold next, wired to those `extern tool` names.


Short version: yes—there are solid, graph-based metrics you can compute on a Go package import graph to quantify “how well structured” it is. Below is a pragmatic, standards-ish set you can ship, plus tiny NS snippets showing how to get each number from the `graph` API we just designed.

I’m using conventional names (Martin’s package metrics where applicable) and graph-theory basics. Treat this as a working scorecard; tune thresholds to your codebase, not dogma.

# Core metrics (what, why, how)

## 0) Sanity: Acyclicity

**What:** Go import graphs must be a DAG at build time.
**Metric:** `acyclic_ok ∈ {0,1}` and optional **tangle ratio** below.
**Healthy target:** 1 (no cycles).

```ns
let ig = graph.from_overlay(gao, "imports")
let acyclic_ok = graph.is_dag(ig) ? 1 : 0
```

## 1) Tangle ratio (SCC mass)

**What:** % of nodes in non-trivial SCCs (size≥2 or self-loop).
**Why:** Measures cyclic “clumps” (should be ≈0 for Go; useful during refactors or cross-build-tag checks).
**Healthy:** 0%; warn ≥ 2%.

```ns
let s = graph.scc(ig)
let tangled = sum(C in s.components -> (len(C) >= 2 ? len(C) : 0))
let tangle_ratio = tangled / graph.order(ig)
```

## 2) Layering depth & span

**What:** Depth = longest path length in the condensation DAG. Span = how far edges jump across layers.
**Why:** Deep chains are fragile; huge cross-layer jumps indicate leakage.
**Healthy:** Depth grows \~log N; fraction of jumps >1 layer small.

```ns
let cd = graph.condensation(ig).dag
# Topological layers via Kahn levels (assume helper exists):
let layers = topo_layers(cd)        # list<list<NodeID>>
let depth = len(layers)             # 1..N
let edges = layer_edge_spans(cd)    # returns list<int> of |layer(dst)-layer(src)|
let span_gt1 = count(d in edges -> d > 1) / max(1, len(edges))
```

## 3) Afferent/Efferent coupling and Instability (Martin)

For each package p:

* **Ca(p)** = inbound imports (afferent).
* **Ce(p)** = outbound imports (efferent).
* **Instability:** `I(p) = Ce / (Ca + Ce)` in \[0,1].
  Stable = near 0 (many depend on it; it depends on few). Unstable = near 1.

**Why:** Packages that many others depend on should be stable (low I) and abstract (see next).
**Healthy:** Distribution makes sense for your layering; also check correlation with depth.

```ns
# Assume degree helpers:
extern tool "tool.fdm.graph.InDegree"   # (g,v,label?) -> int
extern tool "tool.fdm.graph.OutDegree"  # (g,v,label?) -> int

fn instability(g: graph, v) -> float {
  let ca = call tool.fdm.graph.InDegree with (g, v, "imports")
  let ce = call tool.fdm.graph.OutDegree with (g, v, "imports")
  return (ca+ce)==0 ? 0.0 : (ce / (ca+ce))
}
```

## 4) Abstractness and “Distance from Main Sequence” (Martin)

For each package p:

* **Abstractness:** `A(p) = (# exported interfaces) / (# exported types)` (Go proxy for “abstract”).
* **Distance:** `D(p) = | A(p) + I(p) - 1 |` (0 is ideal: stable packages are abstract; unstable are concrete).

**Why:** Stable abstractions principle.
**Healthy:** Lower average D; flag high-D outliers.

```ns
# Pull counts from GAO/type overlay (not the import graph).
# Suppose you can fetch per-package stats:
extern tool "tool.fdm.overlay.TypesSummary"  # (pkg) -> {exported_types: int, exported_interfaces: int}

fn abstractness(pkg: string) -> float {
  let ts = call tool.fdm.overlay.TypesSummary with (pkg)
  return ts.exported_types==0 ? 0.0 : (ts.exported_interfaces / ts.exported_types)
}

fn distance_main_sequence(ig: graph, pkg: string) -> float {
  let I = instability(ig, pkg)
  let A = abstractness(pkg)
  return abs(A + I - 1.0)
}
```

## 5) Chokepoints (articulation points / bridges)

**What:** In undirected view, vertices/edges whose removal disconnects the graph.
**Why:** Single points of architectural failure.
**Metrics:** `ap_ratio = |AP|/|V|`, `bridge_ratio = |bridges|/|E|`; list worst offenders by Ca·Ce.

```ns
let ud = graph.as_undirected(ig, "imports")
let aps = graph.articulation_points(ud)
let brs = graph.bridges(ud)
let ap_ratio = len(aps) / graph.order(ud)
let bridge_ratio = len(brs) / max(1, graph.size(ud))
```

## 6) Dependency concentration (inequality)

**What:** How unequal are out-degrees (Ce) and transitive descendants?
**Why:** A few “god packages” create fragility.
**Metric:** Gini coefficient (or 95th/median) over `Ce(p)` and over `|descendants(p)|`.

```ns
fn gini(xs: list<float>) -> float { 
  # standard O(n log n) gini; implement once in stdlib
}
let ces = map(v in packages -> instability_ce_only(ig, v))   # outdegree only
let gini_ce = gini(ces)
```

## 7) Volatility alignment (optional but powerful)

**What:** Compare git change frequency with instability: unstable packages (high I) should change more; stable ones less.
**Metric:** Spearman correlation between `(1−I)` and change-rate; negative is bad (stable code churns).
**Why:** Architectural anti-patterns show up as volatility mismatches.

---

# Roll-up: an “Architecture Health Index” (AHI)

You can roll these into a single 0–100 score to track in CI. Don’t worship the number—use it to spot drifts.

Example weighting (tweak per repo size):

* **Hard fail**: `acyclic_ok == 0` → AHI = 0 and list cycle witness.
* Base 100, then subtract:

  * `20 * clamp01(tangle_ratio / 0.05)`              # up to −20 at 5% tangled
  * `10 * clamp01((depth / (1+log2(N))) - 1)^+`       # penalty if deeper than \~log₂N
  * `15 * mean_p(D(p)) / 0.5`                         # −15 at avg D = 0.5
  * `10 * ap_ratio / 0.10`                            # −10 at 10% APs
  * `10 * (gini_ce - 0.4)^+ / 0.4`                    # penalize heavy inequality
  * `10 * span_gt1`                                   # penalize cross-layer jumps
  * `10 * max(0, -spearman((1−I), change_rate))`      # volatility misalignment

Where `N = |V|`, `mean_p` is a robust mean (trimmed or median of D), and `x^+` is `max(0,x)`.

---

# Tiny NS: per-package table + overall score

```ns
proc dep_health() {
  let ig = graph.from_overlay(gao, "imports")
  if !graph.is_dag(ig) {
    let w = graph.toposort(ig).cycle
    emit("deps.cycle", {"witness": w})
    emit("deps.ahi", {"score": 0})
    return
  }

  let pkgs = pkg_list_from_overlay(gao)          # your helper
  let rows = []
  for p in pkgs {
    let I  = instability(ig, p)
    let A  = abstractness(p)
    let D  = abs(A + I - 1.0)
    rows = append(rows, {"pkg": p, "Ca": indeg(ig,p), "Ce": outdeg(ig,p), "I": I, "A": A, "D": D})
  }

  # Layer/Span
  let cd = graph.condensation(ig).dag
  let layers = topo_layers(cd)
  let depth = len(layers)
  let span_gt1 = frac_span_gt1(cd)

  # Chokepoints / Tangle
  let ud = graph.as_undirected(ig, "imports")
  let aps = graph.articulation_points(ud)
  let brs = graph.bridges(ud)
  let ap_ratio = len(aps)/graph.order(ig)
  let s = graph.scc(ig)
  let tangled = sum(C in s.components -> (len(C) >= 2 ? len(C) : 0))
  let tangle_ratio = tangled / graph.order(ig)

  # Inequality
  let ces = map(r in rows -> r.Ce)
  let gini_ce = gini(ces)

  # AHI roll-up (skip volatility here for brevity)
  let N = graph.order(ig)
  let base = 100.0
  let penalty = 0.0
  penalty += 20 * clamp01(tangle_ratio / 0.05)
  penalty += 10 * clamp01(max(0, (depth / (1+log2(max(2,N)))) - 1))
  penalty += 15 * (robust_mean(map(r in rows -> r.D)) / 0.5)
  penalty += 10 * clamp01(ap_ratio / 0.10)
  penalty += 10 * clamp01(max(0, gini_ce - 0.4) / 0.4)
  penalty += 10 * clamp01(span_gt1)
  let ahi = max(0, base - penalty)

  emit("deps.scorecard", {
    "AHI": round(ahi,1),
    "N": N,
    "depth": depth,
    "span_gt1": round(span_gt1,3),
    "ap_ratio": round(ap_ratio,3),
    "gini_ce": round(gini_ce,3),
    "tangle_ratio": round(tangle_ratio,3),
    "per_package": rows
  })
}

# helpers (you’ll provide backed by extern tools):
fn indeg(g,v) -> int { return call tool.fdm.graph.InDegree with (g, v, "imports") }
fn outdeg(g,v) -> int { return call tool.fdm.graph.OutDegree with (g, v, "imports") }
fn frac_span_gt1(dag) -> float { 
  let L = layer_index(dag)          # map<NodeID->int>
  let spans = edge_spans(dag, L)    # list<int>
  return count(s in spans -> s>1) / max(1,len(spans))
}
```

---

# Reading the tea leaves (BoLB guide)

* **Cycles:** immediate red card. Fix first.
* **Depth too high:** flatten by extracting interfaces and pushing concrete deps downward; break long “beads on a string”.
* **High D(p):** either make it more abstract (add interfaces; narrow exports) or more unstable (push deps outward/up).
* **Many APs/bridges:** split “god packages” by responsibility; introduce “service” boundaries.
* **Gini high:** redistribute outgoing deps; introduce small facades so not everything imports the same kitchen-sink.
* **Span>1 jumps:** introduce intermediate packages to step across layers, or invert dependency via interfaces.

---

# Minimal API additions you’ll want

We used `InDegree`/`OutDegree`, plus a helper to iterate edges when computing spans. Add:

* `graph.indegree(g, v, label?) -> int`
* `graph.outdegree(g, v, label?) -> int`
* `graph.edge_spans(dag) -> list<int>`  (or give access to `(src,dst)` to compute spans client-side)

These keep the rest of the scorecard trivial.

---

Ship this with thresholds that trigger **witness lists** (e.g., top 10 D(p), top APs with Ca·Ce, worst span edges). Feed it into CI and plot AHI over time. That graph tells you whether refactors are paying rent or just moving furniture.
