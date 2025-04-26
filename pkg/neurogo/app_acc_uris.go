// filename: pkg/neurogo/app_agent.go
package neurogo

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/core"
	"github.com/google/generative-ai-go/genai"
)

// ... Constants, runAgentMode setup, other functions ... (largely unchanged) ...

// updateAccumulatedURIs lists files after a sync and updates the shared URI list.
// FIXED AGAIN: Correctly calculates relative path and prefix for root case.
func updateAccumulatedURIs(
	ctx context.Context,
	a *App,
	llmClient *core.LLMClient,
	absSyncedDir string, // Absolute path of the directory that was just synced
	syncFilter string, // The filter used for the sync
	accumulatedContextURIs *[]string, // Pointer to the slice to update
) {
	a.Logger.Info("Listing API files to update context for next turn...")
	a.Logger.Debug("Updating context for synced directory: %s", absSyncedDir)

	cleanSandboxDir := filepath.Clean(a.Config.SandboxDir)
	absCleanSandboxDir, err := filepath.Abs(cleanSandboxDir)
	if err != nil {
		a.Logger.Error("Error getting absolute path for sandbox dir %q: %v", cleanSandboxDir, err)
		fmt.Println("[AGENT] Warning: Could not determine absolute sandbox path for context update.")
		return
	}
	absCleanSandboxDir = filepath.Clean(absCleanSandboxDir)
	a.Logger.Debug("Context update using absolute sandbox root: %s", absCleanSandboxDir)

	// --- Start FIX V3 for Relative Path & Prefix ---
	syncDirRel := ""
	prefix := "" // Default prefix is empty (matches root files)

	// Only calculate relative path if synced dir is NOT the same as sandbox root
	if absSyncedDir != absCleanSandboxDir {
		relPath, relErr := filepath.Rel(absCleanSandboxDir, absSyncedDir)
		if relErr != nil {
			a.Logger.Error("Cannot get relative path for synced dir %q relative to sandbox %q: %v", absSyncedDir, absCleanSandboxDir, relErr)
			fmt.Println("[AGENT] Warning: Could not determine relative path for context update filtering.")
			return // Abort if we can't get relative path when needed
		}
		syncDirRel = filepath.ToSlash(relPath)
		// Ensure prefix ends with slash ONLY if syncDirRel is not empty or "."
		if syncDirRel != "" && syncDirRel != "." {
			prefix = syncDirRel + "/"
		}
		a.Logger.Debug("Calculated relative sync path: %s, using prefix: '%s'", syncDirRel, prefix)
	} else {
		a.Logger.Debug("Synced directory is the sandbox root. Using empty prefix ''.")
		// syncDirRel remains "", prefix remains ""
	}
	// --- End FIX V3 ---

	apiFiles, listErr := core.HelperListApiFiles(ctx, llmClient.Client(), a.DebugLog)
	if listErr != nil {
		a.Logger.Error("Failed list API files: %v", listErr)
		fmt.Println("[AGENT] Warning: Ctx update failed.")
		return
	}
	a.Logger.Debug("Found %d total API files to filter for context update.", len(apiFiles))

	urisCollected := 0
	newURIs := []string{}
	a.Logger.Debug("Filtering API files using prefix: [%s]", prefix) // Show prefix clearly

	for _, file := range apiFiles {
		if file.DisplayName == "" || file.State != genai.FileStateActive || file.URI == "" {
			continue
		}
		a.Logger.Debug("Checking API DisplayName: [%s]", file.DisplayName)

		// Use the calculated prefix for matching
		if strings.HasPrefix(file.DisplayName, prefix) {
			if syncFilter != "" { /* ... filter check ... */
				match, _ := filepath.Match(syncFilter, filepath.Base(file.DisplayName))
				if !match {
					a.Logger.Debug("-> Matched prefix, SKIPPED by filter: %s", file.DisplayName)
					continue
				}
				a.Logger.Debug("-> Matched prefix AND filter: %s", file.DisplayName)
			} else {
				a.Logger.Debug("-> Matched prefix (no filter): %s", file.DisplayName)
			}
			newURIs = append(newURIs, file.URI)
			urisCollected++
		} else {
			a.Logger.Debug("-> Did NOT match prefix.") // Log non-matches
		}
	} // End file loop

	// Update accumulated URIs (unchanged)
	uriSet := make(map[string]bool)
	for _, uri := range *accumulatedContextURIs {
		uriSet[uri] = true
	}
	for _, uri := range newURIs {
		uriSet[uri] = true
	}
	*accumulatedContextURIs = (*accumulatedContextURIs)[:0]
	for uri := range uriSet {
		*accumulatedContextURIs = append(*accumulatedContextURIs, uri)
	}
	a.Logger.Info("Collected %d URIs from sync (Dir: '%s', Filter: '%s'). Total accumulated URIs: %d", urisCollected, absSyncedDir, syncFilter, len(*accumulatedContextURIs))
	displayDir := filepath.Base(absSyncedDir)
	if absSyncedDir == absCleanSandboxDir && a.Config.SandboxDir == "." {
		displayDir = "."
	}
	fmt.Printf("[AGENT] Context updated with %d files from '%s'.\n", urisCollected, displayDir)
}

// ... rest of app_agent.go, handleMultilineInput, etc. ...
