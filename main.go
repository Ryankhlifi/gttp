package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
)

type ResponseWriter struct {
	conn    net.Conn
	headers http.Header
	status  int
}

var routes = &Route{
	children: nil,
	handler:  nil,
	segment:  "/test",
	end:      false,
	method:   http.MethodGet,
}

func (rw *ResponseWriter) Header() http.Header {
	return rw.headers
}

func (rw *ResponseWriter) WriteHeader(status int) {
	rw.status = status
	_, err := fmt.Fprintf(rw.conn, "HTTP/1.1 %d %s\r\n", status, http.StatusText(status))
	if err != nil {
		fmt.Println("Error setting status code :", err)
		return
	}
	err = rw.headers.Write(rw.conn)
	if err != nil {
		fmt.Println("Error writing header :", err)
		return
	}

	_, err = fmt.Fprintf(rw.conn, "\r\n")
	if err != nil {
		fmt.Println("Error seperating headers from body :", err)
		return
	}

}

func (rw *ResponseWriter) Write(b []byte) (int, error) {
	if rw.status == 0 {
		rw.WriteHeader(http.StatusOK)
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
		headers: make(http.Header),
	}
}

func handleRequest(conn net.Conn) {

	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			fmt.Println("Error closing connection")
		}
	}(conn)

	reader := bufio.NewReader(conn)
	request, err := http.ReadRequest(reader)

	if err != nil {
		if err == io.EOF {
			fmt.Println("EOF , closing connection")
			return
		}

		fmt.Println("Error reading request: " + err.Error())
		return
	}

	//figure out how to invoke the function upon an http request
	node, err := routes.findRouteNode(request.Method, request.URL.Path)
	if err != nil {
		fmt.Println("Error finding route : " + err.Error())
		return
	}

	writer := makeResponseWriter(conn)

	node.handler.ServeHTTP(writer, request)

}

func main() {

	routes.Handle("GET", "/test", testHandler)

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

	}

}

// a simple test handler to send a JSON payload
func testHandler(w http.ResponseWriter, r *http.Request) {
	payload := `{"message":"hello world"}`
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(len(payload)))
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(payload))

}
