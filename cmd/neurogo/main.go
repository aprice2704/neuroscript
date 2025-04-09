// filename: cmd/neurogo/main.go
package main

import (
	"context"
	"os"

	// Import the new neurogo library package
	"github.com/aprice2704/neuroscript/pkg/neurogo" // Adjusted import path
)

func main() {
	// 1. Create a new application instance from the library
	app := neurogo.NewApp()

	// 2. Parse command-line flags using the App's method
	//    os.Args[1:] excludes the program name itself
	err := app.ParseFlags(os.Args[1:])
	if err != nil {
		// Error message and usage already printed by ParseFlags on error
		// os.Stderr likely already has the usage info.
		os.Exit(1) // Exit if flag parsing failed
	}

	// 3. Run the application logic (which handles mode selection & logger init)
	//    Pass a background context for now. Agent mode might need a real context later.
	err = app.Run(context.Background())
	if err != nil {
		// App.Run logs errors via its configured ErrorLog.
		// Just exit non-zero.
		os.Exit(1)
	}

	// Exit 0 on success
	os.Exit(0)
}
