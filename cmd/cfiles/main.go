package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings" // Added for checking API key prefix

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

func main() {
	ctx := context.Background()
	log.Println("Starting program...") // DEBUG

	// --- Get API Key ---
	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		log.Fatal("DEBUG: API key not found in environment variable GOOGLE_API_KEY. Please ensure it is set correctly (e.g., using 'export GOOGLE_API_KEY=YOUR_KEY').")
	} else {
		// DEBUG: Confirm key is loaded, show first few chars for verification (don't log the whole key!)
		prefix := apiKey[:min(5, len(apiKey))] + strings.Repeat("*", max(0, len(apiKey)-5))
		log.Printf("DEBUG: Found API key starting with: %s\n", prefix)
	}

	// --- Create Client ---
	log.Println("DEBUG: Creating Gemini client...") // DEBUG
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		// Log the specific error during client creation
		log.Fatalf("DEBUG: Failed to create Gemini client: %v\n", err)
	}
	// Ensure client is closed even if subsequent operations fail
	defer func() {
		log.Println("DEBUG: Closing client...") // DEBUG
		client.Close()
		log.Println("DEBUG: Client closed.") // DEBUG
	}()
	log.Println("DEBUG: Gemini client created successfully.") // DEBUG

	fmt.Println("Listing files...")

	// --- List Files ---
	log.Println("DEBUG: Calling client.ListFiles(ctx)...") // DEBUG
	iter := client.ListFiles(ctx)
	if iter == nil {
		log.Fatal("DEBUG: client.ListFiles returned a nil iterator!") // Should not happen, but check
	}
	log.Println("DEBUG: Received iterator from ListFiles.") // DEBUG

	// --- Iterate and Print ---
	fileCount := 0
	log.Println("DEBUG: Starting iteration over files...") // DEBUG
	for {
		log.Println("DEBUG: Calling iter.Next()...") // DEBUG
		file, err := iter.Next()
		if err == iterator.Done {
			// Reached the end of the list
			log.Println("DEBUG: iterator.Done received. End of list.") // DEBUG
			break
		}
		if err != nil {
			// Handle potential errors during iteration - Log more prominently
			log.Printf("ERROR: Error fetching next file during iteration: %v\n", err)
			// Decide if you want to stop on the first error or just log and continue
			// For debugging, let's break here to see the first error clearly.
			log.Println("DEBUG: Breaking loop due to iteration error.") // DEBUG
			break
		}

		// If we get here, a file was successfully retrieved
		log.Printf("DEBUG: Successfully fetched file: Name=%s, DisplayName=%s\n", file.Name, file.DisplayName) // DEBUG
		fileCount++

		// Print file details
		fmt.Println("--- File ---")
		fmt.Printf("  Name:        %s\n", file.Name)
		fmt.Printf("  Display Name: %s\n", file.DisplayName)
		fmt.Printf("  URI:         %s\n", file.URI)
		fmt.Printf("  MIME Type:   %s\n", file.MIMEType)
		fmt.Printf("  Size:        %d bytes\n", file.SizeBytes)
		fmt.Printf("  Create Time: %s\n", file.CreateTime)
		fmt.Printf("  Update Time: %s\n", file.UpdateTime)
		fmt.Println("------------")
	}

	log.Printf("DEBUG: Finished iteration. Total files found: %d\n", fileCount) // DEBUG

	if fileCount == 0 {
		fmt.Println("No files found.")
	} else {
		fmt.Printf("\nFound %d file(s).\n", fileCount)
	}
	log.Println("DEBUG: ast.Program finished.") // DEBUG
}

// Helper functions for showing API key prefix safely
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
