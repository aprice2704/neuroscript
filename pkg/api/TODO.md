# API todos

## 1. Expand preflight check to policy

That's a fantastic idea, and it directly addresses the root cause of this entire debugging nightmare.

You're completely right. The silent fails are happening because the `AgentRuntime` successfully loads the script (the tools *are* registered), but the handler *panics at runtime* because the **policy** denies the tool call.

Your suggestion is the correct architectural fix:

The "preflight" check, `api.CheckScriptTools`, should be enhanced. Right now, it only checks if the tools in the script *exist* in the interpreter's tool registry.

We need to upgrade it to *also* check if the tools are *allowed* by the interpreter's `ExecPolicy`.

The `AgentRuntime.LoadUnit` function (which calls `api.CheckScriptTools`) would then look something like this:

1.  `api.LoadFromUnit(interp, unit)` (Loads the script AST)
2.  `policy = interp.ExecPolicy()` (Gets the active policy)
3.  `api.CheckScriptTools(tree, interp)` (The existing check for registration)
4.  **`api.CheckScriptPolicy(tree, policy)` (The new check you're proposing)**

If this new `CheckScriptPolicy` function iterates the AST and finds a `tool.str.inspect` call that isn't allowed by the policy, `LoadUnit` would fail immediately with a clear error:

`ERROR: AgentRuntime.LoadUnit failed: script check failed: tool.str.inspect not allowed by policy`

This would have caught our bug instantly and prevented the agent from ever starting, which is infinitely better than it starting "deaf" and failing silently.


## Improve API usability

1. The current words used seem to confuse everyone, we should rationalize them. 

e.g."You're right, I'm still getting the API usage wrong. My apologies.

The test failure procedure 'run_ask' not found proves that my previous assumption was incorrect. Calling api.ExecWithInterpreter does not permanently load definitions into the interpreter's state for the next call. It appears to be a one-shot execution.

The correct pattern, which is used in daemon_handlers.go and interpreter_persistence_test.go, is to use interp.AppendScript() to load the code (both definitions and commands) and then call interp.ExecuteCommands() to run the appended command blocks.

I have corrected cmd/zadeh/wiring_provider_test.go to use this proper AppendScript -> ExecuteCommands pattern."

