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

type JobManager struct {
	jobs     map[string]*Job
	jobMutex sync.Mutex
}

func NewJobManager() *JobManager {
	return &JobManager{
		jobs: make(map[string]*Job),
	}
}

func (jm *JobManager) StartJob(ctx context.Context, command string, args ...string) uuid.UUID {
	jm.jobMutex.Lock()
	defer jm.jobMutex.Unlock()

	id := uuid.New()

	job := NewJob(id.String())
	jm.jobs[id.String()] = job

	go jm.start(ctx, job, command, args...)

	return id
}

func (jm *JobManager) GetJob(id uuid.UUID) (*Job, error) {
	jm.jobMutex.Lock()
	defer jm.jobMutex.Unlock()

	job, ok := jm.jobs[id.String()]; 
	if !ok {
		return nil, errors.New("job not found")
	}

	return job, nil
}

func (jm *JobManager) start(ctx context.Context, job *Job, command string, args ...string) {
	cmd := exec.CommandContext(ctx, command, args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatalf("Failed to get stdout pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		log.Fatalf("Failed to start command: %v", err)
	}

	// Log command output
	go jm.logOutput(job, stdout)

	// Wait for the command to finish
	if err := cmd.Wait(); err != nil {
		log.Fatalf("Command failed: %v", err)
	}

	// Simulate job completion
	close(job.DoneChannel)

	jm.jobMutex.Lock()
	defer jm.jobMutex.Unlock()

	delete(jm.jobs, job.ID)
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
