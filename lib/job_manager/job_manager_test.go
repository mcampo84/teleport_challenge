package jobmanager_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	jobmanager "github.com/mcampo84/teleport_challenge/lib/job_manager"
	pb "github.com/mcampo84/teleport_challenge/lib/job_manager/pb/v1"
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
	suite.Equal(jobmanager.JobStatusStopped, status)
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

type mockStreamOutputServer struct {
	pb.CommandService_StreamOutputServer
	received []*pb.StreamOutputResponse
}

func (m *mockStreamOutputServer) Send(resp *pb.StreamOutputResponse) error {
	m.received = append(m.received, resp)
	fmt.Printf("Received: %v\n", resp)
	return nil
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
	streamCount := 20

	// Stream the output asynchronously across 20 clients
	done := make(chan bool)
	for i := 0; i < streamCount; i++ {
		client := &mockStreamOutputServer{}

		go func(c pb.CommandService_StreamOutputServer, streamID int) {
			err := suite.jobManager.StreamOutput(jobID, c)
			suite.Require().NoError(err, "Client %d", streamID)
			done <- true
		}(client, i)
	}

	// Wait for the job to complete
	for i := 0; i < 10; i++ {
		jobStatus, _ := suite.jobManager.GetJobStatus(jobID)
		if jobStatus == jobmanager.JobStatusDone {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	// Wait for both streams to complete
	for i := 0; i < streamCount; i++ {
		<-done
	}
}

func (suite *JobManagerTestSuite) TestStreamOutput_NotFound() {
	client := &mockStreamOutputServer{}
	jobID := uuid.New()

	err := suite.jobManager.StreamOutput(jobID, client)
	suite.EqualError(err, "job not found")
}

func (suite *JobManagerTestSuite) TestStreamOutput_NotRunning() {
	client := &mockStreamOutputServer{}
	jobID, err := suite.jobManager.StartJob(suite.ctx, "echo", "Hello, world!")
	suite.Require().NoError(err)

	// Wait for the job to complete
	time.Sleep(1 * time.Second)

	err = suite.jobManager.StreamOutput(jobID, client)
	suite.NoError(err)
}
