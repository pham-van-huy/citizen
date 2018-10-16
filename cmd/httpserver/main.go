// Copyright 2018 https://gophersland.com
// All rights reserved.
// Use of this source code is governed by an Apache License that can be found in the LICENSE file.
package main

import (
	"context"
	"fmt"
	"github.com/gophersland/citizen/httpserver"
	"os"
)

func main() {
	cfg := httpserver.NewConfig(
		9093,
		fmt.Sprintf("%s/src/github.com/gophersland/citizen/httpserver/localhost.crt", os.Getenv("GOPATH")),
		fmt.Sprintf("%s/src/github.com/gophersland/citizen/httpserver/localhost.key", os.Getenv("GOPATH")),
	)
	reqHandlersDependencies := httpserver.NewReqHandlersDependencies("pong")

	err := httpserver.RunServerImpl(context.Background(), cfg, httpserver.ServeReqsImpl, reqHandlersDependencies)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	os.Exit(0)
}
