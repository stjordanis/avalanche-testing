package ipcs

import (
	"time"

	"github.com/ava-labs/avalanche-testing/gecko_client/utils"
	"github.com/ava-labs/gecko/api"
	"github.com/ava-labs/gecko/api/ipcs"
)

// Client ...
type Client struct {
	requester utils.EndpointRequester
}

// NewClient returns a Client for interacting with the IPCS endpoint
func NewClient(uri string, requestTimeout time.Duration) *Client {
	return &Client{
		requester: utils.NewEndpointRequester(uri, "/ext/ipcs", "ipcs", requestTimeout),
	}
}

// PublishBlockchain requests the node to begin publishing consensus and decision events
func (c *Client) PublishBlockchain(blockchainID string) (*ipcs.PublishBlockchainReply, error) {
	res := &ipcs.PublishBlockchainReply{}
	err := c.requester.SendRequest("publishBlockchain", &ipcs.PublishBlockchainArgs{
		BlockchainID: blockchainID,
	}, res)
	return res, err
}

// UnpublishBlockchain requests the node to stop publishing consensus and decision events
func (c *Client) UnpublishBlockchain(blockchainID string) (bool, error) {
	res := &api.SuccessResponse{}
	err := c.requester.SendRequest("unpublishBlockchain", &ipcs.UnpublishBlockchainArgs{
		BlockchainID: blockchainID,
	}, res)
	return res.Success, err
}
