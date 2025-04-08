// checklist_goyacc_test.go
package checklist

import (
	"reflect"
	"strings"
	"testing"
	// "fmt" // Uncomment for debug prints within tests
)

// Helper function for running tests
func runGoyaccTest(t *testing.T, testName string, input string, wantItems []CheckItem, wantMetadataLines MetadataLines) {
	t.Run(testName, func(t *testing.T) {
		lexer := NewChecklistLexer(strings.NewReader(input))
		// yyDebug = 1 // Uncomment for parser trace
		parseResult := yyParse(lexer)
		// yyDebug = 0

		// Check parser success code first
		if parseResult != 0 {
			// Lexer should have printed details via Error() method
			t.Fatalf("yyParse failed with code %d. Check console output for syntax errors.", parseResult)
		}

		// Check parsed items
		if !reflect.DeepEqual(lexer.Result, wantItems) {
			t.Errorf("Items mismatch:\nInput:\n---\n%s\n---\nGot : %#v\nWant: %#v", input, lexer.Result, wantItems)
		} else {
			t.Logf("Items OK: %#v", lexer.Result)
		}

		// Check collected metadata lines
		if !reflect.DeepEqual(lexer.MetadataLines, wantMetadataLines) {
			t.Errorf("MetadataLines mismatch:\nInput:\n---\n%s\n---\nGot : %#v\nWant: %#v", input, lexer.MetadataLines, wantMetadataLines)
		} else {
			t.Logf("MetadataLines OK: %#v", lexer.MetadataLines)
		}

		// Optional: Log full state if any mismatch occurred
		if t.Failed() {
			t.Logf("Full Lexer State on Failure: Result=%#v, MetadataLines=%#v", lexer.Result, lexer.MetadataLines)
		}
	})
}

// --- Test Cases ---

func TestGoyaccParseChecklist_Minimal(t *testing.T) {
	const input = `- [ ] Task 1
- [x] Task 2
`
	wantItems := []CheckItem{
		{Text: "Task 1", Status: "pending", Indent: 0},
		{Text: "Task 2", Status: "done", Indent: 0},
	}
	wantMetadataLines := []string{}
	runGoyaccTest(t, "Minimal_Goyacc_Checklist_Parse", input, wantItems, wantMetadataLines)
}

func TestGoyaccParseChecklist_WithSkippedComments(t *testing.T) {
	const input = `# This is a generic comment (skipped by lexer)
- [ ] First real task
# Another generic comment (skipped by lexer)
  - [x] Second real task indented with 2 spaces
# Generic comment at the end (skipped by lexer)
  # Indented generic comment (skipped by lexer)
`
	wantItems := []CheckItem{
		{Text: "First real task", Status: "pending", Indent: 0},
		{Text: "Second real task indented with 2 spaces", Status: "done", Indent: 2},
	}
	wantMetadataLines := []string{}
	runGoyaccTest(t, "Goyacc_Checklist_With_Skipped_Comments", input, wantItems, wantMetadataLines)
}

func TestGoyaccParseChecklist_WithIndentation(t *testing.T) {
	const input = `
- [ ] Item 1 (Indent 0)
  - [x] Item 2 (Indent 2 spaces)
	- [ ] Item 3 (Indent 1 tab - counted as 1 char here)
    - [X] Item 4 (Indent 4 spaces)
		- [ ] Item 5 (Indent 2 tabs - counted as 2 chars here)
  # Comment with indent - should be skipped by lexer
  - [ ] Item 6 (Indent 2 spaces) after comment
`
	wantItems := []CheckItem{
		{Text: "Item 1 (Indent 0)", Status: "pending", Indent: 0},
		{Text: "Item 2 (Indent 2 spaces)", Status: "done", Indent: 2},
		{Text: "Item 3 (Indent 1 tab - counted as 1 char here)", Status: "pending", Indent: 1},
		{Text: "Item 4 (Indent 4 spaces)", Status: "done", Indent: 4},
		{Text: "Item 5 (Indent 2 tabs - counted as 2 chars here)", Status: "pending", Indent: 2},
		{Text: "Item 6 (Indent 2 spaces) after comment", Status: "pending", Indent: 2},
	}
	wantMetadataLines := []string{}
	runGoyaccTest(t, "Goyacc_Checklist_With_Indentation", input, wantItems, wantMetadataLines)
}

func TestGoyaccParseChecklist_WithMetadata(t *testing.T) {
	// Using ':: ' prefix for metadata lines, ALL before items
	const input = `
:: id: test-checklist-123
:: version : 0.1.0
:: owner:  user_a
:: another-key : some other value with spaces
:: final-meta : end value
# This is a regular comment (skipped by lexer)

- [ ] Task Alpha
  - [x] Task Beta (indent 2)

# final comment (skipped)
`
	wantItems := []CheckItem{
		{Text: "Task Alpha", Status: "pending", Indent: 0},
		{Text: "Task Beta (indent 2)", Status: "done", Indent: 2},
	}
	// Expect raw metadata lines as stored by lexer
	wantMetadataLines := []string{
		":: id: test-checklist-123",
		":: version : 0.1.0",
		":: owner:  user_a",
		":: another-key : some other value with spaces",
		":: final-meta : end value",
	}
	runGoyaccTest(t, "Goyacc_Checklist_With_DoubleColon_Metadata", input, wantItems, wantMetadataLines)
}

func TestGoyaccParseChecklist_Empty(t *testing.T) {
	const input = ``
	wantItems := []CheckItem{}
	wantMetadataLines := []string{}
	runGoyaccTest(t, "Goyacc_Checklist_Empty_Input", input, wantItems, wantMetadataLines)
}

func TestGoyaccParseChecklist_MetaAndCommentsOnly(t *testing.T) {
	const input = `
:: key1: val1
# comment 1 (skipped)
:: key2 : val2
 # comment 2 indented (skipped)
`
	wantItems := []CheckItem{} // Expect no items
	wantMetadataLines := []string{
		":: key1: val1",
		":: key2 : val2",
	}
	runGoyaccTest(t, "Goyacc_Checklist_MetaCommentsOnly", input, wantItems, wantMetadataLines)
}
