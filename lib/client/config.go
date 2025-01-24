package client

type Config struct {
	CaFile        string
	CertFile      string
	KeyFile       string
	ServerAddress string
}

const (
	// defaultServerAddress is the default address of the server.
	defaultServerAddress = "localhost:50051"

	// defaultCertFile is the default path to the client's certificate file.
	defaultCertFile = "conf/client/client.crt"
	
	// defaultKeyFile is the default path to the client's key file.
	defaultKeyFile = "conf/client/client.key"

	// defaultCaFile is the default path to the certificate authority's certificate file.
	defaultCaFile = "conf/ca.crt"
)

var defaultConfig = NewConfig(
	defaultServerAddress,
	defaultCertFile,
	defaultKeyFile,
	defaultCaFile,
)

func GetDefaultConfig() Config {
	return defaultConfig
}

// NewConfig returns a new Config with the given values.
//
// Parameters:
//   - serverAddress: The address of the server.
//   - certFile: The path to the client's certificate file.
//   - keyFile: The path to the client's key file.
//   - caFile: The path to the certificate authority's certificate file.
//
// Returns:
//   - Config: A new Config instance.
func NewConfig(serverAddress, certFile, keyFile, caFile string) Config {
	return Config{
		ServerAddress: serverAddress,
		CertFile:      certFile,
		KeyFile:       keyFile,
		CaFile:        caFile,
	}
}
