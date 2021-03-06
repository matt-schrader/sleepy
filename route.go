package sleepy

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

type Endpoint struct {
	Root   string
	Routes []Route
}

func (endpoint *Endpoint) AddRoute(route Route) {
	if route == (Route{}) {
		fmt.Printf("Error: Cannot pass empty route to Endpoint.AddRoute\n")
		return
	}
	if len(endpoint.Routes) == 0 {
		endpoint.Routes = make([]Route, 0)
	}
	fmt.Printf("%s %s\n", route.Path, route.Method)
	endpoint.Routes = append(endpoint.Routes, route)
}

func (endpoint *Endpoint) FindRoute(path string, method string) (*Route, url.Values) {
	matchPath := path
	if strings.HasSuffix(matchPath, "/") {
		matchPath = matchPath[:len(matchPath)-1]
	}
	for _, route := range endpoint.Routes {
		if route.Method == method {
			values, ok := route.Match(matchPath)
			if ok {
				return &route, values
			}
		}
	}
	return nil, nil
}

type Route struct {
	Path       string
	Method     string
	PathRegexp *regexp.Regexp
}

func NewRoute(path string, method string) Route {
	route := Route{
		Path:       path,
		Method:     method,
		PathRegexp: regexp.MustCompile(pathToRegexpString(path)),
	}
	return route
}

func (route Route) Match(path string) (url.Values, bool) {
	values := make(url.Values)

	matches := route.PathRegexp.FindAllStringSubmatch(path, -1)
	if matches != nil {
		names := route.PathRegexp.SubexpNames()
		for i := 1; i < len(names); i++ {
			name := names[i]
			value := matches[0][i]
			if len(name) > 0 && len(value) > 0 {
				values.Set(name, value)
			}
		}

		return values, true
	}

	return values, false
}
