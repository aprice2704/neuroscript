package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

const (
	colorGreen = "\033[32m"
	colorReset = "\033[0m"
)

// goplsRe parses the standard `gopls check` output.
var goplsRe = regexp.MustCompile(`^([^:]+):(\d+):\d+-\d+:\s+undefined:\s+(.*)$`)

// --- Data Structures for the Symbol Index ---

type PackageIndex struct {
	Index CategorizedSymbolIndex `json:"index"`
}
type CategorizedSymbolIndex struct {
	Constants  map[string]string `json:"constants"`
	Variables  map[string]string `json:"variables"`
	Types      map[string]string `json:"types"`
	Functions  map[string]string `json:"functions"`
	Interfaces map[string]string `json:"interfaces"`
}

// VetError is a simplified, common format for an error.
type VetError struct {
	File       string
	LineNumber int
	Symbol     string
}

// main orchestrates the entire process.
func main() {
	dryRun := flag.Bool("dry", false, "Display suggestions with color instead of writing to files.")
	flag.Parse()

	if len(flag.Args()) < 1 {
		log.Fatalln("Usage: go run . [-dry] <path_to_index.json>")
	}
	indexPath := flag.Args()[0]

	symbolToFileMap, err := readAndFlattenSymbolIndex(indexPath)
	if err != nil {
		log.Fatalf("Error reading symbol index file '%s': %v", indexPath, err)
	}
	log.Printf("Successfully loaded and indexed %d unique symbols.", len(symbolToFileMap))

	goFiles, err := findGoFiles(".")
	if err != nil {
		log.Fatalf("Failed to find Go files: %v", err)
	}
	log.Printf("Found %d Go files to check.", len(goFiles))
	fmt.Println()

	totalErrorsFound := 0

	// ### MAIN PROCESSING LOOP ###
	// Process each file completely before moving to the next.
	for i, file := range goFiles {
		log.Printf("Processing file (%d/%d): %s", i+1, len(goFiles), file)

		// Run gopls check on the single file
		cmd := exec.Command("gopls", "check", file)
		output, _ := cmd.CombinedOutput()
		errorsForThisFile := parseGoplsErrors(output)

		if len(errorsForThisFile) == 0 {
			continue // No errors, move to the next file
		}

		totalErrorsFound += len(errorsForThisFile)
		fmt.Printf("--- Found %d issue(s) in %s ---\n", len(errorsForThisFile), file)

		if err := processFile(file, errorsForThisFile, symbolToFileMap, *dryRun); err != nil {
			log.Printf("ERROR: Could not process file %s: %v", file, err)
		}
		fmt.Println() // Blank line for readability
	}

	log.Printf("Processing complete. Found a total of %d undefined symbols.", totalErrorsFound)
	if *dryRun {
		log.Println("Dry run finished. No files were modified.")
	} else {
		log.Println("Fixes applied.")
	}
}

// processFile handles all suggestions or modifications for a single file.
func processFile(file string, errors []VetError, symbolMap map[string]string, dryRun bool) error {
	content, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	lines := strings.Split(string(content), "\n")

	currentPkg, err := getFilePackageName(file)
	if err != nil {
		return err
	}

	if dryRun {
		// In dry-run mode, just print suggestions for the file's errors.
		for _, vetErr := range errors {
			printSuggestion(lines, vetErr, symbolMap, currentPkg)
		}
		return nil
	}

	// In write mode, apply all fixes and then save the file.
	modifiedLines := make([]string, len(lines))
	copy(modifiedLines, lines)
	modificationsMade := 0

	for _, vetErr := range errors {
		defFile, ok := symbolMap[vetErr.Symbol]
		if !ok {
			continue
		}
		symbolPkg := getPackageFromPath(defFile)

		if currentPkg != symbolPkg {
			lineIndex := vetErr.LineNumber - 1
			originalLine := modifiedLines[lineIndex]
			fixedLine := applyFix(originalLine, vetErr.Symbol, symbolPkg, "")
			if fixedLine != originalLine {
				modifiedLines[lineIndex] = fixedLine
				modificationsMade++
			}
		}
	}

	if modificationsMade > 0 {
		output := strings.Join(modifiedLines, "\n")
		if err := os.WriteFile(file, []byte(output), 0644); err != nil {
			return fmt.Errorf("failed to write changes: %w", err)
		}
		log.Printf("Modified %s to fix %d undefined symbols.", file, modificationsMade)

		cmd := exec.Command("gopls", "imports", "-w", file)
		if err := cmd.Run(); err != nil {
			log.Printf("Warning: 'gopls imports -w %s' failed: %v", file, err)
		} else {
			log.Printf("Updated imports for %s.", file)
		}
	}

	return nil
}

// printSuggestion calculates and prints a single colored suggestion for dry-run mode.
func printSuggestion(lines []string, vetErr VetError, symbolMap map[string]string, currentPkg string) {
	defFile, ok := symbolMap[vetErr.Symbol]
	if !ok {
		return
	}
	symbolPkg := getPackageFromPath(defFile)

	if currentPkg != symbolPkg && vetErr.LineNumber > 0 && vetErr.LineNumber <= len(lines) {
		originalLine := lines[vetErr.LineNumber-1]
		displayLine := applyFix(originalLine, vetErr.Symbol, symbolPkg, colorGreen)

		if displayLine != originalLine {
			fmt.Printf("L%d: Missing import for symbol '%s' (package: %s)\n", vetErr.LineNumber, vetErr.Symbol, symbolPkg)
			fmt.Printf("  Original: %s\n", strings.TrimSpace(originalLine))
			fmt.Printf("  Proposed: %s\n", strings.TrimSpace(displayLine))
		}
	}
}

// applyFix intelligently replaces a symbol on a line, skipping struct fields.
func applyFix(line, symbol, pkg, color string) string {
	re := regexp.MustCompile(`\b` + regexp.QuoteMeta(symbol) + `\b`)
	matches := re.FindAllStringIndex(line, -1)

	for i := len(matches) - 1; i >= 0; i-- {
		match := matches[i]
		start, end := match[0], match[1]

		if end < len(line) && line[end] == ':' {
			continue
		}

		var replacement string
		if color != "" {
			replacement = color + pkg + colorReset + "." + symbol
		} else {
			replacement = pkg + "." + symbol
		}
		line = line[:start] + replacement + line[end:]
	}
	return line
}

// findGoFiles recursively finds all Go files from a root directory.
func findGoFiles(root string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".go") {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

// parseGoplsErrors parses the text output from a `gopls check` command.
func parseGoplsErrors(goplsOutput []byte) []VetError {
	var errors []VetError
	for _, line := range strings.Split(string(goplsOutput), "\n") {
		matches := goplsRe.FindStringSubmatch(line)
		if len(matches) == 4 {
			absPath := matches[1]
			relPath, err := filepath.Rel(".", absPath)
			if err != nil {
				relPath = absPath
			}
			lineNum, _ := strconv.Atoi(matches[2])
			errors = append(errors, VetError{
				File:       relPath,
				LineNumber: lineNum,
				Symbol:     strings.TrimSpace(matches[3]),
			})
		}
	}
	return errors
}

// readAndFlattenSymbolIndex reads the pkg_idx.json file and creates a unified map of symbols.
func readAndFlattenSymbolIndex(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("could not open file: %w", err)
	}
	defer file.Close()

	var pkgIndex PackageIndex
	if err := json.NewDecoder(file).Decode(&pkgIndex); err != nil {
		return nil, fmt.Errorf("could not decode JSON: %w", err)
	}

	flatIndex := make(map[string]string)
	maps := []map[string]string{
		pkgIndex.Index.Constants, pkgIndex.Index.Variables, pkgIndex.Index.Types,
		pkgIndex.Index.Functions, pkgIndex.Index.Interfaces,
	}
	for _, m := range maps {
		for symbol, file := range m {
			flatIndex[symbol] = file
		}
	}
	return flatIndex, nil
}

// getFilePackageName reads a Go file to find its package declaration.
func getFilePackageName(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	re := regexp.MustCompile(`^package\s+([a-zA-Z0-9_]+)`)
	for scanner.Scan() {
		line := scanner.Text()
		if matches := re.FindStringSubmatch(line); len(matches) == 2 {
			return matches[1], nil
		}
	}
	return "", fmt.Errorf("package declaration not found in %s", filePath)
}

// getPackageFromPath extracts the package name from a file path.
func getPackageFromPath(filePath string) string {
	return filepath.Base(filepath.Dir(filePath))
}
