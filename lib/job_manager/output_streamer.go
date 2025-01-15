//go:generate mockgen -destination=output_streamer.mockgen.go -package=jobmanager -source=output_streamer.go

package jobmanager

type OutputStreamer interface {
	Send(output []byte) error
}
