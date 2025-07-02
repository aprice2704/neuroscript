// filename: pkg/testutil/universal_test_helpers_test.go
package testutil

import (
	// Needed for scanner in DiffStrings test
	"fmt"
	"strings"
	"testing"
)

// --- Tests for NormalizeString ---

func TestNormalizeString(t *testing.T) {
	testCases := []struct {
		name		string
		input		string
		flags		NormalizationFlags
		expected	string
	}{
		{
			name:	"No flags (expect default)",
			input: ` line 1 // go comment
			  line 2 # ns comment

			line 4   with   spaces `,
			flags:	0,	// Uses NormDefault
			expected: `line 1
line 2
line 4 with spaces`,
		},
		{
			name:	"TrimSpace only",
			input: `  line 1  // comment stays
			line 2 # comment stays

			  line 4   `,
			flags:	NormTrimSpace,
			expected: `line 1  // comment stays
line 2 # comment stays

line 4`,
		},
		{
			name:	"CompressSpace only",
			input: ` line   1 // comment   stays
			  line 2    # comment stays

			line 4   with   spaces `,
			flags:	NormCompressSpace,	// Implies NormTrimSpace
			expected: `line 1 // comment stays
line 2 # comment stays

line 4 with spaces`, // Blank line stays as NormRemoveBlankLines not set
		},
		{
			name:	"RemoveGoComments only",
			input: ` line 1 // go comment
			  line 2 # ns comment stays

			line 4   `,
			flags:	NormRemoveGoComments,
			expected: ` line 1 
			  line 2 # ns comment stays

			line 4   `,
		},
		{
			name:	"RemoveNSComments only",
			input: ` line 1 // go comment stays
			  line 2 # ns comment

			line 4   `,
			flags:	NormRemoveNSComments,
			expected: ` line 1 // go comment stays
			  line 2 

			line 4   `,
		},
		{
			name:	"RemoveBlankLines only",
			// NOTE: Input line 5 uses a *literal* tab here
			input: `line 1

			line 3
			  // comment line only (whitespace)
			# ns comment line only (whitespace)
			  	  
			line 7`,
			flags:	NormRemoveBlankLines,
			// FIX V11: Correct expectation - blank lines (incl. tab-only) are removed. Others keep leading ws.
			expected: `line 1
			line 3
			  // comment line only (whitespace)
			# ns comment line only (whitespace)
			line 7`,
		},
		{
			name:	"Remove Comments and Blank Lines",
			// NOTE: Input line 4 uses a *literal* tab here
			input: `line 1 // go comment
			# ns comment line (becomes blank)
			line 3

			  	  // Line with only whitespace (Keep if Trim OFF)

			line 7`,
			flags:	NormRemoveGoComments | NormRemoveNSComments | NormRemoveBlankLines,	// No TrimSpace
			// FIX V14: Remove leading space from line 1, keep trailing space. Preserve leading tabs.
			expected: `line 1 
			line 3
			line 7`,
		},
		{
			name:	"Remove Comments, Blank Lines, AND Trim",	// More common combo
			// NOTE: Input line 4 uses a *literal* tab here
			input: `line 1 // go comment
			# ns comment line
			line 3

			  	  // Line with only whitespace

			line 7`,
			flags:	NormRemoveGoComments | NormRemoveNSComments | NormRemoveBlankLines | NormTrimSpace,
			// FIX V11: Confirmed expectation - blank lines removed, others trimmed.
			expected: `line 1
line 3
line 7`,
		},
		{
			name:	"All flags (NormDefault)",
			input: `
			  first line   // go comment here
			second  line # ns comment here

			  third line with   extra spaces


			fourth line ends.  `,
			flags:	NormDefault,
			expected: `first line
second line
third line with extra spaces
fourth line ends.`,
		},
		{
			name:		"Empty Input",
			input:		"",
			flags:		NormDefault,
			expected:	"",
		},
		{
			name:		"Whitespace only input",
			input:		" \n \t \n ",	// Uses actual tab
			flags:		NormDefault,
			expected:	"",	// Default includes Trim and RemoveBlanks - regex matches this as blank
		},
		{
			name:		"Input with only comments",
			input:		"// line 1\n# line 2\n  // line 3",
			flags:		NormDefault,
			expected:	"",	// Default includes comment removal, trim, blank removal
		},
		{
			name:	"No normalization needed",
			input:	"line 1\nline 2",
			flags:	NormDefault,
			expected: `line 1
line 2`,
		},
		{
			name:	"Windows Line Endings",
			input:	"line 1\r\nline 2\r\n",
			flags:	NormDefault,
			expected: `line 1
line 2`,
		},
		{
			name:	"Mixed Line Endings with Space",
			input:	"line 1 \r\n  line 2 \n line 3 \r\n",
			flags:	NormDefault,
			expected: `line 1
line 2
line 3`,
		},
		{
			name:		"Tab handling",
			input:		"line\t1\nline\t\t2",			// Input with tabs
			flags:		NormDefault &^ NormCompressSpace,	// Keep tabs, but trim/etc
			expected:	"line\t1\nline\t\t2",			// Expect tabs to be preserved
		},
		{
			name:		"Tab handling with compression",
			input:		"line\t1\nline\t\t2",	// Input with tabs
			flags:		NormDefault,		// Compress implies Trim
			expected:	"line 1\nline 2",	// Expect tabs to be compressed like spaces
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := NormalizeString(tc.input, tc.flags)
			if actual != tc.expected {
				// Use Errorf for failure message
				t.Errorf("NormalizeString failed for case: %s\nInput:\n%s\nExpected:\n%s\nActual:\n%s", tc.name, tc.input, tc.expected, actual)
				// Log the diff using Logf for supplementary info if -v is used
				// Pass tc.flags here so DiffStrings uses the same normalization for comparison as the test did.
				DiffStrings(t, tc.expected, actual, tc.flags, DiffAnsiColor|DiffVisibleSpace)
			}
		})
	}
}

// --- Tests for DiffStrings ---

// mockTB (remains the same)
type mockTB struct {
	testing.TB
	logs	[]string
	errors	[]string
	failed	bool
}

func newMockTB(t *testing.T) *mockTB	{ return &mockTB{TB: t} }
func (m *mockTB) Logf(format string, args ...interface{}) {
	s := fmt.Sprintf(format, args...)
	m.logs = append(m.logs, s)
}
func (m *mockTB) Errorf(format string, args ...interface{}) {
	s := fmt.Sprintf(format, args...)
	m.errors = append(m.errors, s)
	m.failed = true
}
func (m *mockTB) FailNow()	{ m.failed = true; panic("mockTB FailNow") }	// Panic to stop test
func (m *mockTB) Fail()		{ m.failed = true }
func (m *mockTB) Failed() bool	{ return m.failed }
func (m *mockTB) Helper()	{}

func TestDiffStrings(t *testing.T) {
	t.Run("Equal strings after normalization", func(t *testing.T) {
		mockT := newMockTB(t)
		expected := "line 1 // comment\n  line 2"
		actual := "line 1\nline 2   # comment"
		// Use default normalization for DiffStrings comparison
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
		hasErrorMsg := false
		for _, log := range mockT.logs {	// Check logs for the diff block
			if strings.Contains(log, "--- Diff ---") {
				hasDiffLog = true
				break
			}
		}
		for _, err := range mockT.errors {	// Check errors for the mismatch message
			if strings.Contains(err, "Content mismatch") {
				hasErrorMsg = true
				break
			}
		}
		if !hasErrorMsg {
			t.Errorf("FAIL: Expected DiffStrings to call t.Errorf for different strings, but it didn't seem to.")
		}
		if !hasDiffLog {
			t.Errorf("FAIL: Expected DiffStrings to log the diff output via t.Logf, but it didn't seem to.")
		}
	})

	t.Run("DiffShowFull flag", func(t *testing.T) {
		mockT := newMockTB(t)
		expected := "line 1"
		actual := "line 2"
		// Use DiffShowFull flag here
		areEqual := DiffStrings(mockT, expected, actual, NormDefault, DiffShowFull)

		if areEqual {
			t.Errorf("FAIL: Expected DiffStrings to return false for different strings, but got true")
		}
		if !mockT.Failed() {
			t.Errorf("FAIL: Expected DiffStrings to mark test as failed for different strings, but it didn't.")
		}

		hasExpectedLog := false
		hasActualLog := false
		hasNormExpectedLog := false
		hasNormActualLog := false
		for _, log := range mockT.logs {
			if strings.Contains(log, "--- Expected (Original) ---") {
				hasExpectedLog = true
			}
			if strings.Contains(log, "--- Actual (Original) ---") {
				hasActualLog = true
			}
			if strings.Contains(log, "--- Expected (Normalized) ---") {
				hasNormExpectedLog = true
			}
			if strings.Contains(log, "--- Actual (Normalized) ---") {
				hasNormActualLog = true
			}
		}
		if !hasExpectedLog || !hasActualLog || !hasNormExpectedLog || !hasNormActualLog {
			t.Errorf("FAIL: Expected DiffStrings with DiffShowFull to log Original and Normalized strings, but it didn't.")
		}
	})

	t.Run("DiffAnsiColor flag runs", func(t *testing.T) {
		mockT := newMockTB(t)
		expected := "line 1"
		actual := "line 2"
		// Use DiffAnsiColor flag
		areEqual := DiffStrings(mockT, expected, actual, NormDefault, DiffAnsiColor)

		if areEqual {
			t.Errorf("FAIL: Expected DiffStrings to return false, got true")
		}
		if !mockT.Failed() {
			t.Errorf("FAIL: Expected test to be marked as failed")
		}
		hasAnsi := false
		for _, log := range mockT.logs {	// Diff is logged via t.Logf
			if strings.Contains(log, "\x1b[") {	// Check for ANSI escape code
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
		expected := "line with space\t and cr\r\nnext line"	// \r\n becomes \n
		actual := "line with  extra   space\t\nnext line"	// \n ok, extra spaces
		// Use flags that keep internal spacing variations but remove comments/blanks for clearer diff
		normFlags := NormDefault &^ NormCompressSpace

		// Use DiffAnsiColor and DiffVisibleSpace flags
		areEqual := DiffStrings(mockT, expected, actual, normFlags, DiffAnsiColor|DiffVisibleSpace)

		if areEqual {
			t.Errorf("FAIL: Expected DiffStrings to return false, got true")
		}
		if !mockT.Failed() {
			t.Errorf("FAIL: Expected test to be marked as failed")
		}

		diffLogged := false
		logFound := ""
		for _, log := range mockT.logs {	// Diff is logged via t.Logf
			if strings.Contains(log, "--- Diff ---") {
				diffLogged = true
				logFound = log	// Capture the whole diff block log entry
				break
			}
		}

		if !diffLogged {
			t.Errorf("FAIL: DiffStrings did not log a diff block.")
		} else {
			symbolsExpected := []string{spaceSym, tabSym, nlSym}	// Check for ␣, ␉, ␤
			allFound := true
			symbolsMissing := []string{}
			for _, sym := range symbolsExpected {
				coloredSym := colorVisibleWs + sym + colorReset
				if !strings.Contains(logFound, coloredSym) {
					allFound = false
					symbolsMissing = append(symbolsMissing, sym)
				}
			}
			coloredCRSym := colorVisibleWs + crSym + colorReset
			if strings.Contains(logFound, coloredCRSym) {
				t.Errorf("FAIL: DiffStrings with DiffVisibleSpace showed CR symbol '%s', but CRs should be normalized away.", crSym)
				allFound = false
			}

			if !allFound {
				t.Errorf("FAIL: Expected DiffStrings with DiffVisibleSpace to show symbols ('%s'), but missing: %v.",
					strings.Join(symbolsExpected, "', '"), strings.Join(symbolsMissing, ", "))
				t.Logf("Relevant Log Chunk:\n%s", logFound)	// Log the diff block where symbols were checked
			}
		}
	})
}