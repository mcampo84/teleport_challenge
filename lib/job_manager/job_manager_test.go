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

	// Wait up to 1 second until the job status changes to "Running"
	for i := 0; i < 10; i++ {
		jobStatus, _ := suite.jobManager.GetJobStatus(jobID)
		if jobStatus == jobmanager.JobStatusRunning {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	// Create two OutputStreamer clients
	ctrl := gomock.NewController(suite.T())
	client1 := jobmanager.NewMockOutputStreamer(ctrl)
	client2 := jobmanager.NewMockOutputStreamer(ctrl)

	// Set up expectations for the clients
	client1.EXPECT().Send(gomock.Any()).DoAndReturn(func(data []byte) error {
		fmt.Printf("Client1 received: %s", data)
		return nil
	}).AnyTimes()
	client2.EXPECT().Send(gomock.Any()).DoAndReturn(func(data []byte) error {
		fmt.Printf("Client2 received: %s", data)
		return nil
	}).AnyTimes()

	// Stream the output asynchronously
	done := make(chan bool)
	go func() {
		err := suite.jobManager.StreamOutput(jobID, client1)
		suite.Require().NoError(err)
		done <- true
	}()
	go func() {
		err := suite.jobManager.StreamOutput(jobID, client2)
		suite.Require().NoError(err)
		done <- true
	}()

	// Wait for the job to complete
	for i := 0; i < 10; i++ {
		jobStatus, _ := suite.jobManager.GetJobStatus(jobID)
		if jobStatus == jobmanager.JobStatusDone {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	// Wait for both streams to complete
	<-done
	<-done
}

func (suite *JobManagerTestSuite) TestStreamOutput_NotFound() {
	client := jobmanager.NewMockOutputStreamer(gomock.NewController(suite.T()))
	jobID := uuid.New()

	err := suite.jobManager.StreamOutput(jobID, client)
	suite.EqualError(err, "job not found")
}

func (suite *JobManagerTestSuite) TestStreamOutput_NotRunning() {
	client := jobmanager.NewMockOutputStreamer(gomock.NewController(suite.T()))
	jobID, err := suite.jobManager.StartJob(suite.ctx, "echo", "Hello, world!")
	suite.Require().NoError(err)

	// Wait for the job to complete
	time.Sleep(1 * time.Second)

	err = suite.jobManager.StreamOutput(jobID, client)
	suite.EqualError(err, "job is not running")
}
