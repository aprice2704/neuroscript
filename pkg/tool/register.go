// NeuroScript Version: 0.3.1
// File version: 0.1.3 // Remove bootstrap log.Printf INFO messages.
// filename: pkg/tool/register.go

// Package toolsets provides central registration for extended NeuroScript tool.
// It decouples core interpreter initialization from specific non-core tool implementations
// by allowing tool packages to register themselves via init().
package tool

import (
	"errors"
	"fmt"
	"log" // Using standard log package for bootstrap messages
	"sync"
)

const bootstrapLogPrefix = "[TOOLSET_REGISTRY] "

// ToolRegisterFunc defines the function signature expected for registering a toolset.
type ToolRegisterFunc func(registry ToolRegistrar) error

// --- Registry for Toolset Registration Functions ---

var (
	registrationMu       sync.RWMutex
	toolsetRegistrations = make(map[string]ToolRegisterFunc)
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
func CreateRegistrationFunc(toolsetName string, tools []Implementation) ToolRegisterFunc {
	return func(registry Registrar) error {
		if registry == nil {
			err := fmt.Errorf("CreateRegistrationFunc for %s: registry is nil", toolsetName)
			log.Printf(bootstrapLogPrefix+"ERROR: %v", err) // Log error before returning
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
// via AddToolsetRegistration. It uses the provided  Registrar (which should be
// the interpreter's tool registry).
func RegisterExtendedTools(registry Registrar) error {
	registrationMu.RLock()
	defer registrationMu.RUnlock()

	if registry == nil {
		err := fmt.Errorf("cannot register extended tools: registry is nil")
		log.Printf(bootstrapLogPrefix+"ERROR: %v", err)
		return err
	}

	// REMOVED: log.Printf(bootstrapLogPrefix+"INFO: Registering %d discovered extended tool...", len(toolsetRegistrations))

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

	// REMOVED: log.Printf(bootstrapLogPrefix + "INFO: Extended tools registered successfully via tool.")
	return nil
}

// registerCoreToolsInternal is called by the public RegisterCoreTools.
// It now accepts the ToolRegistry interface.
// Its role is to register any core tools that *cannot* use the
// init() + AddToolImplementations pattern.
// Processing of globalToolImplementations is handled by NewToolRegistry.
func registerCoreToolsInternal(registry ToolRegistry) error {
	// If this function needs to log, it should obtain a logger via the registry
	// interface if such a method is added to ToolRegistry, or handle logging externally.

	// --- Legacy or Non-Init Registration Section ---
	// Example:
	// specificTool := ToolImplementation{Spec: ToolSpec{Name: "MySpecialTool", ...}, Func: toolMySpecialTool}
	// if err := registry.RegisterTool(specificTool); err != nil { // Uses the interface method
	//     return fmt.Errorf("failed registering MySpecialTool: %w", err)
	// }
	// --- End Legacy or Non-Init Registration Section ---

	// For now, assume NewToolRegistry (called during Interpreter init) handles all core tool registrations
	// that use the init-time pattern. If there are other core tools needing explicit registration
	// after the registry is created, they would be registered here using registry.RegisterTool().
	return nil
}

// RegisterCoreTools is the public entry point for ensuring core tools are available.
// It accepts the ToolRegistry interface.
// With the init-based registration pattern, NewToolRegistry (called by Interpreter)
// handles the primary registration from globalToolImplementations.
func RegisterCoreTools(registry ToolRegistry) error {
	if registry == nil {
		return fmt.Errorf("RegisterCoreTools called with a nil registry")
	}

	// Call the internal function, passing the interface.
	if err := registerCoreToolsInternal(registry); err != nil {
		return err
	}

	// Logging: If the ToolRegistry interface had a Logger() method, we could use it.
	// Example:
	// if logger := registry.Logger(); logger != nil {
	//     logger.Debug("RegisterCoreTools: Explicit core tool registration phase complete.")
	// }
	// For now, this specific logging is removed as the interface doesn't expose the logger directly.
	// The creation of the ToolRegistryImpl within NewInterpreter already logs.

	return nil
}
