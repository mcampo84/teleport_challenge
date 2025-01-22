//go:generate mockgen -destination=receiver.mockgen.go -package=client -source=receiver.go

package client

import pb "github.com/mcampo84/teleport_challenge/lib/job_manager/pb/v1"

// A receiver is capable of receiving a stream of inputs and handle them.
// The primary use-case we envision is a command-line interface that can display the output of a job as it is being executed.
type Receiver interface {
	// Receive takes a string and processes it.
	Receive(*pb.StreamOutputResponse) error
}
