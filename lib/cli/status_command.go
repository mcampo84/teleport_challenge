package cli

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/mcampo84/teleport_challenge/lib/client"
)

type StatusCommand struct {
	jobId  uuid.UUID
	client *client.Client
}

func NewStatusCommand(client *client.Client, jobId uuid.UUID) *StatusCommand {
	return &StatusCommand{
		client: client,
		jobId: jobId,
	}
}

func (c *StatusCommand) Execute(ctx context.Context) error {
	resp, err := c.client.Status(ctx, c.jobId)
	if err != nil {
		return err
	}

	// Print the job status to the console.
	fmt.Printf("Job ID: %s\nStatus: %s\n", c.jobId, resp.Status)

	return nil
}
