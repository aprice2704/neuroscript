// NeuroScript Version: 0.3.1
// File version: 0.1.0
// Purpose: Main application entry point for the NeuroScript Language Server (nslsp).
// filename: cmd/nslsp/main.go
// nlines: 50 // Approximate, will vary with actual implementation
// risk_rating: LOW // Primarily setup and boilerplate for LSP communication.

package main

import (
	"context"
	"log"
	"os"

	"github.com/aprice2704/neuroscript/pkg/nslsp"
	"github.com/sourcegraph/jsonrpc2"
)

func main() {
	logger := log.New(os.Stderr, "[nslsp] ", log.LstdFlags|log.Lshortfile)
	logger.Println("Starting NeuroScript Language Server...")

	server := nslsp.NewServer(logger)

	var connOpt []jsonrpc2.ConnOpt
	// Example: Add a logger for jsonrpc2 messages (can be verbose)
	// connOpt = append(connOpt, jsonrpc2.LogMessages(log.New(os.Stderr, "[jsonrpc2] ", 0)))

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
