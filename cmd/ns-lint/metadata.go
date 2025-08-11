// NeuroScript Version: 0.3.0
// File version: 1
// Purpose: Parse ::metadata headers from .ns files for ns-lint, collect .ns file list from paths.
// filename: cmd/ns-lint/metadata.go
// nlines: 104
// risk_rating: LOW

package main

import (
	"bufio"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// FileMeta is a map of metadata keys (lowercased) to values.
type FileMeta map[string]string

// FuncMeta holds metadata associated with a single function or block.
type FuncMeta struct {
	Name string
	Meta map[string]string
}

// ParseMetadata reads the leading metadata lines (::key: value) from a file
// until the first non-metadata line, returning a FileMeta map and function-level
// metadata slice (currently empty; reserved for future function-level parsing).
func ParseMetadata(path string) (FileMeta, []FuncMeta, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()

	meta := FileMeta{}
	var funcs []FuncMeta

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		s := strings.TrimSpace(sc.Text())
		if strings.HasPrefix(s, "::") {
			kv := strings.TrimPrefix(s, "::")
			i := strings.IndexByte(kv, ':')
			if i > 0 {
				k := strings.ToLower(strings.TrimSpace(kv[:i]))
				v := strings.TrimSpace(kv[i+1:])
				meta[k] = v
			}
			continue
		}
		// Stop when we hit first non-metadata line
		break
	}
	if err := sc.Err(); err != nil {
		return nil, nil, err
	}
	return meta, funcs, nil
}

// collectNS recursively collects all .ns files from the given paths.
func collectNS(paths []string) ([]string, error) {
	seen := make(map[string]struct{})
	var files []string

	add := func(p string) {
		if _, ok := seen[p]; !ok {
			seen[p] = struct{}{}
			files = append(files, p)
		}
	}

	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			return nil, err
		}
		if info.IsDir() {
			err = filepath.WalkDir(p, func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if d.IsDir() {
					return nil
				}
				if strings.HasSuffix(strings.ToLower(d.Name()), ".ns") {
					add(path)
				}
				return nil
			})
			if err != nil {
				return nil, err
			}
		} else if strings.HasSuffix(strings.ToLower(p), ".ns") {
			add(p)
		}
	}

	return files, nil
}
