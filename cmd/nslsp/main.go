// NeuroScript Version: 0.3.1
// File version: 0.1.5
// Purpose: CRITICAL FIX - Silence all default logging at the very start of main to prevent stdout pollution that breaks LSP communication.
// filename: cmd/nslsp/main.go
// nlines: 66

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
	// completely clean for JSON-RPC communication. By setting the default logger's
	// output to discard at the absolute start of main(), we ensure no library
	// code can pollute stdout.
	log.SetOutput(io.Discard)

	// Application-specific logging for the server's own logic can still be
	// directed to a file for debugging purposes.
	logFile, err := os.OpenFile("/tmp/nslsp.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		// If we can't create the log file, we can't run.
		// A silent exit is bad, but logging to stderr would break the client.
		return
	}
	defer logFile.Close()

	// This new logger instance, which writes to the file, will be used by the server.
	logger := log.New(logFile, "[nslsp] ", log.LstdFlags|log.Lshortfile)

	logger.Printf("--- NeuroScript Language Server ---\n\tVersion: %s",
		buildDate)

	// When NewServer is called, it will use our file-based logger. Any underlying
	// calls in the api package that use the default logger will be silenced.
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
