package httpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const (
	pingRoute = "/ping"
)

type ReqHandlersDependencies struct {
	pingRouteResponseMessage string
}

func NewReqHandlersDependencies(pingRouteResponseMessage string) ReqHandlersDependencies {
	return ReqHandlersDependencies{
		pingRouteResponseMessage,
	}
}

type ServeReqs func(ctx context.Context, cfg Config, deps ReqHandlersDependencies) error

var _ ServeReqs = ServeReqsImpl

var RunServerImpl = func(ctx context.Context, cfg Config, serveRequests ServeReqs, deps ReqHandlersDependencies) error {
	fmt.Println(fmt.Sprintf("Starting GophersLand HTTP server listening on port: %v.", cfg.port))

	return serveRequests(ctx, cfg, deps)
}

var ServeReqsImpl = func(ctx context.Context, cfg Config, deps ReqHandlersDependencies) error {
	http.Handle(pingRoute, decorateHttpRes(pingHandlerImpl(deps.pingRouteResponseMessage), addJsonHeader()))

	server := &http.Server{Addr: fmt.Sprintf(":%d", cfg.port), Handler: nil}

	go func() {
		<-ctx.Done()
		fmt.Println("Shutting down the HTTP server...")
		server.Shutdown(ctx)
	}()

	err := server.ListenAndServe()

	// Shutting down the server is not something bad ffs Go...
	if err == http.ErrServerClosed {
		return nil
	}

	return err
}

func pingHandlerImpl(pingRouteResponseMessage string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pingReq := pingReq{}
		err := readRequest(r, &pingReq)
		if err != nil {
			writeResponse(w, pingRes{"", err.Error()}, http.StatusBadRequest)
			return
		}

		if len(pingReq.Value) == 0 {
			writeResponse(w, pingRes{"", fmt.Sprintf("ping request value must be at least 1 char")}, http.StatusBadRequest)
			return
		}

		writeResponse(w, pingRes{fmt.Sprintf("request: %s; response: %s", pingReq.Value, pingRouteResponseMessage), ""}, http.StatusOK)
	})
}

type httpResDecorator func(http.Handler) http.Handler

func decorateHttpRes(handler http.Handler, decorators ...httpResDecorator) http.Handler {
	for _, decorator := range decorators {
		handler = decorator(handler)
	}

	return handler
}

func addJsonHeader() httpResDecorator {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			handler.ServeHTTP(w, r)
		})
	}
}

func readRequest(r *http.Request, reqBody interface{}) error {
	reqBodyJson, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		return fmt.Errorf("unable to read request body. %s", err.Error())
	}

	err = json.Unmarshal(reqBodyJson, reqBody)
	if err != nil {
		return fmt.Errorf("unable to unmarshal request body. %s", err.Error())
	}

	return nil
}

func writeResponse(w http.ResponseWriter, res interface{}, statusCode int) {
	jsonRes, jsonMarshalErr := json.Marshal(res)
	if jsonMarshalErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("unable to marshal response. %s", jsonMarshalErr.Error())))
		return
	}

	w.WriteHeader(statusCode)
	w.Write(jsonRes)
	w.Write([]byte("\n"))
}
