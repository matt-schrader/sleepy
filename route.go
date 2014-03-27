package sleepy

import (
  "net/url"
  "regexp"
  "fmt"
  "strings"
)

type Endpoint struct {
    Root        string
    Routes      []Route
}

func (endpoint *Endpoint) AddRoute(route Route) {
    if(endpoint.Routes == nil) {
        fmt.Printf("Initializing routes array\n")
        endpoint.Routes = make([]Route, 0)
    }
    fmt.Printf("Appending %v to %v\n", endpoint.Routes, route)
    endpoint.Routes = append(endpoint.Routes, route)
    fmt.Printf("Routes %v\n", endpoint.Routes)
}

func (endpoint Endpoint) FindRoute(path string) (func(url.Values) (int, interface{}), url.Values) {
    matchPath := path
    if strings.HasSuffix(matchPath, "/") {
        matchPath = matchPath[:len(matchPath) - 1]
    }
    fmt.Printf("Finding route to %s\n", matchPath)
    fmt.Printf("%v\n", endpoint.Routes)
    for _, route := range endpoint.Routes {
        values, ok := route.Match(matchPath)
        fmt.Printf("Got %v---%v", values, ok)
        if ok {
            return route.Handler, values
        }
    }
    return nil, nil
}

type Route struct {
  Path          string
  PathRegexp    *regexp.Regexp
  Handler       func(url.Values) (int, interface{})
}

func NewRoute(path string, handler func(url.Values) (int, interface{})) Route {
  route := Route{
    Path:       path,
    Handler:    handler,
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
