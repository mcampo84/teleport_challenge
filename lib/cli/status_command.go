package cli

import (
	"context"
	"log"

	"github.com/google/uuid"
	"github.com/mcampo84/teleport_challenge/lib/client"
)

type StatusCommand struct {
	BaseCommand

	jobId uuid.UUID
}

func NewStatusCommand(client *client.Client, jobId uuid.UUID) *StatusCommand {
	return &StatusCommand{
		BaseCommand: BaseCommand{
			client: client,
		},
		jobId: jobId,
	}
}

func (c *StatusCommand) Execute(ctx context.Context) error {
	resp, err := c.client.Status(ctx, c.jobId)
	if err != nil {
		return err
	}

	// Print the job status to the console.
	log.Printf("Job ID: %s\nStatus: %s\n", c.jobId, resp.Status)

	return nil
}
