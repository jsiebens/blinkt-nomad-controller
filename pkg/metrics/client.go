package metrics

import (
	"github.com/armon/go-metrics"
)

type resourceAllocation struct {
	allocated   float32
	unallocated float32
}

func (r *resourceAllocation) percentage() float32 {
	return r.allocated / (r.unallocated + r.allocated)
}

func (c *Client) metrics() (*metrics.MetricsSummary, error) {
	m := metrics.MetricsSummary{}
	err := c.get("/v1/metrics", &m)

	if err != nil {
		return nil, err
	} else {
		return &m, nil
	}
}

func (c *Client) PercentageOfAllocatedResource(r string, max int) (float32, error) {
	if r == "allocations" {
		return c.percentageOfRunningAllocations(max)
	} else {
		m, err := c.metrics()
		if err != nil {
			return -1, err
		}

		result := &resourceAllocation{}
		for _, g := range m.Gauges {
			if g.Name == "nomad.client.allocated."+r {
				result.allocated = g.Value
			}
			if g.Name == "nomad.client.unallocated."+r {
				result.unallocated = g.Value
			}
		}

		return result.percentage(), nil
	}
}

func (c *Client) percentageOfRunningAllocations(max int) (float32, error) {
	m, err := c.metrics()
	if err != nil {
		return -1, err
	}

	result := float32(0)
	for _, g := range m.Gauges {
		if g.Name == "nomad.client.allocations.running" {
			result = g.Value
			break
		}
	}

	return result / float32(max), nil
}
