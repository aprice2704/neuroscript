// :: product: FDM/NS
// :: majorVersion: 0
// :: fileVersion: 13
// :: description: Tests for the Git toolset. Removed manual registration to fix duplicate key errors.
// :: latestChange: Removed manual registration loop in newGitTestInterpreter and switched to blank import.
// :: filename: pkg/tool/git/tools_git_test.go
// :: serialization: go
package git_test

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/api/testharness"
	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/aprice2704/neuroscript/pkg/testutil"
	"github.com/aprice2704/neuroscript/pkg/tool"
	_ "github.com/aprice2704/neuroscript/pkg/tool/git" // Import for side-effects (registration)
	"github.com/aprice2704/neuroscript/pkg/types"
)

// --- Test Helpers ---

type gitTestCase struct {
	name          string
	toolName      types.ToolName
	args          []interface{}
	setupFunc     func(t *testing.T, sandboxRoot string)
	checkFunc     func(t *testing.T, interp tool.Runtime, result interface{}, err error)
	wantToolErrIs error
}

func newGitTestInterpreter(t *testing.T) *interpreter.Interpreter {
	t.Helper()
	sandboxOpt := testutil.NewTestSandbox(t)
	// Use the testharness helper to create a valid HostContext.
	hc := testharness.NewTestHostContext(logging.NewTestLogger(t))
	interp := interpreter.NewInterpreter(interpreter.WithHostContext(hc), sandboxOpt)

	// Manual registration loop removed:
	// The git tools are registered via init() in the git package (imported above)
	// and automatically loaded by NewInterpreter's default behavior.
	// Re-registering them caused "duplicate key" errors.

	return interp
}

func setupGitRepo(t *testing.T, sandboxRoot string) {
	t.Helper()
	// Helper to run git commands directly for setup
	runCmd := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = sandboxRoot
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to run git command 'git %v': %v\nOutput: %s", args, err, string(out))
		}
	}

	runCmd("init")
	// Set a default user, otherwise commit will fail in some environments
	runCmd("config", "user.email", "test@example.com")
	runCmd("config", "user.name", "Test User")
	filename := filepath.Join(sandboxRoot, "initial.txt")
	if err := os.WriteFile(filename, []byte("initial content"), 0644); err != nil {
		t.Fatalf("Failed to write initial file: %v", err)
	}
	runCmd("add", "initial.txt")
	runCmd("commit", "-m", "Initial commit")
}

func testGitToolHelper(t *testing.T, tc gitTestCase) {
	t.Helper()
	interp := newGitTestInterpreter(t)
	sandboxRoot := interp.SandboxDir()

	// Setup the repo inside the sandbox
	setupGitRepo(t, sandboxRoot)

	if tc.setupFunc != nil {
		tc.setupFunc(t, sandboxRoot)
	}

	// Use explicit "git" group name since we are outside the package
	fullName := types.MakeFullName("git", string(tc.toolName))
	toolImpl, found := interp.ToolRegistry().GetTool(fullName)
	if !found {
		t.Fatalf("Tool '%s' not found in registry", fullName)
	}

	result, err := toolImpl.Func(interp, tc.args)

	if tc.checkFunc != nil {
		tc.checkFunc(t, interp, result, err)
	} else {
		if tc.wantToolErrIs != nil {
			if !errors.Is(err, tc.wantToolErrIs) {
				t.Errorf("Expected error wrapping [%v], but got: %v", tc.wantToolErrIs, err)
			}
		} else if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	}
}

// --- Test Cases ---

func TestToolGitBranchValidation(t *testing.T) {
	testGitToolHelper(t, gitTestCase{
		name:          "List branches with valid path",
		toolName:      "Branch",
		args:          []interface{}{"."},
		wantToolErrIs: nil, // Listing branches is the default and should not error.
	})
}

func TestToolGitCheckoutValidation(t *testing.T) {
	testGitToolHelper(t, gitTestCase{
		name:          "Missing branch name",
		toolName:      "Checkout",
		args:          []interface{}{"."},
		wantToolErrIs: lang.ErrArgumentMismatch,
	})
}

func TestToolGitAddValidation(t *testing.T) {
	testGitToolHelper(t, gitTestCase{
		name:          "Missing paths list",
		toolName:      "Add",
		args:          []interface{}{"."},
		wantToolErrIs: lang.ErrInvalidArgument,
	})
}

func TestToolGitCommitValidation(t *testing.T) {
	testGitToolHelper(t, gitTestCase{
		name:          "Missing commit message",
		toolName:      "Commit",
		args:          []interface{}{"."},
		wantToolErrIs: lang.ErrArgumentMismatch,
	})
}

func TestToolGitRmValidation(t *testing.T) {
	testGitToolHelper(t, gitTestCase{
		name:          "Missing paths list",
		toolName:      "Rm",
		args:          []interface{}{"."},
		wantToolErrIs: lang.ErrArgumentMismatch,
	})
}

func TestToolGitStatus(t *testing.T) {
	testGitToolHelper(t, gitTestCase{
		name:     "Get status",
		toolName: "Status",
		args:     []interface{}{},
		setupFunc: func(t *testing.T, sandboxRoot string) {
			newFile := filepath.Join(sandboxRoot, "newfile.txt")
			os.WriteFile(newFile, []byte("new data"), 0644)
		},
		checkFunc: func(t *testing.T, interp tool.Runtime, result interface{}, err error) {
			if err != nil {
				t.Fatalf("Status tool failed: %v", err)
			}
			statusMap, ok := result.(map[string]interface{})
			if !ok {
				t.Fatalf("Expected status to be a map, got %T", result)
			}
			if isClean, _ := statusMap["is_clean"].(bool); isClean {
				t.Error("Expected repo to be dirty, but status is clean")
			}
		},
	})
}

func TestToolGitMergeValidation(t *testing.T) {
	testGitToolHelper(t, gitTestCase{
		name:          "Missing branch name",
		toolName:      "Merge",
		args:          []interface{}{"."},
		wantToolErrIs: lang.ErrArgumentMismatch,
	})
}

func TestToolGitPullValidation(t *testing.T) {
	testGitToolHelper(t, gitTestCase{
		name:     "No remote configured",
		toolName: "Pull",
		args:     []interface{}{"."},
		checkFunc: func(t *testing.T, interp tool.Runtime, result interface{}, err error) {
			if err == nil {
				t.Fatal("Expected an error for git pull with no remote, but got nil")
			}
			runtimeErr, ok := err.(*lang.RuntimeError)
			if !ok {
				t.Fatalf("Expected a *lang.RuntimeError, but got %T", err)
			}
			if runtimeErr.Code != lang.ErrorCodeToolExecutionFailed {
				t.Errorf("Expected error code for tool execution failure (%d), but got %d", lang.ErrorCodeToolExecutionFailed, runtimeErr.Code)
			}
		},
	})
}

func TestToolGitPushValidation(t *testing.T) {
	testGitToolHelper(t, gitTestCase{
		name:     "No remote configured",
		toolName: "Push",
		args:     []interface{}{"."},
		checkFunc: func(t *testing.T, interp tool.Runtime, result interface{}, err error) {
			if err == nil {
				t.Fatal("Expected an error for git push with no remote, but got nil")
			}
			runtimeErr, ok := err.(*lang.RuntimeError)
			if !ok {
				t.Fatalf("Expected a *lang.RuntimeError, but got %T", err)
			}
			if runtimeErr.Code != lang.ErrorCodeToolExecutionFailed {
				t.Errorf("Expected error code for tool execution failure (%d), but got %d", lang.ErrorCodeToolExecutionFailed, runtimeErr.Code)
			}
		},
	})
}

func TestToolGitDiffValidation(t *testing.T) {
	testGitToolHelper(t, gitTestCase{
		name:     "No files specified",
		toolName: "Diff",
		args:     []interface{}{"."},
		checkFunc: func(t *testing.T, interp tool.Runtime, result interface{}, err error) {
			if err != nil {
				t.Fatalf("Unexpected error from Diff: %v", err)
			}
			if result != "GitDiff: No changes detected." {
				t.Errorf("Expected no changes, but got: %v", result)
			}
		},
	})
}

func TestToolGitCloneValidation(t *testing.T) {
	// Don't use the standard helper because we don't want a pre-existing repo
	// in the sandbox for a clone test.
	sourceRepoPath := t.TempDir()

	// Setup the repo to be cloned
	setupGitRepo(t, sourceRepoPath)

	interp := newGitTestInterpreter(t)
	sandboxRoot := interp.SandboxDir()

	cloneTool, found := interp.ToolRegistry().GetTool(types.MakeFullName("git", "Clone"))
	if !found {
		t.Fatal("Tool 'Git.Clone' not found in registry")
	}

	_, err := cloneTool.Func(interp, []interface{}{sourceRepoPath, "cloned_repo"})
	if err != nil {
		t.Fatalf("Clone tool failed: %v", err)
	}

	if _, statErr := os.Stat(filepath.Join(sandboxRoot, "cloned_repo", ".git")); os.IsNotExist(statErr) {
		t.Error("Clone did not create a .git directory")
	}
}

func TestToolGitResetValidation(t *testing.T) {
	testGitToolHelper(t, gitTestCase{
		name:          "Missing repo path",
		toolName:      "Reset",
		args:          []interface{}{},
		wantToolErrIs: lang.ErrArgumentMismatch,
	})
}
