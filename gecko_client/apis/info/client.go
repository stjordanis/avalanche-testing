package info

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/ava-labs/avalanche-e2e-tests/gecko_client/utils"
	"github.com/ava-labs/gecko/api/info"
	"github.com/ava-labs/gecko/network"
)

// Client is an Info API Client
type Client struct {
	requester utils.EndpointRequester
}

// NewClient returns a new Info API Client
func NewClient(uri string, requestTimeout time.Duration) *Client {
	return &Client{
		requester: utils.NewEndpointRequester(uri, "/ext/info", "info", requestTimeout),
	}
}

// GetNodeID ...
func (c *Client) GetNodeID() (string, error) {
	res := &info.GetNodeIDReply{}
	err := c.requester.SendRequest("getNodeID", struct{}{}, res)
	return res.NodeID, err
}

// GetNetworkID ...
func (c *Client) GetNetworkID() (uint32, error) {
	res := &info.GetNetworkIDReply{}
	err := c.requester.SendRequest("getNetworkID", struct{}{}, res)
	return uint32(res.NetworkID), err
}

// GetNetworkName ...
func (c *Client) GetNetworkName() (string, error) {
	res := &info.GetNetworkNameReply{}
	err := c.requester.SendRequest("getNetworkName", struct{}{}, res)
	return res.NetworkName, err
}

// GetBlockchainID ...
func (c *Client) GetBlockchainID() (string, error) {
	res := &info.GetBlockchainIDReply{}
	err := c.requester.SendRequest("getBlockchainID", struct{}{}, res)
	return res.BlockchainID, err
}

// Peers ...
func (c *Client) Peers() ([]network.PeerID, error) {
	res := &info.PeersReply{}
	err := c.requester.SendRequest("peers", struct{}{}, res)
	resVal, err := json.Marshal(res)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	} else {
		fmt.Printf("\n%s\n", resVal)
	}
	return res.Peers, err
}

// IsBootstrapped ...
func (c *Client) IsBootstrapped(chain string) (bool, error) {
	res := &info.IsBootstrappedResponse{}
	err := c.requester.SendRequest("isBootstrapped", &info.IsBootstrappedArgs{
		Chain: chain,
	}, res)
	return res.IsBootstrapped, err
}
