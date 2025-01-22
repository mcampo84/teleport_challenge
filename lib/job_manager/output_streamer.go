//go:generate mockgen -destination=output_streamer.mockgen.go -package=jobmanager -source=output_streamer.go

package jobmanager

import "context"

// OutputStreamer is an interface for streaming output to a client. 
// While we intend for this to be used with a gRPC stream, it is not strictly tied to gRPC.
type OutputStreamer interface {
	Send(ctx context.Context, output []byte) error
}
