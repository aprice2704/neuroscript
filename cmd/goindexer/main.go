// NeuroScript Go Indexer
// File version: 2.3.1 // Correct compiler nits, re-add helper functions.
// Purpose: Main package for the Go code indexer. Defines data structures and orchestrates parsing.
// filename: cmd/goindexer/main.go
// nlines: 305 // Approximate
// risk_rating: MEDIUM
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"go/token"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/aprice2704/neuroscript/pkg/adapters"
	"github.com/aprice2704/neuroscript/pkg/core"
	"github.com/aprice2704/neuroscript/pkg/goindex"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	// You might need to import your toolsets package if there's a central registration function like:
	// "github.com/aprice2704/neuroscript/pkg/toolsets"
)

var (
	repoRootPath   string
	repoModulePath string
)

// ComponentDefinition defines a logical component of the project for indexing.
type ComponentDefinition struct {
	Name             string
	PathPrefixes     []string
	IndexFile        string
	ComponentRelPath string
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
	{Name: "commands", PathPrefixes: []string{"cmd/ns", "cmd/ns-lsp", "cmd/goindexer", "cmd/nspatch", "cmd/nsinput"}, IndexFile: "commands_index.json", ComponentRelPath: "cmd"},
	{Name: "project_other", PathPrefixes: []string{"."}, IndexFile: "project_other_index.json", ComponentRelPath: "."},
}

const defaultComponentName = "project_other"

// mapCoreArgSpecsToNeuroScriptArgDetails converts core.ArgSpec to the index's NeuroScriptArgDetail.
// TODO: Move this helper to pkg/goindex for sharing between goindexer and pkg/goindex/reader.
func mapCoreArgSpecsToNeuroScriptArgDetails(coreArgs []core.ArgSpec) []goindex.NeuroScriptArgDetail {
	if coreArgs == nil {
		return nil
	}
	nsArgs := make([]goindex.NeuroScriptArgDetail, len(coreArgs))
	for i, ca := range coreArgs {
		nsArgs[i] = goindex.NeuroScriptArgDetail{
			Name:         ca.Name,
			Type:         string(ca.Type),
			Description:  ca.Description,
			Required:     ca.Required,
			DefaultValue: ca.DefaultValue,
		}
	}
	return nsArgs
}

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
	gitBranch, gitCommitHash := getGitInfo(repoRootPath) // Re-added call

	projectIndex := goindex.ProjectIndex{
		ProjectRootModulePath: repoModulePath,
		IndexSchemaVersion:    "project_index_v2.0.0",
		Components:            make(map[string]goindex.ComponentIndexFileEntry),
		LastIndexedTimestamp:  time.Now().Format(time.RFC3339),
		GitBranch:             gitBranch,
		GitCommitHash:         gitCommitHash,
	}

	for _, def := range defaultComponentDefs {
		componentIndexes[def.Name] = &goindex.ComponentIndex{
			ComponentName:      def.Name,
			ComponentPath:      def.ComponentRelPath,
			IndexSchemaVersion: "component_index_v2.0.0",
			Packages:           make(map[string]*goindex.PackageDetail),
			NeuroScriptTools:   make([]goindex.NeuroScriptToolDetail, 0),
			LastIndexed:        projectIndex.LastIndexedTimestamp,
			GitBranch:          gitBranch,
			GitCommitHash:      gitCommitHash,
		}
	}

	log.Println("Starting to scan and parse Go files...")
	err = filepath.Walk(repoRootPath, func(path string, info os.FileInfo, errWalk error) error {
		if errWalk != nil {
			log.Printf("Prevented directory walk from failing: %v (Path: %s)", errWalk, path)
			return nil
		}

		if info.IsDir() {
			dirName := info.Name()
			if dirName == "vendor" || dirName == ".git" || (strings.HasPrefix(dirName, ".") && dirName != "." && dirName != "..") {
				return filepath.SkipDir
			}
			return nil
		}

		if strings.HasSuffix(info.Name(), ".go") && !strings.HasSuffix(info.Name(), "_test.go") {
			relPathToRepoRoot, errPath := filepath.Rel(repoRootPath, path)
			if errPath != nil {
				log.Printf("    Error getting relative path for %q: %v", path, errPath)
				return nil
			}
			relPathToRepoRoot = filepath.ToSlash(relPathToRepoRoot)

			assignedComponentDef := assignFileToComponent(relPathToRepoRoot, defaultComponentDefs) // Re-added call
			currentComponentIndex := componentIndexes[assignedComponentDef.Name]

			filePathInComponent := relPathToRepoRoot
			foundPrefix := false
			for _, prefix := range assignedComponentDef.PathPrefixes {
				if strings.HasPrefix(relPathToRepoRoot, prefix) {
					filePathInComponent = strings.TrimPrefix(relPathToRepoRoot, prefix)
					filePathInComponent = strings.TrimPrefix(filePathInComponent, "/")
					foundPrefix = true
					break
				}
			}
			if !foundPrefix && assignedComponentDef.ComponentRelPath != "." {
				if strings.HasPrefix(relPathToRepoRoot, assignedComponentDef.ComponentRelPath) {
					filePathInComponent = strings.TrimPrefix(relPathToRepoRoot, assignedComponentDef.ComponentRelPath)
					filePathInComponent = strings.TrimPrefix(filePathInComponent, "/")
				}
			} else if assignedComponentDef.ComponentRelPath == "." {
				filePathInComponent = relPathToRepoRoot
			}
			processFile(fileSet, path, repoRootPath, repoModulePath, filePathInComponent, currentComponentIndex)
		}
		return nil
	})
	if err != nil {
		log.Fatalf("Error walking the path %q: %v\n", repoRootPath, err)
	}
	log.Println("Go file parsing complete.")

	// --- Generate and Link NeuroScript Tool Details (In-Memory) ---
	log.Println("Generating and linking NeuroScript tool details...")

	var indexerLogger interfaces.Logger
	indexerLogger = adapters.NewNoOpLogger() // Use NoOp for indexer's internal interpreter

	interpreterSandboxDir, err := os.MkdirTemp("", "goindexer_interpreter_")
	if err != nil {
		log.Fatalf("Failed to create temp dir for interpreter: %v", err)
	}
	defer os.RemoveAll(interpreterSandboxDir)

	llmClientForIndexer := adapters.NewNoOpLLMClient()

	interpreter, err := core.NewInterpreter(indexerLogger, llmClientForIndexer, interpreterSandboxDir, nil, nil)
	if err != nil {
		log.Fatalf("Failed to create core.Interpreter for tool indexing: %v", err)
	}

	// Register all known toolsets with the interpreter for the indexer
	if errRegCore := core.RegisterCoreTools(interpreter.ToolRegistry()); errRegCore != nil {
		log.Printf("Warning: Failed to register core tools for goindexer: %v", errRegCore)
	}
	if errRegAI := core.RegisterAIWorkerTools(interpreter); errRegAI != nil {
		log.Printf("Warning: Failed to register AI Worker tools for goindexer: %v", errRegAI)
	}
	// Example if you have a pkg/toolsets:
	// if errRegAll := toolsets.RegisterAllToolsets(interpreter.ToolRegistry()); errRegAll != nil {
	//    log.Printf("Warning: Failed to register custom toolsets for goindexer: %v", errRegAll)
	// }

	allToolSpecs := interpreter.ListTools()
	log.Printf("Found %d tools in interpreter for detail generation.", len(allToolSpecs))

	componentPathToNameMap := make(map[string]string)
	for _, def := range defaultComponentDefs {
		componentPathToNameMap[filepath.ToSlash(def.ComponentRelPath)] = def.Name
	}

	processedToolCount := 0
	for _, spec := range allToolSpecs {
		toolImpl, exists := interpreter.GetTool(spec.Name)
		if !exists || toolImpl.Func == nil {
			log.Printf("Warning (goindexer): ToolSpec '%s' found but implementation missing or Func is nil. Creating partial detail.", spec.Name)
			partialDetail := goindex.NeuroScriptToolDetail{
				Name: spec.Name, Description: spec.Description, Category: spec.Category,
				Args:       mapCoreArgSpecsToNeuroScriptArgDetails(spec.Args),
				ReturnType: string(spec.ReturnType), ReturnHelp: spec.ReturnHelp,
				Variadic: spec.Variadic, Example: spec.Example, ErrorConditions: spec.ErrorConditions,
				ImplementingGoFunctionFullName: "N/A - Implementation not found by goindexer",
			}
			targetCompNameForUnlinked := "core" // Default for unlinked
			if _, ok := componentIndexes[targetCompNameForUnlinked]; !ok && len(defaultComponentDefs) > 0 {
				targetCompNameForUnlinked = defaultComponentDefs[0].Name // Fallback to first defined
			}
			if compIdxTarget, ok := componentIndexes[targetCompNameForUnlinked]; ok {
				compIdxTarget.NeuroScriptTools = append(compIdxTarget.NeuroScriptTools, partialDetail)
			}
			continue
		}

		runtimeGoFuncName := runtime.FuncForPC(reflect.ValueOf(toolImpl.Func).Pointer()).Name()
		var foundItem *goindex.FoundCallableItem
		var foundInComponentName string
		var itemErr error

		searchNameForMethod := runtimeGoFuncName
		if strings.HasSuffix(runtimeGoFuncName, "-fm") {
			searchNameForMethod = strings.TrimSuffix(runtimeGoFuncName, "-fm")
		}

	itemSearchLoop:
		for compNameIter, compIndex := range componentIndexes {
			for _, pkgDetail := range compIndex.Packages {
				for i := range pkgDetail.Functions {
					if pkgDetail.Functions[i].Name == runtimeGoFuncName {
						f := pkgDetail.Functions[i]
						foundItem = &goindex.FoundCallableItem{
							FQN: f.Name, SourceFile: f.SourceFile, Returns: f.Returns, IsExported: f.IsExported,
						}
						foundInComponentName = compNameIter
						break itemSearchLoop
					}
				}
				for i := range pkgDetail.Methods {
					if pkgDetail.Methods[i].FQN == searchNameForMethod {
						m := pkgDetail.Methods[i]
						foundItem = &goindex.FoundCallableItem{
							FQN: m.FQN, SourceFile: m.SourceFile, Returns: m.Returns, IsExported: m.IsExported,
						}
						foundInComponentName = compNameIter
						break itemSearchLoop
					}
				}
			}
		}

		toolDetail := goindex.NeuroScriptToolDetail{
			Name: spec.Name, Description: spec.Description, Category: spec.Category,
			Args: mapCoreArgSpecsToNeuroScriptArgDetails(spec.Args), ReturnType: string(spec.ReturnType),
			ReturnHelp: spec.ReturnHelp, Variadic: spec.Variadic, Example: spec.Example, ErrorConditions: spec.ErrorConditions,
			ImplementingGoFunctionFullName: runtimeGoFuncName,
		}

		if foundItem != nil {
			toolDetail.ImplementingGoFunctionFullName = foundItem.FQN
			toolDetail.ImplementingGoFunctionSourceFile = foundItem.SourceFile
			for _, retType := range foundItem.Returns {
				if retType == "error" {
					toolDetail.GoImplementingFunctionReturnsError = true
					break
				}
			}
			actualComponentRelPath := ""
			for _, def := range defaultComponentDefs {
				if def.Name == foundInComponentName {
					actualComponentRelPath = def.ComponentRelPath
					break
				}
			}
			toolDetail.ComponentPath = actualComponentRelPath
			log.Printf("  Linked tool '%s' to Go func '%s' in component '%s' (path: '%s', file: '%s')",
				spec.Name, foundItem.FQN, foundInComponentName, actualComponentRelPath, foundItem.SourceFile)
		} else {
			itemErr = fmt.Errorf("callable '%s' (method search '%s') not found in in-memory parsed index", runtimeGoFuncName, searchNameForMethod)
			log.Printf("  Warning (goindexer): Could not find Go function for tool '%s' (runtime FQN: %s): %v. Partial detail created.", spec.Name, runtimeGoFuncName, itemErr)

			// Fallback component assignment
			targetCompNameForUnlinked := "core"
			compIdx, compExists := componentIndexes[targetCompNameForUnlinked] // compIdx was declared but not used here.
			if !compExists && len(defaultComponentDefs) > 0 {
				targetCompNameForUnlinked = defaultComponentDefs[0].Name
				compIdx, compExists = componentIndexes[targetCompNameForUnlinked] // Re-assign compIdx and compExists
			}
			_ = compIdx
			if compExists { // Use the (potentially reassigned) compIdx
				foundInComponentName = targetCompNameForUnlinked // So it gets added to this component
			} else {
				log.Printf("    Cannot assign unlinked tool '%s' to any component.", spec.Name)
				continue
			}
		}

		if targetComponent, ok := componentIndexes[foundInComponentName]; ok {
			targetComponent.NeuroScriptTools = append(targetComponent.NeuroScriptTools, toolDetail)
			processedToolCount++
		} else {
			log.Printf("Error: Component '%s' (for tool '%s') not found in componentIndexes. Tool detail not added.", foundInComponentName, spec.Name)
		}
	}
	log.Printf("Finished processing %d tool details.", processedToolCount)

	// --- Write Final Enriched Indexes ---
	log.Println("Writing final component and project indexes...")
	projectIndex.LastIndexedTimestamp = time.Now().Format(time.RFC3339)

	for componentNameWrite, compIndexDataWrite := range componentIndexes {
		compIndexDataWrite.LastIndexed = projectIndex.LastIndexedTimestamp
		var currentDefWrite ComponentDefinition
		foundDef := false
		for _, def := range defaultComponentDefs {
			if def.Name == componentNameWrite {
				currentDefWrite = def
				foundDef = true
				break
			}
		}
		if !foundDef {
			log.Printf("Warning: Component '%s' not predefined. Skipping.", componentNameWrite)
			continue
		}
		if componentNameWrite == defaultComponentName && len(compIndexDataWrite.Packages) == 0 && len(compIndexDataWrite.NeuroScriptTools) == 0 {
			log.Printf("Skipping empty catch-all component: %s", componentNameWrite)
			continue
		}
		outputFileName := currentDefWrite.IndexFile
		componentFilePath := filepath.Join(*outputDir, outputFileName)
		jsonData, err := json.MarshalIndent(compIndexDataWrite, "", "  ")
		if err != nil {
			log.Fatalf("Error marshaling final component index %s: %v", componentNameWrite, err)
		}
		if err = os.WriteFile(componentFilePath, jsonData, 0644); err != nil {
			log.Fatalf("Error writing final component index %s to %s: %v", componentNameWrite, componentFilePath, err)
		}
		log.Printf("Successfully wrote final component index for '%s' to %s", componentNameWrite, componentFilePath)
		projectIndex.Components[componentNameWrite] = goindex.ComponentIndexFileEntry{
			Name: componentNameWrite, Path: currentDefWrite.ComponentRelPath, IndexFile: outputFileName, Description: "",
		}
	}

	projectIndexFilePath := filepath.Join(*outputDir, "project_index.json")
	projectJsonData, err := json.MarshalIndent(projectIndex, "", "  ")
	if err != nil {
		log.Fatalf("Error marshaling final project index: %v", err)
	}
	if err = os.WriteFile(projectIndexFilePath, projectJsonData, 0644); err != nil {
		log.Fatalf("Error writing final project index to %s: %v", projectIndexFilePath, err)
	}
	log.Printf("Successfully wrote final project index to %s", projectIndexFilePath)

	duration := time.Since(startTime)
	log.Printf("GoIndexer finished in %s. Indexed %d components.", duration, len(projectIndex.Components))
}

// assignFileToComponent determines which component a file belongs to.
func assignFileToComponent(relFilePathToRepoRoot string, componentDefs []ComponentDefinition) ComponentDefinition {
	normalizedPath := filepath.ToSlash(relFilePathToRepoRoot)
	bestMatchDef := componentDefs[len(componentDefs)-1]
	longestPrefixLen := 0

	for _, def := range componentDefs {
		isProjectOtherCatchAll := def.Name == defaultComponentName && len(def.PathPrefixes) == 1 && def.PathPrefixes[0] == "."
		for _, prefix := range def.PathPrefixes {
			normalizedPrefix := filepath.ToSlash(prefix)
			if normalizedPrefix == "." && isProjectOtherCatchAll && longestPrefixLen > 0 { // True catch-all only if no other prefix matched
				continue
			}
			if normalizedPrefix == "." && !isProjectOtherCatchAll { // Specific component claims "." (root files)
				if !strings.Contains(normalizedPath, "/") { // File is directly in root
					if len(normalizedPrefix) > longestPrefixLen { // Should be 1 for "."
						longestPrefixLen = len(normalizedPrefix)
						bestMatchDef = def
					}
				}
				continue
			}
			if normalizedPrefix != "." && strings.HasPrefix(normalizedPath, normalizedPrefix+"/") || normalizedPath == normalizedPrefix {
				if len(normalizedPrefix) > longestPrefixLen {
					longestPrefixLen = len(normalizedPrefix)
					bestMatchDef = def
				}
			}
		}
	}
	return bestMatchDef
}

// getGitInfo retrieves the current git branch and commit hash.
// Placeholder implementation.
func getGitInfo(repoPath string) (branch string, commitHash string) {
	cmdBranch := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmdBranch.Dir = repoPath
	branchOut, errBranch := cmdBranch.Output()
	if errBranch == nil {
		branch = strings.TrimSpace(string(branchOut))
	} else {
		branch = "unknown-branch"
		log.Printf("Warning: could not get git branch: %v", errBranch)
	}

	cmdCommit := exec.Command("git", "rev-parse", "--short", "HEAD")
	cmdCommit.Dir = repoPath
	commitOut, errCommit := cmdCommit.Output()
	if errCommit == nil {
		commitHash = strings.TrimSpace(string(commitOut))
	} else {
		commitHash = "unknown-commit"
		log.Printf("Warning: could not get git commit hash: %v", errCommit)
	}
	return branch, commitHash
}

// getModulePath (simplified, assuming it's available if needed by findRepoPaths or parser)
// findRepoPaths would typically call a more robust version like the one in your finder.go
