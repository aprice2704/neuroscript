// NeuroScript Version: 0.6.0
// File version: 1
// Purpose: Provides integration tests for the toolsmeta export functionality.
// filename: pkg/api/toolsmeta/export_test.go
// nlines: 65
// risk_rating: LOW

package toolsmeta_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/api/toolsmeta"
	"github.com/aprice2704/neuroscript/pkg/tool"
	"github.com/aprice2704/neuroscript/pkg/types"

	// This blank import is necessary for the test to find the standard tools
	// and verify that the export works correctly.
	_ "github.com/aprice2704/neuroscript/pkg/toolbundles/all"
)

func TestExportTools(t *testing.T) {
	// --- Setup ---
	tempDir := t.TempDir()
	outputFile := filepath.Join(tempDir, "test-tools.json")

	// --- Execute ---
	reg := tool.NewToolRegistry(nil)
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

	// Spot-check for a well-known tool to ensure the registry was populated.
	foundReadTool := false
	expectedFullName := types.FullName("tool.fs.read")
	for _, impl := range toolImpls {
		canonicalName := tool.CanonicalizeToolName(string(impl.Spec.Group) + "." + string(impl.Spec.Name))
		if types.FullName(canonicalName) == expectedFullName {
			foundReadTool = true
			break
		}
	}

	if !foundReadTool {
		t.Errorf("Did not find the expected standard tool '%s' in the exported metadata.", expectedFullName)
	}
}
