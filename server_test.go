package gttp

import (
	"bufio"
	"bytes"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"
)

// mockConn implements net.Conn for testing
type mockConn struct {
	readBuf  *bytes.Buffer
	writeBuf *bytes.Buffer
}

func (m *mockConn) Read(b []byte) (n int, err error)   { return m.readBuf.Read(b) }
func (m *mockConn) Write(b []byte) (n int, err error)  { return m.writeBuf.Write(b) }
func (m *mockConn) Close() error                       { return nil }
func (m *mockConn) LocalAddr() net.Addr                { return nil }
func (m *mockConn) RemoteAddr() net.Addr               { return nil }
func (m *mockConn) SetDeadline(t time.Time) error      { return nil }
func (m *mockConn) SetReadDeadline(t time.Time) error  { return nil }
func (m *mockConn) SetWriteDeadline(t time.Time) error { return nil }

func TestResponseWriter(t *testing.T) {
	writeBuf := new(bytes.Buffer)
	conn := &mockConn{writeBuf: writeBuf}
	bw := bufio.NewWriter(conn)

	rw := makeResponseWriter(conn, bw)

	// test Header
	rw.Header().Set("X-Custom-Header", "TestValue")

	// test WriteHeader
	rw.WriteHeader(http.StatusCreated)

	// test Write
	body := []byte("hello world")
	n, err := rw.Write(body)
	if err != nil {
		t.Fatalf("Write returned error: %v", err)
	}
	if n != len(body) {
		t.Errorf("Expected to write %d bytes, wrote %d", len(body), n)
	}

	// flush the bufio.Writer to push everything to the mockConn
	err = bw.Flush()
	if err != nil {
		t.Fatalf("Flush error: %v", err)
	}

	output := writeBuf.String()

	// check headers and status
	if !strings.Contains(output, "HTTP/1.1 201 Created") {
		t.Errorf("Expected status '201 Created' in output, got:\n%s", output)
	}
	if !strings.Contains(output, "X-Custom-Header: TestValue") {
		t.Errorf("Expected 'X-Custom-Header' in output, got:\n%s", output)
	}
	if !strings.Contains(output, "Content-Length: 11") {
		t.Errorf("Expected 'Content-Length: 11' in output, got:\n%s", output)
	}

	// check body
	if !strings.HasSuffix(output, "hello world") {
		t.Errorf("Expected body 'hello world' at the end of output, got:\n%s", output)
	}
}

func TestSendError(t *testing.T) {
	writeBuf := new(bytes.Buffer)
	conn := &mockConn{writeBuf: writeBuf}
	bw := bufio.NewWriter(conn)
	rw := makeResponseWriter(conn, bw)

	reqErr := &RequestError{
		Message:    "Not Found",
		StatusCode: http.StatusNotFound,
	}

	if reqErr.Error() != "Not Found" {
		t.Errorf("Expected Error() to return 'Not Found', got '%s'", reqErr.Error())
	}

	sendError(rw, reqErr)
	rw.flushHeaders()
	bw.Flush()

	output := writeBuf.String()

	if !strings.Contains(output, "HTTP/1.1 404 Not Found") {
		t.Errorf("Expected 404 status in output, got:\n%s", output)
	}
	if !strings.Contains(output, `{"message":"Not Found"}`) {
		t.Errorf("Expected JSON error message in output, got:\n%s", output)
	}
}

func TestResponseWriter_EmptyWrite(t *testing.T) {
	writeBuf := new(bytes.Buffer)
	conn := &mockConn{writeBuf: writeBuf}
	bw := bufio.NewWriter(conn)

	rw := makeResponseWriter(conn, bw)
	rw.WriteHeader(http.StatusOK)

	n, err := rw.Write([]byte{})
	if err != nil {
		t.Fatalf("Expected nil error on empty write, got: %v", err)
	}
	if n != 0 {
		t.Errorf("Expected to write 0 bytes, wrote %d", n)
	}
}
