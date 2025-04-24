// filename: pkg/core/universal_test_helpers_test.go
package core

import (
	// Needed for scanner in DiffStrings test
	"fmt"
	"strings"
	"testing"
)

// --- Tests for NormalizeString ---

func TestNormalizeString(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		flags    uint32
		expected string
	}{
		{
			name: "No flags (expect default)",
			input: ` line 1 // go comment
			  line 2 # ns comment

			line 4   with   spaces `,
			flags: 0, // Uses NormDefault
			expected: `line 1
line 2
line 4 with spaces`,
		},
		{
			name: "TrimSpace only",
			input: `  line 1  // comment stays
			line 2 # comment stays

			  line 4   `,
			flags: NormTrimSpace,
			expected: `line 1  // comment stays
line 2 # comment stays

line 4`,
		},
		{
			name: "CompressSpace only",
			input: ` line   1 // comment   stays
			  line 2    # comment stays

			line 4   with   spaces `,
			flags: NormCompressSpace, // Implies NormTrimSpace
			// --- FIX V8: Expect TrimSpace to be implied, Blank lines NOT removed ---
			expected: `line 1 // comment stays
line 2 # comment stays

line 4 with spaces`, // Blank line stays as NormRemoveBlankLines not set
		},
		{
			name: "RemoveGoComments only",
			input: ` line 1 // go comment
			  line 2 # ns comment stays

			line 4   `,
			flags: NormRemoveGoComments,
			expected: ` line 1 
			  line 2 # ns comment stays

			line 4   `,
		},
		{
			name: "RemoveNSComments only",
			input: ` line 1 // go comment stays
			  line 2 # ns comment

			line 4   `,
			flags: NormRemoveNSComments,
			expected: ` line 1 // go comment stays
			  line 2 

			line 4   `,
		},
		{
			name: "RemoveBlankLines only",
			input: `line 1

			line 3
			  // comment line only (whitespace)
			# ns comment line only (whitespace)
			  \t  
			line 7`,
			flags: NormRemoveBlankLines,
			// --- FIX V8: TrimSpace is OFF, so only initially blank lines removed ---
			expected: `line 1
			line 3
			  // comment line only (whitespace)
			# ns comment line only (whitespace)
			  \t  
			line 7`, // Lines 1 and 5 (0-indexed, originally empty) were removed. Others remain.
		},
		{
			name: "Remove Comments and Blank Lines",
			input: `line 1 // go comment
			# ns comment line (becomes blank)
			line 3

			  \t  // Line with only whitespace (Keep if Trim OFF)

			line 7`,
			flags: NormRemoveGoComments | NormRemoveNSComments | NormRemoveBlankLines, // No TrimSpace
			// --- FIX V8: Based on refined implementation (TrimSpace is OFF here) ---
			expected: `line 1 
			line 3
			  \t  
			line 7`, // Line 2 becomes "", removed. Line 4 empty, removed. Line 5 has "\t", kept. Line 6 empty, removed.
		},
		{
			name: "Remove Comments, Blank Lines, AND Trim", // More common combo
			input: `line 1 // go comment
			# ns comment line
			line 3

			  \t  // Line with only whitespace

			line 7`,
			flags: NormRemoveGoComments | NormRemoveNSComments | NormRemoveBlankLines | NormTrimSpace,
			// --- FIX V8: Based on refined implementation ---
			expected: `line 1
line 3
line 7`, // Line with only tab is REMOVED because TrimSpace makes it blank first
		},
		{
			name: "All flags (NormDefault)",
			input: `
			  first line   // go comment here
			second  line # ns comment here

			  third line with   extra spaces


			fourth line ends.  `,
			flags: NormDefault,
			expected: `first line
second line
third line with extra spaces
fourth line ends.`,
		},
		{
			name:     "Empty Input",
			input:    "",
			flags:    NormDefault,
			expected: "",
		},
		{
			name:     "Whitespace only input",
			input:    " \n \t \n ",
			flags:    NormDefault,
			expected: "", // Default includes Trim and RemoveBlanks
		},
		{
			name:     "Input with only comments",
			input:    "// line 1\n# line 2\n  // line 3",
			flags:    NormDefault,
			expected: "", // Default includes comment removal, trim, blank removal
		},
		{
			name:  "No normalization needed",
			input: "line 1\nline 2",
			flags: NormDefault,
			expected: `line 1
line 2`,
		},
		{
			name:  "Windows Line Endings",
			input: "line 1\r\nline 2\r\n",
			flags: NormDefault,
			expected: `line 1
line 2`,
		},
		{
			name:  "Mixed Line Endings with Space",
			input: "line 1 \r\n  line 2 \n line 3 \r\n",
			flags: NormDefault,
			expected: `line 1
line 2
line 3`,
		},
		{
			name:     "Tab handling",
			input:    "line\t1\nline\t\t2",             // Input with tabs
			flags:    NormDefault &^ NormCompressSpace, // Keep tabs, but trim/etc
			expected: "line\t1\nline\t\t2",             // Expect tabs to be preserved
		},
		{
			name:     "Tab handling with compression",
			input:    "line\t1\nline\t\t2", // Input with tabs
			flags:    NormDefault,          // Compress implies Trim
			expected: "line 1\nline 2",     // Expect tabs to be compressed like spaces
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := NormalizeString(tc.input, tc.flags)
			if actual != tc.expected {
				// Pass 't' directly here, not mockT
				t.Logf("NormalizeString failed for case: %s", tc.name)
				DiffStrings(t, tc.expected, actual, 0, DiffAnsiColor|DiffVisibleSpace)
				t.Fail() // Mark test as failed
			}
		})
	}
}

// --- Tests for DiffStrings ---

// mockTB (remains the same)
type mockTB struct {
	testing.TB
	logs   []string
	errors []string
	failed bool
}

func newMockTB(t *testing.T) *mockTB { return &mockTB{TB: t} }
func (m *mockTB) Logf(format string, args ...interface{}) {
	s := fmt.Sprintf(format, args...)
	m.logs = append(m.logs, s)
}
func (m *mockTB) Errorf(format string, args ...interface{}) {
	s := fmt.Sprintf(format, args...)
	m.errors = append(m.errors, s)
	m.failed = true
}
func (m *mockTB) FailNow()     { m.failed = true }
func (m *mockTB) Fail()        { m.failed = true }
func (m *mockTB) Failed() bool { return m.failed }
func (m *mockTB) Helper()      {}

func TestDiffStrings(t *testing.T) {
	t.Run("Equal strings after normalization", func(t *testing.T) {
		mockT := newMockTB(t)
		expected := "line 1 // comment\n  line 2"
		actual := "line 1\nline 2   # comment"
		areEqual := DiffStrings(mockT, expected, actual, NormDefault, 0)

		if !areEqual {
			t.Errorf("FAIL: Expected DiffStrings to return true for equal strings, but got false")
		}
		if mockT.Failed() {
			t.Errorf("FAIL: Expected DiffStrings not to mark test as failed for equal strings, but it did.")
		}
	})

	t.Run("Different strings after normalization", func(t *testing.T) {
		mockT := newMockTB(t)
		expected := "line 1\nline 2"
		actual := "line 1\nline THREE"
		areEqual := DiffStrings(mockT, expected, actual, NormDefault, 0)

		if areEqual {
			t.Errorf("FAIL: Expected DiffStrings to return false for different strings, but got true")
		}
		if !mockT.Failed() {
			t.Errorf("FAIL: Expected DiffStrings to mark test as failed for different strings, but it didn't.")
		}

		hasDiffLog := false
		for _, log := range mockT.logs {
			if strings.Contains(log, "--- Diff ---") {
				hasDiffLog = true
				break
			}
		}
		if !hasDiffLog {
			t.Errorf("FAIL: Expected DiffStrings to log the diff output, but it didn't seem to.")
		}
	})

	t.Run("DiffShowFull flag", func(t *testing.T) {
		mockT := newMockTB(t)
		expected := "line 1"
		actual := "line 2"
		areEqual := DiffStrings(mockT, expected, actual, 0, DiffShowFull)

		if areEqual {
			t.Errorf("FAIL: Expected DiffStrings to return false for different strings, but got true")
		}
		if !mockT.Failed() {
			t.Errorf("FAIL: Expected DiffStrings to mark test as failed for different strings, but it didn't.")
		}

		hasExpectedLog := false
		hasActualLog := false
		for _, log := range mockT.logs {
			if strings.Contains(log, "--- Expected (Original) ---") {
				hasExpectedLog = true
			}
			if strings.Contains(log, "--- Actual (Original) ---") {
				hasActualLog = true
			}
		}
		if !hasExpectedLog || !hasActualLog {
			t.Errorf("FAIL: Expected DiffStrings with DiffShowFull to log original strings, but it didn't.")
		}
	})

	t.Run("DiffAnsiColor flag runs", func(t *testing.T) {
		mockT := newMockTB(t)
		expected := "line 1"
		actual := "line 2"
		areEqual := DiffStrings(mockT, expected, actual, 0, DiffAnsiColor)

		if areEqual {
			t.Errorf("FAIL: Expected DiffStrings to return false for different strings, but got true")
		}
		if !mockT.Failed() {
			t.Errorf("FAIL: Expected DiffStrings to mark test as failed for different strings, but it didn't.")
		}
		hasAnsi := false
		for _, log := range mockT.logs {
			if strings.Contains(log, "\x1b[") {
				hasAnsi = true
				break
			}
		}
		if !hasAnsi {
			t.Errorf("FAIL: Expected ANSI codes in log output for DiffAnsiColor, but none found.")
		}
	})

	t.Run("DiffVisibleSpace flag shows space tab nl", func(t *testing.T) {
		mockT := newMockTB(t)
		// Input strings designed to show different whitespace chars after normalization
		expected := "line with space\t and cr\r\nnext line" // \r\n becomes \n
		actual := "line with  extra   space\t\nnext line"   // \n ok, extra spaces
		// Use flags that keep internal spacing variations but remove comments/blanks for clearer diff
		normFlags := NormDefault &^ NormCompressSpace

		areEqual := DiffStrings(mockT, expected, actual, normFlags, DiffAnsiColor|DiffVisibleSpace)

		if areEqual {
			t.Errorf("FAIL: Expected DiffStrings to return false for different strings, but got true")
		}
		if !mockT.Failed() {
			t.Errorf("FAIL: Expected DiffStrings to mark test as failed for different strings, but it didn't.")
		}

		diffLogged := false
		logFound := ""
		for _, log := range mockT.logs {
			if strings.Contains(log, "--- Diff ---") {
				diffLogged = true
				logFound = log
				break
			}
		}

		if !diffLogged {
			t.Errorf("FAIL: DiffStrings did not log a diff block.")
		} else {
			// --- FIX: Check for space, tab, and newline; NOT CR ---
			symbolsExpected := []string{spaceSym, tabSym, nlSym} // Check for ␣, ␉, ␤
			allFound := true
			symbolsMissing := []string{}
			for _, sym := range symbolsExpected {
				// Check if the specific *colored* symbol exists in the log
				coloredSym := colorVisibleWs + sym + colorReset
				if !strings.Contains(logFound, coloredSym) {
					allFound = false
					symbolsMissing = append(symbolsMissing, sym)
				}
			}
			// Explicitly check CR symbol is NOT present
			coloredCRSym := colorVisibleWs + crSym + colorReset
			if strings.Contains(logFound, coloredCRSym) {
				t.Errorf("FAIL: DiffStrings with DiffVisibleSpace showed '%s', but CRs should be normalized away.", crSym)
				allFound = false // Mark as failed if CR is unexpectedly present
			}

			if !allFound {
				t.Errorf("FAIL: Expected DiffStrings with DiffVisibleSpace to show symbols ('%s'), but missing: %v.",
					strings.Join(symbolsExpected, "', '"), strings.Join(symbolsMissing, ", "))
				t.Logf("Relevant Log Chunk:\n%s", logFound)
			}
		}
	})
}
