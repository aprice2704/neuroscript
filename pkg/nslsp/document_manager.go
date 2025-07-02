// NeuroScript Version: 0.3.1
// File version: 0.1.0
// Purpose: Manages in-memory store of open document contents for the LSP server.
// filename: pkg/nslsp/document_manager.go
// nlines: 35 // Approximate
// risk_rating: LOW // Simple map with mutex.

package nslsp

import (
	"sync"

	lsp "github.com/sourcegraph/go-lsp"
)

type DocumentManager struct {
	mu		sync.RWMutex
	documents	map[lsp.DocumentURI]string
}

func NewDocumentManager() *DocumentManager {
	return &DocumentManager{
		documents: make(map[lsp.DocumentURI]string),
	}
}

func (dm *DocumentManager) Get(uri lsp.DocumentURI) (string, bool) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	content, found := dm.documents[uri]
	return content, found
}

func (dm *DocumentManager) Set(uri lsp.DocumentURI, content string) {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	dm.documents[uri] = content
}

func (dm *DocumentManager) Delete(uri lsp.DocumentURI) {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	delete(dm.documents, uri)
}