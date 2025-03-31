// gonsi/main.go
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

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
	debugTokens  = flag.Bool("debug-tokens", false, "Enable verbose token logging during lexing (ignored if -debug-on-error is set)")
	debugAST     = flag.Bool("debug-ast", false, "Enable verbose AST node logging during parsing (ignored if -debug-on-error is set)")
	debugOnError = flag.Bool("debug-on-error", false, "Only show token/AST debug output for files with parse errors (overrides other debug flags)") // Priority flag
)

func initLoggers(enableDebug bool) {
	infoLog = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime)
	errorLog = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

	debugOutput := io.Discard
	if enableDebug {
		debugOutput = os.Stderr
	}
	debugLog = log.New(debugOutput, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)

	log.SetOutput(os.Stderr)
	log.SetPrefix("FATAL: ")
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

func main() {
	flag.Parse()

	// Determine if the main debug channel should be enabled
	// Enable if any debug flag is set, as we buffer and decide later
	enableMainDebug := *debugTokens || *debugAST || *debugOnError
	initLoggers(enableMainDebug)

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

	interpreter := core.NewInterpreter()
	infoLog.Printf("Loading procedures from directory: %s", skillsDir)
	var firstParseError error

	walkErr := filepath.Walk(skillsDir, func(path string, info os.FileInfo, err error) error {
		if firstParseError != nil {
			return firstParseError
		}
		if err != nil {
			errorLog.Printf("Error accessing path %q: %v", path, err)
			return nil
		}

		if !info.IsDir() && strings.HasSuffix(info.Name(), ".ns.txt") {
			fileName := filepath.Base(path)
			debugLog.Printf("--- Processing File: %s ---", fileName)

			var fileDebugBuffer bytes.Buffer
			fileDebugLogger := log.New(&fileDebugBuffer, "", 0)

			// Determine if debug info *needs* to be generated/captured, even if not immediately flushed.
			// We capture if any debug flag is on, because -debug-on-error might need it later.
			captureTokens := *debugTokens || *debugOnError
			captureAST := *debugAST || *debugOnError

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
			basicLexerErrorListener := &errorListener{sourceName: fileName}
			lexer.RemoveErrorListeners()
			lexer.AddErrorListener(basicLexerErrorListener)
			tokenStream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
			tokenReadStart := time.Now()
			tokenStream.Fill()
			tokenReadDuration := time.Since(tokenReadStart)

			// --- Conditional Token Logging (to buffer) ---
			if captureTokens { // Capture token info if potentially needed
				fileDebugLogger.Printf("    Tokens for %s (Read in %v):", fileName, tokenReadDuration)
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
					fileDebugLogger.Printf("      %s (%q) L%d:%d Ch:%d",
						tokenName, token.GetText(), token.GetLine(), token.GetColumn(), token.GetChannel())
				}
				fileDebugLogger.Println("    --- End Tokens ---")
			}

			// Check for lexer errors before parsing
			hasLexerError := len(basicLexerErrorListener.errors) > 0
			if hasLexerError {
				errStr := strings.Join(basicLexerErrorListener.errors, "\n")
				errorLog.Printf("Skipping file '%s' due to lexer errors:\n%s", fileName, errStr)
				// --- Flush Debug Buffer on Lexer Error ONLY if -debug-on-error is set ---
				if *debugOnError { // Check the priority flag
					debugLog.Printf("--- Buffered Debug Output for %s (Lexer Error with -debug-on-error) ---", fileName)
					bufferContent := fileDebugBuffer.Bytes()
					if len(bufferContent) > 0 && len(strings.TrimSpace(string(bufferContent))) > 0 {
						debugLog.Writer().Write(bufferContent)
					} else {
						debugLog.Printf("(No non-whitespace buffered debug output to show)")
					}
					debugLog.Printf("--- End Buffered Debug Output ---")
				}
				// --- End Flush ---
				firstParseError = fmt.Errorf("lexer error in '%s'", fileName)
				return firstParseError // Stop walk
			}

			// --- Parsing ---
			tokenStream.Seek(0)
			stringReader := strings.NewReader(inputString)
			parseOptions := core.ParseOptions{
				DebugAST: captureAST,      // Tell listener to capture if potentially needed
				Logger:   fileDebugLogger, // Always log to buffer if captureAST is true
			}
			procedures, parseErr := core.ParseNeuroScript(stringReader, fileName, parseOptions)
			hasParseError := parseErr != nil

			// --- Conditional Flushing of Buffered Debug Logs (Revised Priority Logic) ---
			shouldFlushBuffer := false
			flushReason := ""

			if *debugOnError {
				// If -debug-on-error is set, ONLY flush if there was an error
				if hasParseError {
					shouldFlushBuffer = true
					flushReason = "(Parse Error with -debug-on-error)"
				}
			} else {
				// If -debug-on-error is NOT set, flush if explicit flags were set
				if *debugTokens || *debugAST {
					shouldFlushBuffer = true
					reasons := []string{}
					if *debugTokens {
						reasons = append(reasons, "-debug-tokens")
					}
					if *debugAST {
						reasons = append(reasons, "-debug-ast")
					}
					flushReason = "(" + strings.Join(reasons, ", ") + ")"
				}
			}

			if shouldFlushBuffer {
				debugLog.Printf("--- Buffered Debug Output for %s %s ---", fileName, flushReason)
				bufferContent := fileDebugBuffer.Bytes()
				if len(bufferContent) > 0 {
					trimmedContent := strings.TrimSpace(string(bufferContent))
					if len(trimmedContent) > 0 {
						debugLog.Writer().Write(bufferContent)
					} else {
						debugLog.Printf("(No non-whitespace buffered debug output to show)")
					}
				} else {
					debugLog.Printf("(No buffered debug output to show)")
				}
				debugLog.Printf("--- End Buffered Debug Output ---")
			}
			// --- End Conditional Flushing ---

			if hasParseError {
				errorLog.Printf("Parse error processing %s.", fileName)
				firstParseError = parseErr
				return firstParseError // Stop walk
			}

			// Load successfully parsed procedures
			if loadErr := interpreter.LoadProcedures(procedures); loadErr != nil {
				wrappedErr := fmt.Errorf("loading from '%s': %w", fileName, loadErr)
				errorLog.Printf("Load error: %v", wrappedErr)
				firstParseError = wrappedErr
				return firstParseError // Stop walk
			}
			// Log successful load only if needed (main debug enabled)
			debugLog.Printf("Successfully parsed and loaded %d procedures from %s", len(procedures), fileName)
		}
		return nil
	})

	// --- Post-Walk Handling ---
	if walkErr != nil && firstParseError != nil {
		infoLog.Printf("Directory walk stopped due to first error encountered.")
		os.Exit(1)
	} else if walkErr != nil {
		errorLog.Printf("Error walking skills directory '%s': %v", skillsDir, walkErr)
		os.Exit(1)
	}

	if firstParseError == nil {
		infoLog.Println("Finished loading procedures successfully.")
	} else {
		infoLog.Println("Directory walk stopped due to errors.")
		os.Exit(1)
	}

	// --- Execute Requested Procedure ---
	infoLog.Printf("Executing procedure: %s with args: %v", procToRun, procArgs)
	fmt.Println("--------------------")

	result, runErr := interpreter.RunProcedure(procToRun, procArgs...)

	fmt.Println("--------------------")
	infoLog.Println("Execution finished.")

	if runErr != nil {
		errorLog.Printf("Execution Error: %v", runErr)
		os.Exit(1)
	}

	infoLog.Printf("Final Result: %v", result)
}

// --- Minimal Error Listener for Lexer ---
type errorListener struct {
	*antlr.DefaultErrorListener
	errors     []string
	sourceName string
}

func (l *errorListener) SyntaxError(recognizer antlr.Recognizer, offendingSymbol interface{}, line, column int, msg string, e antlr.RecognitionException) {
	errMsg := fmt.Sprintf("%s:%d:%d %s", l.sourceName, line, column, msg)
	l.errors = append(l.errors, errMsg)
}
