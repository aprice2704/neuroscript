package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// Package holds the information we care about from the `go list -json` command.
type Package struct {
	ImportPath string   `json:"ImportPath"`
	Imports    []string `json:"Imports"`
	Dir        string   `json:"Dir"` // We'll use this to find the module root
}

// buildDepGraph analyzes a list of packages and constructs a dependency graph.
// The graph is a map where the key is a package's import path and the value is a
// slice of its internal (same-module) dependencies.
func buildDepGraph(packages []string) (map[string][]string, error) {
	graph := make(map[string][]string)
	pkgSet := make(map[string]struct{})
	for _, p := range packages {
		pkgSet[p] = struct{}{}
	}

	// Use `go list -json` to get details for all packages at once.
	args := []string{"list", "-json"}
	args = append(args, packages...)
	cmd := exec.Command("go", args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("could not run 'go list': %v\n%s", err, stderr.String())
	}

	// The JSON output is a stream of JSON objects, not a single array.
	// We need to decode them one by one.
	decoder := json.NewDecoder(&stdout)
	for decoder.More() {
		var pkg Package
		if err := decoder.Decode(&pkg); err != nil {
			return nil, fmt.Errorf("error decoding package json: %w", err)
		}

		var internalImports []string
		for _, imp := range pkg.Imports {
			// An internal import is one that is also in our set of packages.
			if _, ok := pkgSet[imp]; ok {
				internalImports = append(internalImports, imp)
			}
		}
		// Sort for consistent output
		sort.Strings(internalImports)
		graph[pkg.ImportPath] = internalImports
	}

	return graph, nil
}

// findRoots identifies the "root" packages in the graph - those that are not
// imported by any other package within the module. This is only used for the 'tree' format.
func findRoots(graph map[string][]string) []string {
	// A set to keep track of all packages that are dependencies.
	imported := make(map[string]struct{})
	for _, deps := range graph {
		for _, dep := range deps {
			imported[dep] = struct{}{}
		}
	}

	var roots []string
	for pkg := range graph {
		if _, ok := imported[pkg]; !ok {
			roots = append(roots, pkg)
		}
	}
	// Sort for consistent output
	sort.Strings(roots)
	return roots
}

// trimPath shortens the full import path to be relative to the user's CWD.
func trimPath(fullPath, modulePath, relCwd string) string {
	// 1. Trim the module path prefix from the full import path.
	// e.g., "github.com/a/b/c/d" with module "github.com/a/b" -> "c/d"
	trimmed := strings.TrimPrefix(fullPath, modulePath)
	trimmed = strings.TrimPrefix(trimmed, "/")

	// 2. Trim the relative path from module root to CWD.
	// e.g., "c/d" with relCwd "c" -> "d"
	if relCwd != "." && relCwd != "" {
		// Convert OS-specific separator to forward slash for path matching
		relCwdSlash := filepath.ToSlash(relCwd)
		trimmed = strings.TrimPrefix(trimmed, relCwdSlash)
		trimmed = strings.TrimPrefix(trimmed, "/")
	}

	// If after all trimming, the path is empty, it means we're looking
	// at the package for the current directory. Let's represent it with "."
	if trimmed == "" {
		return "."
	}

	return trimmed
}

// printSlim prints the dependency graph as a simple, token-efficient list.
func printSlim(graph map[string][]string, modulePath, relCwd string) {
	var relations []string
	for pkg, deps := range graph {
		// Only include packages that have dependencies in the output
		if len(deps) == 0 {
			continue
		}
		parentPath := trimPath(pkg, modulePath, relCwd)
		for _, dep := range deps {
			childPath := trimPath(dep, modulePath, relCwd)
			relations = append(relations, fmt.Sprintf("%s -> %s", parentPath, childPath))
		}
	}
	// Sort the final list for deterministic output
	sort.Strings(relations)
	for _, relation := range relations {
		fmt.Println(relation)
	}
}

// printTree recursively prints the dependency tree for a given package. (for 'tree' format)
func printTree(pkg, prefix string, visited map[string]bool, graph map[string][]string, modulePath, relCwd string) {
	// Check for circular dependencies.
	if visited[pkg] {
		displayPath := trimPath(pkg, modulePath, relCwd)
		fmt.Printf("%s└─ %s ... (cycle detected)\n", prefix, displayPath)
		return
	}
	visited[pkg] = true

	deps := graph[pkg]
	for i, dep := range deps {
		connector := "├─"
		newPrefix := prefix + "│  "
		if i == len(deps)-1 {
			connector = "└─"
			newPrefix = prefix + "   "
		}
		displayPath := trimPath(dep, modulePath, relCwd)
		fmt.Printf("%s%s %s\n", prefix, connector, displayPath)
		printTree(dep, newPrefix, visited, graph, modulePath, relCwd)
	}

	// Backtrack: remove from visited map for the current path.
	delete(visited, pkg)
}

func main() {
	// --- Setup and Argument Parsing ---
	log.SetFlags(0)
	repoPath := flag.String("path", ".", "Path to the Go repository/module root.")
	format := flag.String("format", "slim", "Output format: 'slim' or 'tree'.")
	flag.Parse()

	absPath, err := filepath.Abs(*repoPath)
	if err != nil {
		log.Fatalf("Error getting absolute path for %s: %v", *repoPath, err)
	}

	// --- Step 1: Find module, packages, and relevant paths ---
	cmdModule := exec.Command("go", "list", "-m")
	cmdModule.Dir = absPath
	var moduleOut, cmdErr bytes.Buffer
	cmdModule.Stdout = &moduleOut
	cmdModule.Stderr = &cmdErr
	if err := cmdModule.Run(); err != nil {
		log.Fatalf("Error finding module path in %s. Are you in a Go module?: %v\n%s", absPath, err, cmdErr.String())
	}
	modulePath := strings.TrimSpace(moduleOut.String())

	cmdModDir := exec.Command("go", "list", "-m", "-f", "{{.Dir}}")
	cmdModDir.Dir = absPath
	var modDirOut bytes.Buffer
	cmdModDir.Stdout = &modDirOut
	cmdModDir.Stderr = &cmdErr
	if err := cmdModDir.Run(); err != nil {
		log.Fatalf("Error finding module root directory: %v\n%s", err, cmdErr.String())
	}
	moduleRootDir := strings.TrimSpace(modDirOut.String())

	relCwd, err := filepath.Rel(moduleRootDir, absPath)
	if err != nil {
		log.Printf("Warning: could not determine relative path from module root: %v. Paths will not be shortened.", err)
		relCwd = "."
	}

	cmdPkgs := exec.Command("go", "list", "./...")
	cmdPkgs.Dir = absPath
	var pkgsOut bytes.Buffer
	cmdPkgs.Stdout = &pkgsOut
	cmdPkgs.Stderr = &cmdErr
	if err := cmdPkgs.Run(); err != nil {
		log.Fatalf("Error listing packages in %s: %v\n%s", absPath, err, cmdErr.String())
	}
	packages := strings.Split(strings.TrimSpace(pkgsOut.String()), "\n")

	if len(packages) == 0 || (len(packages) == 1 && packages[0] == "") {
		log.Fatalf("No Go packages found in %s", absPath)
	}

	// --- Step 2: Build the dependency graph ---
	graph, err := buildDepGraph(packages)
	if err != nil {
		log.Fatalf("Error building dependency graph: %v", err)
	}

	// --- Step 3: Print the graph based on the selected format ---
	switch *format {
	case "slim":
		printSlim(graph, modulePath, relCwd)
	case "tree":
		fmt.Printf("Analyzing dependencies for module: %s\n(Paths relative to: %s)\n\n", modulePath, absPath)
		roots := findRoots(graph)
		if len(roots) == 0 {
			fmt.Println("No clear root packages found. Listing all packages and their dependencies:")
			for pkg := range graph {
				roots = append(roots, pkg)
			}
			sort.Strings(roots)
		}
		for _, root := range roots {
			displayPath := trimPath(root, modulePath, relCwd)
			fmt.Println(displayPath)
			printTree(root, "", make(map[string]bool), graph, modulePath, relCwd)
			fmt.Println()
		}
	default:
		log.Fatalf("Invalid format: %q. Please use 'slim' or 'tree'.", *format)
	}
}
