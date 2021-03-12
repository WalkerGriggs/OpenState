package openstate

import (
	"encoding/json"
	"fmt"
	"net/http"

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

// tasksList returns the list of tasks from the server's FSM.
// TODO return a list of metadata, not just the total count of tasks.
// TODO allow stale reads from the follower, so the request isn't forwarded to
//      the leader.
func (s *HTTPServer) tasksList(resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	if ok, err := s.forward(resp, req); ok {
		return nil, err
	}

	names := make([]string, len(s.server.fsm.tasks))
	for i, task := range s.server.fsm.tasks {
		names[i] = task.Name
	}

	res := &api.TaskListResponse{
		Len:   len(s.server.fsm.tasks),
		Names: names,
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
	var out api.TaskDefineRequest
	dec := json.NewDecoder(req.Body)
	if err := dec.Decode(&out); err != nil {
		return nil, err
	}

	// TaskAddRequest
	args := &TaskDefineRequest{
		Task: out.Task,
	}

	fsmErr, index, err := s.server.raftApply(TaskDefineRequestType, args)
	if err, ok := fsmErr.(error); ok && err != nil {
		s.logger.Error("Failed to update FSM", "error", err, "fsm", true)
		return nil, err
	}

	if err != nil {
		s.logger.Error("Failed to update FSM", "error", err, "raft", true)
		return nil, err
	}

	res := &api.TaskDefineResponse{
		Index: index,
		Name:  out.Task.Name,
		Tags:  out.Task.Tags,
	}

	return res, nil
}
