// NeuroScript Version: 0.6.0
// File version: 5
// Purpose: Corrects assertions to align with "next-node" comment association.
// filename: pkg/parser/comment_prefix_test.go
// nlines: 87
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

			// --- FIX: Updated logic for "next-node" association ---
			totalComments := len(program.Comments) // Should be 0
			mainProc, ok := program.Procedures["main"]
			if !ok {
				t.Fatalf("Failed to find 'main' procedure")
			}

			totalComments += len(mainProc.Comments) // Should be 2 ("Line 1", "Line 2")
			if len(mainProc.Steps) > 0 {
				totalComments += len(mainProc.Steps[0].Comments) // Should be 1 ("In function")
			}
			// --- End Fix ---

			if totalComments != tc.expectedCommentCount {
				t.Errorf("Expected %d total comments, but got %d", tc.expectedCommentCount, totalComments)
			}

			// --- FIX: Check mainProc.Comments instead of program.Comments ---
			if len(mainProc.Comments) > 0 {
				firstCommentText := strings.TrimSpace(mainProc.Comments[0].Text)
				if firstCommentText != tc.expectedFirstComment {
					t.Errorf("Expected first comment to be '%s', but got '%s'", tc.expectedFirstComment, firstCommentText)
				}
			} else if tc.expectedCommentCount > 0 {
				t.Errorf("Expected procedure comments, but found none.")
			}
		})
	}
}
