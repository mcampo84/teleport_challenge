package jobmanager

import (
	"context"
	"errors"
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

	job.Start(ctx, command, args...)

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

	return job.Stop()
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
