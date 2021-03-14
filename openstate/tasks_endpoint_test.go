package openstate

import (
	"math/rand"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	log "github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
)

func NewTestServer() (*Server, error) {
	config := DefaultConfig()

	level := log.Trace
	opts := &log.LoggerOptions{
		Level:           level,
		Output:          os.Stdout,
		IncludeLocation: true,
	}

	config.NodeName = "foo"
	config.Logger = log.NewInterceptLogger(opts)
	config.LogOutput = opts.Output
	config.DevMode = true
	config.BootstrapExpect = 1

	config.SerfConfig.MemberlistConfig.BindAddr = "127.0.0.1"
	config.SerfConfig.MemberlistConfig.SuspicionMult = 2
	config.SerfConfig.MemberlistConfig.RetransmitMult = 2
	config.SerfConfig.MemberlistConfig.ProbeTimeout = 50 * time.Millisecond
	config.SerfConfig.MemberlistConfig.ProbeInterval = 100 * time.Millisecond
	config.SerfConfig.MemberlistConfig.GossipInterval = 100 * time.Millisecond

	config.RaftConfig.LeaderLeaseTimeout = 50 * time.Millisecond
	config.RaftConfig.HeartbeatTimeout = 50 * time.Millisecond
	config.RaftConfig.ElectionTimeout = 50 * time.Millisecond
	config.RaftTimeout = 500 * time.Millisecond

	// Wait so the server has time to bootstrap
	defer func() {
		wait := time.Duration(rand.Int31n(2000)) * time.Millisecond
		time.Sleep(wait)
	}()

	return NewServer(config)
}

func TestTasksEndpoint_List(t *testing.T) {
	t.Parallel()

	// Create a new server
	server, err := NewTestServer()
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	// Serve HTTP endpoints. We won't make any requests over the network, but we
	// need the httpServer as a function receiver.
	httpServer, err := NewHTTPServer(server, server.config)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	// Define a no-op task
	task := &Task{
		FSM:  nil,
		Name: "TestingTask",
		Tags: []string{"test"},
	}

	// Hack to insert a new task
	server.fsm.tasks = append(server.fsm.tasks, task)

	// Mock out the ResponseWriter and Request
	req := httptest.NewRequest("GET", "http://example.com/foo", nil)
	w := httptest.NewRecorder()

	// Grab the list
	out, err := httpServer.tasksList(w, req)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	// Assertions
	assert.Equal(t, len(out.Names), 1, "There should we one defined task")
	assert.Equal(t, out.Names[0], task.Name, "The task names should match")
}
