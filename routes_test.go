package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// helper to create a fresh router for each test
func newRouter() *Route {
	return &Route{
		children: make(map[string]*Route),
	}
}

// dummy handler for registration tests
func dummyHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// helper to build a handler that writes a specific body, useful to verify
// the *correct* handler was matched (not just any handler)
func handlerWithBody(body string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(body))
	}
}

// --- Handle / registration ---

func TestHandle_RegistersRoute(t *testing.T) {
	r := newRouter()
	r.Handle("GET", "/users", dummyHandler)

	node, err := r.findRouteNode("GET", "/users")
	if err != nil {
		t.Fatalf("expected route to be found, got error: %v", err)
	}
	if node == nil {
		t.Fatal("expected non-nil node")
	}
}

func TestHandle_PanicsOnDuplicateRoute(t *testing.T) {
	r := newRouter()
	r.Handle("GET", "/users", dummyHandler)

	defer func() {
		if rec := recover(); rec == nil {
			t.Error("expected panic on duplicate route, got none")
		}
	}()

	r.Handle("GET", "/users", dummyHandler)
}

func TestHandle_PanicsOnNilHandler(t *testing.T) {
	r := newRouter()

	defer func() {
		if rec := recover(); rec == nil {
			t.Error("expected panic on nil handler, got none")
		}
	}()

	r.Handle("GET", "/users", nil)
}

func TestHandle_SamePathDifferentMethods(t *testing.T) {
	r := newRouter()
	// These should not panic — different methods on the same path are valid
	r.Handle("GET", "/users", dummyHandler)
	r.Handle("POST", "/users", dummyHandler)

	_, errGet := r.findRouteNode("GET", "/users")
	_, errPost := r.findRouteNode("POST", "/users")

	if errGet != nil {
		t.Errorf("GET /users should be found, got: %v", errGet)
	}
	if errPost != nil {
		t.Errorf("POST /users should be found, got: %v", errPost)
	}
}

// --- findRouteNode ---

func TestFindRouteNode_NotFound(t *testing.T) {
	r := newRouter()

	_, err := r.findRouteNode("GET", "/nonexistent")
	if err == nil {
		t.Fatal("expected error for unregistered route, got nil")
	}
}

func TestFindRouteNode_WrongMethod(t *testing.T) {
	r := newRouter()
	r.Handle("GET", "/users", dummyHandler)

	_, err := r.findRouteNode("POST", "/users")
	if err == nil {
		t.Fatal("expected error for wrong methods, got nil")
	}
}

func TestFindRouteNode_NestedPath(t *testing.T) {
	r := newRouter()
	r.Handle("GET", "/api/v1/users", dummyHandler)

	node, err := r.findRouteNode("GET", "/api/v1/users")
	if err != nil {
		t.Fatalf("expected nested route to be found, got: %v", err)
	}
	if node == nil {
		t.Fatal("expected non-nil node")
	}
}

func TestFindRouteNode_PartialPathNotFound(t *testing.T) {
	r := newRouter()
	r.Handle("GET", "/api/v1/users", dummyHandler)

	// /api/v1 exists as an intermediate node but has no handler/methods
	_, err := r.findRouteNode("GET", "/api/v1")
	if err == nil {
		t.Fatal("expected error for partial path, got nil")
	}
}

// --- Handler invocation ---

func TestHandler_IsCalledOnMatch(t *testing.T) {
	r := newRouter()
	r.Handle("GET", "/hello", handlerWithBody("hello world"))

	node, err := r.findRouteNode("GET", "/hello")
	if err != nil {
		t.Fatalf("route not found: %v", err)
	}

	req := httptest.NewRequest("GET", "/hello", nil)
	rec := httptest.NewRecorder()
	node.methods["GET"].ServeHTTP(rec, req)

	if rec.Body.String() != "hello world" {
		t.Errorf("expected 'hello world', got '%s'", rec.Body.String())
	}
}

func TestHandler_CorrectHandlerMatchedAmongMultiple(t *testing.T) {
	r := newRouter()
	r.Handle("GET", "/foo", handlerWithBody("foo"))
	r.Handle("GET", "/bar", handlerWithBody("bar"))

	node, err := r.findRouteNode("GET", "/bar")
	if err != nil {
		t.Fatalf("route not found: %v", err)
	}

	req := httptest.NewRequest("GET", "/bar", nil)
	rec := httptest.NewRecorder()
	node.methods["GET"].ServeHTTP(rec, req)

	if rec.Body.String() != "bar" {
		t.Errorf("expected 'bar', got '%s'", rec.Body.String())
	}
}

// --- registerRoute internals ---

func TestRegisterRoute_SetsEndFlag(t *testing.T) {
	r := newRouter()
	r.Handle("GET", "/users", dummyHandler)

	node, err := r.findRouteNode("GET", "/users")
	if err != nil {
		t.Fatalf("route not found: %v", err)
	}
	if !node.end {
		t.Error("expected end flag to be true on terminal node")
	}
}

func TestRegisterRoute_SetsSegment(t *testing.T) {
	r := newRouter()
	r.Handle("GET", "/users", dummyHandler)

	node, err := r.findRouteNode("GET", "/users")
	if err != nil {
		t.Fatalf("route not found: %v", err)
	}
	if node.segment != "users" {
		t.Errorf("expected segment 'users', got '%s'", node.segment)
	}
}

func TestRegisterRoute_SharedPrefixRoutes(t *testing.T) {
	r := newRouter()
	r.Handle("GET", "/api/users", handlerWithBody("users"))
	r.Handle("GET", "/api/posts", handlerWithBody("posts"))

	for _, tc := range []struct {
		path string
		want string
	}{
		{"/api/users", "users"},
		{"/api/posts", "posts"},
	} {
		node, err := r.findRouteNode("GET", tc.path)
		if err != nil {
			t.Errorf("route %s not found: %v", tc.path, err)
			continue
		}
		req := httptest.NewRequest("GET", tc.path, nil)
		rec := httptest.NewRecorder()
		node.methods["GET"].ServeHTTP(rec, req)
		if rec.Body.String() != tc.want {
			t.Errorf("path %s: expected '%s', got '%s'", tc.path, tc.want, rec.Body.String())
		}
	}
}
