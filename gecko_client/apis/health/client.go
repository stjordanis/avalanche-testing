package health

import (
	"time"

	"github.com/ava-labs/avalanche-e2e-tests/gecko_client/utils"
	"github.com/ava-labs/gecko/api/health"
)

type Client struct {
	requester utils.EndpointRequester
}

// Returns Client to interact with Info API endpoint
func NewClient(uri string, requestTimeout time.Duration) *Client {
	return &Client{
		requester: utils.NewEndpointRequester(uri, "/ext/health", "health", requestTimeout),
	}
}

// GetLiveness...
func (c *Client) GetLiveness() (*health.GetLivenessReply, error) {
	res := &health.GetLivenessReply{}
	err := c.requester.SendRequest("getLiveness", struct{}{}, res)
	return res, err
}
