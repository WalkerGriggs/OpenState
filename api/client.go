package api

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// Client provideds the OpenState API client.
type Client struct {
	httpClient *http.Client
	config     Config
}

// NewClient returns a new client with the default httpClient.
// TODO allow the method to take in an optional config instead of the default
func NewClient() (*Client, error) {
	config := DefaultConfig()

	httpClient := &http.Client{}

	client := &Client{
		config:     *config,
		httpClient: httpClient,
	}

	return client, nil
}

// query builds an http.Request and perofrmance the query itself. The response
// body is decoded into the optional, provided out interface. It raises an error
// if the reponse status code is anything but 200.
func (c *Client) query(endpoint string, out interface{}, q *QueryOptions) error {
	r, err := c.newRequest("GET", endpoint)
	if err != nil {
		return err
	}

	r.setQueryOptions(q)

	resp, err := c.doRequest(r)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		var buf bytes.Buffer
		io.Copy(&buf, resp.Body)
		return fmt.Errorf("Unexpected response code: %d (%s)", resp.StatusCode, buf)
	}

	if out != nil {
		if err := decodeBody(resp, &out); err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) write(endpoint string, in, out interface{}, q *WriteOptions) error {
	r, err := c.newRequest("POST", endpoint)
	if err != nil {
		return err
	}

	r.obj = in
	r.setWriteOptions(q)

	resp, err := c.doRequest(r)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		var buf bytes.Buffer
		io.Copy(&buf, resp.Body)
		return fmt.Errorf("Unexpected response code: %d (%s)", resp.StatusCode, buf)
	}

	if out != nil {
		if err := decodeBody(resp, &out); err != nil {
			return err
		}
	}

	return nil
}

// newRequest creates a new request.
func (c *Client) newRequest(method, path string) (*request, error) {
	base, err := url.Parse(c.config.Address)
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(path)
	if err != nil {
		return nil, err
	}

	r := &request{
		config: &c.config,
		method: method,
		url: &url.URL{
			Scheme:  base.Scheme,
			User:    base.User,
			Host:    base.Host,
			Path:    u.Path,
			RawPath: u.RawPath,
		},
		params: make(map[string][]string),
	}

	if c.config.WaitTime != 0 {
		r.params.Set("wait", r.config.WaitTime.String())
	}

	for key, values := range u.Query() {
		for _, value := range values {
			r.params.Add(key, value)
		}
	}

	return r, nil
}

// doRequest converts the request object to a standard http.Request and actually
// performs the request.
func (c *Client) doRequest(r *request) (*http.Response, error) {
	req, err := r.toHTTP()
	if err != nil {
		return nil, err
	}

	return c.httpClient.Do(req)
}
