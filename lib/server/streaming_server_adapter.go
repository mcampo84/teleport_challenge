package server

import (
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

func (s *StreamingServerAdapter) Send(output []byte) error {
	s.part++

	fmt.Printf("Received %s as part %d\n", output, s.part)

	return s.streamingServer.Send(&pb.StreamOutputResponse{
		Part: s.part,
		Buffer: output,
	})
}
