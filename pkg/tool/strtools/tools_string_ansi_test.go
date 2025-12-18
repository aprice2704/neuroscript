// :: product: FDM/NS
// :: majorVersion: 1
// :: fileVersion: 1
// :: description: Tests for ANSI colorization and stripping tools.
// :: latestChange: Initial tests for Colorize and StripAnsi.
// :: filename: pkg/tool/strtools/tools_string_ansi_test.go
// :: serialization: go

package strtools

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/lang"
)

func TestToolStringAnsi(t *testing.T) {
	interp := newStringTestInterpreter(t)

	tests := []struct {
		name       string
		toolName   string
		args       []interface{}
		wantResult interface{}
		wantErrIs  error
	}{
		// --- Colorize Tests ---
		{
			name:       "Colorize Simple Red",
			toolName:   "Colorize",
			args:       MakeArgs("[red]Hello[reset]"),
			wantResult: "\x1b[31mHello\x1b[0m",
		},
		{
			name:       "Colorize Multiple Tags",
			toolName:   "Colorize",
			args:       MakeArgs("[bold][blue]Info:[reset] text"),
			wantResult: "\x1b[1m\x1b[34mInfo:\x1b[0m text",
		},
		{
			name:       "Colorize Unknown Tag",
			toolName:   "Colorize",
			args:       MakeArgs("[unknown] tag"),
			wantResult: "[unknown] tag", // Should remain unchanged
		},
		{
			name:       "Colorize No Tags",
			toolName:   "Colorize",
			args:       MakeArgs("plain text"),
			wantResult: "plain text",
		},
		{
			name:       "Colorize Empty String",
			toolName:   "Colorize",
			args:       MakeArgs(""),
			wantResult: "",
		},
		{
			name:      "Colorize Wrong Type",
			toolName:  "Colorize",
			args:      MakeArgs(123),
			wantErrIs: lang.ErrArgumentMismatch,
		},

		// --- StripAnsi Tests ---
		{
			name:       "StripAnsi Simple",
			toolName:   "StripAnsi",
			args:       MakeArgs("\x1b[31mHello\x1b[0m"),
			wantResult: "Hello",
		},
		{
			name:       "StripAnsi Complex",
			toolName:   "StripAnsi",
			args:       MakeArgs("\x1b[1m\x1b[34mInfo:\x1b[0m text"),
			wantResult: "Info: text",
		},
		{
			name:       "StripAnsi No Codes",
			toolName:   "StripAnsi",
			args:       MakeArgs("plain text"),
			wantResult: "plain text",
		},
		{
			name:       "StripAnsi Empty String",
			toolName:   "StripAnsi",
			args:       MakeArgs(""),
			wantResult: "",
		},
		{
			name:      "StripAnsi Wrong Type",
			toolName:  "StripAnsi",
			args:      MakeArgs(123),
			wantErrIs: lang.ErrArgumentMismatch,
		},
	}

	for _, tt := range tests {
		testStringToolHelper(t, interp, tt)
	}
}
