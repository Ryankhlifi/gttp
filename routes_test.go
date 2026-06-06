package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRegisterRoute(t *testing.T) {
	root := &Route{
		children: make(map[string]*Route),
		segment:  "/",
		methods:  make(map[string]http.Handler),
	}

	handler := func(w http.ResponseWriter, r *http.Request) {}

	err := root.registerRoute(http.MethodGet, "/hello/world", handler)
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}

	// ambiguous route
	err = root.registerRoute(http.MethodGet, "/hello/world", handler)
	if err == nil {
		t.Error("Expected ambiguous path error, got nil")
	}

	// nil handler
	err = root.registerRoute(http.MethodGet, "/hello/nil", nil)
	if err == nil {
		t.Error("Expected nil handler error, got nil")
	}
}

func TestFindRouteNode(t *testing.T) {
	root := &Route{
		children: make(map[string]*Route),
		segment:  "/",
		methods:  make(map[string]http.Handler),
	}

	handler := func(w http.ResponseWriter, r *http.Request) {}
	root.registerRoute(http.MethodGet, "/hello", handler)
	root.registerRoute(http.MethodPost, "/users/{id}", handler)

	// test exact match
	req1 := httptest.NewRequest(http.MethodGet, "/hello", nil)
	node1 := root.findRouteNode(req1)
	if node1 == nil {
		t.Error("Expected route node, got nil")
	}

	// test parameter match
	req2 := httptest.NewRequest(http.MethodPost, "/users/123", nil)
	node2 := root.findRouteNode(req2)
	if node2 == nil {
		t.Error("Expected route node for parametrized path, got nil")
	} else if req2.PathValue("id") != "123" {
		t.Errorf("Expected PathValue 'id' to be 123, got %s", req2.PathValue("id"))
	}

	// test method not found
	req3 := httptest.NewRequest(http.MethodPost, "/hello", nil)
	node3 := root.findRouteNode(req3)
	if node3 != nil {
		t.Error("Expected nil node for wrong method, got non-nil")
	}

	// test path not found
	req4 := httptest.NewRequest(http.MethodGet, "/missing", nil)
	node4 := root.findRouteNode(req4)
	if node4 != nil {
		t.Error("Expected nil node for missing path, got non-nil")
	}
}

func TestHandle(t *testing.T) {
	root := &Route{
		children: make(map[string]*Route),
		segment:  "/",
		methods:  make(map[string]http.Handler),
	}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Handle panicked: %v", r)
		}
	}()

	root.Handle(http.MethodGet, "/test", func(w http.ResponseWriter, r *http.Request) {})
}

func TestHandlePanic(t *testing.T) {
	root := &Route{
		children: make(map[string]*Route),
		segment:  "/",
		methods:  make(map[string]http.Handler),
	}

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected Handle to panic on nil handler")
		}
	}()

	root.Handle(http.MethodGet, "/test", nil)
}
