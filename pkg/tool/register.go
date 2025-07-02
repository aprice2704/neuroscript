// NeuroScript Version: 0.3.1
// File version: 0.1.3 // Remove bootstrap log.Printf INFO messages.
// filename: pkg/tool/register.go

// Package toolsets provides central registration for extended NeuroScript toolsets.
// It decouples core interpreter initialization from specific non-core tool implementations
// by allowing tool packages to register themselves via init().
package tool

import (
	"errors"
	"fmt"
	"log"	// Using standard log package for bootstrap messages
	"sync"

	"github.com/aprice2704/neuroscript/pkg/core"
)

const bootstrapLogPrefix = "[TOOLSET_REGISTRY] "

// ToolRegisterFunc defines the function signature expected for registering a toolset.
type ToolRegisterFunc func(registry core.ToolRegistrar) error

// --- Registry for Toolset Registration Functions ---

var (
	registrationMu		sync.RWMutex
	toolsetRegistrations	= make(map[string]ToolRegisterFunc)
)

// AddToolsetRegistration is called by tool packages (typically in their init() function)
// to register their main registration function with the toolsets package.
func AddToolsetRegistration(name string, regFunc ToolRegisterFunc) {
	registrationMu.Lock()
	defer registrationMu.Unlock()

	if regFunc == nil {
		log.Panicf(bootstrapLogPrefix+"PANIC: Attempted to register nil registration function for toolset '%s'", name)
	}
	if _, exists := toolsetRegistrations[name]; exists {
		log.Printf(bootstrapLogPrefix+"WARN: Toolset registration function for '%s' overwritten.", name)
	}
	toolsetRegistrations[name] = regFunc
	// REMOVED: log.Printf(bootstrapLogPrefix+"INFO: Toolset registration function added for: %s", name)
}

// CreateRegistrationFunc is a helper that takes a toolset name and a slice of ToolImplementations
// and returns a ToolRegisterFunc. This simplifies the registration logic within each toolset package.
func CreateRegistrationFunc(toolsetName string, tools []core.ToolImplementation) ToolRegisterFunc {
	return func(registry core.ToolRegistrar) error {
		if registry == nil {
			err := fmt.Errorf("CreateRegistrationFunc for %s: registry is nil", toolsetName)
			log.Printf(bootstrapLogPrefix+"ERROR: %v", err)	// Log error before returning
			return err
		}
		var errs []error
		for _, toolImpl := range tools {
			if err := registry.RegisterTool(toolImpl); err != nil {
				detailedErr := fmt.Errorf("failed to add tool %q from %s toolset: %w", toolImpl.Spec.Name, toolsetName, err)
				log.Printf(bootstrapLogPrefix+"ERROR: In toolset '%s': %v", toolsetName, detailedErr)
				errs = append(errs, detailedErr)
			}
		}
		if len(errs) > 0 {
			return errors.Join(errs...)
		}
		// REMOVED: log.Printf(bootstrapLogPrefix+"INFO: --- %s Tools Registered ---", toolsetName)
		return nil
	}
}

// RegisterExtendedTools registers all non-core toolsets that have added themselves
// via AddToolsetRegistration. It uses the provided core.ToolRegistrar (which should be
// the interpreter's tool registry).
func RegisterExtendedTools(registry core.ToolRegistrar) error {
	registrationMu.RLock()
	defer registrationMu.RUnlock()

	if registry == nil {
		err := fmt.Errorf("cannot register extended tools: registry is nil")
		log.Printf(bootstrapLogPrefix+"ERROR: %v", err)
		return err
	}

	// REMOVED: log.Printf(bootstrapLogPrefix+"INFO: Registering %d discovered extended toolsets...", len(toolsetRegistrations))

	var allErrors []error

	for name, regFunc := range toolsetRegistrations {
		// REMOVED: log.Printf(bootstrapLogPrefix+"INFO: Calling registration function for toolset: %s", name)
		if err := regFunc(registry); err != nil {
			wrappedErr := fmt.Errorf("failed registering %s toolset: %w", name, err)
			allErrors = append(allErrors, wrappedErr)
			log.Printf(bootstrapLogPrefix+"ERROR: During extended tool registration for toolset '%s': %v", name, err)
		}
	}

	if len(allErrors) > 0 {
		log.Printf(bootstrapLogPrefix+"ERROR: Encountered %d error(s) during extended toolset registration.", len(allErrors))
		return errors.Join(allErrors...)
	}

	// REMOVED: log.Printf(bootstrapLogPrefix + "INFO: Extended tools registered successfully via toolsets.")
	return nil
}