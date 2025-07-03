// NeuroScript Version: 0.4.0
// File version: 1
// Purpose: Fix nil return type in toolGoGetModuleInfo for 'not found' case to match test expectations.
// filename: pkg/tool/gotools/tools_go_mod.go
// nlines: 128
// risk_rating: LOW

package gotools

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	// Assumed import from original file
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/security"
	"github.com/aprice2704/neuroscript/pkg/tool"
	"golang.org/x/mod/modfile"
)

// --- Helper to find and parse go.mod ---

// FindAndParseGoMod searches upwards from startDir for go.mod, parses it, and returns the parsed file,
// the directory it was found in (module root), and any error.
// If the file is not found, it returns a specific error wrapping os.ErrNotExist.
func FindAndParseGoMod(startDir string, log interfaces.Logger) (*modfile.File, string, error) {
	if startDir == "" {
		return nil, "", fmt.Errorf("start directory cannot be empty")
	}
	if log == nil {
		return nil, "", fmt.Errorf("logger cannot be nil for FindAndParseGoMod")
	}

	logPrefix := "[FindGoMod]"
	log.Debug("%s Starting search for go.mod from: %s", logPrefix, startDir)

	absStartDir, err := filepath.Abs(startDir)
	if err != nil {
		log.Error("%s Error making start path absolute %q: %v", logPrefix, startDir, err)
		return nil, "", fmt.Errorf("internal error resolving start path '%s': %w", startDir, err)
	}
	currentDir := filepath.Clean(absStartDir)
	log.Debug("%s Absolute starting directory: %s", logPrefix, currentDir)

	for {
		goModPath := filepath.Join(currentDir, "go.mod")
		log.Debug("%s Checking: %s", logPrefix, goModPath)

		modContent, readErr := os.ReadFile(goModPath)
		if readErr == nil {
			log.Debug("%s Found go.mod at: %s", logPrefix, goModPath)
			modF, parseErr := modfile.Parse(goModPath, modContent, nil)
			if parseErr != nil {
				log.Error("%s Failed to parse %s: %v", logPrefix, goModPath, parseErr)
				return nil, "", fmt.Errorf("failed to parse go.mod at %s: %w", goModPath, parseErr)
			}
			if modF.Module == nil || modF.Module.Mod.Path == "" {
				log.Error("%s Parsed %s but Module path is missing or empty.", logPrefix, goModPath)
				return nil, "", fmt.Errorf("parsed go.mod at %s but module path is missing", goModPath)
			}
			log.Debug("%s Successfully parsed go.mod. Module: %s", logPrefix, modF.Module.Mod.Path)
			return modF, currentDir, nil
		}

		if !errors.Is(readErr, os.ErrNotExist) {
			log.Error("%s Error reading potential go.mod at %s: %v", logPrefix, goModPath, readErr)
			return nil, "", fmt.Errorf("error reading file %s: %w", goModPath, readErr)
		}

		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			log.Debug("%s Reached filesystem root, go.mod not found.", logPrefix)
			return nil, "", fmt.Errorf("go.mod not found in or above %s: %w", absStartDir, os.ErrNotExist)
		}
		currentDir = parentDir
	}
}

// --- Tool: GoGetModuleInfo ---

func toolGoGetModuleInfo(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	logPrefix := "[TOOL-GoGetModuleInfo]"
	logger := interpreter.GetLogger()
	startDirRel := "."

	if len(args) > 0 && args[0] != nil {
		dirStr, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("%w: directory argument must be a string, got %T", lang.ErrValidationTypeMismatch, args[0])
		}
		startDirRel = dirStr
	}

	logger.Debug("%s Called with relative directory: %q", logPrefix, startDirRel)

	sandboxRoot := interpreter.SandboxDir()
	if sandboxRoot == "" {
		return nil, fmt.Errorf("%w: interpreter sandbox directory not set", lang.ErrInternalSecurity)
	}
	// Assuming ResolveAndSecurePath is available in the core package.
	// If it's from security.go or security_helpers.go, it should be fine.
	absStartDir, secErr := security.ResolveAndSecurePath(startDirRel, sandboxRoot)
	if secErr != nil {
		logger.Error("%s Path security error for start directory %q: %v", logPrefix, startDirRel, secErr)
		return nil, fmt.Errorf("invalid start directory: %w", secErr)
	}
	logger.Debug("%s Resolved start directory to: %s", logPrefix, absStartDir)

	modF, modRootDir, err := FindAndParseGoMod(absStartDir, logger)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			logger.Warn("%s No go.mod found starting from %q.", logPrefix, startDirRel)
			// Return a typed nil map to match the test expectation and function's successful return type.
			return (map[string]interface{})(nil), nil
		}
		logger.Error("%s Failed to find or parse go.mod starting from %q: %v", logPrefix, startDirRel, err)
		return nil, fmt.Errorf("%w: %w", lang.ErrInternalTool, err)
	}

	resultMap := make(map[string]interface{})
	if modF.Module != nil && modF.Module.Mod.Path != "" {
		resultMap["modulePath"] = modF.Module.Mod.Path
	} else {
		resultMap["modulePath"] = ""
	}
	if modF.Go != nil && modF.Go.Version != "" {
		resultMap["goVersion"] = modF.Go.Version
	} else {
		resultMap["goVersion"] = ""
	}

	resultMap["rootDir"] = modRootDir

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
