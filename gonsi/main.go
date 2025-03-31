// gonsi/main.go
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time" // Added for token timing if needed, optional

	"github.com/antlr4-go/antlr/v4"
	// Use your actual module path here if different
	"github.com/aprice2704/neuroscript/pkg/core"
	gen "github.com/aprice2704/neuroscript/pkg/core/generated" // Import generated package for token names
)

func main() {
	// Updated arguments: main.go <skills_directory> <ProcedureToRun> [args...]
	if len(os.Args) < 3 {
		fmt.Println("Usage: gonsi <skills_directory> <ProcedureToRun> [args...]")
		os.Exit(1)
	}
	skillsDir := os.Args[1]
	procToRun := os.Args[2]
	procArgs := []string{} // Initialize as empty string slice
	if len(os.Args) > 3 {
		procArgs = os.Args[3:]
	}

	// --- Load Procedures ---
	interpreter := core.NewInterpreter()
	fmt.Printf("Loading procedures from directory: %s\n", skillsDir)
	var firstParseError error // Variable to store the first error encountered

	// Walk the directory to find .ns.txt files
	walkErr := filepath.Walk(skillsDir, func(path string, info os.FileInfo, err error) error {
		// Stop immediately if a parse error has already occurred in a previous file
		if firstParseError != nil {
			return firstParseError // Propagate the error to stop the walk
		}

		if err != nil {
			fmt.Printf("    Warning: Error accessing path %q: %v\n", path, err)
			return nil // Continue walking if possible on access error
		}

		if !info.IsDir() && strings.HasSuffix(info.Name(), ".ns.txt") {
			fmt.Printf("\n--- Processing File: %s ---\n", filepath.Base(path)) // Identify file being processed

			// Read file content first
			contentBytes, readErr := os.ReadFile(path)
			if readErr != nil {
				fmt.Printf("    [Error] Could not read file %s: %v\n", filepath.Base(path), readErr)
				return nil // Continue walking
			}
			inputString := string(contentBytes)
			inputStream := antlr.NewInputStream(inputString)

			// Create Lexer
			lexer := gen.NewNeuroScriptLexer(inputStream)
			lexerErrorListener := core.NewDiagnosticErrorListener() // Use custom listener if available, otherwise basic
			lexer.RemoveErrorListeners()
			lexer.AddErrorListener(lexerErrorListener)

			// Create Token Stream
			tokenStream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
			tokenReadStart := time.Now() // Optional timing
			tokenStream.Fill()           // Fill the stream to get all tokens
			tokenReadDuration := time.Since(tokenReadStart)

			// --- Print Token Stream ---
			fmt.Printf("    Tokens (Read in %v):\n", tokenReadDuration)
			tokens := tokenStream.GetAllTokens()
			tokenNames := lexer.GetSymbolicNames() // Use lexer for names
			for _, token := range tokens {
				tokenName := ""
				tokenType := token.GetTokenType()
				if tokenType >= 0 && tokenType < len(tokenNames) {
					tokenName = tokenNames[tokenType]
				} else if tokenType == antlr.TokenEOF {
					tokenName = "EOF"
				} else {
					tokenName = fmt.Sprintf("Type(%d)", tokenType)
				}
				// Print details: Type Name (Text) Line:Col Channel
				fmt.Printf("      %s (%q) L%d:%d Ch:%d\n",
					tokenName,
					token.GetText(),
					token.GetLine(),
					token.GetColumn(),
					token.GetChannel(),
				)
			}
			fmt.Println("    --- End Tokens ---")
			// --- End Print Token Stream ---

			// Check for lexer errors before parsing
			if lexerErrorListener.HasErrors() {
				errStr := lexerErrorListener.GetErrors()
				fmt.Printf("    [Warning] Skipping file due to lexer errors:\n%s\n", errStr)
				firstParseError = fmt.Errorf("lexer error in '%s'", filepath.Base(path)) // Set flag to stop
				return firstParseError                                                   // Stop walk
			}

			// IMPORTANT: Reset the token stream to the beginning for the parser
			tokenStream.Seek(0)

			// Parse the file using the API (which creates its own parser+listener)
			// We pass the already-read content via a strings.Reader
			stringReader := strings.NewReader(inputString)
			procedures, parseErr := core.ParseNeuroScript(stringReader) // Pass reader to parser API

			if parseErr != nil {
				// ** Stop walk on first parse error **
				wrappedErr := fmt.Errorf("parsing '%s': %w", filepath.Base(path), parseErr)
				fmt.Printf("    [Error] %v\n", wrappedErr) // Print the detailed error
				firstParseError = wrappedErr               // Store the error
				return firstParseError                     // <<< Return error here to STOP walking
			}

			// Load successfully parsed procedures
			if loadErr := interpreter.LoadProcedures(procedures); loadErr != nil {
				// This error shouldn't happen if parsing succeeded, but handle anyway
				wrappedErr := fmt.Errorf("loading from '%s': %w", filepath.Base(path), loadErr)
				fmt.Printf("    [Error] %v\n", wrappedErr)
				firstParseError = wrappedErr // Store the error
				return firstParseError       // Stop walk
			}
			fmt.Printf("    Successfully parsed and loaded %d procedures from %s\n", len(procedures), filepath.Base(path))
		}
		return nil // Continue walking normally if no error occurred in this step
	})

	// Check if the walk was stopped by an error (either access or parse error)
	if walkErr != nil && firstParseError != nil {
		fmt.Printf("\nDirectory walk stopped due to first error encountered: %v\n", firstParseError)
		os.Exit(1) // Exit with error status
	} else if walkErr != nil {
		// Handle other potential walk errors (e.g., permissions)
		fmt.Printf("\nError walking skills directory '%s': %v\n", skillsDir, walkErr)
		os.Exit(1)
	}

	// If walk completed without parse errors stopping it:
	if firstParseError == nil {
		fmt.Println("Finished loading procedures successfully.")
	} else {
		// This case shouldn't be reached if stopping logic works, but included for safety
		fmt.Println("Finished loading procedures (some files skipped due to errors).")
		os.Exit(1)
	}

	// --- Execute Requested Procedure ---
	// This part will only be reached if all files parsed correctly OR
	// if the walk completed before finding the target file (which is unlikely if stopping on error)
	fmt.Printf("\nExecuting procedure: %s with args: %v\n", procToRun, procArgs)
	fmt.Println("--------------------")

	result, runErr := interpreter.RunProcedure(procToRun, procArgs...) // Use variadic slice expansion

	fmt.Println("--------------------")
	fmt.Println("Execution finished.")

	if runErr != nil {
		fmt.Printf("Execution Error: %v\n", runErr)
		os.Exit(1) // Exit with error code if execution failed
	}

	// Only print result if execution was successful
	fmt.Printf("Final Result: %v\n", result)
	// Exit normally
}

// Helper type for diagnostic listener (copy from parser_api.go if needed)
// Ensure this matches the definition used in parser_api.go
type diagnosticErrorListener struct {
	*antlr.DefaultErrorListener
	errors []string
}

func newDiagnosticErrorListener() *diagnosticErrorListener {
	return &diagnosticErrorListener{errors: make([]string, 0)}
}
func (l *diagnosticErrorListener) SyntaxError(recognizer antlr.Recognizer, offendingSymbol interface{}, line, column int, msg string, e antlr.RecognitionException) {
	detailedMsg := fmt.Sprintf("line %d:%d %s", line, column, msg) // Basic message
	l.errors = append(l.errors, detailedMsg)
}
func (l *diagnosticErrorListener) HasErrors() bool   { return len(l.errors) > 0 }
func (l *diagnosticErrorListener) GetErrors() string { return strings.Join(l.errors, "\n") }
