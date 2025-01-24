package server_test

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"google.golang.org/grpc"
	"github.com/stretchr/testify/suite"

	"github.com/mcampo84/teleport_challenge/lib/job_manager"
	pb "github.com/mcampo84/teleport_challenge/lib/job_manager/pb/v1"
	"github.com/mcampo84/teleport_challenge/lib/server"
)

type ServerTestSuite struct {
	suite.Suite
	
	jobManager *jobmanager.JobManager
	server *server.Server
}

func (suite *ServerTestSuite) SetupSuite() {
	suite.jobManager = jobmanager.NewJobManager()

	var err error
	suite.server, err = server.NewServer(server.GetTestConfig(), suite.jobManager)
	suite.Require().NoError(err)
}

func TestServerTestSuite(t *testing.T) {
	suite.Run(t, new(ServerTestSuite))
}

func (suite *ServerTestSuite) TestStreamOutput() {
	ctx := context.Background()

	msg := "Hello, world!"
	// Start a job
	jobID, _ := suite.jobManager.StartJob(ctx, "./test_fixtures/test_script.sh", msg)

	// Stream the output
	req := &pb.StreamOutputRequest{Uuid: jobID.String()}

	stream := NewMockStreamOutputServer(suite)

	// wait for the job to start running
	for {
		if s, _ := suite.jobManager.GetJobStatus(jobID); s == jobmanager.JobStatusRunning {
			break;
		}
	}

	err := suite.server.StreamOutput(req, stream)
	suite.NoError(err)

	// wait for the job to finish
	for {
		if s, _ := suite.jobManager.GetJobStatus(jobID); s == jobmanager.JobStatusDone {
			break;
		}
	}
	stream.Validate(msg)
}

type mockStreamOutputServer struct {
	grpc.ServerStream

	responses []byte
	s *ServerTestSuite
}

func NewMockStreamOutputServer(s *ServerTestSuite) *mockStreamOutputServer {
	return &mockStreamOutputServer{
		responses: []byte{},
		s: s,
	}
}

func (m *mockStreamOutputServer) Send(response *pb.StreamOutputResponse) error {
	// Implement the Send method
	m.responses = append(m.responses, response.Buffer...)
	return nil
}

func (m *mockStreamOutputServer) Context() context.Context {
	// Implement the Context method
	return context.TODO()
}

func (m *mockStreamOutputServer) Validate(msg string) {
	lines := bytes.Split(m.responses, []byte("\n"))
	m.s.Len(lines, 6) // 5 lines of output + 1 empty line at the end
	for i := 0; i < 5; i++ {
		m.s.Equal(fmt.Sprintf("Line %d: %s", i+1, msg), string(lines[i]))
	}
}
