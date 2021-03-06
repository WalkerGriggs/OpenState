package api

type Names struct {
	client *Client
}

func (c *Client) Names() *Names {
	return &Names{client: c}
}

func (s *Names) List() (string, error) {
	var resp string
	err := s.client.query("/v1/names", &resp, nil)
	if err != nil {
		return "", err
	}
	return resp, nil
}
