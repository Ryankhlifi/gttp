# gttp

`gttp` is a simple, lightweight, and performance-focused HTTP server framework built from scratch in Go.

It provides a minimal API for registering routes and starting an HTTP server, while keeping the internals clear and easy to understand.

## Why gttp?

The goal of `gttp` is to make HTTP server development simple while exploring how routing, request handling, response writing, and TCP-level server behavior work under the hood.

Instead of exposing a large framework API, `gttp` keeps things minimal:

```go
gttp.Handle(method, path, handler)
gttp.Listen(port)
```

## Features

- Simple route registration
- Method-based handlers
- Minimal public API
- Custom HTTP response writer
- Buffered response writing
- Basic dynamic route parameter support
- TCP connection handling
- Keep-alive connection support
- Lightweight internal routing
- Built from scratch in Go

## Installation

```bash
go get github.com/Ryankhlifi/gttp
```

## Usage

```go
package main

import (
    "net/http"

    "github.com/Ryankhlifi/gttp"
)

func main() {
    gttp.Handle(http.MethodGet, "/", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Hello from gttp"))
    })

    gttp.Handle(http.MethodGet, "/users/{id}", func(w http.ResponseWriter, r *http.Request) {
        id := r.PathValue("id")
        w.Write([]byte("User ID: " + id))
    })

    gttp.Listen("8080")
}
```

Then run the server:

```bash
go run main.go
```

Visit:

```txt
http://localhost:8080
```

## API

### `gttp.Handle(method, path, handler)`

Registers a route with an HTTP method, path, and handler function.

```go
gttp.Handle(http.MethodGet, "/hello", func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("Hello world"))
})
```

### `gttp.Listen(port)`

Starts the HTTP server on the given port.

```go
gttp.Listen("8080")
```

## Dynamic Routes

`gttp` supports simple dynamic route parameters.

```go
gttp.Handle(http.MethodGet, "/users/{id}", func(w http.ResponseWriter, r *http.Request) {
    id := r.PathValue("id")
    w.Write([]byte("User ID: " + id))
})
```

Example request:

```txt
GET /users/42
```

Response:

```txt
User ID: 42
```

## Performance

Performance is one of the main goals of `gttp`.

A basic load test was performed using `bombardier` against a simple `GET /users` route.

Benchmark command:

```powershell
PS C:\Users\Lenovo\GolandProjects\awesomeProject> & "$env:USERPROFILE\go\bin\bombardier.exe" -c 100 -n 1000000 http://localhost:8080/users
```

Benchmark result:

```txt
Bombarding http://localhost:8080/users with 1000000 request(s) using 100 connection(s)
1000000 / 1000000 [==========================================================================================================================================] 100.00% 226201/s 4s

Done!

Statistics        Avg        Stdev       Max
Reqs/sec      232295.32   20709.41   278288.30
Latency        429.09us   189.98us     26.70ms

HTTP codes:
  1xx - 0
  2xx - 1000000
  3xx - 0
  4xx - 0
  5xx - 0
  others - 0

Throughput: 44.28MB/s
```

Summary:

- Total requests: `1,000,000`
- Concurrent connections: `100`
- Average requests/sec: `232,295.32`
- Average latency: `429.09µs`
- Successful responses: `1,000,000`
- Throughput: `44.28MB/s`

These results may vary depending on hardware, operating system, Go version, and benchmark route behavior.

## Project Focus

`gttp` focuses on:

- keeping the developer API simple
- reducing unnecessary framework complexity
- improving request handling performance
- understanding HTTP server internals
- experimenting with routing and connection handling in Go

## License

This project is licensed under the Apache-2.0 License.
