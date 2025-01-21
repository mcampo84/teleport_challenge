package client

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	pb "github.com/mcampo84/teleport_challenge/lib/job_manager/pb/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
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
