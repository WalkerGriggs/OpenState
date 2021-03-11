package api

import (
	"context"
)

type WriteOptions struct {
	// ctx is passed through to http.NewRequestWithContext. Defaults to the
	// background context when building the http.Request
	ctx context.Context
}

func (o *WriteOptions) Context() context.Context {
	if o != nil && o.ctx != nil {
		return o.ctx
	}

	return context.Background()
}
