// :: product: FDM/NS
// :: majorVersion: 1
// :: fileVersion: 2
// :: description: CRITICAL FIX - Redirects logging to Stderr so it appears in VS Code Output.
// :: latestChange: Changed logger to use os.Stderr instead of /tmp/nslsp.log.
// :: filename: cmd/nslsp/main.go
// :: serialization: go
package main

import (
	"context"
	"io"
	"log"
	"os"

	"github.com/aprice2704/neuroscript/pkg/nslsp"
	_ "github.com/aprice2704/neuroscript/pkg/toolbundles/all" // register the tools
	"github.com/sourcegraph/jsonrpc2"
)

var buildDate string // These variables are set by the linker during the build process.

func main() {
	// CRITICAL FIX: The NeuroScript API's initialization routines can log debug
	// messages to the default global logger. For the LSP, stdout MUST be kept
	// completely clean for JSON-RPC communication.
	log.SetOutput(io.Discard)

	// We use Stderr for logs, which VS Code (and most LSP clients) capture
	// and display in the Output tab. This avoids file permission issues and
	// allows immediate feedback.
	logger := log.New(os.Stderr, "[nslsp] ", log.LstdFlags|log.Lshortfile)

	logger.Printf("--- NeuroScript Language Server ---\n\tVersion: %s",
		buildDate)

	// When NewServer is called, it will use our stderr logger. Any underlying
	// calls in the api package that use the default logger (if using the adapter)
	// will also route here.
	server := nslsp.NewServer(logger)

	var connOpt []jsonrpc2.ConnOpt

	<-jsonrpc2.NewConn(
		context.Background(),
		jsonrpc2.NewBufferedStream(stdrwc{}, jsonrpc2.VSCodeObjectCodec{}),
		jsonrpc2.HandlerWithError(server.Handle),
		connOpt...,
	).DisconnectNotify()

	logger.Println("NeuroScript Language Server stopped.")
}

// stdrwc is a simple ReadWriteCloser for os.Stdin and os.Stdout.
type stdrwc struct{}

func (stdrwc) Read(p []byte) (int, error) {
	return os.Stdin.Read(p)
}

func (stdrwc) Write(p []byte) (int, error) {
	return os.Stdout.Write(p)
}

func (stdrwc) Close() error {
	if err := os.Stdin.Close(); err != nil {
		return err
	}
	return os.Stdout.Close()
}
