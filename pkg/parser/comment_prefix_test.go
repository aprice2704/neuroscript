// NeuroScript Version: 0.6.0
// File version: 4
// Purpose: Adds source code logging to the comment prefix tests.
// filename: pkg/parser/comment_prefix_test.go
// nlines: 85
// risk_rating: LOW

package parser

import (
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/logging"
)

func TestCommentPrefixes(t *testing.T) {
	testCases := map[string]struct {
		script               string
		expectedCommentCount int
		expectedFirstComment string
	}{
		"Hash Comments": {
			script: `
# Line 1
# Line 2

func main() means
    # In function
    set _ = nil
endfunc
`,
			expectedCommentCount: 3,
			expectedFirstComment: "# Line 1",
		},
		"Dash Comments": {
			script: `
-- Line 1
-- Line 2

func main() means
    -- In function
    set _ = nil
endfunc
`,
			expectedCommentCount: 3,
			expectedFirstComment: "-- Line 1",
		},
		"Slash Comments": {
			script: `
// Line 1
// Line 2

func main() means
    // In function
    set _ = nil
endfunc
`,
			expectedCommentCount: 3,
			expectedFirstComment: "// Line 1",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Logf("--- Source Code for TestCommentPrefixes/%s ---\n%s\n------------------", name, tc.script)
			logger := logging.NewTestLogger(t)
			parserAPI := NewParserAPI(logger)
			tree, tokenStream, errs := parserAPI.parseInternal(name+".ns", tc.script)
			if len(errs) > 0 {
				t.Fatalf("Parse failed with errors: %v", errs)
			}

			builder := NewASTBuilder(logger)
			program, _, err := builder.BuildFromParseResult(tree, tokenStream)
			if err != nil {
				t.Fatalf("Build failed unexpectedly: %v", err)
			}

			totalComments := len(program.Comments)
			if mainProc, ok := program.Procedures["main"]; ok {
				totalComments += len(mainProc.Comments)
				if len(mainProc.Steps) > 0 {
					totalComments += len(mainProc.Steps[0].Comments)
				}
			}

			if totalComments != tc.expectedCommentCount {
				t.Errorf("Expected %d total comments, but got %d", tc.expectedCommentCount, totalComments)
			}

			if len(program.Comments) > 0 {
				firstCommentText := strings.TrimSpace(program.Comments[0].Text)
				if firstCommentText != tc.expectedFirstComment {
					t.Errorf("Expected first comment to be '%s', but got '%s'", tc.expectedFirstComment, firstCommentText)
				}
			} else if tc.expectedCommentCount > 0 {
				t.Errorf("Expected program comments, but found none.")
			}
		})
	}
}
