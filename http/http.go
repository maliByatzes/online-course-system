package http

import (
	"context"
	"io"
	"net/http"

	"github.com/maliByatzes/ocs"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	errorCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "ocs_http_error_count",
		Help: "Total Number of errors by error code",
	}, []string{"code"})
)

type Client struct {
	URL string
}

func NewClient(u string) *Client {
	return &Client{URL: u}
}

func (c *Client) newRequest(ctx context.Context, method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, c.URL+url, body)
	if err != nil {
		return nil, err
	}

	if student := ocs.StudentFromContext(ctx); student != nil && student.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+student.APIKey)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}
