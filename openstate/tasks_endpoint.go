package openstate

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/walkergriggs/openstate/api"
)

// tasksRequest routes a request to the various functions which apply to all tasks.
func (s *HTTPServer) tasksRequest(resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	switch req.Method {
	case "GET":
		return s.tasksList(resp, req)
	case "POST", "PUT":
		return s.tasksUpdate(resp, req)
	default:
		return nil, fmt.Errorf("ErrInvalidMethod")
	}
}

// taskSpecificRequest routes a request to various functions which apply to an
// individual task definition.
func (s *HTTPServer) taskSpecificRequest(resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	path := strings.TrimPrefix(req.URL.Path, "/v1/task/")
	switch {
	case strings.HasSuffix(path, "/run"):
		taskName := strings.TrimSuffix(path, "/run")
		return s.taskRun(resp, req, taskName)
	case strings.HasSuffix(path, "/ps"):
		taskName := strings.TrimSuffix(path, "/ps")
		return s.taskPs(resp, req, taskName)
	default:
		return nil, fmt.Errorf("No such endpoint")
		// return s.jobCRUD(resp, req, path)
	}
}

// tasksList returns the list of tasks from the server's FSM.
// TODO return a list of metadata, not just the total count of tasks.
// TODO allow stale reads from the follower, so the request isn't forwarded to
//      the leader.
func (s *HTTPServer) tasksList(resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	if ok, err := s.forward(resp, req); ok {
		return nil, err
	}

	defs := make([]*Definition, 0)

	for _, def := range s.server.fsm.definitions {
		defs = append(defs, def)
	}

	res := &TaskListResponse{
		Definitions: defs,
	}

	return res, nil
}

// taskUpdate applies a task definition request to the raft logs which either
// adds a new or updates and existing task.
func (s *HTTPServer) tasksUpdate(resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	if ok, err := s.forward(resp, req); ok {
		return nil, err
	}

	// Decode and repackage
	var out TaskDefineRequest
	dec := json.NewDecoder(req.Body)
	if err := dec.Decode(&out); err != nil {
		return nil, err
	}

	fsmErr, _, err := s.server.raftApply(TaskDefineRequestType, out)
	if err, ok := fsmErr.(error); ok && err != nil {
		s.logger.Error("Failed to update FSM", "error", err, "fsm", true)
		return nil, err
	}

	if err != nil {
		s.logger.Error("Failed to update FSM", "error", err, "raft", true)
		return nil, err
	}

	// TODO return more than the requested definition
	res := TaskDefineResponse{
		Definition: out.Definition,
	}

	return res, nil
}

// taskRun initializes a new task instance given a task definition and applies
// the instance to the raft logs.
func (s *HTTPServer) taskRun(resp http.ResponseWriter, req *http.Request, defName string) (interface{}, error) {
	if ok, err := s.forward(resp, req); ok {
		return nil, err
	}

	// We current don't need to decode the request body, because we're not sending
	// anything from the API. Maybe in the future when instances get more complex..
	def, ok := s.server.fsm.definitions[defName]
	if !ok {
		return nil, fmt.Errorf("No task definition with name %s", defName)
	}

	// Convert from api.FSM to fsm.FSM
	fsm, err := api.Ftof(def.FSM)
	if err != nil {
		return nil, err
	}

	instance := &Instance{
		ID:         fmt.Sprintf("%s-%s", def.Metadata.Name, generateUUID()),
		Definition: def,
		FSM:        fsm,
	}

	args := &TaskRunRequest{
		Instance: instance,
	}

	fsmErr, _, err := s.server.raftApply(TaskRunRequestType, args)
	if err, ok := fsmErr.(error); ok && err != nil {
		s.logger.Error("Failed to update FSM", "error", err, "fsm", true)
		return nil, err
	}

	if err != nil {
		s.logger.Error("Failed to update FSM", "error", err, "raft", true)
		return nil, err
	}

	res := TaskRunResponse{
		Instance: instance,
	}

	return res, nil
}

// taskPs returns the list of running task instances for a given task definition
// from the server's FSM.
func (s *HTTPServer) taskPs(resp http.ResponseWriter, req *http.Request, defName string) (interface{}, error) {
	if ok, err := s.forward(resp, req); ok {
		return nil, err
	}

	if _, ok := s.server.fsm.definitions[defName]; !ok {
		return nil, fmt.Errorf("No task definition with name %s", defName)
	}

	instances := make([]*Instance, 0)
	for id, instance := range s.server.fsm.instances {
		if strings.HasPrefix(id, defName) {
			instances = append(instances, instance)
		}
	}

	res := &TaskPsResponse{
		Instances: instances,
	}

	return res, nil
}
