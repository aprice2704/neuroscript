// filename: pkg/core/tools_git_status.go
package core

import (
	"fmt" // Keep for toolGitStatus
	"regexp"
	"strconv"
	"strings"
)

// --- toolGitStatus Implementation ---

// toolGitStatus implements the TOOL.GitStatus command.
// It executes the git command and then calls parseGitStatusOutput for parsing.
func toolGitStatus(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Validate arguments (expecting zero)
	if len(args) != 0 {
		return nil, fmt.Errorf("TOOL.GitStatus expects 0 arguments, got %d", len(args))
	}

	interpreter.logger.Println("[TOOL GitStatus] Executing: git status --porcelain -b --untracked-files=all")

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
			if interpreter.sandboxDir != "" {
				errMsg = fmt.Sprintf("%s in sandbox '%s'", errMsg, interpreter.sandboxDir)
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
		interpreter.logger.Printf("[TOOL GitStatus] Error parsing git status output: %v", parseErr)
		resultMap["error"] = fmt.Sprintf("Error parsing git status output: %v", parseErr)
		resultMap["is_clean"] = false // Cannot reliably determine cleanliness if parsing fails
	}

	interpreter.logger.Printf("[TOOL GitStatus] Result: %+v", resultMap)
	return resultMap, nil // Return the result map, no Go error if command itself succeeded
}

// --- Git Status Parsing Logic ---

// branchInfoRegex parses the branch status line from `git status -b --porcelain`
var branchInfoRegex = regexp.MustCompile(`^## ([\w\-/.\(\)]+)(?:\.\.\.([\w\-/.]+))?(?: \[(?:ahead (\d+))?(?:, )?(?:behind (\d+))?\])?`)

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
		"is_clean":                true,             // Assume clean initially
		"error":                   interface{}(nil), // For soft errors detected during parsing
	}
	filesList := []map[string]interface{}{}

	// Trim leading/trailing whitespace and split into lines
	trimmedOutput := strings.TrimSpace(output)
	if trimmedOutput == "" {
		// Empty output implies clean (and likely not unborn, as that has a header)
		return resultMap, nil
	}
	lines := strings.Split(trimmedOutput, "\n")

	branchLine := ""
	if len(lines) > 0 {
		branchLine = lines[0]
	} else {
		// Should not happen if trimmedOutput was not empty, but handle defensively
		resultMap["is_clean"] = true
		return resultMap, nil
	}

	// Handle unborn branch specifically first
	if strings.HasPrefix(branchLine, "## No commits yet on ") {
		parts := strings.Fields(branchLine)
		if len(parts) >= 4 {
			resultMap["branch"] = parts[3] // Extract branch name
		} else {
			resultMap["branch"] = "(unknown unborn)" // Fallback
		}
		resultMap["is_clean"] = true // Unborn is considered "clean" for status purposes
		// No files to process
		return resultMap, nil
	}

	// --- Parse Branch Line (for normal repos) ---
	matches := branchInfoRegex.FindStringSubmatch(branchLine)
	if len(matches) > 1 {
		branchName := matches[1]
		if strings.Contains(branchName, "HEAD (no branch)") {
			resultMap["branch"] = "(detached HEAD)"
		} else {
			resultMap["branch"] = branchName
		}
		if len(matches) > 2 && matches[2] != "" {
			resultMap["remote_branch"] = matches[2]
		}
		if len(matches) > 3 && matches[3] != "" {
			aheadCount, _ := strconv.ParseInt(matches[3], 10, 64) // Ignore error, defaults to 0
			resultMap["ahead"] = aheadCount
		}
		if len(matches) > 4 && matches[4] != "" {
			behindCount, _ := strconv.ParseInt(matches[4], 10, 64) // Ignore error, defaults to 0
			resultMap["behind"] = behindCount
		}
	} else {
		// Fallback if branch line format is unexpected but starts with ##
		if strings.HasPrefix(branchLine, "## ") {
			if strings.Contains(branchLine, "HEAD (no branch)") {
				resultMap["branch"] = "(detached HEAD)"
			} else {
				parts := strings.Fields(branchLine)
				if len(parts) > 1 {
					resultMap["branch"] = parts[1]
				} else {
					resultMap["branch"] = strings.TrimPrefix(branchLine, "## ")
				}
			}
		} else {
			// This indicates a potential parsing issue or unexpected format
			resultMap["error"] = fmt.Sprintf("Failed to parse branch information from line: %s", branchLine)
			// Don't return error here, just set it in the map
		}
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

			if indexStatus == "R" || indexStatus == "C" {
				parts := strings.SplitN(pathPart, " -> ", 2)
				if len(parts) == 2 {
					path = parts[0]
					originalPath = parts[1]
				} else {
					path = pathPart // Fallback
				}
			} else {
				path = pathPart
			}

			// Unquote path if necessary
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

			if indexStatus == "?" && worktreeStatus == "?" {
				untrackedFound = true
			}

			// Check for any staged or unstaged changes to tracked files
			if indexStatus != " " || (worktreeStatus != " " && worktreeStatus != "?") {
				changesFound = true
			}

			fileMap := map[string]interface{}{
				"path":            path,
				"index_status":    indexStatus,
				"worktree_status": worktreeStatus,
				"original_path":   originalPath,
			}
			filesList = append(filesList, fileMap)
		}
	} // End file loop

	resultMap["files"] = filesList
	resultMap["untracked_files_present"] = untrackedFound
	resultMap["is_clean"] = !changesFound // Clean if no tracked file changes

	return resultMap, nil // Return the map, nil Go error (parsing errors are in map["error"])
}
