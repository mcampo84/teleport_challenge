package cli

import (
	"context"
	"log"

	"github.com/mcampo84/teleport_challenge/lib/client"
)

type StartCommand struct {
	BaseCommand

	command string
	args    []string
}

// NewStart
func NewStartCommand(client *client.Client, command string, args ...string) *StartCommand {
	return &StartCommand{
		BaseCommand: BaseCommand{
			client: client,
		},
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
	log.Printf("Job ID %s started\n", resp.Uuid)

	return nil
}
