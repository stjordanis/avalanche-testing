package main

import (
	"errors"
	"fmt"
	"time"
	"github.com/ava-labs/avalanche-testing/avalanche_client/apis"
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
func awaitSetup(clients ...*apis.Client) error {
	for _, client := range clients {
		if healthy, _ := client.HealthAPI().AwaitHealthy(4, 3*time.Second); healthy {
			return nil
		}
	}
	return errUnhealthy
}

// return clients to interact with each avash node
func fiveNodeAvashClients() []*apis.Client {
	clients := make([]*apis.Client, 5)
	for i := 0; i < 5; i++ {
		uri := fmt.Sprintf("http://127.0.0.1:%d", 9650+i*2)
		clients[i] = apis.NewClient(uri, requestTimeout)
	}
	return clients
}
/**
[leveldb]
52.53.202.86
54.183.208.204
[rocksdb]
13.52.254.181
18.144.100.250
[rocksdb_default]
54.183.154.212
13.52.98.57
[rocksdb_multi]
13.52.250.71
54.183.6.92
 */
// return clients for each of the devnet nodes
func privateNetClients() []*apis.Client {
	return []*apis.Client{
		//leveldb
		apis.NewClient("http://52.53.202.86:21000", requestTimeout),
		apis.NewClient("http://54.183.208.204:21000", requestTimeout),
		//rocksdb
		apis.NewClient("http://13.52.254.181:21000", requestTimeout),
		apis.NewClient("http://18.144.100.250:21000", requestTimeout),
	}
}

func bombardNetwork(clients []*apis.Client, numTxs, txFee uint64, done chan error) {
	executor := bombard.NewBombardExecutor(clients, numTxs, txFee, 10*time.Second)
	if err := executor.ExecuteTest(); err != nil {
		fmt.Printf("Test failed: %s\n", err)
		done <- fmt.Errorf("bombardNetwork failed due to: %s", err)
		return
	}
	done <- nil
}

func main() {
	// clients := fiveNodeAvashClients()
	clients := privateNetClients()

	if err := awaitSetup(clients...); err != nil {
		fmt.Printf("Test did not run due to: %s\n", err)
		return
	}
	numTxs := uint64(100000)
	txFee := uint64(1000000)

	executor := bombard.NewBombardExecutor(clients, numTxs, txFee, 10*time.Second)
	if err := executor.ExecuteTest(); err != nil {
		fmt.Printf("Test failed: %s\n", err)
	}

	fmt.Printf("Test finished successfully.")
}
