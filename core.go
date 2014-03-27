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

// GetSupported is the interface that provides the Get
// method a resource must support to receive HTTP GETs.
type GetSupported interface {
	Get(url.Values) (int, interface{})
}

// PostSupported is the interface that provides the Post
// method a resource must support to receive HTTP POSTs.
type PostSupported interface {
	Post(url.Values) (int, interface{})
}

// PutSupported is the interface that provides the Put
// method a resource must support to receive HTTP PUTs.
type PutSupported interface {
	Put(url.Values) (int, interface{})
}

// DeleteSupported is the interface that provides the Delete
// method a resource must support to receive HTTP DELETEs.
type DeleteSupported interface {
	Delete(url.Values) (int, interface{})
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

		handler, values := endpoint.FindRoute(request.URL.Path)

//		var handler func(url.Values) (int, interface{})
//
//		switch request.Method {
//		case GET:
//			if resource, ok := resource.(GetSupported); ok {
//				handler = resource.Get
//			}
//		case POST:
//			if resource, ok := resource.(PostSupported); ok {
//				handler = resource.Post
//			}
//		case PUT:
//			if resource, ok := resource.(PutSupported); ok {
//				handler = resource.Put
//			}
//		case DELETE:
//			if resource, ok := resource.(DeleteSupported); ok {
//				handler = resource.Delete
//			}
//		}

		if handler == nil {
			rw.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		params := request.Form
		for k, v := range values {
		    params[k] = v
		}

		code, data := handler(params)

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
	
	endpoint := Endpoint{ Root: path }
	if resource, ok := resource.(GetSupported); ok {
	    route := NewRoute(path, resource.Get)
	    endpoint.AddRoute(route)
	}
	
	api.mux.HandleFunc(fmt.Sprintf("%s/", path), api.requestHandler(resource, endpoint))
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

