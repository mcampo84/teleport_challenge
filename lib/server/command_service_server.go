package server

import (
	"context"
	"log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/mcampo84/teleport_challenge/lib/job_manager/pb/v1"
)

func (s *Server) Start(ctx context.Context, req *pb.StartRequest) (*pb.StartResponse, error) {
	log.Println("Executing command:", req.Command)

	uid, err := s.jobManager.StartJob(ctx, req.Command, req.Arguments...)
	if err != nil {
		// Convert the error to a include a grpc status and return it
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.StartResponse{Uuid: uid.String()}, nil
}

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

func (s *Server) Stop(ctx context.Context, req *pb.StopRequest) (*pb.StopResponse, error) {
	log.Println("Stopping job:", req.Uuid)

	err := s.jobManager.StopJob(ctx, req.UuidFromUuid())
	if err != nil {
		// Convert the error to a include a grpc status and return it
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.StopResponse{}, nil
}

func (s *Server) StreamOutput(req *pb.StreamOutputRequest, streamingServer pb.CommandService_StreamOutputServer) error {
	log.Println("Streaming output for job:", req.Uuid)

	stream := NewStreamingServerAdapter(streamingServer)

	err := s.jobManager.StreamOutput(req.UuidFromUuid(), stream)
	if err != nil {
		// Convert the error to a include a grpc status and return it
		return status.Error(codes.Internal, err.Error())
	}

	return nil
}
