// :: product: FDM/NS
// :: majorVersion: 1
// :: fileVersion: 14
// :: description: Corrects the file reading logic to satisfy the io.ReadSeeker interface required by the metadata parser. Adds filtering for valid extensions.
// :: latestChange: Added file extension and hidden file filtering to prevent noisy logs from .bash_history etc.
// :: filename: pkg/capsule/loader.go
// :: serialization: go

package capsule

import (
	"bytes"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"log"
	"path/filepath"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/metadata"
)

//go:embed all:content
var contentFS embed.FS

func init() {
	// THE FIX: Use the BuiltInRegistry, which is the one intended for loading.
	// The DefaultStore (in registry.go) will then consume this.
	reg := BuiltInRegistry()

	err := fs.WalkDir(contentFS, "content", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		// SKIP hidden files (like .bash_history) and non-capsule file types
		name := d.Name()
		if strings.HasPrefix(name, ".") {
			return nil
		}
		ext := filepath.Ext(name)
		if ext != ".md" && ext != ".ns" {
			return nil
		}

		file, err := contentFS.Open(path)
		if err != nil {
			return fmt.Errorf("failed to open embedded file %s: %w", path, err)
		}
		defer file.Close()

		// Read the file content into a byte slice first.
		fileBytes, err := io.ReadAll(file)
		if err != nil {
			return fmt.Errorf("failed to read embedded file %s: %w", path, err)
		}

		// Now, create a bytes.Reader, which implements io.ReadSeeker.
		reader := bytes.NewReader(fileBytes)
		meta, contentBody, _, err := metadata.ParseWithAutoDetect(reader)
		if err != nil {
			log.Printf("Skipping file %s: could not parse with auto-detect: %v", path, err)
			return nil // Continue walking
		}

		extractor := metadata.NewExtractor(meta)
		if err := extractor.CheckRequired("id", "version", "description"); err != nil {
			log.Printf("Skipping file %s due to missing metadata: %v", path, err)
			return nil
		}

		priority, _ := extractor.GetIntOr("priority", 100)

		reg.MustRegister(Capsule{
			Name:        extractor.MustGet("id"),
			Version:     extractor.MustGet("version"),
			Description: extractor.MustGet("description"),
			MIME:        extractor.GetOr("mime", "text/plain; charset=utf-8"),
			Content:     string(bytes.TrimSpace(contentBody)),
			Priority:    priority,
		})

		return nil
	})

	if err != nil {
		log.Fatalf("FATAL: Failed to walk and load embedded capsule content: %v", err)
	}
}
