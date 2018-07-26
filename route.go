package main

import (
	"net/http"
	"regexp"
)

type Request struct {
	*http.Request
}

func (r *Request) GetParams(route route) map[string]string {
	match := route.pattern.FindStringSubmatch(r.URL.Path)
	params := make(map[string]string)
	for i, name := range route.pattern.SubexpNames() {
		if i != 0 {
			params[name] = match[i]
		}
	}

	return params
}

type HandlerFunc func(http.ResponseWriter, *http.Request, map[string]string)

func (f HandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request, p map[string]string) {
	f(w, r, p)
}

type Handler interface {
	ServeHTTP(http.ResponseWriter, *http.Request, map[string]string)
}

// http://stackoverflow.com/questions/6564558/
// wildcards-in-the-pattern-for-http-handlefunc
type route struct {
	pattern *regexp.Regexp
	handler Handler
}

type RegexpHandler struct {
	routes []*route
}

func (h *RegexpHandler) Handler(pattern string, handler Handler) {
	h.routes = append(h.routes, &route{regexp.MustCompile(pattern), handler})
}

func (h *RegexpHandler) HandleFunc(
	pattern string,
	handler func(http.ResponseWriter, *http.Request, map[string]string),
) {
	h.routes = append(
		h.routes,
		&route{
			regexp.MustCompile(pattern),
			HandlerFunc(handler),
		},
	)
}

func (h *RegexpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, route := range h.routes {
		if route.pattern.MatchString(r.URL.Path) {
			match := route.pattern.FindStringSubmatch(r.URL.Path)
			params := make(map[string]string)
			for i, name := range route.pattern.SubexpNames() {
				if i != 0 {
					params[name] = match[i]
				}
			}

			route.handler.ServeHTTP(w, r, params)
			return
		}
	}

	http.NotFound(w, r)
}
