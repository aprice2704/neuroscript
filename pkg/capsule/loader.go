// NeuroScript Version: 0.7.0
// File version: 2
// Purpose: Loads and registers all capsules from the content directory on init.
// filename: pkg/capsule/loader.go
// nlines: 57
// risk_rating: MEDIUM
package capsule

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"log"

	"github.com/aprice2704/neuroscript/pkg/metadata"
)

//go:embed all:content
var contentFS embed.FS

func init() {
	parser := metadata.NewMarkdownParser()
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

		meta, contentBody, err := parser.Parse(file)
		if err != nil {
			return fmt.Errorf("failed to parse metadata for %s: %w", path, err)
		}

		extractor := metadata.NewExtractor(meta)
		if err := extractor.CheckRequired(metadata.RequiredCapsuleKeys...); err != nil {
			log.Printf("Skipping file %s due to missing metadata: %v", path, err)
			return nil
		}

		priority, _ := extractor.GetIntOr("priority", 100)

		MustRegister(Capsule{
			Name:     extractor.MustGet("id"), // 'id' from markdown is now the 'Name'
			Version:  extractor.MustGet("version"),
			MIME:     extractor.GetOr("mime", "text/markdown; charset=utf-8"),
			Content:  string(bytes.TrimSpace(contentBody)),
			Priority: priority,
		})

		return nil
	})

	if err != nil {
		log.Fatalf("FATAL: Failed to walk and load embedded capsule content: %v", err)
	}
}
