package main

import (
	"fmt"
	"net/http"
	"strings"
)

type Route struct {
	children map[string]*Route
	handler  http.Handler
	segment  string
	end      bool
	method   string
}

// Handle : this registers a routes and adds a handler/**
func (r *Route) Handle(method string, path string, handler func(http.ResponseWriter, *http.Request)) {
	node, _ := r.findRouteNode(method, path)
	if node != nil && node.method == method {
		panic("Ambigious path: " + path)
	}
	if handler == nil {
		panic("Nil handler given for " + path)
	}
	if node != nil {
		node.handler = http.HandlerFunc(handler)
	}

	r.registerRoute(method, path, handler)

}
func (r *Route) findRouteNode(method string, path string) (*Route, error) {
	node := r
	for _, seg := range strings.Split(strings.Trim(path, "/"), "/") {
		child, ok := node.children[seg]
		if !ok {
			return nil, fmt.Errorf("404 not found")
		}
		node = child
	}

	if node.method != method {
		return nil, fmt.Errorf("method %s not allowed", node.method)
	}

	return node, nil
}

func (r *Route) registerRoute(method string, path string, handler http.HandlerFunc) {
	segments := strings.Split(strings.Trim(path, "/"), "/")
	node := r

	for index, segment := range segments {

		if _, ok := node.children[segment]; ok {
			node = node.children[segment]
			if node.handler == nil {
				node.handler = handler
			}
			continue
		}

		node.children = make(map[string]*Route)
		node.children[segment] = &Route{
			children: make(map[string]*Route),
			segment:  segment,
		}

		node = node.children[segment]
		if index == len(segments)-1 {
			node.end = true
			node.handler = handler
			node.method = method
		}

	}

}
