// Fixpac fixes some broken package refs

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
	colorGreen  = "\033[32m"
	colorOrange = "\033[33m"
	colorGrey   = "\033[90m"
	colorReset  = "\033[0m"
)

var goplsRe = regexp.MustCompile(`^([^:]+):(\d+):\d+-\d+:\s+undefined:\s+(.*)$`)

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
type VetError struct {
	File       string
	LineNumber int
	Symbol     string
}

func main() {
	dryRun := flag.Bool("dry", false, "Display suggestions instead of writing to files.")
	rememberCleared := flag.Bool("remcleared", false, "Remember successfully cleared files by writing them to cleared.json.")
	ignoreCleared := flag.Bool("igcleared", false, "Ignore files listed in a root cleared.json on startup.")
	flag.Parse()

	if len(flag.Args()) < 1 {
		log.Fatalln("Usage: go run . [-dry] [-remcleared] [-igcleared] <path_to_index.json>")
	}
	indexPath := flag.Args()[0]

	// This list will contain all files to skip for this session.
	skipList := make(map[string]bool)
	if *ignoreCleared {
		cleared, err := readClearedFiles("./cleared.json")
		if err != nil {
			log.Printf("Warning: could not read root cleared.json, proceeding without skipping files. Error: %v", err)
		} else {
			for _, file := range cleared {
				skipList[file] = true
			}
			log.Printf("Ignoring %d file(s) from root cleared.json.", len(skipList))
		}
	}

	symbolToFileMap, err := readAndFlattenSymbolIndex(indexPath)
	if err != nil {
		log.Fatalf("Error reading symbol index file '%s': %v", indexPath, err)
	}
	log.Printf("Successfully loaded and indexed %d unique symbols.", len(symbolToFileMap))

	// Find Go files and dynamically populate the skip list from any nested cleared.json files.
	goFiles, err := findFilesAndPopulateSkipList(".", skipList)
	if err != nil {
		log.Fatalf("Failed to find Go files: %v", err)
	}
	log.Printf("Found %d Go files to check.", len(goFiles))
	fmt.Println()

	totalErrorsFound := 0
	var sessionClearedFiles []string // Only files cleared in this run get added here.

	for i, file := range goFiles {
		if skipList[file] {
			log.Printf("Skipping file (%d/%d): %s%s%s", i+1, len(goFiles), colorGrey, file, colorReset)
			continue
		}

		log.Printf("Processing file (%d/%d): %s", i+1, len(goFiles), file)
		cmd := exec.Command("gopls", "check", file)
		output, _ := cmd.CombinedOutput()
		errorsForThisFile := parseGoplsErrors(output)

		if len(errorsForThisFile) == 0 {
			log.Printf("%sFile clean: %s%s", colorGreen, file, colorReset)
			sessionClearedFiles = append(sessionClearedFiles, file)
			fmt.Println()
			continue
		}

		totalErrorsFound += len(errorsForThisFile)
		fmt.Printf("--- Found %d issue(s) in %s ---\n", len(errorsForThisFile), file)
		if err := processFile(file, errorsForThisFile, symbolToFileMap, *dryRun); err != nil {
			log.Printf("ERROR: Could not process file %s: %v", file, err)
		}
		fmt.Println()
	}

	log.Printf("Processing complete. Found a total of %d undefined symbols.", totalErrorsFound)
	if *rememberCleared {
		if err := writeClearedFiles(sessionClearedFiles); err != nil {
			log.Fatalf("ERROR: Could not write to cleared.json: %v", err)
		}
		log.Printf("Wrote %d clean filenames to ./cleared.json.", len(sessionClearedFiles))
	}
	if *dryRun {
		log.Println("Dry run finished. No files were modified.")
	} else {
		log.Println("Fixes applied.")
	}
}

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
		for _, vetErr := range errors {
			printSuggestion(lines, vetErr, symbolMap, currentPkg)
		}
		return nil
	}

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
		} else {
			log.Printf("L%d: Symbol '%s' is in the same package ('%s'). Skipping prefix.", vetErr.LineNumber, vetErr.Symbol, currentPkg)
		}
	}

	if modificationsMade > 0 {
		output := strings.Join(modifiedLines, "\n")
		if err := os.WriteFile(file, []byte(output), 0644); err != nil {
			return fmt.Errorf("failed to write changes: %w", err)
		}
		log.Printf("Modified %s to fix %s%d%s symbols.", file, colorOrange, modificationsMade, colorReset)
	}

	log.Println("Running 'gopls imports' to clean up...")
	cmd := exec.Command("gopls", "imports", "-w", file)
	if err := cmd.Run(); err != nil {
		log.Printf("Warning: 'gopls imports -w %s' failed: %v", file, err)
	} else {
		log.Printf("Formatted imports for %s.", file)
	}
	return nil
}

func printSuggestion(lines []string, vetErr VetError, symbolMap map[string]string, currentPkg string) {
	defFile, ok := symbolMap[vetErr.Symbol]
	if !ok {
		return
	}
	symbolPkg := getPackageFromPath(defFile)
	lineIndex := vetErr.LineNumber - 1
	if lineIndex < 0 || lineIndex >= len(lines) {
		return
	}
	originalLine := lines[lineIndex]
	if currentPkg != symbolPkg {
		displayLine := applyFix(originalLine, vetErr.Symbol, symbolPkg, colorGreen)
		if displayLine != originalLine {
			fmt.Printf("L%d: Missing import for symbol '%s' (package: %s)\n", vetErr.LineNumber, vetErr.Symbol, symbolPkg)
			fmt.Printf("  Original: %s\n", strings.TrimSpace(originalLine))
			fmt.Printf("  Proposed: %s\n", strings.TrimSpace(displayLine))
		}
	} else {
		log.Printf("L%d: Symbol '%s' is in the same package ('%s'). No fix needed.", vetErr.LineNumber, vetErr.Symbol, currentPkg)
	}
}

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

func findFilesAndPopulateSkipList(root string, skipList map[string]bool) ([]string, error) {
	var goFiles []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		// If we find a cleared.json, read it and add its contents to the skip list.
		if filepath.Base(path) == "cleared.json" {
			// Don't process the root one if it was already handled by the -igcleared flag.
			if isRoot, _ := filepath.Match(".", filepath.Dir(path)); isRoot && skipList[path] {
				return nil
			}

			subCleared, readErr := readClearedFiles(path)
			if readErr != nil {
				log.Printf("Warning: could not read sub-directory %s: %v", path, readErr)
				return nil // Continue walking
			}
			parentDir := filepath.Dir(path)
			for _, subFile := range subCleared {
				fullPathToSkip := filepath.Join(parentDir, subFile)
				if !skipList[fullPathToSkip] {
					log.Printf("Ignoring '%s' due to local cleared.json", fullPathToSkip)
					skipList[fullPathToSkip] = true
				}
			}
		} else if strings.HasSuffix(path, ".go") {
			goFiles = append(goFiles, path)
		}
		return nil
	})
	return goFiles, err
}

func parseGoplsErrors(goplsOutput []byte) []VetError {
	var errors []VetError
	for _, line := range strings.Split(string(goplsOutput), "\n") {
		matches := goplsRe.FindStringSubmatch(line)
		if len(matches) == 4 {
			absPath, _ := filepath.Abs(matches[1])
			relPath, err := filepath.Rel(".", absPath)
			if err != nil {
				relPath = absPath
			}
			lineNum, _ := strconv.Atoi(matches[2])
			errors = append(errors, VetError{File: relPath, LineNumber: lineNum, Symbol: strings.TrimSpace(matches[3])})
		}
	}
	return errors
}

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
	maps := []map[string]string{pkgIndex.Index.Constants, pkgIndex.Index.Variables, pkgIndex.Index.Types, pkgIndex.Index.Functions, pkgIndex.Index.Interfaces}
	for _, m := range maps {
		for symbol, file := range m {
			flatIndex[symbol] = file
		}
	}
	return flatIndex, nil
}

func getFilePackageName(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	re := regexp.MustCompile(`^package\s+([a-zA-Z0-9_]+)`)
	for scanner.Scan() {
		if matches := re.FindStringSubmatch(scanner.Text()); len(matches) == 2 {
			return matches[1], nil
		}
	}
	return "", fmt.Errorf("package declaration not found in %s", filePath)
}

func getPackageFromPath(filePath string) string {
	return filepath.Base(filepath.Dir(filePath))
}

func readClearedFiles(path string) ([]string, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return []string{}, nil
	}
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cleared []string
	if err := json.Unmarshal(bytes, &cleared); err != nil {
		return nil, err
	}
	return cleared, nil
}

func writeClearedFiles(cleared []string) error {
	bytes, err := json.MarshalIndent(cleared, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile("./cleared.json", bytes, 0644)
}
