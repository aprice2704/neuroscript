// NeuroScript Version: 0.3.0
// Last Modified: 2025-05-01 20:56:53 PDT // Split file: Go Module tools
// filename: pkg/core/tools_go_mod.go

package core

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/aprice2704/neuroscript/pkg/logging"
	"golang.org/x/mod/modfile"
)

// --- Helper to find and parse go.mod ---

// FindAndParseGoMod searches upwards from startDir for go.mod, parses it, and returns the parsed file,
// the directory it was found in (module root), and any error.
// If the file is not found, it returns a specific error wrapping os.ErrNotExist.
func FindAndParseGoMod(startDir string, log logging.Logger) (*modfile.File, string, error) {
	if startDir == "" {
		return nil, "", fmt.Errorf("start directory cannot be empty")
	}
	if log == nil {
		// Create a discard logger if none provided? Or error? Let's error for safety.
		return nil, "", fmt.Errorf("logger cannot be nil for FindAndParseGoMod")
	}

	logPrefix := "[FindGoMod]"
	log.Debug("%s Starting search for go.mod from: %s", logPrefix, startDir)

	// Ensure startDir is absolute and clean first for reliable traversal
	absStartDir, err := filepath.Abs(startDir)
	if err != nil {
		log.Error("%s Error making start path absolute %q: %v", logPrefix, startDir, err)
		return nil, "", fmt.Errorf("internal error resolving start path '%s': %w", startDir, err)
	}
	currentDir := filepath.Clean(absStartDir)
	log.Debug("%s Absolute starting directory: %s", logPrefix, currentDir)

	// Loop upwards until root or go.mod is found
	for {
		goModPath := filepath.Join(currentDir, "go.mod")
		log.Debug("%s Checking: %s", logPrefix, goModPath)

		modContent, readErr := os.ReadFile(goModPath)
		if readErr == nil {
			// Found go.mod, try parsing
			log.Debug("%s Found go.mod at: %s", logPrefix, goModPath)
			modF, parseErr := modfile.Parse(goModPath, modContent, nil)
			if parseErr != nil {
				log.Error("%s Failed to parse %s: %v", logPrefix, goModPath, parseErr)
				// Return specific parse error
				return nil, "", fmt.Errorf("failed to parse go.mod at %s: %w", goModPath, parseErr)
			}
			// Check if Module path is actually set after successful parsing
			if modF.Module == nil || modF.Module.Mod.Path == "" {
				log.Error("%s Parsed %s but Module path is missing or empty.", logPrefix, goModPath)
				return nil, "", fmt.Errorf("parsed go.mod at %s but module path is missing", goModPath)
			}
			log.Debug("%s Successfully parsed go.mod. Module: %s", logPrefix, modF.Module.Mod.Path)
			// Return parsed file and the directory it was found in (module root)
			return modF, currentDir, nil // Success
		}

		// Error reading file - if it's *not* ErrNotExist, it's a real error
		if !errors.Is(readErr, os.ErrNotExist) {
			log.Error("%s Error reading potential go.mod at %s: %v", logPrefix, goModPath, readErr)
			// Return specific read error
			return nil, "", fmt.Errorf("error reading file %s: %w", goModPath, readErr)
		}

		// If ErrNotExist, go up one directory
		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			// Reached root directory (e.g., "/", "C:\") without finding go.mod
			log.Debug("%s Reached filesystem root, go.mod not found.", logPrefix)
			// Return a specific "not found" error wrapping os.ErrNotExist for clarity
			return nil, "", fmt.Errorf("go.mod not found in or above %s: %w", absStartDir, os.ErrNotExist)
		}
		currentDir = parentDir
	}
}

// --- Tool: GoGetModuleInfo ---

func toolGoGetModuleInfo(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	logPrefix := "[TOOL-GoGetModuleInfo]"
	logger := interpreter.Logger()
	startDirRel := "." // Default search start is sandbox root

	// Parse optional directory argument
	if len(args) > 0 && args[0] != nil {
		dirStr, ok := args[0].(string)
		if !ok {
			// Return nil result and validation Go error
			return nil, fmt.Errorf("%w: directory argument must be a string, got %T", ErrValidationTypeMismatch, args[0])
		}
		startDirRel = dirStr
	}

	logger.Debug("%s Called with relative directory: %q", logPrefix, startDirRel)

	// Validate the starting directory relative to the sandbox
	sandboxRoot := interpreter.SandboxDir()
	if sandboxRoot == "" {
		// Return nil result and security Go error
		return nil, fmt.Errorf("%w: interpreter sandbox directory not set", ErrInternalSecurity)
	}
	absStartDir, secErr := ResolveAndSecurePath(startDirRel, sandboxRoot)
	if secErr != nil {
		logger.Error("%s Path security error for start directory %q: %v", logPrefix, startDirRel, secErr)
		// Return nil result and security Go error (ErrPathViolation is wrapped)
		return nil, fmt.Errorf("invalid start directory: %w", secErr)
	}
	logger.Debug("%s Resolved start directory to: %s", logPrefix, absStartDir)

	// Find and parse go.mod using the EXPORTED helper
	modF, modRootDir, err := FindAndParseGoMod(absStartDir, logger)
	if err != nil {
		// Check if the error was specifically "not found"
		if errors.Is(err, os.ErrNotExist) {
			logger.Warn("%s No go.mod found starting from %q.", logPrefix, startDirRel)
			// Return nil (representing not found) and no Go error for NeuroScript
			return nil, nil
		}
		// For other errors (read error, parse error), log and return nil result + Go error
		logger.Error("%s Failed to find or parse go.mod starting from %q: %v", logPrefix, startDirRel, err)
		// Wrap the specific error (read/parse/missing module path) from FindAndParseGoMod
		return nil, fmt.Errorf("%w: %w", ErrInternalTool, err)
	}

	// Extract information into a map
	resultMap := make(map[string]interface{})
	if modF.Module != nil && modF.Module.Mod.Path != "" {
		resultMap["modulePath"] = modF.Module.Mod.Path
	} else {
		// Should be caught by FindAndParseGoMod, but double-check defensively
		logger.Warn("%s Parsed go.mod but module path is missing (unexpected).", logPrefix)
		resultMap["modulePath"] = ""
	}
	if modF.Go != nil && modF.Go.Version != "" {
		resultMap["goVersion"] = modF.Go.Version
	} else {
		resultMap["goVersion"] = "" // Go version might not be specified
	}

	// Return the absolute path to the directory containing the go.mod
	resultMap["rootDir"] = modRootDir

	// Format requires
	reqList := []map[string]interface{}{}
	if modF.Require != nil {
		for _, req := range modF.Require {
			reqMap := map[string]interface{}{
				"path":     req.Mod.Path,
				"version":  req.Mod.Version,
				"indirect": req.Indirect,
			}
			reqList = append(reqList, reqMap)
		}
	}
	resultMap["requires"] = reqList

	// Format replaces
	repList := []map[string]interface{}{}
	if modF.Replace != nil {
		for _, rep := range modF.Replace {
			repMap := map[string]interface{}{
				"oldPath":    rep.Old.Path,
				"oldVersion": rep.Old.Version,
				"newPath":    rep.New.Path,
				"newVersion": rep.New.Version,
			}
			repList = append(repList, repMap)
		}
	}
	resultMap["replaces"] = repList

	logPath := ""
	if p, ok := resultMap["modulePath"].(string); ok {
		logPath = p
	}
	logger.Debug("%s Successfully retrieved module info for %q", logPrefix, logPath)
	return resultMap, nil
}
