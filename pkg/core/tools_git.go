// NeuroScript Version: 0.3.1
// File version: 5 // Fixed argument parsing in toolGitRm and toolGitDiff.
// Purpose: Implements all Git tool functions with corrected command execution logic.
// filename: pkg/core/tools_git.go
// nlines: 710
// risk_rating: MEDIUM

package core

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// runGitCommand executes a git command within a specific repository path inside the sandbox.
func runGitCommand(interpreter *Interpreter, repoPath string, args ...string) (string, error) {
	absRepoPath, err := SecureFilePath(repoPath, interpreter.SandboxDir())
	if err != nil {
		return "", NewRuntimeError(ErrorCodePathViolation, fmt.Sprintf("invalid repository path '%s'", repoPath), err)
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
func getCurrentGitBranch(interpreter *Interpreter, repoPath string) (string, error) {
	output, err := runGitCommand(interpreter, repoPath, "symbolic-ref", "--short", "HEAD")
	if err != nil {
		stderrLower := strings.ToLower(err.Error())
		if strings.Contains(stderrLower, "fatal: head is a detached symbolic reference") || strings.Contains(stderrLower, "fatal: ref head is not a symbolic ref") {
			hash, hashErr := runGitCommand(interpreter, repoPath, "rev-parse", "--short", "HEAD")
			if hashErr != nil {
				return "", fmt.Errorf("failed to get current branch or commit hash (detached HEAD?): %w", hashErr)
			}
			return hash, nil
		}
		return "", fmt.Errorf("failed to get current git branch: %w", err)
	}
	return output, nil
}

func toolGitAdd(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("%w: GitAdd requires a repo path and a list of paths to add", ErrInvalidArgument)
	}
	repoPath, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("%w: expected repo path as first argument, got %T", ErrInvalidArgument, args[0])
	}
	pathsRaw, ok := args[1].([]interface{})
	if !ok {
		return nil, fmt.Errorf("%w: GitAdd requires a list of paths as its second argument, got %T", ErrInvalidArgument, args[1])
	}

	var paths []string
	for _, pathRaw := range pathsRaw {
		pathStr, ok := pathRaw.(string)
		if !ok {
			return nil, fmt.Errorf("%w: GitAdd path list contained non-string element: %T", ErrInvalidArgument, pathRaw)
		}
		paths = append(paths, pathStr)
	}
	if len(paths) == 0 {
		return "GitAdd: No valid file paths provided or list was empty.", nil
	}
	cmdArgs := append([]string{"add"}, paths...)
	output, err := runGitCommand(interpreter, repoPath, cmdArgs...)
	if err != nil {
		return nil, fmt.Errorf("GitAdd failed: %w", err)
	}
	return fmt.Sprintf("GitAdd successful for paths: %v.\nOutput:\n%s", paths, output), nil
}

func toolGitCommit(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	if len(args) < 2 || len(args) > 3 {
		return nil, fmt.Errorf("%w: Git.Commit requires 2 or 3 arguments (repoPath, message, [add_all])", ErrInvalidArgument)
	}
	repoPath, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("%w: expected repo path as first argument, got %T", ErrInvalidArgument, args[0])
	}
	message, okM := args[1].(string)
	if !okM || message == "" {
		return nil, fmt.Errorf("%w: invalid type or empty value for 'message', expected non-empty string", ErrInvalidArgument)
	}
	addAll := false
	if len(args) == 3 {
		if args[2] != nil {
			addAllOpt, okA := args[2].(bool)
			if !okA {
				return nil, fmt.Errorf("%w: invalid type for 'add_all', expected boolean or nil", ErrInvalidArgument)
			}
			addAll = addAllOpt
		}
	}

	if addAll {
		interpreter.logger.Debug("[Tool: GitCommit] Staging all changes (git add .)")
		_, errAdd := runGitCommand(interpreter, repoPath, "add", ".")
		if errAdd != nil {
			return nil, fmt.Errorf("failed during 'git add .' before commit: %w", errAdd)
		}
	}

	output, err := runGitCommand(interpreter, repoPath, "commit", "-m", message)
	if err != nil {
		stderrLower := strings.ToLower(err.Error())
		if strings.Contains(stderrLower, "nothing to commit") || strings.Contains(stderrLower, "no changes added to commit") {
			interpreter.logger.Warn("[Tool: GitCommit] Commit attempted but no changes were staged/detected.")
			return "GitCommit: Nothing to commit.", nil
		}
		return nil, fmt.Errorf("GitCommit failed: %w", err)
	}
	return fmt.Sprintf("GitCommit successful. Message: %q.\nOutput:\n%s", message, output), nil
}

func toolGitBranch(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("%w: Git.Branch requires at least a repository path", ErrInvalidArgument)
	}
	repoPath, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("%w: expected repo path as first argument, got %T", ErrInvalidArgument, args[0])
	}
	opArgs := args[1:]

	var branchNameOpt interface{}
	if len(opArgs) > 0 {
		branchNameOpt = opArgs[0]
	}

	name := ""
	if branchNameOpt != nil {
		if n, ok := branchNameOpt.(string); ok {
			name = n
		} else {
			return nil, fmt.Errorf("%w: invalid type for 'name', expected string or nil, got %T", ErrInvalidArgument, branchNameOpt)
		}
	}

	if name != "" {
		action := "create"
		gitArgs := []string{"branch", name}
		_, err := runGitCommand(interpreter, repoPath, gitArgs...)
		if err != nil {
			return nil, fmt.Errorf("failed to %s branch '%s': %w", action, name, err)
		}
		return fmt.Sprintf("Successfully %s branch '%s'.", action, name), nil
	} else {
		gitArgs := []string{"branch", "--no-color"}
		output, err := runGitCommand(interpreter, repoPath, gitArgs...)
		if err != nil {
			return nil, fmt.Errorf("failed to list branches: %w", err)
		}
		var branches []interface{}
		rawLines := strings.Split(output, "\n")
		for _, line := range rawLines {
			trimmedLine := strings.TrimSpace(line)
			trimmedLine = strings.TrimPrefix(trimmedLine, "* ")
			if trimmedLine != "" && !strings.Contains(trimmedLine, "->") {
				branches = append(branches, trimmedLine)
			}
		}
		return branches, nil
	}
}
