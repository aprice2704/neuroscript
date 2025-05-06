// NeuroScript Version: 0.3.1
// File version: 0.0.1
// Registration for nspatch tools.
// filename: pkg/nspatch/register.go

package nspatch

import (
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/core"
	"github.com/aprice2704/neuroscript/pkg/toolsets"
)

// init registers the nspatch toolset.
func init() {
	toolsets.AddToolsetRegistration("nspatch", RegisterNsPatchTools)
}

// RegisterNsPatchTools registers the tools in this package.
func RegisterNsPatchTools(registry core.ToolRegistrar) error {
	var registrationErrors []error
	// Add toolGeneratePatchImpl
	if err := registry.RegisterTool(toolGeneratePatchImpl); err != nil {
		registrationErrors = append(registrationErrors, err)
	}

	// TODO: Add toolApplyPatchImpl wrapper here later

	if len(registrationErrors) > 0 {
		// Consider joining errors properly if multiple tools are registered
		return fmt.Errorf("errors registering nspatch tools: %w", registrationErrors[0])
	}
	fmt.Println("--- NsPatch Tools Registered ---") // Log registration
	return nil
}
