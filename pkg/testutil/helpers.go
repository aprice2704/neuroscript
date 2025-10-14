// NeuroScript Version: 0.8.0
// File version: 2
// Purpose: Updated the test interpreter helper to use the new public api.New() constructor and HostContext, resolving build failures.
// filename: pkg/testutil/helpers.go
// nlines: 40
// risk_rating: MEDIUM

package testutil

import (
	"os"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/api"
	"github.com/aprice2704/neuroscript/pkg/logging"
	_ "github.com/aprice2704/neuroscript/pkg/toolbundles/all" // Ensures all tools are linked
	"github.com/aprice2704/neuroscript/pkg/types"
)

// NewTestInterpreterWithAllTools creates a new interpreter instance for testing using the public API
// and verifies that the tool registry has been fully populated.
func NewTestInterpreterWithAllTools(t *testing.T) *api.Interpreter {
	t.Helper()

	// The new API requires a HostContext. We'll create a minimal one.
	hc, err := api.NewHostContextBuilder().
		WithLogger(logging.NewTestLogger(t)).
		WithStdout(os.Stdout).
		WithStderr(os.Stderr).
		WithStdin(os.Stdin).
		Build()
	if err != nil {
		t.Fatalf("Failed to build HostContext for test interpreter: %v", err)
	}

	// Create the interpreter using the new public API.
	// Standard tools are registered by default.
	interp := api.New(api.WithHostContext(hc))

	if interp.ToolRegistry() == nil {
		t.Fatal("FATAL: New() returned an interpreter with a nil tool registry.")
	}

	// Verification step: Check for a known extended tool.
	expectedTool := types.FullName("tool.meta.listtools")
	if _, found := interp.ToolRegistry().GetTool(expectedTool); !found {
		var availableTools []string
		for _, spec := range interp.ToolRegistry().ListTools() {
			availableTools = append(availableTools, string(spec.Name()))
		}
		t.Fatalf("FATAL: Test interpreter's tool registry is incomplete. Expected to find '%s', but it was missing. "+
			"Available tools (%d): %s",
			expectedTool, len(availableTools), strings.Join(availableTools, ", "))
	}

	return interp
}
