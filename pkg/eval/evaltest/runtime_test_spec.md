RuntimeTestSpec.md
This document specifies the contract for a valid eval.Runtime implementation.

1. GetVariable(name string) (lang.Value, bool)
The variable store must be correct, type-safe, and isolated.

Test 1.1: Get Existing Variable

Action: Pre-load the runtime with {"foo": lang.StringValue{"bar"}}.

Test: Call rt.GetVariable("foo").

Assert: Returns (lang.StringValue{"bar"}, true).

Test 1.2: Get Non-Existent Variable

Action: Call rt.GetVariable("non_existent_var").

Assert: Returns (any, false). The returned value can be anything, but the boolean must be false.

Test 1.3: Contract: No Raw Types (Variables)

Action: Pre-load the runtime with a "raw" Go type (e.g., map[string]any{"my_raw_map": "raw_string"}). This setup is invalid and should be caught by the host's "set variable" logic if possible.

Test: Call rt.GetVariable("my_raw_map").

Assert: The call must never return ("raw_string", true). It must either return a correctly wrapped lang.Value (e.g., lang.MapValue{...}) or (any, false) if the invalid set was rejected.

Test 1.4: Collection Value-Type Correctness

Action: Pre-load the runtime with {"my_map": lang.MapValue{Value: ...}} (a value type).

Test: Call rt.GetVariable("my_map").

Assert: Returns (lang.MapValue{...}, true). It must not incorrectly return a pointer type.

2. GetToolSpec(toolName types.FullName) (eval.ToolSpec, bool)
The runtime must accurately report the specifications of available tools.

Test 2.1: Get Existing Spec

Action: Pre-load a tool with a known spec (e.g., tool.test.add(a, b)).

Test: Call rt.GetToolSpec("tool.test.add").

Assert: Returns (spec, true), where spec accurately reflects the arguments a and b, including their Required status.

Test 2.2: Get Non-Existent Spec

Action: Call rt.GetToolSpec("tool.fake.nonexistent").

Assert: Returns (eval.ToolSpec{}, false).

3. ExecuteTool(toolName types.FullName, args map[string]lang.Value) (lang.Value, error)
The runtime must execute tools safely and adhere to the lang.Value wrapping contract.

Test 3.1: Execute Valid Tool

Action: Pre-load a tool tool.test.add that adds two numbers.

Test: Call rt.ExecuteTool("tool.test.add", {"a": lang.NumberValue{10}, "b": lang.NumberValue{5}}).

Assert: Returns (lang.NumberValue{15}, nil).

Test 3.2: Execute Non-Existent Tool

Action: Call rt.ExecuteTool("tool.fake.nonexistent", nil).

Assert: Returns (nil, err) where errors.Is(err, lang.ErrToolNotFound).

Test 3.3: Contract: No Raw Return Values

Action: Pre-load a tool tool.test.get_raw_map that returns a raw map[string]any.

Test: Call rt.ExecuteTool("tool.test.get_raw_map", nil).

Assert: The runtime must intercept the raw value and wrap it. Returns (lang.MapValue{...}, nil). It must never return (map[string]any{...}, nil).

Test 3.4: Contract: Wrap Tool Error

Action: Pre-load a tool tool.test.get_error that returns (nil, errors.New("tool_panic")).

Test: Call rt.ExecuteTool("tool.test.get_error", nil).

Assert: Returns (nil, err) where err is a *lang.RuntimeError that wraps the original "tool_panic" error.

Test 3.5: Contract: Recover from Tool Panic

Action: Pre-load a tool tool.test.get_panic that calls panic("oh no").

Test: Call rt.ExecuteTool("tool.test.get_panic", nil).

Assert: The runtime must recover from the panic. Returns (nil, err) where err is a *lang.RuntimeError indicating a panic occurred.

4. RunProcedure(procName string, args ...lang.Value) (lang.Value, error)
The runtime must be able to manage and call user-defined procedures.

Test 4.1: Execute Valid Procedure

Action: Pre-load a user-defined procedure my_proc(a) that returns a + 1.

Test: Call rt.RunProcedure("my_proc", lang.NumberValue{10}).

Assert: Returns (lang.NumberValue{11}, nil).

Test 4.2: Execute Non-Existent Procedure

Action: Call rt.RunProcedure("fake_proc").

Assert: Returns (nil, err) where errors.Is(err, lang.ErrProcedureNotFound).

Test 4.3: Argument Mismatch

Action: Pre-load my_proc(a) (arity 1).

Test: Call rt.RunProcedure("my_proc", lang.NumberValue{10}, lang.NumberValue{11}).

Assert: Returns (nil, err) where errors.Is(err, lang.ErrArgumentMismatch).