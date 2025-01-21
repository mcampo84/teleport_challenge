package server

type Config struct {
	Address  string
	CertFile string
	KeyFile  string
	CaFile   string
}

const (
	// defaultAddress is the default address to listen on.
	defaultAddress = "localhost:50051"

	// defaultCertFile is the default path to the server's certificate file.
	defaultCertFile = "conf/server/server.crt"

	// defaultKeyFile is the default path to the server's key file.
	defaultKeyFile = "conf/server/server.key"

	// defaultCaFile is the default path to the certificate authority's certificate file.
	defaultCaFile = "conf/ca.crt"
)


var defaultConfig = NewConfig(
	defaultAddress,
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
//   - address: The address to listen on.
//   - certFile: The path to the server's certificate file.
//   - keyFile: The path to the server's key file.
//   - caFile: The path to the certificate authority's certificate file.
//
// Returns:
//   - Config: A new Config instance.
func NewConfig(address, certFile, keyFile, caFile string) Config {
	return Config{
		Address:  address,
		CertFile: certFile,
		KeyFile:  keyFile,
		CaFile:   caFile,
	}
}
