// NeuroScript Version: 0.6.0
// File version: 1
// Purpose: Adds tests for edge cases like scripts with only comments or metadata.
// filename: pkg/parser/additional_parsing_test.go
// nlines: 70
// risk_rating: LOW

package parser

import (
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/logging"
)

func TestEdgeCaseParsing(t *testing.T) {
	testCases := map[string]struct {
		scriptContent         string
		expectedProgramErrMsg string
		expectedFileMetaCount int
		expectedCommentCount  int
	}{
		"Only Comments": {
			scriptContent: `
# This is a file with only comments.
-- Another comment.
# And a third.
`,
			expectedCommentCount: 3,
		},
		"Only Metadata": {
			scriptContent: `
:: key1: value1
:: key2: value2
`,
			expectedFileMetaCount: 2,
		},
		"Mixed Comments and Metadata": {
			scriptContent: `
# A leading comment.
:: key1: value1
-- A comment between metadata lines.
:: key2: value2
# A trailing comment.
`,
			expectedFileMetaCount: 2,
			expectedCommentCount:  3,
		},
		"Blank Lines Between Functions": {
			scriptContent: `
func FirstFunc() means
	emit "first"
endfunc


func SecondFunc() means
	emit "second"
endfunc
`,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			logger := logging.NewTestLogger(t)
			parserAPI := NewParserAPI(logger)
			tree, tokenStream, errs := parserAPI.parseInternal(name+".ns", tc.scriptContent)
			if len(errs) > 0 {
				t.Fatalf("Parse failed with errors: %v", errs)
			}

			builder := NewASTBuilder(logger)
			program, fileMetadata, err := builder.BuildFromParseResult(tree, tokenStream)
			if err != nil {
				if tc.expectedProgramErrMsg == "" || !strings.Contains(err.Error(), tc.expectedProgramErrMsg) {
					t.Fatalf("Build failed unexpectedly: %v", err)
				}
				return
			}
			if tc.expectedProgramErrMsg != "" {
				t.Fatalf("Expected build to fail with '%s', but it succeeded.", tc.expectedProgramErrMsg)
			}

			if len(fileMetadata) != tc.expectedFileMetaCount {
				t.Errorf("Expected %d file metadata entries, but got %d", tc.expectedFileMetaCount, len(fileMetadata))
			}

			if len(program.Comments) != tc.expectedCommentCount {
				t.Errorf("Expected %d program-level comments, but got %d", tc.expectedCommentCount, len(program.Comments))
			}
		})
	}
}
