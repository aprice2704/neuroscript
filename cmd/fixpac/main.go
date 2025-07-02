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

	allErrors, err := checkAllFilesOneByOne()
	if err != nil {
		log.Fatalf("Failed during file checking: %v", err)
	}

	if len(allErrors) == 0 {
		fmt.Println("\nNo 'undefined' symbol errors found in any files.")
		return
	}
	fmt.Printf("\nFound %d 'undefined' symbol error(s) across all files.\n\n", len(allErrors))

	if err := suggestAndFix(allErrors, symbolToFileMap, *dryRun); err != nil {
		log.Fatalf("Error suggesting or applying fixes: %v", err)
	}

	if *dryRun {
		fmt.Println("Dry run complete. No files were modified.")
	} else {
		fmt.Println("Fixes applied successfully.")
	}
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

// checkAllFilesOneByOne finds all Go files and runs `gopls check` on each one.
func checkAllFilesOneByOne() ([]VetError, error) {
	goFiles, err := findGoFiles(".")
	if err != nil {
		return nil, fmt.Errorf("error finding go files: %w", err)
	}

	log.Printf("Found %d Go files to check.", len(goFiles))
	var allErrors []VetError
	re := regexp.MustCompile(`^([^:]+):(\d+):\d+-\d+:\s+undefined:\s+(.*)$`)

	for i, file := range goFiles {
		log.Printf("Checking file (%d/%d): %s", i+1, len(goFiles), file)
		cmd := exec.Command("gopls", "check", file)
		output, _ := cmd.CombinedOutput()

		for _, line := range strings.Split(string(output), "\n") {
			matches := re.FindStringSubmatch(line)
			if len(matches) == 4 {
				absPath := matches[1]
				relPath, err := filepath.Rel(".", absPath)
				if err != nil {
					relPath = absPath
				}

				lineNum, _ := strconv.Atoi(matches[2])
				allErrors = append(allErrors, VetError{
					File:       relPath,
					LineNumber: lineNum,
					Symbol:     strings.TrimSpace(matches[3]),
				})
			}
		}
	}
	return allErrors, nil
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

// suggestAndFix iterates through errors, then either prints suggestions or applies fixes.
func suggestAndFix(errors []VetError, symbolToFileMap map[string]string, dryRun bool) error {
	errorsByFile := make(map[string][]VetError)
	for _, err := range errors {
		errorsByFile[err.File] = append(errorsByFile[err.File], err)
	}

	for file, errs := range errorsByFile {
		fmt.Printf("--- Processing: %s ---\n", file)

		currentPkg, err := getFilePackageName(file)
		if err != nil {
			log.Printf("Warning: Could not determine package for %s: %v", file, err)
			continue
		}

		fileContent, err := os.ReadFile(file)
		if err != nil {
			log.Printf("Warning: Could not read file %s: %v", file, err)
			continue
		}
		lines := strings.Split(string(fileContent), "\n")
		modifiedLines := make([]string, len(lines))
		copy(modifiedLines, lines)
		modificationsMade := 0

		for _, vetErr := range errs {
			defFile, ok := symbolToFileMap[vetErr.Symbol]
			if !ok {
				log.Printf("Warning: Symbol '%s' from L%d not found in the index.", vetErr.Symbol, vetErr.LineNumber)
				continue
			}
			symbolPkg := getPackageFromPath(defFile)

			if currentPkg != symbolPkg && vetErr.LineNumber > 0 && vetErr.LineNumber <= len(lines) {
				originalLine := modifiedLines[vetErr.LineNumber-1] // Use potentially modified line for subsequent fixes
				re := regexp.MustCompile(`\b` + regexp.QuoteMeta(vetErr.Symbol) + `\b`)

				if dryRun {
					coloredReplacement := colorGreen + symbolPkg + colorReset + "." + vetErr.Symbol
					displayLine := re.ReplaceAllString(originalLine, coloredReplacement)
					fmt.Printf("L%d: Missing import for symbol '%s' (package: %s)\n", vetErr.LineNumber, vetErr.Symbol, symbolPkg)
					fmt.Printf("  Original: %s\n", strings.TrimSpace(originalLine))
					fmt.Printf("  Proposed: %s\n", strings.TrimSpace(displayLine))
				} else {
					fixedLine := re.ReplaceAllString(originalLine, symbolPkg+"."+vetErr.Symbol)
					modifiedLines[vetErr.LineNumber-1] = fixedLine
					modificationsMade++
				}
			}
		}

		if !dryRun && modificationsMade > 0 {
			output := strings.Join(modifiedLines, "\n")
			if err := os.WriteFile(file, []byte(output), 0644); err != nil {
				log.Printf("ERROR: Failed to write changes to %s: %v", file, err)
				continue
			}
			log.Printf("Modified %s to fix %d undefined symbols.", file, modificationsMade)

			cmd := exec.Command("gopls", "imports", "-w", file)
			if err := cmd.Run(); err != nil {
				log.Printf("Warning: 'gopls imports -w %s' failed: %v", file, err)
			} else {
				log.Printf("Updated imports for %s.", file)
			}
		}
		fmt.Println() // Add a blank line for readability
	}
	return nil
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
