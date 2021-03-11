package openstate

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"

	log "github.com/hashicorp/go-hclog"
)

// HTTPServer wraps Server and exposes it over an HTTP interface.
type HTTPServer struct {
	server *Server

	// Address the HTTPServer listens on
	addr string

	// Http multiplexer
	mux *http.ServeMux

	// logger is an hclog instance for parity with the server logger
	logger log.Logger

	// Network listener and channel to link the HTTPServer and an actual
	// http.Server
	listener   net.Listener
	listenerCh chan struct{}
}

// newHTTPServer returns a new HTTPServ object.
func NewHTTPServer(s *Server, c *Config) (*HTTPServer, error) {
	mux := http.NewServeMux()

	addr, err := net.ResolveTCPAddr("tcp", c.HTTPAdvertise.String())
	if err != nil {
		return nil, err
	}

	ln, err := net.Listen("tcp", net.JoinHostPort(addr.IP.String(), strconv.Itoa(addr.Port)))
	if err != nil {
		return nil, fmt.Errorf("Failed to start HTTP listener: %v", err)
	}

	srv := &HTTPServer{
		server:     s,
		mux:        mux,
		logger:     c.Logger,
		listener:   ln,
		listenerCh: make(chan struct{}),
	}

	srv.registerHandlers()

	httpServer := http.Server{
		Addr:    srv.addr,
		Handler: srv.mux,
	}

	go func() {
		defer close(srv.listenerCh)
		httpServer.Serve(ln)
	}()

	return srv, nil
}

// registerHandlers maps each handler to an endpoint on the mux.
func (s *HTTPServer) registerHandlers() {
	s.mux.HandleFunc("/v1/tasks", s.wrap(s.tasksRequest))
}

// wrap wraps the handler function with some quality-of-life improvements. It
// returns a net/http ServeMux compliant handler function.
func (s *HTTPServer) wrap(handler func(resp http.ResponseWriter, req *http.Request) (interface{}, error)) func(resp http.ResponseWriter, req *http.Request) {
	f := func(resp http.ResponseWriter, req *http.Request) {
		reqURL := req.URL.String()
		start := time.Now()

		defer func() {
			s.logger.Debug("request complete", "method", req.Method, "path", reqURL, "duration", time.Since(start))
		}()

		obj, err := handler(resp, req)

	HAS_ERR:
		if err != nil {
			// TODO return more than just a 500
			resp.WriteHeader(500)
			resp.Write([]byte(err.Error()))

			s.logger.Error("request failed", "method", req.Method, "path", reqURL, "error", err, "code", 500)
			return
		}

		if obj != nil {
			bytes, err := json.Marshal(obj)
			if err != nil {
				goto HAS_ERR
			}

			resp.Header().Set("Content-Type", "application/json")
			resp.Write(bytes)
		}
	}

	return f
}

func (s *HTTPServer) forward(resp http.ResponseWriter, req *http.Request) (bool, error) {
	isLeader, address := s.server.getLeader()
	if isLeader {
		return false, nil
	}

	parts := s.server.peers[address]

	// TODO
	//   - scheme shouldn't be hardcoded
	//   - we shouldn't assume the request URL is a relative path
	u := url.URL{
		Scheme: "http",
		Host:   parts.http_addr.String(),
		Path:   req.URL.String(),
	}

	header := resp.Header()
	header.Set("Location", u.String())
	header.Set("Content-Type", "text/html; charset=utf-8")
	resp.WriteHeader(http.StatusPermanentRedirect)

	return true, nil
}
