// NeuroScript Version: 0.5.2
// File version: 1
// Purpose: Implements string utility tools (LineCount).
// filename: pkg/tool/strtools/tools_string_utils.go
// nlines: 39
// risk_rating: LOW

package strtools

import (
	"fmt"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

func toolLineCountString(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	// Corresponds to "LineCount" tool with args: content_string
	if len(args) != 1 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "String.LineCount: expected 1 argument (content_string)", lang.ErrArgumentMismatch)
	}
	content, ok := args[0].(string)
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("String.LineCount: content_string argument must be a string, got %T", args[0]), lang.ErrArgumentMismatch)
	}

	if content == "" {
		// interpreter.GetLogger().Debug("Tool: String.LineCount", "content", content, "line_count", 0)
		return float64(0), nil
	}
	// Count occurrences of newline character
	lineCount := float64(strings.Count(content, "\n"))
	// Add 1 if the string doesn't end with a newline (to count the last line)
	if !strings.HasSuffix(content, "\n") {
		lineCount++
	}

	// interpreter.GetLogger().Debug("Tool: String.LineCount", "content_len", len(content), "line_count", lineCount)
	return lineCount, nil
}
