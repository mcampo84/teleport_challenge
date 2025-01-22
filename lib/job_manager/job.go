package jobmanager

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os/exec"
	"sync"
	"syscall"
	
	"github.com/google/uuid"
)

// Job represents a job that can be managed by the job manager.
// It contains information about the job's ID, log buffers, log channels,
// and channels to signal when the job is done. It also includes synchronization
// primitives to manage concurrent access to the job's status and command execution.
//
// Fields:
// - ID: A unique identifier for the job.
// - LogBuffer: A buffer to store log entries as byte slices.
// - LogChannels: A list of channels to send log entries to.
// - DoneChannel: A channel to signal when the job is completed.
// - status: The current status of the job.
// - mu: A mutex to protect access to the job's status.
// - cond: A condition variable to signal changes in the job's status.
// - wg: A wait group to wait for all goroutines associated with the job to complete.
// - cmd: The command to be executed by the job.
type Job struct {
	id          uuid.UUID
	logBuffer   []byte
	logChannels []chan []byte
	doneChannel chan struct{}
	status      JobStatus
	mu          sync.Mutex
	cond        *sync.Cond
	wg          sync.WaitGroup
	cmd         *exec.Cmd
}

// NewJob creates a new Job with the given ID and initializes its fields.
//
// Returns:
//   - *Job: A new Job instance.
func NewJob() *Job {
	id := uuid.New()
	job := &Job{
		id:          id,
		logBuffer:   make([]byte, 0),
		logChannels: make([]chan []byte, 0),
		doneChannel: make(chan struct{}),
		status:      JobStatusInitializing,
	}
	job.cond = sync.NewCond(&job.mu)
	return job
}

// setStatus sets the status of the job.
//
// Parameters:
//   - status: The status to set for the job.
func (j *Job) setStatus(status JobStatus) {
	j.mu.Lock()
	
	defer func() {
		j.mu.Unlock()
		j.cond.Broadcast()
	}()
	j.status = status
}

// GetStatus returns the current status of the job.
//
// Returns:
//   - JobStatus: The current status of the job.
func (j *Job) GetStatus() JobStatus {
	j.mu.Lock()
	defer j.mu.Unlock()

	return j.status
}

// Start begins the execution of the job with the specified command and arguments.
// It logs the command output and updates the job status accordingly.
//
// Parameters:
//   - ctx: The context to control the job's lifecycle.
//   - command: The command to be executed
//   - args: Additional arguments for the command
func (j *Job) Start(ctx context.Context, command string, args ...string) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				j.setDone(JobStatusError)
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
		j.setStatus(JobStatusRunning)

		// Log command output
		log.Print("Logging command output")
		j.wg.Add(1)
		go j.logOutput(stdout)

		// Wait for the command to finish
		if err := j.cmd.Wait(); err != nil {
			// Update Job Status
			log.Print("Setting job status to error")
			j.setStatus(JobStatusError)
			log.Printf("Command failed: %v", err)
		}

		j.mu.Lock()
		defer func() {
			j.mu.Unlock()
			j.cond.Broadcast()
		}()

		log.Print("Closing job")
		j.setDone(JobStatusDone)
	}()
}

// Stop attempts to stop the job's command if it is running and waits for all goroutines to complete.
// It closes the log and done channels to stop logging and streaming of logs.
//
// Returns:
//   - error: Any error encountered during the stopping process.
func (j *Job) Stop() error {
	j.mu.Lock()
	defer func() {
		j.mu.Unlock()
		j.cond.Broadcast()
	}()

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

	// Closing channels to stop logging and streaming of logs
	j.setDone(JobStatusDone)

	return nil
}

// setDone marks the job as done, then closes the log and done channels.
//
// Parameters:
//   - status: The status to set for the job.
func (j *Job) setDone(status JobStatus) {
	// close the done channel if not already closed, and set the job status
	select {
	case <-j.doneChannel:
		// already closed
	default:
		close(j.doneChannel)

		// do not call j.SetStatus() here. SetStatus() will lock the mutex, which is already locked.
		j.status = status
	}
}

// logOutput reads the command's stdout and forwards the output to the log buffer and channels.
//
// Parameters:
//   - stdout: The stdout pipe of the command.
func (j *Job) logOutput(stdout io.ReadCloser) {
	defer j.wg.Done()
	buffer := make([]byte, 1024)

	// read the output in 1024-byte chunks, and forward to the log buffer (and all waiting channels)
	for {
		n, err := stdout.Read(buffer)
		if err != nil {
			// if EOF is reached, we can close the log channels
			if err == io.EOF {
				for _, ch := range j.logChannels {
					close(ch)
				}

				return
			}
			log.Printf("Error reading command output: %v", err)
			return
		}

		logLine := buffer[:n]
		fmt.Println(string(logLine))

		j.mu.Lock()
		// add lines to the LogBuffer, to support new clients requesting an output stream
		j.logBuffer = append(j.logBuffer, logLine...)

		// for existing clients, send the new log line to the channel for streaming
		for _, ch := range j.logChannels {
			select {
			case ch <- logLine:
			default:
				log.Printf("Channel is full, skipping log line")
			}
		}
		j.mu.Unlock()
		j.cond.Broadcast()
	}
}

// streamOutput streams the job's output to the provided OutputStreamer.
//
// Parameters:
//   - streamer: The OutputStreamer to stream output to.
//
// Returns:
//   - error: Any error encountered during the streaming process.
func (j *Job) streamOutput(streamer OutputStreamer) error {
	if j.status != JobStatusRunning {
		return errors.New("job is not running")
	}

	// create & add a channel to the Job so we can stream the output as it comes
	logChannel := make(chan []byte)
	j.mu.Lock()
	j.logChannels = append(j.logChannels, logChannel)
	var logBuffer []byte 
	copy(logBuffer, j.logBuffer)
	j.mu.Unlock()

	// Stream the existing log buffer
	if err := streamer.Send(logBuffer); err != nil {
		return err
	}

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
		case <-j.doneChannel:
			 // Drain the logChannel before returning
			 for logLine := range logChannel {
				if err := streamer.Send(logLine); err != nil {
					return err
				}
			}
			return nil
		}
	}
}
