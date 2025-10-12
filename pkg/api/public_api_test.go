// NeuroScript Version: 0.8.0
// File version: 3
// Purpose: Corrects the event handler callback test to use the new HostContextBuilder pattern.
// filename: pkg/api/public_api_test.go
// nlines: 50
// risk_rating: LOW

package api_test

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/api"
)

// TestPublicAPI_PolicyBuilderIsAccessible confirms that the fluent policy builder,
// a key part of the public API, is correctly re-exported and can be used.
func TestPublicAPI_PolicyBuilderIsAccessible(t *testing.T) {
	// This test simply needs to compile and run without panicking.
	_ = api.NewPolicyBuilder(api.ContextNormal).
		Allow("tool.fs.read").
		Grant("fs:read:/tmp/*").
		Build()
}

// TestPublicAPI_CapabilityHelpersAreAccessible confirms that the various helpers
// for creating capabilities are available through the api package.
func TestPublicAPI_CapabilityHelpersAreAccessible(t *testing.T) {
	// This test simply needs to compile and run without panicking.
	_ = api.NewCapability(api.ResFS, api.VerbRead, "/tmp/*")
	_, err := api.ParseCapability("net:read:*.example.com")
	if err != nil {
		t.Fatalf("api.ParseCapability failed: %v", err)
	}
}

// TestPublicAPI_ReExportedTypes confirms that key types can be instantiated
// via the api package, which is crucial for consumers like FDM.
func TestPublicAPI_ReExportedTypes(t *testing.T) {
	// This test simply needs to compile and run without panicking.
	var _ api.ToolImplementation
	var _ api.AIProvider
	var _ api.Logger
}

// TestPublicAPI_EventHandlerCallbackIsAccessible verifies that the event handler
// error callback can be set via the HostContextBuilder.
func TestPublicAPI_EventHandlerCallbackIsAccessible(t *testing.T) {
	// This test just needs to compile. It now uses the correct builder pattern.
	_ = api.NewHostContextBuilder().WithEventHandlerErrorCallback(func(eventName, source string, err *api.RuntimeError) {})
}
