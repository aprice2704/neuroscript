// filename: pkg/core/universal_test_helpers.go
package core

import (
	"bufio"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/pmezard/go-difflib/difflib"
)

// --- String Normalization ---

// Normalization flags
const (
	NormTrimSpace         uint32 = 1 << 0 // Trim leading/trailing whitespace from each line and the whole string
	NormCompressSpace     uint32 = 1 << 1 // Replace multiple consecutive whitespace chars with a single space (implies TrimSpace)
	NormRemoveGoComments  uint32 = 1 << 2 // Remove // comments
	NormRemoveNSComments  uint32 = 1 << 3 // Remove # comments (assuming # for NeuroScript)
	NormRemoveBlankLines  uint32 = 1 << 4 // Remove lines that become empty after comment removal and potential trimming
	NormSpaceAroundTokens uint32 = 1 << 5 // Ensure single space around common tokens like {}, (), [], =, ==, etc. (More advanced)

	NormDefault uint32 = NormTrimSpace | NormCompressSpace | NormRemoveGoComments | NormRemoveNSComments | NormRemoveBlankLines
)

var (
	goCommentRegex  = regexp.MustCompile(`//.*`)
	nsCommentRegex  = regexp.MustCompile(`#.*`)
	multiSpaceRegex = regexp.MustCompile(`\s+`)
)

// NormalizeString applies various normalization options to a string.
// Revised logic V8: Final attempt at blank line logic.
func NormalizeString(content string, flags uint32) string {
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

		// 2. Determine if the line *is* effectively blank based on the original trim flag intent
		isEffectivelyBlank := false
		// Check based on the explicit TrimSpace flag requested by the user for blank check logic
		if originalTrimFlag != 0 {
			isEffectivelyBlank = (strings.TrimSpace(lineAfterComments) == "")
		} else {
			// If TrimSpace was NOT explicitly requested, blank only if originally empty after comments
			isEffectivelyBlank = (lineAfterComments == "")
		}

		// 3. Skip blank lines if requested AND the line is effectively blank
		if shouldRemoveBlankLines && isEffectivelyBlank {
			continue // Skip this line entirely
		}

		// 4. Process the non-blank line for spacing using the *effective* trim status for output
		lineForOutput := lineAfterComments // Start with comment-removed version
		if shouldTrim {                    // Use effective trim (includes implied by compress) for processing
			lineForOutput = strings.TrimSpace(lineForOutput)
		}
		if shouldCompress {
			// Compress implies trim, TrimSpace above already handled edges
			lineForOutput = multiSpaceRegex.ReplaceAllString(lineForOutput, " ")
			// Final safety trim if compress created single space edge case (and trim is effectively on)
			if shouldTrim {
				lineForOutput = strings.TrimSpace(lineForOutput)
			}
		}

		processedLines = append(processedLines, lineForOutput)
	}

	result := strings.Join(processedLines, "\n")

	// No final trim needed.
	return result
}

// --- String Diffing ---

// Diff display flags
const (
	DiffShowFull     uint32 = 1 << 0 // Show full expected and actual strings before the diff
	DiffAnsiColor    uint32 = 1 << 1 // Add ANSI color codes
	DiffNoContext    uint32 = 1 << 2 // Use difflib context=0 for minimal diff (N/A for custom diff)
	DiffVisibleSpace uint32 = 1 << 3 // Replace spaces/tabs/CR/NL with visible symbols (requires DiffAnsiColor)
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
// It normalizes inputs using NormalizeString before diffing.
// Returns true if strings are equal after normalization, false otherwise.
func DiffStrings(t testing.TB, expected, actual string, normFlags, diffFlags uint32) bool {
	t.Helper()

	normExpected := NormalizeString(expected, normFlags)
	normActual := NormalizeString(actual, normFlags)

	if normExpected == normActual {
		return true
	}

	t.Errorf("Content mismatch after normalization (flags: %d):", normFlags)

	if diffFlags&DiffShowFull != 0 {
		t.Logf("--- Expected (Original) ---\n%s\n--------------------------", expected)
		t.Logf("--- Actual (Original) ---\n%s\n------------------------", actual)
	}

	var diffBuilder strings.Builder
	separator := strings.Repeat("-", 60)
	diffBuilder.WriteString(separator + "\n")
	diffBuilder.WriteString(fmt.Sprintf("%-4s %-4s | Content\n", "EXP", "ACT"))
	diffBuilder.WriteString(separator + "\n")

	// Use SplitAfter to keep newlines attached for visualization
	aLines := strings.SplitAfter(normExpected, "\n")
	bLines := strings.SplitAfter(normActual, "\n")
	if len(aLines) > 0 && aLines[len(aLines)-1] == "" {
		aLines = aLines[:len(aLines)-1]
	}
	if len(bLines) > 0 && bLines[len(bLines)-1] == "" {
		bLines = bLines[:len(bLines)-1]
	}

	matcher := difflib.NewMatcher(aLines, bLines)
	opcodes := matcher.GetOpCodes() // Corrected method name

	useColor := diffFlags&DiffAnsiColor != 0
	showVisible := diffFlags&DiffVisibleSpace != 0 && useColor

	// Helper to format lines with optional color and visible whitespace
	formatDiffLine := func(line string, linePrefixColor string) string {
		// Preserve trailing newline if showVisible is true, because replacer handles it.
		// Otherwise, trim it for cleaner non-visible output.
		originalNewline := ""
		lineContent := line
		if strings.HasSuffix(line, "\n") {
			originalNewline = "\n"
			lineContent = strings.TrimSuffix(line, "\n")
		}
		// CR should already be removed by NormalizeString

		if showVisible {
			// Apply replacements only to the content part first
			lineContent = strings.ReplaceAll(lineContent, "\t", tabSym) // Replace individually before coloring
			lineContent = strings.ReplaceAll(lineContent, " ", spaceSym)
			// Add color around symbols
			lineContent = strings.ReplaceAll(lineContent, tabSym, colorVisibleWs+tabSym+colorReset+linePrefixColor) // Add back line color
			lineContent = strings.ReplaceAll(lineContent, spaceSym, colorVisibleWs+spaceSym+colorReset+linePrefixColor)

			// Add back newline symbol if needed, applying color correctly
			if originalNewline == "\n" {
				lineContent += colorVisibleWs + nlSym + colorReset
			}
		} else {
			// If not showing visible, add back the original newline
			lineContent += originalNewline
		}

		// Apply overall line color (Red/Green) *after* potential internal coloring of symbols
		if useColor {
			return fmt.Sprintf("%s%s%s", linePrefixColor, lineContent, colorReset) // Add final reset for safety
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
				lineContent := strings.TrimSuffix(aLines[i], "\n") // Display without newline char
				// Optionally show visible whitespace on equal lines too? Maybe dim?
				// if showVisible { lineContent = strings.ReplaceAll(strings.ReplaceAll(lineContent, " ", spaceSym), "\t", tabSym)}
				diffBuilder.WriteString(fmt.Sprintf("%-4d %-4d |  %s\n", lineNumA, lineNumB, lineContent))
				lineNumA++
				lineNumB++
			}
		case 'd': // Delete lines (only in A/Expected)
			for i := i1; i < i2; i++ {
				formattedLine := formatDiffLine(aLines[i], colorRed)
				diffBuilder.WriteString(fmt.Sprintf("%-4d %-4s |-%s\n", lineNumA, "", formattedLine))
				lineNumA++
			}
		case 'i': // Insert lines (only in B/Actual)
			for j := j1; j < j2; j++ {
				formattedLine := formatDiffLine(bLines[j], colorGreen)
				diffBuilder.WriteString(fmt.Sprintf("%-4s %-4d |+%s\n", "", lineNumB, formattedLine))
				lineNumB++
			}
		case 'r': // Replace lines (differs between A and B)
			// Optional: Limit replaced block size for readability
			delCount := i2 - i1
			addCount := j2 - j1
			// maxLines := 10 // Example limit

			for i := 0; i < delCount; i++ {
				// Add logic here to omit lines if needed (using maxLines)
				idx := i1 + i
				formattedLine := formatDiffLine(aLines[idx], colorRed)
				diffBuilder.WriteString(fmt.Sprintf("%-4d %-4s |-%s\n", lineNumA, "", formattedLine))
				lineNumA++
			}
			for j := 0; j < addCount; j++ {
				// Add logic here to omit lines if needed (using maxLines)
				idx := j1 + j
				formattedLine := formatDiffLine(bLines[idx], colorGreen)
				diffBuilder.WriteString(fmt.Sprintf("%-4s %-4d |+%s\n", "", lineNumB, formattedLine))
				lineNumB++
			}
		}
	}
	diffBuilder.WriteString(separator + "\n")

	t.Logf("--- Diff ---\n%s", diffBuilder.String())

	return false // Indicate that strings were different
}

// AssertEqualStrings is a convenience helper that uses DiffStrings and fails the test if needed.
func AssertEqualStrings(t *testing.T, expected, actual string, msgAndArgs ...interface{}) {
	t.Helper()
	defaultDiffFlags := DiffAnsiColor | DiffVisibleSpace // Enable visible space by default
	if !DiffStrings(t, expected, actual, NormDefault, defaultDiffFlags) {
		if len(msgAndArgs) > 0 {
			format, ok := msgAndArgs[0].(string)
			if !ok {
				t.Fatalf("First argument to AssertEqualStrings message must be a format string.")
			}
			t.Fatalf(format, msgAndArgs[1:]...)
		} else {
			t.FailNow()
		}
	}
}
