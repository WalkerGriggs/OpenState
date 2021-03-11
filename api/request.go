package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// request abstracts an http.Request
type request struct {
	// config is used to inject client config options into the request struct
	config *Config

	// method is the HTTP request method
	method string

	// url is the query endpoint
	url *url.URL

	// Values maps a string key to a list of values. It is used for query
	// parameters and form values. Unlike in the http.Header map, the keys in a
	// Values map are case-sensitive.
	params url.Values

	// body is the request body formatted as an io.Reader
	body io.Reader

	// obj is a body alternative. If provided and body is left empty, this object
	// will be json encoded and used as the request.Body
	obj interface{}

	// ctx is passed to http.NewRequestWithContext. Defaults to context.Background
	ctx context.Context
}

// setQueryOptions unpacks QueryOptions into the request params map.
func (r *request) setQueryOptions(q *QueryOptions) {
	if q == nil {
		return
	}

	if q.WaitTime != 0 {
		r.params.Set("wait", q.WaitTime.String())
	}

	for k, v := range q.Params {
		r.params.Set(k, v)
	}

	r.ctx = q.Context()
}

// setWriteOptions unpacks WriteOptions into the request params map.
func (r *request) setWriteOptions(w *WriteOptions) {
	if w == nil {
		return
	}

	r.ctx = w.Context()
}

// toHTTP converts our request abstraction to an actual http.Request.
func (r *request) toHTTP() (*http.Request, error) {
	r.url.RawQuery = r.params.Encode()

	// encode the obj as the request body if applicable.
	if r.body == nil && r.obj != nil {
		if b, err := encodeBody(r.obj); err != nil {
			return nil, err
		} else {
			r.body = b
		}
	}

	if r.ctx == nil {
		r.ctx = context.Background()
	}

	req, err := http.NewRequestWithContext(r.ctx, r.method, r.url.RequestURI(), r.body)
	if err != nil {
		return nil, err
	}

	req.URL.Host = r.url.Host
	req.URL.Scheme = r.url.Scheme
	req.Host = r.url.Host

	return req, nil
}

// decodeBody unmarshals the response body into the provided interface. It
// raises an error if the response body is empty and the provided interface is
// not nil.
func decodeBody(resp *http.Response, out interface{}) error {
	switch resp.ContentLength {
	case 0:
		if out == nil {
			return nil
		}
		return fmt.Errorf("Got 0 byte response with non-nil decode object")
	default:
		dec := json.NewDecoder(resp.Body)
		return dec.Decode(out)
	}
}

// encodeBody marshals the provided interface to an io.Reader.
func encodeBody(obj interface{}) (io.Reader, error) {
	if reader, ok := obj.(io.Reader); ok {
		return reader, nil
	}

	buf := bytes.NewBuffer(nil)
	enc := json.NewEncoder(buf)
	if err := enc.Encode(obj); err != nil {
		return nil, err
	}
	return buf, nil
}
