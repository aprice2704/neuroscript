// pkg/neurodata/checklist/scanner_parser_test.go
package checklist

import (
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

var testlogger = slog.New(slog.NewTextHandler(os.Stderr, nil))

// Helper to compare Checklists, ignoring line numbers
func checklistsEqual(t *testing.T, got, want *ParsedChecklist) bool {
	t.Helper()
	if got == nil && want == nil {
		return true
	}
	if got == nil || want == nil {
		t.Errorf("Nil mismatch: got=%v, want=%v", got == nil, want == nil)
		return false
	}
	if !reflect.DeepEqual(got.Metadata, want.Metadata) {
		t.Errorf("Metadata mismatch:\n got: %#v\nwant: %#v", got.Metadata, want.Metadata)
		return false
	}
	if len(got.Items) != len(want.Items) {
		t.Errorf("Item count mismatch: got %d, want %d", len(got.Items), len(want.Items))
		t.Logf("Got Items: %#v", got.Items)
		t.Logf("Want Items: %#v", want.Items)
		return false
	}
	equal := true
	for i := range got.Items {
		if i >= len(want.Items) {
			break
		}
		gotItem := got.Items[i]
		wantItem := want.Items[i]
		if gotItem.Text != wantItem.Text ||
			gotItem.Status != wantItem.Status ||
			gotItem.Symbol != wantItem.Symbol ||
			gotItem.Indent != wantItem.Indent ||
			gotItem.IsAutomatic != wantItem.IsAutomatic {
			t.Errorf("Item %d mismatch:\n got: {Text:%q Status:%q Symbol:'%c' Indent:%d Auto:%t Line:%d}\nwant: {Text:%q Status:%q Symbol:'%c' Indent:%d Auto:%t Line:IGNORED}",
				i,
				gotItem.Text, gotItem.Status, gotItem.Symbol, gotItem.Indent, gotItem.IsAutomatic, gotItem.LineNumber,
				wantItem.Text, wantItem.Status, wantItem.Symbol, wantItem.Indent, wantItem.IsAutomatic,
			)
			equal = false
		}
	}
	return equal
}

func TestParseChecklistScannerFixtures(t *testing.T) {

	fixtureBaseDir := "test_fixtures"

	tests := []struct {
		name             string
		fixtureFile      string
		want             *ParsedChecklist // Expected result if no error
		wantErr          bool             // Error reading file
		wantParseErr     bool             // Error during ParseChecklist call
		wantParseErrType error            // Specific error type expected
	}{
		// --- V12: Updated Empty Content Test ---
		{
			name:             "Empty Content - EXPECT PARSE ERROR", // Renamed
			fixtureFile:      "empty.txt",
			want:             nil, // Expect nil result because of error
			wantErr:          false,
			wantParseErr:     true,         // Expect ParseChecklist to return an error now
			wantParseErrType: ErrNoContent, // Expect the specific defined error
		},
		// --- V12: Added Whitespace/Comment Only Test ---
		{
			name:             "Only Whitespace and Comments - EXPECT PARSE ERROR",
			fixtureFile:      "whitespace_comments.txt", // Needs creating
			want:             nil,
			wantErr:          false,
			wantParseErr:     true,
			wantParseErrType: ErrNoContent,
		},
		// --- Happy Path / Valid Input Cases ---
		{
			name:        "Basic Items",
			fixtureFile: "basic_items.txt",
			want: &ParsedChecklist{
				Items: []ChecklistItem{
					{Text: "Task 1", Status: "pending", Symbol: ' '},
					{Text: "Task 2", Status: "done", Symbol: 'x'},
					{Text: "Task 3", Status: "done", Symbol: 'x'},
				},
				Metadata: map[string]string{},
			},
			wantErr: false, wantParseErr: false,
		},
		{
			name:        "Partial and Special Status",
			fixtureFile: "partial_special.txt",
			want: &ParsedChecklist{
				Items: []ChecklistItem{
					{Text: "Partial", Status: "partial", Symbol: '-'},
					{Text: "Needs Info", Status: "special", Symbol: '?'},
					{Text: "Blocked", Status: "special", Symbol: '!'},
				},
				Metadata: map[string]string{},
			},
			wantErr: false, wantParseErr: false,
		},
		{
			name:        "Automatic Marker (| |)",
			fixtureFile: "automatic.txt",
			want: &ParsedChecklist{
				Items: []ChecklistItem{
					{Text: "Parent Task", Status: "pending", Symbol: ' ', Indent: 0, IsAutomatic: true},
					{Text: "Child", Status: "done", Symbol: 'x', Indent: 2, IsAutomatic: false},
				},
				Metadata: map[string]string{},
			},
			wantErr: false, wantParseErr: false,
		},
		// --- V11 Passed: This test now passes with string manipulation ---
		{
			name:        "Indentation and Whitespace",
			fixtureFile: "indentation.txt",
			want: &ParsedChecklist{
				Items: []ChecklistItem{
					{Text: "Item 1", Status: "pending", Symbol: ' ', Indent: 0},
					{Text: "Item 1.1", Status: "done", Symbol: 'x', Indent: 2},
					{Text: "Item 1.1.1", Status: "partial", Symbol: '-', Indent: 4},
					{Text: "Item 2", Status: "pending", Symbol: ' ', Indent: 0},
				},
				Metadata: map[string]string{},
			},
			wantErr: false, wantParseErr: false,
		},
		{
			name:        "Metadata, Comments, Headings",
			fixtureFile: "metadata_comments.txt",
			want: &ParsedChecklist{
				Items: []ChecklistItem{
					{Text: "Item A", Status: "pending", Symbol: ' ', Indent: 0},
					{Text: "Item B (Special)", Status: "special", Symbol: '?', Indent: 2},
					{Text: "Item C", Status: "done", Symbol: 'x', Indent: 0},
				},
				Metadata: map[string]string{"version": "1.1", "type": "Checklist"},
			},
			wantErr: false, wantParseErr: false,
		},
		{
			name:        "Checklist stops at non-item/meta/comment/heading",
			fixtureFile: "stops_at_content.txt",
			want: &ParsedChecklist{
				Items: []ChecklistItem{
					{Text: "Item 1", Status: "pending", Symbol: ' ', Indent: 0},
				},
				Metadata: map[string]string{"key": "value"},
			},
			wantErr: false, wantParseErr: false,
		},
		{
			name:        "Malformed Brackets (empty)",
			fixtureFile: "malformed_empty.txt",
			want: &ParsedChecklist{
				Items:    []ChecklistItem{{Text: "Empty", Status: "pending", Symbol: ' '}},
				Metadata: map[string]string{},
			},
			wantErr: false, wantParseErr: false,
		},
		{
			name:        "Malformed Pipes (empty)",
			fixtureFile: "malformed_empty_pipes.txt",
			want: &ParsedChecklist{
				Items:    []ChecklistItem{{Text: "Empty Pipe", Status: "pending", Symbol: ' ', IsAutomatic: true}},
				Metadata: map[string]string{},
			},
			wantErr: false, wantParseErr: false,
		},
		// --- Error Expectation Cases ---
		{
			name:             "Malformed Brackets (multi-char) - EXPECT PARSE ERROR",
			fixtureFile:      "malformed_multichar.txt",
			want:             nil,
			wantErr:          false,
			wantParseErr:     true,
			wantParseErrType: ErrMalformedItem,
		},
		{
			name:             "Malformed Pipes (multi-char) - EXPECT PARSE ERROR",
			fixtureFile:      "malformed_pipes.txt",
			want:             nil,
			wantErr:          false,
			wantParseErr:     true,
			wantParseErrType: ErrMalformedItem,
		},
		// --- File System Error Case ---
		{
			name:         "File Not Found",
			fixtureFile:  "nonexistent_file.txt",
			want:         nil,
			wantErr:      true, // Expect error from ReadFile
			wantParseErr: false,
		},
	}

	// --- Test Setup ---
	// Ensure base fixture directory exists
	if _, err := os.Stat(fixtureBaseDir); os.IsNotExist(err) {
		if mkErr := os.Mkdir(fixtureBaseDir, 0755); mkErr != nil {
			t.Fatalf("Failed to create fixture base dir %s: %v", fixtureBaseDir, mkErr)
		}
	}
	// Create/ensure required fixture files exist before running tests
	fixtureContentMap := map[string]string{
		"empty.txt":                 "",                                     // For ErrNoContent test
		"whitespace_comments.txt":   "\n  # Comment\n\n-- Another\n  \t \n", // For ErrNoContent test
		"basic_items.txt":           "- [ ] Task 1\n- [x] Task 2\n- [X] Task 3",
		"partial_special.txt":       "- [-] Partial\n- [?] Needs Info\n- [!] Blocked",
		"automatic.txt":             "- | | Parent Task\n  - [x] Child",
		"indentation.txt":           "\n- [ ] Item 1\n  - [x] Item 1.1\n    - [-] Item 1.1.1\n- [ ] Item 2\n",
		"metadata_comments.txt":     ":: version: 1.1\n:: type: Checklist\n\n# Section 1\n- [ ] Item A\n-- Some comment\n  # Subsection 1.1\n  - [?] Item B (Special)\n\n# Section 2 (Empty)\n\n- [x] Item C",
		"stops_at_content.txt":      ":: key: value\n- [ ] Item 1\nThis is just plain text.\n- [x] Item 2 (will not be parsed)",
		"malformed_multichar.txt":   "- [xx] Malformed",
		"malformed_pipes.txt":       "- |xx| Malformed Pipe",
		"malformed_empty.txt":       "- [] Empty",
		"malformed_empty_pipes.txt": "- || Empty Pipe",
	}
	for name, content := range fixtureContentMap {
		fp := filepath.Join(fixtureBaseDir, name)
		if errWrite := os.WriteFile(fp, []byte(content), 0644); errWrite != nil {
			t.Fatalf("Failed to write setup fixture %s: %v", fp, errWrite)
		}
	}
	// --- End Test Setup ---

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixturePath := filepath.Join(fixtureBaseDir, tt.fixtureFile)

			// Read File
			contentBytes, readErr := os.ReadFile(fixturePath)
			if (readErr != nil) != tt.wantErr {
				t.Fatalf("ReadFile(%q) error = %v, wantErr %v", fixturePath, readErr, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if readErr != nil && !tt.wantErr {
				t.Fatalf("Unexpected error reading fixture %q: %v", fixturePath, readErr)
			}

			// Parse Content
			content := string(contentBytes)
			got, parseErr := ParseChecklist(content, testlogger)

			// Check Parse Error Expectation
			if tt.wantParseErr {
				if parseErr == nil {
					t.Fatalf("ParseChecklist() expected an error but got nil")
				}
				if tt.wantParseErrType != nil {
					if !errors.Is(parseErr, tt.wantParseErrType) {
						t.Fatalf("ParseChecklist() error type mismatch:\n Got error: %v (%T)\nWant error type: %v", parseErr, parseErr, tt.wantParseErrType)
					}
					t.Logf("ParseChecklist() correctly returned expected error type: %v", parseErr)
				} else {
					t.Logf("ParseChecklist() correctly returned an error: %v", parseErr)
				}
			} else { // Expect NO parse error
				if parseErr != nil {
					t.Fatalf("ParseChecklist() unexpected error: %v", parseErr)
				}

				// Compare Results
				if tt.want == nil {
					if got != nil && (len(got.Items) > 0 || len(got.Metadata) > 0) {
						t.Errorf("Expected nil or empty result, but got: %+v", *got)
					}
				} else {
					if got == nil {
						t.Errorf("Expected non-nil result (items/metadata), but got nil")
					} else {
						for i := range tt.want.Items {
							tt.want.Items[i].LineNumber = 0
						}
						if !checklistsEqual(t, got, tt.want) {
							t.Logf("Comparison failed for fixture: %s", tt.fixtureFile)
						}
					}
				}
			}
		})
	}

	// Cleanup created fixture files? Or rely on test framework?
	// Let's add cleanup for the files we created explicitly in setup.
	// Note: This runs *after* all subtests.
	for name := range fixtureContentMap {
		fp := filepath.Join(fixtureBaseDir, name)
		os.Remove(fp) // Ignore error on cleanup
	}

}
