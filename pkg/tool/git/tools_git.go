// NeuroScript Version: 0.3.1
// File version: 8 // Corrected type conversion for variadic string arguments in pull and push.
// Purpose: Implements all Git tool functions.
// filename: pkg/tool/git/tools_git.go
// nlines: 300+
// risk_rating: MEDIUM

package git

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/security"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

// runGitCommand executes a git command within a specific repository path inside the sandbox.
func runGitCommand(interpreter tool.Runtime, repoPath string, args ...string) (string, error) {
	absRepoPath, err := security.SecureFilePath(repoPath, interpreter.SandboxDir())
	if err != nil {
		return "", lang.NewRuntimeError(lang.ErrorCodePathViolation, fmt.Sprintf("invalid repository path '%s'", repoPath), err)
	}

	cmd := exec.Command("git", args...)
	cmd.Dir = absRepoPath

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	runErr := cmd.Run()
	stderrStr := strings.TrimSpace(stderr.String())

	if runErr != nil {
		return "", fmt.Errorf("git command failed (in %s): 'git %s' -> %v: %s", repoPath, strings.Join(args, " "), runErr, stderrStr)
	}

	return strings.TrimSpace(stdout.String()), nil
}

// getCurrentGitBranch determines the current branch name.
func getCurrentGitBranch(interpreter tool.Runtime, repoPath string) (string, error) {
	output, err := runGitCommand(interpreter, repoPath, "symbolic-ref", "--short", "HEAD")
	if err != nil {
		stderrLower := strings.ToLower(err.Error())
		if strings.Contains(stderrLower, "fatal: head is a detached symbolic reference") || strings.Contains(stderrLower, "fatal: ref head is not a symbolic ref") {
			hash, hashErr := runGitCommand(interpreter, repoPath, "rev-parse", "--short", "HEAD")
			if hashErr != nil {
				return "", fmt.Errorf("failed to get current commit hash (detached HEAD?): %w", hashErr)
			}
			return hash, nil
		}
		return "", fmt.Errorf("failed to get current git branch: %w", err)
	}
	return output, nil
}

func toolGitAdd(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	if len(args) != 2 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "Git.Add requires a repo path and a list of paths", lang.ErrInvalidArgument)
	}
	repoPath, ok := args[0].(string)
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, "repo_path must be a string", lang.ErrInvalidArgument)
	}
	pathsRaw, ok := args[1].([]interface{})
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, "paths must be a list", lang.ErrInvalidArgument)
	}

	var paths []string
	for _, pathRaw := range pathsRaw {
		pathStr, ok := pathRaw.(string)
		if !ok {
			return nil, lang.NewRuntimeError(lang.ErrorCodeType, "all paths in the list must be strings", lang.ErrInvalidArgument)
		}
		paths = append(paths, pathStr)
	}
	if len(paths) == 0 {
		return "Git.Add: No valid file paths provided.", nil
	}
	cmdArgs := append([]string{"add"}, paths...)
	output, err := runGitCommand(interpreter, repoPath, cmdArgs...)
	if err != nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeToolExecutionFailed, fmt.Sprintf("Git.Add failed: %v", err), err)
	}
	return fmt.Sprintf("Git.Add successful for paths: %v.\n%s", paths, output), nil
}

func toolGitCommit(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	if len(args) < 2 || len(args) > 3 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "Git.Commit requires 2 or 3 arguments", lang.ErrArgumentMismatch)
	}
	repoPath, ok := args[0].(string)
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, "repo_path must be a string", lang.ErrInvalidArgument)
	}
	message, okM := args[1].(string)
	if !okM || message == "" {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "commit_message must be a non-empty string", lang.ErrInvalidArgument)
	}
	addAll := false
	if len(args) == 3 && args[2] != nil {
		addAll, ok = args[2].(bool)
		if !ok {
			return nil, lang.NewRuntimeError(lang.ErrorCodeType, "add_all must be a boolean", lang.ErrInvalidArgument)
		}
	}

	if addAll {
		if _, err := runGitCommand(interpreter, repoPath, "add", "."); err != nil {
			return nil, lang.NewRuntimeError(lang.ErrorCodeToolExecutionFailed, fmt.Sprintf("failed 'git add .': %v", err), err)
		}
	}

	output, err := runGitCommand(interpreter, repoPath, "commit", "-m", message)
	if err != nil {
		if strings.Contains(err.Error(), "nothing to commit") {
			return "Git.Commit: Nothing to commit.", nil
		}
		return nil, lang.NewRuntimeError(lang.ErrorCodeToolExecutionFailed, fmt.Sprintf("Git.Commit failed: %v", err), err)
	}
	return fmt.Sprintf("Git.Commit successful: %s", output), nil
}

func toolGitBranch(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	if len(args) < 1 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "Git.Branch requires at least a repository path", lang.ErrArgumentMismatch)
	}
	repoPath, ok := args[0].(string)
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, "repo_path must be a string", lang.ErrInvalidArgument)
	}

	if len(args) == 1 { // List branches
		output, err := runGitCommand(interpreter, repoPath, "branch", "--no-color")
		if err != nil {
			return nil, lang.NewRuntimeError(lang.ErrorCodeToolExecutionFailed, fmt.Sprintf("failed to list branches: %v", err), err)
		}
		var branches []interface{}
		for _, line := range strings.Split(output, "\n") {
			trimmed := strings.TrimSpace(strings.TrimPrefix(line, "* "))
			if trimmed != "" {
				branches = append(branches, trimmed)
			}
		}
		return branches, nil
	}

	// Create branch
	branchName, ok := args[1].(string)
	if !ok || branchName == "" {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "branch name must be a non-empty string", lang.ErrInvalidArgument)
	}
	_, err := runGitCommand(interpreter, repoPath, "branch", branchName)
	if err != nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeToolExecutionFailed, fmt.Sprintf("failed to create branch '%s': %v", branchName, err), err)
	}
	return fmt.Sprintf("Successfully created branch '%s'.", branchName), nil
}

func toolGitCheckout(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	if len(args) < 2 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "Git.Checkout requires at least repo_path and branch", lang.ErrArgumentMismatch)
	}
	repoPath, ok := args[0].(string)
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, "repo_path must be a string", lang.ErrInvalidArgument)
	}
	branch, ok := args[1].(string)
	if !ok || branch == "" {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "branch must be a non-empty string", lang.ErrInvalidArgument)
	}
	create := false
	if len(args) > 2 && args[2] != nil {
		create, ok = args[2].(bool)
		if !ok {
			return nil, lang.NewRuntimeError(lang.ErrorCodeType, "create flag must be a boolean", lang.ErrInvalidArgument)
		}
	}

	gitArgs := []string{"checkout"}
	if create {
		gitArgs = append(gitArgs, "-b")
	}
	gitArgs = append(gitArgs, branch)

	output, err := runGitCommand(interpreter, repoPath, gitArgs...)
	if err != nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeToolExecutionFailed, fmt.Sprintf("failed to checkout '%s': %v", branch, err), err)
	}
	return fmt.Sprintf("Successfully checked out '%s'.\n%s", branch, output), nil
}

func toolGitRm(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	if len(args) != 2 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "Git.Rm requires repo_path and paths", lang.ErrArgumentMismatch)
	}
	repoPath, ok := args[0].(string)
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, "repo_path must be a string", lang.ErrInvalidArgument)
	}
	paths, ok := args[1].([]interface{})
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, "paths must be a list", lang.ErrInvalidArgument)
	}
	if len(paths) == 0 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "paths list cannot be empty", lang.ErrInvalidArgument)
	}
	strPaths := make([]string, len(paths))
	for i, p := range paths {
		strPaths[i], ok = p.(string)
		if !ok {
			return nil, lang.NewRuntimeError(lang.ErrorCodeType, "all paths in list must be strings", lang.ErrInvalidArgument)
		}
	}
	cmdArgs := append([]string{"rm"}, strPaths...)
	output, err := runGitCommand(interpreter, repoPath, cmdArgs...)
	if err != nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeToolExecutionFailed, fmt.Sprintf("failed to remove paths: %v", err), err)
	}
	return fmt.Sprintf("Successfully removed: %s\n%s", strings.Join(strPaths, ", "), output), nil
}

func toolGitMerge(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	if len(args) != 2 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "Git.Merge requires repo_path and branch", lang.ErrArgumentMismatch)
	}
	repoPath, ok := args[0].(string)
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, "repo_path must be a string", lang.ErrInvalidArgument)
	}
	branch, ok := args[1].(string)
	if !ok || branch == "" {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "branch must be a non-empty string", lang.ErrInvalidArgument)
	}
	output, err := runGitCommand(interpreter, repoPath, "merge", branch)
	if err != nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeToolExecutionFailed, fmt.Sprintf("failed to merge branch '%s': %v", branch, err), err)
	}
	return fmt.Sprintf("Successfully merged '%s'.\n%s", branch, output), nil
}

func toolGitPull(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	if len(args) < 1 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "Git.Pull requires at least a repo_path", lang.ErrArgumentMismatch)
	}
	repoPath, ok := args[0].(string)
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, "repo_path must be a string", lang.ErrInvalidArgument)
	}
	gitArgs := []string{"pull"}
	for _, arg := range args[1:] {
		if strArg, ok := arg.(string); ok {
			gitArgs = append(gitArgs, strArg)
		} else {
			return nil, lang.NewRuntimeError(lang.ErrorCodeType, "all optional arguments for Pull must be strings", lang.ErrInvalidArgument)
		}
	}
	output, err := runGitCommand(interpreter, repoPath, gitArgs...)
	if err != nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeToolExecutionFailed, fmt.Sprintf("Git.Pull failed: %v", err), err)
	}
	return fmt.Sprintf("Git.Pull successful.\n%s", output), nil
}

func toolGitPush(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	if len(args) < 1 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "Git.Push requires at least a repo_path", lang.ErrArgumentMismatch)
	}
	repoPath, ok := args[0].(string)
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, "repo_path must be a string", lang.ErrInvalidArgument)
	}
	gitArgs := []string{"push"}
	for _, arg := range args[1:] {
		if strArg, ok := arg.(string); ok {
			gitArgs = append(gitArgs, strArg)
		} else {
			return nil, lang.NewRuntimeError(lang.ErrorCodeType, "all optional arguments for Push must be strings", lang.ErrInvalidArgument)
		}
	}
	output, err := runGitCommand(interpreter, repoPath, gitArgs...)
	if err != nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeToolExecutionFailed, fmt.Sprintf("Git.Push failed: %v", err), err)
	}
	return fmt.Sprintf("Git.Push successful.\n%s", output), nil
}

func toolGitDiff(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	if len(args) < 1 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "Git.Diff requires at least a repo_path", lang.ErrArgumentMismatch)
	}
	repoPath, ok := args[0].(string)
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, "repo_path must be a string", lang.ErrInvalidArgument)
	}
	gitArgs := []string{"diff"}
	if len(args) > 1 {
		if cached, ok := args[1].(bool); ok && cached {
			gitArgs = append(gitArgs, "--cached")
		}
	}
	output, err := runGitCommand(interpreter, repoPath, gitArgs...)
	if err != nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeToolExecutionFailed, err.Error(), err)
	}
	if output == "" {
		return "GitDiff: No changes detected.", nil
	}
	return output, nil
}

func toolGitClone(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	if len(args) != 2 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "Git.Clone: expected 2 arguments (repository_url, relative_path)", lang.ErrArgumentMismatch)
	}
	repositoryURL, okURL := args[0].(string)
	if !okURL || repositoryURL == "" {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "repository_url is required", lang.ErrInvalidArgument)
	}
	relativePath, okPath := args[1].(string)
	if !okPath || relativePath == "" {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "relative_path is required", lang.ErrInvalidArgument)
	}

	absTargetPath, secErr := security.SecureFilePath(relativePath, interpreter.SandboxDir())
	if secErr != nil {
		return nil, secErr
	}
	if _, err := os.Stat(absTargetPath); err == nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodePathExists, fmt.Sprintf("target path '%s' already exists", relativePath), lang.ErrPathExists)
	}

	cmd := exec.Command("git", "clone", repositoryURL, absTargetPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeToolExecutionFailed, fmt.Sprintf("git clone failed: %s", string(output)), err)
	}
	return fmt.Sprintf("Successfully cloned '%s' to '%s'", repositoryURL, relativePath), nil
}

func toolGitReset(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	if len(args) < 1 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "Git.Reset requires at least a repo_path", lang.ErrArgumentMismatch)
	}
	repoPath, ok := args[0].(string)
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, "repo_path must be a string", lang.ErrInvalidArgument)
	}

	gitArgs := []string{"reset"}
	if len(args) > 1 {
		if mode, ok := args[1].(string); ok && mode != "" {
			gitArgs = append(gitArgs, "--"+mode)
		}
	}
	if len(args) > 2 {
		if commit, ok := args[2].(string); ok && commit != "" {
			gitArgs = append(gitArgs, commit)
		}
	}

	output, err := runGitCommand(interpreter, repoPath, gitArgs...)
	if err != nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeToolExecutionFailed, fmt.Sprintf("Git.Reset failed: %v", err), err)
	}
	return fmt.Sprintf("Git.Reset successful.\n%s", output), nil
}
