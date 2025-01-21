//go:generate mockgen -destination=receiver.mockgen.go -package=client -source=receiver.go

package client

// A receiver is capable of receiving a stream of inputs and handle them.
// The primary use-case we envision is a command-line interface that can display the output of a job as it is being executed.
type Receiver interface {
	// Receive takes a string and processes it.
	Receive([]byte) error
}
