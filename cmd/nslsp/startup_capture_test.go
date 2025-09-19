// NeuroScript Version: 0.7.0
// File version: 1
// Purpose: A diagnostic test to build and capture the raw initial stdout of the nslsp executable to identify unexpected startup text.
// filename: cmd/nslsp/startup_capture_test.go
// nlines: 70
// risk_rating: LOW

package main

import (
	"bufio"
	"context"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestCaptureServerStartupOutput(t *testing.T) {
	// 1. Build the nslsp executable to a temporary directory.
	// This ensures we are testing the current state of the code.
	tempDir := t.TempDir()
	executableName := "nslsp_test_capture"
	if runtime.GOOS == "windows" {
		executableName += ".exe"
	}
	executablePath := filepath.Join(tempDir, executableName)

	buildCmd := exec.Command("go", "build", "-o", executablePath, ".")
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build nslsp executable for test: %v\nOutput:\n%s", err, string(output))
	}
	t.Logf("Successfully built test executable at: %s", executablePath)

	// 2. Run the newly built executable and capture its output.
	t.Log("This test will capture stdout for 2 seconds to check for unexpected shell startup messages.")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	runCmd := exec.CommandContext(ctx, executablePath)
	stdout, err := runCmd.StdoutPipe()
	if err != nil {
		t.Fatalf("Failed to get stdout pipe: %v", err)
	}

	if err := runCmd.Start(); err != nil {
		t.Fatalf("Failed to start server process: %v", err)
	}

	// 3. Read from stdout and report any findings.
	scanner := bufio.NewScanner(stdout)
	var capturedLines []string
	for scanner.Scan() {
		capturedLines = append(capturedLines, scanner.Text())
	}

	// Wait for the process to be cancelled by the context.
	_ = runCmd.Wait()

	if len(capturedLines) > 0 {
		// If we captured *any* output, the test fails. Stdout must be clean.
		capturedText := strings.Join(capturedLines, "\n")
		t.Fatalf("FAIL: Captured unexpected output on stdout, which will break LSP communication.\n--- OFFENDING TEXT ---\n%s\n----------------------", capturedText)
	}

	t.Log("SUCCESS: No output was captured from stdout.")
}
