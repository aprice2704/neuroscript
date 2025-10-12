// NeuroScript Version: 0.8.0
// File version: 2
// Purpose: Defines a general-purpose, thread-safe manager for named text buffers.
// filename: pkg/interpreter/buffers.go
// nlines: 60
// risk_rating: MEDIUM

package interpreter

import (
	"bytes"
	"sync"
)

// BufferManager provides a general facility for managing named, writable text buffers.
type BufferManager struct {
	buffers map[string]*bytes.Buffer
	mu      sync.Mutex
}

// NewBufferManager creates and initializes a new BufferManager.
func NewBufferManager() *BufferManager {
	return &BufferManager{
		buffers: make(map[string]*bytes.Buffer),
	}
}

// Create registers a new buffer with the given handle.
func (m *BufferManager) Create(handle string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.buffers[handle]; !exists {
		m.buffers[handle] = &bytes.Buffer{}
	}
}

// Write appends data to the named buffer.
func (m *BufferManager) Write(handle, data string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if buffer, exists := m.buffers[handle]; exists {
		buffer.WriteString(data)
	}
}

// GetAndClear retrieves all content from a buffer and then resets it.
func (m *BufferManager) GetAndClear(handle string) string {
	m.mu.Lock()
	defer m.mu.Unlock()

	if buffer, exists := m.buffers[handle]; exists {
		content := buffer.String()
		buffer.Reset()
		return content
	}
	return ""
}
