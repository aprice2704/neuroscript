// NeuroScript Version: 0.6.0
// File version: 0.1.1
// Purpose: Manages in-memory store of open document contents, with a new method to retrieve all documents.
// filename: pkg/nslsp/document_manager.go
// nlines: 45
// risk_rating: LOW

package nslsp

import (
	"sync"

	lsp "github.com/sourcegraph/go-lsp"
)

type DocumentManager struct {
	mu        sync.RWMutex
	documents map[lsp.DocumentURI]string
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

// GetAll returns a copy of the current documents map.
func (dm *DocumentManager) GetAll() map[lsp.DocumentURI]string {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	docsCopy := make(map[lsp.DocumentURI]string)
	for uri, content := range dm.documents {
		docsCopy[uri] = content
	}
	return docsCopy
}
