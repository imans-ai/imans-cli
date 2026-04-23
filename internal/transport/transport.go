package transport

import (
	"net/http"
	"time"
)

type Options struct {
	Timeout time.Duration
}

func NewHTTPClient(opts Options) *http.Client {
	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	return &http.Client{Timeout: timeout}
}
