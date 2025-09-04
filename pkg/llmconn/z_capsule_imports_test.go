// NeuroScript Version: 0.7.0
// File version: 1
// Purpose: Ensures the bootstrap capsule package is imported, triggering its init() function to populate the capsule registry before tests are run.
// filename: pkg/llmconn/z_capsule_imports_test.go
// nlines: 12
// risk_rating: LOW

package llmconn_test

import (
	// This blank import is critical. It forces the Go compiler to include
	// the bootstrap capsule package in the test binary, which in turn
	// triggers its init() function, populating the global capsule registry.
	// Without this, capsule.Get() would fail to find the required prompts.
	_ "github.com/aprice2704/neuroscript/pkg/capsule"
)
