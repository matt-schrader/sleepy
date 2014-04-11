package sleepy

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
)

const (
	GET    = "GET"
	POST   = "POST"
	PUT    = "PUT"
	DELETE = "DELETE"
)

type Restful interface {
	GetResource() interface{}
}

// An API manages a group of resources by routing requests
// to the correct method on a matching resource and marshalling
// the returned data to JSON for the HTTP response.
//
// You can instantiate multiple APIs on separate ports. Each API
// will manage its own set of resources.
type API struct {
	mux *http.ServeMux
}

// NewAPI allocates and returns a new API.
func NewAPI() *API {
	return &API{}
}

func (api *API) requestHandler(resource interface{}, endpoint Endpoint) http.HandlerFunc {
	listMethod := reflect.ValueOf(resource).MethodByName("List")
	getMethod := reflect.ValueOf(resource).MethodByName("Get")
	postMethod := reflect.ValueOf(resource).MethodByName("Post")
	putMethod := reflect.ValueOf(resource).MethodByName("Put")
	deleteMethod := reflect.ValueOf(resource).MethodByName("Delete")

	return func(rw http.ResponseWriter, request *http.Request) {
		if request.ParseForm() != nil {
			rw.WriteHeader(http.StatusBadRequest)
			return
		}

		path := request.URL.Path
		if path[len(path)-1] == '/' {
			fmt.Printf("%s %d\n", path, len(path))
			path = path[:len(path)-2]
		}
		route, values := endpoint.FindRoute(request.URL.Path, request.Method)
		if route == nil || values == nil {
			rw.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		params := request.Form
		for k, v := range values {
			params[k] = v
		}

		var code int
		var data interface{}
		if request.Method == GET && getMethod != (reflect.Value{}) {
			var method reflect.Value
			if len(values) == 0 {
				method = listMethod
			} else {
				method = getMethod
			}
			inputs := make([]reflect.Value, 1)
			inputs[0] = reflect.ValueOf(params)
			codeData := method.Call(inputs)
			r := codeData[1]
			code = int(codeData[0].Int())
			data = reflect.Value(r).Interface()
		} else if request.Method == POST || request.Method == PUT {
			var resourceProxy interface{}
			if resource, ok := resource.(Restful); ok {
				resourceProxy = resource.GetResource()
			}
			if resourceProxy == nil {
				fmt.Printf("Route(%s) does not implement GetResource properly", route.Path)
				rw.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			decoder := json.NewDecoder(request.Body)

			err := decoder.Decode(&resourceProxy)
			fmt.Printf("proxy: %v\n", resourceProxy)
			if err != nil {
				fmt.Printf("Error occurred: %v\n", err)
				http.Error(rw, err.Error(), http.StatusInternalServerError)
				return
			}

			if request.Method == POST {
				inputs := make([]reflect.Value, 2)
				inputs[0] = reflect.ValueOf(resourceProxy)
				inputs[1] = reflect.ValueOf(params)
				codeData := postMethod.Call(inputs)
				code = int(codeData[0].Int())
				data = reflect.Value(codeData[1]).Interface()
			} else if request.Method == POST {
				inputs := make([]reflect.Value, 2)
				inputs[0] = reflect.ValueOf(resourceProxy)
				inputs[1] = reflect.ValueOf(params)
				codeData := putMethod.Call(inputs)
				code = int(codeData[0].Int())
				data = reflect.Value(codeData[1]).Interface()
			}
		} else if request.Method == DELETE {
			inputs := make([]reflect.Value, 1)
			inputs[0] = reflect.ValueOf(params)
			deleteResult := deleteMethod.Call(inputs)
			code = int(deleteResult[0].Int())
		}

		var content []byte
		var dataError error
		if code != 404 {
			content, dataError = json.Marshal(data)
			if dataError != nil {
				rw.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(code)
		rw.Write(content)
	}
}

// AddResource adds a new resource to an API. The API will route
// requests that match one of the given paths to the matching HTTP
// method on the resource.
func (api *API) AddResource(resource interface{}, path string) {

	if api.mux == nil {
		api.mux = http.NewServeMux()
	}

	endpoint := Endpoint{Root: path}

	if listMethod := reflect.ValueOf(resource).MethodByName("List"); listMethod != (reflect.Value{}) {
		listRoute := NewRoute(path, GET)
		endpoint.AddRoute(listRoute)
	}

	if getMethod := reflect.ValueOf(resource).MethodByName("Get"); getMethod != (reflect.Value{}) {
		getRoute := NewRoute(fmt.Sprintf("%s/:id", path), GET)
		endpoint.AddRoute(getRoute)
	}

	if postMethod := reflect.ValueOf(resource).MethodByName("Post"); postMethod != (reflect.Value{}) {
		postRoute := NewRoute(fmt.Sprintf("%s/:id", path), POST)
		endpoint.AddRoute(postRoute)
	}

	if putMethod := reflect.ValueOf(resource).MethodByName("Put"); putMethod != (reflect.Value{}) {
		route := NewRoute(fmt.Sprintf("%s/:id", path), PUT)
		endpoint.AddRoute(route)
	}

	if deleteMethod := reflect.ValueOf(resource).MethodByName("Delete"); deleteMethod != (reflect.Value{}) {
		route := NewRoute(fmt.Sprintf("%s/:id", path), DELETE)
		endpoint.AddRoute(route)
	}

	if len(endpoint.Routes) != 0 {
		api.mux.HandleFunc(fmt.Sprintf("%s/", path), api.requestHandler(resource, endpoint))
	}
}

// Start causes the API to begin serving requests on the given port.
func (api *API) Start(port int) error {
	if api.mux == nil {
		return errors.New("You must add at least one resource to this API.")
	}
	portString := fmt.Sprintf(":%d", port)
	fmt.Printf("Running server on port %d\n", port)
	return http.ListenAndServe(portString, api.mux)
}
