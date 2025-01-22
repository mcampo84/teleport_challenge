package server

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"

	jobmanager "github.com/mcampo84/teleport_challenge/lib/job_manager"
	pb "github.com/mcampo84/teleport_challenge/lib/job_manager/pb/v1"
)

type Server struct {
	pb.UnimplementedCommandServiceServer

	config     Config
	grpcServer *grpc.Server
	jobManager *jobmanager.JobManager
}

func NewServer(config Config, jobManager *jobmanager.JobManager) (*Server, error) {
	s := &Server{config: config, jobManager: jobManager}

	err := s.setup()
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Server) setup() error {
	// Load server certificate and key
	cert, err := tls.LoadX509KeyPair(s.config.CertFile, s.config.KeyFile)
	if (err != nil) {
		return fmt.Errorf("failed to load server certificate: %v", err)
	}

	// Load CA certificate
	caCert, err := os.ReadFile(s.config.CaFile)
	if err != nil {
		return fmt.Errorf("failed to read CA certificate: %v", err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	// Create TLS credentials
	creds := credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientCAs:    caCertPool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
	})

	opts := []grpc.ServerOption{
		grpc.Creds(creds),
	}
	s.grpcServer = grpc.NewServer(opts...)

	pb.RegisterCommandServiceServer(s.grpcServer, s)

	return nil
}

func (s *Server) StartServer() error {
	lis, err := net.Listen("tcp", s.config.Address)
	if err != nil {
		return err
	}

	log.Printf("Server listening at %v", lis.Addr())
	return s.grpcServer.Serve(lis)
}

func (s *Server) GracefulStop() {
	s.grpcServer.GracefulStop()
}

// Below are the functions necessary to implement the gRPC server functionality



// Start implements the Start RPC method by starting a new job via the job manager library
//
// Parameters:
//  - ctx: The context of the request
//  - req: The StartRequest message containing the command and arguments
//
// Returns:
//  - The StartResponse message containing the UUID of the job
//  - An error if the job could not be started
func (s *Server) Start(ctx context.Context, req *pb.StartRequest) (*pb.StartResponse, error) {
	log.Println("Executing command:", req.Command)

	var args []string
	for _, arg := range req.Arguments {
		args = append(args, string(arg))
	}

	uid, err := s.jobManager.StartJob(ctx, string(req.Command), args...)
	if err != nil {
		// Convert the error to a include a grpc status and return it
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.StartResponse{Uuid: uid.String()}, nil
}

// Status implements the Status RPC method by retrieving the status of a job via the job manager library
//
// Parameters:
//  - ctx: The context of the request
//  - req: The StatusRequest message containing the UUID of the job
//
// Returns:
//  - The StatusResponse message containing the status of the job
//  - An error if the status could not be retrieved
func (s *Server) Status(ctx context.Context, req *pb.StatusRequest) (*pb.StatusResponse, error) {
	// Implement your command status logic here
	log.Println("Getting status for job:", req.Uuid)

	jobStatus, err := s.jobManager.GetJobStatus(req.UuidFromUuid())
	if err != nil {
		// Convert the error to a include a grpc status and return it
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.StatusResponse{Status: jobStatus.String()}, nil
}

// Stop implements the Stop RPC method by stopping a job via the job manager library
//
// Parameters:
//  - ctx: The context of the request
//  - req: The StopRequest message containing the UUID of the job
//
// Returns:
//  - The StopResponse message
//  - An error if the job could not be stopped
func (s *Server) Stop(ctx context.Context, req *pb.StopRequest) (*pb.StopResponse, error) {
	log.Println("Stopping job:", req.Uuid)

	err := s.jobManager.StopJob(ctx, req.UuidFromUuid())
	if err != nil {
		// Convert the error to a include a grpc status and return it
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.StopResponse{}, nil
}

// StreamOutput implements the StreamOutput RPC method by streaming the output of a job via the job manager library
//
// Parameters:
//  - req: The StreamOutputRequest message containing the UUID of the job
//  - streamingServer: The streaming server to stream the output to
//
// Returns:
//  - An error if the output could not be streamed
func (s *Server) StreamOutput(req *pb.StreamOutputRequest, streamingServer pb.CommandService_StreamOutputServer) error {
	log.Println("Streaming output for job:", req.Uuid)

	stream := NewStreamingServerAdapter(streamingServer)

	ctx := streamingServer.Context()
	if ctx.Err() != nil {
		log.Printf("Context error before starting stream for job %s: %v", req.Uuid, ctx.Err())
		return status.Error(codes.Canceled, "context canceled")
	}

	err := s.jobManager.StreamOutput(ctx, req.UuidFromUuid(), stream)
	if err != nil {
		log.Printf("Error streaming output for job %s: %v", req.Uuid, err)
	}

	log.Println("Successfully started streaming output for job:", req.Uuid)
	return nil
}

// Runtime check that Server implements pb.CommandServiceServer
var _ pb.CommandServiceServer = &Server{}
