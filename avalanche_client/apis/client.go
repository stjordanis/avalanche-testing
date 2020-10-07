package apis

import (
	"time"

	"github.com/ava-labs/avalanche-testing/avalanche_client/apis/admin"
	"github.com/ava-labs/avalanche-testing/avalanche_client/apis/avm"
	"github.com/ava-labs/avalanche-testing/avalanche_client/apis/health"
	"github.com/ava-labs/avalanche-testing/avalanche_client/apis/info"
	"github.com/ava-labs/avalanche-testing/avalanche_client/apis/ipcs"
	"github.com/ava-labs/avalanche-testing/avalanche_client/apis/keystore"
	"github.com/ava-labs/avalanche-testing/avalanche_client/apis/platform"
)

// Client ...
type Client struct {
	admin    *admin.Client
	xChain   *avm.Client
	health   *health.Client
	info     *info.Client
	ipcs     *ipcs.Client
	keystore *keystore.Client
	platform *platform.Client
}

// NewClient returns a Client for interacting with the P Chain endpoint
func NewClient(uri string, requestTimeout time.Duration) *Client {
	return &Client{
		admin:    admin.NewClient(uri, requestTimeout),
		xChain:   avm.NewClient(uri, "X", requestTimeout),
		health:   health.NewClient(uri, requestTimeout),
		info:     info.NewClient(uri, requestTimeout),
		ipcs:     ipcs.NewClient(uri, requestTimeout),
		keystore: keystore.NewClient(uri, requestTimeout),
		platform: platform.NewClient(uri, requestTimeout),
	}
}

// PChainAPI ...
func (c *Client) PChainAPI() *platform.Client {
	return c.platform
}

// XChainAPI ...
func (c *Client) XChainAPI() *avm.Client {
	return c.xChain
}

// InfoAPI ...
func (c *Client) InfoAPI() *info.Client {
	return c.info
}

// HealthAPI ...
func (c *Client) HealthAPI() *health.Client {
	return c.health
}

// IPCSAPI ...
func (c *Client) IPCSAPI() *ipcs.Client {
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
