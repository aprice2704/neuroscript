// :: product: FDM/NS
// :: majorVersion: 1
// :: fileVersion: 1
// :: description: Implements ANSI colorization and manipulation tools using efficient string replacement.
// :: latestChange: Initial implementation of Colorize and StripAnsi.
// :: filename: pkg/tool/strtools/tools_string_ansi.go
// :: serialization: go

package strtools

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

var (
	// ansiReplacer handles the efficient replacement of color tags with ANSI codes.
	ansiReplacer *strings.Replacer
	// ansiRegex is used to strip ANSI codes from strings.
	ansiRegex *regexp.Regexp
)

func init() {
	// Define the mapping of tags to ANSI codes.
	// We use pairs of "tag", "code".
	replacements := []string{
		// --- Resets ---
		"[reset]", "\x1b[0m", // Reset all attributes
		"[default]", "\x1b[39m", // Default foreground color
		"[bg-default]", "\x1b[49m", // Default background color

		// --- Styles ---
		"[bold]", "\x1b[1m",
		"[bright]", "\x1b[1m", // Alias for bold/bright
		"[dim]", "\x1b[2m",
		"[italic]", "\x1b[3m",
		"[underline]", "\x1b[4m",
		"[blink]", "\x1b[5m",
		"[reverse]", "\x1b[7m",
		"[hidden]", "\x1b[8m",
		"[strike]", "\x1b[9m",

		// --- Foreground Colors (Standard) ---
		"[black]", "\x1b[30m",
		"[red]", "\x1b[31m",
		"[green]", "\x1b[32m",
		"[yellow]", "\x1b[33m",
		"[blue]", "\x1b[34m",
		"[magenta]", "\x1b[35m",
		"[cyan]", "\x1b[36m",
		"[white]", "\x1b[37m",

		// --- Foreground Colors (Bright/High Intensity) ---
		"[bright-black]", "\x1b[90m",
		"[gray]", "\x1b[90m", // Common alias
		"[bright-red]", "\x1b[91m",
		"[bright-green]", "\x1b[92m",
		"[bright-yellow]", "\x1b[93m",
		"[bright-blue]", "\x1b[94m",
		"[bright-magenta]", "\x1b[95m",
		"[bright-cyan]", "\x1b[96m",
		"[bright-white]", "\x1b[97m",

		// --- Background Colors (Standard) ---
		"[bg-black]", "\x1b[40m",
		"[bg-red]", "\x1b[41m",
		"[bg-green]", "\x1b[42m",
		"[bg-yellow]", "\x1b[43m",
		"[bg-blue]", "\x1b[44m",
		"[bg-magenta]", "\x1b[45m",
		"[bg-cyan]", "\x1b[46m",
		"[bg-white]", "\x1b[47m",

		// --- Background Colors (Bright/High Intensity) ---
		"[bg-bright-black]", "\x1b[100m",
		"[bg-bright-red]", "\x1b[101m",
		"[bg-bright-green]", "\x1b[102m",
		"[bg-bright-yellow]", "\x1b[103m",
		"[bg-bright-blue]", "\x1b[104m",
		"[bg-bright-magenta]", "\x1b[105m",
		"[bg-bright-cyan]", "\x1b[106m",
		"[bg-bright-white]", "\x1b[107m",
	}

	ansiReplacer = strings.NewReplacer(replacements...)

	// Regex to match ANSI escape sequences (CSI codes ending in m) for stripping.
	ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*m`)
}

// toolStringColorize replaces supported color tags (e.g., [red], [bold]) with their ANSI escape codes.
func toolStringColorize(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "Colorize: expected 1 argument (input_string)", lang.ErrArgumentMismatch)
	}
	inputStr, ok := args[0].(string)
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("Colorize: input_string argument must be a string, got %T", args[0]), lang.ErrArgumentMismatch)
	}

	// Efficiently replace all tags
	result := ansiReplacer.Replace(inputStr)

	//interpreter.GetLogger().Debug("Tool: Colorize", "input_len", len(inputStr), "result_len", len(result))
	return result, nil
}

// toolStringStripAnsi removes all ANSI escape codes from the string.
func toolStringStripAnsi(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "StripAnsi: expected 1 argument (input_string)", lang.ErrArgumentMismatch)
	}
	inputStr, ok := args[0].(string)
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("StripAnsi: input_string argument must be a string, got %T", args[0]), lang.ErrArgumentMismatch)
	}

	result := ansiRegex.ReplaceAllString(inputStr, "")

	interpreter.GetLogger().Debug("Tool: StripAnsi", "input_len", len(inputStr), "result_len", len(result))
	return result, nil
}
