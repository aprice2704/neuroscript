package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"path/filepath"
	"strings"
)

// List of common stdlib package prefixes to filter out from call graph
var commonStdlibPrefixes = []string{
	// Keep log.* for now, per user request (not removed)
	"fmt.", // Filter Printf/Println below
	"strings.",
	"os.",
	"errors.",
	"path/filepath.",
	"bytes.",
	"bufio.",
	"encoding/json.",
	"strconv.",
	"context.",
	"time.",
	"sync.",
	"net/http.",
	"io.",
	"sort.",
	"regexp.",
}

// Specific function names to filter
var filteredFunctionNames = map[string]bool{
	"fmt.Printf":  true,
	"fmt.Println": true,
	"log.Printf":  true, // Add log.Printf
	"log.Println": true, // Add log.Println
	// Add other specific functions if needed, e.g., "fmt.Sprintf" ?
}

// getRelativePackagePath determines the package path relative to the repo root.
func getRelativePackagePath(filePath string) (string, error) {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return "", err
	}
	dir := filepath.Dir(absPath)
	if repoRootPath == "" {
		return "", fmt.Errorf("repoRootPath not set globally")
	}
	relDir, err := filepath.Rel(repoRootPath, dir)
	if err != nil {
		return "", fmt.Errorf("could not get relative path for dir %s from root %s: %w", dir, repoRootPath, err)
	}
	if relDir == "." {
		return ".", nil
	}
	return filepath.ToSlash(relDir), nil
}

// getRelativeFilePath determines the file path relative to the repo root.
func getRelativeFilePath(filePath string) (string, error) {
	if repoRootPath == "" {
		return "", fmt.Errorf("repoRootPath not set globally")
	}
	relPath, err := filepath.Rel(repoRootPath, filePath)
	if err != nil {
		return "", fmt.Errorf("could not get relative path for file %s from root %s: %w", filePath, repoRootPath, err)
	}
	return filepath.ToSlash(relPath), nil
}

// constructShortName manually constructs the short name string.
func constructShortName(pkgPathRel, receiverName, funcName string) string {
	cleanReceiver := strings.TrimPrefix(receiverName, "*")
	var parts []string
	if pkgPathRel != "." && pkgPathRel != "" {
		parts = append(parts, pkgPathRel)
	}
	if cleanReceiver != "" {
		parts = append(parts, cleanReceiver)
	}
	parts = append(parts, funcName)
	return strings.Join(parts, ".")
}

// resolveCallTarget attempts to determine the shortName string of the called function/method.
// Returns "" if the call should be filtered.
// *** MODIFIED: Added filtering for Printf/Println, strip local package path ***
func resolveCallTarget(fset *token.FileSet, call *ast.CallExpr, currentPkgRel string, imports map[string]string) string {
	var targetName string           // Store resolved name before filtering
	var isLocalPkgCall bool = false // Track if the call is within the same repo

	switch fun := call.Fun.(type) {
	case *ast.Ident:
		// Simple identifier - function in the same package or built-in
		resolvedName := fun.Name
		if isBuiltin(resolvedName) {
			targetName = resolvedName // Builtins have no package path
		} else {
			// Call within the current package
			isLocalPkgCall = true
			targetName = constructShortName(currentPkgRel, "", resolvedName)
		}

	case *ast.SelectorExpr:
		// Selector: pkg.Func or obj.Method
		if pkgIdent, ok := fun.X.(*ast.Ident); ok {
			pkgAliasOrTypeName := pkgIdent.Name
			funcOrMethodName := fun.Sel.Name
			if fullPkgPath, found := imports[pkgAliasOrTypeName]; found {
				// Identified as pkg.Func based on import alias
				if repoModulePath != "" && strings.HasPrefix(fullPkgPath, repoModulePath) {
					// Local package call
					isLocalPkgCall = true
					relPkgPath := strings.TrimPrefix(fullPkgPath, repoModulePath)
					relPkgPath = strings.TrimPrefix(relPkgPath, "/")
					if relPkgPath == "" {
						targetName = constructShortName(".", "", funcOrMethodName)
					} else {
						targetName = constructShortName(relPkgPath, "", funcOrMethodName)
					}
				} else {
					// External package call
					targetName = fmt.Sprintf("%s.%s", fullPkgPath, funcOrMethodName)
				}
			} else {
				// Assume obj.Method or Type.Method in current package
				isLocalPkgCall = true                                                                        // Assume local if not found in imports
				targetName = constructShortName(currentPkgRel, pkgAliasOrTypeName, funcOrMethodName) + "(?)" // Mark as uncertain
			}
		} else {
			// Complex selector like a.B().C()
			targetName = formatNode(fset, fun) + "(?)"
			// Can we determine if it's local? Difficult without type info. Assume external for filtering.
			isLocalPkgCall = false
		}
	default:
		// Other complex cases
		targetName = formatNode(fset, fun) + "(?)"
		isLocalPkgCall = false // Assume external for filtering
	}

	// --- Filtering Logic ---
	if isBuiltin(targetName) {
		// log.Printf("    Filtering built-in call: %s", targetName)
		return "" // Filter builtins
	}
	if filteredFunctionNames[targetName] {
		// log.Printf("    Filtering specific function call: %s", targetName)
		return "" // Filter specific Printf/Println etc.
	}
	// Only check common prefixes for EXTERNAL packages
	if !isLocalPkgCall {
		for _, prefix := range commonStdlibPrefixes {
			if strings.HasPrefix(targetName, prefix) {
				// log.Printf("    Filtering common stdlib call: %s", targetName)
				return "" // Filter common external calls
			}
		}
	}

	// --- Strip local package prefix if required ---
	// This is the final formatting step before returning
	if isLocalPkgCall {
		// Construct the expected prefix for the current package
		currentPrefix := ""
		if currentPkgRel != "." && currentPkgRel != "" {
			currentPrefix = currentPkgRel + "."
		}
		// If the targetName starts with the current package prefix, strip it
		if currentPrefix != "" && strings.HasPrefix(targetName, currentPrefix) {
			// Check for methods - strip only the package part
			parts := strings.SplitN(targetName, ".", 3)       // Split into max 3: pkg, type/receiver, method
			if len(parts) == 3 && parts[0] == currentPkgRel { // Method call like pkg.Type.Method
				return strings.Join(parts[1:], ".") // Return Type.Method
			} else if len(parts) == 2 && parts[0] == currentPkgRel { // Function call like pkg.Function
				return parts[1] // Return Function
			}
			// Fallback or handle other cases if necessary
		} else if currentPrefix == "" && !strings.Contains(targetName, ".") {
			// If in root package (.) and target has no dots, it's already simplified
			// (Handles calls like MyFunction() within root package)
			return targetName
		}
		// If prefix doesn't match (e.g., uncertain call marked with ?), return as is
	}

	return targetName // Return the resolved, potentially simplified, and not filtered name
}

// isBuiltin checks if a name corresponds to a Go built-in function.
func isBuiltin(name string) bool {
	// Based on https://golang.org/ref/spec#Predeclared_identifiers
	switch name {
	// Functions
	case "append", "cap", "clear", "close", "complex", "copy", "delete",
		"imag", "len", "make", "max", "min", "new", "panic", "print",
		"println", "real", "recover":
		return true
	default:
		return false
	}
}
