// NeuroScript Version: 0.3.1
// File version: 0.1.4 // Corrected Registrar typo to ToolRegistrar.
// filename: pkg/tool/register.go

// Package tool provides central registration for extended NeuroScript tool.
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
}

// CreateRegistrationFunc is a helper that takes a toolset name and a slice of ToolImplementations
// and returns a ToolRegisterFunc. This simplifies the registration logic within each toolset package.
func CreateRegistrationFunc(toolsetName string, tools []ToolImplementation) ToolRegisterFunc {
	// FIX: Corrected type to ToolRegistrar
	return func(registry ToolRegistrar) error {
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
		return nil
	}
}

// RegisterExtendedTools registers all non-core toolsets that have added themselves
// via AddToolsetRegistration. It uses the provided ToolRegistrar (which should be
// the interpreter's tool registry).
// FIX: Corrected type to ToolRegistrar
func RegisterExtendedTools(registry ToolRegistrar) error {
	registrationMu.RLock()
	defer registrationMu.RUnlock()

	if registry == nil {
		err := fmt.Errorf("cannot register extended tools: registry is nil")
		log.Printf(bootstrapLogPrefix+"ERROR: %v", err)
		return err
	}

	var allErrors []error

	for name, regFunc := range toolsetRegistrations {
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

	return nil
}

// registerCoreToolsInternal is called by the public RegisterCoreTools.
// It now accepts the ToolRegistry interface.
func registerCoreToolsInternal(registry ToolRegistry) error {
	// This function can be expanded to register core tools that don't use the init() pattern.
	return nil
}

// RegisterCoreTools is the public entry point for ensuring core tools are available.
// It accepts the ToolRegistry interface.
func RegisterCoreTools(registry ToolRegistry) error {
	if registry == nil {
		return fmt.Errorf("RegisterCoreTools called with a nil registry")
	}

	if err := registerCoreToolsInternal(registry); err != nil {
		return err
	}

	return nil
}
