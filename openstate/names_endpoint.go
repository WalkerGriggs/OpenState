package openstate

import (
	"net/http"
)

// HACK hardcoded for endpoint development
func (s *HTTPServer) NamesRequest(resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	return "Hello!", nil
}
