package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4" // Import doublestar for matching
)

// --- Data Structures (Assume defined, same as previous version) ---
type Index struct {
	SchemaVersion string                 `json:"schemaVersion"`
	RepoRoot      string                 `json:"repoRoot"`
	Packages      map[string]PackageInfo `json:"packages"`
}
type PackageInfo struct {
	Files map[string]FileInfo `json:"files"`
}
type FileInfo struct {
	FunctionShortNames []string     `json:"functions,omitempty"`
	Methods            []MethodInfo `json:"methods"`
}
type MethodInfo struct {
	Receiver string
	Name     string
	Calls    []string `json:"calls,omitempty"`
}

// CallInfo struct was removed

// --- Global Variables / Config (remains the same) ---
var (
	repoRootPath   string
	repoModulePath string
)

// --- Custom Flag Type (remains the same) ---
type CommaSeparatedFlag []string

func (f *CommaSeparatedFlag) String() string { return strings.Join(*f, ",") }
func (f *CommaSeparatedFlag) Set(value string) error {
	if value == "" {
		return fmt.Errorf("empty value for comma-separated flag")
	}
	parts := strings.Split(value, ",")
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			*f = append(*f, trimmed)
		}
	}
	return nil
}

func main() {
	log.SetFlags(0)

	// --- Command Line Flags ---
	var dirPaths CommaSeparatedFlag
	var excludePaths CommaSeparatedFlag // For exclusion glob patterns
	flag.Var(&dirPaths, "dirs", "Comma-separated list of directories to scan")
	// *** UPDATED HELP TEXT for -exclude ***
	flag.Var(&excludePaths, "exclude", "Comma-separated doublestar glob patterns relative to repo root to exclude (e.g., '**/*_test.go', 'pkg/core/generated/**')")
	outputFile := flag.String("o", "neuroscript_index.json", "Output JSON file path")
	repoRootOverride := flag.String("repo-root", "", "Override detected repository root path")
	repoModuleOverride := flag.String("repo-module", "", "Override detected repository module path")

	// Custom Usage message to include the example
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(flag.CommandLine.Output(), "\nExample:\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  %s -dirs . -exclude 'pkg/core/generated/**,**/*_test.go' -o index.json\n", os.Args[0])
	}

	flag.Parse()

	if len(dirPaths) == 0 {
		log.Fatal("Error: -dirs flag is required")
	}
	if *outputFile == "" {
		log.Fatal("Error: -o flag is required")
	}

	// --- Detect repo paths (logic remains same) ---
	var detectErr error
	startDir := dirPaths[0]
	absStartDir, err := filepath.Abs(startDir)
	if err != nil {
		log.Fatalf("Error getting absolute path for %q: %v", startDir, err)
	}
	repoRootPath, repoModulePath, detectErr = findRepoPaths(absStartDir) // Assumes defined in finder.go
	rootFound := repoRootPath != ""
	moduleParsed := repoModulePath != ""
	if detectErr != nil {
		log.Printf("Warning: Problem during repo path detection: %v", detectErr)
	}
	if *repoRootOverride != "" {
		repoRootPath = *repoRootOverride
		rootFound = true
		log.Printf("Using overridden Repo Root Path: %s", repoRootPath)
	}
	if *repoModuleOverride != "" {
		repoModulePath = *repoModuleOverride
		moduleParsed = true
		log.Printf("Using overridden Repo Module Path: %s", repoModulePath)
	}
	if !rootFound {
		log.Fatal("Error: Repository root path could not be determined.")
	}
	if !moduleParsed {
		if detectErr != nil {
			log.Fatalf("Error: Failed to determine module path. %v", detectErr)
		} else {
			log.Fatal("Error: Repository module path could not be determined.")
		}
	}
	absRepoRootPath, err := filepath.Abs(repoRootPath)
	if err != nil {
		log.Fatalf("Error getting absolute path for final repo root %q: %v", repoRootPath, err)
	}
	repoRootPath = filepath.Clean(absRepoRootPath)
	log.Printf("Final Repo Root Path: %s", repoRootPath)
	log.Printf("Final Repo Module Path: %s", repoModulePath)

	// Store exclude patterns directly (convert to forward slashes for consistency)
	forwardSlashExcludePaths := make([]string, len(excludePaths))
	for i, pattern := range excludePaths {
		trimmedPattern := strings.TrimPrefix(pattern, "./")
		forwardSlashExcludePaths[i] = filepath.ToSlash(trimmedPattern)
	}
	log.Printf("Using exclude patterns: %v", forwardSlashExcludePaths)

	// --- Initialize Index ---
	index := Index{
		SchemaVersion: "1.7", // Version remains same as only logging/help changed
		RepoRoot:      repoModulePath,
		Packages:      make(map[string]PackageInfo),
	}

	// --- Process Directories ---
	fileSet := token.NewFileSet()

	for _, dir := range dirPaths {
		absDir, err := filepath.Abs(dir)
		if err != nil {
			log.Printf("Error getting abs path for %q: %v. Skipping.", dir, err)
			continue
		}
		if !strings.HasPrefix(absDir, repoRootPath) {
			log.Printf("Warning: Dir %q (%s) outside root %s. Skipping.", dir, absDir, repoRootPath)
			continue
		}
		log.Printf("Scanning directory: %s", absDir)

		// Use filepath.Walk, apply doublestar.PathMatch inside
		walkFunc := func(path string, info os.FileInfo, errWalk error) error {
			if errWalk != nil {
				log.Printf("Error accessing path %q: %v", path, errWalk)
				return nil
			}

			// Get path relative to root for exclusion checks
			relPath, err := filepath.Rel(repoRootPath, path)
			if err != nil {
				log.Printf("Warning: Cannot get relative path for %q: %v", path, err)
				return nil
			} // Skip if rel path fails
			relPath = filepath.ToSlash(relPath) // Use forward slashes

			// --- Exclusion Check using doublestar.PathMatch ---
			for _, pattern := range forwardSlashExcludePaths {
				// PathMatch expects forward slashes in both pattern and path
				match, errMatch := doublestar.PathMatch(pattern, relPath)
				if errMatch != nil {
					log.Printf("Warning: Invalid exclude pattern '%s': %v", pattern, errMatch)
					continue // Skip invalid pattern
				}
				if match {
					// *** UNCOMMENTED LOGGING ***
					log.Printf("  Excluding '%s' matching pattern '%s'", relPath, pattern)
					if info.IsDir() {
						return filepath.SkipDir // Skip the whole directory if it matches
					}
					return nil // Skip this file
				}
			}

			// Standard skips (after exclusion check)
			if info.IsDir() {
				if name := info.Name(); name == "vendor" || name == ".git" {
					// log.Printf("  Skipping common directory: %s", path)
					return filepath.SkipDir
				}
				return nil // Continue into directory
			}
			if !strings.HasSuffix(info.Name(), ".go") {
				return nil
			} // Only process .go files

			// Process the file if not excluded
			processFile(fileSet, path, &index) // Call helper (defined in parser.go)
			return nil
		}

		err = filepath.Walk(absDir, walkFunc) // Use filepath.Walk
		if err != nil {
			log.Printf("Error walking directory %q: %v", dir, err)
		}
	}

	// --- Write Output JSON ---
	jsonData, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		log.Fatalf("Error marshaling index to JSON: %v", err)
	}
	err = os.WriteFile(*outputFile, jsonData, 0644)
	if err != nil {
		log.Fatalf("Error writing index to file %q: %v", *outputFile, err)
	}
	log.Printf("Successfully wrote index to %s", *outputFile)
}
