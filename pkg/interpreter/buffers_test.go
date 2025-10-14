// NeuroScript Version: 0.8.0
// File version: 1
// Purpose: Adds concurrency tests for the BufferManager to ensure thread-safe writes.
// filename: pkg/interpreter/buffers_test.go
// nlines: 55
// risk_rating: MEDIUM

package interpreter_test

import (
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/interpreter"
)

func TestBufferManager_Concurrency(t *testing.T) {
	t.Run("Concurrent writes to the same buffer", func(t *testing.T) {
		bm := interpreter.NewBufferManager()
		bufferName := "concurrent_buffer"
		bm.Create(bufferName)

		var wg sync.WaitGroup
		numGoroutines := 100
		wg.Add(numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func(n int) {
				defer wg.Done()
				// Each goroutine writes a unique, identifiable string.
				writeData := fmt.Sprintf("[goroutine-%d]", n)
				bm.Write(bufferName, writeData)
			}(i)
		}

		wg.Wait()

		// Retrieve the final content of the buffer.
		finalContent := bm.GetAndClear(bufferName)

		// Verify that the content from every goroutine is present.
		// This confirms that writes were not lost or garbled.
		for i := 0; i < numGoroutines; i++ {
			expectedString := fmt.Sprintf("[goroutine-%d]", i)
			if !strings.Contains(finalContent, expectedString) {
				t.Errorf("Buffer content is missing expected string for goroutine %d", i)
			}
		}

		// Also verify that the buffer is now empty.
		if contentAfterClear := bm.GetAndClear(bufferName); contentAfterClear != "" {
			t.Errorf("Buffer was not empty after GetAndClear. Content: %s", contentAfterClear)
		}
	})
}
