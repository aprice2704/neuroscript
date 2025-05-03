// NeuroScript Version: 0.3.0
// Last Modified: 2025-05-03 15:55:00 PM PDT // Use SLogAdapter for actual log output in tests
// filename: pkg/neurodata/checklist/test_helpers.go
package checklist

import (
	"fmt"
	"log/slog" // <<< Added import
	"os"       // <<< Added import
	"testing"

	// Import necessary packages WITHOUT causing cycles
	"github.com/aprice2704/neuroscript/pkg/adapters"
	"github.com/aprice2704/neuroscript/pkg/core" // <<< Added import for LevelDebug
	"github.com/aprice2704/neuroscript/pkg/toolsets"
)

// newTestInterpreterWithAllTools creates a new interpreter instance for checklist testing,
// initializing it with BOTH core AND extended tools, using a functional logger.
func newTestInterpreterWithAllTools(t *testing.T) (*core.Interpreter, *core.ToolRegistry) {
	t.Helper() // Mark this as a test helper

	tempDir := t.TempDir()

	// --- Logger Setup ---
	// <<< FIX: Create a real SLog logger that outputs Debug messages to Stderr >>>
	slogHandler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level:     slog.LevelDebug, // Ensure Debug messages are logged
		AddSource: false,           // Optional: Add source file/line to logs
	})
	slogger := slog.New(slogHandler)
	// Create the adapter implementing our logging.Logger interface
	logger, err := adapters.NewSlogAdapter(slogger)
	if err != nil {
		t.Fatalf("Setup Error: Failed to create SLogAdapter: %v", err)
	}
	// --- End Logger Setup ---

	llmClient := adapters.NewNoOpLLMClient() // Assuming constructor doesn't require logger

	// Initialize the ToolRegistry from core
	registry := core.NewToolRegistry()

	// Register CORE tools first
	err = core.RegisterCoreTools(registry)
	assertNoErrorSetup(t, err, "Failed to register core tools") // Use local helper

	// Register EXTENDED tools (Checklist, Blocks, etc.) via toolsets
	err = toolsets.RegisterExtendedTools(registry)
	assertNoErrorSetup(t, err, "Failed to register extended toolsets") // Use local helper

	// Create the core interpreter instance, passing the REAL logger
	interp := core.NewInterpreter(logger, llmClient) // <<< Passes the SLogAdapter

	// Inject the registry containing ALL tools
	interp.SetToolRegistry(registry) // <<< ADDED CALL

	// Create/Set FileAPI and Sandbox Dir
	interp.SetSandboxDir(tempDir)

	return interp, registry
}

// Helper function to get a node's value from the map returned by getNodeViaTool
// (Implementation unchanged)
func getNodeValue(t *testing.T, nodeData map[string]interface{}) interface{} {
	t.Helper()
	if nodeData == nil {
		t.Fatalf("getNodeValue: called with nil nodeData")
	}
	val, ok := nodeData["value"]
	if !ok {
		return nil
	}
	return val
}

// Helper function to get node attributes from the map returned by getNodeViaTool
// (Implementation unchanged)
func getNodeAttributesMap(t *testing.T, nodeData map[string]interface{}) map[string]string {
	t.Helper()
	if nodeData == nil {
		t.Fatalf("getNodeAttributesMap: called with nil nodeData")
	}
	attrsVal, exists := nodeData["attributes"]
	if !exists {
		return make(map[string]string)
	}
	if attrsVal == nil {
		return make(map[string]string)
	}
	rawAttrsMap, ok := attrsVal.(map[string]interface{})
	if !ok {
		t.Fatalf("getNodeAttributesMap: 'attributes' field is not a map[string]interface{}: %T", attrsVal)
	}
	stringAttrsMap := make(map[string]string)
	for k, v := range rawAttrsMap {
		if vStr, ok := v.(string); ok {
			stringAttrsMap[k] = vStr
		} else {
			stringAttrsMap[k] = fmt.Sprintf("%v", v)
		}
	}
	return stringAttrsMap
}

// Helper to fail test immediately if error occurs during setup
func assertNoErrorSetup(t *testing.T, err error, msgFormat string, args ...interface{}) {
	t.Helper()
	if err != nil {
		message := fmt.Sprintf(msgFormat, args...)
		t.Fatalf("Setup Error: %s: %v", message, err)
	}
}

// --- ADDED: Helper to check if a tool was found in the registry ---
// assertToolFound fails the test if the tool was not found (found is false).
func assertToolFound(t *testing.T, found bool, toolName string) {
	t.Helper()
	if !found {
		t.Fatalf("Setup Error: Required tool '%s' not found in registry", toolName)
	}
}

// --- ADDED: Local pstr helper ---
// pstr returns a pointer to the given string. Useful for optional string args.
func pstr(s string) *string {
	return &s
}

// pbool returns a pointer to the given boolean.
func pbool(b bool) *bool {
	return &b
}

// pint returns a pointer to the given integer.
func pint(i int) *int {
	return &i
}

// --- End Pointer Helpers ---
