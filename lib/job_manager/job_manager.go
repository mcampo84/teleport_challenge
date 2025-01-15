package jobmanager

import (
	"context"
	"errors"
	"io"
	"log"
	"os/exec"
	"sync"

	"github.com/google/uuid"
)

type JobStatus string

const (
	JobStatusInitializing JobStatus = "initializing"
	JobStatusRunning      JobStatus = "running"
	JobStatusDone         JobStatus = "done"
	JobStatusError        JobStatus = "error"
	JobStatusNotFound     JobStatus = "not found"
)

type JobManager struct {
	jobs     map[string]*Job
	jobMutex sync.Mutex
}

func NewJobManager() *JobManager {
	return &JobManager{
		jobs: make(map[string]*Job),
	}
}

// StartJob kicks off a job with the given command and arguments, and returns a unique identifier while the job executes in the background.
func (jm *JobManager) StartJob(ctx context.Context, command string, args ...string) (uuid.UUID, error) {
	jm.jobMutex.Lock()
	defer jm.jobMutex.Unlock()

	id := uuid.New()

	job := NewJob(id.String())
	jm.jobs[id.String()] = job

	go func() {
		defer func() {
			if r := recover(); r != nil {
				job.SetStatus(JobStatusError)
				log.Printf("Recovered from panic: %v", r)
			}
		}()
		jm.start(ctx, job, command, args...)
	}()

	return id, nil
}

func (jm *JobManager) StreamOutput(id uuid.UUID, streamer OutputStreamer) error {
	job, err := jm.getJob(id)
	if err != nil {
		return err
	}

	if job.GetStatus() != JobStatusRunning {
		return errors.New("job is not running")
	}

	go job.streamOutput(streamer)

	return nil
}

func (jm *JobManager) StopJob(ctx context.Context, id uuid.UUID) error {
	job, err := jm.getJob(id)
	if err != nil {
		return err
	}

	if job.GetStatus() != JobStatusRunning {
		return errors.New("job is not running")
	}

	job.mu.Lock()
	defer job.mu.Unlock()

	// close the log channels
	for _, ch := range job.LogChannels {
		close(ch)
	}

	// close the log buffer
	job.LogBuffer = nil

	// close the done channel
	close(job.DoneChannel)

	return nil
}

func (jm *JobManager) GetJobStatus(id uuid.UUID) (JobStatus, error) {
	job, err := jm.getJob(id)
	if err != nil {
		return JobStatusNotFound, err
	}

	return job.GetStatus(), nil
}

func (jm *JobManager) getJob(id uuid.UUID) (*Job, error) {
	jm.jobMutex.Lock()
	defer jm.jobMutex.Unlock()

	job, ok := jm.jobs[id.String()]
	if !ok {
		return nil, errors.New("job not found")
	}

	return job, nil
}

// Start a job and update its status as it progresses. Send a signal to all clients when the job is done.
func (jm *JobManager) start(ctx context.Context, job *Job, command string, args ...string) {
	cmd := exec.CommandContext(ctx, command, args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatalf("Failed to get stdout pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		log.Fatalf("Failed to start command: %v", err)
	}

	// Update job status
	job.SetStatus(JobStatusRunning)

	// Log command output
	go jm.logOutput(job, stdout)

	// Wait for the command to finish
	if err := cmd.Wait(); err != nil {
		// Update Job Status
		job.SetStatus(JobStatusError)

		log.Fatalf("Command failed: %v", err)
	}

	// Update job status
	job.SetStatus(JobStatusDone)

	// Signal job completion
	close(job.DoneChannel)
}

func (jm *JobManager) logOutput(job *Job, stdout io.ReadCloser) {
	buffer := make([]byte, 1024)

	// read the output in 1024-byte chunks, and forward to the log buffer (and all waiting channels)
	for {
		n, err := stdout.Read(buffer)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatalf("Error reading command output: %v", err)
		}

		logLine := buffer[:n]
		job.mu.Lock()

		// add lines to the LogBuffer, to support new clients requesting an output stream
		job.LogBuffer = append(job.LogBuffer, logLine)

		// for existing clients, send the new log line to the channel for streaming
		for _, ch := range job.LogChannels {
			ch <- logLine
		}
		job.mu.Unlock()
	}
}
