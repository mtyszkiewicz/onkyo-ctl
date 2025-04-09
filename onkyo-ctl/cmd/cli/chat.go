// chat.go
package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/chzyer/readline"
	"github.com/mtyszkiewicz/eiscp/internal/pkg/eiscp"
)

// StartChatSession initiates an interactive chat session with the Onkyo device
func StartChatSession(client *eiscp.EISCPClient) error {
	fmt.Println("Chat session with Onkyo TX-L20D established.")
	fmt.Println("Type EISCP commands or 'exit' to quit.")
	fmt.Println("Use Ctrl+C or Ctrl+D to terminate the session.")
	fmt.Println("Use arrow up/down to navigate command history.")

	// Setup readline with history support
	rl, err := readline.New("> ")
	if err != nil {
		return fmt.Errorf("failed to initialize readline: %w", err)
	}
	defer rl.Close()

	// Setup signal handling for Ctrl+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\nTerminating chat session...")
		rl.Close()
		os.Exit(0)
	}()

	// Start the chat loop
	for {
		line, err := rl.Readline()
		if err != nil {
			// Handle Ctrl+D or EOF
			fmt.Println("\nTerminating chat session...")
			break
		}

		input := strings.TrimSpace(line)
		if input == "" {
			continue
		}

		if strings.ToLower(input) == "exit" {
			fmt.Println("Terminating chat session...")
			break
		}

		// Add command to history
		rl.SaveHistory(input)

		// Send the command to the Onkyo device
		response, err := client.SendReceiveCommand(input)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		// Format and display the response
		fmt.Printf("TX-L20D: %s\n", response)
	}

	return nil
}
