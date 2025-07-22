// NeuroScript Version: 0.5.4
// File version: 2
// Purpose: Correctly parse git status, including untracked files for is_clean check.
// filename: pkg/tool/git/tools_git_status.go
// nlines: 218
// risk_rating: MEDIUM
package git

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

// --- toolGitStatus Implementation ---

// toolGitStatus implements the TOOL.GitStatus command.
// It executes the git command via the shared runGitCommand and then calls parseGitStatusOutput for parsing.
func toolGitStatus(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	// Validate arguments to accept an optional repo_path
	repoPath := "." // Default path
	if len(args) > 1 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "Git.Status expects at most one argument (repo_path)", lang.ErrInvalidArgument)
	}
	if len(args) == 1 {
		var ok bool
		repoPath, ok = args[0].(string)
		if !ok {
			return nil, lang.NewRuntimeError(lang.ErrorCodeType, "repo_path must be a string", lang.ErrInvalidArgument)
		}
	}

	interpreter.GetLogger().Debug("[GitStatus] Executing", "command", "git status --porcelain -b --untracked-files=all", "repoPath", repoPath)

	// Use the shared runGitCommand function, which correctly handles sandboxed paths.
	// This was the primary bug fix.
	output, err := runGitCommand(interpreter, repoPath, "status", "--porcelain", "-b", "--untracked-files=all")

	// Check if runGitCommand itself returned an error (command failed, etc.)
	if err != nil {
		// Check for the common "not a git repository" error message.
		errStr := err.Error()
		if strings.Contains(errStr, "not a git repository") || strings.Contains(errStr, "fatal: not a Git repository") {
			errMsg := "Not a git repository (or any of the parent directories)"
			if interpreter.SandboxDir() != "" {
				errMsg = fmt.Sprintf("%s in sandbox '%s'", errMsg, interpreter.SandboxDir())
			}
			// Return a map with the error, as is the convention for this specific tool on this error.
			return map[string]interface{}{
				"branch":                  nil,
				"remote_branch":           nil,
				"ahead":                   int64(0),
				"behind":                  int64(0),
				"files":                   []map[string]interface{}{},
				"untracked_files_present": false,
				"is_clean":                false,
				"error":                   errMsg,
			}, nil
		}
		// For other git errors, wrap and return them.
		return nil, lang.NewRuntimeError(lang.ErrorCodeToolExecutionFailed, fmt.Sprintf("Git.Status failed: %v", err), err)
	}

	// If command succeeded, pass the output to the parser
	resultMap, parseErr := parseGitStatusOutput(output)
	if parseErr != nil {
		interpreter.GetLogger().Debug("[GitStatus] Error parsing git status output", "error", parseErr)
		if resultMap == nil {
			resultMap = map[string]interface{}{
				"branch":                  nil,
				"remote_branch":           nil,
				"ahead":                   int64(0),
				"behind":                  int64(0),
				"files":                   []map[string]interface{}{},
				"untracked_files_present": false,
				"is_clean":                false,
				"error":                   nil,
			}
		}
		resultMap["error"] = fmt.Sprintf("Error parsing git status output: %v", parseErr)
		resultMap["is_clean"] = false
	}

	interpreter.GetLogger().Debug("[GitStatus] Result", "map", resultMap)
	return resultMap, nil // Return the result map
}

// --- Git Status Parsing Logic ---

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
		return resultMap, nil
	}
	branchLine := lines[0]
	branchInfo := ""

	// --- Parse Branch Line ---
	if strings.HasPrefix(branchLine, "## ") {
		branchInfo = strings.TrimPrefix(branchLine, "## ")

		if strings.HasPrefix(branchInfo, "No commits yet on ") {
			resultMap["branch"] = strings.TrimPrefix(branchInfo, "No commits yet on ")
		} else if strings.Contains(branchInfo, "HEAD (no branch)") {
			resultMap["branch"] = "(detached HEAD)"
		} else {
			aheadBehindPart := ""
			remotePart := ""
			localBranchPart := ""

			if strings.Contains(branchInfo, " [") && strings.HasSuffix(branchInfo, "]") {
				bracketStart := strings.LastIndex(branchInfo, " [")
				aheadBehindPart = branchInfo[bracketStart+2 : len(branchInfo)-1]
				branchInfo = branchInfo[:bracketStart]

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

			if strings.Contains(branchInfo, "...") {
				parts := strings.SplitN(branchInfo, "...", 2)
				localBranchPart = parts[0]
				remotePart = parts[1]
				resultMap["remote_branch"] = remotePart
			} else {
				localBranchPart = branchInfo
			}
			resultMap["branch"] = localBranchPart
		}
	} else {
		resultMap["error"] = fmt.Sprintf("Failed to parse branch information from line: %s", branchLine)
	}

	// --- Parse File Status Lines ---
	untrackedFound := false
	changesFound := false
	if len(lines) > 1 {
		for _, line := range lines[1:] {
			line = strings.TrimSuffix(line, "\r")
			if len(line) < 4 || line == "" {
				continue
			}

			indexStatus := string(line[0])
			worktreeStatus := string(line[1])
			pathPart := line[3:]
			path := ""
			originalPath := interface{}(nil)

			if (indexStatus == "R" || indexStatus == "C") && strings.Contains(pathPart, " -> ") {
				parts := strings.SplitN(pathPart, " -> ", 2)
				path = parts[0]
				originalPath = parts[1]
			} else {
				path = pathPart
			}

			if strings.HasPrefix(path, "\"") && strings.HasSuffix(path, "\"") {
				if unquotedPath, err := strconv.Unquote(path); err == nil {
					path = unquotedPath
				}
			}
			if origPathStr, ok := originalPath.(string); ok && strings.HasPrefix(origPathStr, "\"") && strings.HasSuffix(origPathStr, "\"") {
				if unquotedOrig, err := strconv.Unquote(origPathStr); err == nil {
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
				changesFound = true
			}

			filesList = append(filesList, fileMap)
		}
	}

	resultMap["files"] = filesList
	resultMap["untracked_files_present"] = untrackedFound
	resultMap["is_clean"] = !changesFound && !untrackedFound

	return resultMap, nil
}
