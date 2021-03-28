package openstate

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/walkergriggs/openstate/openstate/structs"
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

func (s *HTTPServer) instanceEvent(resp http.ResponseWriter, req *http.Request, id string) (interface{}, error) {
	if ok, err := s.forward(resp, req); ok {
		return nil, err
	}

	var out structs.InstanceEventRequest
	dec := json.NewDecoder(req.Body)
	if err := dec.Decode(&out); err != nil {
		return nil, err
	}

	instance, err := s.server.fsm.state.GetInstanceByID(id)
	if err != nil {
		return nil, err
	}

	if instance == nil {
		return nil, fmt.Errorf("No such instance %s\n", id)
	}

	if err := instance.FSM.Do(out.EventName); err != nil {
		return nil, err
	}

	res := structs.InstanceEventResponse{
		Instance: instance,
	}

	return res, nil
}
