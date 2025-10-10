// NeuroScript Version: 0.8.0
// File version: 7
// Purpose: FIX: Accesses the internal interpreter to get the ToolRegistry for the test, aligning with the new facade pattern.
// filename: pkg/api/toolsmeta/export_test.go
// nlines: 65
// risk_rating: LOW

package toolsmeta_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/api"
	"github.com/aprice2704/neuroscript/pkg/api/toolsmeta"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

type internalInterpreter interface {
	ToolRegistry() tool.ToolRegistry
}

func TestExportTools(t *testing.T) {
	// --- Setup ---
	tempDir := t.TempDir()
	outputFile := filepath.Join(tempDir, "test-tools.json")

	// Create a new interpreter via the facade to get a populated tool registry.
	interp := api.New()
	internal := interp.InternalRuntime().(internalInterpreter)
	reg := internal.ToolRegistry()
	if reg == nil {
		t.Fatal("Tool registry from internal interpreter was nil")
	}

	// --- Execute ---
	err := toolsmeta.ExportTools(reg, outputFile)
	if err != nil {
		t.Fatalf("ExportTools() returned an unexpected error: %v", err)
	}

	// --- Verify File Exists ---
	info, err := os.Stat(outputFile)
	if os.IsNotExist(err) {
		t.Fatalf("Expected output file '%s' was not created.", outputFile)
	}
	if info.Size() == 0 {
		t.Fatalf("Expected output file '%s' to not be empty, but it was.", outputFile)
	}

	// --- Verify File Content ---
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file '%s': %v", outputFile, err)
	}

	var toolImpls []tool.ToolImplementation
	if err := json.Unmarshal(content, &toolImpls); err != nil {
		t.Fatalf("Failed to unmarshal JSON from output file: %v", err)
	}

	if len(toolImpls) == 0 {
		t.Fatal("Expected to find at least one tool implementation in the output, but found none.")
	}

	// Spot-check for a well-known tool.
	foundReadTool := false
	expectedFullName := "tool.fs.read"
	for _, impl := range toolImpls {
		actualFullName := api.MakeToolFullName(string(impl.Spec.Group), string(impl.Spec.Name))
		if strings.EqualFold(string(actualFullName), expectedFullName) {
			foundReadTool = true
			break
		}
	}

	if !foundReadTool {
		t.Errorf("Did not find the expected standard tool '%s' in the exported metadata.", expectedFullName)
	}
}
