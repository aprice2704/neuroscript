# Guide: Building Execution Policies

**Audience:** FDM and NeuroScript host application developers.
**Purpose:** Explains how to use the fluent builder API to create security policies for the NeuroScript interpreter.

---

## 1. Overview & Core Concepts

Every NeuroScript interpreter is governed by an **Execution Policy**. This policy is a security-critical component that defines what a script is allowed to do. It is composed of three main parts:

1.  **Allow/Deny Lists:** A simple mechanism to permit or forbid the execution of tools based on their names (e.g., `tool.fs.read`).
2.  **Capabilities:** A fine-grained permission model that grants specific rights to resources (e.g., granting read access to a specific file path).
3.  **Resource Limits:** Quantitative limits on operations to prevent abuse (e.g., setting a maximum number of network calls).

The `api` package provides a fluent builder, `api.NewPolicyBuilder`, to construct these policies.

### Core Concepts

* **Resource:** A category of system asset a script might access. Examples: `fs` (filesystem), `net` (network), `env` (environment variables).
* **Verb:** An action that can be performed on a Resource. Examples: `read`, `write`, `use`, `admin`.
* **Scope:** A specific instance of a resource the verb applies to. Examples: a file path for `fs`, a hostname for `net`, or an environment variable key for `env`.
* **Capability:** The combination of a **Resource**, one or more **Verbs**, and one or more **Scopes** that defines a single, granular permission.

---

## 2. Allow/Deny Lists & The "Deny-by-Default" Principle

The simplest part of a policy is the list of tools a script is allowed or forbidden to run. **Deny rules always override allow rules**.

**Crucially, all policies created with the builder are "deny-by-default."** This means that a tool is only allowed to run if it explicitly matches a pattern in the `.Allow()` list. A new, unmodified policy will deny all tool calls.

Both lists support wildcard matching for tool names.

| Pattern | Example | Description |
| :--- | :--- | :--- |
| Exact | `tool.fs.read` | Matches the specific tool name (case-insensitive). |
| Prefix | `tool.fs.*` | Matches all tools in the `fs` group. |
| Suffix | `*.delete` | Matches any tool ending in `delete`. |
| Substring | `*agent*` | Matches any tool containing `agent`. |
| Universal | `*` | Matches all tools. |

---

## 3. Capabilities In-Depth

Capabilities provide fine-grained control over a script's permissions. A script can only perform a privileged action if one of its tool's `RequiredCaps` is satisfied by a `Capability` granted in the policy.

### Standard Resources & Verbs

The policy system defines a standard set of resources and verbs:

| Resource | Verbs | Description |
| :--- | :--- | :--- |
| `fs` | `read`, `write` | Filesystem access. |
| `net` | `read`, `write` | Network access. |
| `env` | `read` | Read access to environment variables. |
| `secret` | `use` | Access to decrypted secrets. |
| `model` | `use`, `admin` | Use of or administrative access to AgentModels. |
| `tool` | `exec` | General permission to execute tools (rarely used). |
| `budget`| `use` | Permission related to spending limits. |
| `bus` | `read`, `write`| Permissions for the internal event bus. |


### Scope Matching Rules

The power of capabilities lies in how a **granted scope** is matched against a **required scope**. The rules vary by resource type:

| Resource(s) | Grant Scope Syntax | Description & Examples |
| :--- | :--- | :--- |
| `env`, `secret`, `model`, `sandbox`, `proc` | Simple Wildcards | Uses the same wildcard patterns as Allow/Deny lists. **Example:** A grant with scope `stripe_*` will satisfy a need for scope `STRIPE_API_KEY`. |
| `fs` | Glob Pattern | The grant is a standard filesystem glob pattern. **Example:** A grant with scope `/data/*.log` will satisfy a need for `/data/app.log` but not `/data/config/app.log`. |
| `net` | Hostname Wildcards | Matches a hostname and optional port. The pattern `*.example.com` is special: it matches the base domain (`example.com`) and any subdomain (`api.example.com`). Other wildcards follow the "Simple Wildcard" rules. If ports are specified in both grant and need, they must match exactly. |
| `clock`, `rand`, `budget` | Exact or `*` | The scope must be an exact match (e.g., `seed:123`), `"true"`, or the universal wildcard `*`. |

---

## 4. Resource Limits

To prevent runaway scripts from consuming excessive resources, you can set quantitative limits. If a limit is exceeded during execution, the script will halt with a policy error.

| Limit | Builder Method | Description |
| :--- | :--- | :--- |
| Max Tool Calls | `.LimitToolCalls(name, max)` | Maximum number of times a specific tool can be called. |
| Network Limits | `.LimitNet(maxCalls, maxBytes)` | Total number of network calls and total bytes transferred. |
| Filesystem Limits | `.LimitFS(maxCalls, maxBytes)` | Total number of filesystem calls and total bytes transferred. |
| Budget (per-call) | `.LimitPerCallCents(curr, cents)`| Maximum cost for a single operation (e.g., an `ask`). |
| Budget (per-run) | `.LimitPerRunCents(curr, cents)`| Maximum total cost for the entire script execution. |

---

## 5. Examples

### Secure, Sandboxed Policy (Default)

This is the **recommended default** for running untrusted scripts. Since the builder is deny-by-default, this policy allows only the `tool.strtools.Concat` tool and nothing else.

```go
// Create a policy for a normal, untrusted execution context.
securePolicy := api.NewPolicyBuilder(api.ContextNormal).
    Allow("tool.strtools.Concat"). // Allow only a single, safe tool.
    Build()

interp := api.New(api.WithExecPolicy(securePolicy))
```

### Trusted Configuration Policy

This policy is for trusted setup scripts that need privileged operations. It runs in a special context, uses a specific, minimal `Allow` list, and grants only the precise capabilities needed.

```go
// Create a policy for a trusted 'config' context.
trustedPolicy := api.NewPolicyBuilder(api.ContextConfig).
    Allow( // Explicitly list the ONLY tools this script is allowed to run.
        "tool.account.Register",
        "tool.agentmodel.Register",
        "tool.os.Getenv",
    ).
    // Grant capabilities using the convenient string parser.
    Grant("account:admin:*").
    Grant("model:admin:*").
    Grant("env:read:OPENAI_API_KEY,ANTHROPIC_API_KEY").
    // Set quantitative limits.
    LimitToolCalls("tool.account.Register", 10).
    LimitNet(50, 1024*1024). // 50 calls, 1MB total
    Build()

interp := api.New(api.WithExecPolicy(trustedPolicy))
```

---

**End of file**