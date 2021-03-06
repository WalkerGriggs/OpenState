package openstate

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	log "github.com/hashicorp/go-hclog"
)

type HTTPServer struct {
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

// newHTTPServer returns a new HTTPServ object. Note,
func newHTTPServer(c *Config) (*HTTPServer, error) {
	mux := http.NewServeMux()

	addr, err := net.ResolveTCPAddr("tcp", c.HTTPAdvertise.String())
	if err != nil {
		return nil, err
	}

	ln, err := net.Listen("tcp", net.JoinHostPort(addr.IP.String(), strconv.Itoa(addr.Port)))
	if err != nil {
		return nil, fmt.Errorf("Failed to start HTTP listener: %v", err)
	}

	s := &HTTPServer{
		mux:        mux,
		logger:     c.Logger,
		listener:   ln,
		listenerCh: make(chan struct{}),
	}

	s.registerHandlers()

	return s, nil
}

// serve wraps a net/http Server's Serve() method using the HTTPServer's mux,
// address, and listener.
func (s *HTTPServer) serve() error {
	httpServer := http.Server{
		Addr:    s.addr,
		Handler: s.mux,
	}

	defer close(s.listenerCh)
	return httpServer.Serve(s.listener)
}

// registerHandlers maps each handler to an endpoint on the mux.
func (s *HTTPServer) registerHandlers() {
	s.mux.HandleFunc("/v1/names", s.wrap(s.NamesRequest))
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
