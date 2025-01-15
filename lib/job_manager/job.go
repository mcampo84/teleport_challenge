package jobmanager

import "sync"

type Job struct {
	ID          string
	LogBuffer   [][]byte
	LogChannels []chan []byte
	DoneChannel chan struct{}

	status JobStatus
	mu     sync.Mutex
	cond   *sync.Cond
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

func (j *Job) SetStatus(status JobStatus) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.status = status
	j.cond.Broadcast()
}

func (j *Job) GetStatus() JobStatus {
	j.mu.Lock()
	defer j.mu.Unlock()

	return j.status
}

func (j *Job) streamOutput(streamer OutputStreamer) error {
	// create & add a channel to the Job so we can stream the output as it comes
	logChannel := make(chan []byte)
	j.mu.Lock()
	j.LogChannels = append(j.LogChannels, logChannel)
	j.mu.Unlock()

	// Stream the existing log buffer
	j.mu.Lock()
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
