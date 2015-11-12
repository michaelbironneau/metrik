package metrik

import (
	"net/http"
	"regexp"
)

type route struct {
	pattern *regexp.Regexp
	handler http.Handler
}

type regexpHandler struct {
	routes []*route
}

func (h *regexpHandler) Route(pattern string, handler func(http.ResponseWriter, *http.Request)) *regexpHandler {
	compiledRegexp, err := regexp.Compile(pattern)
	if err != nil {
		panic(err) //this will be a compile time panic
	}
	h.routes = append(h.routes, &route{
		pattern: compiledRegexp,
		handler: http.HandlerFunc(handler),
	})
	return h
}

func (h *regexpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, route := range h.routes {
		if route.pattern.MatchString(r.URL.Path) {
			route.handler.ServeHTTP(w, r)
			return
		}
	}
	http.NotFound(w, r)
}
