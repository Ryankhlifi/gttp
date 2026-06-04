package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"time"
)

type ResponseWriter struct {
	conn    net.Conn
	headers http.Header
	status  int
}

type RequestError struct {
	Message    string
	StatusCode int
}

func (e *RequestError) Error() string {
	return e.Message
}

var routes = &Route{
	children: make(map[string]*Route),
	segment:  "/",
	end:      false,
	methods:  make(map[string]http.Handler),
}

func (rw *ResponseWriter) Header() http.Header {
	return rw.headers
}

func (rw *ResponseWriter) WriteHeader(status int) {
	rw.Header().Set("Content-Type", "application/json")
	rw.Header().Set("Connection", "Keep-Alive")
	rw.status = status
	_, err := fmt.Fprintf(rw.conn, "HTTP/1.1 %d %s\r\n", status, http.StatusText(status))

	if err != nil {
		fmt.Println("Error setting status code :", err)
		return
	}

}

func (rw *ResponseWriter) Write(b []byte) (int, error) {
	if b != nil {
		bodyLength := strconv.Itoa(len(b))
		rw.Header().Add("Content-Length", bodyLength)
		err := rw.Header().Write(rw.conn)
		if err != nil {
			fmt.Println("Error writing content length header :", err)
			return 0, err
		}
	}
	_, err := fmt.Fprintf(rw.conn, "\r\n")
	if err != nil {
		fmt.Println("Error seperating headers from body :", err)
		return 0, err
	}

	n, err := rw.conn.Write(b)
	if err != nil {
		fmt.Println("Error writing body :", err)
		return 0, err
	}
	return n, nil
}

func makeResponseWriter(conn net.Conn) *ResponseWriter {
	return &ResponseWriter{
		conn:    conn,
		headers: http.Header{},
	}
}

func handleRequest(conn net.Conn) {

	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			fmt.Println("Error closing connection : ", err)
		}
	}(conn)

	reader := bufio.NewReader(conn)
	for {

		err := conn.SetReadDeadline(time.Now().Add(30 * time.Second))
		if err != nil {
			fmt.Println("Error setting read deadline :", err)
			return
		}

		writer := makeResponseWriter(conn)
		request, err := http.ReadRequest(reader)

		if err != nil {
			if err == io.EOF {
				fmt.Println("client closed connection , closing connection")
			} else {
				fmt.Println("Error reading request: " + err.Error())
			}
			return
		}

		node := routes.findRouteNode(request.Method, request.URL.Path)
		if node == nil {
			sendError(writer, &RequestError{"route not found", http.StatusNotFound})
			continue
		}

		node.methods[request.Method].ServeHTTP(writer, request)
	}
}

func main() {

	routes.Handle("POST", "/test", postTestHandler)
	routes.Handle("GET", "/test", getTestHandler)

	if len(os.Args) != 2 {
		panic("unspecified port or too many arguments. Usage : go run main.go [PORT]")
	}

	port := ":" + os.Args[1]
	listener, err := net.Listen("tcp", port)
	if err != nil {
		panic(err.Error())
	}
	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {
			fmt.Println("Error closing listener: " + err.Error())
		}
	}(listener)

	fmt.Println("Listening on " + listener.Addr().String())
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: " + err.Error())
			continue
		}

		go handleRequest(conn)

		//runtime memory benchmark

		go func() {
			ticker := time.NewTicker(1 * time.Second)
			defer ticker.Stop()

			for range ticker.C {
				memoryBenchmark()
			}
		}()

	}

}

// a simple test handler to send a JSON payload
func getTestHandler(w http.ResponseWriter, _ *http.Request) {
	payload := `{"message":" /test , methods is GET"}`
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(payload))
	if err != nil {
		return
	}

}

func postTestHandler(w http.ResponseWriter, _ *http.Request) {
	payload := `{"message":" /test , methods is POST"}`
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(payload))
	if err != nil {
		return
	}

}

func sendError(w http.ResponseWriter, requestError *RequestError) {
	payload := `{"message":"` + requestError.Message + `"}`
	w.WriteHeader(requestError.StatusCode)
	_, err := w.Write([]byte(payload))
	if err != nil {
		fmt.Println("Error writing body :", err)
		return
	}
}

func memoryBenchmark() {
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)
	fmt.Printf(
		"Alloc = %v MiB | TotalAlloc = %v MiB | Sys = %v MiB | NumGC = %v\n",
		bToMb(stats.Alloc),
		bToMb(stats.TotalAlloc),
		bToMb(stats.Sys),
		stats.NumGC,
	)
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}
