package health

import (
	"time"

	"github.com/ava-labs/avalanche-testing/avalanche_client/utils"
	"github.com/ava-labs/avalanchego/api/health"
)

// Client for Avalanche Health API Endpoint
type Client struct {
	requester utils.EndpointRequester
}

// NewClient returns a client to interact with Health API endpoint
func NewClient(uri string, requestTimeout time.Duration) *Client {
	return &Client{
		requester: utils.NewEndpointRequester(uri, "/ext/health", "health", requestTimeout),
	}
}

// GetLiveness returns a health check on the Avalanche node
func (c *Client) GetLiveness() (*health.GetLivenessReply, error) {
	res := &health.GetLivenessReply{}
	err := c.requester.SendRequest("getLiveness", struct{}{}, res)
	return res, err
}

// AwaitHealthy queries the GetLiveness endpoint [checks] times, with a pause of [interval]
// in between checks and returns early if GetLiveness returns healthy
func (c *Client) AwaitHealthy(checks int, interval time.Duration) (bool, error) {
	for i := 0; i < checks; i++ {
		time.Sleep(interval)
		res, err := c.GetLiveness()
		if err != nil {
			continue
		}

		if res.Healthy {
			return true, nil
		}
	}

	return false, nil
}
