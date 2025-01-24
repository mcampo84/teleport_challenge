//go:generate mockgen -destination=output_streamer.mockgen.go -package=jobmanager -source=output_streamer.go

package jobmanager

import pb "github.com/mcampo84/teleport_challenge/lib/job_manager/pb/v1"

// OutputStreamer is an interface for streaming output to a client.
// While we intend for this to be used with a gRPC stream, it is not strictly tied to gRPC.
type OutputStreamer interface {
	Send(output *pb.StreamOutputResponse) error
}
