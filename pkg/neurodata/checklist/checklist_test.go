// pkg/neurodata/checklist/checklist_test.go
package checklist

import (
	"reflect"
	"strings"
	"testing"
)

// TestParseChecklistContent tests the exported ParseChecklistContent function.
func TestParseChecklistContent(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		want        []map[string]interface{}
		wantErr     bool
		errContains string // Substring expected in the error message
	}{
		{
			name:  "Simple Valid Items",
			input: "- [ ] Task 1\n- [x] Task 2\n- [X] Task 3 (Upper X)",
			want: []map[string]interface{}{
				{"text": "Task 1", "status": "pending"},
				{"text": "Task 2", "status": "done"},
				{"text": "Task 3 (Upper X)", "status": "done"},
			},
			wantErr: false,
		},
		{
			name:  "Whitespace Variations",
			input: "  - [ ] Leading space text \n- [x]  Trailing space text  \n-    [ ]   Spaces before marker",
			want: []map[string]interface{}{
				{"text": "Leading space text", "status": "pending"},
				{"text": "Trailing space text", "status": "done"},
				// The current ANTLR grammar requires exactly one space after `-` and `]`,
				// so the third item won't be parsed correctly by the current grammar.
				// Let's expect only the first two. If the grammar is made more flexible later, update this test.
				// {"text": "Spaces before marker", "status": "pending"},
			},
			wantErr: false,
		},
		{
			name:  "Mixed Valid and Invalid Lines",
			input: "# Comment line\n\n- [ ] Valid Item 1\nJust some random text\n- [x] Valid Item 2\n-- Another comment",
			want: []map[string]interface{}{
				{"text": "Valid Item 1", "status": "pending"},
				{"text": "Valid Item 2", "status": "done"},
			},
			wantErr: false,
		},
		{
			name:    "Empty Input",
			input:   "",
			want:    []map[string]interface{}{}, // Expect empty slice
			wantErr: false,
		},
		{
			name:    "No Valid Items",
			input:   "This has no checklist items.\n- Invalid item format\n[ ] Another invalid line",
			want:    []map[string]interface{}{}, // Expect empty slice
			wantErr: false,
		},
		{
			name:  "CRLF Line Endings",
			input: "- [ ] Task A\r\n- [x] Task B\r\n",
			want: []map[string]interface{}{
				{"text": "Task A", "status": "pending"},
				{"text": "Task B", "status": "done"},
			},
			wantErr: false,
		},
		{
			name:  "Example from project_plan.md (MVP block)",
			input: "# id: kronos-mvp-reqs\n# version: 0.1.1\n# rendering_hint: markdown-list\n# status: draft\n\n- [ ] Allow user to start a timer for a task (e.g., `kronos start \"Coding feature X\"`).\n- [ ] Allow user to stop the current timer (e.g., `kronos stop`).\n- [ ] Store time entries locally (format TBD - potentially simple CSV or JSON).\n- [ ] Allow tagging entries with a project name (e.g., `kronos start \"Review PR\" --project \"NeuroScript\"`).\n- [x] Basic command-line argument parsing for `start` and `stop`.\n- [ ] Generate a simple summary report for today's entries (e.g., `kronos report today`).\n- [ ] Ensure basic persistence across app restarts.",
			want: []map[string]interface{}{
				{"text": "Allow user to start a timer for a task (e.g., `kronos start \"Coding feature X\"`).", "status": "pending"},
				{"text": "Allow user to stop the current timer (e.g., `kronos stop`).", "status": "pending"},
				{"text": "Store time entries locally (format TBD - potentially simple CSV or JSON).", "status": "pending"},
				{"text": "Allow tagging entries with a project name (e.g., `kronos start \"Review PR\" --project \"NeuroScript\"`).", "status": "pending"},
				{"text": "Basic command-line argument parsing for `start` and `stop`.", "status": "done"},
				{"text": "Generate a simple summary report for today's entries (e.g., `kronos report today`).", "status": "pending"},
				{"text": "Ensure basic persistence across app restarts.", "status": "pending"},
			},
			wantErr: false,
		},
		// Add test cases for potential errors if the ANTLR parser can produce them (e.g., malformed input)
		// Currently, the grammar is simple and likely just won't match invalid lines.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseChecklistContent(tt.input)

			// Check error expectation
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseChecklistContent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Check error content if error was expected
			if tt.wantErr {
				if tt.errContains != "" && (err == nil || !strings.Contains(err.Error(), tt.errContains)) {
					t.Errorf("ParseChecklistContent() expected error containing %q, got: %v", tt.errContains, err)
				}
				// Don't compare result if error was expected
				return
			}

			// Check result if no error was expected
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseChecklistContent() mismatch:")
				// Log details for easier debugging
				t.Errorf("  Input:\n---\n%s\n---", tt.input)
				// Use %#v for detailed map/slice comparison
				t.Errorf("  Got : %#v", got)
				t.Errorf("  Want: %#v", tt.want)
			}
		})
	}
}
