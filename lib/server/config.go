package server

type Config struct {
	Address  string
	CertFile string
	KeyFile  string
	CaFile   string
}

func NewConfig(address, certFile, keyFile, caFile string) *Config {
	return &Config{
		Address:  address,
		CertFile: certFile,
		KeyFile:  keyFile,
		CaFile:   caFile,
	}
}
