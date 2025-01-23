// Package jobmanager provides a high-level interface for managing jobs.
// It includes a JobManager type that keeps track of all jobs and provides methods for starting, stopping, and checking the status of jobs.
// It also includes a JobStatus type that represents the status of a job, and an OutputStreamer interface for streaming output to a client.
package jobmanager

import (
	"context"
	"errors"
	"log"
	"sync"

	"github.com/google/uuid"
)

// JobStatus represents the status of a job.
type JobStatus string

const (
	JobStatusInitializing JobStatus = "initializing"
	JobStatusRunning      JobStatus = "running"
	JobStatusDone         JobStatus = "done"
	JobStatusError        JobStatus = "error"
	JobStatusNotFound     JobStatus = "not found"
	JobStatusStopped      JobStatus = "stopped"
)

var FinalStates = []JobStatus{
	JobStatusDone,
	JobStatusError,
	JobStatusStopped,
}

// String returns the string representation of a JobStatus.
//
// Returns:
//   - string: The string representation of the JobStatus.
func (js JobStatus) String() string {
	return string(js)
}

// JobManager is the high-level interface for managing jobs.
// It keeps track of all jobs and provides methods for starting, stopping, and checking the status of jobs.
//
// Fields:
// - jobs: A map of job UUIDs to Job instances.
// - jobMutex: A mutex to protect concurrent access to the job map.
type JobManager struct {
	jobs     map[string]*Job
	jobMutex sync.Mutex
}

// NewJobmanager creates a new JobManager instance.
//
// Returns:
//   - *JobManager: A new JobManager instance.
func NewJobManager() *JobManager {
	return &JobManager{
		jobs: make(map[string]*Job),
	}
}

// StartJob kicks off a job with the given command and arguments, and returns a unique identifier while the job executes in the background.
// StartJob starts a new job with the given command and arguments.
// It locks the job manager's mutex to ensure thread safety, generates a new UUID for the job,
// creates a new job instance, and starts the job with the provided context, command, and arguments.
// It returns the UUID of the newly created job and any error encountered during the process.
//
// Parameters:
//   - ctx: The context to control the job's lifecycle.
//   - command: The command to be executed by the job.
//   - args: Additional arguments for the command.
//
// Returns:
//   - uuid.UUID: The UUID of the newly created job.
//   - error: Any error encountered during the job creation or start process.
func (jm *JobManager) StartJob(ctx context.Context, command string, args ...string) (uuid.UUID, error) {
	jm.jobMutex.Lock()
	defer jm.jobMutex.Unlock()

	job := NewJob()
	jm.jobs[job.id.String()] = job

	job.Start(ctx, command, args...)

	return job.id, nil
}

// StreamOutput streams the output of a job to the provided OutputStreamer.
//
// Parameters:
//   - id: The UUID of the job to stream output from.
//   - streamer: The OutputStreamer to stream output to.
//
// Returns:
//   - error: Any error encountered during the streaming process.
func (jm *JobManager) StreamOutput(ctx context.Context, id uuid.UUID, streamer OutputStreamer) error {
	job, err := jm.getJob(id)
	if err != nil {
		return err
	}

	if job.GetStatus() != JobStatusRunning {
		return errors.New("job is not running")
	}

	if err := ctx.Err(); err != nil {
		log.Printf("Context error before starting output stream: %v\n", err)
		return err
	}

	go func() {
		streamCtx, cancel := context.WithCancel(ctx)
		defer cancel()

		log.Printf("Starting streamOutput for job %s\n", id)
		err := job.streamOutput(streamCtx, streamer)
		if err != nil {
			log.Printf("Error streaming output for job %s: %v\n", id, err)
		}

		if err := streamCtx.Err(); err != nil {
			log.Printf("Context error after streaming output: %v\n", err)
		}
	}()

	return nil
}

// StopJob stops a job with the given UUID.
//
// Parameters:
//   - ctx: The context to control the job's lifecycle.
//   - id: The UUID of the job to stop.
//
// Returns:
//   - error: Any error encountered during the stopping process.
func (jm *JobManager) StopJob(ctx context.Context, id uuid.UUID) error {
	job, err := jm.getJob(id)
	if err != nil {
		return err
	}

	if job.GetStatus() != JobStatusRunning {
		return errors.New("job is not running")
	}

	return job.Stop()
}

// GetJobStatus returns the status of a job with the given UUID.
//
// Parameters:
//   - id: The UUID of the job to get the status of.
//
// Returns:
//   - JobStatus: The status of the job.
//   - error: Any error encountered during the process.
func (jm *JobManager) GetJobStatus(id uuid.UUID) (JobStatus, error) {
	job, err := jm.getJob(id)
	if err != nil {
		return JobStatusNotFound, err
	}

	return job.GetStatus(), nil
}

// getJob retrieves a job with the given UUID, and returns an error if the job is not found.
//
// Parameters:
//   - id: The UUID of the job to retrieve.
//
// Returns:
//   - *Job: The job with the given UUID.
//   - error: Any error encountered during the process.
func (jm *JobManager) getJob(id uuid.UUID) (*Job, error) {
	jm.jobMutex.Lock()
	defer jm.jobMutex.Unlock()

	job, ok := jm.jobs[id.String()]
	if !ok {
		return nil, errors.New("job not found")
	}

	return job, nil
}
