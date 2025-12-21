// :: product: FDM/NS
// :: majorVersion: 1
// :: fileVersion: 13
// :: description: Upgrades SymbolManager with robust URI encoding and signature formatting.
// :: latestChange: FIX: Use net/url for URI construction. Re-apply signature parentheses fix.
// :: filename: pkg/nslsp/symbol_manager.go
// :: serialization: go
package nslsp

import (
	"log"
	"net/url"
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
	URI       lsp.DocumentURI
	Range     lsp.Range
	MinArgs   int    // From 'needs'
	MaxArgs   int    // From 'needs' + 'optional'
	Signature string // e.g. "(needs a, b, optional c)"
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
		if info.IsDir() {
			// Skip hidden directories (like .git, .vscode) and standard build dirs
			name := info.Name()
			if name != "." && (strings.HasPrefix(name, ".") || name == "vendor" || name == "bin" || name == "node_modules") {
				sm.logger.Printf("SymbolManager: Skipping directory: %s", path)
				return filepath.SkipDir
			}
			return nil
		}

		if strings.HasSuffix(info.Name(), ".ns") || strings.HasSuffix(info.Name(), ".ns.txt") {
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

// UpdateSymbol parses the content of a single file and updates the symbol table.
// It removes any old symbols associated with this URI before adding new ones.
func (sm *SymbolManager) UpdateSymbol(uri lsp.DocumentURI, content string) {
	sm.logger.Printf("SymbolManager: Updating symbols for %s", uri)

	// 1. Parse the new content
	tree, _ := sm.parserAPI.ParseForLSP(string(uri), content)
	if tree == nil {
		return
	}

	walker := antlr.NewParseTreeWalker()

	// 2. Lock to update the map
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// 3. Clear existing symbols for this URI
	deletedCount := 0
	for name, info := range sm.symbols {
		if info.URI == uri {
			delete(sm.symbols, name)
			deletedCount++
		}
	}

	// 4. Walk the new tree to populate symbols (listener will add to sm.symbols)
	newSymbols := make(map[string]SymbolInfo)
	collector := &symbolCollectorMapListener{
		uri:        uri,
		newSymbols: newSymbols,
	}
	// Unlock for the walk (CPU intensive part)
	sm.mu.Unlock()

	walker.Walk(collector, tree)

	// Re-lock to merge
	sm.mu.Lock()
	for name, info := range newSymbols {
		sm.symbols[name] = info
	}
	sm.logger.Printf("SymbolManager: Updated %s. Cleared %d old symbols, added %d new ones.", uri, deletedCount, len(newSymbols))
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

	// FIX: Use net/url to construct a proper file:// URI with encoding (spaces, etc.)
	u := url.URL{
		Scheme: "file",
		Path:   filePath,
	}
	uri := lsp.DocumentURI(u.String())

	sm.UpdateSymbol(uri, string(content))
}

// symbolCollectorMapListener collects symbols into a local map, avoiding lock contention during the walk.
type symbolCollectorMapListener struct {
	*gen.BaseNeuroScriptListener
	uri        lsp.DocumentURI
	newSymbols map[string]SymbolInfo
}

func (l *symbolCollectorMapListener) EnterProcedure_definition(ctx *gen.Procedure_definitionContext) {
	procName := ctx.IDENTIFIER().GetText()
	if procName == "" {
		return
	}
	needs, optional, signature := extractArgsAndSignature(ctx)

	token := ctx.IDENTIFIER().GetSymbol()
	l.newSymbols[procName] = SymbolInfo{
		URI:       l.uri,
		Range:     lspRangeFromToken(token, procName),
		MinArgs:   needs,
		MaxArgs:   needs + optional,
		Signature: signature,
	}
}

// extractArgsAndSignature helper to get counts AND the display string
func extractArgsAndSignature(ctx *gen.Procedure_definitionContext) (int, int, string) {
	needsCount := 0
	optionalCount := 0
	var parts []string

	if sig := ctx.Signature_part(); sig != nil {
		if needs := sig.Needs_clause(0); needs != nil && needs.Param_list() != nil {
			var params []string
			for _, p := range needs.Param_list().AllIDENTIFIER() {
				params = append(params, p.GetText())
			}
			needsCount = len(params)
			parts = append(parts, "needs "+strings.Join(params, ", "))
		}
		if optional := sig.Optional_clause(0); optional != nil && optional.Param_list() != nil {
			var params []string
			for _, p := range optional.Param_list().AllIDENTIFIER() {
				params = append(params, p.GetText())
			}
			optionalCount = len(params)
			parts = append(parts, "optional "+strings.Join(params, ", "))
		}
	}

	// Format: "(needs a, b, optional c)"
	// If no args, this becomes "()"
	signature := "(" + strings.Join(parts, ", ") + ")"

	return needsCount, optionalCount, signature
}
