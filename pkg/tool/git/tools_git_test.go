// NeuroScript Version: 0.4.1
// File version: 19
// Purpose: Final corrected test file to pass all remaining Git tool tests, with a more robust setup.
// filename: pkg/tool/git/tools_git_test.go
// nlines: 235
// risk_rating: MEDIUM

package git

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

// MakeArgs is a convenience function to create a slice of interfaces, useful for constructing tool arguments programmatically.
func MakeArgs(vals ...interface{}) []interface{} {
	if vals == nil {
		return []interface{}{}
	}
	return vals
}

// initGitRepoForTest creates a standard repo with an initial commit containing a README.
func initGitRepoForTest(t *testing.T, baseDir string, repoSubDir string) (string, error) {
	t.Helper()
	repoPath := filepath.Join(baseDir, repoSubDir)
	if err := os.MkdirAll(repoPath, 0755); err != nil {
		return "", err
	}
	gitCmds := [][]string{
		{"init"},
		{"checkout", "-b", "main"},
		{"config", "user.email", "test@example.com"},
		{"config", "user.name", "Test User"},
	}
	for _, cmdArgs := range gitCmds {
		cmd := exec.Command("git", cmdArgs...)
		cmd.Dir = repoPath
		if _, err := cmd.CombinedOutput(); err != nil {
			return "", err
		}
	}
	// Create and commit a README file for a more standard initial state.
	if err := os.WriteFile(filepath.Join(repoPath, "README.md"), []byte("init"), 0644); err != nil {
		return "", err
	}
	addCmd := exec.Command("git", "add", "README.md")
	addCmd.Dir = repoPath
	if _, err := addCmd.CombinedOutput(); err != nil {
		return "", err
	}
	commitCmd := exec.Command("git", "commit", "-m", "initial commit")
	commitCmd.Dir = repoPath
	if _, err := commitCmd.CombinedOutput(); err != nil {
		return "", err
	}

	return repoPath, nil
}

// gitTestCase defines the structure for git tool tests.
type gitTestCase struct {
	Name      string
	Args      []interface{}
	WantErrIs error
	Setup     func(t *testing.T, repoPath string)
}

// testGitToolHelper runs a git tool test case with proper repo setup.
func testGitToolHelper(t *testing.T, toolName string, tc gitTestCase) {
	t.Helper()
	interp := interpreter.NewInterpreter()
	repoSubDir, ok := tc.Args[0].(string)
	if !ok {
		t.Fatalf("Test case '%s' must have a string repo path as first argument.", tc.Name)
	}
	// Use a fresh temp dir for each test to avoid conflicts
	sandboxDir := t.TempDir()
	interp.SetSandboxDir(sandboxDir)

	absRepoPath, err := initGitRepoForTest(t, interp.SandboxDir(), repoSubDir)
	if err != nil {
		t.Fatalf("Test case '%s': Failed to set up git repo: %v", tc.Name, err)
	}
	if tc.Setup != nil {
		tc.Setup(t, absRepoPath)
	}
	toolImpl, found := interp.ToolRegistry().GetTool(toolName)
	if !found {
		t.Fatalf("Tool '%s' not found in registry", toolName)
	}
	t.Run(tc.Name, func(t *testing.T) {
		_, testErr := toolImpl.Func(interp, tc.Args)
		if tc.WantErrIs != nil {
			if testErr == nil {
				t.Errorf("Expected error wrapping [%v], but got nil", tc.WantErrIs)
			} else if !errors.Is(testErr, tc.WantErrIs) {
				t.Errorf("Expected error wrapping [%v], but got: %v", tc.WantErrIs, testErr)
			}
		} else if testErr != nil {
			t.Errorf("Unexpected error: %v", testErr)
		}
	})
}

const dummyRepoPath = "repo"

func TestToolGitBranchValidation(t *testing.T) {
	testCases := []gitTestCase{
		{Name: "Correct_Args_(Create)", Args: MakeArgs(dummyRepoPath, "new-feature"), WantErrIs: nil},
		{Name: "Wrong_Arg_Type_(Name)", Args: MakeArgs(dummyRepoPath, 123), WantErrIs: lang.ErrInvalidArgument},
	}
	for _, tc := range testCases {
		testGitToolHelper(t, "Git.Branch", tc)
	}
}

func TestToolGitCheckoutValidation(t *testing.T) {
	testCases := []gitTestCase{
		{Name: "Correct_Args_(Checkout)", Args: MakeArgs(dummyRepoPath, "main"), WantErrIs: nil},
		{Name: "Correct_Args_(Create_and_Checkout)", Args: MakeArgs(dummyRepoPath, "new-feature", true), WantErrIs: nil},
	}
	for _, tc := range testCases {
		testGitToolHelper(t, "Git.Checkout", tc)
	}
}

func TestToolGitRmValidation(t *testing.T) {
	setupWithFile := func(t *testing.T, repoPath string) {
		t.Helper()
		// Create nested directory for the file.
		nestedDir := filepath.Join(repoPath, "path", "to")
		if err := os.MkdirAll(nestedDir, 0755); err != nil {
			t.Fatalf("Setup failed to create nested directory: %v", err)
		}
		filePath := filepath.Join(nestedDir, "file.txt")
		if err := os.WriteFile(filePath, []byte("content"), 0644); err != nil {
			t.Fatalf("Setup failed to write file: %v", err)
		}
		addCmd := exec.Command("git", "add", ".")
		addCmd.Dir = repoPath
		if out, err := addCmd.CombinedOutput(); err != nil {
			t.Fatalf("Setup failed to 'git add': %v\nOutput: %s", err, string(out))
		}
		commitCmd := exec.Command("git", "commit", "-m", "add file")
		commitCmd.Dir = repoPath
		if out, err := commitCmd.CombinedOutput(); err != nil {
			t.Fatalf("Setup failed to 'git commit': %v\nOutput: %s", err, string(out))
		}
	}
	testCases := []gitTestCase{
		{Name: "Correct_Args_(Single_Path_String)", Args: MakeArgs(dummyRepoPath, "path/to/file.txt"), WantErrIs: nil, Setup: setupWithFile},
		{Name: "Wrong_Arg_Type_(Path)", Args: MakeArgs(dummyRepoPath, 123), WantErrIs: lang.ErrInvalidArgument},
	}
	for _, tc := range testCases {
		testGitToolHelper(t, "Git.Rm", tc)
	}
}

func TestToolGitMergeValidation(t *testing.T) {
	setupWithBranch := func(t *testing.T, repoPath string) {
		t.Helper()
		// Create and switch to the develop branch
		checkoutCmd := exec.Command("git", "checkout", "-b", "develop")
		checkoutCmd.Dir = repoPath
		if out, err := checkoutCmd.CombinedOutput(); err != nil {
			t.Fatalf("Setup failed to create 'develop' branch: %v\nOutput: %s", err, string(out))
		}
		// Create a new commit on the develop branch
		filePath := filepath.Join(repoPath, "dev-file.txt")
		if err := os.WriteFile(filePath, []byte("dev content"), 0644); err != nil {
			t.Fatalf("Setup failed to write dev file: %v", err)
		}
		addCmd := exec.Command("git", "add", ".")
		addCmd.Dir = repoPath
		if out, err := addCmd.CombinedOutput(); err != nil {
			t.Fatalf("Setup failed to 'git add' on develop: %v\nOutput: %s", err, string(out))
		}
		commitCmd := exec.Command("git", "commit", "-m", "commit on develop")
		commitCmd.Dir = repoPath
		if out, err := commitCmd.CombinedOutput(); err != nil {
			t.Fatalf("Setup failed to 'git commit' on develop: %v\nOutput: %s", err, string(out))
		}
		// Switch back to main branch to be ready for the merge
		checkoutMainCmd := exec.Command("git", "checkout", "main")
		checkoutMainCmd.Dir = repoPath
		if out, err := checkoutMainCmd.CombinedOutput(); err != nil {
			t.Fatalf("Setup failed to checkout main: %v\nOutput: %s", err, string(out))
		}
	}
	testCases := []gitTestCase{
		{Name: "Correct_Args", Args: MakeArgs(dummyRepoPath, "develop"), WantErrIs: nil, Setup: setupWithBranch},
		{Name: "Wrong_Arg_Type_(Branch)", Args: MakeArgs(dummyRepoPath, 123), WantErrIs: lang.ErrInvalidArgument},
	}
	for _, tc := range testCases {
		testGitToolHelper(t, "Git.Merge", tc)
	}
}

func TestToolGitPullValidation(t *testing.T) {
	// Only testing validation, as functional test requires a remote.
	testCases := []gitTestCase{
		{Name: "Wrong_Arg_Type_(Remote_Name)", Args: MakeArgs(dummyRepoPath, 123, "main"), WantErrIs: lang.ErrInvalidArgument},
	}
	for _, tc := range testCases {
		testGitToolHelper(t, "Git.Pull", tc)
	}
}

func TestToolGitPushValidation(t *testing.T) {
	// Only testing validation, as functional test requires a remote.
	testCases := []gitTestCase{
		{Name: "Wrong_Arg_Type_(Branch_Name)", Args: MakeArgs(dummyRepoPath, "origin", false), WantErrIs: lang.ErrInvalidArgument},
	}
	for _, tc := range testCases {
		testGitToolHelper(t, "Git.Push", tc)
	}
}

func TestToolGitDiffValidation(t *testing.T) {
	testCases := []gitTestCase{
		{Name: "Correct_Args_(Cached_Only)", Args: MakeArgs(dummyRepoPath, true), WantErrIs: nil},
		{Name: "Wrong_Arg_Type", Args: MakeArgs(dummyRepoPath, "not-bool"), WantErrIs: lang.ErrInvalidArgument},
	}
	for _, tc := range testCases {
		testGitToolHelper(t, "Git.Diff", tc)
	}
}
