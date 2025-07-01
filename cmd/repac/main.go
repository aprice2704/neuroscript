package repac

import (
	"bytes"
	"fmt"
	"go/parser"
	"go/printer"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	// 1. Get the current working directory.
	currentDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current directory: %v", err)
	}

	// 2. Extract the base name of the directory for the new package name.
	newPackageName := filepath.Base(currentDir)
	fmt.Printf("Target package name will be: %s\n", newPackageName)

	// 3. Read all entries in the current directory.
	files, err := os.ReadDir(".")
	if err != nil {
		log.Fatalf("Failed to read directory: %v", err)
	}

	// 4. Loop through each file/directory.
	for _, file := range files {
		// Skip directories and non-Go files.
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".go") {
			continue
		}

		filePath := file.Name()
		fmt.Printf("Processing file: %s\n", filePath)

		// --- AST-based modification ---

		// a. Create a FileSet. This is needed by the parser to track source lang.Positions.
		fset := token.NewFileSet()

		// b. Parse the file to build an Abstract Syntax Tree (AST).
		// We use ParseComments so that comments are preserved when we write the file back.
		node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
		if err != nil {
			log.Printf("WARNING: Could not parse %s: %v", filePath, err)
			continue
		}

		// c. Check if the package name is already correct.
		if node.Name.Name == newPackageName {
			fmt.Printf("Package in %s is already correct. Skipping.\n", filePath)
			continue
		}

		fmt.Printf("Changing package from '%s' to '%s' in %s\n", node.Name.Name, newPackageName, filePath)

		// d. Modify the package name directly in the AST.
		// This is the core of the change.
		node.Name.Name = newPackageName

		// e. Create a buffer to write the modified AST back to.
		var buf bytes.Buffer

		// f. Use the printer to format the AST and write it to the buffer.
		// This reconstructs the Go source code from the modified tree.
		if err := printer.Fprint(&buf, fset, node); err != nil {
			log.Printf("WARNING: Failed to format modified code for %s: %v", filePath, err)
			continue
		}

		// g. Write the buffer's content back to the original file.
		info, err := file.Info()
		if err != nil {
			log.Printf("WARNING: Could not get file info for %s: %v", filePath, err)
			continue
		}
		if err := os.WriteFile(filePath, buf.Bytes(), info.Mode()); err != nil {
			log.Printf("WARNING: Failed to write updated file %s: %v", filePath, err)
			continue
		}

		fmt.Printf("Successfully updated package in %s\n", filePath)
	}

	fmt.Println("\nDone.")
}
