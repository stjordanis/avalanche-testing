package bombard

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/ava-labs/avalanche-e2e-tests/commons/ava_testsuite/rpc_workflow_runner"
	"github.com/ava-labs/avalanche-e2e-tests/gecko_client/apis"
	"github.com/ava-labs/gecko/api"
	"github.com/palantir/stacktrace"
)

var (
	requestTimeout  = 2 * time.Second
	uri             = "http://127.0.0.1:9650"
	privateKey      = "PrivateKey-ewoqjP7PxY4yr3iLTpLisriqt94hdyDFNgchSxGGztUrTXtNN"
	genesisUsername = "aoufbnoeur2r"
	genesisPassword = "vwnlu24ir723irg23irfg3ho"
	userPass        = api.UserPass{Username: genesisUsername, Password: genesisPassword}
)

type executor struct {
	normalClients     []*apis.Client
	acceptanceTimeout time.Duration
	numTxs            int
	txFee             uint64
}

func createRandomString() string {
	return fmt.Sprintf("rand:%d", rand.Int())
}

func (e *executor) ExecuteTest() error {
	genesisClient := e.normalClients[0]
	secondaryClients := make([]*rpc_workflow_runner.RPCWorkFlowRunner, len(e.normalClients)-1)
	xChainAddrs := make([]string, len(e.normalClients)-1)
	for i, client := range e.normalClients[1:] {
		secondaryClients[i] = rpc_workflow_runner.NewRPCWorkFlowRunner(
			client,
			api.UserPass{Username: createRandomString(), Password: createRandomString()},
			e.acceptanceTimeout,
		)
		xChainAddress, _, err := secondaryClients[i].CreateDefaultAddresses()
		if err != nil {
			return stacktrace.Propagate(err, "Failed to create default addresses for client: %d", i)
		}
		xChainAddrs[i] = xChainAddress
	}

	highLevelGenesisClient := rpc_workflow_runner.NewRPCWorkFlowRunner(
		genesisClient,
		api.UserPass{Username: genesisUsername, Password: genesisPassword},
		e.acceptanceTimeout,
	)

	if _, err := highLevelGenesisClient.ImportGenesisFunds(); err != nil {
		return stacktrace.Propagate(err, "Failed to fund genesis client.")
	}
	// Fund X Chain Addresses enough to issue numTxs and 1 new asset
	seedAmount := uint64(e.numTxs+1) * e.txFee
	if err := highLevelGenesisClient.FundXChainAddresses(xChainAddrs, seedAmount); err != nil {
		return stacktrace.Propagate(err, "Failed to fund X Chain Addresses for Clients")
	}

	for i, client := range secondaryClients {
		if err := client.VerifyXChainAVABalance(xChainAddrs[i], seedAmount); err != nil {
			return stacktrace.Propagate(err, "Failed to verify X Chain Balane for Client: %d", i)
		}
	}

	return nil
}
