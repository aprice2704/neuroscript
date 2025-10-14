// NeuroScript Version: 0.8.0
// File version: 1
// Purpose: Provides a centralized test harness for API tests to reduce boilerplate.
// filename: pkg/api/test_harness_test.go
// nlines: 25
// risk_rating: LOW

package api_test

import (
	"io"
	"os"

	"github.com/aprice2704/neuroscript/pkg/api"
	"github.com/aprice2704/neuroscript/pkg/logging"
)

// newTestHostContext creates a minimal, valid HostContext for use in API tests.
// It provides non-nil, discarding I/O streams to satisfy the builder's requirements.
func newTestHostContext(logger api.Logger) *api.HostContext {
	if logger == nil {
		logger = logging.NewNoOpLogger()
	}
	hc, err := api.NewHostContextBuilder().
		WithLogger(logger).
		WithStdout(io.Discard).
		WithStdin(os.Stdin).
		WithStderr(io.Discard).
		Build()
	if err != nil {
		// This should never happen in a test context with these values.
		panic("failed to build test host context: " + err.Error())
	}
	return hc
}
