// cmd/repac/main.go
//
// repac rewrites the leading “// filename:” header and the package clause
// of every .go file in the current directory (and, optionally, its
// sub-directories).
//
// Usage
//   repac [-recurse] <path-from-project-root-to-dot>
//
// Example (run from neuroscript/pkg/tool):
//   repac -recurse pkg/tool
//
// A file located at pkg/tool/errtools/tools_error.go becomes:
//
//   // filename: pkg/tool/errtools/tools_error.go
//   package errtools
//
// The tool skips “.git/” and “vendor/” trees and preserves gofmt spacing.

package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var recurse = flag.Bool("recurse", false, "recurse into sub-directories")

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(),
			"Usage: %s [-recurse] <path-from-project-root-to-dot>\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(2)
	}

	rootToDot := filepath.Clean(flag.Arg(0)) // e.g. pkg/tool
	startDir, err := os.Getwd()              // absolute path to "."
	must(err)

	var goFiles []string
	if *recurse {
		err = filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				switch d.Name() {
				case ".git", "vendor":
					return filepath.SkipDir
				}
				return nil
			}
			if strings.HasSuffix(path, ".go") {
				goFiles = append(goFiles, path)
			}
			return nil
		})
		must(err)
	} else {
		ents, err := os.ReadDir(".")
		must(err)
		for _, e := range ents {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".go") {
				goFiles = append(goFiles, e.Name())
			}
		}
	}

	for _, rel := range goFiles {
		abs, err := filepath.Abs(rel) // absolute path for stable processing
		must(err)
		processFile(abs, startDir, rootToDot)
	}
}

// processFile rewrites one file’s header and package clause.
func processFile(absPath, startDir, rootToDot string) {
	fset := token.NewFileSet()
	astFile, err := parser.ParseFile(fset, absPath, nil, parser.ParseComments)
	if err != nil {
		log.Printf("skip %s (cannot parse): %v", absPath, err)
		return
	}

	// derive package name from the file’s own directory
	dirName := filepath.Base(filepath.Dir(absPath))

	// For test files, append "_test" to the package name
	if strings.HasSuffix(absPath, "_test.go") {
		dirName += "_test"
	}

	if astFile.Name.Name != dirName {
		astFile.Name = &ast.Ident{Name: dirName}
	}

	// re-print the AST back to source (preserves formatting)
	var buf bytes.Buffer
	must(printer.Fprint(&buf, fset, astFile))

	// fix or insert the // filename: header
	newSrc := fixHeader(buf.Bytes(), absPath, startDir, rootToDot)

	// write back, preserving file permissions
	info, err := os.Stat(absPath)
	must(err)
	must(os.WriteFile(absPath, newSrc, info.Mode()))
	fmt.Printf("updated %s\n", absPath)
}

// ───────────────────────────────────────────────────────────────────────────────
// fixHeader ensures the first header line reads:
//
//   // filename: <rootToDot>/<relPathFromDotToFile>
//
// It removes any legacy “// Filename:” header.

func fixHeader(src []byte, absPath, startDir, rootToDot string) []byte {
	relBelowStart, _ := filepath.Rel(startDir, absPath)
	headerPath := filepath.ToSlash(filepath.Join(rootToDot, relBelowStart))
	want := "// filename: " + headerPath

	sc := bufio.NewScanner(bytes.NewReader(src))
	sc.Split(bufio.ScanLines)

	var outLines []string
	inserted := false
	legacyHdr := regexp.MustCompile(`^//\s*Filename:`) // strip these
	curHdr := regexp.MustCompile(`^//\s*filename:`)    // replace these

	for sc.Scan() {
		line := sc.Text()

		// drop legacy header entirely
		if legacyHdr.MatchString(line) {
			continue
		}

		// replace existing lowercase header
		if !inserted && curHdr.MatchString(line) {
			line = want
			inserted = true
		}

		outLines = append(outLines, line)

		// insert header before first non-comment if needed
		if !inserted && strings.TrimSpace(line) == "" {
			// allow blank line before header
		} else if !inserted && !strings.HasPrefix(line, "//") {
			outLines = append([]string{want}, outLines...)
			inserted = true
		}
	}
	if !inserted { // empty file edge-case
		outLines = append([]string{want}, outLines...)
	}
	return []byte(strings.Join(outLines, "\n"))
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
