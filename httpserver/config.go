// Copyright 2018 https://gophersland.com
// All rights reserved.
// Use of this source code is governed by an Apache License that can be found in the LICENSE file.
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
