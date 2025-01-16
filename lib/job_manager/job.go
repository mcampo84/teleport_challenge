package jobmanager

import (
	"context"
	"os/exec"
	"sync"
	"syscall"
	"log"
	"io"
)

type Job struct {
	ID          string
	LogBuffer   [][]byte
	LogChannels []chan []byte
	DoneChannel chan struct{}

	status JobStatus
	mu     sync.Mutex
	cond   *sync.Cond

	cmd *exec.Cmd
}

func NewJob(id string) *Job {
	job := &Job{
		ID:          id,
		LogBuffer:   make([][]byte, 0),
		LogChannels: make([]chan []byte, 0),
		DoneChannel: make(chan struct{}),
		status:      JobStatusInitializing,
	}
	job.cond = sync.NewCond(&job.mu)
	return job
}

func (j *Job) Lock() {
	j.mu.Lock()
}

func (j *Job) Unlock() {
	j.mu.Unlock()
	j.cond.Broadcast()
}

func (j *Job) SetStatus(status JobStatus) {
	j.Lock()
	defer j.Unlock()
	
	j.status = status
}

func (j *Job) GetStatus() JobStatus {
	j.Lock()
	defer j.Unlock()

	return j.status
}

func (j *Job) Start(ctx context.Context, command string, args ...string) {
	j.cmd = exec.CommandContext(ctx, command, args...)
	stdout, err := j.cmd.StdoutPipe()
	if err != nil {
		log.Fatalf("Failed to get stdout pipe: %v", err)
	}

	if err := j.cmd.Start(); err != nil {
		log.Fatalf("Failed to start command: %v", err)
	}

	// Update job status
	j.SetStatus(JobStatusRunning)

	// Log command output
	go j.logOutput(stdout)

	// Wait for the command to finish
	if err := j.cmd.Wait(); err != nil {
		// Update Job Status
		j.SetStatus(JobStatusError)

		log.Printf("Command failed: %v", err)
	}

	j.SetDone(JobStatusDone)
}

func (j *Job) SetDone(status JobStatus) {
	// close the log channels if not already closed
	for _, ch := range j.LogChannels {
		close(ch)
	}

	j.LogBuffer = nil

	// close the done channel if not already closed, and set the job status
	select {
	case <-j.DoneChannel:
		// already closed
	default:
		close(j.DoneChannel)
		j.SetStatus(status)
	}
}

func (j *Job) Stop() error {
	j.Lock()
	defer j.Unlock()

	// Attempt to stop the command if it's running
	if j.cmd != nil && j.cmd.Process != nil {
		if err := j.cmd.Process.Signal(syscall.SIGTERM); err != nil {
			return err
		}
	}

	j.SetDone(JobStatusDone)

	return nil
}

func (j *Job) logOutput(stdout io.ReadCloser) {
	buffer := make([]byte, 1024)

	// read the output in 1024-byte chunks, and forward to the log buffer (and all waiting channels)
	for {
		select {
		case <-j.DoneChannel:
			return
		default:
			n, err := stdout.Read(buffer)
			if err != nil {
				if err == io.EOF {
					return
				}
				log.Printf("Error reading command output: %v", err)
				return
			}

			logLine := buffer[:n]
			j.Lock()

			// add lines to the LogBuffer, to support new clients requesting an output stream
			j.LogBuffer = append(j.LogBuffer, logLine)

			// for existing clients, send the new log line to the channel for streaming
			for _, ch := range j.LogChannels {
				ch <- logLine
			}
			j.Unlock()
		}
	}
}

func (j *Job) streamOutput(streamer OutputStreamer) error {
	// create & add a channel to the Job so we can stream the output as it comes
	logChannel := make(chan []byte)
	j.Lock()
	j.LogChannels = append(j.LogChannels, logChannel)

	// Stream the existing log buffer
	for _, logLine := range j.LogBuffer {
		if err := streamer.Send(logLine); err != nil {
			j.mu.Unlock()
			return err
		}
	}
	j.mu.Unlock()

	// Stream new log lines and job completion
	for {
		select {
		case logLine := <-logChannel:
			if err := streamer.Send(logLine); err != nil {
				return err
			}
		case <-j.DoneChannel:
			return nil
		}
	}
}
