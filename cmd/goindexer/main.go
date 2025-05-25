// NeuroScript Go Indexer
// File version: 2.2.6 // Ensure ComponentIndex uses IndexSchemaVersion and new fields.
// Purpose: Main package for the Go code indexer. Defines data structures and orchestrates parsing.
// filename: main.go
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"go/token"
	"log"
	"os"
	"path/filepath"

	// "sort" // Not used directly in this main, but parser might use it via helpers
	"strings"
	"time"

	"github.com/aprice2704/neuroscript/pkg/goindex" // Using the types from the separate package
	// "github.com/bmatcuk/doublestar/v4" // Not used in the provided main.go snippet
)

var (
	repoRootPath   string
	repoModulePath string
)

// ComponentDefinition defines a logical component of the project for indexing.
type ComponentDefinition struct {
	Name             string   // e.g., "core", "nslsp"
	PathPrefixes     []string // Directory prefixes relative to repo root (e.g., "pkg/core", "pkg/nslsp")
	IndexFile        string   // Filename for this component's index (e.g., "core_index.json")
	ComponentRelPath string   // The representative relative path for the component from repo root for ProjectIndex
}

var defaultComponentDefs = []ComponentDefinition{
	{Name: "core", PathPrefixes: []string{"pkg/core"}, IndexFile: "core_index.json", ComponentRelPath: "pkg/core"},
	{Name: "nslsp", PathPrefixes: []string{"pkg/nslsp"}, IndexFile: "nslsp_index.json", ComponentRelPath: "pkg/nslsp"},
	{Name: "goindex_types", PathPrefixes: []string{"pkg/goindex"}, IndexFile: "goindex_types_index.json", ComponentRelPath: "pkg/goindex"},
	{Name: "neurodata", PathPrefixes: []string{"pkg/neurodata"}, IndexFile: "neurodata_index.json", ComponentRelPath: "pkg/neurodata"},
	{Name: "neurogo", PathPrefixes: []string{"pkg/neurogo"}, IndexFile: "neurogo_index.json", ComponentRelPath: "pkg/neurogo"},
	{Name: "nspatch_pkg", PathPrefixes: []string{"pkg/nspatch"}, IndexFile: "nspatch_pkg_index.json", ComponentRelPath: "pkg/nspatch"},
	{Name: "adapters", PathPrefixes: []string{"pkg/adapters"}, IndexFile: "adapters_index.json", ComponentRelPath: "pkg/adapters"},
	{Name: "toolsets", PathPrefixes: []string{"pkg/toolsets"}, IndexFile: "toolsets_index.json", ComponentRelPath: "pkg/toolsets"},
	// Assuming "commands" component aggregates multiple top-level cmd directories
	{Name: "commands", PathPrefixes: []string{"cmd/ns", "cmd/ns-lsp", "cmd/goindexer"}, IndexFile: "commands_index.json", ComponentRelPath: "cmd"},
	{Name: "project_other", PathPrefixes: []string{"."}, IndexFile: "project_other_index.json", ComponentRelPath: "."}, // Catch-all
}

const defaultComponentName = "project_other" // Changed default to project_other to align with its definition

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	startTime := time.Now()

	rootDir := flag.String("root", ".", "Root directory of the Go project")
	outputDir := flag.String("output", ".", "Directory to save the index files")
	flag.Parse()

	var err error
	repoRootPath, repoModulePath, err = findRepoPaths(*rootDir)
	if err != nil {
		log.Fatalf("Error finding repo paths: %v", err)
	}
	log.Printf("Repository root: %s, Module path: %s", repoRootPath, repoModulePath)

	err = os.MkdirAll(*outputDir, 0755)
	if err != nil {
		log.Fatalf("Error creating output directory %q: %v", *outputDir, err)
	}

	fileSet := token.NewFileSet()
	componentIndexes := make(map[string]*goindex.ComponentIndex)
	gitBranch, gitCommitHash := getGitInfo(repoRootPath)

	projectIndex := goindex.ProjectIndex{
		ProjectRootModulePath: repoModulePath,
		IndexSchemaVersion:    "project_index_v2.0.0", // Align with types.go version if desired
		Components:            make(map[string]goindex.ComponentIndexFileEntry),
		LastIndexedTimestamp:  time.Now().Format(time.RFC3339),
		GitBranch:             gitBranch,
		GitCommitHash:         gitCommitHash,
	}

	for _, def := range defaultComponentDefs {
		componentIndexes[def.Name] = &goindex.ComponentIndex{
			ComponentName:      def.Name,
			ComponentPath:      def.ComponentRelPath,
			IndexSchemaVersion: "component_index_v2.0.0", // Align with types.go version
			Packages:           make(map[string]*goindex.PackageDetail),
			NeuroScriptTools:   make([]goindex.NeuroScriptToolDetail, 0),
			LastIndexed:        projectIndex.LastIndexedTimestamp, // Use same timestamp for initial setup
			GitBranch:          gitBranch,
			GitCommitHash:      gitCommitHash,
		}
	}
	// No separate defaultComponentName initialization if "project_other" with path "." covers it.
	// Ensure "project_other" is appropriately defined in defaultComponentDefs.

	log.Println("Starting to scan and parse Go files...")
	err = filepath.Walk(repoRootPath, func(path string, info os.FileInfo, errWalk error) error {
		if errWalk != nil {
			log.Printf("Prevented directory walk from failing: %v (Path: %s)", errWalk, path)
			return nil
		}

		if info.IsDir() {
			dirName := info.Name()
			if dirName == "vendor" || dirName == ".git" || (strings.HasPrefix(dirName, ".") && dirName != "." && dirName != "..") {
				// log.Printf("Skipping directory: %s", path)
				return filepath.SkipDir
			}
			return nil
		}

		if strings.HasSuffix(info.Name(), ".go") && !strings.HasSuffix(info.Name(), "_test.go") {
			relPathToRepoRoot, err := filepath.Rel(repoRootPath, path)
			if err != nil {
				log.Printf("    Error getting relative path for %q: %v", path, err)
				return nil
			}
			relPathToRepoRoot = filepath.ToSlash(relPathToRepoRoot)

			assignedComponentDef := assignFileToComponent(relPathToRepoRoot, defaultComponentDefs)
			currentComponentIndex := componentIndexes[assignedComponentDef.Name]

			// Path of the file relative to its assigned component's primary path prefix
			filePathInComponent := relPathToRepoRoot
			// Try to make it relative to the component's specific path prefix
			foundPrefix := false
			for _, prefix := range assignedComponentDef.PathPrefixes {
				if strings.HasPrefix(relPathToRepoRoot, prefix) {
					filePathInComponent = strings.TrimPrefix(relPathToRepoRoot, prefix)
					filePathInComponent = strings.TrimPrefix(filePathInComponent, "/")
					foundPrefix = true
					break
				}
			}
			if !foundPrefix && assignedComponentDef.ComponentRelPath != "." { // If not matched to a prefix but component has a path
				if strings.HasPrefix(relPathToRepoRoot, assignedComponentDef.ComponentRelPath) {
					filePathInComponent = strings.TrimPrefix(relPathToRepoRoot, assignedComponentDef.ComponentRelPath)
					filePathInComponent = strings.TrimPrefix(filePathInComponent, "/")
				}
			} else if assignedComponentDef.ComponentRelPath == "." { // For components at root (like project_other)
				filePathInComponent = relPathToRepoRoot
			}

			// log.Printf("  Processing %s (Component: %s, PathInComp: %s)", relPathToRepoRoot, assignedComponentDef.Name, filePathInComponent)
			processFile(fileSet, path, repoModulePath, filePathInComponent, currentComponentIndex)
		}
		return nil
	})

	if err != nil {
		log.Fatalf("Error walking the path %q: %v\n", repoRootPath, err)
	}

	for componentName, compIndexData := range componentIndexes {
		// Update LastIndexed to current time before writing, if desired
		compIndexData.LastIndexed = time.Now().Format(time.RFC3339)

		isPredefined := false
		var currentDef ComponentDefinition
		for _, def := range defaultComponentDefs {
			if def.Name == componentName {
				isPredefined = true
				currentDef = def
				break
			}
		}

		if !isPredefined { // Should not happen if all components are pre-initialized from defaultComponentDefs
			log.Printf("Warning: Component '%s' was not predefined. Skipping index write.", componentName)
			continue
		}

		// Skip "project_other" if it's empty and was the catch-all
		if componentName == "project_other" && len(compIndexData.Packages) == 0 && len(compIndexData.NeuroScriptTools) == 0 {
			isActuallyDefinedAsCatchAll := false
			for _, def := range defaultComponentDefs {
				if def.Name == "project_other" && len(def.PathPrefixes) == 1 && def.PathPrefixes[0] == "." {
					isActuallyDefinedAsCatchAll = true
					break
				}
			}
			if isActuallyDefinedAsCatchAll {
				log.Printf("Skipping empty catch-all component: %s", componentName)
				continue
			}
		}

		outputFileName := currentDef.IndexFile
		if outputFileName == "" { // Fallback if IndexFile wasn't set in definition
			outputFileName = strings.ToLower(componentName) + "_index.json"
		}

		componentFilePath := filepath.Join(*outputDir, outputFileName)
		jsonData, err := json.MarshalIndent(compIndexData, "", "  ")
		if err != nil {
			log.Fatalf("Error marshaling component index %s to JSON: %v", componentName, err)
		}
		err = os.WriteFile(componentFilePath, jsonData, 0644)
		if err != nil {
			log.Fatalf("Error writing component index %s to file %q: %v", componentName, componentFilePath, err)
		}
		log.Printf("Successfully wrote component index for '%s' to %s", componentName, componentFilePath)

		projectIndex.Components[componentName] = goindex.ComponentIndexFileEntry{
			Name:        componentName, // Added Name field to ComponentIndexFileEntry
			Path:        currentDef.ComponentRelPath,
			IndexFile:   outputFileName,
			Description: "", // Optional description can be added to ComponentDefinition
		}
	}

	projectIndexFilePath := filepath.Join(*outputDir, "project_index.json")
	projectJsonData, err := json.MarshalIndent(projectIndex, "", "  ")
	if err != nil {
		log.Fatalf("Error marshaling project index to JSON: %v", err)
	}
	err = os.WriteFile(projectIndexFilePath, projectJsonData, 0644)
	if err != nil {
		log.Fatalf("Error writing project index to file %q: %v", projectIndexFilePath, err)
	}
	log.Printf("Successfully wrote project index to %s", projectIndexFilePath)

	duration := time.Since(startTime)
	log.Printf("GoIndexer finished in %s. Indexed %d components.", duration, len(projectIndex.Components))
}

func assignFileToComponent(relFilePathToRepoRoot string, componentDefs []ComponentDefinition) ComponentDefinition {
	normalizedPath := filepath.ToSlash(relFilePathToRepoRoot)
	bestMatchDef := componentDefs[len(componentDefs)-1] // Assume last one is "project_other" or a default
	longestPrefixLen := 0

	for _, def := range componentDefs {
		// "project_other" with "." prefix should be a last resort.
		// If its path is just ".", it will match everything if not careful.
		isProjectOtherCatchAll := def.Name == "project_other" && len(def.PathPrefixes) == 1 && def.PathPrefixes[0] == "."

		for _, prefix := range def.PathPrefixes {
			normalizedPrefix := filepath.ToSlash(prefix)
			if normalizedPrefix == "." && !isProjectOtherCatchAll { // A specific component claims "."
				if normalizedPath != "" { // Match if file is in a subdir of this "." component.
					// This logic is tricky. Generally "." prefix means it's at the root of the component.
					// If file is "a.go", normalizedPath is "a.go".
					// If prefix is ".", and component path is also ".", it matches.
					// Let's assume specific prefixes are more concrete than ".".
				}
			}

			if normalizedPrefix != "." && strings.HasPrefix(normalizedPath, normalizedPrefix) {
				// Check if this path is more specific (longer) than current best match
				// or if the current best match is the generic "project_other"
				if len(normalizedPrefix) > longestPrefixLen {
					longestPrefixLen = len(normalizedPrefix)
					bestMatchDef = def
				}
			}
		}
	}

	// If no specific component matched (longestPrefixLen is still 0), it defaults to project_other
	// (which should be the last in componentDefs with path prefix ".")
	if longestPrefixLen == 0 {
		for _, def := range componentDefs { // Find the actual "project_other" or default
			if def.Name == "project_other" && len(def.PathPrefixes) > 0 && def.PathPrefixes[0] == "." {
				return def
			}
		}
		// Fallback to the last defined component if "project_other" with "." not found.
		if len(componentDefs) > 0 {
			return componentDefs[len(componentDefs)-1]
		}
	}

	return bestMatchDef
}

func getGitInfo(repoPath string) (branch string, commitHash string) {
	// Actual implementation would use:
	// cmdBranch := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	// cmdBranch.Dir = repoPath
	// ... and similar for commit hash ...
	return "unknown-branch", "unknown-commit"
}

func getModulePath(goModPath string) (string, error) {
	content, err := os.ReadFile(goModPath)
	if err != nil {
		return "", err
	}
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module ")), nil
		}
	}
	return "", fmt.Errorf("module declaration not found in %s", goModPath)
}
