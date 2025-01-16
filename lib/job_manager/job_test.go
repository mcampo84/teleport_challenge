package jobmanager_test

import (
	"context"
	"testing"
	"time"

	jobmanager "github.com/mcampo84/teleport_challenge/lib/job_manager"
	"github.com/stretchr/testify/suite"
)

type JobTestSuite struct {
	suite.Suite

	ctx context.Context
}

func (suite *JobTestSuite) SetupTest() {
	suite.ctx = context.Background()
}

func TestJob(t *testing.T) {
	suite.Run(t, new(JobTestSuite))
}

func (suite *JobTestSuite) TestStart() {
	job := jobmanager.NewJob("test")
	job.Start(suite.ctx, "sleep", "1")
	suite.Equal(jobmanager.JobStatusInitializing, job.GetStatus())
}

func (suite *JobTestSuite) TestStop() {
	job := jobmanager.NewJob("test")
	job.Start(suite.ctx, "sleep", "10")
	time.Sleep(100 * time.Millisecond) // Give the job some time to start

	err := job.Stop()
	suite.NoError(err)
	suite.Equal(jobmanager.JobStatusDone, job.GetStatus())
}

func (suite *JobTestSuite) TestGetStatus() {
	job := jobmanager.NewJob("test")
	job.Start(suite.ctx, "sleep", "2")
	suite.Equal(jobmanager.JobStatusInitializing, job.GetStatus())

	time.Sleep(time.Second) // Wait for the job to start
	suite.Equal(jobmanager.JobStatusRunning, job.GetStatus())

	time.Sleep(2 * time.Second) // Wait for the job to complete
	suite.Equal(jobmanager.JobStatusDone, job.GetStatus())
}