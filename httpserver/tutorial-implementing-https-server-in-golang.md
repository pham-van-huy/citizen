# Tutorial: Implementing an HTTPS server in Go

In this tutorial we will:

- implement an advanced, custom HTTPS Server in Go without using any 3th party library
- create a `/ping` route with JSON request/response
- validate the request
- develop a flexible, route-granular response decorator
- secure the server with self-signed, SSL certificate
- cover the server with an integration test

Let's get started, shall we ;)

## Running an HTTP server

To launch the HTTP server running on port 9090 we will leverage the standard [net](https://golang.org/pkg/net/http/?m=all) package. No dependency on any other 3th party package!

```go
server := &http.Server{Addr: fmt.Sprintf(":%d", 9090), Handler: nil}
server.ListenAndServe()
```

## Registering a /route

Adding a `/route` is possible by registering a route handler anywhere before calling `server.ListenAndServe`.

```go
http.Handle("/ping", pingHandlerImpl())
```

## Representing route request/response with proper DTO (data transfer object)

Let's say each ping Request should contain a value and each Response should contain a message and a error, if any occurred. We want both the request and the response to be sent/received in a JSON format.

```go
type pingReq struct {
	Value string `json:"value"`
}

type pingRes struct {
	Message string `json:"message"`
	Error   string `json:"error"`
}
```

Go allows to also define a third parameter in a Struct configuring how a Struct should be serialized.

## Handling the /route

The `pingHandlerImpl()` will be a function returning a `http.Handler` as required by previously called `func Handle(pattern string, handler Handler)`.

```go
func pingHandlerImpl() http.Handler {
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

		writeResponse(w, pingRes{fmt.Sprintf("request: %s; response: %s", pingReq.Value, "some response..."), ""}, http.StatusOK)
	})
}
```

The `pingHandlerImpl()` will read a request, validate it and return an appropriate response.

## Reading an HTTP request

The `readRequest` could be implemented as:

```go
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
```

- we read the whole HTTP body, all the bytes using standard `ioutil` package
- the body will be unmarshal into our previously defined `pingReq` struct
- the method signature is generic enough to be reusable by every other additional route handler such as `/user/signup` etc

## Returning a response back to the client

```go
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
```

- a generic method so any other route handler can reuse it
- in case of a fatal error, a proper 500 status code is returned back to the client
- we allow a specific `statusCode` to be passed around in case we want to return a 400 if ping request validation determines so inside the handler, or 200 if everything went OK

## Decorating each route response in a modular manner

We said that each response should be in a JSON format.

One potential solution would be surely to write a response header in each handler like so:

```go
w.Header().Set("Content-Type", "application/json")
```

But that is not very maintainable and also doesn't help us add additional extensions if needed without modifying the code. [Open-closed principle](https://en.wikipedia.org/wiki/Open%E2%80%93closed_principle).

Also developing a nice, modular decorator system is quite a lot of fun and pretty straight forward. Pay close attention to the functions signatures!

**First,**

The HTTP package expects the following signature:

```go
func Handle(pattern string, handler Handler)
```

That we previously satisfied as:

```go
http.Handle("/ping", pingHandlerImpl())
```

Where `pingHandlerImpl()` was:

```go
func pingHandlerImpl() http.Handler {
```

**Second,**

So far so good. Now we will create a `typed func` and call it `httpResDecorator` which is basically an abstract interface but on a function level! Oh yes, Go is really cool.

```go
type httpResDecorator func(http.Handler) http.Handler
```

And a function called `decorateHttpRes` that will wrap any possible route handler as a first argument (`handler http.Handler`) and an array of our decorators as a second argument, `decorators ...httpResDecorator`.

```go
func decorateHttpRes(handler http.Handler, decorators ...httpResDecorator) http.Handler {
	for _, decorator := range decorators {
		handler = decorator(handler)
	}

	return handler
}
```

**Result,**

Given the above abstraction, we can now re-write our previous implementation and decorate responses on route level, full lego style!!!

```go
http.Handle("/ping", decorateHttpRes(pingHandlerImpl(), addJsonHeader()))
```

Or if we e.g, fancy to add a specific authorization header to each response? We can:

```go
http.Handle("/authorize", decorateHttpRes(authorizeHandlerImpl(), addJsonHeader(), addAuthorizationHeader()))
```

... without affecting any other route, or touching the existing logic.

The decorator (middleware logic) is inspired by Mat Ryer [implementation](https://medium.com/@matryer/the-http-handler-wrapper-technique-in-golang-updated-bc7fbcffa702). Recommended read.

## Securing the HTTP server with an SSL certificate

Change the current `server.ListenAndServe` to `server.ListenAndServeTLS`:

```go
err := server.ListenAndServeTLS(
    cfg.certificatePemFilePath,
    cfg.certificatePemPrivKeyFilePath,
)
```

And create a new config file to encapsulate the server properties such as port and certificate paths:

```go
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
```

**TIP: Always introduce a constructor factory function to initialize a Struct in a valid state.**

But wait a minute, what are those certificate paths? Well, an SSL certificate that secures your traffic is basically an cryptographic combination of a certificate (Public key) in combination, as usual, with a Private key.

Let's generate one right one using OpenSSL that should be pre-installed on your linux distribution so I don't count it as an 3th party tool.

Generating `localhost.crt` certificate file and its private key `localhost.key`: 

```
openssl req -x509 -out localhost.crt -keyout localhost.key \
  -newkey rsa:2048 -nodes -sha256 \
  -subj '/CN=localhost' -extensions EXT -config <( \
   printf "[dn]\nCN=localhost\n[req]\ndistinguished_name = dn\n[EXT]\nsubjectAltName=DNS:localhost\nkeyUsage=digitalSignature\nextendedKeyUsage=serverAuth")
```

[source](https://letsencrypt.org/docs/certificates-for-localhost/)

![openssl result](https://rablater.sirv.com/Images/GophersLand/Screenshot%20from%202018-10-16%2021-20-57.png)

Fill the config, compile, launch the server.

*PS: Storing the keys inside a repository is obviously a horrible idea, only suitable for localhost demonstrations.*

```go
func main() {
	cfg := httpserver.NewConfig(
		9093,
		"/home/enchanter/go/src/github.com/gophersland/citizen/httpserver/localhost.crt",
		"/home/enchanter/go/src/github.com/gophersland/citizen/httpserver/localhost.key",
	)
	reqHandlersDependencies := httpserver.NewReqHandlersDependencies("pong")

	err := httpserver.RunServerImpl(context.Background(), cfg, httpserver.ServeReqsImpl, reqHandlersDependencies)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	os.Exit(0)
}
```

Perform a request using cURL: 

```
curl -k -d '{"Value":"pinging gophersland server"}' -H "Content-Type: application/json" -X POST https://localhost:9093/ping

> {"message":"request: pinging gophersland server; response: pong","error":""}
```

Wuhu :star:!

**Note:**

We added a "**-k, --insecure (SSL)**" flag. This option explicitly allows cURL to perform "insecure" SSL connections and transfers as our certificate is self-signed and not trusted by any authority. This is fine for demonstration purposes. 

For production use, we would generate the certificate using some official public authority such as previously mentioned, Let's Encrypt and perform their DNS challenge, or buy it $$$. The DNS challenge consist from creating a DNS TXT record with specific value to validate domain ownership. In such a case, the cURL would work without any problems with the same exact code we have right now. 

The only difference would be the certificates themselves. Have a look at the official [CertBot](https://certbot.eff.org/) docs for more info on generating a self signed publicly recognised certificate supporting EVEN wildcard rules such as `*.gophersland.com`.

## Writing integration test to avoid boring, slow manual testing

Tests are amazing and save time. Anyone who says otherwise is probably an uneducated Product Owner :)

WIP. To be continued...

## Extra things to notice

- no need for creating a Server `Struct`. **Functions first, approach!**
- dependencies are clearly encapsulated and passed around without affecting the server functionality, functions signature. Even if more routes would be added and custom dependencies for each route would be needed, only one place changes, The isolated deps Struct.