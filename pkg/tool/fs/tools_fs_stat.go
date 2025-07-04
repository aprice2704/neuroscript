// filename: pkg/tool/fs/tools_fs_stat.go
package fs

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/security"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

// toolStat implements the FS.Stat tool.
func toolStat(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "Stat: expected 1 argument (path)", lang.ErrArgumentMismatch)
	}

	relPath, ok := args[0].(string)
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("Stat: path argument must be a string, got %T", args[0]), lang.ErrInvalidArgument)
	}
	if relPath == "" {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "Stat: path cannot be empty", lang.ErrInvalidArgument)
	}

	sandboxRoot := interpreter.SandboxDir()
	absPath, secErr := security.ResolveAndSecurePath(relPath, sandboxRoot)
	if secErr != nil {
		return nil, secErr
	}

	info, err := os.Stat(absPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, lang.NewRuntimeError(lang.ErrorCodeFileNotFound, fmt.Sprintf("Stat: path not found '%s'", relPath), lang.ErrFileNotFound)
		}
		return nil, lang.NewRuntimeError(lang.ErrorCodeIOFailed, fmt.Sprintf("Stat: failed to get info for '%s'", relPath), errors.Join(lang.ErrIOFailed, err))
	}

	return map[string]interface{}{
		"name":             info.Name(),
		"path":             relPath,
		"size_bytes":       info.Size(),
		"is_dir":           info.IsDir(),
		"modified_unix":    info.ModTime().Unix(),
		"modified_rfc3339": info.ModTime().Format(time.RFC3339Nano),
		"mode_string":      info.Mode().String(),
		"mode_perm":        info.Mode().Perm().String(),
	}, nil
}
