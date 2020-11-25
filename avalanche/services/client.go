package services

import (
	"fmt"
	"time"

	"github.com/ava-labs/avalanchego/api/admin"
	"github.com/ava-labs/avalanchego/api/health"
	"github.com/ava-labs/avalanchego/api/info"
	"github.com/ava-labs/avalanchego/api/ipcs"
	"github.com/ava-labs/avalanchego/api/keystore"
	"github.com/ava-labs/avalanchego/vms/avm"
	"github.com/ava-labs/avalanchego/vms/platformvm"
	"github.com/ava-labs/coreth/ethclient"
	"github.com/ava-labs/coreth/plugin/evm"
)

// Chain names
const (
	XChain = "X"
	CChain = "C"
)

// Client is a general client for avalanche
type Client struct {
	admin     *admin.Client
	xChain    *avm.Client
	health    *health.Client
	info      *info.Client
	ipcs      *ipcs.Client
	keystore  *keystore.Client
	platform  *platformvm.Client
	cChain    *evm.Client
	cChainEth *ethclient.Client
}

// NewClient returns a Client for interacting with the P Chain endpoint
func NewClient(ipAddr string, port int, requestTimeout time.Duration) (*Client, error) {
	uri := fmt.Sprintf("http://%s:%d", ipAddr, port)
	cClient, err := ethclient.Dial(fmt.Sprintf("ws://%s:%d/ext/bc/C/ws", ipAddr, port))
	if err != nil {
		return nil, err
	}
	return &Client{
		admin:     admin.NewClient(uri, requestTimeout),
		xChain:    avm.NewClient(uri, XChain, requestTimeout),
		health:    health.NewClient(uri, requestTimeout),
		info:      info.NewClient(uri, requestTimeout),
		ipcs:      ipcs.NewClient(uri, requestTimeout),
		keystore:  keystore.NewClient(uri, requestTimeout),
		platform:  platformvm.NewClient(uri, requestTimeout),
		cChain:    evm.NewCChainClient(uri, requestTimeout),
		cChainEth: cClient,
	}, nil
}

// PChainAPI ...
func (c *Client) PChainAPI() *platformvm.Client {
	return c.platform
}

// XChainAPI ...
func (c *Client) XChainAPI() *avm.Client {
	return c.xChain
}

// CChainAPI ...
func (c *Client) CChainAPI() *evm.Client {
	return c.cChain
}

// CChainEthAPI ...
func (c *Client) CChainEthAPI() *ethclient.Client {
	return c.cChainEth
}

// InfoAPI ...
func (c *Client) InfoAPI() *info.Client {
	return c.info
}

// HealthAPI ...
func (c *Client) HealthAPI() *health.Client {
	return c.health
}

// IpcsAPI ...
func (c *Client) IpcsAPI() *ipcs.Client {
	return c.ipcs
}

// KeystoreAPI ...
func (c *Client) KeystoreAPI() *keystore.Client {
	return c.keystore
}

// AdminAPI ...
func (c *Client) AdminAPI() *admin.Client {
	return c.admin
}
