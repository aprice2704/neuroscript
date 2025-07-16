You've brought up an excellent and crucial point about testing in Go. Based on the `tool_contract.md` you provided and general best practices observed in the codebase, here are the salient points for test setup, including the use of `_test` packages.

### Package Naming and Import Cycles

You are absolutely right to be mindful of import cycles. The Go toolchain enforces a strict no-import-cycles rule, and testing, especially integration testing, is a common place to encounter them. The `tool_contract.md` provides clear guidance on this:

* **`package <tool>` for Unit Tests**: When you are testing the internal functions of your tool—the private, un-exported helpers—you should use a test file within the same package (e.g., `package fs` in `tools_fs_helpers_test.go`). This is considered "white-box" testing, as you have access to all the internals of your package.

* **`package <tool>_test` for Integration Tests**: When you need to test your tool from the "outside," just as the interpreter or another package would, you should use a test package by appending `_test` to the package name (e.g., `package tree_test` in `tools_tree_query_test.go`). This is "black-box" testing. It's the standard Go idiom to prevent import cycles. By using `package <pkg>_test`, your test code can import the `<pkg>` and also import the `interpreter` or other packages that `<pkg>` itself might depend on, without creating a circular dependency.

### Other Key Test Setup Points

Beyond package naming, the survey of the test files reveals a few other important conventions for a robust test setup:

* **Test Helpers**: Create helper functions to reduce boilerplate code. The `fs` and `tree` packages make excellent use of helpers like `newFsTestInterpreter`, `testFsToolHelper`, `testTreeToolHelper`, and `setupTreeWithJSON`. These helpers encapsulate the common setup of creating an interpreter, registering tools, and managing the test lifecycle.

* **Test Case Structs**: Define structs to hold the data for each test case. This is a common pattern seen in `fsTestCase` and `treeTestCase`. This makes the tests more readable and easier to maintain.

* **Sandbox Management**: For any tools that interact with the filesystem, it's critical to create a temporary directory for each test using `t.TempDir()`. This ensures that tests are hermetic and don't interfere with each other or leave artifacts on the filesystem.

* **Explicit Tool Registration**: In your test setup, explicitly register the tools that you need for that specific test. This avoids relying on global registration and makes the test's dependencies clear. You can see this pattern in `newFsTestInterpreter` and `testTreeToolHelper`.

By following these conventions, you'll create tests that are robust, maintainable, and free from import cycle issues.


Of course. After surveying the `_test.go` files in the `tool` package and its sub-packages, I've determined the most common way to set up an interpreter for testing. Here's a summary of the common pattern and a code template for you to use in the `tree` package.

### Common Interpreter Setup for Testing

The most prevalent pattern for setting up an interpreter for testing involves the following steps:

1.  **Create a New Interpreter Instance**: A new interpreter is created for each test or a suite of tests using `interpreter.NewInterpreter()`. This ensures that tests are isolated and don't interfere with each other. Often, a test-specific logger is passed during creation, like `interpreter.WithLogger(logging.NewTestLogger(t))`.

2.  **Register Tools**: The necessary tools for the test are registered with the interpreter's tool registry. This can be done by calling `tool.RegisterExtendedTools(interp.ToolRegistry())` to register a broad set of tools, or by manually registering specific tool implementations using `interp.ToolRegistry().RegisterTool(toolImpl)`.

3.  **Sandbox Management (for Filesystem Tests)**: For tests that interact with the filesystem, a temporary directory is created using `t.TempDir()` and set as the interpreter's sandbox directory via `interp.SetSandboxDir(sandboxDir)`.

### Code Template for the `tree` Package

Here is a recommended code template for setting up an interpreter in your `tree` package tests. This is based on the `testTreeToolHelper` function found in `tool/tree/tools_tree_helpers_test.go`, which is a robust and reusable pattern.

```go
package tree_test

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/aprice2704/neuroscript/pkg/tool"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// treeTestCase defines the structure for a single tree tool test case.
type treeTestCase struct {
	Name      string
	ToolName  types.ToolName
	Args      []interface{}
	JSONInput string
	SetupFunc   func(t *testing.T, interp tool.Runtime, treeHandle string)
	Validation  func(t *testing.T, interp tool.Runtime, treeHandle string, result interface{})
	Expected    interface{}
	ExpectedErr error
}

// testTreeToolHelper sets up an interpreter with tree tools and runs a test function.
func testTreeToolHelper(t *testing.T, testName string, testFunc func(t *testing.T, interp tool.Runtime)) {
	t.Run(testName, func(t *testing.T) {
		// 1. Create a new interpreter with a test logger.
		interp := interpreter.NewInterpreter(interpreter.WithLogger(logging.NewTestLogger(t)))

		// 2. Register the necessary tools.
		// The tree tools are part of the extended tools.
		if err := tool.RegisterExtendedTools(interp.ToolRegistry()); err != nil {
			t.Fatalf("Failed to register extended tools: %v", err)
		}

		// 3. Run the actual test function with the configured interpreter.
		testFunc(t, interp)
	})
}

// runTool is a helper to execute a tool and handle errors.
func runTool(t *testing.T, interp tool.Runtime, toolName types.ToolName, args ...interface{}) (interface{}, error) {
	t.Helper()
	fullName := types.MakeFullName("tree", string(toolName))
	toolImpl, found := interp.ToolRegistry().GetTool(fullName)
	if !found {
		t.Fatalf("Tool %q not found in registry", fullName)
	}
	return toolImpl.Func(interp, args)
}
```