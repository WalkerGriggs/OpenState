package openstate

import (
	"bytes"
	"encoding/json"
	"math/rand"
	"net"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	log "github.com/hashicorp/go-hclog"
	"github.com/hashicorp/nomad/helper/freeport"
	"github.com/stretchr/testify/require"

	"github.com/walkergriggs/openstate/openstate/mocks"
	"github.com/walkergriggs/openstate/openstate/structs"
)

func NewTestServer() (*Server, error) {
	config := DefaultConfig()

	level := log.Trace
	opts := &log.LoggerOptions{
		Level:           level,
		Output:          os.Stdout,
		IncludeLocation: true,
	}

	ports := freeport.MustTake(3)

	config.RaftAdvertise = &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: ports[0]}
	config.SerfAdvertise = &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: ports[1]}
	config.HTTPAdvertise = &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: ports[2]}

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
	require.NoError(t, err)
	require.NotNil(t, server)

	// Create the http server to use as the method receiver
	httpServer, err := NewHTTPServer(server, server.config)
	require.NoError(t, err)
	require.NotNil(t, httpServer)

	// Mock the definition
	def := mocks.Definition()

	// Insert the definition
	err = server.fsm.state.InsertDefinition(def)
	require.NoError(t, err)

	// Mock out the ResponseWriter and Request
	req := httptest.NewRequest("GET", "http://example.com/", nil)
	w := httptest.NewRecorder()

	// Grab the list
	out, err := httpServer.tasksList(w, req)
	require.NoError(t, err)
	require.NotNil(t, out)

	res := out.(*structs.TaskListResponse)

	require.Equal(t, len(res.Definitions), 1, "there should be a single definition")
	require.Equal(t, res.Definitions[0], def, "the definitions should match")
}

func TestTaskEndpoints_Update(t *testing.T) {
	t.Parallel()

	// Create a new server
	server, err := NewTestServer()
	require.NoError(t, err)
	require.NotNil(t, server)

	// Create the http server to use as the method receiver
	httpServer, err := NewHTTPServer(server, server.config)
	require.NoError(t, err)
	require.NotNil(t, httpServer)

	// Mock the definition
	def := mocks.Definition()

	body := structs.TaskDefineRequest{
		Definition: def,
	}

	// Write TaskDefineRequest to io.Reader for http.Request
	buf := bytes.NewBuffer(nil)
	enc := json.NewEncoder(buf)
	err = enc.Encode(body)
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "http://example.com/", buf)
	w := httptest.NewRecorder()

	out, err := httpServer.tasksUpdate(w, req)
	require.NoError(t, err)
	require.NotNil(t, out)

	res := out.(structs.TaskDefineResponse)

	require.Equal(t, res.Definition.Name, def.Name, "the definitions should match")
	require.Equal(t, res.Definition.FSM, def.FSM, "the definition FSMs should match")
}

func TestTaskEndpoints_Run(t *testing.T) {
	t.Parallel()

	// Create a new server
	server, err := NewTestServer()
	require.NoError(t, err)
	require.NotNil(t, server)

	// Create the http server to use as the method receiver
	httpServer, err := NewHTTPServer(server, server.config)
	require.NoError(t, err)
	require.NotNil(t, httpServer)

	// Mock the definition
	def := mocks.Definition()

	// Insert the definition
	err = server.fsm.state.InsertDefinition(def)
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "http://example.com/", nil)
	w := httptest.NewRecorder()

	out, err := httpServer.taskRun(w, req, def.Name)
	require.NoError(t, err)
	require.NotNil(t, out)

	res := out.(structs.TaskRunResponse)

	instances, err := server.fsm.state.GetInstances()
	require.NoError(t, err)
	require.NotNil(t, instances)

	require.Equal(t, len(instances), 1, "it should insert the running instance")
	require.Equal(t, instances[0].ID, res.Instance.ID, "it should return the running instance")
	require.Equal(t, res.Instance.FSM.State(), def.FSM.Initial, "it should be in the initial state")
}

func TestTaskEndpoints_Ps(t *testing.T) {
	t.Parallel()

	// Create a new server
	server, err := NewTestServer()
	require.NoError(t, err)
	require.NotNil(t, server)

	// Create the http server to use as the method receiver
	httpServer, err := NewHTTPServer(server, server.config)
	require.NoError(t, err)
	require.NotNil(t, httpServer)

	// Mock two instances from two different definitions
	i1 := mocks.Instance()

	i2 := mocks.Instance()
	i2.Definition.Name = "i2-definition"
	i2.ID = "i2-instance"

	for _, instance := range []*structs.Instance{i1, i2} {
		err = server.fsm.state.InsertDefinition(instance.Definition)
		require.NoError(t, err)

		err = server.fsm.state.InsertInstance(instance)
		require.NoError(t, err)
	}

	instances, err := server.fsm.state.GetInstances()
	require.NoError(t, err)
	require.NotNil(t, instances)
	require.Equal(t, len(instances), 2)

	req := httptest.NewRequest("GET", "http://example.com/", nil)
	w := httptest.NewRecorder()

	out, err := httpServer.taskPs(w, req, i1.Definition.Name)
	require.NoError(t, err)
	require.NotNil(t, out)

	res := out.(*structs.TaskPsResponse)

	require.Equal(t, len(res.Instances), 1, "it should return a single instance")
	require.Equal(t, res.Instances[0].ID, i1.ID, "it should return the only matching instance")
}
