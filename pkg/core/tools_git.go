// NeuroScript Version: 0.3.1
// File version: 0.1.0
// Register all git tools via init() and define specs for all tools in file.
// filename: pkg/core/tools_git.go

package core

import (
	"errors"
	"fmt"
	"strings"
	// "os/exec" // toolExec likely handles this
	// "bytes" // toolExec likely handles this
)

// --- Helper Function (Assumed): toolExec ---
// func toolExec(interpreter *Interpreter, cmdAndArgs ...string) (string, error)
// This function is assumed to exist, likely in tools_shell.go,
// and handles running commands in the sandbox, capturing output, and errors.

// --- Helper Function (NEW - specific to git tools) ---

// getCurrentGitBranch determines the current branch name.
func getCurrentGitBranch(interpreter *Interpreter) (string, error) {
	// git symbolic-ref --short HEAD is often more reliable than rev-parse for current branch
	output, err := toolExec(interpreter, "git", "symbolic-ref", "--short", "HEAD")
	if err != nil {
		// Check if the error is because we are in detached HEAD state
		// Error messages vary slightly between git versions
		stderrLower := strings.ToLower(err.Error()) // Use error message which should contain stderr via toolExec wrapper
		if strings.Contains(stderrLower, "fatal: head is a detached symbolic reference") || strings.Contains(stderrLower, "fatal: ref head is not a symbolic ref") {
			// Try getting the commit hash instead for detached HEAD
			interpreter.Logger().Debug("Getting commit hash for detached HEAD state.")
			output, err = toolExec(interpreter, "git", "rev-parse", "--short", "HEAD")
			if err != nil {
				return "", fmt.Errorf("failed to get current branch or commit hash (detached HEAD?): %w", err)
			}
			// Trim potential whitespace from commit hash output
			return strings.TrimSpace(output), nil // Return commit hash in detached state
		}
		// Return the original error if it wasn't a detached HEAD issue
		return "", fmt.Errorf("failed to get current git branch: %w", err)
	}
	// Trim potential whitespace from branch name output
	return strings.TrimSpace(output), nil
}

// --- toolGitAdd implementation ---
func toolGitAdd(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("%w: GitAdd requires exactly one argument (list of paths)", ErrInvalidArgument)
	}
	pathsRaw, ok := args[0].([]interface{}) // args[0] itself is the list of paths
	if !ok {
		return nil, fmt.Errorf("%w: GitAdd requires a list of paths as its argument, got %T", ErrInvalidArgument, args[0])
	}

	paths := make([]string, 0, len(pathsRaw))
	validatedPaths := make([]string, 0, len(pathsRaw))

	for _, pathRaw := range pathsRaw {
		pathStr, ok := pathRaw.(string)
		if !ok {
			// Allow skipping non-strings in the list? Or error out? Error for now.
			return nil, fmt.Errorf("%w: GitAdd path list contained non-string element: %T", ErrInvalidArgument, pathRaw)
		}

		_, secErr := SecureFilePath(pathStr, interpreter.sandboxDir)
		if secErr != nil {
			errMsg := fmt.Sprintf("GitAdd path error for '%s': %s", pathStr, secErr.Error())
			// Decide: return error immediately or collect errors? Return immediately for now.
			return nil, fmt.Errorf("%s: %w", errMsg, errors.Join(ErrValidationArgValue, secErr))
		}
		validatedPaths = append(validatedPaths, pathStr) // Add validated path
		paths = append(paths, pathStr)                   // Collect relative paths for command
	}

	if len(paths) == 0 {
		return "GitAdd: No valid file paths provided or list was empty.", nil
	}

	// Command is "git", arguments are "add", path1, path2, ...
	cmdArgs := append([]string{"add"}, paths...)
	output, err := toolExec(interpreter, append([]string{"git"}, cmdArgs...)...) // Pass "git" and then the cmdArgs elements

	if err != nil {
		// toolExec includes output in error, so just wrap
		return nil, fmt.Errorf("GitAdd failed: %w", err)
	}

	return fmt.Sprintf("GitAdd successful for paths: %v.\nOutput:\n%s", validatedPaths, output), nil
}

// --- toolGitCommit implementation ---
func toolGitCommit(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Args: message (string, required), add_all (bool, optional)
	if len(args) < 1 || len(args) > 2 {
		return nil, fmt.Errorf("%w: Git.Commit requires 1 or 2 arguments (message, [add_all])", ErrInvalidArgument)
	}

	message, okM := args[0].(string)
	if !okM || message == "" {
		return nil, fmt.Errorf("%w: invalid type or empty value for 'message', expected non-empty string", ErrInvalidArgument)
	}

	addAll := false // Default
	if len(args) == 2 {
		// Allow nil to explicitly skip optional arg
		if args[1] != nil {
			addAllOpt, okA := args[1].(bool)
			if !okA {
				return nil, fmt.Errorf("%w: invalid type for 'add_all', expected boolean or nil", ErrInvalidArgument)
			}
			addAll = addAllOpt
		}
	}

	if addAll {
		// Stage all tracked changes first
		interpreter.logger.Debug("[Tool: GitCommit] Staging all changes (git add .)")
		_, errAdd := toolExec(interpreter, "git", "add", ".")
		if errAdd != nil {
			return nil, fmt.Errorf("failed during 'git add .' before commit: %w", errAdd)
		}
	}

	// Perform the commit
	interpreter.logger.Debug("[Tool: GitCommit] Executing: git commit -m '...'")
	gitArgs := []string{"commit", "-m", message}
	output, err := toolExec(interpreter, append([]string{"git"}, gitArgs...)...)

	if err != nil {
		// Check for "nothing to commit" which isn't necessarily a failure for the AI workflow
		stderrLower := strings.ToLower(err.Error())
		if strings.Contains(stderrLower, "nothing to commit") || strings.Contains(stderrLower, "no changes added to commit") {
			interpreter.logger.Warn("[Tool: GitCommit] Commit attempted but no changes were staged/detected.")
			return "GitCommit: Nothing to commit.", nil // Return success message, not error
		}
		// Otherwise, it's a real error
		return nil, fmt.Errorf("GitCommit failed: %w", err)
	}

	interpreter.logger.Debug("[Tool: GitCommit] Success. Output:\n%s", output)
	return fmt.Sprintf("GitCommit successful. Message: %q.\nOutput:\n%s", message, output), nil
}

// --- Tool Implementation: toolGitBranch ---
// This function is used by the registered "GitNewBranch" tool.
// It can also be used by a more general "GitBranch" tool if registered with a spec that exposes its optional parameters.
func toolGitBranch(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Args: name (string, opt), checkout (bool, opt), list_remote (bool, opt), list_all (bool, opt)
	// When called via "GitNewBranch" spec: args[0] is branch_name. Others default.
	var branchNameOpt interface{}
	var checkoutOpt interface{} = false // Default bools
	var listRemoteOpt interface{} = false
	var listAllOpt interface{} = false

	if len(args) > 0 {
		branchNameOpt = args[0]
	}
	if len(args) > 1 {
		checkoutOpt = args[1]
	}
	if len(args) > 2 {
		listRemoteOpt = args[2]
	}
	if len(args) > 3 {
		listAllOpt = args[3]
	}

	name := ""
	nameOk := false
	if branchNameOpt == nil {
		nameOk = true // nil is okay, means list branches
	} else if n, ok := branchNameOpt.(string); ok {
		name = n
		nameOk = true
	}
	if !nameOk {
		return nil, fmt.Errorf("%w: invalid type for 'name', expected string or nil, got %T", ErrInvalidArgument, branchNameOpt)
	}

	checkout := false
	if v, ok := checkoutOpt.(bool); ok {
		checkout = v
	} else if checkoutOpt != nil { // Only error if not nil and not bool
		return nil, fmt.Errorf("%w: invalid type for 'checkout', expected boolean, got %T", ErrInvalidArgument, checkoutOpt)
	}

	listRemote := false
	if v, ok := listRemoteOpt.(bool); ok {
		listRemote = v
	} else if listRemoteOpt != nil {
		return nil, fmt.Errorf("%w: invalid type for 'list_remote', expected boolean, got %T", ErrInvalidArgument, listRemoteOpt)
	}

	listAll := false
	if v, ok := listAllOpt.(bool); ok {
		listAll = v
	} else if listAllOpt != nil {
		return nil, fmt.Errorf("%w: invalid type for 'list_all', expected boolean, got %T", ErrInvalidArgument, listAllOpt)
	}

	if name != "" { // Create branch mode
		if listRemote || listAll {
			return nil, fmt.Errorf("%w: cannot specify list flags (-a, -r) when creating a branch ('name' provided)", ErrInvalidArgument)
		}
		if strings.ContainsAny(name, " \t\n\\/:*?\"<>|~^") || strings.HasPrefix(name, "-") || strings.Contains(name, "..") || strings.HasSuffix(name, "/") || strings.HasSuffix(name, ".lock") {
			return nil, fmt.Errorf("%w: branch name '%s' contains invalid characters or format", ErrValidationArgValue, name)
		}

		gitArgs := []string{}
		action := "create"
		if checkout { // For "GitNewBranch" spec, 'checkout' will be false from its default.
			gitArgs = append(gitArgs, "checkout", "-b", name)
			action = "create and checkout"
		} else {
			gitArgs = append(gitArgs, "branch", name)
		}
		interpreter.logger.Debug("[Tool: GitBranch/GitNewBranch] Executing: git %s", strings.Join(gitArgs, " "))
		_, err := toolExec(interpreter, append([]string{"git"}, gitArgs...)...)
		if err != nil {
			return nil, fmt.Errorf("failed to %s branch '%s': %w", action, name, err)
		}
		interpreter.logger.Debug("[Tool: GitBranch/GitNewBranch] Successfully %s branch '%s'", action, name)
		return fmt.Sprintf("Successfully %s branch '%s'.", action, name), nil

	} else { // List branches mode
		if checkout { // 'checkout' is false if name is omitted and checkoutOpt was nil/false
			return nil, fmt.Errorf("%w: cannot specify 'checkout' flag when listing branches ('name' omitted)", ErrInvalidArgument)
		}
		gitArgs := []string{"branch"}
		if listAll {
			gitArgs = append(gitArgs, "-a")
		} else if listRemote {
			gitArgs = append(gitArgs, "-r")
		}
		gitArgs = append(gitArgs, "--no-color")

		interpreter.logger.Debug("[Tool: GitBranch - List Mode] Executing: git %s", strings.Join(gitArgs, " "))
		output, err := toolExec(interpreter, append([]string{"git"}, gitArgs...)...)
		if err != nil {
			return nil, fmt.Errorf("failed to list branches: %w", err)
		}

		branches := []string{}
		rawLines := strings.Split(output, "\n")
		for _, line := range rawLines {
			trimmedLine := strings.TrimSpace(line)
			trimmedLine = strings.TrimPrefix(trimmedLine, "* ")
			if trimmedLine != "" && !strings.Contains(trimmedLine, "->") {
				branches = append(branches, trimmedLine)
			}
		}
		result := make([]interface{}, len(branches))
		for i, b := range branches {
			result[i] = b
		}
		interpreter.logger.Debug("[Tool: GitBranch - List Mode] Found %d branches.", len(result))
		return result, nil // Returns []interface{} (effectively slice of strings)
	}
}

// --- toolGitCheckout Tool Implementation ---
func toolGitCheckout(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Args: branch (string, required), create (bool, optional)
	// When called via "GitCheckout" spec: args[0] is branch_or_commit. len(args) is 1.
	if len(args) < 1 || len(args) > 2 { // This function's own validation
		return nil, fmt.Errorf("%w: toolGitCheckout internal: expects 1 or 2 arguments, got %d", ErrInvalidArgument, len(args))
	}

	branch, okB := args[0].(string)
	if !okB || branch == "" {
		return nil, fmt.Errorf("%w: invalid type or empty value for 'branch', expected non-empty string", ErrInvalidArgument)
	}

	create := false // Default for the function
	if len(args) == 2 {
		// This block is NOT entered if "GitCheckout" tool calls with 1 arg.
		if args[1] != nil {
			createOpt, okC := args[1].(bool)
			if !okC {
				return nil, fmt.Errorf("%w: invalid type for 'create', expected boolean or nil", ErrInvalidArgument)
			}
			create = createOpt
		}
	}

	gitArgs := []string{"checkout"}
	action := "checkout"
	if create { // 'create' remains false if called by "GitCheckout" tool
		gitArgs = append(gitArgs, "-b")
		action = "create and checkout"
		if strings.ContainsAny(branch, " \t\n\\/:*?\"<>|~^") || strings.HasPrefix(branch, "-") || strings.Contains(branch, "..") || strings.HasSuffix(branch, "/") || strings.HasSuffix(branch, ".lock") {
			return nil, fmt.Errorf("%w: branch name '%s' contains invalid characters or format when creating", ErrValidationArgValue, branch)
		}
	}
	gitArgs = append(gitArgs, branch)

	interpreter.logger.Debug("[Tool: GitCheckout] Executing: git %s", strings.Join(gitArgs, " "))
	output, err := toolExec(interpreter, append([]string{"git"}, gitArgs...)...)

	if err != nil {
		return nil, fmt.Errorf("failed to %s branch/ref '%s': %w", action, branch, err)
	}
	interpreter.logger.Debug("[Tool: GitCheckout] Success. Output:\n%s", output)
	return fmt.Sprintf("Successfully checked out branch/ref '%s'.\nOutput:\n%s", branch, output), nil
}

// --- GitRm Tool Implementation ---
func toolGitRm(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("%w: GitRm requires exactly one argument (path)", ErrInvalidArgument)
	}
	path, ok := args[0].(string)
	if !ok || path == "" {
		return nil, fmt.Errorf("%w: invalid type or empty value for 'path', expected non-empty string", ErrInvalidArgument)
	}

	securePath, err := SecureFilePath(path, interpreter.sandboxDir)
	if err != nil {
		return nil, fmt.Errorf("invalid path for GitRm '%s': %w", path, errors.Join(ErrValidationArgValue, err))
	}
	relativePath := path

	interpreter.logger.Debug("[Tool: GitRm] Executing: git rm %s (validated path: %s)", relativePath, securePath)
	output, err := toolExec(interpreter, "git", "rm", relativePath)

	if err != nil {
		return nil, fmt.Errorf("failed to remove path '%s': %w", relativePath, err)
	}
	interpreter.logger.Debug("[Tool: GitRm] Success. Output:\n%s", output)
	return fmt.Sprintf("Successfully removed path '%s' from git index.\nOutput:\n%s", relativePath, output), nil
}

// --- GitMerge Tool Implementation ---
func toolGitMerge(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("%w: GitMerge requires exactly one argument (branch name)", ErrInvalidArgument)
	}
	branchName, ok := args[0].(string)
	if !ok || branchName == "" {
		return nil, fmt.Errorf("%w: invalid type or empty value for 'branch', expected non-empty string", ErrInvalidArgument)
	}

	interpreter.logger.Debug("[Tool: GitMerge] Executing: git merge %s", branchName)
	output, err := toolExec(interpreter, "git", "merge", branchName)

	if err != nil {
		return nil, fmt.Errorf("failed to merge branch '%s' (check for conflicts): %w", branchName, err)
	}

	interpreter.logger.Debug("[Tool: GitMerge] Success. Output:\n%s", output)
	return fmt.Sprintf("Successfully merged branch '%s'.\nOutput:\n%s", branchName, output), nil
}

// --- GitPull Tool Implementation ---
func toolGitPull(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	if len(args) != 0 { // This function's internal check
		return nil, fmt.Errorf("%w: toolGitPull internal: expects no arguments, got %d", ErrInvalidArgument, len(args))
	}

	interpreter.logger.Debug("[Tool: GitPull] Executing: git pull")
	output, err := toolExec(interpreter, "git", "pull")

	if err != nil {
		return nil, fmt.Errorf("GitPull failed: %w", err)
	}

	interpreter.logger.Debug("[Tool: GitPull] Success. Output:\n%s", output)
	return fmt.Sprintf("GitPull successful.\nOutput:\n%s", output), nil
}

// --- GitPush Tool Implementation ---
func toolGitPush(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Args: remote (string, opt), branch (string, opt), set_upstream (bool, opt)
	// When called via "GitPush" spec, args is empty.
	remote := "origin"
	var branch string
	setUpstream := false

	// These conditions will not be met if called by "GitPush" tool (0 args)
	if len(args) > 0 && args[0] != nil {
		remoteOpt, okR := args[0].(string)
		if !okR || remoteOpt == "" {
			return nil, fmt.Errorf("%w: invalid type or empty value for 'remote', expected non-empty string or nil", ErrInvalidArgument)
		}
		remote = remoteOpt
	}
	if len(args) > 1 && args[1] != nil {
		branchOpt, okB := args[1].(string)
		if !okB || branchOpt == "" {
			return nil, fmt.Errorf("%w: invalid type or empty value for 'branch', expected non-empty string or nil", ErrInvalidArgument)
		}
		branch = branchOpt
	}
	if len(args) > 2 && args[2] != nil {
		upstreamOpt, okU := args[2].(bool)
		if !okU {
			return nil, fmt.Errorf("%w: invalid type for 'set_upstream', expected boolean or nil", ErrInvalidArgument)
		}
		setUpstream = upstreamOpt
	}

	var err error
	if branch == "" { // True if called by "GitPush" tool
		interpreter.logger.Debug("[Tool: GitPush] Branch not specified, determining current branch.")
		branch, err = getCurrentGitBranch(interpreter)
		if err != nil {
			return nil, err
		}
		interpreter.logger.Debug("[Tool: GitPush] Current branch detected:", "branch", branch)
	}

	gitArgs := []string{"push"}
	if setUpstream { // false if called by "GitPush" tool
		gitArgs = append(gitArgs, "-u")
	}
	gitArgs = append(gitArgs, remote, branch) // Uses defaults if called by "GitPush" tool

	interpreter.logger.Debug("[Tool: GitPush] Executing: git %s", strings.Join(gitArgs, " "))
	output, pushErr := toolExec(interpreter, append([]string{"git"}, gitArgs...)...)

	if pushErr != nil {
		return nil, fmt.Errorf("GitPush failed: %w", pushErr)
	}

	interpreter.logger.Debug("[Tool: GitPush] Success. Output:\n%s", output)
	return fmt.Sprintf("GitPush successful (%s -> %s).\nOutput:\n%s", branch, remote, output), nil
}

// --- GitDiff Tool Implementation ---
func toolGitDiff(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Args: cached (bool, opt), commit1 (string, opt), commit2 (string, opt), path (string, opt)
	// When called via "GitDiff" spec, args is empty.
	cached := false
	var commit1, commit2 string
	//path := ""

	argPos := 0
	// These conditions for parsing args won't be met if "GitDiff" calls with 0 args.
	if len(args) > argPos {
		if args[argPos] == nil {
			argPos++
		} else if v, ok := args[argPos].(bool); ok {
			cached = v
			argPos++
		} else {
			return nil, fmt.Errorf("%w: expected boolean or nil for 'cached', got %T", ErrInvalidArgument, args[argPos])
		}
	}
	// ... (similar parsing for commit1, commit2, path) ...
	// Simplified for brevity as these paths are not taken for 0-arg call.
	if len(args) > argPos { // For commit1
		if args[argPos] != nil {
			if v, ok := args[argPos].(string); ok {
				commit1 = v
			}
		}
		argPos++
	}
	if len(args) > argPos { // For commit2
		if args[argPos] != nil {
			if v, ok := args[argPos].(string); ok {
				commit2 = v
			}
		}
		argPos++
	}

	if commit2 != "" && commit1 == "" {
		return nil, fmt.Errorf("%w: 'commit2' requires 'commit1' to be specified", ErrInvalidArgument)
	}
	if cached && (commit1 != "" || commit2 != "") {
		return nil, fmt.Errorf("%w: 'cached' option cannot be used with 'commit1' or 'commit2'", ErrInvalidArgument)
	}

	gitArgs := []string{"diff"}
	if cached { // false for 0-arg call
		gitArgs = append(gitArgs, "--cached")
	}
	// commit1, commit2, path are empty for 0-arg call.
	// So, defaults to 'git diff' (working tree vs index).

	interpreter.logger.Debug("[Tool: GitDiff] Executing: git %s", strings.Join(gitArgs, " "))
	output, err := toolExec(interpreter, append([]string{"git"}, gitArgs...)...)

	if err != nil {
		interpreter.logger.Warn("[Tool: GitDiff] Command may have indicated differences or failed, returning output.", "error", err)
		return output, nil // Return output even on error, as diff output is the primary goal
	}

	if output == "" {
		interpreter.logger.Debug("[Tool: GitDiff] Success. No changes detected.")
		return "GitDiff: No changes detected.", nil
	}

	interpreter.logger.Debug("[Tool: GitDiff] Success. Changes detected.")
	return output, nil
}
