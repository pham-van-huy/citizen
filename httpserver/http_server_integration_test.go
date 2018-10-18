package httpserver

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestHttpServerLifeCycle(t *testing.T) {
	ctx, closeServer := context.WithCancel(context.Background())
	cfg := NewConfig(
		9093,
		fmt.Sprintf("%s/src/github.com/gophersland/citizen/httpserver/localhost.crt", os.Getenv("GOPATH")),
		fmt.Sprintf("%s/src/github.com/gophersland/citizen/httpserver/localhost.key", os.Getenv("GOPATH")),
	)

	go func() {
		reqHandlersDependencies := NewReqHandlersDependencies("test pong")
		err := RunServerImpl(ctx, cfg, ServeReqsImpl, reqHandlersDependencies)
		if err != nil {
			t.Fatal(err)
		}
	}()

	time.Sleep(time.Second * 2)

	req, err := http.NewRequest("POST", createURL(cfg, pingRoute), createPingReq())
	if err != nil {
		closeServer()
		t.Fatal(err)
	}

	resp, err := newHttpClient().Do(req)
	if err != nil {
		closeServer()
		t.Fatal(err)
	}
	defer resp.Body.Close()

	var pingRes pingRes
	body, _ := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(body, &pingRes)
	if err != nil {
		closeServer()
		t.Fatal(err)
	}

	if len(pingRes.Error) != 0 {
		closeServer()
		t.Fatal(pingRes.Error)
	}

	if len(pingRes.Message) == 0 {
		closeServer()
		t.Fatal("returned response is not suppose to be empty")
	}

	if resp.StatusCode != http.StatusOK {
		closeServer()
		t.Fatalf("returned response code '%v' is not as expected one '%v'", resp.StatusCode, http.StatusOK)
	}

	if resp.Header.Get("Content-Type") != "application/json" {
		closeServer()
		t.Fatalf("returned response header '%v' is not '%v'", resp.Header.Get("Content-Type"), "application/json")
	}

	closeServer()
}

func createPingReq() *bytes.Reader {
	reqBodyJson, _ := json.Marshal(pingReq{"test ping value"})
	return bytes.NewReader(reqBodyJson)
}

func createURL(cfg Config, route string) string {
	return fmt.Sprintf("https://%s:%d%s", "localhost", cfg.port, route)
}

func newHttpClient() *http.Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	return &http.Client{Transport: tr}
}
