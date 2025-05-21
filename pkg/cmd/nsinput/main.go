// Filename: ns-input/main.go
package main

import (
	// Import errors package
	"fmt"
	"log"
	"os"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

// Represents the state of our TUI application
type model struct {
	textarea   textarea.Model
	err        error
	submitted  bool   // Flag to indicate if input was submitted vs cancelled
	outputFile string // ADDED: File path to write output to
}

// Creates the initial state of the application model
func initialModel(outputFile string) model { // ADDED: outputFile parameter
	ti := textarea.New()
	ti.Placeholder = "Enter your multi-line prompt here..."
	ti.Focus()
	ti.CharLimit = 0
	ti.SetHeight(15) // Increased height example
	// Width will be set dynamically via WindowSizeMsg

	return model{
		textarea:   ti,
		err:        nil,
		submitted:  false,
		outputFile: outputFile, // Store the output file path
	}
}

// Init is the first command run when the program starts.
func (m model) Init() tea.Cmd {
	return textarea.Blink
}

// Update handles messages and updates the model.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			m.submitted = false
			return m, tea.Quit // Quit the program (will not write output)

		case tea.KeyCtrlD, tea.KeyEsc:
			m.submitted = true
			return m, tea.Quit // Quit the program (will write output)

		default:
			m.textarea, cmd = m.textarea.Update(msg)
			cmds = append(cmds, cmd)
		}

	case tea.WindowSizeMsg:
		m.textarea.SetWidth(msg.Width)
		// Example: Adjust height, leaving room for instructions
		// m.textarea.SetHeight(msg.Height - 3)
	}

	return m, tea.Batch(cmds...)
}

// View renders the UI based on the model's state.
func (m model) View() string {
	return fmt.Sprintf(
		"Enter prompt. Submit: Ctrl+D or Esc | Cancel: Ctrl+C\n\n%s",
		m.textarea.View(),
	)
}

func main() {
	// --- Get output file path from argument ---
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "Usage: nsinput <output_file_path>")
		os.Exit(1)
	}
	outputFilePath := os.Args[1]
	// ---

	// --- Optional: Log Bubble Tea internal stuff to a file ---
	// f, errLog := tea.LogToFile("ns-input-debug.log", "debug")
	// if errLog != nil { fmt.Println("could not create log file:", errLog); os.Exit(1) }
	// defer f.Close()
	// ---

	p := tea.NewProgram(initialModel(outputFilePath)) // Pass output path to model

	mFinal, err := p.Run()
	if err != nil {
		log.Printf("Bubbletea exited with error: %v", err)
		os.Exit(1)
	}

	// --- Output the result to the specified file ---
	if m, ok := mFinal.(model); ok {
		if m.submitted {
			// Write the final textarea value to the output file
			writeErr := os.WriteFile(m.outputFile, []byte(m.textarea.Value()), 0644)
			if writeErr != nil {
				// Write error to stderr so neurogo might see it
				fmt.Fprintf(os.Stderr, "Error writing output file '%s': %v\n", m.outputFile, writeErr)
				os.Exit(1) // Exit with error if write fails
			}
			// fmt.Fprintf(os.Stderr, "Successfully wrote %d bytes to %s\n", len(m.textarea.Value()), m.outputFile) // Debug message to stderr
		} else {
			// If cancelled (Ctrl+C), ensure the output file is empty or doesn't exist
			// Writing an empty file is a simple way to signal cancellation clearly.
			_ = os.WriteFile(m.outputFile, []byte(""), 0644) // Write empty file on cancel
			// fmt.Fprintln(os.Stderr, "Input cancelled. Wrote empty output file.") // Debug message to stderr
		}
	} else {
		fmt.Fprintln(os.Stderr, "Error retrieving final model state")
		os.Exit(1)
	}
	// Exit with 0 status on success or clean cancellation
}
