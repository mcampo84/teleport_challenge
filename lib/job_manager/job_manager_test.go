package jobmanager_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	jobmanager "github.com/mcampo84/teleport_challenge/lib/job_manager"
	"github.com/stretchr/testify/suite"
)

type JobManagerTestSuite struct {
	suite.Suite

	ctx        context.Context
	jobManager *jobmanager.JobManager
}

func (suite *JobManagerTestSuite) SetupSuite() {
	suite.jobManager = jobmanager.NewJobManager()
}

func (suite *JobManagerTestSuite) SetupTest() {
	suite.ctx = context.Background()
}

func TestJobManager(t *testing.T) {
	suite.Run(t, new(JobManagerTestSuite))
}

func (suite *JobManagerTestSuite) TestStartJob() {
	jobID, _ := suite.jobManager.StartJob(suite.ctx, "echo", "Hello, world!")
	suite.NotNil(jobID)
}

// Simulate starting a job and streaming the output to multiple clients simultaneously.
// Both clients should receive the same output.
// As new output lines are written by the job, they should be streamed to both clients.
func (suite *JobManagerTestSuite) TestStreamOutput() {
	// Start ./test_fixtures/test_script.sh with the argument "Hello, world!"
	jobID, err := suite.jobManager.StartJob(suite.ctx, "./test_fixtures/test_script.sh", "Hello, world!")
	suite.Require().NoError(err)

	// Create two OutputStreamer clients
	ctrl := gomock.NewController(suite.T())
	client1 := jobmanager.NewMockOutputStreamer(ctrl)
	client2 := jobmanager.NewMockOutputStreamer(ctrl)
	for i := 0; i < 5; i++ {
		client1.EXPECT().Send([]byte(fmt.Sprintf("Line %d: Hello, world!\n", i))).Return(nil).AnyTimes()
		client2.EXPECT().Send([]byte(fmt.Sprintf("Line %d: Hello, world!\n", i))).Return(nil).AnyTimes()
	}

	// Wait up to 1 second until the job status changes to "Running"
	for i := 0; i < 10; i++ {
		jobStatus, _ := suite.jobManager.GetJobStatus(jobID)
		if jobStatus == jobmanager.JobStatusRunning {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	// Stream output to both clients
	err = suite.jobManager.StreamOutput(jobID, client1)
	suite.Require().NoError(err)
	err = suite.jobManager.StreamOutput(jobID, client2)
	suite.Require().NoError(err)	
}

// Stop a job that is currently running.
// The job should stop streaming output to clients.
func (suite *JobManagerTestSuite) TestStopJob() {
	jobID, _ := suite.jobManager.StartJob(suite.ctx, "./test_fixtures/test_script.sh", "Hello, world!")

	// Wait up to 1 second until the job status changes to "Running"
	for i := 0; i < 10; i++ {
		jobStatus, _ := suite.jobManager.GetJobStatus(jobID)
		if jobStatus == jobmanager.JobStatusRunning {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	// Stop the job
	err := suite.jobManager.StopJob(suite.ctx, jobID)
	suite.Require().NoError(err)
}
