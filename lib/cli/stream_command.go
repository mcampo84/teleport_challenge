package cli

import (
	"context"

	"github.com/google/uuid"

	"github.com/mcampo84/teleport_challenge/lib/client"
)

type StreamCommand struct {
	BaseCommand

	jobId uuid.UUID
}

func NewStreamCommand(client *client.Client, jobId uuid.UUID) *StreamCommand {
	return &StreamCommand{
		BaseCommand: BaseCommand{
			client: client,
		},
		jobId: jobId,
	}
}

func (c *StreamCommand) Execute(ctx context.Context) (err error) {
	err = c.client.StreamOutput(ctx, c.jobId)

	if err != nil {
		return err
	}

	return nil
}
