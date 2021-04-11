package openstate

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/walkergriggs/openstate/openstate/mocks"
	"github.com/walkergriggs/openstate/openstate/structs"
)

func TestInstancesEndpoint_Event(t *testing.T) {
	t.Parallel()

	// Create a new server
	server, err := NewTestServer()
	require.NoError(t, err)
	require.NotNil(t, server)

	// Create the http server to use as the method receiver
	httpServer, err := NewHTTPServer(server, server.config)
	require.NoError(t, err)
	require.NotNil(t, httpServer)

	// Mock the instance
	instance := mocks.Instance()

	// Insert the instance
	err = server.fsm.state.InsertInstance(instance)
	require.NoError(t, err)

	// Define an event request
	body := structs.InstanceEventRequest{
		EventName: "turn_yellow",
	}

	// Write InstanceEventRequest to io.Reader for http.Request
	buf := bytes.NewBuffer(nil)
	enc := json.NewEncoder(buf)
	err = enc.Encode(body)
	require.NoError(t, err)

	// Mock out the ResponseWriter and Request
	req := httptest.NewRequest("GET", "http://example.com/", buf)
	w := httptest.NewRecorder()

	// Grab the list
	out, err := httpServer.instanceEvent(w, req, instance.ID)
	require.NoError(t, err)
	require.NotNil(t, out)

	res := out.(structs.InstanceEventResponse)

	require.Equal(t, "yellow", res.Instance.FSM.State(), "it should transition to yellow")
}
