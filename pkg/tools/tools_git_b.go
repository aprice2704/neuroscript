// NeuroScript Version: 0.3.1
// File version: 8 // Fixed validation and error handling in toolGitDiff.
// Purpose: Implements the second half of the Git tool functions.
// filename: pkg/core/tools_git_b.go
// nlines: 243
// risk_rating: MEDIUM

package core

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/lang"
)

func toolGitCheckout(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("%w: Git.Checkout requires at least a repo path and branch name", ErrInvalidArgument)
	}
	repoPath, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("%w: expected repo path as first argument, got %T", ErrInvalidArgument, args[0])
	}
	opArgs := args[1:]

	if len(opArgs) < 1 || len(opArgs) > 2 {
		return nil, fmt.Errorf("%w: Git.Checkout requires 1 or 2 operational arguments (branch, [create]), got %d", ErrInvalidArgument, len(opArgs))
	}
	branch, okB := opArgs[0].(string)
	if !okB || branch == "" {
		return nil, fmt.Errorf("%w: invalid type or empty value for 'branch', expected non-empty string", ErrInvalidArgument)
	}
	create := false
	if len(opArgs) == 2 {
		if opArgs[1] != nil {
			createOpt, okC := opArgs[1].(bool)
			if !okC {
				return nil, fmt.Errorf("%w: invalid type for 'create', expected boolean or nil, got %T", ErrInvalidArgument, opArgs[1])
			}
			create = createOpt
		}
	}

	gitArgs := []string{"checkout"}
	action := "checkout"
	if create {
		if strings.ContainsAny(branch, " \t\n\\/:*?\"<>|~^") {
			return nil, fmt.Errorf("%w: branch name '%s' contains invalid characters", ErrValidationArgValue, branch)
		}
		gitArgs = append(gitArgs, "-b")
		action = "create and checkout"
	}
	gitArgs = append(gitArgs, branch)

	output, err := runGitCommand(interpreter, repoPath, gitArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to %s branch/ref '%s': %w", action, branch, err)
	}
	return fmt.Sprintf("Successfully checked out branch/ref '%s'.\nOutput:\n%s", branch, output), nil
}

func toolGitRm(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("%w: GitRm requires two arguments (repoPath, path or paths)", ErrInvalidArgument)
	}
	repoPath, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("%w: expected repo path as first argument, got %T", ErrInvalidArgument, args[0])
	}

	var pathsToRemove []string
	switch pathsArg := args[1].(type) {
	case string:
		if pathsArg != "" {
			pathsToRemove = append(pathsToRemove, pathsArg)
		}
	case []interface{}:
		for _, p := range pathsArg {
			if pathStr, ok := p.(string); ok && pathStr != "" {
				pathsToRemove = append(pathsToRemove, pathStr)
			} else {
				return nil, fmt.Errorf("%w: path list for GitRm contained a non-string element: %T", ErrInvalidArgument, p)
			}
		}
	default:
		return nil, fmt.Errorf("%w: invalid type for 'path(s)' argument, expected string or list, got %T", ErrInvalidArgument, args[1])
	}

	if len(pathsToRemove) == 0 {
		return nil, fmt.Errorf("%w: no valid paths provided to GitRm", ErrInvalidArgument)
	}
	cmdArgs := append([]string{"rm"}, pathsToRemove...)
	output, err := runGitCommand(interpreter, repoPath, cmdArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to remove paths: %w", err)
	}
	return fmt.Sprintf("Successfully removed paths from git index.\nOutput:\n%s", output), nil
}

func toolGitMerge(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("%w: GitMerge requires two arguments (repoPath, branch name)", ErrInvalidArgument)
	}
	repoPath, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("%w: expected repo path as first argument, got %T", ErrInvalidArgument, args[0])
	}
	branchName, ok := args[1].(string)
	if !ok || branchName == "" {
		return nil, fmt.Errorf("%w: invalid type or empty value for 'branch', expected non-empty string", ErrInvalidArgument)
	}
	output, err := runGitCommand(interpreter, repoPath, "merge", branchName)
	if err != nil {
		return nil, fmt.Errorf("failed to merge branch '%s' (check for conflicts): %w", branchName, err)
	}
	return fmt.Sprintf("Successfully merged branch '%s'.\nOutput:\n%s", branchName, output), nil
}

func toolGitPull(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("%w: Git.Pull requires at least a repository path", ErrInvalidArgument)
	}
	repoPath, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("%w: expected repo path as first argument, got %T", ErrInvalidArgument, args[0])
	}
	opArgs := args[1:]
	gitArgs := []string{"pull"}
	if len(opArgs) > 0 {
		remote, okR := opArgs[0].(string)
		if !okR {
			return nil, fmt.Errorf("%w: invalid type for remote, expected string, got %T", ErrInvalidArgument, opArgs[0])
		}
		gitArgs = append(gitArgs, remote)
	}
	if len(opArgs) > 1 {
		branch, okB := opArgs[1].(string)
		if !okB {
			return nil, fmt.Errorf("%w: invalid type for branch, expected string, got %T", ErrInvalidArgument, opArgs[1])
		}
		gitArgs = append(gitArgs, branch)
	}
	output, err := runGitCommand(interpreter, repoPath, gitArgs...)
	if err != nil {
		return nil, fmt.Errorf("GitPull failed: %w", err)
	}
	return fmt.Sprintf("GitPull successful.\nOutput:\n%s", output), nil
}

func toolGitPush(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("%w: Git.Push requires at least a repository path", ErrInvalidArgument)
	}
	repoPath, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("%w: expected repo path as first argument, got %T", ErrInvalidArgument, args[0])
	}
	opArgs := args[1:]
	remote := "origin"
	var branch string
	setUpstream := false

	if len(opArgs) > 0 && opArgs[0] != nil {
		remoteOpt, okR := opArgs[0].(string)
		if !okR || remoteOpt == "" {
			return nil, fmt.Errorf("%w: invalid type or empty value for 'remote', got %T", ErrInvalidArgument, opArgs[0])
		}
		remote = remoteOpt
	}
	if len(opArgs) > 1 && opArgs[1] != nil {
		branchOpt, okB := opArgs[1].(string)
		if !okB || branchOpt == "" {
			return nil, fmt.Errorf("%w: invalid type or empty value for 'branch', got %T", ErrInvalidArgument, opArgs[1])
		}
		branch = branchOpt
	}
	if len(opArgs) > 2 && opArgs[2] != nil {
		upstreamOpt, okU := opArgs[2].(bool)
		if !okU {
			return nil, fmt.Errorf("%w: invalid type for 'set_upstream', expected boolean or nil", ErrInvalidArgument)
		}
		setUpstream = upstreamOpt
	}

	if branch == "" {
		var err error
		branch, err = getCurrentGitBranch(interpreter, repoPath)
		if err != nil {
			return nil, err
		}
	}

	gitArgs := []string{"push"}
	if setUpstream {
		gitArgs = append(gitArgs, "-u")
	}
	gitArgs = append(gitArgs, remote, branch)

	output, err := runGitCommand(interpreter, repoPath, gitArgs...)
	if err != nil {
		return nil, fmt.Errorf("GitPush failed: %w", err)
	}
	return fmt.Sprintf("GitPush successful (%s -> %s).\nOutput:\n%s", branch, remote, output), nil
}

func toolGitDiff(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, fmt.Errorf("%w: Git.Diff requires a repo path and an optional boolean 'cached' flag", ErrInvalidArgument)
	}
	repoPath, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("%w: expected repo path as first argument, got %T", ErrInvalidArgument, args[0])
	}

	gitArgs := []string{"diff"}
	if len(args) == 2 {
		if cached, isBool := args[1].(bool); isBool {
			if cached {
				gitArgs = append(gitArgs, "--cached")
			}
		} else {
			return nil, fmt.Errorf("%w: optional second argument to Git.Diff must be a boolean, got %T", ErrInvalidArgument, args[1])
		}
	}

	output, err := runGitCommand(interpreter, repoPath, gitArgs...)
	if err != nil {
		return nil, err
	}

	if output == "" {
		return "GitDiff: No changes detected.", nil
	}
	return output, nil
}

func toolGitClone(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	if len(args) != 2 {
		return nil, lang.NewRuntimeError(ErrorCodeArgMismatch, "Git.Clone: expected 2 arguments (repository_url, relative_path)", ErrArgumentMismatch)
	}
	repositoryURL, okURL := args[0].(string)
	if !okURL || repositoryURL == "" {
		return nil, lang.NewRuntimeError(ErrorCodeArgMismatch, "Git.Clone: repository_url (string) is required and cannot be empty", ErrInvalidArgument)
	}
	relativePath, okPath := args[1].(string)
	if !okPath || relativePath == "" {
		return nil, lang.NewRuntimeError(ErrorCodeArgMismatch, "Git.Clone: relative_path (string) is required and cannot be empty", ErrInvalidArgument)
	}

	sandboxRoot := interpreter.SandboxDir()
	if sandboxRoot == "" {
		return nil, lang.NewRuntimeError(ErrorCodeConfiguration, "Git.Clone: interpreter sandbox directory is not set", ErrConfiguration)
	}
	absTargetPath, secErr := SecureFilePath(relativePath, sandboxRoot)
	if secErr != nil {
		return nil, secErr
	}
	if _, err := os.Stat(absTargetPath); err == nil {
		return nil, lang.NewRuntimeError(ErrorCodePathExists, fmt.Sprintf("Git.Clone: target path '%s' already exists", relativePath), ErrPathExists)
	} else if !os.IsNotExist(err) {
		return nil, lang.NewRuntimeError(ErrorCodeIOFailed, fmt.Sprintf("Git.Clone: error checking target path '%s'", relativePath), errors.Join(ErrIOFailed, err))
	}

	cmd := exec.Command("git", "clone", repositoryURL, absTargetPath)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		errMsg := fmt.Sprintf("Git.Clone: 'git clone' command failed for URL '%s' into '%s'", repositoryURL, relativePath)
		stderrStr := strings.TrimSpace(stderr.String())
		if stderrStr != "" {
			errMsg = fmt.Sprintf("%s: %s", errMsg, stderrStr)
		}
		return nil, lang.NewRuntimeError(ErrorCodeToolExecutionFailed, errMsg, errors.Join(ErrToolExecutionFailed, err))
	}

	return fmt.Sprintf("Successfully cloned '%s' to '%s'", repositoryURL, relativePath), nil
}
