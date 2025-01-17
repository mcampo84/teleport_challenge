package server

import (
	"log"
	"net"

	jobmanager "github.com/mcampo84/teleport_challenge/lib/job_manager"
	pb "github.com/mcampo84/teleport_challenge/lib/job_manager/pb/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type Server struct {
	pb.UnimplementedCommandServiceServer

	config     Config
	grpcServer *grpc.Server
	jobManager *jobmanager.JobManager
}

func NewServer(config Config, jobManager *jobmanager.JobManager) *Server {
	return &Server{config: config, jobManager: jobManager}
}

func (s *Server) Listen() error {
	creds, err := credentials.NewServerTLSFromFile(s.config.CertFile, s.config.KeyFile)
	if err != nil {
		return err
	}

	opts := []grpc.ServerOption{grpc.Creds(creds)}
	s.grpcServer = grpc.NewServer(opts...)

	pb.RegisterCommandServiceServer(s.grpcServer, s)

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
