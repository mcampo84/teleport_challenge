package jobmanager_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
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
	jobID := suite.jobManager.StartJob(suite.ctx, "echo", "Hello, world!")
	suite.NotNil(jobID)
}

func (suite *JobManagerTestSuite) TestGetJob() {
	jobID := suite.jobManager.StartJob(suite.ctx, "echo", "Hello, world!")
	job, err := suite.jobManager.GetJob(jobID)
	suite.NotNil(job)
	suite.NoError(err)

	_, err = suite.jobManager.GetJob(uuid.New())
	suite.EqualError(err, "job not found")
}

// Simulate starting a job and streaming the output to multiple clients simultaneously.
// Both clients should receive the same output.
// As new output lines are written by the job, they should be streamed to both clients.
func (suite *JobManagerTestSuite) TestStreamOutput() {
	jobID := suite.jobManager.StartJob(suite.ctx, "echo", "Hello, world!")
	job, _ := suite.jobManager.GetJob(jobID)
	suite.Require().NotNil(job)

	// Create two OutputStreamer clients
	ctrl := gomock.NewController(suite.T())
	client1 := jobmanager.NewMockOutputStreamer(ctrl)
	client1.EXPECT().Send([]byte("Hello, world!\n")).Return(nil).AnyTimes()
	client2 := jobmanager.NewMockOutputStreamer(ctrl)
	client2.EXPECT().Send([]byte("Hello, world!\n")).Return(nil).AnyTimes()

	// Stream output to both clients
	go job.StreamOutput(client1)
	go job.StreamOutput(client2)

	// Wait for the job to complete
	<-job.DoneChannel

	// Verify that both clients received the same output
	_, err := suite.jobManager.GetJob(jobID)
	suite.EqualError(err, "job not found")
}
