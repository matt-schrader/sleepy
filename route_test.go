package sleepy

import (
	"testing"
)

func TestNewRoute(t *testing.T) {
	endpoint := Endpoint{}

	endpoint.AddRoute(Route{})
	if len(endpoint.Routes) > 0 {
		t.Errorf("Expected routes to be empty but got %d routes", len(endpoint.Routes))
	}

	bookRoute := Route{Path: "/book", Method: GET}
	endpoint.AddRoute(bookRoute)
	if len(endpoint.Routes) != 1 {
		t.Errorf("Expected routes to have 1 route but actually has %d routes", len(endpoint.Routes))
	}
}
