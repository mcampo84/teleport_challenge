// Package cli provides a CLI interface for executing commands remotely via the job manager service.
package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/mcampo84/teleport_challenge/lib/client"
)

type CLI struct {
	client *client.Client
}

func NewCLI(jobManagerClient *client.Client) *CLI {
	return &CLI{client: jobManagerClient}
}

// Start initializes the CLI, outputs a helpful usage message, then awaits input. There should be no context timeout.
func (c *CLI) Start(ctx context.Context) error {
	fmt.Println("Welcome to the Job Manager CLI")
	fmt.Println("Available commands:")
	fmt.Println("  - start [command] [args...]")
	fmt.Println("  - stop [job UUID]")
	fmt.Println("  - status [job UUID]")
	fmt.Println("  - stream [job UUID]")

	c.awaitInput(ctx)

	return nil
}

// awaitInput waits for the user to input a command, with arguments following.
// Examples of commands are:
//   - start [command] [args...]
//   - stop [job UUID]
//   - status [job UUID]
//   - stream [job UUID]
func (c *CLI) awaitInput(ctx context.Context) {
	reader := bufio.NewReader(os.Stdin)

	for {
		// Read the user input.
		fmt.Print("> ")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading command:", err)
			continue
		}

		// Trim the newline character and split the input into command and arguments.
		input = strings.TrimSpace(input)
		parts := strings.Fields(input)
		if len(parts) == 0 {
			fmt.Println("Error: no command provided")
			continue
		}

		cmd, err := NewCommand(c.client, parts)
		if err != nil {
			fmt.Printf("Error: %s\n", err.Error())
			continue
		}

		err = cmd.Execute(ctx)
		if err != nil {
			fmt.Printf("Error: %s\n", err.Error())
			continue
		}
	}
}
