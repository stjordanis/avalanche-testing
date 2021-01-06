package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/ava-labs/avalanche-testing/avalanche/services"
	"github.com/ava-labs/avalanche-testing/testsuite/tests/bombard"
)

var (
	errUnhealthy = errors.New("node was not marked as healthy")
)

const (
	stakerClientURI    = "http://127.0.0.1:9660"
	delegatorClientURI = "http://127.0.0.1:9662"
	requestTimeout     = 10 * time.Second
)

// cd $GOPATH
// go get -v -d github.com/ava-labs/avalanche-testing
// cd $GOPATH/src/github.com/ava-labs/avalanche-testing
// go mod download
// copy the contents of this file into a file called bombard_once.go
// go build -o bombard_once bombard_once.go
// ./bombard_once

// This will start a script to pre-generate a list of 100K txs for each node
// and then send the transactions in order to one of the nodes (one list for each node)
// In order to push the TPS as high as possible, start around 10 of these scripts.

// The funds for these transactions will come from the funded genesis address on the
// default local network. Therefore, enough funds for 100K txs are transferred to a newly
// generated address for each node. This way, the transactions will never conflict and there
// will not be any missing UTXOs since the transactions are issued in order.

// Since the funds are transferred from the genesis address with the Send API, starting
// multiple instances of this script must be spaced out until the transactions from the
// genesis address have been accepted on the network to avoid causing the Send API to create
// a conflicting transaction that would then be dropped before being issued to the network.

// Therefore, for best results and to push the TPS, you should start one script after the other
// once the first script says: "Creating string of X transactions..."

// You can start the scripts in the background as follows:
// nohup ./bombard-once > b0.out 2>&1 &

// wait until each node reports healthy and return an error if it takes too long
func awaitSetup(clients ...*services.Client) error {
	for _, client := range clients {
		if healthy, _ := client.HealthAPI().AwaitHealthy(4, 3*time.Second); healthy {
			return nil
		}
	}
	return errUnhealthy
}

// return clients to interact with each avash node
func fiveNodeAvashClients() ([]*services.Client, error) {
	clients := make([]*services.Client, 5)
	for i := 0; i < 5; i++ {
		client, err := services.NewClient("127.0.0.1", 9650+i*2, requestTimeout)
		if err != nil {
			return nil, err
		}
		clients[i] = client
	}
	return clients, nil
}

// return clients for each of the devnet nodes
func devNetClients() ([]*services.Client, error) {
	clients := make([]*services.Client, 0, 5)
	for _, ip := range []string{"34.201.137.119", "3.90.138.89", "54.145.81.99", "54.146.1.110", "54.91.255.231"} {
		client, err := services.NewClient(ip, 21000, requestTimeout)
		if err != nil {
			return nil, err
		}
		clients = append(clients, client)
	}
	return clients, nil
}

// bombardNetwork bombars the clients with [numTxs] each and adds an error or nil
// to [done] when it has completed
func bombardNetwork(clients []*services.Client, numTxs, txFee uint64, done chan error) {
	executor := bombard.NewBombardExecutor(clients, numTxs, txFee, 10*time.Second)
	if err := executor.ExecuteTest(); err != nil {
		fmt.Printf("Test failed: %s\n", err)
		done <- fmt.Errorf("bombardNetwork failed due to: %s", err)
		return
	}
	done <- nil
}

func main() {
	// clients, err := devNetClients()
	clients, err := fiveNodeAvashClients()
	if err != nil {
		fmt.Printf("Failed to create clients due to %s\n", err)
		return
	}

	if err := awaitSetup(clients...); err != nil {
		fmt.Printf("Test did not run due to: %s\n", err)
		return
	}
	numTxs := uint64(7500)
	txFee := uint64(1000000)

	executor := bombard.NewBombardExecutor(clients, numTxs, txFee, 10*time.Second)
	if err := executor.ExecuteTest(); err != nil {
		fmt.Printf("Test failed: %s\n", err)
		return
	}

	fmt.Printf("Test finished successfully.\n")
}
