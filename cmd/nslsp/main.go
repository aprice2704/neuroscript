// NeuroScript Version: 0.3.1
// File version: 0.1.3
// Purpose: CRITICAL FIX - Add blank import for toolbundles to ensure tools are registered at startup. FEAT: Add build-time variable injection for versioning.
// filename: cmd/nslsp/main.go
// nlines: 63
// risk_rating: LOW

package main

import (
	"context"
	"log"
	"os"

	"github.com/aprice2704/neuroscript/pkg/nslsp"
	_ "github.com/aprice2704/neuroscript/pkg/toolbundles/all" // register the tools
	"github.com/sourcegraph/jsonrpc2"
)

var buildDate string // These variables are set by the linker during the build process.

func main() {

	// THE FIX IS HERE: Log to a file instead of stderr.
	logFile, err := os.OpenFile("/tmp/nslsp.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		// If we can't create the log file, we can't run.
		// A silent exit is bad, but logging to stderr would break the client.
		// In a real-world scenario, you might try a fallback path.
		return
	}
	defer logFile.Close()

	logger := log.New(logFile, "[nslsp] ", log.LstdFlags|log.Lshortfile)

	logger.Printf("--- NeuroScript Language Server ---\n\tVersion: %s",
		buildDate)

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
