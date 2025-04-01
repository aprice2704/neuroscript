// gonsi/main.go
package main

import (
	// "bytes" // No longer needed for buffering
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	// "time" // No longer needed for token timing here

	"github.com/antlr4-go/antlr/v4"
	"github.com/aprice2704/neuroscript/pkg/core"
	gen "github.com/aprice2704/neuroscript/pkg/core/generated"
)

// --- Logger Setup ---
var (
	infoLog  *log.Logger
	debugLog *log.Logger
	errorLog *log.Logger
)

// Define command-line flags
var (
	debugTokens      = flag.Bool("debug-tokens", false, "Enable verbose token logging during lexing")
	debugAST         = flag.Bool("debug-ast", false, "Enable verbose AST node logging during parsing")
	debugInterpreter = flag.Bool("debug-interpreter", false, "Enable verbose interpreter execution logging")
	// debugOnError      = flag.Bool("debug-on-error", false, "(No longer used effectively)") // Kept for compatibility but logic removed
)

func initLoggers(enableDebug bool) {
	infoLog = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime)
	errorLog = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

	// --- Simplified Debug Logger Setup ---
	// If any debug flag is true, log DEBUG messages to stderr. Otherwise, discard.
	debugOutput := io.Discard
	if enableDebug {
		debugOutput = os.Stderr
		fmt.Fprintf(os.Stderr, "--- Debug Logging Enabled ---\n") // Add clear indicator
	}
	debugLog = log.New(debugOutput, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)
	// --- End Simplified Setup ---

	log.SetOutput(os.Stderr) // Fatal logs
	log.SetPrefix("FATAL: ")
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

func main() {
	flag.Parse()

	// --- Simplified Debug Enable Logic ---
	// Enable debug logger if ANY specific debug flag is set.
	enableMainDebug := *debugTokens || *debugAST || *debugInterpreter
	initLoggers(enableMainDebug)
	// --- End Simplified Logic ---

	args := flag.Args()
	if len(args) < 2 {
		fmt.Println("Usage: gonsi [flags] <skills_directory> <ProcedureToRun> [args...]")
		flag.PrintDefaults()
		os.Exit(1)
	}
	skillsDir := args[0]
	procToRun := args[1]
	procArgs := []string{}
	if len(args) > 2 {
		procArgs = args[2:]
	}

	// Pass the potentially enabled debugLog to Interpreter and Parser
	interpreter := core.NewInterpreter(debugLog)
	infoLog.Printf("Loading procedures from directory: %s", skillsDir)
	var firstErrorEncountered error // Track first error to stop walk

	walkErr := filepath.Walk(skillsDir, func(path string, info os.FileInfo, walkErrIn error) error {
		// Stop walk immediately if an error has already occurred
		if firstErrorEncountered != nil {
			return firstErrorEncountered
		}
		if walkErrIn != nil {
			errorLog.Printf("Error accessing path %q: %v", path, walkErrIn)
			return nil // Continue walking other paths if possible
		}

		if !info.IsDir() && strings.HasSuffix(info.Name(), ".ns.txt") {
			fileName := filepath.Base(path)
			// Log directly using the main debugLog now
			debugLog.Printf("--- Processing File: %s ---", fileName)

			// Read file content
			contentBytes, readErr := os.ReadFile(path)
			if readErr != nil {
				errorLog.Printf("Could not read file %s: %v", fileName, readErr)
				return nil // Continue walk
			}
			inputString := string(contentBytes)
			inputStream := antlr.NewInputStream(inputString)

			// --- Lexing ---
			lexer := gen.NewNeuroScriptLexer(inputStream)
			lexerErrorListener := core.NewDiagnosticErrorListener(fileName) // Use exported listener
			lexer.RemoveErrorListeners()
			lexer.AddErrorListener(lexerErrorListener)
			tokenStream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)

			// Optionally log tokens if requested
			if *debugTokens {
				tokenStream.Fill() // Need to fill before getting all tokens
				debugLog.Printf("    Tokens for %s:", fileName)
				tokens := tokenStream.GetAllTokens()
				tokenNames := lexer.GetSymbolicNames()
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
					debugLog.Printf("      %s (%q) L%d:%d Ch:%d", tokenName, token.GetText(), token.GetLine(), token.GetColumn(), token.GetChannel())
				}
				debugLog.Println("    --- End Tokens ---")
				tokenStream.Seek(0) // Reset stream for parser
			}

			// --- Parsing ---
			parser := gen.NewNeuroScriptParser(tokenStream)
			parserErrorListener := core.NewDiagnosticErrorListener(fileName) // Use exported listener
			parser.RemoveErrorListeners()
			parser.AddErrorListener(parserErrorListener)

			// Pass DebugAST flag and the main debugLog to the parser API
			parseOptions := core.ParseOptions{
				DebugAST: *debugAST, // Generate AST debug logs if flag is set
				Logger:   debugLog,  // Use the main debug logger
			}
			stringReader := strings.NewReader(inputString)
			procedures, parseErr := core.ParseNeuroScript(stringReader, fileName, parseOptions)

			// Combine and check errors
			allErrors := append(lexerErrorListener.Errors, parserErrorListener.Errors...)
			hasError := len(allErrors) > 0 || parseErr != nil
			if parseErr != nil && len(allErrors) == 0 {
				allErrors = append(allErrors, parseErr.Error())
			}

			if hasError {
				errorMsg := fmt.Sprintf("Parse error processing %s:\n%s", fileName, strings.Join(allErrors, "\n"))
				errorLog.Print(errorMsg)          // Log the specific error
				if firstErrorEncountered == nil { // Store the first error
					if parseErr != nil {
						firstErrorEncountered = parseErr
					} else {
						firstErrorEncountered = fmt.Errorf(errorMsg)
					}
				}
				return firstErrorEncountered // Stop walk
			}

			// Load successfully parsed procedures
			if loadErr := interpreter.LoadProcedures(procedures); loadErr != nil {
				wrappedErr := fmt.Errorf("loading from '%s': %w", fileName, loadErr)
				errorLog.Printf("Load error: %v", wrappedErr)
				if firstErrorEncountered == nil {
					firstErrorEncountered = wrappedErr
				}
				return firstErrorEncountered // Stop walk
			}
			debugLog.Printf("Successfully parsed and loaded %d procedures from %s", len(procedures), fileName)
		}
		return nil // Continue walk
	})

	// --- Post-Walk Handling ---
	// Check if the walk was stopped by an error stored in firstErrorEncountered
	if firstErrorEncountered != nil {
		infoLog.Printf("Directory walk stopped due to error: %v", firstErrorEncountered)
		os.Exit(1)
	} else if walkErr != nil { // Handle errors from filepath.Walk itself (rare)
		errorLog.Printf("Error walking skills directory '%s': %v", skillsDir, walkErr)
		os.Exit(1)
	}

	// Log success only if no errors occurred
	infoLog.Println("Finished loading procedures successfully.")

	// --- Execute Requested Procedure ---
	infoLog.Printf("Executing procedure: %s with args: %v", procToRun, procArgs)
	fmt.Println("--------------------") // Separator before execution output

	result, runErr := interpreter.RunProcedure(procToRun, procArgs...)

	fmt.Println("--------------------") // Separator after execution output
	infoLog.Println("Execution finished.")

	if runErr != nil {
		// Error is already formatted with context by the interpreter
		errorLog.Printf("Execution Error: %v", runErr)
		os.Exit(1)
	}

	infoLog.Printf("Final Result: %v", result)
}
