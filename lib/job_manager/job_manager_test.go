package jobmanager_test

import (
	"context"
	"fmt"
	"testing"
	"time"

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
	jobID, err := suite.jobManager.StartJob(suite.ctx, "echo", "Hello, world!")
	suite.Require().NoError(err)
	suite.NotNil(jobID)
}

func (suite *JobManagerTestSuite) TestGetJobStatus() {
	// Start ./test_fixtures/test_script.sh with the argument "Hello, world!"
	jobID, err := suite.jobManager.StartJob(suite.ctx, "./test_fixtures/test_script.sh", "Hello, world!")
	suite.Require().NoError(err)

	// Wait up to 1 second until the job status changes to "Running"
	for i := 0; i < 10; i++ {
		jobStatus, _ := suite.jobManager.GetJobStatus(jobID)
		if jobStatus == jobmanager.JobStatusRunning {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	status, err := suite.jobManager.GetJobStatus(jobID)
	suite.Require().NoError(err)
	suite.Equal(jobmanager.JobStatusRunning, status)
}

func (suite *JobManagerTestSuite) TestGetJobStatus_NotFound() {
	_, err := suite.jobManager.GetJobStatus(uuid.New())
	suite.EqualError(err, "job not found")
}

func (suite *JobManagerTestSuite) TestStopJob() {
	// Start ./test_fixtures/test_script.sh with the argument "Hello, world!"
	jobID, err := suite.jobManager.StartJob(suite.ctx, "./test_fixtures/test_script.sh", "Hello, world!")
	suite.Require().NoError(err)

	// Wait up to 1 second until the job status changes to "Running"
	for i := 0; i < 10; i++ {
		jobStatus, _ := suite.jobManager.GetJobStatus(jobID)
		if jobStatus == jobmanager.JobStatusRunning {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	// Stop the job
	err = suite.jobManager.StopJob(suite.ctx, jobID)
	suite.Require().NoError(err)

	// Verify the job status is "Done"
	status, err := suite.jobManager.GetJobStatus(jobID)
	suite.Require().NoError(err)
	suite.Equal(jobmanager.JobStatusDone, status)
}

func (suite *JobManagerTestSuite) TestStopJob_NotFound() {
	err := suite.jobManager.StopJob(suite.ctx, uuid.New())
	suite.EqualError(err, "job not found")
}

func (suite *JobManagerTestSuite) TestStopJob_NotRunning() {
	jobID, err := suite.jobManager.StartJob(suite.ctx, "echo", "Hello, world!")
	suite.Require().NoError(err)

	// Wait for the job to complete
	time.Sleep(1 * time.Second)

	// Try to stop the job
	err = suite.jobManager.StopJob(suite.ctx, jobID)
	suite.EqualError(err, "job is not running")
}

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

func (suite *JobManagerTestSuite) TestStreamOutput_NotFound() {
	err := suite.jobManager.StreamOutput(uuid.New(), nil)
	suite.EqualError(err, "job not found")
}

func (suite *JobManagerTestSuite) TestStreamOutput_NotRunning() {
	jobID, err := suite.jobManager.StartJob(suite.ctx, "echo", "Hello, world!")
	suite.Require().NoError(err)

	// Wait for the job to complete
	time.Sleep(1 * time.Second)

	// Try to stream the output
	err = suite.jobManager.StreamOutput(jobID, nil)
	suite.EqualError(err, "job is not running")
}
