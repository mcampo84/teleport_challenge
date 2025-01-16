package jobmanager

import (
	"context"
	"errors"
	"io"
	"log"
	"os/exec"
	"sync"
	"syscall"
)

type Job struct {
	ID          string
	LogBuffer   [][]byte
	LogChannels []chan []byte
	DoneChannel chan struct{}

	status JobStatus
	mu     sync.Mutex
	cond   *sync.Cond
	wg     sync.WaitGroup

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
	go func() {
		defer func() {
			if r := recover(); r != nil {
				j.SetDone(JobStatusError)
			}
		}()
		j.cmd = exec.CommandContext(ctx, command, args...)
		stdout, err := j.cmd.StdoutPipe()
		if err != nil {
			log.Fatalf("Failed to get stdout pipe: %v", err)
		}

		if err := j.cmd.Start(); err != nil {
			log.Fatalf("Failed to start command: %v", err)
		}

		// Update job status
		log.Print("Setting job status to running")
		j.SetStatus(JobStatusRunning)

		// Log command output
		log.Print("Logging command output")
		j.wg.Add(1)
		go j.logOutput(stdout)

		// Wait for the command to finish
		if err := j.cmd.Wait(); err != nil {
			// Update Job Status
			log.Print("Setting job status to error")
			j.SetStatus(JobStatusError)

			log.Printf("Command failed: %v", err)
		}

		j.Lock()
		defer j.Unlock()

		log.Print("Closing job")
		j.SetDone(JobStatusDone)
	}()
}

func (j *Job) SetDone(status JobStatus) {
	// close the log channels if not already closed
	for _, ch := range j.LogChannels {
		close(ch)
	}

	// close the done channel if not already closed, and set the job status
	select {
	case <-j.DoneChannel:
		// already closed
	default:
		close(j.DoneChannel)
		// do not call j.SetStatus() here. SetStatus() will lock the mutex, which is already locked.
		j.status = status
	}

	j.LogBuffer = nil
}

func (j *Job) Stop() error {
	j.Lock()
	defer j.Unlock()

	// Closing channels to stop logging and streaming of logs
	j.SetDone(JobStatusDone)

	// Attempt to stop the command if it's running
	if j.cmd != nil && j.cmd.Process != nil {
		log.Printf("Sending SIGTERM to process %d", j.cmd.Process.Pid)
		if err := j.cmd.Process.Signal(syscall.SIGTERM); err != nil {
			log.Printf("Failed to send SIGTERM: %v", err)
			return err
		}
	}

	// Wait for all goroutines to complete
	j.wg.Wait()

	return nil
}

func (j *Job) logOutput(stdout io.ReadCloser) {
	defer j.wg.Done()
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
	if j.status != JobStatusRunning {
		return errors.New("job is not running")
	}

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
		case logLine, ok := <-logChannel:
			if !ok {
				return nil
			}
			if err := streamer.Send(logLine); err != nil {
				return err
			}
		case <-j.DoneChannel:
			return nil
		}
	}
}
