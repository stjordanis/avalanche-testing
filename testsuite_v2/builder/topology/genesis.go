package topology

import (
	"time"

	"github.com/ava-labs/avalanche-testing/testsuite_v2/builder/chainhelper"

	"github.com/ava-labs/avalanche-testing/avalanche/services"

	"github.com/kurtosis-tech/kurtosis-go/lib/testsuite"

	avalancheNetwork "github.com/ava-labs/avalanche-testing/avalanche/networks"
	"github.com/ava-labs/avalanchego/api"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
)

// Genesis is a single attribute of the Topology (only one Genesis) that holds the Genesis data
type Genesis struct {
	id       string
	client   *services.Client
	Address  string
	userPass api.UserPass
	context  *testsuite.TestContext
}

func newGenesis(id string, userName string, password string, client *services.Client, context *testsuite.TestContext) *Genesis {
	return &Genesis{
		id: id,
		userPass: api.UserPass{
			Username: userName,
			Password: password,
		},
		client:  client,
		context: context,
	}
}

// ImportGenesisFunds fetches the default funded funds and imports them
func (g *Genesis) ImportGenesisFunds() error {

	var err error

	keystore := g.client.KeystoreAPI()
	if _, err = keystore.CreateUser(g.userPass); err != nil {
		return stacktrace.Propagate(err, "Failed to take create genesis user account.")
	}

	g.Address, err = g.client.XChainAPI().ImportKey(
		g.userPass,
		avalancheNetwork.DefaultLocalNetGenesisConfig.FundedAddresses.PrivateKey)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to take control of genesis account.")
	}
	logrus.Infof("Genesis Address: %s.", g.Address)
	return nil
}

// FundXChainAddresses funds the genesis funds into an address on the XChain
func (g *Genesis) FundXChainAddresses(addresses []string, amount uint64) *Genesis {
	for _, address := range addresses {
		txID, err := g.client.XChainAPI().Send(
			g.userPass,
			nil, // from addrs
			"",  // change addr
			amount,
			"AVAX",
			address,
			"",
		)
		if err != nil {
			g.context.Fatal(stacktrace.Propagate(err, "Failed to fund addresses in genesis."))
			return g
		}

		// wait for the tx to go through
		err = chainhelper.XChain().AwaitTransactionAcceptance(g.client, txID, 30*time.Second)
		if err != nil {
			g.context.Fatal(err)
			return g
		}

		// verify the balance
		err = chainhelper.XChain().CheckBalance(g.client, address, "AVAX", amount)
		if err != nil {
			g.context.Fatal(err)
			return g
		}

		logrus.Infof("Funded X Chain Address: %s with %d.", address, amount)
	}

	return g
}
