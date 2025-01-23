package cli

import (
	"context"
	"log"

	"github.com/google/uuid"
	"github.com/mcampo84/teleport_challenge/lib/client"
)

type StopCommand struct {
	BaseCommand

	jobId uuid.UUID
}

func NewStopCommand(client *client.Client, jobId uuid.UUID) *StopCommand {
	return &StopCommand{
		BaseCommand: BaseCommand{
			client: client,
		},
		jobId: jobId,
	}
}

func (c *StopCommand) Execute(ctx context.Context) (err error) {
	_, err = c.client.Stop(ctx, c.jobId)
	if err != nil {
		return err
	}

	// Print the job ID to the console.
	log.Printf("Job ID %s successfully stopped.\n", c.jobId)

	return nil
}
