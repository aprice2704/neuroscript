// NeuroScript Version: 0.7.2
// File version: 12
// Purpose: Corrects the file reading logic to satisfy the io.ReadSeeker interface required by the metadata parser.
// filename: pkg/capsule/loader.go
// nlines: 75
// risk_rating: HIGH
package capsule

import (
	"bytes"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"log"

	"github.com/aprice2704/neuroscript/pkg/metadata"
)

//go:embed all:content
var contentFS embed.FS

func init() {
	reg := DefaultRegistry()

	err := fs.WalkDir(contentFS, "content", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
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
			Name:     extractor.MustGet("id"),
			Version:  extractor.MustGet("version"),
			MIME:     extractor.GetOr("mime", "text/plain; charset=utf-8"),
			Content:  string(bytes.TrimSpace(contentBody)),
			Priority: priority,
		})

		return nil
	})

	if err != nil {
		log.Fatalf("FATAL: Failed to walk and load embedded capsule content: %v", err)
	}
}
