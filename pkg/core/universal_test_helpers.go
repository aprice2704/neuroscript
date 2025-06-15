// filename: pkg/core/universal_test_helpers.go
package core

import (
	"bufio"
	"fmt"
	"regexp"
	"strings"
	"testing"

	// Keep for potential future use if regex fails, but regex is better
	"github.com/pmezard/go-difflib/difflib"
)

// --- Test Logging Helpers ---

// testWriter redirects log output to t.Logf.
type testWriter struct {
	t *testing.T
}

// Write implements io.Writer, sending log output to the test's log.
func (tw testWriter) Write(p []byte) (n int, err error) {
	// Trim trailing newline added by log package if present, t.Logf adds its own.
	trimmed := strings.TrimSuffix(string(p), "\n")
	tw.t.Logf("%s", trimmed) // Use t.Logf to print the log message
	return len(p), nil
}

// logTest is a simple helper for logging during tests using t.Logf.
// It prevents needing to pass 't' everywhere explicitly just for logging.
// Deprecated: Prefer direct use of t.Logf for clarity unless this provides significant utility.
func LogTest(t *testing.T, format string, args ...interface{}) {
	t.Helper()
	t.Logf("[TEST LOG] "+format, args...)
}

type NormalizationFlags uint32
type DiffFlags uint32

// --- String Normalization ---
// (Normalization flags and functions remain unchanged)
// Normalization flags
const (
	NormTrimSpace         NormalizationFlags = 1 << 0 // Trim leading/trailing space characters (ASCII 32) from each line.
	NormCompressSpace     NormalizationFlags = 1 << 1 // Replace multiple consecutive whitespace chars with a single space (implies NormTrimSpace).
	NormRemoveGoComments  NormalizationFlags = 1 << 2 // Remove // comments.
	NormRemoveNSComments  NormalizationFlags = 1 << 3 // Remove # comments (assuming # for NeuroScript).
	NormRemoveBlankLines  NormalizationFlags = 1 << 4 // Remove lines containing only whitespace after comment removal.
	NormSpaceAroundTokens NormalizationFlags = 1 << 5 // Ensure single space around common tokens like {}, (), [], =, ==, etc. (More advanced, NYI).

	// NormDefault combines common normalization options.
	NormDefault          NormalizationFlags = NormTrimSpace | NormCompressSpace | NormRemoveGoComments | NormRemoveNSComments | NormRemoveBlankLines
	DefaultNormalization                    = NormDefault
)

var (
	goCommentRegex      = regexp.MustCompile(`//.*`)
	nsCommentRegex      = regexp.MustCompile(`#.*`)
	multiSpaceRegex     = regexp.MustCompile(`\s+`)
	onlyWhitespaceRegex = regexp.MustCompile(`^\s*$`) // Regex to check for lines containing only whitespace
)

// NormalizeString applies various normalization options to a string.
// Revised logic V12: Restore Debug + Fix compiler nits. (Debug Removed V13)
func NormalizeString(content string, flags NormalizationFlags) string {
	if flags == 0 {
		flags = NormDefault
	}

	// Determine effective flags *before* processing loop
	originalTrimFlag := flags & NormTrimSpace
	shouldCompress := flags&NormCompressSpace != 0
	shouldRemoveGoComments := flags&NormRemoveGoComments != 0
	shouldRemoveNSComments := flags&NormRemoveNSComments != 0
	shouldRemoveBlankLines := flags&NormRemoveBlankLines != 0

	// Effective trim status (Compress implies Trim)
	shouldTrim := (originalTrimFlag != 0) || shouldCompress

	var processedLines []string
	// Replace \r\n and \r with \n first
	content = strings.ReplaceAll(content, "\r\n", "\n")
	content = strings.ReplaceAll(content, "\r", "\n")
	scanner := bufio.NewScanner(strings.NewReader(content))

	for scanner.Scan() {
		line := scanner.Text() // Original line for this iteration

		// 1. Remove Comments
		lineAfterComments := line
		if shouldRemoveGoComments {
			lineAfterComments = goCommentRegex.ReplaceAllString(lineAfterComments, "")
		}
		if shouldRemoveNSComments {
			lineAfterComments = nsCommentRegex.ReplaceAllString(lineAfterComments, "")
		}

		// 2. Determine if the line *is* effectively blank (contains only whitespace chars)
		isEffectivelyBlank := false
		if shouldRemoveBlankLines { // Check only if flag is set
			isEffectivelyBlank = onlyWhitespaceRegex.MatchString(lineAfterComments)
		}

		// 3. Skip blank lines if requested AND the line is effectively blank
		if shouldRemoveBlankLines && isEffectivelyBlank {
			continue // Skip this line entirely
		}

		// 4. Process the non-blank line for spacing using the *effective* trim status for output
		lineForOutput := lineAfterComments // Start with comment-removed version

		// Apply compression first if needed, as it implies trimming logic.
		if shouldCompress {
			// Replace any sequence of one or more whitespace chars with a single space.
			lineForOutput = multiSpaceRegex.ReplaceAllString(lineForOutput, " ")
			// Compression includes trimming edges, so TrimSpace handles residual edge spaces.
			lineForOutput = strings.TrimSpace(lineForOutput)
		} else if shouldTrim {
			// Apply standard TrimSpace if only trimming (not compressing) was requested.
			// This only removes leading/trailing ASCII 32 spaces.
			lineForOutput = strings.TrimSpace(lineForOutput)
		}

		// Only append if the line wasn't skipped (redundant check, but safe)
		// and handle the edge case where trimming/compression results in an empty string
		// which should still be skipped if NormRemoveBlankLines is active.
		// Re-check blankness *after* processing, but only if removing blank lines.
		finalBlankCheckNeeded := shouldRemoveBlankLines
		if finalBlankCheckNeeded {
			if onlyWhitespaceRegex.MatchString(lineForOutput) {
				continue // Skip lines that *become* blank after processing
			}
		}

		processedLines = append(processedLines, lineForOutput)
	}
	// Ignore scanner errors for this helper

	result := strings.Join(processedLines, "\n")

	// No final trim needed on the whole result.
	return result
}

// --- String Diffing ---
// (Diff flags and functions remain unchanged)
// Diff display flags
const (
	DiffShowFull     DiffFlags = 1 << 0 // Show full expected and actual strings before the diff
	DiffAnsiColor    DiffFlags = 1 << 1 // Add ANSI color codes
	DiffNoContext    DiffFlags = 1 << 2 // Use difflib context=0 for minimal diff (N/A for custom diff)
	DiffVisibleSpace DiffFlags = 1 << 3 // Replace spaces/tabs/CR/NL with visible symbols (requires DiffAnsiColor)
	DefaultDiff                = DiffShowFull | DiffAnsiColor | DiffVisibleSpace
)

const (
	colorReset     = "\x1b[0m"
	colorRed       = "\x1b[31m" // Deletions
	colorGreen     = "\x1b[32m" // Additions
	colorGray      = "\x1b[90m" // Context / Equal lines
	colorVisibleWs = "\x1b[95m" // Bright Magenta/Pink for whitespace symbols
	spaceSym       = "␣"        // U+2423 OPEN BOX symbol for space
	tabSym         = "␉"        // U+2409 SYMBOL FOR HORIZONTAL TABULATION
	crSym          = "␍"        // U+240D SYMBOL FOR CARRIAGE RETURN (Unlikely after Norm)
	nlSym          = "␤"        // U+2424 SYMBOL FOR NEWLINE
)

// Package-level initialized replacer for visible whitespace symbols
var visibleWsReplacer = strings.NewReplacer(
	// Order matters: handle \n before space if space symbol includes space itself.
	// Color reset is crucial after each symbol.
	"\n", colorVisibleWs+nlSym+colorReset,
	"\r", colorVisibleWs+crSym+colorReset, // Should be removed by NormalizeString
	"\t", colorVisibleWs+tabSym+colorReset,
	" ", colorVisibleWs+spaceSym+colorReset,
)

// DiffStrings compares two strings (typically expected and actual test results)
// and logs a formatted diff using t.Logf or t.Errorf.
// It normalizes inputs using NormalizeString before diffing if normFlags != 0.
// Returns true if strings are equal after optional normalization, false otherwise.
func DiffStrings(t testing.TB, expected, actual string, normFlags NormalizationFlags, diffFlags DiffFlags) bool {
	t.Helper()

	normExpected := expected
	normActual := actual
	if normFlags != 0 { // Only normalize if flags are provided
		normExpected = NormalizeString(expected, normFlags)
		normActual = NormalizeString(actual, normFlags)
	}

	if normExpected == normActual {
		return true
	}

	// Use constant format string for Errorf
	t.Errorf("Content mismatch after normalization (norm flags: %d, diff flags: %d):", normFlags, diffFlags) // Mark test as failed

	// Use Logf for supplementary info, Errorf already marked the failure.
	if diffFlags&DiffShowFull != 0 {
		t.Logf("--- Expected (Original) ---\n%s\n--------------------------", expected)
		t.Logf("--- Actual (Original) ---\n%s\n------------------------", actual)
		if normFlags != 0 {
			t.Logf("--- Expected (Normalized) ---\n%s\n--------------------------", normExpected)
			t.Logf("--- Actual (Normalized) ---\n%s\n------------------------", normActual)
		}
	}

	var diffBuilder strings.Builder
	separator := strings.Repeat("-", 60)
	diffBuilder.WriteString("\n" + separator + "\n") // Add newline before diff for clarity
	diffBuilder.WriteString(fmt.Sprintf("%-4s %-4s | Content\n", "EXP", "ACT"))
	diffBuilder.WriteString(separator + "\n")

	// Use SplitAfter to keep newlines attached for visualization
	aLines := strings.SplitAfter(normExpected, "\n")
	bLines := strings.SplitAfter(normActual, "\n")
	// Remove trailing empty string if present after split (difflib common issue)
	if len(aLines) > 0 && aLines[len(aLines)-1] == "" {
		aLines = aLines[:len(aLines)-1]
	}
	if len(bLines) > 0 && bLines[len(bLines)-1] == "" {
		bLines = bLines[:len(bLines)-1]
	}

	matcher := difflib.NewMatcher(aLines, bLines)
	opcodes := matcher.GetOpCodes()

	useColor := diffFlags&DiffAnsiColor != 0
	showVisible := diffFlags&DiffVisibleSpace != 0 && useColor

	// Helper to format lines with optional color and visible whitespace
	formatDiffLine := func(line string, linePrefixColor string) string {
		hasNewline := strings.HasSuffix(line, "\n")
		lineContent := strings.TrimSuffix(line, "\n")
		// CR should already be removed by NormalizeString

		if showVisible {
			// Apply replacements only to the content part first
			lineContent = strings.ReplaceAll(lineContent, "\t", tabSym)
			lineContent = strings.ReplaceAll(lineContent, " ", spaceSym)
			// Add color around symbols
			coloredLineContent := ""
			for _, r := range lineContent {
				symStr := string(r)
				switch r {
				case []rune(tabSym)[0]:
					coloredLineContent += colorVisibleWs + symStr + colorReset + linePrefixColor
				case []rune(spaceSym)[0]:
					coloredLineContent += colorVisibleWs + symStr + colorReset + linePrefixColor
				default:
					coloredLineContent += symStr
				}
			}
			lineContent = coloredLineContent

			// Add back newline symbol if needed, applying color correctly
			if hasNewline {
				lineContent += colorVisibleWs + nlSym + colorReset
			}
		} else {
			// If not showing visible, add back the original newline if it existed
			if hasNewline {
				lineContent += "\n"
			}
		}

		// Apply overall line color (Red/Green) *after* potential internal coloring of symbols
		if useColor {
			// Ensure reset at the very end of the line
			return fmt.Sprintf("%s%s%s", linePrefixColor, lineContent, colorReset)
		}
		return lineContent
	}

	lineNumA := 1
	lineNumB := 1

	for _, opcode := range opcodes {
		tag, i1, i2, j1, j2 := opcode.Tag, opcode.I1, opcode.I2, opcode.J1, opcode.J2

		switch tag {
		case 'e': // Equal lines
			for i := i1; i < i2; i++ {
				// Display equal lines without color/prefix for clarity
				lineContent := strings.TrimSuffix(aLines[i], "\n") // Display without newline char
				// Optionally format equal lines for visible WS
				prefix := " " // Default prefix for equal lines
				formattedLine := lineContent
				if showVisible {
					prefix = ""                                             // formatDiffLine will handle spacing/symbols
					formattedLine = formatDiffLine(aLines[i], colorGray)    // Use gray for context
					formattedLine = strings.TrimSuffix(formattedLine, "\n") // Remove extra newline from helper
				} else if useColor {
					formattedLine = colorGray + lineContent + colorReset
				}

				diffBuilder.WriteString(fmt.Sprintf("%-4d %-4d |%s %s\n", lineNumA, lineNumB, prefix, formattedLine))
				lineNumA++
				lineNumB++
			}
		case 'd': // Delete lines (only in A/Expected)
			for i := i1; i < i2; i++ {
				formattedLine := formatDiffLine(aLines[i], colorRed)
				diffBuilder.WriteString(fmt.Sprintf("%-4d %-4s |-%s", lineNumA, "", strings.TrimSuffix(formattedLine, "\n")))
				if strings.HasSuffix(aLines[i], "\n") || (showVisible && strings.HasSuffix(formattedLine, nlSym+colorReset)) {
					diffBuilder.WriteString("\n")
				}
				lineNumA++
			}
		case 'i': // Insert lines (only in B/Actual)
			for j := j1; j < j2; j++ {
				formattedLine := formatDiffLine(bLines[j], colorGreen)
				diffBuilder.WriteString(fmt.Sprintf("%-4s %-4d |+%s", "", lineNumB, strings.TrimSuffix(formattedLine, "\n")))
				if strings.HasSuffix(bLines[j], "\n") || (showVisible && strings.HasSuffix(formattedLine, nlSym+colorReset)) {
					diffBuilder.WriteString("\n")
				}
				lineNumB++
			}
		case 'r': // Replace lines (differs between A and B)
			// Simpler 'r' handling: Show deletes then inserts
			for i := i1; i < i2; i++ {
				formattedLineA := formatDiffLine(aLines[i], colorRed)
				diffBuilder.WriteString(fmt.Sprintf("%-4d %-4s |-%s", lineNumA, "", strings.TrimSuffix(formattedLineA, "\n")))
				if strings.HasSuffix(aLines[i], "\n") || (showVisible && strings.HasSuffix(formattedLineA, nlSym+colorReset)) {
					diffBuilder.WriteString("\n")
				}
				lineNumA++
			}
			for j := j1; j < j2; j++ {
				formattedLineB := formatDiffLine(bLines[j], colorGreen)
				diffBuilder.WriteString(fmt.Sprintf("%-4s %-4d |+%s", "", lineNumB, strings.TrimSuffix(formattedLineB, "\n")))
				if strings.HasSuffix(bLines[j], "\n") || (showVisible && strings.HasSuffix(formattedLineB, nlSym+colorReset)) {
					diffBuilder.WriteString("\n")
				}
				lineNumB++
			}
		}
	}
	diffBuilder.WriteString(separator + "\n")

	// Log the final diff string using Logf as it's supplementary info
	t.Logf("--- Diff ---\n%s", diffBuilder.String())

	return false // Indicate that strings were different
}

// AssertEqualStrings is a convenience helper that uses DiffStrings with default flags
// and fails the test immediately if strings are not equal after normalization.
func AssertEqualStrings(t *testing.T, expected, actual string, msgAndArgs ...interface{}) {
	t.Helper()
	// Use default normalization (NormDefault) and default diff flags for assertion convenience
	defaultDiffFlags := DiffAnsiColor | DiffVisibleSpace
	if !DiffStrings(t, expected, actual, NormDefault, defaultDiffFlags) {
		// DiffStrings already called t.Errorf. We just need to stop the test.
		// Construct message if provided
		message := "Assertion failed: Strings not equal after default normalization."
		var finalMsg string
		if len(msgAndArgs) > 0 {
			format, ok := msgAndArgs[0].(string)
			if !ok {
				// Handle case where first arg is not a format string - just prepend our message
				finalMsg = fmt.Sprintf("%s (%+v)", message, msgAndArgs)
			} else {
				// Format the user's message and prepend ours
				userMsg := fmt.Sprintf(format, msgAndArgs[1:]...)
				finalMsg = fmt.Sprintf("%s (%s)", message, userMsg)
			}
		} else {
			finalMsg = message
		}
		// Use constant format string "%s" for Fatalf
		t.Fatalf("%s", finalMsg)
	}
}

// Helper for max function used in diff replace logic (if needed)
// func max(a, b int) int { // Moved to end to avoid conflict
// 	if a > b {
// 		return a
// 	}
// 	return b
// }

// writeFileHelper writes content to a file, creating directories if needed.
// Added from tools_go_ast_symbol_helpers_test.go.txt
// func writeFileHelper(t *testing.T, path string, content string) {
// 	t.Helper()
// 	dir := filepath.Dir(path)
// 	if err := os.MkdirAll(dir, 0755); err != nil {
// 		t.Fatalf("writeFileHelper: failed to create directory %s: %v", dir, err)
// 	}
// 	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
// 		t.Fatalf("writeFileHelper: failed to write file %s: %v", path, err)
// 	}
// }

// max helper function
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// --- REMOVED deepEqualWithTolerance function definition ---
// func deepEqualWithTolerance(a, b interface{}) bool { ... }
