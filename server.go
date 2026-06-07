package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"time"
)

type ResponseWriter struct {
	conn        net.Conn
	writer      *bufio.Writer
	headers     http.Header
	status      int
	wroteHeader bool
}

type RequestError struct {
	Message    string
	StatusCode int
}

func (e *RequestError) Error() string {
	return e.Message
}

func (rw *ResponseWriter) flushHeaders() {
	if rw.wroteHeader {
		return
	}

	if rw.status == 0 {
		rw.status = http.StatusOK
	}

	if rw.Header().Get("Content-Type") == "" {
		rw.Header().Set("Content-Type", "application/json")
	}
	if rw.Header().Get("Connection") == "" {
		rw.Header().Set("Connection", "Keep-Alive")
	}

	// default to 0 length if no length is set
	if rw.Header().Get("Content-Length") == "" && rw.Header().Get("Transfer-Encoding") == "" {
		rw.Header().Set("Content-Length", "0")
	}

	_, err := fmt.Fprintf(rw.writer, "HTTP/1.1 %d %s\r\n", rw.status, http.StatusText(rw.status))
	if err != nil {
		fmt.Println("Error setting status code :", err)
		return
	}

	err = rw.Header().Write(rw.writer)
	if err != nil {
		fmt.Println("Error writing headers:", err)
		return
	}

	_, err = rw.writer.WriteString("\r\n")
	if err != nil {
		fmt.Println("Error separating headers from body:", err)
	}

	rw.wroteHeader = true
}

func (rw *ResponseWriter) Header() http.Header {
	return rw.headers
}

func (rw *ResponseWriter) WriteHeader(status int) {
	rw.status = status
}

func (rw *ResponseWriter) Write(b []byte) (int, error) {
	if !rw.wroteHeader {
		if rw.Header().Get("Content-Length") == "" {
			rw.Header().Set("Content-Length", strconv.Itoa(len(b)))
		}
		rw.flushHeaders()
	}

	if b == nil || len(b) == 0 {
		return 0, nil
	}

	n, err := rw.writer.Write(b)
	if err != nil {
		fmt.Println("Error writing body :", err)
		return n, err
	}

	return n, nil
}

func makeResponseWriter(conn net.Conn, bw *bufio.Writer) *ResponseWriter {
	return &ResponseWriter{
		conn:        conn,
		writer:      bw,
		headers:     http.Header{},
		wroteHeader: false,
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
	bw := bufio.NewWriter(conn)
	for {

		err := conn.SetReadDeadline(time.Now().Add(30 * time.Second))
		if err != nil {
			fmt.Println("Error setting read deadline :", err)
			return
		}

		writer := makeResponseWriter(conn, bw)
		request, err := http.ReadRequest(reader)

		if err != nil {
			if err != io.EOF {
				go fmt.Println("Error reading request: " + err.Error())
			}
			return
		}

		node := routes.findRouteNode(request)
		if node == nil {
			sendError(writer, &RequestError{"route not found", http.StatusNotFound})
			writer.flushHeaders() // ensure headers are written
			err := bw.Flush()
			if err != nil {
				fmt.Println("Error flushing buffer :", err)
				return
			}
			continue
		}

		node.methods[request.Method].ServeHTTP(writer, request)
		writer.flushHeaders() // ensure headers are written even if Write() wasn't called
		err = bw.Flush()
		if err != nil {
			fmt.Println("Error flushing buffer :", err)
			return
		}
	}
}

func listen(port string) {

	if port == "" {
		panic("You must provide a port to listen on.")
	}

	port = ":" + port
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

func sendError(w http.ResponseWriter, requestError *RequestError) {
	payload := `{"message":"` + requestError.Message + `"}`
	w.WriteHeader(requestError.StatusCode)
	_, err := w.Write([]byte(payload))
	if err != nil {
		fmt.Println("Error writing body :", err)
		return
	}
}
