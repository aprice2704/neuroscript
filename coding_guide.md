# FDM Coding Guide & Architectural Invariants

**Version:** 2.4 (Concurrency & Testing Anti-Patterns)
**Original Version:** 2.3 (Local Cancel Law & Shutdown Determinism)
**Status:** Ratified
**Scope:** Core Go architecture, Schema/DSL, NeuroScript tools, Service patterns, and Subsystem protocols.
**Purpose:** This is the authoritative "Law" for contributing to FDM. It consolidates rules from `graph.md`, `services.md`, and subsystem guides into a single operational manual.

---

## 1. Glossary & Identity Primitives

The FDM system relies on a precise vocabulary for identity. Mixing these concepts is the primary source of bugs.

### 1.1 The Identity Triad

| Term | Prefix | Definition | Properties | Usage Context |
| :--- | :--- | :--- | :--- | :--- |
| **NodeID** | `N_` | **Physical Identity.** A specific, immutable version snapshot of an object. | Content-addressed (hash-based). Never changes once created. | Used for historical lookups, audit logs, `Prev` pointers, and `NodeRef` fields. |
| **EntityID** | `E_` | **Semantic Identity.** The stable identity of a "thing" across time. | Stable. Persists across versions. Links a chain of NodeIDs together. | Used for "current state" operations, mutations, structural links, and `EntityRef` fields. |
| **Handle** | n/a | **Human Identity.** A mutable alias (e.g., `workq/main`, `users/bob`). | Mutable. Not unique across time. Strictly a UI/Config layer concept. | **NEVER** used in compiled logic or core storage. Must be resolved at the boundary. |

### 1.2 Graph Concepts

* **Head:** The most recent `NodeID` for a given `EntityID`. Stored in the `heads` index in memory. This is the authoritative "Now".
* **NodeRef (Kind 5):** A reference field pointing to a specific snapshot (`N_`). It is "Frozen in time". Guaranteed to exist. Use for things like `replies_to` in forums or `result_of` in tasks.
* **EntityRef (Kind 11):** A reference field pointing to the live entity (`E_`). It "Always points to current head". Use for structural links like `owned_by` or `member_of`.
* **NodeDef:** A validated, typed structure ready to be persisted. It serves as the input to `CreateEntity` and `MutateEntity`.
* **Ghost Node:** A node that exists in the append-only log (`mem`) but is not pointed to by any `EntityID` in the `heads` map. Effectively invisible to the system.

---

## 2. The Graph Laws (Architectural Invariants)

These are the non-negotiable rules of the system. Violation of these laws leads to data corruption, "ghost nodes," or "hallucinations."

### Law 1: The Canonical Creation Law
**Rule:** You **MUST** use the generated `nodes.Build<Type>` functions (in Go) or `fdm.nodes.<domain>.<type>.create` tools (in NeuroScript).
**The Violation:** Manually constructing `map[string]any` or using `ix.NewNodeDef` directly for known types.
**The Consequence:** Bypasses schema validation, default values, and type checking. Leads to runtime crashes when fields are missing.
**Fix:** Refactor to use the generated builder found in `code/nodes/gen_<type>.go`.

### Law 2: Ghost Node Prohibition
**Rule:** **NEVER** use `graph.CreateNode` to create a new Entity.
**The Violation:** Calling `graphSvc.CreateNode(def)` for a new object.
**The Consequence:** The node is saved to the log, but the `heads` index is not updated. `ReadHead(EntityID)` returns "Not Found."
**Fix:** Use `graph.CreateEntity(EntityID, NodeDef)` for new things. Use generated mutators for updates.

### Law 3: The Indexing Law (The History Trap)
**Rule:** **NEVER** iterate `mem` (the node log) to find the "current state" of the system.
**The Violation:** Using `graph.Memory().IterateNodes()` or `QueryService` (which scans mem) to find the "latest" version of a node.
**The Consequence:** You will find deleted items ("Hallucinations") and stale versions ("Echoes").
**Fix:**
* To find a specific thing: `graph.ReadHead(EntityID)` (O(1) lookup).
* To scan all active things: `graph.GetHeads(NodeType)` (Returns only the tips of active chains).

### Law 4: The Handle Prohibition
**Rule:** Compiled Go code **MUST NOT** accept or resolve Handles as identifiers.
**The Violation:** A function signature like `func GetQueue(handle string)`.
**The Consequence:** Ambiguity (is it an ID? a handle?), breakage if the handle is renamed, and tighter coupling to the UI layer.
**Fix:** Handles are data, not keys. Resolution happens at the system boundary (UI, CLI, or generated Tool implementing `upsert`), converting Handle -> EntityID before calling the Service Layer.

### Law 5: The Strong Typing Law
**Rule:** In Go, use `ix.EntityID` and `ix.NodeID` types. **NEVER** use raw `string` for IDs in internal APIs.
**The Violation:** `func UpdateUser(id string, data *ix.Entity)`
**The Consequence:** "Stringly typed" errors where a NodeID is passed to a function expecting an EntityID, causing a Ghost Node on write.
**Fix:** Use the types defined in `interfaces`.

### Law 6: The Resolution Law
**Rule:** Use canonical helpers for ID interaction.
* **Validation:** `ix.IsEntityID(s)` / `ix.IsNodeID(s)`.
* **Mutation Targets:** `ix.ResolveToEntityID(g, id)`. If passed a NodeID, it looks up the node to find its owner EntityID.
* **State Reads:** `ix.ResolveToNode(g, id)`. If passed an EntityID, it looks up the current Head.

---

## 3. Identity & Naming Implementation

### 3.1 Canonical ID Formats
All serialized IDs must strictly adhere to the prefixing scheme.

* **NodeID**: `N_<encoded_payload>`
  * Example: `N_01JABC...`
  * Generated by: System (hash of content + sequence).
* **EntityID**: `E_<encoded_payload>`
  * Example: `E_01JXYZ...`
  * Generated by: `ix.NewEntityID()` (typically ULID-based).

### 3.2 The Handle Law (Reserved Field)
The field `handle` is a system-reserved keyword.
* **Constraint:** You **MUST NOT** define a field named `handle` in any DSL schema.
* **Management:** The system automatically manages this field when you use the Handle Service (`handlesvc`).
* **Uniqueness:** A handle must be unique within its `NodeType`.
* **Resolution:** Handles are aliases. They exist to be resolved to `EntityID`s immediately.
* **Display Names:** For UI labels that don't need uniqueness or resolution, use the standard `fdm:displayname` field.

---

## 4. Go Graph Interaction Patterns

### 4.1 Creation Pattern
Use the generated builder from `code/nodes/` to create a definition, then submit it to the graph service.

```go
import (
    "context"
    "[github.com/aprice2704/fdm/code/nodes](https://github.com/aprice2704/fdm/code/nodes)"
    ix "[github.com/aprice2704/fdm/code/interfaces](https://github.com/aprice2704/fdm/code/interfaces)"
)

func CreateMyTask(ctx context.Context, graphSvc ix.Graph) (*ix.Entity, error) {
    // 1. Build Definition (Type-Safe, Validated)
    def, err := nodes.BuildTask(nodes.BuildTaskParams{
        TaskStatus:      nodes.TaskStatusTodo,
        TaskDescription: "Review the new architecture docs",
        TaskPriority:    10,
    })
    if err != nil {
        return nil, err
    }

    // 2. Mint a new EntityID
    eid := ix.NewEntityID()

    // 3. Create Entity (Atomic)
    node, err := graphSvc.CreateEntity(eid, def)
    if err != nil {
        return nil, err
    }

    return ix.NewEntity(node), nil
}
```

### 4.2 Mutation Pattern
Use the generated mutator. It handles the complexity of fetching the current head, applying changes, maintaining the version chain, and updating the index.

```go
func MarkTaskDone(ctx context.Context, graphSvc ix.Graph, taskID ix.EntityID) (*ix.Entity, error) {
    // The generated Mutate function wraps the generic graphSvc.MutateEntity.
    newNode, err := nodes.MutateTask(ctx, graphSvc, taskID, map[string]any{
        nodes.FieldTaskStatus: nodes.TaskStatusDone,
        nodes.FieldTaskProgress: map[string]any{"percent": 100},
    })
    if err != nil {
        return nil, err
    }
    return ix.NewEntity(newNode), nil
}
```

### 4.3 Read Pattern
Choose the read method based on whether you need the "Now" or the "Past".

* **For Current State:**
  `node, found := graphSvc.ReadHead(entityID)`
  * This is an O(1) lookup in the `heads` map. It is always authoritative.

* **For History / Specific Version:**
  `node, found := graphSvc.GetNode(nodeID)`
  * This retrieves an immutable snapshot. **Never** treat this as current state.

* **For Scanning / Indexing:**
  `nodes := graphSvc.GetHeads(nodeType)`
  * This returns the current head of *every* active entity of that type. Use this for building overlays or finding items.

**ANTI-PATTERN WARNING:** Do NOT use `querysvc` to "find the latest version" of a node you just wrote. The write operation (`CreateEntity` or `MutateEntity`) *returns* the new head. Use that return value. `querysvc` is for searching, not for transactional consistency.

---

## 5. Schema & DSL Authoring

Node definitions live in `code/definitions/`. They are the source of truth for the system.

### 5.1 The Definition Anatomy
```go
// code/definitions/work_task.go

var Task = dsl.Define("work", "task", func(t *dsl.TypeBuilder) {
    t.Desc("Represents a unit of work to be performed.")

    // Primitives
    t.Field("work:title", dsl.Text, "The task title").Required()
    t.Field("work:priority", dsl.Int, "Priority level").Range(0, 100)

    // Enums (Lowercase snake_case MANDATORY)
    t.Field("work:status", dsl.Text, "Current status").
        Enum("todo", "in_progress", "done", "blocked")

    // References
    t.Field("work:queue_id", dsl.EntityRef, "Owning queue").Required()
    t.Field("work:result_ref", dsl.NodeRef, "Result snapshot").PastOnly()
})
```

### 5.2 Field Constraints & Types

| Constraint | Effect |
| :--- | :--- |
| **`.Required()`** | Field must be present and non-zero. Validated at write time. |
| **`.Immutable()`**| Field cannot be changed in future versions (enforced by validator). |
| **`.Enum(v...)`** | Restricts value to set. **MUST** use lowercase snake_case values. |
| **`.PastOnly()`** | For `NodeRef`. Enforces `target.Seq < source.Seq`. Guarantees a DAG. |
| **`.List()`** | Field is a slice (e.g., `[]string`). |
| **`.Map()`** | Field is a map (e.g., `map[string]any`). |
| **`.Range(min, max)`** | For numbers. Enforces value bounds. |
| **`.Pattern(regex)`** | For strings. Enforces regex match. |
| **`.RequiredIf(sib, val)`** | Conditional requirement based on sibling field value. |

### 5.3 Naming Conventions (Strict)
* **Node Types**: `owner/name` (e.g., `work/task`, `code/file`).
* **Fields**: `owner:field_name` (e.g., `work:status`, `code:path`).

---

## 6. fdm-gen and Generated Artifacts

The `fdm-gen` tool (`code/cmd/fdm-gen/`) is the bridge between your DSL definitions and the executable code.

### 6.1 The Generation Pipeline
1.  **Parse**: It parses `code/interfaces/` to extract manually defined constants.
2.  **Merge**: It loads the DSL definitions and merges them with existing constants.
3.  **Emit**: It generates strict Go code and NeuroScript bindings.

### 6.2 Generated Outputs (The "Why")

* **`interfaces/fdm_constants_gen.go`**
  * **What:** The **Single Source of Truth** for `NodeType` and `FieldName` constants.
  * **Use:** Use these constants (e.g., `ix.NodeTypeTask`) instead of raw strings.

* **`nodes/gen_<type>.go`**
  * **What:** Typed builders and mutators.
  * **Use:** `nodes.BuildTask(...)`, `nodes.MutateTask(...)`, `nodes.TaskStatusTodo`.

* **`noderegistry/gen_*.go`**
  * **What:** NeuroScript tools (`fdm.nodes.<domain>.<type>.create`) and Schema Registry functions.
  * **Use:** Used internally by the registry to serve tools to agents.

**Rule**: If you change a DSL definition file, you **MUST** run `go run ./code/cmd/fdm-gen`.

### 6.3 Troubleshooting Generation
* `missing required field`: You missed a `.Required()` field in the Params struct.
* `violates PastOnly`: Your `NodeRef` points to a node that doesn't exist yet or is newer than the source.
* `unknown field`: You are using `Strict` schema and tried to set a field not in the DSL.

---

## 7. Runtime Validation

The Graph Service enforces integrity at write time via `ValidateNode`. This logic is not bypassable.

* **Schema Checks**: Every field is matched against the DSL. Unknown fields are rejected (in strict mode). Missing required fields cause failure.
* **Self-Reference**: A node cannot point to its own NodeID.
* **Recursive Links**: A link node (`link/edge`) cannot point to another link node.
* **PastOnly**: Sequence invariants are verified. If a node tries to reference a "future" node (which is impossible in a DAG), it is rejected.

---

## 8. Metadata Rules

Metadata provides machine-readable context for files.

### 8.1 File-Level Metadata
* **Format**: `:: key: value` (Strict spacing: one space after `::`, one space after `:`).
* **Character Set**: Keys must match `[a-zA-Z0-9_.-]+`.

**For NeuroScript (`.ns`):**
* Must appear at the **absolute start** of the file.
* Must be followed by **exactly one blank line**.
* Must NOT be commented out.

**For Markdown (`.md`):**
* Must appear at the **absolute end** of the file (footer).
* Must be preceded by **exactly one blank line**.
* Nothing may follow it (except a newline).

### 8.2 Embedded NeuroData Blocks
Inside a fenced code block (e.g., within a composite markdown file), metadata must appear at the **top** of the block content, before any data items.

### 8.3 Capsule Metadata
Capsules (`schema: capsule`) have mandatory keys:
* `:: id`: Must start with `capsule/` and use lowercase alphanumerics.
* `:: version`: Must be an integer.
* `:: description`: Must be present.
* `:: serialization`: Must match file type (`ns` or `md`).

---

## 9. NeuroScript Graph Contract (The Entity-First Law)

In NeuroScript, we abstract away the Copy-On-Write mechanics. Scripts interact with "Entity Objects".

### 9.1 The Entity Object
An entity is represented as a Map containing:
* `id`: The stable `EntityID`.
* `_version`: The opaque `NodeID` (used for Optimistic Locking).
* `_type`: Informational type string.
* `fields`: The data payload.

### 9.2 Tool Contracts
* **Arguments**: Domain tools MUST define arguments as map-based parameters to eliminate positional errors.
* **Pass the Object**: Scripts MUST pass the full entity object (the map) to tools requiring an identity.
* **Return Values**: Query and mutation tools MUST return `NSEntity` objects (maps), not raw strings.

### 9.3 Entity Splitting (Forbidden)
* ❌ **BAD:** `set id = task["id"]; call tool.update({"id": id, ...})`
  * This discards the `_version` lock and forces a blind read.
* ✅ **GOOD:** `call tool.update({"entity": task, "updates": {...}})`
  * This passes the full object, enabling optimistic concurrency checks.

---

## 10. Services Framework

FDM uses a strict Service Oriented Architecture (SOA).

### 10.1 The Interface
All services must implement the `services.Service` interface:
* `ID() ServiceID`: Unique name.
* `Requires() []ServiceID`: Dependencies.
* `Start(env Env) error`: Initialization.
* `Stop(ctx) error`: Shutdown.

### 10.2 Lifecycle & Dependency Injection
* **No Globals**: Services must not access global state.
* **Injection**: Dependencies are acquired **only** inside `Start()`, using `services.Require[T](env.Registry(), Key)`.
* **Factory Rule**: `New()` should only allocate the struct. Do not fetch dependencies in `New()`.
* **Readiness**: Call `s.MarkReady()` only when fully initialized. The registry blocks dependent services until this happens.
* **Persistence**: Services **MUST** implement `Snapshotter`.
  * If stateful: Implement `Snapshot()`/`Restore()` (use `GraphCache` helper if possible).
  * If stateless: Explicitly return `nil, nil` in `Snapshot()` to confirm design intent.

### 10.3 The Service Key Law (Strict Typing)
The Registry enforces strict type safety on lookup keys to prevent dependency injection errors.

* **The Type:** Keys are defined as `services.Key[T]`, where `T` is the interface being requested (e.g., `Key[interfaces.Graph]`).
* **The Prohibition:** You **MUST NOT** pass raw strings (e.g., `"workq"`) to `Require` or `InjectService`. This will cause an immediate runtime panic ("unsupported key type").
* **The Conversion:** If you are calling a generic API that accepts `Key[any]` (like the untyped `Require` on the Registry interface), you **MUST** convert your specific key using `.AsAny()`.

### 10.4 The Local Cancel Law (Deterministic Shutdown)
To prevent "The Deadly Embrace" deadlock during system shutdown, services that manage long-running background tasks (e.g., pollers, reconcilers, or `GraphCache` helpers) must maintain strict control over their lifecycle.

**The Trap (Heffalump Trap):**
A service's `Stop()` method calls a blocking `Wait()` on a background helper. If that helper is using the global `env.Context()`, it will never receive a cancellation signal because the global context is only cancelled *after* the `Stop()` method of all services has completed.

**The Rules:**
1. **Local Contexts:** Services **MUST** create a local child context (e.g., `context.WithCancel(env.Context())`) for any background helper that requires an exit signal.
2. **Signal Before Wait:** In the `Stop()` method, the service **MUST** call the local `cancel()` function **BEFORE** calling any blocking `Wait()` or `wg.Wait()` on the helper.
3. **Hardened Helpers:** When using helpers like `GraphCache`, prefer internalized lifecycle methods (e.g., `cache.Stop()`) which handle the signal-and-wait sequence atomically.

**Correct Pattern:**
```go
func (s *Service) Start(env Env) error {
    // 1. Create a child context to control background tasks
    s.stopCtx, s.stopCancel = context.WithCancel(env.Context())
    
    // 2. Pass the LOCAL context to the helper
    s.cache.RunPeriodicReconciler(s.stopCtx, ...)
    return nil
}

func (s *Service) Stop(ctx context.Context) error {
    // 3. Turn the key FIRST
    s.stopCancel() 
    
    // 4. Wait for the door to close SECOND
    s.cache.Wait() 
    return nil
}
```

---

## 11. WorkQ Subsystem

The Work Queue (`workq`) is the asynchronous task engine.

### 11.1 Core Architecture
* **Storage**: WorkQ is graph-backed. `Tasks` (`task/task`), `Queues` (`workq/queue`), and `WorkOrders` (`task/work_order`) are standard nodes.
* **Startup**: On boot, the service performs a `Reindex` to scan `Heads` and build an in-memory priority heap of pending work.
* **Signaling**: The "Bird" component monitors the heap and emits `workq.available` events ("chirps") when work is ready.
* **Budgets**: Execution limits (`max_fuel`, `max_tokens`) are enforced by the service.

### 11.2 Lifecycle
1.  **Enqueue**: Producer calls `tool.fdm.workq.enqueue`. Creates a `Task` node.
2.  **Signal**: Bird emits `workq.available`.
3.  **Dequeue**: Worker agent calls `tool.fdm.workq.dequeue`. Service mints a `WorkOrder` node (the lease) and sets `Task.ActiveOrder`.
4.  **Execute**: Worker performs logic.
5.  **Settlement**: Worker calls `ack` (Success) or `nack` (Failure). Service updates Task status.

---

## 12. AEIOU & Host Loop

The `ask` statement is powered by the AEIOU v3 Host Loop.

### 12.1 The "Proxy" Architecture
* **Hook**: When `ask` is called, the interpreter delegates to `aeiouService.RunAskLoop`.
* **Controller**: The `aeiou` service (Go) runs the multi-turn loop, not the interpreter.
* **Validation**: When the LLM returns `ACTIONS`, the service parses and validates the AST *before* execution.
* **Execution**: The service uses `api.ExecuteSandboxedAST` to run the validated code in a **sandboxed fork** of the interpreter.
* **Providers**: LLM providers are injected into the service via the host.

### 12.2 Safety Guards
* **Termination**: Loop ends on `<<<LOOP:DONE>>>` signal or max turns (capped at 25).
* **Progress Guard**: The loop aborts if the digest of outputs/whispers repeats (stuck loop detection).

---

## 13. Actor Identity

Security depends on knowing *who* is acting.

* **Interface**: `interfaces.Actor` exposes `DID()` (Decentralized Identifier).
* **Runtime Binding**: The `IDInterpreter` binds an Actor to a runtime instance.
* **Discovery**: In a tool implementation, use `rt.Actor()` to retrieve the current caller.
* **Enforcement**: Tools marked `RequiresID: true` will automatically fail if called from a runtime with no bound actor (e.g., a generic console script).

---

## 14. Forum Invariants

The Forum domain has specific structural rules.

* **Snapshots**: Replies (`parent_id`) and Votes (`msg`) MUST target a specific `NodeID` (Snapshot) using `NodeRef`. History is immutable; you vote on *what was written*, not what it might become.
* **Live Links**: Ownership (`author`), Projects (`project_id`), and Containers (`thread_id`) MUST target an `EntityID` (Live) using `EntityRef`. Structure follows the head.
* **Head Calculation**: The `forumsvc` is responsible for cryptographically validating the `ForumHead` node, which summarizes the thread's state.

---

## 15. Consolidated Forbidden Patterns (The "Anti-Patterns")

1.  **Ghost Nodes**: Using `CreateNode` to create a new entity.
    * **Fix**: Use `CreateEntity`.
2.  **History Trap**: Scanning `mem` to find current state.
    * **Fix**: Use `ReadHead` or `GetHeads`.
3.  **Handle-as-ID**: Storing handles in `_id` fields or map keys.
    * **Fix**: Use `NodeID` or `EntityID` + Handle Service for lookup.
4.  **Raw String IDs**: Passing `string` to Go internal APIs where an ID type exists.
    * **Fix**: Cast to `ix.EntityID` or `ix.NodeID`.
5.  **Entity Splitting**: Extracting IDs in NeuroScript and passing strings to tools.
    * **Fix**: Pass the full Entity Object (map).
6.  **Schema Errors**: Defining `handle` in DSL; Mixed-case enum values.
    * **Fix**: Remove `handle` field; use lowercase enums.
7.  **Query-for-Head**: Using `QueryService` to find the node you just wrote.
    * **Fix**: The write operation returns the head. Use it.
8.  **Manual Map Construction**: Building nodes without generated builders.
    * **Fix**: Use `nodes.Build<Type>(params).`
9.  **Silent Failure**: Returning `nil`/`0` on error without logging.
    * **Fix**: Log errors to `Stderr` if returning a default value is necessary.
10. **Orphaned Anchors**: Anchoring derived data (contracts) to `EntityID`.
    * **Fix**: Anchor to the specific source `NodeID` (Snapshot).
11. **Premature Event (The "Ghost Listener")**: Publishing events to a bus before the target Agent is running.
    * **Fix**: In tests/setup, call `manager.StartAgent()` and **wait** for the `AgentRegistered` event or poll `manager.GetAgent()` until it returns `Running`.
12. **Async Data Race (The "Dirty Buffer")**: Using `bytes.Buffer` to capture output from an Agent Runtime (which runs in a goroutine).
    * **Fix**: Use a thread-safe implementation like `LockedBuffer` or `io.Pipe`.

---

## 16. Testing & Concurrency

### 16.1 The "Wait for Readiness" Rule
When writing integration tests involving agents or async services:
* **Never** assume an agent is ready immediately after `StartAgent()` returns.
* **Always** gate subsequent actions (like `bus.Publish`) behind a readiness check.
* **Use**: `waitForAgent(t, manager, did)` helper loop.

### 16.2 Lock Hygiene
* **Zero Inference**: Never read shared state outside of a lock.
* **Reserve & Release**: For long-running operations (like I/O or verification), acquire the lock to mark the operation as "in-progress", release it to do the work, and re-acquire it to commit the result.

## 17. NeuroScript Host Integration Invariants

Integrating the NeuroScript interpreter into FDM services requires strict adherence to I/O and lifecycle rules to prevent data corruption and "Zombie" state.

### 17.1 The Output Precedence Law
The `HostContext` provides two distinct output pipelines. They MUST NOT be confused.

* **The Semantic Channel (`EmitFunc`)**: Used exclusively by the `emit` keyword in scripts. This is the structured path for programmatic results. If provided, it consumes all emissions and the interpreter will **NOT** write them to `Stdout`.
* **The Streamed Channel (`Stdout`)**: Used by `tool.io.print`, `Runtime.Println`, or as a fallback for `emit` if `EmitFunc` is nil. This is for unstructured human-readable logs and traces.
* **The Violation**: Expecting programmatic results to appear in a buffer captured via `WithStdout` when an `EmitFunc` is also defined.
* **The Consequence**: Timeouts in integration tests as the result buffer remains empty while the data is routed to the structured callback.

### 17.2 The Dynamic Binding Law (Anti-Zombie Logger)
The `HostContext` is **immutable** once passed to `api.New()`. The interpreter captures the specific function or object pointer provided at creation.

* **The Violation (Snapshotting)**: Creating a `HostContext` callback that closes over a local variable or a specific logger instance (e.g., `NewNSLoggerAdapter(core.Log.With(...))`).
* **The Consequence**: If the host later redirects its global logger (common in integration tests), the interpreter remains bound to the original "Zombie Logger," making it immune to test redirection and causing assertion failures.
* **Fix**: Ensure all `HostContext` callbacks perform a **dynamic lookup** at call-time rather than snapshotting state at boot.

### 17.3 HostContext Thread-Safety (Dirty Buffer Protection)
Because `HostContext` is shared by reference among all forked interpreters, every component provided to it MUST be thread-safe.

* **The Violation**: Passing a raw `bytes.Buffer` to `WithStdout`, `WithStderr`, or utilizing non-thread-safe slices/maps within an `EmitFunc` closure.
* **The Consequence**: Data races and garbled output when multiple event handlers or asynchronous tools execute in parallel.
* **Fix**: Use a `LockedBuffer` or mutex-protected collectors for all host-provided I/O.


:: id: capsule/coding_guide
:: version: 2.4
:: description: FDM Coding Guide & Architectural Invariants. Added Concurrency & Testing Anti-Patterns (Premature Event, Dirty Buffer).
:: serialization: md
:: filename: coding_guide.md