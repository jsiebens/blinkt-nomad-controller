package metrics

import (
	"github.com/armon/go-metrics"
)

func (c *Client) Metrics() (*metrics.MetricsSummary, error) {
	m := metrics.MetricsSummary{}
	err := c.get("/v1/metrics", &m)

	if err != nil {
		return nil, err
	} else {
		return &m, nil
	}
}
