// NeuroScript Version: 0.3.1
// File version: 0.1.2
// Purpose: CRITICAL FIX - Add blank import for toolbundles to ensure tools are registered at startup.
// filename: cmd/nslsp/main.go
// nlines: 55
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

// This version will be our ground truth.
const serverVersion = "1.0.2"

func main() {
	logger := log.New(os.Stderr, "[nslsp] ", log.LstdFlags|log.Lshortfile)
	logger.Printf("--- NeuroScript Language Server STARTING - VERSION %s ---", serverVersion)

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
