package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"os"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	pb "github.com/mcampo84/teleport_challenge/lib/job_manager/pb/v1"
)

type Client struct {
	client pb.CommandServiceClient
	config Config
	conn   *grpc.ClientConn
}

func NewClient(config Config) (*Client, error) {
	c := &Client{
		config: config,
	}

	if err := c.setup(); err != nil {
		return nil, err
	}
	
	return c, nil
}

func (c *Client) setup() error {
	// Load client certificate
	cert, err := tls.LoadX509KeyPair(c.config.CertFile, c.config.KeyFile)
	if err != nil {
		return fmt.Errorf("failed to load client certificate: %v", err)
	}

	// Load CA certificate
	caCert, err := os.ReadFile(c.config.CaFile)
	if err != nil {
		return fmt.Errorf("failed to read CA certificate: %v", err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	// Create TLS credentials
	creds := credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
	})

	// Connect to the server
	c.conn, err = grpc.NewClient(c.config.ServerAddress, grpc.WithTransportCredentials(creds))
	if err != nil {
		return fmt.Errorf("failed to connect to server: %v", err)
	}

	c.client = pb.NewCommandServiceClient(c.conn)

	return nil
}

func (c *Client) GracefulStop() {
	c.conn.Close()
}

// Start takes a command and arguments and starts a new job by calling the Start RPC method on the server.
func (c *Client) Start(ctx context.Context, command string, args ...string) (*pb.StartResponse, error) {
	req := &pb.StartRequest{
		Command:   command,
		Arguments: args,
	}

	return c.client.Start(ctx, req)
}

func (c *Client) Status(ctx context.Context, jobId uuid.UUID) (*pb.StatusResponse, error) {
	req := &pb.StatusRequest{
		Uuid: jobId.String(),
	}

	return c.client.Status(ctx, req)
}

// Stop takes a job UUID and stops the job by calling the Stop RPC method on the server.
func (c *Client) Stop(ctx context.Context, jobId uuid.UUID) (*pb.StopResponse, error) {
	req := &pb.StopRequest{
		Uuid: jobId.String(),
	}

	return c.client.Stop(ctx, req)
}

// StreamOutput takes a job UUID and streams the output to the receiver by calling the StreamOutput RPC method on the server.
//
// Parameters:
//  - ctx: The context of the request
//  - jobId: The UUID of the job
func (c *Client) StreamOutput(ctx context.Context, jobId uuid.UUID) error {
	req := &pb.StreamOutputRequest{
		Uuid: jobId.String(),
	}

	streamCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Send each line of output to the receiver as it is received, until the stream is closed.
	stream, err := c.client.StreamOutput(streamCtx, req)
	if err != nil {
		return err
	}

	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			fmt.Println("")
			break
		} else if err != nil {
			return err
		}

		fmt.Printf("Part %d: %s\n", resp.Part, resp.Buffer)
	}
	
	return nil
}
