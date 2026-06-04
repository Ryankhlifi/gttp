package main

import (
	"net/http"
)

type Route struct {
	children map[string]*Route
	segment  string
	end      bool
	methods  map[string]http.Handler
}

// Handle : this registers a routes and adds a handler function/**
func (r *Route) Handle(method string, path string, handler func(http.ResponseWriter, *http.Request)) {
	handlerFunc := http.HandlerFunc(handler)
	node := r.findRouteNode(method, path)
	if node != nil && node.methods[method] != nil {
		panic("Ambigious path: " + path)
	}
	if handler == nil {
		panic("Nil handler given for " + path)
	}
	if node != nil {
		node.methods[method] = handlerFunc
		return
	}

	r.registerRoute(method, path, handler)

}
func (r *Route) findRouteNode(method string, path string) *Route {
	node := r

	var start int
	var segment string
	i := 0

	for i < len(path) {

		for i < len(path) && path[i] == '/' {
			i++
		}

		if i > len(path) {
			break
		}

		start = i

		for i < len(path) && path[i] != '/' {
			i++
		}

		segment = path[start:i]

		child, ok := node.children[segment]
		if !ok {
			return nil
		}
		node = child
	}

	if node.methods[method] == nil {
		return nil
	}

	return node
}

func (r *Route) registerRoute(method string, path string, handler http.HandlerFunc) {
	node := r

	var start int
	var segment string
	i := 0
	for i < len(path) {

		for i < len(path) && path[i] == '/' {
			i++
		}

		if i >= len(path) {
			break
		}

		start = i

		for i < len(path) && path[i] != '/' {
			i++
		}

		segment = path[start:i]
		_, ok := node.children[segment]
		if !ok {
			node.children[segment] = &Route{
				children: make(map[string]*Route),
				segment:  segment,
				methods:  make(map[string]http.Handler),
			}
		}

		node = node.children[segment]
		if i == len(path) {
			node.end = true
			node.methods[method] = handler
		}

	}

}
