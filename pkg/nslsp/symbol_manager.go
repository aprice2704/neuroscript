// NeuroScript Version: 0.7.0
// File version: 7
// Purpose: Upgrades SymbolManager to differentiate between minimum (needs) and maximum (optional) arguments. FIX: Reworked to perform synchronous, on-demand scanning of individual directories instead of a full workspace scan. FIX: Removed non-recursive logic to correctly scan subdirectories.
// filename: pkg/nslsp/symbol_manager.go
// nlines: 109
// risk_rating: HIGH

package nslsp

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/antlr4-go/antlr/v4"
	gen "github.com/aprice2704/neuroscript/pkg/antlr/generated"
	"github.com/aprice2704/neuroscript/pkg/parser"
	lsp "github.com/sourcegraph/go-lsp"
)

// SymbolInfo stores where a procedure is defined and its signature details.
type SymbolInfo struct {
	URI     lsp.DocumentURI
	Range   lsp.Range
	MinArgs int // From 'needs'
	MaxArgs int // From 'needs' + 'optional'
}

// SymbolManager scans the workspace and maintains a table of all procedure definitions.
type SymbolManager struct {
	mu          sync.RWMutex
	symbols     map[string]SymbolInfo
	scannedDirs map[string]struct{}
	parserAPI   *parser.ParserAPI
	logger      *log.Logger
}

// NewSymbolManager creates a new symbol manager.
func NewSymbolManager(logger *log.Logger) *SymbolManager {
	return &SymbolManager{
		symbols:     make(map[string]SymbolInfo),
		scannedDirs: make(map[string]struct{}),
		parserAPI:   parser.NewParserAPI(nil),
		logger:      logger,
	}
}

// ScanDirectory scans the given directory for procedure definitions if it hasn't been scanned before.
func (sm *SymbolManager) ScanDirectory(dirPath string) {
	sm.mu.RLock()
	_, alreadyScanned := sm.scannedDirs[dirPath]
	sm.mu.RUnlock()

	if alreadyScanned {
		return
	}

	sm.mu.Lock()
	// Double-check after acquiring write lock
	if _, alreadyScanned := sm.scannedDirs[dirPath]; alreadyScanned {
		sm.mu.Unlock()
		return
	}
	sm.scannedDirs[dirPath] = struct{}{}
	sm.mu.Unlock()

	sm.logger.Printf("SymbolManager: Starting scan of directory '%s'", dirPath)
	fileCount := 0
	procCountBefore := len(sm.symbols)

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// THE FIX IS HERE: The non-recursive check was removed. filepath.Walk will now correctly traverse subdirectories.
		if !info.IsDir() && (strings.HasSuffix(info.Name(), ".ns") || strings.HasSuffix(info.Name(), ".ns.txt")) {
			fileCount++
			sm.parseFileForSymbols(path)
		}
		return nil
	})

	if err != nil {
		sm.logger.Printf("ERROR: SymbolManager: Failed to scan directory '%s': %v", dirPath, err)
	}
	procCountAfter := len(sm.symbols)
	sm.logger.Printf("SymbolManager: Scan of '%s' complete. Found %d new procedures in %d files.", dirPath, procCountAfter-procCountBefore, fileCount)
}

// GetSymbolInfo finds a procedure in the workspace.
func (sm *SymbolManager) GetSymbolInfo(name string) (SymbolInfo, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	info, found := sm.symbols[name]
	return info, found
}

// parseFileForSymbols reads a file and adds any procedure definitions to the symbol table.
func (sm *SymbolManager) parseFileForSymbols(filePath string) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		sm.logger.Printf("WARN: SymbolManager: Could not read file %s: %v", filePath, err)
		return
	}

	tree, _ := sm.parserAPI.ParseForLSP(filePath, string(content))
	if tree == nil {
		return
	}

	listener := &symbolScanListener{
		sm:  sm,
		uri: lsp.DocumentURI("file://" + filePath),
	}
	walker := antlr.NewParseTreeWalker()
	walker.Walk(listener, tree)
}

// symbolScanListener is a simple ANTLR listener that extracts procedure names and arg counts.
type symbolScanListener struct {
	*gen.BaseNeuroScriptListener
	sm  *SymbolManager
	uri lsp.DocumentURI
}

func (l *symbolScanListener) EnterProcedure_definition(ctx *gen.Procedure_definitionContext) {
	procName := ctx.IDENTIFIER().GetText()
	if procName == "" {
		return
	}

	needsCount := 0
	optionalCount := 0
	if sig := ctx.Signature_part(); sig != nil {
		if needs := sig.Needs_clause(0); needs != nil && needs.Param_list() != nil {
			needsCount = len(needs.Param_list().AllIDENTIFIER())
		}
		if optional := sig.Optional_clause(0); optional != nil && optional.Param_list() != nil {
			optionalCount = len(optional.Param_list().AllIDENTIFIER())
		}
	}

	l.sm.mu.Lock()
	defer l.sm.mu.Unlock()

	token := ctx.IDENTIFIER().GetSymbol()
	l.sm.symbols[procName] = SymbolInfo{
		URI:     l.uri,
		Range:   lspRangeFromToken(token, procName),
		MinArgs: needsCount,
		MaxArgs: needsCount + optionalCount,
	}
}
