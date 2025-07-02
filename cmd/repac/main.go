// cmd/repac/main.go
//
// Usage:
//   repac [-recurse] <path-from-project-root-to-dot>
//
// Example (run inside neuroscript/pkg/tool/fs):
//   repac -recurse pkg/tool/fs
//
// ──> every .go file under . is rewritten to
//     package fs
//     // filename: pkg/tool/fs/<subpath>/<file.go>

// filename: cmd/repac
package repac

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

var (
	recurse = flag.Bool("recurse", false, "recurse into sub-directories")
)

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

	rootToDot := filepath.Clean(flag.Arg(0))
	pkgName := filepath.Base(rootToDot)

	// absolute path of the starting directory (“.”)
	startDir, err := os.Getwd()
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

	for _, file := range goFiles {
		processFile(file, startDir, rootToDot, pkgName)
	}
}

// ────────────────────────────────────────────────────────────────────────────────

func processFile(path, startDir, rootToDot, pkgName string) {
	fset := token.NewFileSet()
	astFile, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		log.Printf("skip %s (cannot parse): %v", path, err)
		return
	}

	// 1️⃣ package name
	if astFile.Name.Name != pkgName {
		astFile.Name = &ast.Ident{Name: pkgName}
	}

	// 2️⃣ re-print source (keeps formatting)
	var buf bytes.Buffer
	must(printer.Fprint(&buf, fset, astFile))

	// 3️⃣ adjust // filename: comment
	newSrc := fixHeader(buf.Bytes(), path, startDir, rootToDot)

	// 4️⃣ write back preserving perms
	info, err := os.Stat(path)
	must(err)
	must(os.WriteFile(path, newSrc, info.Mode()))
	fmt.Printf("updated %s\n", path)
}

// rewrite/insert the “// filename:” line
func fixHeader(src []byte, absPath, startDir, rootToDot string) []byte {
	relBelowStart, _ := filepath.Rel(startDir, absPath)
	headerPath := filepath.ToSlash(filepath.Join(rootToDot, relBelowStart))
	want := "// filename: " + headerPath

	sc := bufio.NewScanner(bytes.NewReader(src))
	sc.Split(bufio.ScanLines)

	var outLines []string
	inserted := false
	re := regexp.MustCompile(`^//\s*filename:`)

	for sc.Scan() {
		line := sc.Text()
		if !inserted {
			if re.MatchString(line) {
				line = want	// replace
				inserted = true
			} else if strings.TrimSpace(line) == "" {
				// keep empty line before header
			} else if !strings.HasPrefix(line, "//") {
				// first real code line & no header seen – insert now
				outLines = append(outLines, want)
				inserted = true
			}
		}
		outLines = append(outLines, line)
	}
	if !inserted {	// file had no lines (unlikely) or ended before header inserted
		outLines = append([]string{want}, outLines...)
	}
	return []byte(strings.Join(outLines, "\n"))
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}