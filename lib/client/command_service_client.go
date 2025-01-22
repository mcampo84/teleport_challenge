package client

// This file contains the implementation of the client for the command service.

import (
	"context"
	"fmt"
	"io"

	"github.com/google/uuid"
	pb "github.com/mcampo84/teleport_challenge/lib/job_manager/pb/v1"
)

// Start takes a command and arguments and starts a new job by calling the Start RPC method on the server.
func (c *Client) Start(ctx context.Context, command string, args ...string) (*pb.StartResponse, error) {
	req := &pb.StartRequest{
		Command:   command,
		Arguments: args,
	}

	return c.client.Start(ctx, req)
}

func (c *Client) Status(ctx context.Context, jobId uuid.UUID) (*pb.StatusResponse, error) {
	req := &pb.StatusRequest{
		Uuid: jobId.String(),
	}

	return c.client.Status(ctx, req)
}

// Stop takes a job UUID and stops the job by calling the Stop RPC method on the server.
func (c *Client) Stop(ctx context.Context, jobId uuid.UUID) (*pb.StopResponse, error) {
	req := &pb.StopRequest{
		Uuid: jobId.String(),
	}

	return c.client.Stop(ctx, req)
}

// StreamOutput takes a job UUID and streams the output to the receiver by calling the StreamOutput RPC method on the server.
//
// Parameters:
//  - ctx: The context of the request
//  - jobId: The UUID of the job
func (c *Client) StreamOutput(ctx context.Context, jobId uuid.UUID) error {
	req := &pb.StreamOutputRequest{
		Uuid: jobId.String(),
	}

	// Send each line of output to the receiver as it is received, until the stream is closed.
	stream, err := c.client.StreamOutput(ctx, req)
	if err != nil {
		return err
	}

	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			fmt.Println("")
			break
		} else if err != nil {
			return err
		}

		fmt.Printf("Part %d: %s\n", resp.Part, resp.Buffer)
	}
	
	return nil
}

