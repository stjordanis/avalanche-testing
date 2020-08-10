package ipcs

import (
	"time"

	"github.com/ava-labs/avalanche-e2e-tests/gecko_client/utils"
	"github.com/ava-labs/gecko/api"
	"github.com/ava-labs/gecko/api/ipcs"
)

type Client struct {
	requester utils.EndpointRequester
}

// Returns a Client for interacting with the IPCS endpoint
func NewClient(uri string, requestTimeout time.Duration) *Client {
	return &Client{
		requester: utils.NewEndpointRequester(uri, "/ext/ipcs", "ipcs", requestTimeout),
	}
}

func (c *Client) PublishBlockchain(blockchainID string) (string, error) {
	res := &ipcs.PublishBlockchainReply{}
	err := c.requester.SendRequest("publishBlockchain", &ipcs.PublishBlockchainArgs{
		BlockchainID: blockchainID,
	}, res)
	return res.URL, err
}

func (c *Client) UnpublishBlockchain(blockchainID string) (bool, error) {
	res := &api.SuccessResponse{}
	err := c.requester.SendRequest("unpublishBlockchain", &ipcs.UnpublishBlockchainArgs{
		BlockchainID: blockchainID,
	}, res)
	return res.Success, err
}
