package cli

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/mcampo84/teleport_challenge/lib/client"
)

type StreamCommand struct {
	jobId  uuid.UUID
	client *client.Client
}

func NewStreamCommand(client *client.Client, jobId uuid.UUID) *StreamCommand {
	return &StreamCommand{
		client: client,
		jobId:  jobId,
	}
}

func (c *StreamCommand) Execute(ctx context.Context) (err error) {
	err = c.client.StreamOutput(ctx, c.jobId, c)
	if err != nil {
		return err
	}

	return nil
}

func (c *StreamCommand) Receive(output []byte) error {
	fmt.Println(output)

	return nil
}
