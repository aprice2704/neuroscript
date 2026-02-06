// :: product: FDM/NS
// :: majorVersion: 1
// :: fileVersion: 5
// :: description: CLI utility to audit and enforce the filename metadata key, ignoring generated files.
// :: latestChange: Added exclusion for files containing "DO NOT EDIT" and verified traversal of empty subdirectories.
// :: filename: code/cmd/meta-audit/main.go
// :: serialization: go

package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/metadata"
)

var (
	dryRun  = flag.Bool("dry-run", false, "Report inconsistencies without modifying files")
	apply   = flag.Bool("yes", false, "Apply changes to files")
	verbose = flag.Bool("v", false, "Verbose output")
)

func main() {
	flag.Parse()

	if !*dryRun && !*apply {
		fmt.Println("Error: specify --dry-run to check or --yes to apply changes.")
		os.Exit(1)
	}

	root, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get working directory: %v\n", err)
		os.Exit(1)
	}

	auditor := &Auditor{
		Root:    root,
		DryRun:  *dryRun,
		Verbose: *verbose,
	}

	// filepath.WalkDir will descend into all directories including those
	// that are currently empty of target files.
	err = filepath.WalkDir(root, auditor.Walk)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Walk failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nAudit complete. Files checked: %d, Modified: %d\n", auditor.Checked, auditor.Modified)
}

type Auditor struct {
	Root     string
	DryRun   bool
	Verbose  bool
	Checked  int
	Modified int
}

func (a *Auditor) Walk(path string, d fs.DirEntry, err error) error {
	if err != nil {
		return err
	}

	// Ignore hidden directories and .git
	if d.IsDir() {
		if strings.HasPrefix(d.Name(), ".") {
			return filepath.SkipDir
		}
		// Return nil to continue descending into this directory
		return nil
	}

	// Target specified extensions
	ext := filepath.Ext(path)
	name := d.Name()
	isTarget := false
	switch ext {
	case ".go", ".ns", ".md":
		isTarget = true
	case ".txt":
		if strings.HasSuffix(name, ".ns.txt") {
			isTarget = true
		}
	}
	if name == "ndcl.md" {
		isTarget = true
	}

	if !isTarget {
		return nil
	}

	a.Checked++
	return a.AuditFile(path)
}

func (a *Auditor) AuditFile(path string) error {
	relPath, err := filepath.Rel(a.Root, path)
	if err != nil {
		return err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// Never edit a file that says “DO NOT EDIT”
	checkLimit := 1024
	if len(data) < checkLimit {
		checkLimit = len(data)
	}
	if bytes.Contains(data[:checkLimit], []byte("DO NOT EDIT")) {
		if a.Verbose {
			fmt.Printf("[SKIP] %s (Generated/Do Not Edit)\n", relPath)
		}
		return nil
	}

	// Read lines to handle CRLF/LF correctly
	lines, err := metadata.ReadLines(bytes.NewReader(data))
	if err != nil {
		return err
	}

	// Use unified parsing logic to find existing blocks
	store, metaEndLine := metadata.ParseHeaderBlock(lines)
	content := data
	serialization := ""
	isHeader := true

	if store == nil {
		var metaStartLine int
		store, metaStartLine = metadata.ParseFooterBlock(lines)
		if store != nil {
			isHeader = false
			serialization = "md"
			content = []byte(strings.Join(lines[:metaStartLine], "\n"))
		}
	} else {
		isHeader = true
		serialization = "ns"
		if filepath.Ext(path) == ".go" {
			serialization = "go"
		}
		content = []byte(strings.Join(lines[metaEndLine:], "\n"))
	}

	// Fallback for injection
	if store == nil {
		store = make(metadata.Store)
		ext := filepath.Ext(path)
		switch ext {
		case ".go":
			serialization = "go"
		case ".ns":
			serialization = "ns"
		default:
			serialization = "md"
		}
	}

	// Audit only the filename key
	if store["filename"] == relPath {
		return nil
	}

	a.Modified++
	if a.DryRun {
		fmt.Printf("[DRY-RUN] %s: Filename missing/wrong. Want: %q\n", relPath, relPath)
		return nil
	}

	fmt.Printf("[FIX] %s\n", relPath)
	store["filename"] = relPath

	return a.WriteUpdatedFile(path, store, content, serialization, isHeader)
}

func (a *Auditor) WriteUpdatedFile(path string, store metadata.Store, content []byte, serialization string, isHeader bool) error {
	var buf bytes.Buffer
	filename := store["filename"]
	trimmedContent := bytes.TrimSpace(content)

	switch serialization {
	case "go":
		// Top-of-file doc comment format
		fmt.Fprintf(&buf, "// :: filename: %s\n", filename)
		for k, v := range store {
			if k != "filename" {
				fmt.Fprintf(&buf, "// :: %s: %s\n", k, v)
			}
		}
		buf.WriteString("\n")
		buf.Write(trimmedContent)
		buf.WriteString("\n")

	case "ns":
		// Native metadata header
		fmt.Fprintf(&buf, ":: filename: %s\n", filename)
		for k, v := range store {
			if k != "filename" {
				fmt.Fprintf(&buf, ":: %s: %s\n", k, v)
			}
		}
		// Followed by exactly one blank line
		buf.WriteString("\n")
		buf.Write(trimmedContent)
		buf.WriteString("\n")

	case "md":
		// Native metadata footer
		buf.Write(trimmedContent)
		// Preceded by exactly one blank line
		buf.WriteString("\n\n")
		fmt.Fprintf(&buf, ":: filename: %s\n", filename)
		for k, v := range store {
			if k != "filename" {
				fmt.Fprintf(&buf, ":: %s: %s\n", k, v)
			}
		}
	}

	return os.WriteFile(path, buf.Bytes(), 0644)
}
