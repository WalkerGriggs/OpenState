package openstate

import (
	"fmt"
	"net/http"
	"strings"
)

/*
 * /names
 */

// namesRequest routes the request to various functions which apply to all names.
func (s *HTTPServer) namesRequest(resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	switch req.Method {
	case "GET":
		return s.namesList(resp, req)
	default:
		return nil, fmt.Errorf("ErrInvalidMethod")
	}
}

// namesList returns the list of names from the server's FSM.
func (s *HTTPServer) namesList(resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	return s.server.fsm.names, nil
}

/*
 * /name/<name>
 */

// nameSpecificRequest routes the request to various name-specific functions.
func (s *HTTPServer) nameSpecificRequest(resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	path := strings.TrimPrefix(req.URL.Path, "/v1/name/")
	switch {
	default:
		return s.nameCRUD(resp, req, path)
	}
}

// nameCRUD routes the request to method-specific functions.
func (s *HTTPServer) nameCRUD(resp http.ResponseWriter, req *http.Request, name string) (interface{}, error) {
	switch req.Method {
	case "PUT", "POST":
		return s.nameUpdate(resp, req, name)
	default:
		return nil, fmt.Errorf("ErrInvalidMethod")
	}
}

// nameUpdate applies the given name to the Raft's FSM.
func (s *HTTPServer) nameUpdate(resp http.ResponseWriter, req *http.Request, name string) (interface{}, error) {
	if done, err := s.forward(resp, req); done {
		s.logger.Info("Forwarding request to leader")
		return nil, err
	}

	args := NameAddRequest{
		Name: name,
	}

	fsmErr, index, err := s.server.raftApply(NameAddRequestType, args)
	if err, ok := fsmErr.(error); ok && err != nil {
		s.logger.Error("adding name failed", "error", err, "fsm", true)
		return nil, err
	}

	if err != nil {
		s.logger.Error("adding name failed", "error", err, "raft", true)
		return nil, err
	}

	return index, nil
}
