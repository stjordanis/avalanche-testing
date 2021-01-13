package managedasset

import (
	"fmt"
	"strings"
	"time"

	avalancheNetwork "github.com/ava-labs/avalanche-testing/avalanche/networks"
	"github.com/ava-labs/avalanche-testing/avalanche/services"
	"github.com/ava-labs/avalanche-testing/testsuite/helpers"
	"github.com/ava-labs/avalanchego/api"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/snow/choices"
	"github.com/ava-labs/avalanchego/utils/json"
	"github.com/ava-labs/avalanchego/vms/avm"
	"github.com/palantir/stacktrace"
)

const (
	username = "username"
	password = "sjh;lqlfhja;lkajw'd;dwdadw!?!?!?"
)

var (
	fundedKey  = avalancheNetwork.DefaultLocalNetGenesisConfig.FundedAddresses.PrivateKey
	fundedAddr = avalancheNetwork.DefaultLocalNetGenesisConfig.FundedAddresses.Address
)

// Executor executes a test to ensure that the X-Chain's managed asset functionality
// works as intended.
type Executor struct {
	clients           []*services.Client
	acceptanceTimeout time.Duration
	epochDuration     time.Duration
}

// Execute the test. Returns an error if the test fails.
func (e Executor) Execute() error {
	if len(e.clients) < 1 {
		return stacktrace.NewError("executor has 0 clients")
	}
	userPass := api.UserPass{Username: username, Password: password}

	// Used to wait until something is accepted
	var waiters []*helpers.RPCWorkFlowRunner
	for _, client := range e.clients {
		waiters = append(waiters, helpers.NewRPCWorkFlowRunner(
			client,
			userPass,
			e.acceptanceTimeout,
		))
	}

	// On each node:
	// * Create a user
	// * Import the funded key
	// * Create another address and put it in [addrs]
	// * Create a managed asset and put its ID in [managedAssets]
	//   The manager/initial holder/minter of the asset is the address on the node.
	addrs := []string{} // addrs[i] is the address created by e.clients[i]
	managedAssetIDs := []ids.ID{}
	for i, client := range e.clients {
		// Create user
		_, err := client.KeystoreAPI().CreateUser(userPass)
		if err != nil {
			return stacktrace.Propagate(err, "couldn't create user")
		}

		// Import funded key
		_, err = client.XChainAPI().ImportKey(
			userPass,
			fundedKey,
		)
		if err != nil {
			return stacktrace.Propagate(err, "couldn't import funded key")
		}

		// Create another address
		addr, err := client.XChainAPI().CreateAddress(userPass)
		if err != nil {
			return stacktrace.Propagate(err, "couldn't create address")
		}
		addrs = append(addrs, addr)

		// Create a managed asset
		managedAssetID, err := e.clients[i].XChainAPI().CreateAsset(
			userPass,   // user/pass
			nil,        // from addrs
			fundedAddr, // change addr
			"yeet",     // name
			"YEET",     //symbol
			10,         // denomination
			[]*avm.Holder{
				{ // One initial holder: this node's address
					Address: addr,
					Amount:  json.Uint64(100),
				},
			},
			[]avm.Minters{
				{
					Threshold: 1,
					Minters:   []string{addr},
				},
			},
			avm.Manager{
				Threshold: 1,
				Addrs:     []string{addr},
			},
		)
		if err != nil {
			return stacktrace.Propagate(err, "couldn't create asset")

		}
		managedAssetIDs = append(managedAssetIDs, managedAssetID)

		// Wait untilthe tx is accepted
		if err := waiters[i].AwaitXChainTransactionAcceptance(managedAssetID); err != nil {
			return stacktrace.Propagate(err, "create managed asset tx not accepted")
		}

		// Make sure the initial holder got the funds
		reply, err := client.XChainAPI().GetBalance(addr, managedAssetID.String())
		if err != nil {
			return stacktrace.Propagate(err, "couldn't get balance")
		} else if uint64(reply.Balance) != 100 {
			return stacktrace.NewError("expected initial holder balance to be %d but is %d", 100, uint64(reply.Balance))
		}
	}

	// Mint 1 unit of each asset to the next node's address
	for i, client := range e.clients {
		mintAddr := addrs[(i+1)%len(e.clients)]
		txID, err := client.XChainAPI().Mint(
			userPass,                    // user/pass
			nil,                         // from addrs
			fundedAddr,                  // change addr
			1,                           // mint amount
			managedAssetIDs[i].String(), // asset ID
			mintAddr,                    // to (next node's address)
		)
		if err != nil {
			return stacktrace.Propagate(err, "couldn't mint asset")
		}
		// Wait until the tx is accepted
		if err := waiters[i].AwaitXChainTransactionAcceptance(txID); err != nil {
			return stacktrace.Propagate(err, "mint tx not accepted")
		}

		// Make sure the funds were minted
		reply, err := client.XChainAPI().GetBalance(mintAddr, managedAssetIDs[i].String())
		if err != nil {
			return err
		} else if uint64(reply.Balance) != 1 {
			return stacktrace.NewError("expected mint address balance to be %d but is %d", 1, uint64(reply.Balance))
		}

		// Send the 1 newly minted unit as manager
		toAddr := addrs[(i+2)%len(e.clients)]
		txID, err = client.XChainAPI().SendAsManager(
			userPass,                    // user/pass
			[]string{mintAddr},          // from addrs
			fundedAddr,                  // managed asset change addr
			1,                           // mint amount
			managedAssetIDs[i].String(), // asset ID
			toAddr,                      // to (next node's address)
			"",                          // memo
			fundedAddr,                  // fee change addr
			[]string{fundedAddr},        // fee from addrs
		)
		if err != nil {
			return stacktrace.Propagate(err, "couldn't send as manager")
		}
		// Wait until the tx is accepted
		if err := waiters[i].AwaitXChainTransactionAcceptance(txID); err != nil {
			return stacktrace.Propagate(err, "send as manager tx not accepted")
		}

		// Make sure the send worked
		reply, err = client.XChainAPI().GetBalance(toAddr, managedAssetIDs[i].String())
		if err != nil {
			return err
		} else if uint64(reply.Balance) != 1 {
			return stacktrace.NewError("expected balance to be %d but is %d", 1, uint64(reply.Balance))
		}

		// Change this asset's manager to the next node
		txID, err = client.XChainAPI().UpdateManagedAsset(
			userPass,
			nil,
			fundedAddr,
			managedAssetIDs[i].String(),
			false,
			avm.Manager{
				Threshold: 1,
				Addrs:     []string{addrs[(i+1)%len(e.clients)]},
			},
		)
		if err != nil {
			return stacktrace.Propagate(err, "couldn't update managed asset status")
		}
		// Wait until the tx is accepted
		if err := waiters[i].AwaitXChainTransactionAcceptance(txID); err != nil {
			return stacktrace.Propagate(err, "update managed asset tx not accepted")
		}

		status, _, err := client.XChainAPI().GetTxStatus(txID)
		if err != nil {
			return stacktrace.Propagate(err, "couldn't get tx status")
		} else if status != choices.Accepted {
			return stacktrace.NewError("expected status to be accepted but is %s", status)
		}
	}

	// Wait until 2 epochs have passed and the manager change has gone into effect
	time.Sleep(2 * e.epochDuration)

	// Now e.clients[i+1] is the manager of managedAssetIDs[i] and
	// addrs[i+2] has a balance of 1 of managedAssetIDs[i]
	// Send as manager with the new manager
	for i, assetID := range managedAssetIDs {
		client := e.clients[(i+1)%len(e.clients)]
		toAddr := addrs[(i+1)%len(e.clients)]
		txID, err := client.XChainAPI().SendAsManager(
			userPass,                              // user/pass
			[]string{addrs[(i+2)%len(e.clients)]}, // from addrs
			fundedAddr,                            // managed asset change addr
			1,                                     // mint amount
			assetID.String(),                      // asset ID
			toAddr,                                // to
			"",                                    // memo
			fundedAddr,                            // fee change addr
			[]string{fundedAddr},                  // fee from addrs
		)
		if err != nil {
			return stacktrace.Propagate(err, "couldn't send as manager")
		}
		// Wait until the tx is accepted
		if err := waiters[(i+1)%len(e.clients)].AwaitXChainTransactionAcceptance(txID); err != nil {
			return stacktrace.Propagate(err, "send as manager tx not accepted")
		}

		// Make sure the send worked
		reply, err := client.XChainAPI().GetBalance(toAddr, assetID.String())
		if err != nil {
			return err
		} else if uint64(reply.Balance) != 1 {
			return stacktrace.NewError("expected balance to be %d but is %d", 1, uint64(reply.Balance))
		}
	}
	// Now e.clients[i+1] is the manager of managedAssetIDs[i] and
	// addrs[i+1] has a balance of 1 of managedAssetIDs[i]

	// Freeze each asset
	for i, assetID := range managedAssetIDs {
		client := e.clients[(i+1)%len(e.clients)]
		// Change this asset's manager and freeze it
		txID, err := client.XChainAPI().UpdateManagedAsset(
			userPass,
			nil,              // from addrs
			fundedAddr,       // change addr
			assetID.String(), // asset ID
			true,             // frozen
			avm.Manager{ // change the manager
				Threshold: 1,
				Addrs:     []string{addrs[(i+2)%len(e.clients)]},
			},
		)
		if err != nil {
			return stacktrace.Propagate(err, "couldn't update managed asset status")
		}
		// Wait until the tx is accepted
		if err := waiters[(i+1)%len(e.clients)].AwaitXChainTransactionAcceptance(txID); err != nil {
			return stacktrace.Propagate(err, "send as manager tx not accepted")
		}
	}
	// Wait until 2 epochs have passed and the freeze has gone into effect
	time.Sleep(2 * e.epochDuration)

	// Now e.clients[i+2] is the manager of managedAssetIDs[i]
	// Try to send, which should fail due to asset being frozen
	for i, assetID := range managedAssetIDs {
		client := e.clients[(i+2)%len(e.clients)]
		toAddr := addrs[i]
		_, err := client.XChainAPI().SendAsManager(
			userPass,                              // user/pass
			[]string{addrs[(i+1)%len(e.clients)]}, // from addrs
			fundedAddr,                            // managed asset change addr
			1,                                     // mint amount
			assetID.String(),                      // asset ID
			toAddr,                                // to
			"",                                    // memo
			fundedAddr,                            // fee change addr
			[]string{fundedAddr},                  // fee from addrs
		)
		if err == nil {
			return stacktrace.NewError("expected error while sending frozen asset")
		} else if !strings.Contains(err.Error(), "frozen") {
			return stacktrace.NewError(fmt.Sprintf("expected error to mention asset being frozen but error is: %s", err))
		}
	}

	// Unfreeze each asset
	for i, assetID := range managedAssetIDs {
		client := e.clients[(i+2)%len(e.clients)]
		// Change this asset's manager and freeze it
		txID, err := client.XChainAPI().UpdateManagedAsset(
			userPass,
			nil,              // from addrs
			fundedAddr,       // change addr
			assetID.String(), // asset ID
			false,            // frozen
			avm.Manager{ // change the manager, too
				Threshold: 1,
				Addrs:     []string{addrs[(i+1)%len(e.clients)]},
			},
		)
		if err != nil {
			return stacktrace.Propagate(err, "couldn't update managed asset status")
		}

		// Wait until the tx is accepted
		if err := waiters[(i+2)%len(e.clients)].AwaitXChainTransactionAcceptance(txID); err != nil {
			return stacktrace.Propagate(err, "update managed asset tx not accepted")
		}
	}

	// Wait until 2 epochs have passed and the unfreeze / asset manager change have gone into effect
	time.Sleep(2 * e.epochDuration) // TODO change
	// Now e.clients[i+1] is the manager of managedAssetIDs[i]
	// and each asset is unfrozen

	// Try to send as manager
	for i, assetID := range managedAssetIDs {
		client := e.clients[(i+1)%len(e.clients)]
		toAddr := addrs[i]
		txID, err := client.XChainAPI().SendAsManager(
			userPass,                              // user/pass
			[]string{addrs[(i+1)%len(e.clients)]}, // from addrs
			fundedAddr,                            // managed asset change addr
			1,                                     // mint amount
			assetID.String(),                      // asset ID
			toAddr,                                // to
			"",                                    // memo
			fundedAddr,                            // fee change addr
			[]string{fundedAddr},                  // fee from addrs
		)
		if err != nil {
			return stacktrace.Propagate(err, "couldn't send as manager")
		}

		// Wait until the tx is accepted
		if err := waiters[(i+1)%len(e.clients)].AwaitXChainTransactionAcceptance(txID); err != nil {
			return stacktrace.Propagate(err, "send as manager tx not accepted")
		}
	}
	return nil
}
