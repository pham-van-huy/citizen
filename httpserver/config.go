package httpserver

type Config struct {
	port                          int
	certificatePemFilePath        string
	certificatePemPrivKeyFilePath string
}

func NewConfig(port int, certificatePemFilePath string, certificatePemPrivKeyFilePath string) Config {
	return Config{
		port,
		certificatePemFilePath,
		certificatePemPrivKeyFilePath,
	}
}
