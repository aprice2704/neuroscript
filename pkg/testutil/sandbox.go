// NeuroScript Version: 0.6.0
// File version: 1
// Purpose: Provides a centralized, project-wide test helper for creating sandboxed interpreters.
// filename: pkg/testutil/sandbox.go
// nlines: 23
// risk_rating: LOW

package testutil

import (
	"os"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/api"
)

// NewTestSandbox creates a temporary directory for a test and returns an
// api.Option to configure an interpreter with it. It registers a cleanup
// function with the test to remove the directory after the run.
// This function should be used by any test, in any package, that needs a
// sandboxed filesystem.
func NewTestSandbox(t *testing.T) api.Option {
	t.Helper()
	tempDir, err := os.MkdirTemp("", "neuroscript_test_sandbox_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir for sandbox: %v", err)
	}

	t.Cleanup(func() {
		os.RemoveAll(tempDir)
	})

	return api.WithSandboxDir(tempDir)
}
