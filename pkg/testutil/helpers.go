// NeuroScript Version: 0.6.0
// File version: 1.0.1
// Purpose: Provides a centralized helper for creating a fully-initialized interpreter for use in tests across multiple packages. FIX: Removed WithOutStandardTools option to allow the interpreter to correctly initialize with the globally registered tools.
// filename: pkg/testutil/helpers.go
// nlines: 35
// risk_rating: LOW

package testutil

import (
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/interpreter"
	_ "github.com/aprice2704/neuroscript/pkg/toolbundles/all" // Ensures all tools are linked
	"github.com/aprice2704/neuroscript/pkg/types"
)

// NewTestInterpreterWithAllTools creates a new interpreter instance for testing and
// critically verifies that the tool registry has been fully populated. If a known
// extended tool is missing, it fails the test immediately.
func NewTestInterpreterWithAllTools(t *testing.T) *interpreter.Interpreter {
	t.Helper()

	// FIX: The WithOutStandardTools() option was preventing the interpreter from
	// loading the globally registered tools. Removing it allows the default
	// constructor behavior, which is to populate the registry.
	interp := interpreter.NewInterpreter()

	if interp.ToolRegistry() == nil {
		t.Fatal("FATAL: NewInterpreter() returned an interpreter with a nil tool registry.")
	}

	// Verification step: Check for a known extended tool to ensure the
	// toolbundles were correctly linked and initialized.
	expectedTool := types.FullName("tool.meta.listtools")
	if _, found := interp.ToolRegistry().GetTool(expectedTool); !found {
		var availableTools []string
		for _, spec := range interp.ToolRegistry().ListTools() {
			availableTools = append(availableTools, string(spec.Name()))
		}
		t.Fatalf("FATAL: Test interpreter's tool registry is incomplete. Expected to find '%s', but it was missing. "+
			"This usually indicates a problem with test build linkage. Available tools (%d): %s",
			expectedTool, len(availableTools), strings.Join(availableTools, ", "))
	}

	return interp
}
