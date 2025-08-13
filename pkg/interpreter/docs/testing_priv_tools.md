# Testing Privileged and Policy-Restricted Tools

All tools are loaded into the interpreter's registry at startup, but their use is governed by a runtime **Execution Policy**. Privileged tools (e.g., `tool.agentmodel.Register`) are blocked by default in the `normal` execution context.

To test a privileged tool, you must create an interpreter and provide it with a specific policy that grants the necessary permissions for your test.

### Steps for Writing a Privileged Tool Test

1.  **Create a Standard Interpreter**: In your test, instantiate a normal interpreter. It will automatically load all tools.

    ```go
    import "[github.com/aprice2704/neuroscript/pkg/interpreter](https://github.com/aprice2704/neuroscript/pkg/interpreter)"

    // In your test file...
    interp := interpreter.NewInterpreter() // All tools are loaded by default.
    ```

2.  **Define a Permissive `ExecPolicy` for the Test**: Create an `ExecPolicy` that sets the context to `config` (which allows trusted tools) and explicitly allows and grants capabilities for the tool you are testing.

    ```go
    import (
        "[github.com/aprice2704/neuroscript/pkg/runtime](https://github.com/aprice2704/neuroscript/pkg/runtime)"
        "[github.com/aprice2704/neuroscript/pkg/policy/capability](https://github.com/aprice2704/neuroscript/pkg/policy/capability)"
    )

    // Define a policy that allows the agentmodel admin tools to run.
    testPolicy := &runtime.ExecPolicy{
        Context: runtime.ContextConfig, // Use 'config' to enable trusted tools.
        Allow:   []string{"tool.agentmodel.*"}, // Allow the toolset.
        Grants: capability.NewGrantSet(
            []capability.Capability{
                // Grant the specific capability the tool requires.
                {Resource: "model", Verbs: []string{"admin"}, Scopes: []string{"*"}},
            },
            capability.Limits{}, // No specific limits for this test.
        ),
    }
    ```

3.  **Apply the Policy to the Interpreter**: Use the new `WithExecPolicy` option when creating the interpreter to apply your test-specific rules.

    ```go
    interp := interpreter.NewInterpreter(interpreter.WithExecPolicy(testPolicy))
    ```

4.  **Execute and Verify**: With the permissive policy in place, your test can now run scripts or call tools that would otherwise be blocked.

    ```go
    // This script will now succeed because the policy allows it.
    script := `must tool.agentmodel.Register("my_agent", {"provider":"p", "model":"m"})`
    _, err := interp.ExecuteScriptString("my_test", script, nil)
    if err != nil {
        t.Fatalf("Script execution failed: %v", err)
    }
    ```

This approach accurately simulates how the host application will configure the interpreter for different tasks (e.g., a startup script running in a `config` context vs. a user script running in a `normal` context) and ensures your tools are tested under realistic policy conditions.