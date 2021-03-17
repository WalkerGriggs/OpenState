package openstate

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/walkergriggs/openstate/api"
)

// taskSpecificRequest routes a request to various functions which apply to an
// individual task definition.
func (s *HTTPServer) instanceSpecificRequest(resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	path := strings.TrimPrefix(req.URL.Path, "/v1/instance/")
	switch {
	case strings.HasSuffix(path, "/event"):
		instanceName := strings.TrimSuffix(path, "/event")
		return s.instanceEvent(resp, req, instanceName)
	default:
		return nil, fmt.Errorf("No such endpoint")
	}
}

func (s *HTTPServer) instanceEvent(resp http.ResponseWriter, req *http.Request, name string) (interface{}, error) {
	if ok, err := s.forward(resp, req); ok {
		return nil, err
	}

	var out api.InstanceEventRequest
	dec := json.NewDecoder(req.Body)
	if err := dec.Decode(&out); err != nil {
		return nil, err
	}

	instance, ok := s.server.fsm.instances[name]
	if !ok {
		return nil, fmt.Errorf("No such instance %s\n", name)
	}

	if err := instance.FSM.Do(out.EventName); err != nil {
		return nil, err
	}

	res := &api.InstanceEventResponse{
		CurrentState: instance.FSM.State(),
	}

	return res, nil
}
