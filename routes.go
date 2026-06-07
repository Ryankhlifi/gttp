package main

import (
	"errors"
	"net/http"
	"strconv"
)

type Route struct {
	children map[string]*Route
	segment  string
	end      bool
	methods  map[string]http.Handler
}

var routes = &Route{
	children: make(map[string]*Route),
	segment:  "/",
	end:      false,
	methods:  make(map[string]http.Handler),
}

func (r *Route) findRouteNode(request *http.Request) *Route {
	node := r
	path := request.URL.Path
	method := request.Method

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
		_, isNumber := strconv.Atoi(segment)

		// isNumber == nil means the conversion didn't throw an error so segment in a number
		if isNumber == nil {
			request.SetPathValue("id", segment)
			segment = "{id}"
		}
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

func (r *Route) registerRoute(method string, path string, handler http.HandlerFunc) error {
	if handler == nil {
		return errors.New("nil handler given for route " + path)
	}
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
			if node.methods[method] != nil {
				return errors.New("ambiguous path")
			}
			node.end = true
			node.methods[method] = handler
		}

	}
	return nil
}

func Handle(method string, path string, handler func(http.ResponseWriter, *http.Request)) {
	err := routes.registerRoute(method, path, handler)
	if err != nil {
		panic(err)
	}

}
