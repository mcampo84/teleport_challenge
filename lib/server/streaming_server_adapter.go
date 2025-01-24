package server

import (
	"context"
	"fmt"

	pb "github.com/mcampo84/teleport_challenge/lib/job_manager/pb/v1"
)

type StreamingServerAdapter struct {
	part            int32
	streamingServer pb.CommandService_StreamOutputServer
}

func NewStreamingServerAdapter(streamingServer pb.CommandService_StreamOutputServer) *StreamingServerAdapter {
	return &StreamingServerAdapter{part: 0, streamingServer: streamingServer}
}

func (s *StreamingServerAdapter) Send(ctx context.Context, output []byte) error {
	s.part++

	fmt.Printf("Received %s as part %d\n", output, s.part)

	if ctx.Err() != nil {
		fmt.Printf("Context error before sending part %d: %v\n", s.part, ctx.Err())
		return ctx.Err()
	}

	err := s.streamingServer.Send(&pb.StreamOutputResponse{
		Part: s.part,
		Buffer: output,
	})
	if err != nil {
		fmt.Printf("Error sending part %d: %v\n", s.part, err)
		return err
	}

	return nil
}
