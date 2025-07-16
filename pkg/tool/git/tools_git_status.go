// NeuroScript Version: 0.5.4
// File version: 1
// Purpose: Correctly parse git status, including untracked files for is_clean check.
// filename: pkg/tool/git/tools_git_status.go
// nlines: 242
// risk_rating: MEDIUM
package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

// toolExec executes an external command and returns combined stdout/stderr as a string,
// or an error if the command fails to run or exits non-zero.
// This is intended as an *internal* helper for other tools like Git tools.
func toolExec(interpreter tool.Runtime, cmdAndArgs ...string) (string, error) {
	if len(cmdAndArgs) == 0 {
		return "", fmt.Errorf("toolExec requires at least a command")
	}
	commandPath := cmdAndArgs[0]
	commandArgs := cmdAndArgs[1:]

	// Basic security check (can be enhanced)
	if strings.Contains(commandPath, "..") || strings.ContainsAny(commandPath, "|;&$><`\\") {
		errMsg := fmt.Sprintf("toolExec blocked suspicious command path: %q", commandPath)
		if logger := interpreter.GetLogger(); logger != nil {
			logger.Error("[toolExec] " + errMsg)
		}
		// Return error message and a wrapped ErrInternalTool or a specific execution error
		return errMsg, fmt.Errorf("%w: %s", lang.ErrToolExecutionFailed, errMsg)
	}

	if logger := interpreter.GetLogger(); logger != nil {
		// Corrected logging call to use key-value pairs.
		logger.Debug("[toolExec] Executing", "command", commandPath, "args", strings.Join(commandArgs, " "))
	}

	cmd := exec.Command(commandPath, commandArgs...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	execErr := cmd.Run()

	stdoutStr := stdout.String()
	stderrStr := stderr.String()
	combinedOutput := stdoutStr + stderrStr // Combine outputs

	if execErr != nil {
		// Command failed (non-zero exit or execution error)
		errMsg := fmt.Sprintf("command '%s %s' failed with exit error: %v. Output:\n%s",
			commandPath, strings.Join(commandArgs, " "), execErr, combinedOutput)
		if logger := interpreter.GetLogger(); logger != nil {
			logger.Error("[toolExec] " + errMsg)
		}
		// Return the combined output along with the error
		return combinedOutput, fmt.Errorf("%w: %s", lang.ErrToolExecutionFailed, errMsg)
	}

	// Command succeeded
	if logger := interpreter.GetLogger(); logger != nil {
		logger.Debug("[toolExec] Command successful", "output", combinedOutput)
	}
	return combinedOutput, nil
}

// --- toolGitStatus Implementation ---

// toolGitStatus implements the TOOL.GitStatus command.
// It executes the git command and then calls parseGitStatusOutput for parsing.
func toolGitStatus(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	// Validate arguments (expecting zero)
	if len(args) != 0 {
		return nil, fmt.Errorf("TOOL.GitStatus expects 0 arguments, got %d", len(args))
	}

	interpreter.GetLogger().Debug("GitStatus] Executing: git status --porcelain -b --untracked-files=all")

	// Execute using toolExec which handles sandboxing and returns output/error string
	output, err := toolExec(interpreter, "git", "status", "--porcelain", "-b", "--untracked-files=all")

	// Check if toolExec itself returned an error (command failed, stderr etc.)
	if err != nil {
		// Initialize map structure even on error
		resultMap := map[string]interface{}{
			"branch":                  interface{}(nil),
			"remote_branch":           interface{}(nil),
			"ahead":                   int64(0),
			"behind":                  int64(0),
			"files":                   []map[string]interface{}{},
			"untracked_files_present": false,
			"is_clean":                false, // Cannot determine status on error
			"error":                   interface{}(nil),
		}
		// Check common "not a repo" messages within the error string from toolExec
		errStr := err.Error()
		if strings.Contains(errStr, "not a git repository") || strings.Contains(errStr, "fatal: not a Git repository") {
			errMsg := "Not a git repository (or any of the parent directories)"
			if interpreter.SandboxDir() != "" {
				errMsg = fmt.Sprintf("%s in sandbox '%s'", errMsg, interpreter.SandboxDir())
			}
			resultMap["error"] = errMsg
		} else {
			// General command failure
			resultMap["error"] = fmt.Sprintf("git status command failed: %s", errStr)
		}
		// Return the map containing the error string, not a Go error to the interpreter loop
		return resultMap, nil
	}

	// If toolExec succeeded, pass the output to the parser
	resultMap, parseErr := parseGitStatusOutput(output)
	if parseErr != nil {
		// Log parsing error and also include it in the returned map's error field
		interpreter.GetLogger().Debug("Tool: GitStatus] Error parsing git status output", "error", parseErr)
		// Use existing map if partial parsing happened, otherwise create a default one
		if resultMap == nil {
			resultMap = map[string]interface{}{
				"branch":                  interface{}(nil),
				"remote_branch":           interface{}(nil),
				"ahead":                   int64(0),
				"behind":                  int64(0),
				"files":                   []map[string]interface{}{},
				"untracked_files_present": false,
				"is_clean":                false, // Cannot reliably determine cleanliness if parsing fails
				"error":                   interface{}(nil),
			}
		}
		resultMap["error"] = fmt.Sprintf("Error parsing git status output: %v", parseErr)
		resultMap["is_clean"] = false
	}

	interpreter.GetLogger().Debug("Tool: GitStatus] Result", "map", resultMap)
	return resultMap, nil // Return the result map, no Go error if command itself succeeded
}

// --- Git Status Parsing Logic (REVISED) ---

// Regular expressions for extracting ahead/behind counts
var aheadRegex = regexp.MustCompile(`ahead (\d+)`)
var behindRegex = regexp.MustCompile(`behind (\d+)`)

// parseGitStatusOutput takes the raw string output from git status and parses it.
// It returns the structured map or a Go error if parsing fails fundamentally.
func parseGitStatusOutput(output string) (map[string]interface{}, error) {
	resultMap := map[string]interface{}{
		"branch":                  interface{}(nil),
		"remote_branch":           interface{}(nil),
		"ahead":                   int64(0),
		"behind":                  int64(0),
		"files":                   []map[string]interface{}{},
		"untracked_files_present": false,
		"is_clean":                true, // Assume clean initially
		"error":                   interface{}(nil),
	}
	filesList := []map[string]interface{}{}

	trimmedOutput := strings.TrimSpace(output)
	if trimmedOutput == "" {
		return resultMap, nil // Empty output means clean
	}
	lines := strings.Split(trimmedOutput, "\n")

	if len(lines) == 0 {
		return resultMap, nil // Should not happen, but be safe
	}
	branchLine := lines[0]
	branchInfo := "" // Part of the line containing branch/remote/ahead/behind info

	// --- Parse Branch Line ---
	if strings.HasPrefix(branchLine, "## ") {
		branchInfo = strings.TrimPrefix(branchLine, "## ")

		// Check for specific states first
		if strings.HasPrefix(branchInfo, "No commits yet on ") {
			resultMap["branch"] = strings.TrimPrefix(branchInfo, "No commits yet on ")
			// is_clean remains true, ahead/behind 0, remote nil
		} else if strings.Contains(branchInfo, "HEAD (no branch)") {
			resultMap["branch"] = "(detached HEAD)"
			// is_clean depends on files, ahead/behind 0, remote nil
		} else {
			// Normal branch parsing
			aheadBehindPart := ""
			remotePart := ""
			localBranchPart := ""

			// Extract ahead/behind info first
			if strings.Contains(branchInfo, " [") && strings.HasSuffix(branchInfo, "]") {
				bracketStart := strings.LastIndex(branchInfo, " [")
				aheadBehindPart = branchInfo[bracketStart+2 : len(branchInfo)-1] // Content inside brackets
				branchInfo = branchInfo[:bracketStart]                           // Remaining part before brackets

				aheadMatches := aheadRegex.FindStringSubmatch(aheadBehindPart)
				if len(aheadMatches) > 1 {
					aheadCount, _ := strconv.ParseInt(aheadMatches[1], 10, 64)
					resultMap["ahead"] = aheadCount
				}
				behindMatches := behindRegex.FindStringSubmatch(aheadBehindPart)
				if len(behindMatches) > 1 {
					behindCount, _ := strconv.ParseInt(behindMatches[1], 10, 64)
					resultMap["behind"] = behindCount
				}
			}

			// Extract remote tracking info
			if strings.Contains(branchInfo, "...") {
				parts := strings.SplitN(branchInfo, "...", 2)
				localBranchPart = parts[0]
				remotePart = parts[1]
				resultMap["remote_branch"] = remotePart
			} else {
				localBranchPart = branchInfo // No remote tracking info
			}
			resultMap["branch"] = localBranchPart
		}
	} else {
		// Unexpected branch line format
		resultMap["error"] = fmt.Sprintf("Failed to parse branch information from line: %s", branchLine)
	}

	// --- Parse File Status Lines ---
	untrackedFound := false
	changesFound := false // Tracks staged/unstaged changes to *tracked* files
	if len(lines) > 1 {   // Process only if there are file lines
		for _, line := range lines[1:] {
			line = strings.TrimSuffix(line, "\r") // Handle potential CRLF
			if len(line) < 4 || line == "" {
				continue
			}

			indexStatus := string(line[0])
			worktreeStatus := string(line[1])
			pathPart := line[3:]
			path := ""
			originalPath := interface{}(nil)

			// Handle Renamed/Copied paths which have " -> " separator
			if (indexStatus == "R" || indexStatus == "C") && strings.Contains(pathPart, " -> ") {
				parts := strings.SplitN(pathPart, " -> ", 2)
				path = parts[0]
				originalPath = parts[1]
			} else {
				path = pathPart
			}

			// Unquote path if necessary (Git uses C-style quoting for unusual chars)
			if strings.HasPrefix(path, "\"") && strings.HasSuffix(path, "\"") {
				unquotedPath, err := strconv.Unquote(path)
				if err == nil {
					path = unquotedPath
				}
			}
			// Unquote originalPath if necessary
			if origPathStr, ok := originalPath.(string); ok && strings.HasPrefix(origPathStr, "\"") && strings.HasSuffix(origPathStr, "\"") {
				unquotedOrig, err := strconv.Unquote(origPathStr)
				if err == nil {
					originalPath = unquotedOrig
				}
			}

			fileMap := map[string]interface{}{
				"path":            path,
				"index_status":    indexStatus,
				"worktree_status": worktreeStatus,
				"original_path":   originalPath,
			}

			if indexStatus == "?" && worktreeStatus == "?" {
				untrackedFound = true
			} else {
				// Any other status code indicates a change to a tracked file.
				changesFound = true
			}

			filesList = append(filesList, fileMap)
		}
	} // End file loop

	resultMap["files"] = filesList
	resultMap["untracked_files_present"] = untrackedFound
	// FIX: A repo is not clean if there are any changes OR untracked files.
	resultMap["is_clean"] = !changesFound && !untrackedFound

	return resultMap, nil
}
