package sleepy

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

const (
	GET    = "GET"
	POST   = "POST"
	PUT    = "PUT"
	DELETE = "DELETE"
)

type ListSupported interface {
	List(url.Values) (int, interface{})
}

// GetSupported is the interface that provides the Get
// method a resource must support to receive HTTP GETs.
type GetSupported interface {
	Get(url.Values) (int, interface{})
}

// PostSupported is the interface that provides the Post
// method a resource must support to receive HTTP POSTs.
type PostSupported interface {
	Post(interface{}, url.Values) (int, interface{})
}

// PutSupported is the interface that provides the Put
// method a resource must support to receive HTTP PUTs.
type PutSupported interface {
	Put(interface{}, url.Values) (int, interface{})
}

// DeleteSupported is the interface that provides the Delete
// method a resource must support to receive HTTP DELETEs.
type DeleteSupported interface {
	Delete(url.Values) int
}

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
	return func(rw http.ResponseWriter, request *http.Request) {
		if request.ParseForm() != nil {
			rw.WriteHeader(http.StatusBadRequest)
			return
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
		if request.Method == GET {
			code, data = route.RetrieveHandler(params)
		} else if request.Method == POST || request.Method == PUT {
			var resourceProxy interface{}
			if resource, ok := resource.(Restful); ok {
				resourceProxy = resource.GetResource()
				fmt.Println("Resource: %v\n", resourceProxy)
			} else {
				fmt.Println("Not a Resource!!!!!!!!!!!!!")
			}
			//if resourceProxy == nil {
			//	fmt.Printf("Route(%s) does not implement GetResource properly", route.Path)
			//	rw.WriteHeader(http.StatusMethodNotAllowed)
			//	return
			//}
			decoder := json.NewDecoder(request.Body)

			err := decoder.Decode(&resourceProxy)
			fmt.Printf("proxy: %v\n", resourceProxy)
			if err != nil {
				fmt.Printf("Error occurred: %v\n", err)
				http.Error(rw, err.Error(), http.StatusInternalServerError)
				return
			}
			code, data = route.SaveHandler(&resourceProxy, params)
		} else if request.Method == DELETE {
			code = route.DeleteHandler(params)
		}

		content, err := json.Marshal(data)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			return
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

	rootEndpoint := Endpoint{Root: path}
	nestedEndpoint := Endpoint{Root: path}

	if resource, ok := resource.(ListSupported); ok {
		route := NewRetrieveRoute(path, GET, resource.List)
		rootEndpoint.AddRoute(route)
	}

	if resource, ok := resource.(GetSupported); ok {
		route := NewRetrieveRoute(fmt.Sprintf("%s/:id", path), GET, resource.Get)
		nestedEndpoint.AddRoute(route)
	}

	if resource, ok := resource.(PostSupported); ok {
		route := NewSaveRoute(fmt.Sprintf("%s/:id", path), POST, resource.Post)
		nestedEndpoint.AddRoute(route)
	}

	if resource, ok := resource.(PutSupported); ok {
		route := NewSaveRoute(fmt.Sprintf("%s/:id", path), PUT, resource.Put)
		nestedEndpoint.AddRoute(route)
	}

	if resource, ok := resource.(DeleteSupported); ok {
		route := NewDeleteRoute(fmt.Sprintf("%s/:id", path), DELETE, resource.Delete)
		nestedEndpoint.AddRoute(route)
	}

	api.mux.HandleFunc(path, api.requestHandler(resource, rootEndpoint))
	api.mux.HandleFunc(fmt.Sprintf("%s/", path), api.requestHandler(resource, nestedEndpoint))
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
