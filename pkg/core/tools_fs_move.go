// filename: pkg/core/tools_fs_move.go
package core

import (
	"errors"
	"fmt"
	"os"
)

// toolMoveFile implements the TOOL.MoveFile command.
func toolMoveFile(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// --- Argument Validation ---
	if len(args) != 2 {
		return nil, fmt.Errorf("TOOL.MoveFile: expected 2 arguments (source, destination), got %d", len(args))
	}
	sourcePath, okSrc := args[0].(string)
	destPath, okDest := args[1].(string)
	if !okSrc || !okDest {
		return nil, fmt.Errorf("TOOL.MoveFile: both source and destination arguments must be strings")
	}
	if sourcePath == "" || destPath == "" {
		return nil, fmt.Errorf("TOOL.MoveFile: source and destination paths cannot be empty")
	}

	// --- Path Security Validation ---
	// Use interpreter's sandboxDir if set, otherwise current dir "."
	sandboxRoot := interpreter.sandboxDir
	if sandboxRoot == "" {
		interpreter.logger.Warn("TOOL MoveFile] Interpreter sandboxDir is empty, using default relative path validation from current directory.")
		sandboxRoot = "."
	}

	absSource, errSource := SecureFilePath(sourcePath, sandboxRoot)
	if errSource != nil {
		errMsg := fmt.Sprintf("Invalid source path '%s': %v", sourcePath, errSource)
		interpreter.logger.Info("Tool: MoveFile] Error: %s", errMsg)
		// Return map with error, plus Go error for interpreter
		return map[string]interface{}{"error": errMsg}, fmt.Errorf("TOOL.MoveFile: %w", errors.Join(ErrValidationArgValue, errSource))
	}

	absDest, errDest := SecureFilePath(destPath, sandboxRoot)
	if errDest != nil {
		errMsg := fmt.Sprintf("Invalid destination path '%s': %v", destPath, errDest)
		interpreter.logger.Info("Tool: MoveFile] Error: %s", errMsg)
		// Return map with error, plus Go error for interpreter
		return map[string]interface{}{"error": errMsg}, fmt.Errorf("TOOL.MoveFile: %w", errors.Join(ErrValidationArgValue, errDest))
	}

	interpreter.logger.Info("Tool: MoveFile] Validated paths: Source '%s' -> '%s', Dest '%s' -> '%s'", sourcePath, absSource, destPath, absDest)

	// --- Pre-Move Checks (Source Exists, Destination Does Not) ---
	_, srcStatErr := os.Stat(absSource)
	if srcStatErr != nil {
		errMsg := ""
		if errors.Is(srcStatErr, os.ErrNotExist) {
			errMsg = fmt.Sprintf("Source path '%s' does not exist.", sourcePath)
		} else {
			errMsg = fmt.Sprintf("Error checking source path '%s': %v", sourcePath, srcStatErr)
		}
		interpreter.logger.Info("Tool: MoveFile] Error: %s", errMsg)
		return map[string]interface{}{"error": errMsg}, fmt.Errorf("TOOL.MoveFile: %w", srcStatErr)
	}

	_, destStatErr := os.Stat(absDest)
	if destStatErr == nil {
		// Destination exists! Abort as per spec.
		errMsg := fmt.Sprintf("Destination path '%s' already exists.", destPath)
		interpreter.logger.Info("Tool: MoveFile] Error: %s", errMsg)
		// Not strictly a validation error, but a precondition failure
		return map[string]interface{}{"error": errMsg}, fmt.Errorf("TOOL.MoveFile: %s", errMsg)
	} else if !errors.Is(destStatErr, os.ErrNotExist) {
		// Error other than NotExist when checking destination
		errMsg := fmt.Sprintf("Error checking destination path '%s': %v", destPath, destStatErr)
		interpreter.logger.Info("Tool: MoveFile] Error: %s", errMsg)
		return map[string]interface{}{"error": errMsg}, fmt.Errorf("TOOL.MoveFile: %w", destStatErr)
	}
	// If we reach here, source exists and destination does not exist (or we got ErrNotExist)

	// --- Perform Move/Rename ---
	interpreter.logger.Info("Tool: MoveFile] Attempting rename/move: '%s' -> '%s'", absSource, absDest)
	renameErr := os.Rename(absSource, absDest)
	if renameErr != nil {
		errMsg := fmt.Sprintf("Failed to move/rename '%s' to '%s': %v", sourcePath, destPath, renameErr)
		interpreter.logger.Info("Tool: MoveFile] Error: %s", errMsg)
		return map[string]interface{}{"error": errMsg}, fmt.Errorf("TOOL.MoveFile: %w", renameErr)
	}

	// --- Success ---
	interpreter.logger.Info("Tool: MoveFile] Successfully moved/renamed '%s' to '%s'", sourcePath, destPath)
	return map[string]interface{}{"error": nil}, nil
}

// --- Registration ---

func registerFsMoveTools(registry *ToolRegistry) error {
	return registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{
			Name:        "MoveFile",
			Description: "Moves or renames a file or directory within the sandbox.",
			Args: []ArgSpec{
				{Name: "source", Type: ArgTypeString, Required: true, Description: "The current path to the file/directory."},
				{Name: "destination", Type: ArgTypeString, Required: true, Description: "The desired new path for the file/directory."},
			},
			ReturnType: ArgTypeAny, // Returns map[string]interface{} -> Any
		},
		Func: toolMoveFile,
	})
}
