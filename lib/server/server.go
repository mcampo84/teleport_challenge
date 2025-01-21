package server

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net"
	"os"

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
	if err != nil {
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

	opts := []grpc.ServerOption{grpc.Creds(creds)}
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
