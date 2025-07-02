// filename: pkg/tool/git/tools_git_status_test.go
package git

import (
	"reflect"	// For DeepEqual
	"testing"
	// Add other necessary standard library imports if parseGitStatusOutput uses them indirectly
)

func TestParseGitStatusOutput(t *testing.T) {
	testCases := []struct {
		name		string
		input		string
		expectedMap	map[string]interface{}
		expectParseErr	bool	// Expect a Go error from the parser function itself (should be rare)
	}{
		// Using the same examples as before
		{
			name:	"Clean Repository",
			input: `## main...origin/main
`,
			expectedMap: map[string]interface{}{
				"branch":			"main",
				"remote_branch":		"origin/main",
				"ahead":			int64(0),
				"behind":			int64(0),
				"files":			[]map[string]interface{}{},
				"untracked_files_present":	false,
				"is_clean":			true,
				"error":			nil,
			},
			expectParseErr:	false,
		},
		{
			name:	"Clean Repository (No Remote)",
			input: `## main
`,
			expectedMap: map[string]interface{}{
				"branch":			"main",
				"remote_branch":		nil,
				"ahead":			int64(0),
				"behind":			int64(0),
				"files":			[]map[string]interface{}{},
				"untracked_files_present":	false,
				"is_clean":			true,
				"error":			nil,
			},
			expectParseErr:	false,
		},
		{
			name:	"Modified File (Worktree)",
			input: `## main...origin/main
 M go.mod
`,
			expectedMap: map[string]interface{}{
				"branch":		"main",
				"remote_branch":	"origin/main",
				"ahead":		int64(0),
				"behind":		int64(0),
				"files": []map[string]interface{}{
					{"path": "go.mod", "index_status": " ", "worktree_status": "M", "original_path": nil},
				},
				"untracked_files_present":	false,
				"is_clean":			false,
				"error":			nil,
			},
			expectParseErr:	false,
		},
		{
			name:	"Staged Add & Untracked & Ahead",
			input: `## feature/new...origin/feature/new [ahead 1]
A  new_file.txt
?? another_untracked.go
`,
			expectedMap: map[string]interface{}{
				"branch":		"feature/new",
				"remote_branch":	"origin/feature/new",
				"ahead":		int64(1),
				"behind":		int64(0),
				"files": []map[string]interface{}{
					{"path": "new_file.txt", "index_status": "A", "worktree_status": " ", "original_path": nil},
					{"path": "another_untracked.go", "index_status": "?", "worktree_status": "?", "original_path": nil},
				},
				"untracked_files_present":	true,
				"is_clean":			false,	// Staged change makes it not clean
				"error":			nil,
			},
			expectParseErr:	false,
		},
		{
			name:	"Staged Modify, Unstaged Delete, Untracked, Behind",
			input: `## main...origin/main [behind 2]
M  go.mod
 D deleted_file.txt
?? untracked.md
`,
			expectedMap: map[string]interface{}{
				"branch":		"main",
				"remote_branch":	"origin/main",
				"ahead":		int64(0),
				"behind":		int64(2),
				"files": []map[string]interface{}{
					{"path": "go.mod", "index_status": "M", "worktree_status": " ", "original_path": nil},
					{"path": "deleted_file.txt", "index_status": " ", "worktree_status": "D", "original_path": nil},	// Unstaged delete is worktree 'D'
					{"path": "untracked.md", "index_status": "?", "worktree_status": "?", "original_path": nil},
				},
				"untracked_files_present":	true,
				"is_clean":			false,	// Staged and unstaged changes
				"error":			nil,
			},
			expectParseErr:	false,
		},
		{
			name:	"Renamed (Staged)",
			input: `## main
R  new_name.txt -> old_name.txt
`,
			expectedMap: map[string]interface{}{
				"branch":		"main",
				"remote_branch":	nil,
				"ahead":		int64(0),
				"behind":		int64(0),
				"files": []map[string]interface{}{
					{"path": "new_name.txt", "index_status": "R", "worktree_status": " ", "original_path": "old_name.txt"},
				},
				"untracked_files_present":	false,
				"is_clean":			false,	// Staged change
				"error":			nil,
			},
			expectParseErr:	false,
		},
		{
			name:	"Unborn Branch",
			input: `## No commits yet on main
`,
			expectedMap: map[string]interface{}{
				"branch":			"main",
				"remote_branch":		nil,
				"ahead":			int64(0),
				"behind":			int64(0),
				"files":			[]map[string]interface{}{},
				"untracked_files_present":	false,
				"is_clean":			true,	// Unborn is clean
				"error":			nil,
			},
			expectParseErr:	false,
		},
		{
			name:	"Detached HEAD",
			input: `## HEAD (no branch)
 M go.mod
`,
			expectedMap: map[string]interface{}{
				"branch":		"(detached HEAD)",
				"remote_branch":	nil,
				"ahead":		int64(0),
				"behind":		int64(0),
				"files": []map[string]interface{}{
					{"path": "go.mod", "index_status": " ", "worktree_status": "M", "original_path": nil},
				},
				"untracked_files_present":	false,
				"is_clean":			false,	// Changed file
				"error":			nil,
			},
			expectParseErr:	false,
		},
		{
			name:	"Path with Spaces (Quoted)",
			input: `## main
 M "path with spaces/file name.txt"
?? "untracked with spaces.log"
`,
			expectedMap: map[string]interface{}{
				"branch":		"main",
				"remote_branch":	nil,
				"ahead":		int64(0),
				"behind":		int64(0),
				"files": []map[string]interface{}{
					{"path": "path with spaces/file name.txt", "index_status": " ", "worktree_status": "M", "original_path": nil},
					{"path": "untracked with spaces.log", "index_status": "?", "worktree_status": "?", "original_path": nil},
				},
				"untracked_files_present":	true,
				"is_clean":			false,	// Changed file
				"error":			nil,
			},
			expectParseErr:	false,
		},
		{
			name:	"Empty Input String",	// Should be treated as clean
			input:	``,
			expectedMap: map[string]interface{}{
				"branch":			nil,
				"remote_branch":		nil,
				"ahead":			int64(0),
				"behind":			int64(0),
				"files":			[]map[string]interface{}{},
				"untracked_files_present":	false,
				"is_clean":			true,	// Empty means clean
				"error":			nil,
			},
			expectParseErr:	false,
		},
		{
			name:	"Malformed Branch Line",	// Parser should set error in map
			input: `# Not a branch line
 M file.txt`,
			expectedMap: map[string]interface{}{
				"branch":		nil,
				"remote_branch":	nil,
				"ahead":		int64(0),
				"behind":		int64(0),
				"files": []map[string]interface{}{	// Still parses files if possible
					{"path": "file.txt", "index_status": " ", "worktree_status": "M", "original_path": nil},
				},
				"untracked_files_present":	false,
				"is_clean":			false,									// Change detected
				"error":			"Failed to parse branch information from line: # Not a branch line",	// Expect error set in map
			},
			expectParseErr:	false,	// The function itself doesn't return Go error
		},
		{
			name:	"Only Branch Line (Ahead)",	// Test boundary condition
			input:	`## dev...origin/dev [ahead 3]`,
			expectedMap: map[string]interface{}{
				"branch":			"dev",
				"remote_branch":		"origin/dev",
				"ahead":			int64(3),
				"behind":			int64(0),
				"files":			[]map[string]interface{}{},
				"untracked_files_present":	false,
				"is_clean":			true,	// No file changes listed means clean
				"error":			nil,
			},
			expectParseErr:	false,
		},
	}

	for _, tc := range testCases {
		tc := tc	// Capture range variable
		t.Run(tc.name, func(t *testing.T) {
			actualMap, err := parseGitStatusOutput(tc.input)

			// Check the Go error returned by the function
			if tc.expectParseErr {
				if err == nil {
					t.Errorf("Expected a Go error from parseGitStatusOutput, but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected Go error to be nil, but got: %v", err)
				}
			}

			// Now compare the returned map structure using reflect.DeepEqual
			// for the overall structure, and potentially individual fields for clarity
			if !reflect.DeepEqual(tc.expectedMap, actualMap) {
				t.Errorf("Result map mismatch:")
				// Provide more detailed diff if needed (could compare field by field)
				for key, expectedValue := range tc.expectedMap {
					actualValue, keyExists := actualMap[key]
					if !keyExists {
						t.Errorf("  Missing key: %s", key)
						continue
					}
					// Special check for 'files' slice using ElementsMatch semantics
					if key == "files" {
						expectedFiles, okE := expectedValue.([]map[string]interface{})
						actualFiles, okA := actualValue.([]map[string]interface{})
						if !okE || !okA {
							t.Errorf("  Type mismatch for key '%s': expected []map[string]interface{}, got %T and %T", key, expectedValue, actualValue)
							continue
						}
						// Simple length check first
						if len(expectedFiles) != len(actualFiles) {
							t.Errorf("  Length mismatch for key '%s': expected %d, got %d", key, len(expectedFiles), len(actualFiles))
							// Print lists for easier debugging
							t.Errorf("    Expected Files: %+v", expectedFiles)
							t.Errorf("    Actual Files:   %+v", actualFiles)
						} else if !compareFileMaps(expectedFiles, actualFiles) {	// Use helper for element comparison
							t.Errorf("  Element mismatch for key '%s'", key)
							t.Errorf("    Expected Files: %+v", expectedFiles)
							t.Errorf("    Actual Files:   %+v", actualFiles)
						}

					} else if !reflect.DeepEqual(expectedValue, actualValue) {
						t.Errorf("  Value mismatch for key '%s': expected '%v' (%T), got '%v' (%T)", key, expectedValue, expectedValue, actualValue, actualValue)
					}
				}
				// Check for extra keys in actual map
				for key := range actualMap {
					if _, keyExists := tc.expectedMap[key]; !keyExists {
						t.Errorf("  Extra key found in actual map: %s", key)
					}
				}
			}
		})
	}
}

// compareFileMaps checks if two slices of file status maps contain the same elements,
// ignoring order. This is a basic implementation for ElementsMatch semantics.
func compareFileMaps(expected, actual []map[string]interface{}) bool {
	if len(expected) != len(actual) {
		return false
	}
	if len(expected) == 0 {	// Both empty
		return true
	}

	actualCopy := make([]map[string]interface{}, len(actual))
	copy(actualCopy, actual)

	for _, expMap := range expected {
		foundMatch := false
		for i, actMap := range actualCopy {
			if reflect.DeepEqual(expMap, actMap) {
				// Remove found element from copy to handle duplicates correctly
				actualCopy = append(actualCopy[:i], actualCopy[i+1:]...)
				foundMatch = true
				break
			}
		}
		if !foundMatch {
			return false	// Expected map not found in actual
		}
	}

	return len(actualCopy) == 0	// Should be empty if all expected maps were matched
}