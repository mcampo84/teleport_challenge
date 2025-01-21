package cli

import (
	"context"
	"fmt"

	"github.com/mcampo84/teleport_challenge/lib/client"
)

type StartCommand struct {
	command string
	args    []string

	client *client.Client
}

// NewStart
func NewStartCommand(client *client.Client, command string, args ...string) *StartCommand {
	return &StartCommand{
		client: client,
		command: command,
		args:    args,
	}
}

func (c *StartCommand) Execute(ctx context.Context) error {
	resp, err := c.client.Start(ctx, c.command, c.args...)
	if err != nil {
		return err
	}

	// Print the job ID to the console.
	fmt.Printf("Job ID %s started\n", resp.Uuid)

	return nil
}
